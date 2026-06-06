# Step 05: 多租户与 JWT — xcontext / Token / Auth 中间件

## 目标

建立多租户上下文传播机制（JWT Claims → xcontext → Repository），实现 JWT 签发/验证/吊销，Auth 中间件提取认证信息。

## 前置条件

- Step 04 完成，数据模型已生成
- Redis 连接正常（Token 黑名单用）

## 文件清单

```
pkg/
├── xcontext/
│   ├── keys.go                # context key 定义
│   ├── tenant.go              # TenantID / TenantCode
│   ├── user.go                # UserID / UserName / TokenID
│   ├── role.go                # Roles / RoleIDs / HasRole
│   └── copy.go                # CopyContext（异步场景）
├── jwt/
│   ├── config.go              # jwt.Config
│   ├── claims.go              # Claims 结构体
│   └── manager.go             # Manager：签发/验证/吊销
├── idgen/
│   └── idgen.go               # 雪花算法 ID 生成
└── password/
    └── password.go            # bcrypt Hash / Verify

internal/middleware/
└── auth.go                    # JWT 认证中间件
```

## 实现规范

### 1. xcontext — 多租户认证上下文

```go
// pkg/xcontext/keys.go
package xcontext

type contextKey string

const (
    keyTenantID   contextKey = "tenant_id"
    keyTenantCode contextKey = "tenant_code"
    keyUserID     contextKey = "user_id"
    keyUserName   contextKey = "user_name"
    keyTokenID    contextKey = "token_id"
    keyRoles      contextKey = "roles"
    keyRoleIDs    contextKey = "role_ids"
)
```

```go
// pkg/xcontext/tenant.go
package xcontext

import "context"

func SetTenantID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, keyTenantID, id)
}

func GetTenantID(ctx context.Context) string {
    v, _ := ctx.Value(keyTenantID).(string)
    return v
}

func SetTenantCode(ctx context.Context, code string) context.Context {
    return context.WithValue(ctx, keyTenantCode, code)
}

func GetTenantCode(ctx context.Context) string {
    v, _ := ctx.Value(keyTenantCode).(string)
    return v
}
```

```go
// pkg/xcontext/user.go — 同理 SetUserID/GetUserID, SetUserName/GetUserName, SetTokenID/GetTokenID
```

```go
// pkg/xcontext/role.go
package xcontext

import "context"

func SetRoles(ctx context.Context, roles []string) context.Context {
    return context.WithValue(ctx, keyRoles, roles)
}

func GetRoles(ctx context.Context) []string {
    v, _ := ctx.Value(keyRoles).([]string)
    return v
}

func HasRole(ctx context.Context, roleCode string) bool {
    for _, r := range GetRoles(ctx) {
        if r == roleCode {
            return true
        }
    }
    return false
}

func SetRoleIDs(ctx context.Context, ids []string) context.Context {
    return context.WithValue(ctx, keyRoleIDs, ids)
}

func GetRoleIDs(ctx context.Context) []string {
    v, _ := ctx.Value(keyRoleIDs).([]string)
    return v
}
```

```go
// pkg/xcontext/copy.go
package xcontext

import "context"

// CopyContext 拷贝认证上下文到新 context（goroutine 场景必用）
// 新 context 不继承原 context 的取消信号
func CopyContext(ctx context.Context) context.Context {
    newCtx := context.Background()
    newCtx = SetTenantID(newCtx, GetTenantID(ctx))
    newCtx = SetTenantCode(newCtx, GetTenantCode(ctx))
    newCtx = SetUserID(newCtx, GetUserID(ctx))
    newCtx = SetUserName(newCtx, GetUserName(ctx))
    newCtx = SetTokenID(newCtx, GetTokenID(ctx))
    newCtx = SetRoles(newCtx, GetRoles(ctx))
    newCtx = SetRoleIDs(newCtx, GetRoleIDs(ctx))
    return newCtx
}
```

### 2. JWT Manager

```go
// pkg/jwt/config.go
package jwt

import "time"

type Config struct {
    AccessSecret  string
    RefreshSecret string
    AccessTTL     time.Duration
    RefreshTTL    time.Duration
    Issuer        string
}
```

```go
// pkg/jwt/claims.go
package jwt

import jwtgo "github.com/golang-jwt/jwt/v5"

type Claims struct {
    TenantID   string   `json:"tid"`
    TenantCode string   `json:"tcode"`
    UserID     string   `json:"uid"`
    UserName   string   `json:"uname"`
    Roles      []string `json:"roles"`
    RoleIDs    []string `json:"rids"`
    TokenID    string   `json:"jti"`
    jwtgo.RegisteredClaims
}

type TokenPair struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresAt    int64  `json:"expires_at"` // access token 过期时间（毫秒）
}
```

**Claims 字段用缩写**：减少 Token 体积（在每次请求中传输）

