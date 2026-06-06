# 多租户登录设计方案

## 核心设计

- **用户-租户关系**：用户属于单个租户，通过 `tenant_id` 绑定
- **租户 ID**：所有租户统一使用 idgen 生成的 Sonyflake ID（18 位），包括默认租户
- **默认租户**：通过 `tenant_code = 'default'` 标识，应用启动时加载到缓存 `cache.Get().Tenant.GetDefaultTenantID()`
- **权限控制**：Casbin `(username, tenantCode, roleCode)` 三元组
- **超管判断**：Token 中携带 `roles` 数组，包含 `super_admin` 即为超管
- **租户隔离**：登录时通过 URL 路径参数，业务接口从 Token 获取

## 数据库设计

### 租户表

```sql
CREATE TABLE tenants (
    tenant_id VARCHAR(20) PRIMARY KEY,
    tenant_code VARCHAR(50) NOT NULL UNIQUE,
    tenant_name VARCHAR(255) NOT NULL
);

INSERT INTO tenants (tenant_id, tenant_code, tenant_name) VALUES
('153547313510524266', 'default', '默认租户'),
('268447313510524267', 'company-a', '公司A');
```

### 用户表

```sql
CREATE TABLE users (
    user_id VARCHAR(20) PRIMARY KEY,
    tenant_id VARCHAR(20) NOT NULL,
    user_name VARCHAR(100) NOT NULL,
    password VARCHAR(100) NOT NULL,
    name VARCHAR(100) NOT NULL DEFAULT '',
    avatar VARCHAR(255),
    phone VARCHAR(20),
    email VARCHAR(100),
    status SMALLINT NOT NULL DEFAULT 1,
    remark TEXT,
    last_login_time BIGINT,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0,
    UNIQUE KEY uk_tenant_username (tenant_id, user_name) WHERE deleted_at = 0
);
```

### 角色表

```sql
CREATE TABLE roles (
    role_id VARCHAR(20) PRIMARY KEY,
    tenant_id VARCHAR(20) NOT NULL,
    role_code VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    status SMALLINT NOT NULL DEFAULT 1,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0,
    UNIQUE KEY uk_tenant_code(tenant_id, role_code) WHERE deleted_at = 0
);

INSERT INTO roles (role_id, tenant_id, name, role_code) VALUES
('role-super-001', '153547313510524266', '超级管理员', 'super_admin'),
('role-sales-001', '153547313510524266', '销售角色', 'sales');
```

### 用户角色关联表

```sql
CREATE TABLE user_roles (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    role_id VARCHAR(36) NOT NULL,
    tenant_id VARCHAR(20) NOT NULL,
    UNIQUE KEY uk_user_role(user_id, role_id)
);
```

## Casbin 配置

### 模型

```conf
[request_definition]
r = sub, dom, obj, act

[role_definition]
g = _, _, _  # (user, role, domain)
g2 = _, _    # 角色继承

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act)
```

### 策略示例

```conf
# 用户角色
g, admin, default, super_admin
g, zhangsan, company-a, tenant-a-sales

# 角色继承
g2, tenant-a-sales, sales

# 角色权限
p, super_admin, default, *, *
p, sales, default, menu:orders, *
```

## 路由设计

```go
// 公开接口 - 从路径获取租户
publicGroup := r.Group("/api/v1/:tenant_code")
publicGroup.Use(middlewares.TenantMiddleware())
{
    publicGroup.POST("/login", authHandler.Login)
}

// 认证接口 - 从 Token 获取租户
authGroup := r.Group("/api/v1")
authGroup.Use(middlewares.AuthMiddleware())
{
    authGroup.GET("/users", userHandler.ListUsers)
}
```

**URL 示例**:
- 登录: `POST /api/v1/default/login`
- 业务: `GET /api/v1/users`

## Token 设计

```json
{
  "user_id": "user-super-001",
  "username": "admin",
  "tenant_id": "153547313510524266",
  "tenant_code": "default",
  "roles": ["super_admin", "sales"],
  "exp": 1734567890
}
```

## 中间件

### TenantMiddleware（公开接口）

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

### AuthMiddleware（认证接口）

