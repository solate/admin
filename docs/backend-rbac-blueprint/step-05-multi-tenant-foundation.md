# Step 05: 多租户基础设施 — xcontext / JWT / 隔离

## 目标

建立多租户上下文传播机制：JWT Claims → xcontext → Repository 隔离。实现 JWT 工具和认证上下文存取。

## 前置条件

- Step 04 完成，数据模型已生成

## 文件清单

```
pkg/
├── jwt/
│   ├── jwt.go                 # Claims 结构 + GenerateTokenPair
│   └── manager.go             # Manager：签发、验证、黑名单（Redis）
├── xcontext/
│   ├── tenant.go              # TenantID, TenantCode, CopyContext
│   ├── user.go                # UserID, UserName, TokenID
│   └── role.go                # Roles, RoleIDs, HasRole
├── idgen/
│   └── idgen.go               # UUID 生成
└── password/
    └── password.go            # bcrypt 加密/验证

internal/middleware/
└── auth.go                    # JWT 认证中间件（提取 Claims → xcontext）
```

## 实现细节

### 1. JWT Claims 结构

```go
type Claims struct {
    TenantID   string   `json:"tenant_id"`
    TenantCode string   `json:"tenant_code"`
    UserID     string   `json:"user_id"`
    UserName   string   `json:"user_name"`
    Roles      []string `json:"roles"`      // role_code 列表
    RoleIDs    []string `json:"role_ids"`   // role_id 列表
    TokenID    string   `json:"token_id"`   // 会话标识
    jwt.RegisteredClaims
}
```

**为什么同时存 Roles (code) 和 RoleIDs**：
- `Roles`：用于 `HasRole("super_admin")` 超管判断，人类可读
- `RoleIDs`：用于 PermissionCache 查询权限，查缓存用 ID

### 2. JWT Manager

```go
type Manager struct {
    accessSecret   string
    refreshSecret  string
    accessTTL      time.Duration
    refreshTTL     time.Duration
    issuer         string
    rdb            *redis.Client
}

// 核心方法：
func (m *Manager) GenerateTokenPair(ctx, tenantID, tenantCode, userID, userName string, roles, roleIDs []string) (*TokenPair, error)
func (m *Manager) VerifyAccessToken(tokenString string) (*Claims, error)
func (m *Manager) VerifyRefreshToken(tokenString string) (*Claims, error)
func (m *Manager) RevokeToken(ctx context.Context, tokenID string) error  // Redis 黑名单
func (m *Manager) IsRevoked(ctx context.Context, tokenID string) (bool, error)
```

**设计决策**：
- Access Token：短期（30 分钟），Refresh Token：长期（7 天）
- 两个 Token 共享同一个 `TokenID`（会话标识）
- 黑名单存 Redis：key=`token:blacklist:{tokenID}`，TTL=Refresh Token 的 TTL
- 登出时将 TokenID 加入黑名单

### 3. xcontext — 多租户认证上下文

```go
// tenant.go
func SetTenantID(ctx context.Context, tenantID string) context.Context
func GetTenantID(ctx context.Context) string
func SetTenantCode(ctx context.Context, code string) context.Context
func GetTenantCode(ctx context.Context) string

// user.go
func SetUserID(ctx context.Context, userID string) context.Context
func GetUserID(ctx context.Context) string
func SetUserName(ctx context.Context, name string) context.Context
func GetUserName(ctx context.Context) string
func SetTokenID(ctx context.Context, tokenID string) context.Context
func GetTokenID(ctx context.Context) string

// role.go
func SetRoles(ctx context.Context, roles []string) context.Context
func GetRoles(ctx context.Context) []string
func HasRole(ctx context.Context, roleCode string) bool
func SetRoleIDs(ctx context.Context, roleIDs []string) context.Context
func GetRoleIDs(ctx context.Context) []string

// CopyContext — 异步场景必须用
func CopyContext(ctx context.Context) context.Context
```

**CopyContext 的必要性**：goroutine 中使用 `go func()` 时，默认的 `context.Background()` 丢失了认证上下文。`CopyContext` 将所有字段拷贝到新 context。

### 4. Auth 中间件

```go
func JWTAuth(jwtMgr *jwt.Manager) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 提取 Bearer token
        // 2. VerifyAccessToken
        // 3. 检查黑名单 IsRevoked
        // 4. 调用 SetAuthContext(c, claims) 填充所有字段
        // 5. c.Next()
    }
}

func SetAuthContext(c *gin.Context, claims *jwt.Claims) {
    ctx := c.Request.Context()
    ctx = xcontext.SetTenantID(ctx, claims.TenantID)
    ctx = xcontext.SetTenantCode(ctx, claims.TenantCode)
    ctx = xcontext.SetUserID(ctx, claims.UserID)
    ctx = xcontext.SetUserName(ctx, claims.UserName)
    ctx = xcontext.SetRoles(ctx, claims.Roles)
    ctx = xcontext.SetRoleIDs(ctx, claims.RoleIDs)
    ctx = xcontext.SetTokenID(ctx, claims.TokenID)
    c.Set("ctx", ctx)
}
```

### 5. 多租户隔离模式

```
三层隔离：

1. JWT Claims（身份层）：
   登录时 tenant_id 写入 Token → 无法伪造

2. xcontext（传播层）：
   Auth 中间件提取 Claims → 存入 Context → 全链路可用

3. Repository（执行层）：
   所有查询 WHERE tenant_id = xcontext.GetTenantID(ctx)
   所有创建 model.TenantID = xcontext.GetTenantID(ctx)
```

## 验收标准

- [ ] `GenerateTokenPair()` 生成 access + refresh token
- [ ] `VerifyAccessToken()` 正确解析 Claims
- [ ] 过期 Token 验证失败
- [ ] `RevokeToken()` 后 `IsRevoked()` 返回 true
- [ ] xcontext 所有 Set/Get 正确存取
- [ ] `HasRole(ctx, "super_admin")` 在 Roles 包含时返回 true
- [ ] `CopyContext()` 拷贝后 goroutine 中能获取 TenantID
- [ ] Auth 中间件在无 Token 时返回 401
- [ ] Auth 中间件在 Token 过期时返回 401
- [ ] Auth 中间件在 Token 有效时正确填充 xcontext

## AI 协作提示

```
请按 step-05-multi-tenant-foundation.md 实现多租户基础设施。
这是整个系统最关键的基础，必须确保：
1. JWT Claims 结构完整（tenantID, userID, roles, roleIDs, tokenID）
2. xcontext 的 CopyContext 拷贝所有字段
3. Auth 中间件正确提取 Claims 并填充 context
4. Token 黑名单用 Redis 实现
参考现有 backend/pkg/utils/jwt/ 和 backend/pkg/xcontext/ 的模式。
```
