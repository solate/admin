# 多租户登录设计方案（方案一：单租户用户）

> **架构模式**：用户属于单个租户，通过 `tenant_id` 绑定
> **超管支持**：平台超管（`tenant_id` 为 NULL）+ 租户超管
> **租户隔离**：登录时通过 URL 路径参数，业务接口从 Token 获取租户

---

## 1. 设计概述

- **用户绑定租户**：每个用户属于一个租户（平台超管的 `tenant_id` 为 NULL）
- **租户内唯一用户名**：用户名在租户内唯一，不同租户可有同名用户
- **登录时区分租户**：通过 URL 路径 `tenant_code` 区分
- **业务接口简洁**：Token 中已包含 `tenant_id`，无需路径参数

---

## 2. 数据库设计

### 2.1 用户表 (users)

```sql
CREATE TABLE users (
    user_id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(36),               -- 平台超管为 NULL，租户用户有值
    user_name VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    ...
    UNIQUE KEY uk_tenant_username (tenant_id, user_name)
);
```

**用户名唯一性**：
- 平台超管：`(NULL, 'admin')` 全局唯一
- 租户用户：`('tenant-a', 'admin')` 租户内唯一
- 平台超管和租户用户可以有相同的用户名（因为 NULL != 'tenant-a'）

### 2.2 租户表 (tenants)

```sql
CREATE TABLE tenants (
    tenant_id VARCHAR(36) PRIMARY KEY,
    tenant_code VARCHAR(50) NOT NULL UNIQUE,   -- 租户编码，用于 URL
    tenant_name VARCHAR(255) NOT NULL,
    ...
);
```

---

## 3. 路由设计

### 3.1 路由结构

```
公开接口（无 Token）：  /api/v1/:tenant_code/*  → 从路径获取租户
认证接口（有 Token）：  /api/v1/*             → 从 Token 获取租户
```

**示例**：
```go
// 公开接口 - 需要路径参数
publicGroup := r.Group("/api/v1/:tenant_code")
publicGroup.Use(middlewares.TenantMiddleware())
{
    publicGroup.POST("/login", authHandler.Login)
    publicGroup.POST("/captcha", captchaHandler.GetCaptcha)
}

// 认证接口 - 从 Token 获取 tenant_id
authGroup := r.Group("/api/v1")
authGroup.Use(middlewares.AuthMiddleware())
{
    authGroup.GET("/users", userHandler.ListUsers)
    authGroup.GET("/users/:user_id", userHandler.GetUser)
}
```

### 3.2 URL 示例

| 接口 | URL | 租户获取方式 |
|------|-----|-------------|
| 租户A登录 | `POST /api/v1/company-a/login` | 从路径 |
| 平台超管登录 | `POST /api/v1/platform/login` | 从路径 |
| 用户列表 | `GET /api/v1/users` | 从 Token |
| 用户详情 | `GET /api/v1/users/:user_id` | 从 Token |

---

## 4. 中间件设计

### 4.1 TenantMiddleware（仅用于公开接口）

从 URL 路径提取 `tenant_code`，查询租户信息：

```go
func TenantMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tenantCode := c.Param("tenant_code")

        // platform 表示平台超管
        if tenantCode == "platform" {
            c.Set("tenant_id", nil)
            c.Next()
            return
        }

        // 查询租户
        tenant, err := tenantRepo.GetByCode(c, tenantCode)
        if err != nil {
            c.JSON(404, gin.H{"message": "租户不存在"})
            c.Abort()
            return
        }

        c.Set("tenant_id", tenant.TenantID)
        c.Next()
    }
}
```

### 4.2 AuthMiddleware（用于业务接口）

从 Token 解析用户信息（含 `tenant_id`）：

```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")

        // 解析 Token 获取用户信息（包含 tenant_id）
        user, err := parseToken(token)
        if err != nil {
            c.JSON(401, gin.H{"message": "未授权"})
            c.Abort()
            return
        }

        c.Set("user", user)
        c.Set("tenant_id", user.TenantID)
        c.Next()
    }
}
```

---

## 5. 登录流程

```
用户访问 /api/v1/:tenant_code/login
       │
       ▼
提取 tenant_code → 查询租户 → 设置 tenant_id
       │
       ▼
验证用户名+密码（带上 tenant_id 条件）
       │
       ├───────────┬───────────┐
       ▼           ▼           ▼
  平台超管    租户用户    登录失败
  返回 Token  返回 Token   返回错误
```

---

## 6. API 设计

### 登录接口

```http
POST /api/v1/company-a/login
{
  "username": "admin",
  "password": "password"
}
```

**响应**：
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user_id": "user-456",
  "tenant_id": "tenant-123",
  "user_type": "tenant_user"
}
```

**Token Payload**：
```json
{
  "user_id": "user-456",
  "tenant_id": "tenant-123",
  "user_type": "tenant_user",
  "exp": 1734567890
}
```

### 业务接口

```http
GET /api/v1/users
Authorization: Bearer <token>
```

---

## 7. 权限控制

业务逻辑中通过 `tenant_id` 判断：

```go
func (s *UserService) ListUsers(ctx context.Context) ([]*model.User, error) {
    currentUser, _ := ctx.Get("user").(*model.User)

    // 平台超管：查询所有
    if currentUser.TenantID == nil {
        return s.repo.ListAll(ctx)
    }

    // 租户用户：只查询本租户
    return s.repo.ListByTenantID(ctx, *currentUser.TenantID)
}
```

---

## 8. 初始化数据

### 创建租户

```sql
INSERT INTO tenants (tenant_id, tenant_code, tenant_name) VALUES
('tenant-001', 'company-a', '公司A'),
('tenant-002', 'company-b', '公司B');
```

每个租户创建时自动生成 `admin` 用户（`tenant_id` 为该租户ID）。

### 创建平台超管

```sql
INSERT INTO users (user_id, tenant_id, user_name, password) VALUES
('platform-super-1', NULL, 'admin', 'hashed_password');
```

---

## 9. 文件修改清单

### 后端

| 文件 | 修改内容 |
|------|----------|
| `backend/scripts/dev_schema.sql` | 用户表 `tenant_id` 可为 NULL，新增租户表 |
| `backend/internal/middleware/tenant.go` | 租户中间件（登录等公开接口） |
| `backend/internal/middleware/auth.go` | 认证中间件（业务接口） |
| `backend/internal/model/user.go` | `TenantID` 改为指针类型 |
| `backend/internal/router/router.go` | 添加 `:tenant_code` 路由 |
| `backend/internal/repository/user_repo.go` | 支持按 `tenant_id` 条件查询 |

### 前端

| 文件 | 修改内容 |
|------|----------|
| `frontend/src/api/auth.ts` | 更新登录 API |
| `frontend/src/utils/token.ts` | Token 管理 |
| `frontend/src/views/Login.vue` | 登录页面 |
