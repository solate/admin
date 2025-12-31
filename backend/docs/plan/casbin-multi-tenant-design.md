# Casbin 多租户权限系统设计

## 核心设计

**一句话总结**：Casbin 记录 用户→角色→权限 的映射关系，通过中间件自动鉴权，开发者无感知。

---

## 1. Casbin 记录什么？

### 三种策略

| 策略 | 用途 | 格式 | 示例 |
|------|------|------|------|
| **g** | 用户角色绑定 | `g, 用户, 角色, 租户` | `g, alice, admin, tenant_a` |
| **p** | 角色权限定义 | `p, 角色, 租户, 资源, 操作` | `p, admin, tenant_a, /api/v1/users, GET` |
| **g2** | 角色继承 | `g2, 父角色, 子角色` | `g2, senior_admin, admin` |

### 职责划分

```
┌─────────────────────────────────────────────────────────────┐
│  业务数据库                  │  Casbin                        │
│  ─────────────              │  ─────────────                 │
│  users: 用户信息             │  g 策略: 用户属于哪个角色        │
│  roles: 角色元数据           │  p 策略: 角色拥有哪些权限        │
│  permissions: 权限元数据      │  g2 策略: 角色继承关系          │
└─────────────────────────────────────────────────────────────┘
```

---

## 2. Casbin 模型

```conf
[request_definition]
r = sub, dom, obj, act  # 用户, 租户, 资源, 操作

[policy_definition]
p = sub, dom, obj, act  # 角色, 租户, 资源, 操作

[role_definition]
g = _, _, _    # 用户-角色-租户
g2 = _, _      # 角色继承

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act)
```

---

## 3. 完整示例

### 3.1 场景：租户 A 的权限配置

```conf
# 用户角色绑定 (g)
g, alice, admin, tenant_a
g, bob, user, tenant_a

# 角色权限策略 (p)
p, admin, tenant_a, /api/v1/users, GET
p, admin, tenant_a, /api/v1/roles, GET
p, admin, tenant_a, /api/v1/roles, POST
p, user, tenant_a, /api/v1/profile, GET

# 角色继承 (g2)
g2, senior_admin, admin
```

### 3.2 数据存储形式

| ptype | v0 | v1 | v2 | v3 |
|-------|----|----|----|----|
| g | alice | admin | tenant_a | |
| p | admin | tenant_a | /api/v1/users | GET |

---

## 4. JWT + Casbin 完整流程

### 4.1 登录时生成 Token（包含角色）

```go
type Claims struct {
    UserID     string   `json:"user_id"`
    Username   string   `json:"username"`
    TenantID   string   `json:"tenant_id"`
    TenantCode string   `json:"tenant_code"`
    Roles      []string `json:"roles"`       // ⭐ 从 Casbin 查询
}

func GenerateToken(user *User) (string, error) {
    roles := enforcer.GetRolesForUserInDomain(user.Username, user.TenantCode)
    claims := Claims{..., Roles: roles}
    return jwt.NewWithClaims(...).SignedString(secret)
}
```

### 4.2 AuthMiddleware（解析 Token → Context）

```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        claims := auth.ParseToken(c.GetHeader("Authorization"))

        // ⭐ 将用户信息存入 Context
        ctx := c.Request.Context()
        ctx = xcontext.SetUserID(ctx, claims.UserID)
        ctx = xcontext.SetTenantID(ctx, claims.TenantID)
        ctx = xcontext.SetTenantCode(ctx, claims.TenantCode)
        ctx = xcontext.SetRoles(ctx, claims.Roles)
        c.Request = c.Request.WithContext(ctx)

        c.Next()
    }
}
```

### 4.3 CasbinMiddleware（超管跳过，其他走鉴权）

