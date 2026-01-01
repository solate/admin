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

```go
// internal/service/auth_service.go

func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
    // 1. 验证用户凭据（跨租户查询）
    ctx = database.SkipTenantCheck(ctx)
    user, err := s.userRepo.GetByUserName(ctx, req.UserName)
    if err != nil {
        return nil, err
    }

    // 2. 获取用户在当前租户下的角色
    roles, err := s.casbinService.GetRolesForUserInDomain(ctx, user.UserName, user.TenantCode)
    if err != nil {
        return nil, err
    }

    // 3. 生成 Token
    tokenPair, err := s.jwtManager.GenerateTokenPair(
        user.TenantID,
        user.TenantCode,
        user.UserID,
        user.UserName,
        roles,
    )

    return &dto.LoginResponse{
        AccessToken:  tokenPair.AccessToken,
        RefreshToken: tokenPair.RefreshToken,
        ExpiresIn:    tokenPair.ExpiresIn,
        Roles:        roles,
    }, nil
}
```

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

func SetupRouter(r *gin.Engine, db *gorm.DB, enforcer *casbin.Enforcer, jwtManager *jwt.Manager) {
    api := r.Group("/api/v1")

    // ========== 公开接口（无需认证）==========
    public := api.Group("")
    {
        public.POST("/login", handlers.Login)
        public.POST("/refresh", handlers.RefreshToken)
    }

    // ========== 认证接口（需 JWT 验证）==========
    authenticated := api.Group("")
    authenticated.Use(middleware.AuthMiddleware(jwtManager))
    {
        // ========== 超管专用接口（可选）==========
        admin := authenticated.Group("/admin")
        admin.Use(middleware.SuperAdminOnly())
        {
            admin.GET("/tenants", handlers.ListAllTenants)
            admin.POST("/tenants", handlers.CreateTenant)
        }

        // ========== RoleMiddleware（核心）==========
        // 根据角色处理租户切换和数据层检查
        authenticated.Use(middleware.RoleMiddleware(db))

        // ========== Casbin 权限检查 ==========
        authenticated.Use(middleware.CasbinMiddleware(enforcer))

        // ========== 业务接口 ==========
        authenticated.GET("/roles", handlers.ListRoles)
        authenticated.POST("/roles", handlers.CreateRole)
        // ...
    }
}
```

### 5.2 RoleMiddleware 中间件（核心）

```go
// internal/middleware/role_middleware.go

// RoleMiddleware 根据用户角色设置租户上下文
// - 超管(super_admin)：可切换任意租户，跳过数据层检查
// - 监管方(auditor)：可切换授权租户列表
// - 其他角色：强制使用 JWT 中的租户
func RoleMiddleware(db *gorm.DB) gin.HandlerFunc {
    tenantRepo := repository.NewTenantRepo(db)

    return func(c *gin.Context) {
        ctx := c.Request.Context()
        roles := xcontext.GetRoles(ctx)

        // 默认使用 JWT 中的租户
        finalTenantID := xcontext.GetTenantID(ctx)
        finalTenantCode := xcontext.GetTenantCode(ctx)

        // 获取请求的目标租户
        targetTenant := c.GetHeader("X-Target-Tenant")
        if targetTenant != "" {
            // 超管：可以切换任意租户
            if hasRole(roles, constants.SuperAdmin) {
                tenant, err := tenantRepo.GetByCode(ctx, targetTenant)
                if err != nil {
                    response.Error(c, xerr.ErrTenantNotFound)
                    c.Abort()
                    return
                }
                finalTenantID = tenant.TenantID
                finalTenantCode = tenant.TenantCode
            }
            // 监管方：只能切换授权的租户列表
            else if hasRole(roles, constants.Auditor) {
                allowedTenants := s.getAllowedTenants(ctx, xcontext.GetUserID(ctx))
                if contains(allowedTenants, targetTenant) {
                    tenant, err := tenantRepo.GetByCode(ctx, targetTenant)
                    if err != nil {
                        response.Error(c, xerr.ErrTenantNotFound)
                        c.Abort()
                        return
                    }
                    finalTenantID = tenant.TenantID
                    finalTenantCode = tenant.TenantCode
                } else {
                    response.Error(c, xerr.ErrForbidden)
                    c.Abort()
                    return
                }
            }
            // 其他角色：忽略 X-Target-Tenant，使用 JWT 租户
        }

        // 更新 Context
        ctx = xcontext.SetTenantID(ctx, finalTenantID)
        ctx = xcontext.SetTenantCode(ctx, finalTenantCode)

        // 超管跳过数据层租户检查
        if hasRole(roles, constants.SuperAdmin) {
            ctx = database.SkipTenantCheck(ctx)
        }

        c.Request = c.Request.WithContext(ctx)
        c.Next()
    }
}

// hasRole 检查用户是否拥有指定角色
func hasRole(roles []string, targetRole string) bool {
    for _, role := range roles {
        if role == targetRole {
            return true
        }
    }
    return false
}
```

### 5.3 CasbinMiddleware（简化版）

```go
// internal/middleware/casbin_middleware.go