```go
// pkg/jwt/manager.go
package jwt

import (
    "context"
    "fmt"
    "time"

    jwtgo "github.com/golang-jwt/jwt/v5"
    "github.com/redis/go-redis/v9"
)

type Manager struct {
    cfg Config
    rdb *redis.Client
}

func NewManager(cfg Config, rdb *redis.Client) *Manager {
    return &Manager{cfg: cfg, rdb: rdb}
}

// GenerateTokenPair 签发 access + refresh token
func (m *Manager) GenerateTokenPair(claims *Claims) (*TokenPair, error) {
    now := time.Now()

    // Access Token
    claims.RegisteredClaims = jwtgo.RegisteredClaims{
        Issuer:    m.cfg.Issuer,
        IssuedAt:  jwtgo.NewNumericDate(now),
        ExpiresAt: jwtgo.NewNumericDate(now.Add(m.cfg.AccessTTL)),
    }
    accessToken, err := jwtgo.NewWithClaims(jwtgo.SigningMethodHS256, claims).
        SignedString([]byte(m.cfg.AccessSecret))
    if err != nil {
        return nil, fmt.Errorf("sign access token: %w", err)
    }

    // Refresh Token（只含 TokenID + UserID，最小化）
    refreshClaims := &jwtgo.RegisteredClaims{
        Subject:   claims.UserID,
        ID:        claims.TokenID,
        Issuer:    m.cfg.Issuer,
        IssuedAt:  jwtgo.NewNumericDate(now),
        ExpiresAt: jwtgo.NewNumericDate(now.Add(m.cfg.RefreshTTL)),
    }
    refreshToken, err := jwtgo.NewWithClaims(jwtgo.SigningMethodHS256, refreshClaims).
        SignedString([]byte(m.cfg.RefreshSecret))
    if err != nil {
        return nil, fmt.Errorf("sign refresh token: %w", err)
    }

    return &TokenPair{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresAt:    now.Add(m.cfg.AccessTTL).UnixMilli(),
    }, nil
}

// VerifyAccess 验证 access token
func (m *Manager) VerifyAccess(tokenStr string) (*Claims, error) {
    claims := &Claims{}
    token, err := jwtgo.ParseWithClaims(tokenStr, claims, func(t *jwtgo.Token) (interface{}, error) {
        return []byte(m.cfg.AccessSecret), nil
    })
    if err != nil {
        return nil, err
    }
    if !token.Valid {
        return nil, fmt.Errorf("invalid token")
    }
    return claims, nil
}

// VerifyRefresh 验证 refresh token，返回 UserID 和 TokenID
func (m *Manager) VerifyRefresh(tokenStr string) (userID, tokenID string, err error) {
    claims := &jwtgo.RegisteredClaims{}
    token, err := jwtgo.ParseWithClaims(tokenStr, claims, func(t *jwtgo.Token) (interface{}, error) {
        return []byte(m.cfg.RefreshSecret), nil
    })
    if err != nil {
        return "", "", err
    }
    if !token.Valid {
        return "", "", fmt.Errorf("invalid refresh token")
    }
    return claims.Subject, claims.ID, nil
}

// Revoke 将 TokenID 加入黑名单（登出时调用）
func (m *Manager) Revoke(ctx context.Context, tokenID string) error {
    key := "token:blacklist:" + tokenID
    return m.rdb.Set(ctx, key, "1", m.cfg.RefreshTTL).Err()
}

// IsRevoked 检查 TokenID 是否已吊销
func (m *Manager) IsRevoked(ctx context.Context, tokenID string) (bool, error) {
    key := "token:blacklist:" + tokenID
    n, err := m.rdb.Exists(ctx, key).Result()
    if err != nil {
        return false, err
    }
    return n > 0, nil
}
```

### 3. Auth 中间件

```go
// internal/middleware/auth.go
package middleware

import (
    "strings"

    "admin/pkg/jwt"
    "admin/pkg/response"
    "admin/pkg/xcontext"
    "admin/pkg/xerr"
    "github.com/gin-gonic/gin"
)

func JWTAuth(jwtMgr *jwt.Manager) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 提取 Bearer Token
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            response.Fail(c, xerr.New(xerr.CodeUnauthorized, "缺少认证信息"))
            c.Abort()
            return
        }
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            response.Fail(c, xerr.New(xerr.CodeUnauthorized, "认证格式错误"))
            c.Abort()
            return
        }

        // 2. 验证 Token
        claims, err := jwtMgr.VerifyAccess(parts[1])
        if err != nil {
            response.Fail(c, xerr.New(xerr.CodeTokenExpired, "Token 无效或已过期"))
            c.Abort()
            return
        }

        // 3. 检查黑名单
        revoked, err := jwtMgr.IsRevoked(c.Request.Context(), claims.TokenID)
        if err != nil || revoked {
            response.Fail(c, xerr.New(xerr.CodeTokenRevoked, "Token 已失效"))
            c.Abort()
            return
        }

        // 4. 注入上下文
        ctx := c.Request.Context()
        ctx = xcontext.SetTenantID(ctx, claims.TenantID)
        ctx = xcontext.SetTenantCode(ctx, claims.TenantCode)
        ctx = xcontext.SetUserID(ctx, claims.UserID)
        ctx = xcontext.SetUserName(ctx, claims.UserName)
        ctx = xcontext.SetTokenID(ctx, claims.TokenID)
        ctx = xcontext.SetRoles(ctx, claims.Roles)
        ctx = xcontext.SetRoleIDs(ctx, claims.RoleIDs)
        c.Request = c.Request.WithContext(ctx)

        c.Next()
    }
}
```

