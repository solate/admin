# Step 06: 登录认证 — 登录 / 登出 / 刷新 Token / 租户切换

## 目标

实现完整的认证流程：登录获取 Token、刷新 Token、登出注销、租户切换。这是第一个完整的业务域实现，确立 handler → service → repository 的标准模式。

## 前置条件

- Step 05 完成，JWT Manager 和 Auth 中间件可用
- 数据库有 users 表和种子数据

## 文件清单

```
internal/
├── handler/auth/
│   ├── handler.go             # Handler struct + NewHandler()
│   ├── login.go               # Login 登录
│   ├── logout.go              # Logout 登出
│   ├── refresh.go             # RefreshToken 刷新
│   ├── profile.go             # GetProfile 获取当前用户信息
│   └── dto.go                 # 该域 DTO
├── service/auth/
│   ├── service.go             # Service struct + NewService()
│   ├── login.go               # Login 业务逻辑
│   ├── logout.go              # Logout
│   ├── refresh.go             # Refresh
│   └── profile.go             # GetProfile
├── repository/
│   ├── user_repo.go           # 补充：GetByEmail
│   └── tenant_repo.go         # 新增：GetByID, GetByCode

internal/router/
├── router.go                  # 更新：注册 auth 路由
└── handlers.go                # 更新：添加 Auth handler

scripts/seed/
└── main.go                    # 种子数据（超管用户 + 默认租户）
```

## 实现规范

### 1. 登录流程

```
前端 → POST /api/v1/auth/login
       Body: { "email": "...", "password": "..." }

后端流程：
1. 验证参数（email + password 必填）
2. 查用户（不带 tenant_id，email 全局唯一）
3. 验证密码（bcrypt）
4. 检查用户状态（status=1 才允许登录）
5. 检查租户状态（tenant.status=1 才允许登录）
6. 查询用户角色（user_roles → roles）
7. 签发 TokenPair
8. 记录登录日志（异步）
9. 返回 TokenPair + 用户基本信息
```

### 2. DTO 定义

```go
// internal/handler/auth/dto.go
package auth

// LoginRequest 登录请求
type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

// LoginResponse 登录响应
type LoginResponse struct {
    AccessToken  string   `json:"access_token"`
    RefreshToken string   `json:"refresh_token"`
    ExpiresAt    int64    `json:"expires_at"`
    User         UserInfo `json:"user"`
}

type UserInfo struct {
    UserID   string `json:"user_id"`
    UserName string `json:"user_name"`
    Email    string `json:"email"`
    Avatar   string `json:"avatar"`
    TenantID string `json:"tenant_id"`
    Roles    []string `json:"roles"`
}

// RefreshRequest 刷新 Token
type RefreshRequest struct {
    RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshResponse 刷新响应
type RefreshResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresAt    int64  `json:"expires_at"`
}
```

### 3. Handler 模板

```go
// internal/handler/auth/handler.go
package auth

import (
    authsvc "admin/internal/service/auth"
)

type Handler struct {
    svc *authsvc.Service
}

func NewHandler(svc *authsvc.Service) *Handler {
    return &Handler{svc: svc}
}
```

```go
// internal/handler/auth/login.go
package auth

import (
    "admin/pkg/response"
    "github.com/gin-gonic/gin"
)

func (h *Handler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Fail(c, xerr.New(xerr.CodeParamInvalid, "参数错误"))
        return
    }

    result, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
    if err != nil {
        response.Fail(c, err)
        return
    }

    response.OK(c, result)
}
```

### 4. Service 模板

```go
// internal/service/auth/service.go
package auth

import (
    "admin/internal/repository"
    "admin/pkg/jwt"
    "github.com/rs/zerolog"
)

type Service struct {
    userRepo   *repository.UserRepo
    tenantRepo *repository.TenantRepo
    jwtMgr     *jwt.Manager
    log        zerolog.Logger
}

func NewService(
    userRepo *repository.UserRepo,
    tenantRepo *repository.TenantRepo,
    jwtMgr *jwt.Manager,
    log zerolog.Logger,
) *Service {
    return &Service{
        userRepo:   userRepo,
        tenantRepo: tenantRepo,
        jwtMgr:     jwtMgr,
        log:        log,
    }
}
```

