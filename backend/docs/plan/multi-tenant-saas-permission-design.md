# 多租户 SaaS 平台权限系统设计方案

> 从行业最佳实践角度，重新审视多租户权限架构设计

---

## 一、问题背景

### 1.1 业务场景

在多租户 SaaS 平台中，存在三种典型的数据访问需求：

| 用户类型 | 数据访问范围 | 典型场景 |
|---------|-------------|---------|
| **超级管理员** | 所有租户的数据 | 平台运营、技术支持、跨租户数据分析 |
| **监管方/审计员** | 特定租户子集 | 合规审计、集团管控多子公司 |
| **租户管理员** | 本租户内数据 | 租户内部用户管理、权限分配 |
| **普通用户** | 个人数据 | 个人信息查看、个人操作记录 |

### 1.2 核心挑战

1. **租户隔离与跨租户访问的矛盾**：默认需要严格隔离，但特定场景需要跨租户
2. **权限传递的复杂性**：租户上下文如何在整个调用链中正确传递
3. **接口设计的简洁性**：如何不通过额外接口/参数实现复杂权限控制
4. **默认配置继承**：租户管理员创建角色/菜单时，如何看到并使用超管设置的默认值

### 1.3 核心设计思路

**关键发现**：系统已经存在 `tenant_code = 'default'` 的默认租户，我们可以利用它来存储全局模板数据。

```
查询逻辑 = (default 租户的数据) ∪ (当前租户的数据)
```

这样：
- ✅ 不需要额外的 `global_*_templates` 表
- ✅ 数据结构统一，所有数据都在同一张表
- ✅ 查询逻辑简单：`WHERE tenant_id IN (default_tenant_id, current_tenant_id)`
- ✅ 租户可以覆盖默认值（创建同名资源，查询时优先返回租户的）

---

## 二、业界解决方案对比

### 2.1 主流方案对比

| 方案 | 代表产品 | 核心思路 | 优点 | 缺点 |
|------|---------|---------|------|------|
| **行级安全 (RLS)** | PostgreSQL, SQL Server | 数据库层自动添加 WHERE 条件 | 安全、透明、无遗漏 | 数据库绑定、灵活性差 |
| **应用层租户上下文** | Salesforce, Shopify | Context 传递租户ID，ORM/业务层过滤 | 灵活、跨数据库友好 | 需要开发者意识 |
| **双层权限模型** | AWS IAM | 身份认证 + 资源策略分离 | 表达力强、细粒度 | 复杂度高 |
| **Casbin RBAC** | Kubernetes RBAC | 策略引擎 + 中间件拦截 | 灵活、无侵入 | 需要额外配置 |

### 2.2 推荐方案：应用层租户上下文 + Casbin RBAC

结合我们现有的技术栈，采用 **GORM Scope + Casbin** 的组合方案：

```
┌─────────────────────────────────────────────────────────────┐
│                    多租户权限控制层次                         │
├─────────────────────────────────────────────────────────────┤
│  Layer 1: JWT 认证层                                        │
│  └─ Token 包含：userId, tenantId, roles                     │
├─────────────────────────────────────────────────────────────┤
│  Layer 2: RoleMiddleware (角色中间件)                       │
│  └─ 根据角色处理租户切换、跳过数据层检查                      │
├─────────────────────────────────────────────────────────────┤
│  Layer 3: Casbin 权限层 (API 级别)                          │
│  └─ 策略：g(user, role, tenant), p(role, tenant, res, act)  │
│  └─ 超管跳过，其他角色走策略检查                             │
├─────────────────────────────────────────────────────────────┤
│  Layer 4: GORM Scope 层 (数据级别)                         │
│  └─ 自动添加 WHERE tenant_id = ? 或跳过 (skip_tenant_check)  │
└─────────────────────────────────────────────────────────────┘
```

---

## 三、JWT Token 设计

### 3.1 Token结构

```go
type Claims struct {
    // 基础身份信息
    TenantID   string   `json:"tenant_id"`    // 当前登录租户ID
    TenantCode string   `json:"tenant_code"`  // 当前登录租户编码
    UserID     string   `json:"user_id"`
    UserName   string   `json:"user_name"`

    // 权限信息
    Roles      []string `json:"roles"`         // 当前租户下的角色列表

    // 会话管理
    TokenID    string   `json:"token_id"`      // 会话唯一标识
    jwt.RegisteredClaims
}
```

