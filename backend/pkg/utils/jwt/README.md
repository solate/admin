# JWT 与 Gin 框架集成指南

## 概述

这个 JWT 包提供了与 Gin 框架高度集成的 token 认证方案，支持以下特性：

- ✅ 双 token 模式（access + refresh）
- ✅ Token 黑名单管理
- ✅ 跨设备登出
- ✅ 会话管理
- ✅ 灵活的中间件
- ✅ 开箱即用的 Gin 适配

## 快速开始

### 1. 基础配置

```go
package main

import (
	"admin/pkg/jwt"
	"admin/pkg/xredis"
	"github.com/gin-gonic/gin"
)

func main() {
	// 步骤 1: 创建 JWT 配置
	jwtConfig, err := jwt.NewConfigBuilder().
		WithAccessSecret("your-access-secret-key-at-least-32-chars").
		WithAccessExpire(3600).                    // 1 小时
		WithRefreshSecret("your-refresh-secret-key").
		WithRefreshExpire(604800).                 // 7 天
		WithIssuer("my-admin-api").
		Build()
	if err != nil {
		panic(err)
	}

	// 步骤 2: 初始化 Redis 存储
	redisClient := xredis.NewRedisClient(cfg)
	store := jwt.NewRedisStore(redisClient)

	// 步骤 3: 创建 JWT 管理器
	jwtManager := jwt.NewJWTManager(jwtConfig, store)

	// 步骤 4: 初始化 Gin 路由
	router := gin.Default()

	// 步骤 5: 设置认证路由（包含 login、refresh、logout）
	jwt.SetupAuthRoutesWithManager(router, jwtManager)

	// 步骤 6: 设置受保护的路由
	protected := router.Group("/api")
	protected.Use(jwt.GinAuthMiddleware(jwtManager))
	{
		protected.GET("/profile", func(c *gin.Context) {
			userID := jwt.GetUserID(c)
			c.JSON(200, gin.H{"user_id": userID})
		})
	}

	router.Run(":8080")
}
```

## 工作流程

### 登录流程
```
1. POST /auth/login
   ↓
   生成 access token 和 refresh token
   ↓
   返回 token pair
   ↓
   客户端保存 token
```

### 请求认证流程
```
1. 客户端在 Authorization header 中附加 token: "Bearer <access_token>"
   ↓
2. GinAuthMiddleware 拦截请求
   ↓
3. 验证 token 签名和过期时间
   ↓
4. 检查 token 是否在黑名单中
   ↓
5. 验证通过，注入 claims 到 gin context
```

### Token 刷新流程
```
1. POST /auth/refresh
   Body: { "refresh_token": "<refresh_token>" }
   ↓
2. 验证 refresh token
   ↓
3. 撤销旧 token（加入黑名单）
   ↓
4. 生成新的 token 对
   ↓
5. 返回新 token
```

### 登出流程
```
1. POST /auth/logout (需要认证)
   ↓
2. 获取当前 token 的 tokenID
   ↓
3. 将 tokenID 加入黑名单
   ↓
4. 删除对应的 refresh token
```

## API 端点

### 公开端点

#### 登录
```bash
POST /auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password123"
}

# 响应
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "expires_in": 3600,
  "token_type": "Bearer"
}
```

#### 刷新 Token
```bash
POST /auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGc..."
}

# 响应
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "expires_in": 3600,
  "token_type": "Bearer"
}
```

#### 验证 Token
```bash
POST /auth/verify
Content-Type: application/json

{
  "token": "eyJhbGc..."
}

# 响应
{
  "valid": true,
  "user_id": "user-123",
  "tenant_id": "tenant-001",
  "role_id": "role-admin"
}
```

### 受保护端点

#### 获取用户信息
```bash
GET /auth/profile
Authorization: Bearer eyJhbGc...

# 响应
{
  "user_id": "user-123",
  "tenant_id": "tenant-001",
  "role_id": "role-admin",
  "issued_at": "2024-01-01T00:00:00Z",
  "expires_at": "2024-01-01T01:00:00Z"
}
```

#### 登出
```bash
POST /auth/logout
Authorization: Bearer eyJhbGc...

# 响应
{
  "message": "logged out successfully"
}
```