```go
// internal/service/auth/login.go
package auth

import (
    "context"
    "admin/pkg/idgen"
    "admin/pkg/jwt"
    "admin/pkg/password"
    "admin/pkg/xerr"
)

func (s *Service) Login(ctx context.Context, email, pwd string) (*LoginResult, error) {
    // 1. 查用户（跨租户，用 email 查）
    user, err := s.userRepo.GetByEmail(ctx, email)
    if err != nil {
        return nil, xerr.New(xerr.CodeUnauthorized, "用户名或密码错误")
    }

    // 2. 验证密码
    if !password.Verify(user.Password, pwd) {
        return nil, xerr.New(xerr.CodeUnauthorized, "用户名或密码错误")
    }

    // 3. 检查用户状态
    if user.Status != 1 {
        return nil, xerr.New(xerr.CodeForbidden, "用户已被禁用")
    }

    // 4. 检查租户状态
    tenant, err := s.tenantRepo.GetByID(ctx, user.TenantID)
    if err != nil || tenant.Status != 1 {
        return nil, xerr.New(xerr.CodeTenantDisabled, "租户已被禁用")
    }

    // 5. 查询用户角色
    roles, roleIDs, err := s.userRepo.GetUserRoles(ctx, user.UserID, user.TenantID)
    if err != nil {
        return nil, xerr.Wrap(xerr.CodeInternal, "查询角色失败", err)
    }

    // 6. 签发 Token
    tokenID := idgen.NextID()
    claims := &jwt.Claims{
        TenantID:   user.TenantID,
        TenantCode: tenant.TenantCode,
        UserID:     user.UserID,
        UserName:   user.UserName,
        Roles:      roles,
        RoleIDs:    roleIDs,
        TokenID:    tokenID,
    }
    tokenPair, err := s.jwtMgr.GenerateTokenPair(claims)
    if err != nil {
        return nil, xerr.Wrap(xerr.CodeInternal, "签发Token失败", err)
    }

    // 7. 返回
    return &LoginResult{
        TokenPair: tokenPair,
        User: UserInfo{
            UserID:   user.UserID,
            UserName: user.UserName,
            Email:    user.Email,
            Avatar:   user.Avatar,
            TenantID: user.TenantID,
            Roles:    roles,
        },
    }, nil
}

type LoginResult struct {
    TokenPair *jwt.TokenPair
    User      UserInfo
}
```

### 5. 路由注册

```go
// router.go 中追加
public.POST("/auth/login", handlers.Auth.Login)
public.POST("/auth/refresh", handlers.Auth.Refresh)

// 需要认证的路由
auth := engine.Group("/api/v1")
auth.Use(middleware.JWTAuth(jwtMgr))
{
    auth.POST("/auth/logout", handlers.Auth.Logout)
    auth.GET("/auth/profile", handlers.Auth.GetProfile)
}
```

### 6. 种子数据

```go
// scripts/seed/main.go
// 创建：
// 1. 默认租户：tenant_code=default, name=默认租户
// 2. 超管用户：email=admin@example.com, password=Admin123456
// 3. 超管角色：role_code=super_admin
// 4. user_roles 关联
```

### 7. 安全考虑

| 风险 | 对策 |
|------|------|
| 密码暴力破解 | 登录失败 5 次锁定 15 分钟（Redis 计数） |
| Token 泄露 | 短 TTL (30min) + 黑名单吊销 |
| 错误信息泄露 | 统一返回"用户名或密码错误"，不区分用户不存在/密码错 |
| 并发登录 | 允许多端登录，每次签发独立 TokenID |

## 验收标准

```bash
# 1. 种子数据
make seed
# 期望：创建默认租户 + 超管用户 + 角色

# 2. 登录成功
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"Admin123456"}'
# 期望：返回 access_token + refresh_token + user 信息

# 3. 带 Token 访问
TOKEN=$(curl -s ... | jq -r '.data.access_token')
curl -s http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer $TOKEN"
# 期望：返回当前用户信息

# 4. 无 Token 访问
curl -s http://localhost:8080/api/v1/auth/profile
# 期望：{"code":40100,"message":"缺少认证信息"}

# 5. 登出
curl -s -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer $TOKEN"
# 期望：{"code":0,"message":"success"}

# 6. 登出后 Token 失效
curl -s http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer $TOKEN"
# 期望：{"code":40102,"message":"Token 已失效"}

# 7. 刷新 Token
curl -s -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"..."}'
# 期望：返回新的 access_token + refresh_token
```

## AI 协作提示

```
请按 step-06-auth-login.md 实现登录认证模块。

要点：
1. handler/auth/ — handler.go + login.go + logout.go + refresh.go + profile.go + dto.go
2. service/auth/ — service.go + login.go + logout.go + refresh.go + profile.go
3. repository/ — user_repo.go 补充 GetByEmail/GetUserRoles, tenant_repo.go
4. 登录不带 tenant_id 条件（email 全局唯一）
5. 错误信息不区分用户不存在/密码错误
6. Token 签发后返回 access_token + refresh_token + user info
7. 登出将 TokenID 加入 Redis 黑名单
8. scripts/seed/ 创建默认租户 + 超管用户
9. 更新 router.go 注册路由（public + auth 两组）
10. 更新 main.go 组装依赖链
```

---

*上一步：[Step 05 - 多租户与 JWT](step-05-multi-tenant-jwt.md) | 下一步：[Step 07 - RBAC 核心](step-07-rbac-permission-cache.md)*