```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        claims := parseToken(c.GetHeader("Authorization"))
        c.Set("user", claims)
        c.Set("tenant_id", claims.TenantID)
        c.Set("tenant_code", claims.TenantCode)

        // 超管跳过权限检查
        isSuperAdmin := slices.Contains(claims.Roles, "super_admin")
        if !isSuperAdmin {
            allowed, _ := enforcer.Enforce(claims.Username, claims.TenantCode, c.Request.URL.Path, c.Request.Method)
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

## 缓存设计

使用 `pkg/cache` 包管理默认租户缓存：

```go
// 应用启动时初始化
cache.Init(db)

// 获取默认租户ID
tenantID := cache.Get().Tenant.GetDefaultTenantID()

// 判断是否为默认租户
cache.Get().Tenant.IsDefaultTenant(tenantID)
```

**优势**：
- 消除魔法值 `000000000000000000`
- 默认租户和普通租户使用相同的 ID 生成规则
- 便于迁移和管理

## 常量

```go
const (
    DefaultTenantCode = "default"      // 默认租户编码
    SuperAdminRoleCode = "super_admin" // 超级管理员角色编码
)
```

## 实现说明

### 租户查询实现

**方案**：登录接口在 Handler 中直接调用 `TenantRepo.GetByCodeManual()` 查询租户，不使用中间件。

**理由**：
- 登录是唯一需要从路径获取租户的公开接口（注册、验证码不需要租户信息）
- 忘记密码等未来可能添加的公开接口可复用相同方式
- 避免中间件过度设计，代码更直观

**Repo 层封装**：
```go
// GetByCodeManual 根据租户编码获取租户信息（手动模式，用于跨租户查询）
// 使用场景：登录时需要通过租户code查询租户，此时还没有当前租户信息
func (r *TenantRepo) GetByCodeManual(ctx context.Context, tenantCode string) (*model.Tenant, error) {
    // 跨租户查询：使用 ManualTenantMode 禁止自动添加当前租户过滤
    ctx = database.ManualTenantMode(ctx)
    return r.q.Tenant.WithContext(ctx).
        Where(r.q.Tenant.TenantCode.Eq(tenantCode)).
        First()
}
```

### 路由差异

| 项目 | 设计文档 | 实际实现 |
|------|----------|----------|
| 路径格式 | `/api/v1/:tenant_code/login` | `/api/v1/auth/:tenant_code/login` |
| 分组结构 | 独立 `/api/v1/:tenant_code` 分组 | `/auth` 分组下，应用 `SkipTenantCheck` 中间件 |
| 中间件 | `TenantMiddleware` | 不使用租户中间件，Handler 内调用 `GetByCodeManual` |

**实际路由**：
```go
auth := public.Group("/auth")
{
    auth.GET("/captcha", app.Handlers.CaptchaHandler.Get)            // 无需租户
    auth.POST("/refresh", app.Handlers.AuthHandler.Refresh)          // 无需租户
    auth.POST("/:tenant_code/login", app.Handlers.AuthHandler.Login) // Handler 内查询租户
}
```

## 修改清单

| 文件 | 修改内容 |
|------|----------|
| `backend/scripts/dev_schema.sql` | 新增 tenants 表，所有 ID 字段统一为 VARCHAR(20) |
| `backend/scripts/init_data/main.go` | 插入默认租户和角色记录 |
| `backend/pkg/cache/cache.go` | 缓存管理器 |
| `backend/pkg/cache/tenant.go` | 租户缓存实现 |
| `backend/pkg/constants/system.go` | 添加租户常量（移除 DefaultTenantID） |
| `backend/internal/middleware/tenant.go` | 租户中间件（设计参考，实际未使用） |
| `backend/internal/middleware/auth.go` | 认证中间件（包含 super_admin 则跳过权限检查） |
| `backend/internal/model/user.go` | ID 字段改为 VARCHAR(20) |
| `backend/internal/model/role.go` | ID 字段改为 VARCHAR(20) |
| `backend/internal/repository/tenant_repo.go` | 添加 `GetByCodeManual` 方法 |
| `backend/internal/service/auth_service.go` | 登录时查询用户角色并写入 Token |
| `frontend/src/api/auth.ts` | 登录 API 添加 tenant_code |
| `frontend/src/views/Login.vue` | 登录页添加租户输入 |