**设计要点**：
- ✅ 不需要 `scope` 字段
- ✅ 直接通过 `roles` 判断用户权限
- ✅ 超管判断：`contains(roles, "super_admin")`

### 3.2 登录时 Token 生成

登录流程：`POST /api/v1/auth/:tenant_code/login`

1. **TenantFromCode 中间件**：从 URL 获取租户编码，查询租户信息并注入到 context
2. **AuthService.Login**：[auth_service.go:48](backend/internal/service/auth_service.go#L48)
   - 验证码校验
   - 从 context 获取租户信息
   - 查询用户（租户内唯一）
   - 验证密码和状态
   - 从 Casbin 获取用户角色
   - 生成 JWT Token

---

## 四、Casbin 权限模型设计

### 4.1 模型配置

```conf
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act

[role_definition]
g = _, _, _

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act)
```

### 4.2 策略示例

```conf
# 用户-角色-租户绑定 (g 策略)
g, admin, super_admin, platform      # 平台超管
g, alice, tenant_admin, tenant_a      # 租户A管理员
g, bob, user, tenant_a                # 租户A普通用户
g, auditor, viewer, tenant_b          # 监管方（可多租户）
g, auditor, viewer, tenant_c

# 角色权限定义 (p 策略)
p, super_admin, platform, /api/v1/*, *
p, tenant_admin, tenant_a, /api/v1/users/*, *
p, tenant_admin, tenant_a, /api/v1/roles/*, *
p, user, tenant_a, /api/v1/profile, GET
p, viewer, tenant_b, /api/v1/audit/*, GET
p, viewer, tenant_c, /api/v1/audit/*, GET
```

### 4.3 角色权限说明

| 特性 | 超级管理员 | 租户管理员 | 普通用户 |
|------|-----------|-----------|---------|
| **Casbin 检查** | 跳过 | 正常检查 | 正常检查 |
| **数据层过滤** | 默认跳过 | 自动过滤 | 自动过滤 |
| **租户切换** | 任意租户（X-Target-Tenant） | 不能切换 | 不能切换 |
| **default 租户读** | ✅ | ✅ | ❌ |
| **default 租户写** | ✅ | ❌ | ❌ |
| **实现方式** | `contains(roles, "super_admin")` | Casbin 策略 | Casbin 策略 |

---

## 五、中间件链设计

### 5.1 完整中间件链

```go
// internal/router/router.go

func Setup(r *gin.Engine, app *App) {
    // 全局中间件
    r.Use(middleware.RequestIDMiddleware())
    r.Use(middleware.LoggerMiddleware())
    r.Use(middleware.RecoveryMiddleware())
    r.Use(middleware.CORSMiddleware())
    r.Use(middleware.RateLimitMiddleware(...))

    // 公开接口（无需认证）
    public := r.Group("/api/v1")
    {
        // 登录需要从 URL 获取租户编码
        auth := public.Group("/auth/:tenant_code")
        auth.Use(middleware.TenantFromCode(app.DB))
        {
            auth.POST("/login", app.Handlers.AuthHandler.Login)
        }

        // 其他公开接口
        public.GET("/auth/captcha", app.Handlers.CaptchaHandler.Get)
        public.POST("/auth/refresh", app.Handlers.AuthHandler.Refresh)
    }

    // 认证接口（需 JWT + Casbin 权限检查）
    authorized := r.Group("/api/v1")
    authorized.Use(middleware.AuthMiddleware(app.JWT))
    authorized.Use(middleware.CasbinMiddleware(app.Enforcer))
    authorized.Use(middleware.OperationLogMiddleware(app.OperationLogWriter))
    {
        // 业务接口
        authorized.GET("/profile", app.Handlers.UserHandler.GetProfile)

        // 租户切换（超管/审计员）
        authorized.POST("/auth/switch-tenant", app.Handlers.AuthHandler.SwitchTenant)
        authorized.GET("/auth/available-tenants", app.Handlers.AuthHandler.GetAvailableTenants)

        // 租户管理
        authorized.POST("/tenants", app.Handlers.TenantHandler.CreateTenant)
        // ...
    }
}
```

### 5.2 AuthMiddleware（JWT 认证）

```go
// internal/middleware/auth_middleware.go

func AuthMiddleware(jwtManager *jwt.Manager) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            response.Error(c, xerr.ErrUnauthorized)
            c.Abort()
            return
        }

        // 解析 "Bearer <token>" 格式
        parts := strings.SplitN(token, " ", 2)
        if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
            response.Error(c, xerr.ErrTokenInvalid)
            c.Abort()
            return
        }

        // 验证 token（签名、过期、黑名单）
        claims, err := jwtManager.VerifyAccessToken(c.Request.Context(), parts[1])
        if err != nil {
            response.Error(c, err)
            c.Abort()
            return
        }

        // 将认证信息注入到 context（租户、用户、角色等）
        requestCtx := SetAuthContext(c.Request.Context(), claims)
        c.Request = c.Request.WithContext(requestCtx)

        c.Next()
    }
}

// SetAuthContext 一次性设置所有认证信息到context
func SetAuthContext(ctx context.Context, claims *jwt.Claims) context.Context {
    ctx = xcontext.SetTenantID(ctx, claims.TenantID)
    ctx = xcontext.SetTenantCode(ctx, claims.TenantCode)
    ctx = xcontext.SetUserID(ctx, claims.UserID)
    ctx = xcontext.SetUserName(ctx, claims.UserName)
    ctx = xcontext.SetRoles(ctx, claims.Roles)
    ctx = xcontext.SetTokenID(ctx, claims.TokenID)
    return ctx
}
```

### 5.3 TenantFromCode（登录时从 URL 获取租户）

```go
// internal/middleware/tenant_middleware.go

func TenantFromCode(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 从 URL 路径参数获取租户编码
        tenantCode := c.Param("tenant_code")
        if tenantCode == "" {
            response.Error(c, xerr.New(xerr.ErrInvalidParams.Code, "租户编码不能为空"))
            c.Abort()
            return
        }

        // 查询租户信息（使用 Manual 模式，不自动添加租户过滤）
        tenant, err := repository.NewTenantRepo(db).GetByCodeManual(c.Request.Context(), tenantCode)
        if err != nil {
            response.Error(c, xerr.New(xerr.ErrNotFound.Code, "租户不存在"))
            c.Abort()
            return
        }

        // 将租户信息注入到 context
        ctx := c.Request.Context()
        ctx = database.WithTenantID(ctx, tenant.TenantID)
        ctx = xcontext.SetTenantID(ctx, tenant.TenantID)
        ctx = xcontext.SetTenantCode(ctx, tenant.TenantCode)
        c.Request = c.Request.WithContext(ctx)

        c.Next()
    }
}
```

### 5.4 CasbinMiddleware（权限检查）

```go
// internal/middleware/casbin_middleware.go

func CasbinMiddleware(enforcer *casbin.Enforcer) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := c.Request.Context()
        roles := xcontext.GetRoles(ctx)

        // 超管跳过 Casbin 检查
        if xcontext.HasRole(ctx, constants.SuperAdmin) {
            c.Next()
            return
        }

        // 其他角色走 Casbin 策略检查
        ok, err := enforcer.Enforce(
            xcontext.GetUserName(ctx),
            xcontext.GetTenantCode(ctx),
            c.Request.URL.Path,
            c.Request.Method,
        )

        if !ok || err != nil {
            response.Error(c, xerr.ErrForbidden)
            c.Abort()
            return
        }

        c.Next()
    }
}
```

---

## 六、GORM Scope 隔离设计

### 6.1 三种查询模式

| 模式 | 设置方式 | 行为 | 使用场景 |
|------|---------|------|---------|
| **自动模式** | 默认 | 自动添加 `WHERE tenant_id = ?` | 业务数据查询 |
| **跳过模式** | `database.SkipTenantCheck(ctx)` | 不添加租户过滤 | 超管全局查询、登录验证 |
| **手动模式** | `database.WithTenantMode(ctx, "manual")` | 不自动添加，由业务控制 | 跨租户聚合查询 |

### 6.2 数据层回调（现有实现）

```go
// pkg/database/scopes.go

func tenantQueryCallback(db *gorm.DB) {
    if !hasTenantColumn(db) {
        return
    }

    // 1. 优先检查是否跳过租户检查
    if shouldSkipTenantCheck(db) {
        return
    }

    // 2. 检查查询模式
    mode := getTenantMode(db)
    if mode == TenantModeManual {
        // 手动模式：不自动添加 WHERE，由 Repository 精确控制
        return
    }

    // 3. 默认行为：自动添加当前租户
    tenantID, ok := getTenantID(db)
    if !ok {
        db.AddError(ErrMissingTenantID)
        return
    }
    db.Statement.AddClause(clause.Where{Exprs: []clause.Expression{
        clause.Eq{Column: clause.Column{Name: "tenant_id"}, Value: tenantID},
    }})
}
```

### 6.3 Repository 层使用示例

```go
// internal/repository/user_repo.go

// 默认情况：自动添加租户过滤
func (r *UserRepo) List(ctx context.Context) ([]*model.User, error) {
    var users []*model.User
    err := r.db.WithContext(ctx).Find(&users).Error
    // 自动生成：SELECT * FROM users WHERE tenant_id = ?
    return users, err
}

// 超管查询所有租户：使用 SkipTenantCheck
func (r *UserRepo) ListAllTenants(ctx context.Context) ([]*model.User, error) {
    var users []*model.User
    ctx = database.SkipTenantCheck(ctx)
    err := r.db.WithContext(ctx).Find(&users).Error
    // 生成：SELECT * FROM users
    return users, err
}

// 监管方查询指定租户列表：使用 Manual 模式
func (r *UserRepo) ListByTenantCodes(ctx context.Context, tenantCodes []string) ([]*model.User, error) {
    var users []*model.User
    ctx = database.WithTenantMode(ctx, database.TenantModeManual)
    err := r.db.WithContext(ctx).
        Where("tenant_code IN ?", tenantCodes).
        Find(&users).Error
    // 生成：SELECT * FROM users WHERE tenant_code IN (?, ?, ?)
    return users, err
}
```

---

## 七、租户上下文传递策略

### 7.1 Context 设计

```go
// pkg/xcontext/tenant.go

type contextKey string

const (
    TenantIDKey   contextKey = "tenant_id"
    TenantCodeKey contextKey = "tenant_code"
    UserIDKey     contextKey = "user_id"
    UserNameKey   contextKey = "user_name"
    RolesKey      contextKey = "roles"
    TokenIDKey    contextKey = "token_id"
)
```

**实现**：[tenant.go:1](backend/pkg/xcontext/tenant.go)

### 7.2 传递策略

```
┌─────────────────────────────────────────────────────────────┐
│  HTTP Request                                               │
│  URL: /api/v1/auth/:tenant_code/login                       │
│  Headers: Authorization                                     │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│  TenantFromCode (仅登录)                                    │
│  ├─ 从 URL 获取租户编码                                      │
│  ├─ 查询租户信息                                             │
│  └─ 注入: TenantID, TenantCode                              │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│  AuthMiddleware                                             │
│  ├─ 解析 JWT → Claims                                       │
│  ├─ SetAuthContext 一次性设置所有信息                        │
│  └─ Context → request.Context                               │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│  CasbinMiddleware                                           │
│  ├─ 检查角色为超管 → 跳过                                    │
│  └─ 否则执行: Enforce(user, tenant, path, method)           │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│  Handler → Service → Repository                             │
│  └─ 通过 xcontext.GetXxx(ctx) 获取租户/用户信息              │
└─────────────────────────────────────────────────────────────┘
```

---

## 八、API 接口设计原则

### 8.1 核心原则

| 原则 | 说明 | 示例 |
|------|------|------|
| **Context 优先** | 租户ID从 Context 获取，不传参数 | `GET /api/v1/users` |
| **URL 获取租户** | 登录时从 URL 路径获取租户 | `POST /api/v1/auth/:tenant_code/login` |
| **API 切换租户** | 超管/审计员通过专用 API 切换租户 | `POST /api/v1/auth/switch-tenant` |
| **业务无关** | Handler/Service 不关心租户逻辑 | 所有角色共用一套接口 |

### 8.2 接口示例

```bash
# ========== 登录（通过 URL 指定租户）==========
POST /api/v1/auth/tenant_a/login
Body: { "username": "alice", "password": "xxx", "captcha_id": "xxx", "captcha": "xxx" }

# ========== 普通用户/租户管理员 ==========
# 获取当前租户用户列表（自动使用 JWT 租户）
GET /api/v1/users
Headers: Authorization: Bearer <token>

# ========== 超级管理员 ==========
# 查看平台所有租户
GET /api/v1/tenants
Headers: Authorization: Bearer <super_admin_token>

# 获取可切换的租户列表（超管返回所有租户）
GET /api/v1/auth/available-tenants
Headers: Authorization: Bearer <super_admin_token>

# 切换到租户 A（返回新的 Token）
POST /api/v1/auth/switch-tenant
Headers: Authorization: Bearer <super_admin_token>
Body: { "tenant_id": "tenant_a_id" }
# 返回: { "access_token": "...", "refresh_token": "...", "expires_in": 3600 }

# 切换后，使用新 Token 访问租户 A 的数据
GET /api/v1/users
Headers: Authorization: Bearer <new_token>

# ========== 审计员 ==========
# 获取可切换的租户列表（只返回有权限的租户）
GET /api/v1/auth/available-tenants
Headers: Authorization: Bearer <auditor_token>
# 返回: [{ "tenant_id": "...", "tenant_code": "tenant_b", ... }, ...]

# 切换到有权限的租户
POST /api/v1/auth/switch-tenant
Headers: Authorization: Bearer <auditor_token>
Body: { "tenant_id": "tenant_b_id" }
```

### 8.3 Handler 实现（无租户逻辑）

```go
// internal/handler/user_handler.go

func (h *UserHandler) List(c *gin.Context) {
    // 业务逻辑完全不需要关心租户
    users, err := h.userService.List(c.Request.Context())
    if err != nil {
        response.Error(c, err)
        return
    }
    response.Success(c, users)
}

func (h *UserHandler) Create(c *gin.Context) {
    var req dto.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, xerr.ErrInvalidParams)
        return
    }

    // Service 会自动从 context 获取租户ID
    user, err := h.userService.Create(c.Request.Context(), &req)
    if err != nil {
        response.Error(c, err)
        return
    }
    response.Success(c, user)
}
```

```go
// internal/service/user_service.go

func (s *UserService) Create(ctx context.Context, req *dto.CreateUserRequest) (*model.User, error) {
    // 自动从 context 获取租户ID（由 JWT 设置）
    tenantID := xcontext.GetTenantID(ctx)

    user := &model.User{
        UserID:    uuid.New().String(),
        TenantID:  tenantID,  // 使用上下文中的租户ID
        UserName:  req.UserName,
        Password:  req.Password,
        // ... 其他字段
    }

    return s.userRepo.Create(ctx, user)
}
```

---

## 九、典型场景实现

### 9.1 场景一：用户登录

```bash
# 通过 URL 指定租户
POST /api/v1/auth/tenant_a/login
Body: { "username": "alice", "password": "xxx", "captcha_id": "xxx", "captcha": "xxx" }
```

**流程**：
1. `TenantFromCode` 中间件从 URL 获取租户编码 → 查询租户 → 注入 context
2. `AuthService.Login` [auth_service.go:48](backend/internal/service/auth_service.go#L48) 验证并生成 Token
3. 返回 Token（包含租户信息和角色）

### 9.2 场景二：超管跨租户操作

```bash
# 1. 获取可切换的租户列表
GET /api/v1/auth/available-tenants
Headers: Authorization: Bearer <super_admin_token>

# 2. 切换到租户 A
POST /api/v1/auth/switch-tenant
Headers: Authorization: Bearer <super_admin_token>
Body: { "tenant_id": "tenant_a_id" }

# 3. 使用新 Token 访问租户 A 的数据
GET /api/v1/users
Headers: Authorization: Bearer <new_token>
```

**实现**：[auth_service.go:151](backend/internal/service/auth_service.go#L151) - SwitchTenant

### 9.3 场景三：审计员查看多租户数据

```bash
# 1. 获取有权限的租户列表（从 Casbin g 策略获取）
GET /api/v1/auth/available-tenants
Headers: Authorization: Bearer <auditor_token>

# 2. 切换到有权限的租户
POST /api/v1/auth/switch-tenant
Body: { "tenant_id": "tenant_b_id" }
```

**实现**：[auth_service.go:236](backend/internal/service/auth_service.go#L236) - GetAvailableTenants

---

## 十、安全最佳实践

### 10.1 防御层次

| 层次 | 防护措施 | 检查点 |
|------|---------|-------|
| 1 | JWT 签名验证 | AuthMiddleware |
| 2 | Token 黑名单 | JWTManager |
| 3 | 租户切换权限 | RoleMiddleware |
| 4 | API 级别权限 | CasbinMiddleware |
| 5 | 数据级隔离 | GORM Scope |
| 6 | 业务级校验 | Service 层 |

### 10.2 关键安全规则

1. **租户管理员无法越权**：RoleMiddleware 忽略其 X-Target-Tenant 参数
2. **监管方范围受限**：必须预配置授权租户列表
3. **超管可追溯**：所有跨租户操作记录审计日志
4. **Context 只写**：Handler 不应修改 Context 租户信息

### 10.3 审计日志

```go
// pkg/audit/audit.go

type AuditLog struct {
    ID           string
    ActorID      string        // 操作者ID
    ActorTenant  string        // 操作者租户
    TargetTenant string        // 目标租户（超管跨租户时）
    Action       string        // 操作类型
    Resource     string        // 资源
    IP           string
    UserAgent    string
    CreatedAt    time.Time
}

func LogAction(ctx context.Context, action, resource string) {
    log := &AuditLog{
        ActorID:      xcontext.GetUserID(ctx),
        ActorTenant:  xcontext.GetTenantCode(ctx), // JWT 中的原始租户
        TargetTenant: xcontext.GetTenantCode(ctx), // Context 中的目标租户
        Action:       action,
        Resource:     resource,
        // ...
    }
    // 存储
}
```

---

## 十一、性能优化

### 11.1 Casbin 策略缓存

```go
// 使用 Redis 缓存策略
type CachedEnforcer struct {
    *casbin.Enforcer
    cache *redis.Client
}

func (e *CachedEnforcer) Enforce(rvals ...interface{}) (bool, error) {
    key := fmt.Sprintf("casbin:%v", rvals)
    if cached, err := e.cache.Get(key); err == nil {
        return cached == "1", nil
    }

    result, err := e.Enforcer.Enforce(rvals...)
    if err == nil {
        e.cache.Set(key, result, 5*time.Minute)
    }
    return result, err
}
```

### 11.2 角色信息缓存到 JWT

避免每次请求都查询 Casbin：

```go
// 登录时查询一次角色，存入 JWT
roles := enforcer.GetRolesForUserInDomain(username, tenantCode)
token := jwt.WithClaims("roles", roles)

// 中间件直接从 JWT 读取
if len(claims.Roles) > 0 && contains(claims.Roles, "super_admin") {
    // 跳过 Casbin 检查
}
```

---

## 十三、总结

本设计方案通过以下方式实现多租户 SaaS 平台的复杂权限需求：

1. **JWT Roles 字段**：通过角色列表判断用户权限（super_admin/tenant_admin/auditor）
2. **SetAuthContext 统一设置**：AuthMiddleware 一次性注入所有认证信息到 context
3. **Casbin 智能跳过**：超管跳过检查，其他角色走策略
4. **GORM 两种模式**：Auto/Skip 灵活控制数据层过滤（Manual 模式用于登录查询租户）
5. **租户切换 API**：通过 `/auth/switch-tenant` 专用接口实现租户切换，返回新 Token
6. **Context 传递**：租户信息在整个调用链中透明传递

**核心优势**：
- ✅ 单套接口服务所有角色
- ✅ Handler/Service 完全不感知租户逻辑
- ✅ 支持超管、租户管理员、审计员三种角色
- ✅ 安全可控，多层防御
- ✅ 租户切换返回新 Token，无状态设计
- ✅ 登录时通过 URL 指定租户，语义清晰
