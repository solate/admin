# 多租户登录设计方案

> **架构模式**：用户属于单个租户，通过 `tenant_id` 绑定
> **Casbin 集成**：使用 `(username, tenantCode, roleCode)` 三元组，权限由角色控制
> **user_type 冗余**：Token 中携带 user_type，避免每次请求都查询 Casbin 判断是否超管
> **租户隔离**：登录时通过 URL 路径参数，业务接口从 Token 获取租户

---

## 1. 设计概述

- **租户 ID 设计**：
  - **默认租户**：`tenant_id` 为空字符串 `""`，`tenant_code` 为 `"default"`
  - **其他租户**：`tenant_id` 为 UUID 值，`tenant_code` 为自定义编码
  - 这样设计的好处：默认租户数据与普通租户数据在同一表中明确区分
- **权限由角色控制**：通过 Casbin 的角色机制管理权限（角色继承、权限分配）
- **user_type 冗余字段**：`1` 普通用户、`2` 租户管理员、`3` 超级管理员
  - 作用：Token 中携带，中间件直接判断是否超管，避免查询 Casbin + Roles 表
  - 注意：真实权限仍由 Casbin 的 role 控制，user_type 只是性能优化
- **Casbin domain**：直接使用 `tenant_code`，默认租户使用 `"default"`
- **数据隔离**：
  - Repository 层通过 `tenant_id` 过滤数据
  - 默认租户：`WHERE tenant_id = ''`
  - 其他租户：`WHERE tenant_id = '具体的UUID'`

---

## 2. 数据库设计

### 2.1 租户表 (tenants)

```sql
CREATE TABLE tenants (
    tenant_id VARCHAR(36) PRIMARY KEY,
    tenant_code VARCHAR(50) NOT NULL UNIQUE,
    tenant_name VARCHAR(255) NOT NULL
);

-- 初始化数据（默认租户不需要插入，tenant_id 为空即表示默认租户）
INSERT INTO tenants (tenant_id, tenant_code, tenant_name) VALUES
('tenant-001', 'company-a', '公司A'),
('tenant-002', 'company-b', '公司B');
```

### 2.2 用户表 (users)

```sql
CREATE TABLE users (
    user_id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    user_name VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    user_type TINYINT NOT NULL DEFAULT 1,
    UNIQUE KEY uk_tenant_username (tenant_id, user_name)
);

-- 初始化数据
INSERT INTO users (user_id, tenant_id, user_name, password, user_type) VALUES
('user-super-001', '', 'admin', 'hashed_password', 3),              -- 超管（默认租户，tenant_id 为空）
('user-admin-001', 'tenant-001', 'admin', 'hashed_password', 2),     -- 租户管理员
('user-001', 'tenant-001', 'zhangsan', 'hashed_password', 1);        -- 普通用户
```

### 2.3 角色表 (roles)

```sql
CREATE TABLE roles (
    role_id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL,
    UNIQUE KEY uk_tenant_code(tenant_id, code)
);

-- 初始化数据
INSERT INTO roles (role_id, tenant_id, name, code) VALUES
('role-super-001', '', '超级管理员', 'super_admin'),
('role-sales-001', '', '销售角色', 'sales');
```

**注意**：`parent_id` 已删除，角色继承通过 Casbin `g2` 策略管理。

---

## 3. Casbin 设计

### 3.1 模型配置

```conf
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act

[role_definition]
g = _, _, _  # 用户-角色-租户: (user, role, domain)
g2 = _, _    # 角色继承: (child_role, parent_role)

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act)
```

### 3.2 策略示例

```conf
# g 策略：用户角色绑定 (user, role, domain)
g, admin, default, super_admin
g, zhangsan, company-a, tenant-a-sales

# g2 策略：角色继承 (child, parent) - 不需要 domain
g2, tenant-a-sales, sales

# p 策略：角色权限 (role, domain, resource, action)
p, super_admin, default, *, *
p, sales, default, menu:orders, *
p, sales, default, btn:order_create, *
```

---

## 4. 路由设计

```
公开接口（无 Token）：  /api/v1/:tenant_code/*  → 从路径获取租户
认证接口（有 Token）：  /api/v1/*             → 从 Token 获取租户
```

### 4.1 路由示例

```go
// 公开接口
publicGroup := r.Group("/api/v1/:tenant_code")
publicGroup.Use(middlewares.TenantMiddleware())
{
    publicGroup.POST("/login", authHandler.Login)
}

// 认证接口
authGroup := r.Group("/api/v1")
authGroup.Use(middlewares.AuthMiddleware())
{
    authGroup.GET("/users", userHandler.ListUsers)
}
```

### 4.2 URL 示例

| 接口 | URL |
|------|-----|
| 超管登录 | `POST /api/v1/default/login` |
| 租户A登录 | `POST /api/v1/company-a/login` |
| 用户列表 | `GET /api/v1/users` |

---

## 5. JWT Token 设计

### 5.1 Token Payload

```json
{
  "user_id": "user-super-001",
  "username": "admin",
  "tenant_id": "",
  "tenant_code": "default",
  "user_type": 3,
  "exp": 1734567890
}
```