#### 跨设备登出（登出所有设备）
```bash
POST /auth/logout-all
Authorization: Bearer eyJhbGc...

# 响应
{
  "message": "all sessions logged out successfully"
}
```

## 中间件使用

### 强制认证中间件
```go
// 所有经过此中间件的路由都需要提供有效的 token
protected := router.Group("/api")
protected.Use(jwt.GinAuthMiddleware(jwtManager))
{
    protected.GET("/protected-resource", handler)
}
```

### 可选认证中间件
```go
// 如果提供了有效的 token，则验证并注入 claims
// 如果没有提供 token 或 token 无效，仍然继续处理请求
optional := router.Group("/api")
optional.Use(jwt.GinOptionalAuthMiddleware(jwtManager))
{
    optional.GET("/maybe-protected", handler)
}
```

## 在处理函数中获取用户信息

```go
func myHandler(c *gin.Context) {
    // 方法 1: 获取整个 claims 对象
    claims := jwt.GetClaims(c)
    if claims == nil {
        c.JSON(401, gin.H{"error": "unauthorized"})
        return
    }
    userID := claims.UserID
    tenantID := claims.TenantID
    roleID := claims.RoleID

    // 方法 2: 快捷获取特定字段
    userID := jwt.GetUserID(c)
    tenantID := jwt.GetTenantID(c)
    roleID := jwt.GetRoleID(c)

    // 业务逻辑...
    c.JSON(200, gin.H{"user_id": userID})
}
```

## 错误处理

```go
import "errors"

func handleToken(c *gin.Context) {
    claims, err := jwtManager.VerifyAccessToken(c.Request.Context(), tokenString)
    if err != nil {
        if errors.Is(err, jwt.ErrTokenExpired) {
            c.JSON(401, gin.H{"error": "token expired"})
        } else if errors.Is(err, jwt.ErrTokenBlacklisted) {
            c.JSON(401, gin.H{"error": "token revoked"})
        } else if errors.Is(err, jwt.ErrMissingToken) {
            c.JSON(401, gin.H{"error": "missing token"})
        } else {
            c.JSON(401, gin.H{"error": "invalid token"})
        }
        return
    }
    // 验证通过
}
```

## Claims 结构

```go
type Claims struct {
    TenantID string                  // 租户 ID
    UserID   string                  // 用户 ID
    RoleID   string                  // 角色 ID
    TokenID  string                  // Token 唯一标识（用于黑名单/会话管理）
    jwt.RegisteredClaims             // 标准 JWT claims
}
```

## 高级特性

### 跨设备登出
```go
// 撤销某个用户的所有会话
jwtManager.RevokeAllUserTokens(ctx, tenantID, userID)
```

### 单个 Token 撤销
```go
// 撤销单个 token（例如在登出时）
jwtManager.RevokeToken(ctx, tokenID)
```

### 配置构造器
```go
// 使用流式 API 构造配置
config, err := jwt.NewConfigBuilder().
    WithAccessSecret("secret1").
    WithAccessExpire(3600).
    WithRefreshSecret("secret2").
    WithRefreshExpire(604800).
    WithIssuer("my-app").
    Build()
```

## 安全建议

1. **密钥管理**
   - 使用强密钥（至少 32 个字符）
   - 从环境变量读取密钥，不要硬编码
   - 定期轮换密钥

2. **Token 过期时间**
   - Access token: 15 分钟 ~ 1 小时
   - Refresh token: 7 ~ 30 天

3. **HTTPS**
   - 在生产环境中必须使用 HTTPS

4. **Token 存储**
   - 客户端不要在 localStorage 中存储 token
   - 使用 httpOnly cookie 存储 token

5. **CORS 配置**
   - 谨慎配置 CORS，避免跨域攻击

## 性能优化

- Redis 存储确保快速查询
- Token 黑名单使用 TTL 自动清理
- 会话索引便于跨设备操作

## 故障排除

### Token 验证失败
- 检查 Authorization header 格式：`Bearer <token>`
- 验证 token 是否过期
- 检查密钥是否与签名时使用的密钥一致

### Redis 连接错误
- 确保 Redis 服务正常运行
- 检查 Redis 连接配置

### Token 刷新失败
- 检查 refresh token 是否过期
- 确保 refresh token 仍在 Redis 中