```go
func CasbinMiddleware(enforcer *casbin.Enforcer) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := c.Request.Context()

        // 超管跳过
        if xcontext.IsSuperAdmin(ctx) {
            ctx = database.SkipTenantCheck(ctx)
            c.Request = c.Request.WithContext(ctx)
            c.Next()
            return
        }

        // Casbin 鉴权
        ok, _ := enforcer.Enforce(
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

## 5. 接口设计（自动上下文隔离）

### 5.1 核心原则

| 原则 | 说明 |
|------|------|
| **无需传 tenant_id** | 从 Context 自动获取 |
| **无需 /me 端点** | GET /profile 返回当前用户信息 |
| **Casbin 透明** | 开发者无需关心 Casbin 细节 |

### 5.2 路由配置

```go
// ========== 简化后的路由设计 ==========
authenticated := r.Group("/api/v1")
authenticated.Use(middleware.AuthMiddleware())
authenticated.Use(middleware.CasbinMiddleware(enforcer))
{
    // ========== 租户管理（仅超管）==========
    tenant := authenticated.Group("/tenants")
    tenant.Use(middleware.SuperAdminMiddleware())  // 一行中间件搞定
    {
        tenant.POST("", handlers.CreateTenant)
        tenant.GET("", handlers.ListTenants)
        tenant.GET("/:id", handlers.GetTenant)
        tenant.PUT("/:id", handlers.UpdateTenant)
        tenant.DELETE("/:id", handlers.DeleteTenant)
    }

    // ========== 角色管理（租户管理员+超管）==========
    authenticated.POST("/roles", handlers.CreateRole)
    authenticated.GET("/roles", handlers.ListRoles)
    authenticated.GET("/roles/:id", handlers.GetRole)
    authenticated.PUT("/roles/:id", handlers.UpdateRole)
    authenticated.DELETE("/roles/:id", handlers.DeleteRole)
    
    // 权限分配
    authenticated.POST("/roles/:id/permissions", handlers.AssignPermissions)
    authenticated.GET("/roles/:id/permissions", handlers.GetRolePermissions)

    // ========== 用户角色（租户管理员+超管）==========
    authenticated.POST("/users/:id/roles", handlers.AssignRole)
    authenticated.DELETE("/users/:id/roles/:role_id", handlers.RemoveRole)
    authenticated.GET("/users/:id/roles", handlers.GetUserRoles)

    // ========== 当前用户信息（登录后自动获取）==========
    authenticated.GET("/profile", handlers.GetProfile)  // 包含用户信息、角色、权限
}
```

### 5.3 Handler 实现（无需关心 tenant_id）

```go
func (h *RoleHandler) Create(c *gin.Context) {
    var req CreateRoleRequest
    c.ShouldBindJSON(&req)

    // ⭐ Service 层自动从 Context 获取 tenant_id
    role, err := h.roleService.Create(c.Request.Context(), &req)
}

func (h *UserHandler) GetProfile(c *gin.Context) {
    ctx := c.Request.Context()
    user, _ := h.userRepo.GetByID(ctx, xcontext.GetUserID(ctx))
    roles := xcontext.GetRoles(ctx)
    response.Success(c, gin.H{"user": user, "roles": roles})
}
```

---

## 6. 安全保障：即使用户知道路径也无法访问

### 6.1 普通用户鉴权流程（无权限被拦截）

```
用户 bob (普通用户，roles = ["user"])
尝试访问: GET /api/v1/roles

1. AuthMiddleware 解析 Token → roles = ["user"]

2. CasbinMiddleware 检查:
   enforcer.Enforce("bob", "tenant_a", "/api/v1/roles", "GET")
   → bob 的角色是 user
   → user 对 /api/v1/roles 没有配置权限
   → 返回 false

3. 返回 403 Forbidden ❌
   Handler 不会执行
```

### 6.2 超级管理员鉴权流程（跳过 Casbin）

```
用户 root (超级管理员，roles = ["super_admin"])
尝试访问: GET /api/v1/tenants

1. AuthMiddleware 解析 Token → roles = ["super_admin"]

2. CasbinMiddleware 检查:
   → xcontext.IsSuperAdmin(ctx) 返回 true
   → 跳过 Casbin 鉴权 ✅
   → 同时设置 SkipTenantCheck，可跨租户查询数据