func CasbinMiddleware(enforcer *casbin.Enforcer) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := c.Request.Context()
        roles := xcontext.GetRoles(ctx)

        // 超管跳过 Casbin 检查
        if hasRole(roles, constants.SuperAdmin) {
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

        if !ok {
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
    TenantIDKey      contextKey = "tenant_id"
    TenantCodeKey    contextKey = "tenant_code"
    UserIDKey        contextKey = "user_id"
    UserNameKey      contextKey = "user_name"
    RolesKey         contextKey = "roles"
    AllowedTenantsKey contextKey = "allowed_tenants" // 授权租户列表（监管方）
)

// SetAllowedTenants 设置授权租户列表
func SetAllowedTenants(ctx context.Context, tenants []string) context.Context {
    return context.WithValue(ctx, AllowedTenantsKey, tenants)
}

// GetAllowedTenants 获取授权租户列表
func GetAllowedTenants(ctx context.Context) []string {
    if tenants, ok := ctx.Value(AllowedTenantsKey).([]string); ok {
        return tenants
    }
    return nil
}
```

### 7.2 传递策略

```
┌─────────────────────────────────────────────────────────────┐
│  HTTP Request                                               │
│  Headers: Authorization + X-Target-Tenant (可选)            │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│  AuthMiddleware                                             │
│  ├─ 解析 JWT → Claims                                       │
│  ├─ 设置: UserID, TenantID, Roles                           │
│  └─ Context → request.Context                               │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│  RoleMiddleware                                              │
│  ├─ 检查角色是否允许切换租户                                 │
│  ├─ 解析 X-Target-Tenant (如果有)                            │
│  ├─ 验证租户权限（超管任意/监管方授权）                       │
│  └─ 更新: TenantID, TenantCode                              │
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
│  └─ 所有组件通过 context.Value() 获取租户信息                 │
└─────────────────────────────────────────────────────────────┘
```

---

## 八、API 接口设计原则

### 8.1 核心原则

| 原则 | 说明 | 示例 |
|------|------|------|
| **Context 优先** | 租户ID从 Context 获取，不传参数 | `GET /api/v1/users` |
| **Header 切换** | 超管通过 Header 切换目标租户 | `X-Target-Tenant: tenant_a` |
| **路径可选** | 特定场景支持 Path 参数 | `GET /api/v1/tenants/:code/users` |
| **业务无关** | Handler/Service 不关心租户逻辑 | 所有角色共用一套接口 |

### 8.2 接口示例

```bash
# ========== 普通用户/租户管理员 ==========
# 获取当前租户用户列表（自动使用 JWT 租户）
GET /api/v1/users
Headers: Authorization: Bearer <token>

# ========== 超级管理员 ==========
# 切换到租户 A 查看用户
GET /api/v1/users
Headers:
  Authorization: Bearer <super_admin_token>
  X-Target-Tenant: tenant_a

# 查看平台所有租户（不需要租户上下文）
GET /api/v1/admin/tenants
Headers: Authorization: Bearer <super_admin_token>

# 为租户 B 创建用户
POST /api/v1/users
Headers:
  Authorization: Bearer <super_admin_token>
  X-Target-Tenant: tenant_b
Body: { "username": "new_user", ... }

# ========== 监管方 ==========
# 查看授权租户列表（范围限制）
GET /api/v1/audit/tenants
Headers: Authorization: Bearer <auditor_token>
# 返回: tenant_b, tenant_c, tenant_d

# 查看授权租户的数据
GET /api/v1/audit/logs
Headers:
  Authorization: Bearer <auditor_token>
  X-Target-Tenant: tenant_c
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
    // 自动从 context 获取租户ID（由 RoleMiddleware 设置）
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

```go
// 1. 用户登录（需要跨租户查询用户）
POST /api/v1/login
Body: { "username": "alice", "password": "xxx" }

// Handler
func (h *AuthHandler) Login(c *gin.Context) {
    // 使用 SkipTenantCheck，因为登录时还不知道租户
    ctx := database.SkipTenantCheck(c.Request.Context())
    tokenPair, err := h.authService.Login(ctx, &req)
    // ...
}

// 2. 返回 Token（包含当前租户信息）
{
  "access_token": "xxx",
  "refresh_token": "yyy",
  "expires_in": 3600,
  "roles": ["tenant_admin"]  // 或 ["super_admin"]（超管）
}
```

### 9.2 场景二：超管跨租户操作

```bash
# 超管为租户 A 添加用户
POST /api/v1/users
Headers:
  Authorization: Bearer <super_admin_token>  # roles = ["super_admin"]
  X-Target-Tenant: tenant_a
Body: { "username": "bob", "password": "xxx" }

# 流程：
# 1. AuthMiddleware → 解析 Token, roles = ["super_admin"]
# 2. RoleMiddleware → 解析 X-Target-Tenant → 更新 Context.TenantID = tenant_a
# 3. CasbinMiddleware → 角色为超管 → 跳过检查
# 4. UserService.Create → 从 Context 获取 tenant_a → 创建用户
```

### 9.3 场景三：监管方查看多租户数据

```go
// 登录时设置授权租户列表
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
    // ...
    // 查询用户授权的租户列表
    allowedTenants := s.auditService.GetAllowedTenants(ctx, user.UserID)

    tokenPair, err := s.jwtManager.GenerateTokenPair(
        // ...,
        allowedTenants=allowedTenants,
    )
    // ...
}

// 查看时验证
GET /api/v1/audit/logs
Headers:
  Authorization: Bearer <auditor_token>  # roles = ["auditor"]
  X-Target-Tenant: tenant_c

// RoleMiddleware 中验证：
// if hasRole(roles, "auditor") {
//     if !contains(allowedTenants, requestTenantCode) {
//         return ErrForbidden
//     }
// }
```

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
2. **RoleMiddleware 中间件**：统一处理租户切换逻辑，Handler 无感知
3. **Casbin 智能跳过**：超管跳过检查，其他角色走策略
4. **GORM 三种模式**：Auto/Skip/Manual 灵活控制数据层过滤
5. **Context 传递**：租户信息在整个调用链中透明传递

**核心优势**：
- ✅ 单套接口服务所有角色
- ✅ Handler/Service 完全不感知租户逻辑
- ✅ 支持超管、租户管理员、监管方三种角色
- ✅ 安全可控，多层防御
- ✅ 性能优化，缓存友好