**关键设计**：
- 用 `c.Request = c.Request.WithContext(ctx)` 而非 `c.Set()`
- 下游通过 `c.Request.Context()` 获取，与标准库 context 一致
- Repository 层用 `xcontext.GetTenantID(ctx)` 自然获取租户 ID

### 4. ID 生成 — 雪花算法

```go
// pkg/idgen/idgen.go
package idgen

import (
    "fmt"
    "sync"
    "github.com/sony/sonyflake"
)

var (
    sf   *sonyflake.Sonyflake
    once sync.Once
)

func init() {
    sf = sonyflake.NewSonyflake(sonyflake.Settings{})
    if sf == nil {
        panic("sonyflake init failed")
    }
}

// NextID 生成唯一 ID（string 类型）
func NextID() string {
    id, err := sf.NextID()
    if err != nil {
        panic(fmt.Sprintf("generate id: %v", err))
    }
    return fmt.Sprintf("%d", id)
}

// NextIDs 批量生成 ID
func NextIDs(n int) []string {
    ids := make([]string, n)
    for i := range ids {
        ids[i] = NextID()
    }
    return ids
}
```

**为什么用 sonyflake 而非 UUID**：
- 更短（最多 18 位 vs UUID 36 位）
- 有序（时间递增，B+ 树插入性能好）
- 全局唯一（机器标识 + 序列号）

### 5. 密码工具

```go
// pkg/password/password.go
package password

import "golang.org/x/crypto/bcrypt"

// Hash 对密码进行 bcrypt 加密
func Hash(plain string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
    return string(bytes), err
}

// Verify 验证密码
func Verify(hashed, plain string) bool {
    return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}
```

## 多租户隔离三层模型

```
┌─────────────────────────────────────────────────────────┐
│  Layer 1: 身份层 (JWT)                                   │
│  登录时将 tenant_id 写入 Token → 不可伪造                 │
└─────────────────────┬───────────────────────────────────┘
                      │ Auth 中间件提取
┌─────────────────────▼───────────────────────────────────┐
│  Layer 2: 传播层 (xcontext)                              │
│  Claims → context.WithValue → 全链路传递                  │
└─────────────────────┬───────────────────────────────────┘
                      │ Repository 使用
┌─────────────────────▼───────────────────────────────────┐
│  Layer 3: 执行层 (Repository)                            │
│  WHERE tenant_id = xcontext.GetTenantID(ctx)             │
│  model.TenantID = xcontext.GetTenantID(ctx)              │
└─────────────────────────────────────────────────────────┘
```

## 依赖引入

```bash
go get github.com/golang-jwt/jwt/v5
go get github.com/sony/sonyflake
go get golang.org/x/crypto
```

## 验收标准

```bash
# 1. 编译通过
go build ./...

# 2. ID 生成测试
go test ./pkg/idgen/ -v -run TestNextID
# 期望：生成的 ID 为纯数字字符串，长度 <= 20

# 3. 密码工具测试
go test ./pkg/password/ -v
# 期望：Hash 后 Verify 通过

# 4. JWT 签发/验证测试
go test ./pkg/jwt/ -v
# 期望：
#   - GenerateTokenPair 返回 access + refresh token
#   - VerifyAccess 正确解析 Claims
#   - 过期 Token 验证失败
#   - Revoke 后 IsRevoked 返回 true

# 5. xcontext 测试
go test ./pkg/xcontext/ -v
# 期望：
#   - Set/Get 正确存取
#   - HasRole 在包含时返回 true
#   - CopyContext 拷贝后新 context 能获取所有字段

# 6. Auth 中间件集成测试
# 启动后：
curl -s http://localhost:8080/api/v1/protected
# 期望：{"code":40100,"message":"缺少认证信息"}
```

## AI 协作提示

```
请按 step-05-multi-tenant-jwt.md 实现多租户基础设施。

要点：
1. pkg/xcontext/ — 5 个文件：keys.go, tenant.go, user.go, role.go, copy.go
2. pkg/jwt/ — Claims 字段用缩写（tid/uid/uname）减少 Token 体积
3. pkg/jwt/manager.go — 签发/验证/吊销，黑名单存 Redis
4. pkg/idgen/ — sonyflake 雪花算法，NextID() 返回 string
5. pkg/password/ — bcrypt Hash/Verify
6. internal/middleware/auth.go — 提取 Bearer → 验证 → 黑名单 → 注入 context
7. Auth 中间件用 c.Request.WithContext() 而非 c.Set()
8. 所有 pkg 包写单元测试
```

---

*上一步：[Step 04 - 数据模型](step-04-database-schema.md) | 下一步：[Step 06 - 登录认证](step-06-auth-login.md)*