### 5.2 登录响应

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user_id": "user-super-001",
  "username": "admin",
  "tenant_code": "default",
  "user_type": 3
}
```

---

## 6. 中间件设计

### 6.1 TenantMiddleware（公开接口）

```go
func TenantMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tenantCode := c.Param("tenant_code")

        tenant, err := tenantRepo.GetByCode(c, tenantCode)
        if err != nil {
            c.JSON(404, gin.H{"message": "租户不存在"})
            c.Abort()
            return
        }

        c.Set("tenant_id", tenant.TenantID)
        c.Set("tenant_code", tenant.TenantCode)
        c.Next()
    }
}
```

### 6.2 AuthMiddleware（认证接口）

```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        claims, err := parseToken(token)
        if err != nil {
            c.JSON(401, gin.H{"message": "未授权"})
            c.Abort()
            return
        }

        c.Set("user", claims)
        c.Set("username", claims.Username)
        c.Set("tenant_code", claims.TenantCode)
        c.Set("user_type", claims.UserType)

        // 超管跳过权限检查，其他用户需要 Casbin 验证
        if claims.UserType != 3 {
            path := c.Request.URL.Path
            method := c.Request.Method
            allowed, _ := enforcer.Enforce(claims.Username, claims.TenantCode, path, method)
            if !allowed {
                c.JSON(403, gin.H{"message": "无权限"})
                c.Abort()
                return
            }
        }

        c.Next()
    }
}
```

---

## 7. 常量定义

```go
package constants

const (
    // 租户
    DefaultTenantID   = ""       // 默认租户 ID 为空字符串
    DefaultTenantCode = "default" // 默认租户 code

    // 用户类型
    UserTypeUser        = 1 // 普通用户
    UserTypeTenantAdmin = 2 // 租户管理员
    UserTypeSuperAdmin  = 3 // 超级管理员

    // 角色
    SuperAdminRoleCode = "super_admin"
)
```

---

## 8. 核心业务逻辑

### 8.1 登录

```go
func (h *AuthHandler) Login(c *gin.Context) {
    tenantCode := c.Param("tenant_code")

    var req LoginRequest
    c.BindJSON(&req)

    // 查询租户
    tenant, err := h.tenantRepo.GetByCode(c, tenantCode)
    if err != nil {
        c.JSON(404, gin.H{"message": "租户不存在"})
        return
    }

    // 查询用户
    user, err := h.userRepo.GetByUsernameAndTenant(c, req.Username, tenant.TenantID)
    if err != nil || !verifyPassword(req.Password, user.Password) {
        c.JSON(401, gin.H{"message": "用户名或密码错误"})
        return
    }

    // 生成 Token
    token := generateToken JWTClaims{
        UserID:     user.UserID,
        Username:   user.UserName,
        TenantID:   user.TenantID,
        TenantCode: tenant.TenantCode,
        UserType:   user.UserType,
    }

    c.JSON(200, gin.H{
        "access_token": token,
        "user_type":    user.UserType,
    })
}
```

### 8.2 权限检查

```go
func (s *UserService) HasPermission(userID, tenantCode, resource, action string) bool {
    allowed, _ := s.enforcer.Enforce(userID, tenantCode, resource, action)
    return allowed
}
```

### 8.3 租户创建角色（继承超管角色模板）

```go
func (s *RoleService) CreateRole(ctx context.Context, req *CreateRoleRequest) error {
    tenantCode := getTenantCode(ctx)

    // 如果有父角色，验证父角色属于默认租户（tenant_id 为空）
    if req.ParentRoleCode != nil {
        parent, _ := s.roleRepo.GetByCode(ctx, *req.ParentRoleCode)
        if parent.TenantID != constants.DefaultTenantID {
            return errors.New("只能继承平台角色模板")
        }
    }

    // 创建角色
    role := &Role{
        RoleID:   uuid.New(),
        TenantID: getTenantID(ctx),
        Code:     req.Code,
        Name:     req.Name,
    }
    s.roleRepo.Create(ctx, role)

    // Casbin g2: 创建角色继承关系 (child, parent) - 注意 g2 不需要 domain
    if req.ParentRoleCode != nil {
        s.enforcer.AddGroupingPolicy(role.Code, *req.ParentRoleCode)
    }

    return nil
}
```

---

## 9. 文件修改清单

### 后端

| 文件 | 修改内容 |
|------|----------|
| `backend/scripts/dev_schema.sql` | 新增 tenants 表，users 表添加 user_type，tenant_id 改为 NOT NULL |
| `backend/pkg/constants/system.go` | 添加 DefaultTenantID/Code、UserType 常量 |
| `backend/pkg/casbin/super_admin.go` | 使用 "default" 作为 domain |
| `backend/internal/middleware/tenant.go` | 租户中间件（从路径获取） |
| `backend/internal/middleware/auth.go` | 认证中间件（从 Token 获取，user_type=3 跳过权限检查） |
| `backend/internal/model/user.go` | 添加 UserType 字段 |
| `backend/internal/router/router.go` | 添加 `:tenant_code` 路由 |
| `backend/internal/repository/user_repo.go` | 支持按 tenant_id 条件查询 |
| `backend/internal/service/auth_service.go` | 登录逻辑调整 |

### 前端

| 文件 | 修改内容 |
|------|----------|
| `frontend/src/api/auth.ts` | 更新登录 API，添加 tenant_code 参数 |
| `frontend/src/views/Login.vue` | 登录页面添加租户选择或输入 |