3. 直接执行 Handler ✅
   返回租户列表数据
```

### 6.3 租户管理员鉴权流程（有权限通过）

```
用户 alice (租户管理员，roles = ["admin"])
尝试访问: GET /api/v1/roles

1. AuthMiddleware 解析 Token → roles = ["admin"]

2. CasbinMiddleware 检查:
   enforcer.Enforce("alice", "tenant_a", "/api/v1/roles", "GET")
   → alice 的角色是 admin
   → admin 对 /api/v1/roles 有 GET 权限
   → 返回 true

3. 执行 Handler ✅
   返回角色列表数据（自动按 tenant_id 过滤）
```

### 6.4 三种角色的鉴权对比

| 角色 | Token 中的 roles | Casbin 检查 | 结果 |
|------|-----------------|------------|------|
| 超级管理员 | `["super_admin"]` | 跳过（`IsSuperAdmin = true`） | ✅ 直接通过 |
| 租户管理员 | `["admin"]` | 检查策略，有权限则通过 | ✅ 需要配置策略 |
| 普通用户 | `["user"]` | 检查策略，有权限则通过 | ❌ 无权限被拦截 |

### 6.5 Casbin 策略配置

```conf
# 用户角色绑定
g, root, super_admin, default    # root 是超级管理员
g, alice, admin, tenant_a        # alice 是租户管理员
g, bob, user, tenant_a           # bob 是普通用户

# 角色权限策略
p, super_admin, default, /api/v1/*, *    # ✅ 超管拥有所有权限（实际跳过检查）
p, admin, tenant_a, /api/v1/roles, GET    # ✅ admin 可以访问
# user 角色没有配置 /api/v1/roles 权限 = ❌ 无法访问
```

### 6.6 安全保障链

| 层级 | 防护 |
|------|------|
| 1 | JWT 签名验证 |
| 2 | Casbin 策略检查（中间件拦截） |
| 3 | 租户隔离 |
| 4 | 数据层过滤 |

**关键**：
- 超管在第 2 层直接跳过，不走 Casbin 检查
- 普通用户必须在 Casbin 策略中配置权限才能通过

---

## 7. 常用操作

```go
// 用户角色操作
enforcer.AddGroupingPolicy("alice", "admin", "tenant_a")
enforcer.GetRolesForUserInDomain("alice", "tenant_a")  // ["admin"]

// 角色权限操作
enforcer.AddPolicy("admin", "tenant_a", "/api/v1/users", "GET")
enforcer.GetFilteredPolicy(0, "admin", "tenant_a")

// 角色继承
enforcer.AddRoleForUser("admin", "user")  // g2, admin, user
```

---

## 8. 最佳实践

### 资源命名规范

| 类型 | 格式 | 示例 |
|------|------|------|
| API | `/api/v1/:resource` | `/api/v1/users` |
| 菜单 | `menu:模块:功能` | `menu:system:user` |
| 按钮 | `btn:模块:操作` | `btn:user:create` |
| 数据 | `data:模块:范围` | `data:user:all` |

### 性能优化

```go
// 按租户加载策略
enforcer.LoadFilteredPolicy(&gormadapter.Filter{V1: []string{"tenant_a"}})

// 角色缓存到 JWT，避免频繁查询 Casbin
roles := enforcer.GetRolesForUserInDomain("alice", "tenant_a")
token := jwt.WithClaims("roles", roles)
```

### 安全建议

1. **超管跳过 Casbin**，提升性能
2. **所有 p 策略必须包含 tenant_code**，实现租户隔离
3. **记录权限变更审计日志**
4. **新租户自动初始化默认角色和权限**

---

## 9. 实现清单

- [ ] JWT Token 包含角色列表
- [ ] AuthMiddleware 解析 Token → Context
- [ ] CasbinMiddleware 超管跳过，其他走鉴权
- [ ] GET /profile 接口
- [ ] Handler/Service 自动从 Context 获取 tenant_id
- [ ] 新租户自动初始化默认角色和权限
