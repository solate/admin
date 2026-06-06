# 架构问题深度分析：Casbin / 多租户 / GORM+Gen

> 本文基于项目实际代码，分析三个架构层面的核心痛点，给出诊断结论和改进方案。

---

## 目录

- [第一部分：Casbin 使用分析 — 是用错了，还是设计如此？](#第一部分casbin-使用分析--是用错了还是设计如此)
  - [1.1 当前实现全景图](#11-当前实现全景图)
  - [1.2 逐项问题诊断](#12-逐项问题诊断)
  - [1.3 核心矛盾分析](#13-核心矛盾分析)
  - [1.4 为什么很多项目用 Casbin](#14-为什么很多项目用-casbin)
  - [1.5 替代方案：纯数据库 RBAC](#15-替代方案纯数据库-rbac)
- [第二部分：多租户 Scope 问题分析](#第二部分多租户-scope-问题分析)
  - [2.1 当前实现全景图](#21-当前实现全景图)
  - [2.2 逐项问题诊断](#22-逐项问题诊断)
  - [2.3 核心矛盾](#23-核心矛盾)
  - [2.4 替代方案对比](#24-替代方案对比)
- [第三部分：GORM + Gen 代码生成问题](#第三部分gorm--gen-代码生成问题)
  - [3.1 表名映射问题](#31-表名映射问题)
  - [3.2 Tag 定制问题](#32-tag-定制问题)
- [第四部分：综合建议与实施路径](#第四部分综合建议与实施路径)

---

## 第一部分：Casbin 使用分析 — 是用错了，还是设计如此？

### 1.1 当前实现全景图

#### Casbin 在项目中的角色

Casbin 在本项目中承担了三个职责：

| 职责 | Casbin 策略类型 | 数据格式 |
|------|---------------|---------|
| **用户-角色绑定** | `g` 策略 | `g, username, role_code, default` |
| **角色继承** | `g2` 策略 | `g2, child_role, parent_role` |
| **权限策略** | `p` 策略 | `p, role_code, default, resource, action` |

其中权限策略包含三种资源类型：
- **API 路径权限**：`p, admin, default, /api/v1/users, GET` — 中间件自动校验
- **菜单权限**：`p, admin, default, menu:xxx, *` — 前端菜单可见性
- **按钮权限**：`p, admin, default, btn:menuID:action, *` — 前端按钮可见性

#### 数据流向

```
permissions 表（按钮定义）
    ↓
menus 表（菜单定义，含 api_paths JSON）
    ↓ 手动解析 JSON
casbin_rule 表（p 策略存储）
    ↓ 中间件 Enforce
API 请求鉴权（path + method 匹配）
```

**关键发现**：所有域都硬编码为 `"default"`（`casbin_middleware.go:51`）：

```go
authDomain := constants.DefaultTenantCode  // 永远是 "default"
```

Casbin 的 RBAC-with-Domains 特性被完全浪费。多租户隔离实际上由 GORM Callback（数据库层）处理，Casbin 只负责 API 路径鉴权。

#### 涉及文件

| 文件 | 职责 |
|------|------|
| `pkg/casbin/casbin.go` | 初始化 Enforcer + GORM Adapter |
| `pkg/casbin/rbac_model.go` | RBAC 模型定义（sub, dom, obj, act） |
| `internal/middleware/casbin_middleware.go` | Gin 中间件，调用 Enforce |
| `internal/service/role_service.go` | 角色权限 CRUD（操作 p 策略） |
| `internal/service/user_role_service.go` | 用户角色绑定（操作 g 策略） |
| `internal/service/user_menu_service.go` | 菜单权限解析（读取 p 策略 + g2 DFS） |
| `internal/repository/user_role_repo.go` | 封装 Casbin Enforcer API |

### 1.2 逐项问题诊断

#### 问题 1：需要设计 code 来对应数据库

**诊断：设计如此**

Casbin 只认字符串，不认数据库 ID。因此：
- `users` 表必须有 `user_name` 字段（Casbin 用它当 subject）
- `roles` 表必须有 `role_code` 字段（Casbin 用它当 subject/object）
- 菜单权限用 `menu:menuID` 格式，按钮权限用 `btn:menuID:action` 格式

这不是使用错误，而是 Casbin 的设计约束。如果你用数据库 ID 做 Casbin 的标识，就能用 ID 查询了——但 Casbin 的字符串格式意味着你必须维护两套标识系统。

#### 问题 2：查询不能关联 ID，必须先查 Casbin

**诊断：设计如此**

这是最痛的点。以"获取用户的菜单列表"为例（`user_menu_service.go`）：

```go
// 当前实现的查询链路（至少 4 步）：
// 1. Casbin: 获取用户角色 → g 策略
roles := s.enforcer.GetRolesForUserInDomain(userName, tenantCode)

// 2. Casbin: DFS 获取继承角色 → g2 策略
allRoleCodes, inheritedRoles := s.getAllRoleCodes(roles)

// 3. Casbin: 获取每个角色的菜单权限 → p 策略（N 次查询）
policies, _ := s.enforcer.GetFilteredPolicy(0, role, domain)

// 4. 数据库: 用解析出的 menuID 查询菜单详情
menus, err := s.menuRepo.GetByIDs(ctx, menuIDs)
```

**如果用纯数据库，只需一条 SQL**：

```sql
SELECT m.* FROM menus m
JOIN role_permissions rp ON rp.resource = 'menu:' || m.menu_id
JOIN user_roles ur ON ur.role_id = rp.role_id
WHERE ur.user_id = ? AND ur.tenant_id = ?
```

Casbin 的策略存储（`casbin_rule` 表）和业务数据（`menus`、`permissions` 表）分离，导致无法 JOIN 查询。每次都需要"先查 Casbin → 解析字符串 → 再查数据库"。

#### 问题 3：内存查询 + DB 不同步

**诊断：可优化，但根因是架构选择**

Casbin 启动时 `LoadPolicy()` 加载全量策略到内存。之后的查询都是内存操作，确实很快。但代价是：
- 直接修改数据库不生效，必须重启或调用 `LoadPolicy()`
- 多实例部署时，实例 A 修改策略，实例 B 不知道

Casbin 提供了 Watcher 机制来解决同步问题（Redis Pub/Sub），但本项目的单实例部署让这个问题不那么严重。

**关键洞察**：如果你把权限数据放在数据库表里，用一条带索引的 SQL 查询：

```sql
-- 带索引的查询，毫秒级
SELECT resource FROM role_permissions
WHERE role_id IN (?, ?) AND tenant_id = ?
```

PostgreSQL 的 B-tree 索引查询同样是毫秒级。对于管理后台的 QPS（通常 < 100），数据库查询和内存查询的差距人类无法感知。再加上可以加一层应用层缓存（`sync.Map` 或 Redis），性能完全等同。

#### 问题 4：路由与策略手动关联

**诊断：设计如此**

Casbin 是通用策略引擎，不理解"路由"这个概念。它只理解 `(sub, dom, obj, act)` 四元组。

当前的关联方式（`role_service.go:603-635`）：

```go
// 分配菜单权限时，解析菜单的 api_paths JSON，手动创建 API 策略
if menu.APIPaths != "" {
    var apiPaths []APIPath
    json.Unmarshal([]byte(menu.APIPaths), &apiPaths)
    for _, apiPath := range apiPaths {
        for _, method := range apiPath.Methods {
            s.enforcer.AddPolicy(role.RoleCode, tenantCode, apiPath.Path, method)
        }
    }
}
```

这意味着：
1. 每次分配菜单权限，都要解析 JSON → 创建多条 Casbin 策略
2. 菜单的 `api_paths` 变更时，需要同步更新所有相关 Casbin 策略
3. 如果忘记同步，前端菜单隐藏了但后端 API 还能访问（或反过来）

**更好的方式**：菜单和 API 路径的关联直接在数据库中维护，中间件从数据库（或缓存）读取关联关系。

#### 问题 5：RBAC 理解难度大

**诊断：部分用错，主要是过度设计**

本项目的 Casbin 模型包含：
- `g = _, _, _`（用户-角色-域三参数绑定）
- `g2 = _, _`（角色-角色无域继承）
- 跨域权限解析（`user_menu_service.go` 中的 DFS 遍历）

但实际使用时：
- 所有 `g` 策略的域都是 `"default"`
- `g2` 继承在应用层手动 DFS 遍历，而不是让 Casbin 自动处理
- 继承角色的跨域查询（`getMenuPermissionsForRoles` 中 `domains = append(domains, constants.DefaultTenantCode)`）是手动拼接的

这些复杂度来源于"想用 Casbin 的域来实现多租户"和"实际只用 default 域"之间的矛盾。如果你只需要简单 RBAC，一个 `user_roles` 关联表就够了。

#### 问题 6：ABAC 没利用

**诊断：设计如此，项目不需要 ABAC**

ABAC（Attribute-Based Access Control）适用于：
- "只有工作日 9-18 点可以访问"
- "只能操作自己部门的资源"
- "IP 白名单内的用户才有权限"

这些需求在管理后台中极少出现。管理后台的权限模型本质上是：
- **谁能看什么菜单** → 菜单权限
- **谁能调什么 API** → API 权限
- **谁在哪个租户** → 多租户

这三者都是 RBAC（基于角色的访问控制），不需要 ABAC。

### 1.3 核心矛盾分析

| 矛盾 | 说明 |
|------|------|
| **策略与数据无法分离** | Casbin 的核心价值是"策略与代码分离"（如外部 policy 文件），但本项目的策略来自数据库中的菜单/按钮配置，无法分离 |
| **数据双写** | 权限数据存在两处：`permissions` 表 + `casbin_rule` 表。菜单的 `api_paths` 变更时，必须同步更新 Casbin 策略，否则前后端权限不一致 |
| **能力错配** | Casbin 擅长复杂策略（ABAC、条件表达式），项目只需要简单 RBAC。用大炮打蚊子 |
| **中间件只管一半** | CasbinMiddleware 只检查 API 路径权限。菜单可见性和按钮可见性仍需 `UserMenuService` 手动解析 Casbin 策略，再查数据库 |

### 1.4 为什么很多项目用 Casbin

| 场景 | Casbin 优势 | 本项目是否需要 |
|------|------------|-------------|
| 微服务架构 | 集中式策略服务，多语言 SDK | 否，单体应用 |
| 复杂策略需求 | ABAC、属性表达式、时间条件 | 否，只需 RBAC |
| 合规审计 | 策略版本管理、审计日志 | 否，操作日志已覆盖 |
| 多语言系统 | Go/Java/Python/Node 统一策略 | 否，只有 Go |
| 外部策略管理 | 策略文件与代码分离 | 否，策略来自数据库 |

**结论：本项目引入 Casbin 的场景都不成立。**

### 1.5 替代方案：纯数据库 RBAC

#### 新表结构

```sql
-- 用户角色关联（替代 g 策略：g, username, role_code, tenant_code）
CREATE TABLE user_roles (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(20) NOT NULL REFERENCES users(user_id),
    role_id VARCHAR(20) NOT NULL REFERENCES roles(role_id),
    tenant_id VARCHAR(20) NOT NULL REFERENCES tenants(tenant_id),
    created_at BIGINT NOT NULL DEFAULT (extract(epoch from now()) * 1000)::bigint,
    UNIQUE(user_id, role_id, tenant_id)
);

CREATE INDEX idx_user_roles_user_tenant ON user_roles(user_id, tenant_id);
CREATE INDEX idx_user_roles_role ON user_roles(role_id);

-- 角色权限关联（替代 p 策略：p, role_code, domain, resource, action）
CREATE TABLE role_permissions (
    id BIGSERIAL PRIMARY KEY,
    role_id VARCHAR(20) NOT NULL REFERENCES roles(role_id),
    permission_id VARCHAR(20) NOT NULL REFERENCES permissions(permission_id),
    tenant_id VARCHAR(20) NOT NULL REFERENCES tenants(tenant_id),
    created_at BIGINT NOT NULL DEFAULT (extract(epoch from now()) * 1000)::bigint,
    UNIQUE(role_id, permission_id, tenant_id)
);

CREATE INDEX idx_role_permissions_role ON role_permissions(role_id, tenant_id);

-- 角色继承（替代 g2 策略，直接在 roles 表加字段）
ALTER TABLE roles ADD COLUMN parent_role_id VARCHAR(20) REFERENCES roles(role_id);
```

#### 权限缓存层

```go
// pkg/permission/cache.go

type PermissionCache struct {
    mu       sync.RWMutex
    apiPerms map[string][]APIPermission // key: roleID, value: 该角色的 API 权限列表
    menuPerms map[string][]string       // key: roleID, value: 菜单 ID 列表
    expAt    time.Time
    ttl      time.Duration
    db       *gorm.DB
}

type APIPermission struct {
    Path   string
    Method string
}

func NewPermissionCache(db *gorm.DB, ttl time.Duration) *PermissionCache {
    return &PermissionCache{
        apiPerms:  make(map[string][]APIPermission),
        menuPerms: make(map[string][]string),
        ttl:       ttl,
        db:        db,
    }
}

// Refresh 刷新缓存（定时任务调用，如每 30 秒）
func (c *PermissionCache) Refresh(ctx context.Context) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    // 一次性查询所有角色的 API 权限
    // SQL: SELECT r.role_id, p.resource, p.action
    //      FROM role_permissions rp
    //      JOIN permissions p ON p.permission_id = rp.permission_id
    //      WHERE p.type = 'API'
    // 加上继承角色的递归查询
    // ...
    return nil
}

// CheckAPIPermission 检查 API 权限（中间件调用）
func (c *PermissionCache) CheckAPIPermission(roleIDs []string, path, method string) bool {
    c.mu.RLock()
    defer c.mu.RUnlock()

    for _, roleID := range roleIDs {
        for _, perm := range c.apiPerms[roleID] {
            if matchPath(perm.Path, path) && matchMethod(perm.Method, method) {
                return true
            }
        }
    }
    return false
}

// GetUserMenuIDs 获取用户的菜单 ID 列表（替代 UserMenuService 中的 Casbin 查询）
func (c *PermissionCache) GetUserMenuIDs(roleIDs []string) []string {
    c.mu.RLock()
    defer c.mu.RUnlock()

    menuIDSet := make(map[string]bool)
    for _, roleID := range roleIDs {
        for _, menuID := range c.menuPerms[roleID] {
            menuIDSet[menuID] = true
        }
    }

    result := make([]string, 0, len(menuIDSet))
    for id := range menuIDSet {
        result = append(result, id)
    }
    return result
}
```

#### 替代后的中间件

```go
// internal/middleware/rbac_middleware.go

func RBACMiddleware(permCache *permission.PermissionCache) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := c.Request.Context()

        // 超管跳过
        if xcontext.HasRole(ctx, constants.SuperAdmin) {
            c.Next()
            return
        }

        userName := xcontext.GetUserName(ctx)
        tenantCode := xcontext.GetTenantCode(ctx)
        if userName == "" || tenantCode == "" {
            response.Error(c, xerr.ErrUnauthorized)
            c.Abort()
            return
        }

        // 从缓存获取用户的角色 ID 列表（一次查询）
        // 实际上角色信息已经在 JWT 中，可以从 context 获取
        roleIDs := xcontext.GetRoleIDs(ctx)

        // 检查 API 权限（纯内存操作，与 Casbin 性能相同）
        path := c.Request.URL.Path
        method := c.Request.Method

        if !permCache.CheckAPIPermission(roleIDs, path, method) {
            response.Error(c, xerr.ErrForbidden)
            c.Abort()
            return
        }

        c.Next()
    }
}
```

#### 替代后的菜单查询（对比当前实现）

**当前（Casbin，4 步）：**
```go
// user_menu_service.go 中的 GetUserMenu
roles := s.enforcer.GetRolesForUserInDomain(userName, tenantCode)  // 步骤 1
allRoleCodes, inheritedRoles := s.getAllRoleCodes(roles)            // 步骤 2: DFS
menuIDs := s.getMenuPermissionsForRoles(ctx, allRoleCodes, ...)    // 步骤 3: N 次 Casbin 查询
menus, err := s.menuRepo.GetByIDs(ctx, menuIDs)                    // 步骤 4: 数据库查询
```

**替代后（纯数据库，1 条 SQL）：**
```go
func (s *UserMenuService) GetUserMenu(ctx context.Context, userID string) (*dto.UserMenuResponse, error) {
    // 一条 SQL 完成所有查询
    menus, err := s.menuRepo.GetMenusByUserID(ctx, userID)
    // SQL:
    // SELECT DISTINCT m.* FROM menus m
    //   JOIN permissions p ON p.resource = 'menu:' || m.menu_id
    //   JOIN role_permissions rp ON rp.permission_id = p.permission_id
    //   JOIN user_roles ur ON ur.role_id = rp.role_id
    //   LEFT JOIN roles parent ON parent.role_id = r.parent_role_id
    //   LEFT JOIN role_permissions rp2 ON rp2.role_id = parent.role_id
    // WHERE ur.user_id = $1 AND ur.tenant_id = $2
    //    AND m.status = 1
    // ORDER BY m.sort_order
}
```

#### 收益对比

| 维度 | 当前（Casbin） | 替代后（数据库 RBAC） |
|------|-------------|-----------------|
| 查询方式 | 先查 Casbin → 解析 → 再查数据库 | 一条 SQL JOIN |
| 权限数据存储 | `casbin_rule` + `permissions` 双写 | `role_permissions` 单写 |
| 事务支持 | Casbin 与数据库不在同一事务 | 在同一事务中 |
| 代码量 | ~800 行（Casbin 适配 + 服务） | ~400 行（标准 CRUD） |
| AI 编程友好度 | 低（需理解 Casbin 概念） | 高（标准 SQL 模式） |
| 性能 | 内存查询（需全量加载） | 缓存 + 索引查询（同等） |
| 新增表 | 0（复用 `casbin_rule`） | 2（`user_roles` + `role_permissions`） |
| 依赖 | casbin/v2 + gorm-adapter | 无额外依赖 |

---

## 第二部分：多租户 Scope 问题分析

### 2.1 当前实现全景图

#### GORM Callback 机制

`pkg/database/scopes.go` 注册了 4 个 Callback：

| Callback | 触发时机 | 行为 |
|----------|---------|------|
| `tenant:create` | `gorm:create` 之前 | 反射设置 `tenant_id` 字段值 |
| `tenant:query` | `gorm:query` 之前 | 添加 `WHERE tenant_id = ?` |
| `tenant:update` | `gorm:update` 之前 | 添加 `WHERE tenant_id = ?` |
| `tenant:delete` | `gorm:delete` 之前 | 添加 `WHERE tenant_id = ?` |

#### 判断逻辑

```go
func tenantQueryCallback(db *gorm.DB) {
    if !hasTenantColumn(db) { return }          // 表没有 tenant_id → 跳过
    if shouldSkipTenantCheck(db) { return }     // context 标记跳过 → 跳过
    tenantID, ok := getTenantID(db)
    if !ok { db.AddError(ErrMissingTenantID); return }
    // 添加 WHERE tenant_id = ?
    db.Statement.AddClause(clause.Where{Exprs: []clause.Expression{
        clause.Eq{Column: clause.Column{Name: "tenant_id"}, Value: tenantID},
    }})
}
```

#### 表分类

**带 `tenant_id` 的表（17 张，自动过滤）：**
users, roles, departments, positions, tenant_menus, dict_types, dict_items, login_logs, operation_logs, video_files, video_risks, video_signatures, video_risk_operations, devices, device_playlists, device_approvals, access_devices

**全局表（5 张，无 tenant_id，自动跳过）：**
tenants, menus, permissions, casbin_rule, user_positions

#### Skip 机制

| 跳过方式 | 触发场景 | 文件 |
|---------|---------|------|
| `TenantSkipMiddleware` | super_admin / auditor 请求 | `internal/middleware/tenant_middleware.go` |
| `xcontext.SkipTenantCheck(ctx)` | 登录、跨租户查询、后台任务 | 各 Repository 的 `*Manual` 方法 |

### 2.2 逐项问题诊断

#### 问题 1：AI 读不到 scope，写错逻辑

**诊断：设计缺陷 — 隐式逻辑对 AI 不友好**

当 AI 看到：

```go
func (r *UserRepo) List(ctx context.Context, offset, limit int) ([]*model.User, int64, error) {
    return r.q.User.WithContext(ctx).FindByPage(offset, limit)
}
```

AI 理解为"查询所有用户"。它不知道 `FindByPage` 会触发 GORM Callback 自动添加 `WHERE tenant_id = ?`。

AI 可能会：
- 手动加上 `Where(r.q.User.TenantID.Eq(tenantID))`，导致双重过滤
- 忘记在跨租户查询时调用 `SkipTenantCheck`
- 不理解为什么某些表自动过滤、某些不过滤

#### 问题 2：tenantIDs 传入与自动 tenantID 冲突

**诊断：设计缺陷 — Callback 无法感知已有条件**

场景：超级管理员查看多租户数据

```go
// 开发者想查询多个租户的用户
r.q.User.WithContext(ctx).Where(r.q.User.TenantID.In(tenantIDs...)).Find()
```

但 Callback 会额外添加 `WHERE tenant_id = ?`（当前用户的租户），导致最终 SQL 变成：

```sql
WHERE tenant_id IN ('t1', 't2') AND tenant_id = 'admin_tenant'  -- 冲突！
```

注意：`scopes.go` 中有 `hasTenantCondition()` 函数（检测已有 tenant_id 条件），但从未被调用。这是一个未完成的修复。

#### 问题 3：隐式条件导致 AI 错误

**诊断：设计缺陷 — 魔法行为不可预测**

隐式 Callback 的问题不是"好不好用"，而是"读代码看不出它在做什么"。

对比：

```go
// 隐式（当前）— 读代码不知道有 tenant 过滤
r.q.User.WithContext(ctx).Where(r.q.User.Status.Eq(1)).Find()

// 显式 — 一目了然
r.q.User.WithContext(ctx).
    Where(r.q.User.TenantID.Eq(tenantID)).
    Where(r.q.User.Status.Eq(1)).
    Find()
```

AI（和人类开发者）看到显式代码，不会犯"忘记租户"或"重复添加"的错误。

#### 问题 4：大量 skip 点

**诊断：必然代价，但可减少**

当前需要跳过租户检查的场景：

| 场景 | 当前跳过方式 | 频率 |
|------|-----------|------|
| 登录 | `SkipTenantCheck(ctx)` | 必然 |
| 超管查看所有数据 | `TenantSkipMiddleware` | 必然 |
| 审核员跨租户 | `TenantSkipMiddleware` | 必然 |
| 跨租户批量统计 | `SkipTenantCheck(ctx)` + 原生 SQL | 必然 |
| 后台定时任务 | `SkipTenantCheck(ctx)` | 按需 |
| 按用户名/邮箱/手机登录 | `SkipTenantCheck(ctx)` | 必然 |

无论用隐式还是显式方案，这些场景都需要某种形式的"跳过"。区别在于：
- **隐式**：每个查询默认有 tenant_id，需要的地方显式跳过（当前方式）
- **显式**：每个查询默认没有，需要的地方显式添加

显式方案的"跳过"自然为零，因为根本不需要跳过——你只在你需要的地方添加 tenant_id。

#### 问题 5：表需要分带/不带 tenantID

**诊断：这是正确的架构选择，不是问题**

管理后台必然有全局表（tenants、menus、permissions）和租户隔离表（users、roles、departments）。这种分类是业务需求，与实现方式无关。

无论用 Callback 还是显式过滤，都需要知道哪些表有 tenant_id。Callback 的 `hasTenantColumn()` 通过反射自动判断，这其实是它的一个优点——至少不需要维护全局表列表。

#### 问题 6：大量跳过全局表

**诊断：自动处理，无需担心**

全局表（menus、permissions、tenants、casbin_rule）没有 `tenant_id` 列，`hasTenantColumn()` 返回 false，Callback 直接跳过。不需要手动处理。

但 AI 需要知道这个行为，否则可能困惑"为什么查询 menus 表时没有 tenant 过滤"。

### 2.3 核心矛盾

**隐式 vs 显式** — 这是一个零和博弈：

| | 隐式（Callback） | 显式（手动添加） |
|---|---|---|
| **安全性** | 高（不会忘记加 tenant_id） | 中（可能忘记加） |
| **AI 可读性** | 低（看不到隐藏逻辑） | 高（代码即文档） |
| **冲突风险** | 高（已有 tenant_id 时冲突） | 无（完全可控） |
| **Skip 数量** | 多（每个跨租户场景都要跳过） | 零（不需要跳过） |
| **调试难度** | 高（看不到实际 SQL 条件） | 低（代码直接反映 SQL） |

**冲突检测缺失**：`scopes.go` 中 `hasTenantCondition()` 函数已经写了但从未调用。如果加上这个检测，可以避免"双重 tenant_id"问题：

```go
func tenantQueryCallback(db *gorm.DB) {
    if !hasTenantColumn(db) { return }
    if shouldSkipTenantCheck(db) { return }
    if hasTenantCondition(db) { return }  // 已有 tenant_id 条件 → 跳过
    // ...
}
```

### 2.4 替代方案对比

#### 方案 A：显式 Repository 模式（推荐）

```go
// internal/repository/user_repo.go

func (r *UserRepo) List(ctx context.Context, offset, limit int) ([]*model.User, int64, error) {
    tenantID := xcontext.GetTenantID(ctx)
    return r.q.User.WithContext(ctx).
        Where(r.q.User.TenantID.Eq(tenantID)).
        FindByPage(offset, limit)
}

// 跨租户查询（不需要 SkipTenantCheck）
func (r *UserRepo) ListByTenantIDs(ctx context.Context, tenantIDs []string, offset, limit int) ([]*model.User, int64, error) {
    return r.q.User.WithContext(ctx).
        Where(r.q.User.TenantID.In(tenantIDs...)).
        FindByPage(offset, limit)
}

// 登录查询（不需要 SkipTenantCheck）
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
    return r.q.User.WithContext(ctx).
        Where(r.q.User.Email.Eq(email)).
        First()
}
```

**优点**：
- AI 可见所有条件，不会猜错
- 没有 Callback 冲突
- 不需要 SkipTenantCheck
- 跨租户查询和单租户查询用不同方法，逻辑清晰

**缺点**：
- 每个查询都要写 `Where(tenantID)`（但 AI 可以轻松生成）
- 需要移除 Callback（改动较大）

#### 方案 B：Query Wrapper 模式

```go
// pkg/database/tenant_query.go

func TenantScope(ctx context.Context, db *gorm.DB) *gorm.DB {
    if xcontext.ShouldSkipTenantCheck(ctx) {
        return db
    }
    tenantID := xcontext.GetTenantID(ctx)
    if tenantID == "" {
        return db
    }
    return db.Where("tenant_id = ?", tenantID)
}

// 使用
func (r *UserRepo) List(ctx context.Context, offset, limit int) ([]*model.User, int64, error) {
    return database.TenantScope(ctx, r.q.User.WithContext(ctx)).FindByPage(offset, limit)
}
```

**优点**：
- 单行显式调用，AI 可理解
- 保留 SkipTenantCheck 机制（兼容现有代码）
- 可以逐步迁移（新代码用 TenantScope，旧代码继续用 Callback）

**缺点**：
- 仍然依赖 context 中的 SkipTenantCheck
- 和 Callback 共存时可能双重过滤（需要先禁用 Callback）

#### 方案 C：保留 Callback + 增强 CLAUDE.md（过渡方案）

在 `CLAUDE.md` 中添加：

```markdown
## 多租户自动过滤规则

**所有带 tenant_id 列的表，GORM 会自动添加 WHERE tenant_id = ? 条件。**

### 自动过滤的表（17 张）
users, roles, departments, positions, tenant_menus, dict_types, dict_items,
login_logs, operation_logs, video_files, video_risks, video_signatures,
video_risk_operations, devices, device_playlists, device_approvals, access_devices

### 全局表（不自动过滤）
tenants, menus, permissions, casbin_rule, user_positions

### 编写 Repository 代码时注意
1. 不要手动添加 `Where(tenantID)` 条件（Callback 已自动添加）
2. 跨租户查询必须先调用 `xcontext.SkipTenantCheck(ctx)`
3. 登录等无租户上下文的场景，使用 `*Manual` 后缀的方法
4. `hasTenantCondition()` 函数存在但未启用，请勿依赖
```

**优点**：
- 不改代码，立即可用
- 帮助 AI 理解隐式行为

**缺点**：
- 不够可靠（AI 可能忽略长文档）
- 不解决冲突问题
- 新开发者仍需要学习隐式规则

#### 推荐策略

| 阶段 | 方案 | 时间 |
|------|------|------|
| **短期** | 方案 C — 增强 CLAUDE.md | 立即 |
| **中期** | 方案 B — Query Wrapper + 禁用 Callback | 迁移 Casbin 时同步做 |
| **长期** | 方案 A — 完全显式 | 新项目或全面重构时 |

---

## 第三部分：GORM + Gen 代码生成问题

### 3.1 表名映射问题

**问题**：Gen 从数据库表名生成 Go 结构体名和 `TableName()` 方法。如果期望的 Go 名称与 Gen 生成的不一致，无法在不修改生成代码的情况下覆盖。

**示例**：
- 数据库表名：`user_positions`
- Gen 生成：`UserPosition` 结构体，`TableName()` 返回 `user_positions`
- 如果想让 `TableName()` 返回其他值，无法在不编辑生成代码的情况下实现

**解决方式**：

1. **在 `gen_from_db.go` 中配置**（推荐）：

```go
// scripts/gen_from_db.go
g := gen.NewGenerator(gen.Config{
    OutPath: "internal/dal/query",
    Mode:    gen.WithoutContext,
})

// 为特定表自定义配置
g.ApplyInterface(
    func(model.CustomTableName) {},  // 自定义接口
    g.GenerateModel("user_positions"),
)
```

2. **创建 Wrapper 类型**（不修改生成代码）：

```go
// internal/model/custom/user_position.go
type UserPositionEx struct {
    model.UserPosition
}

func (UserPositionEx) TableName() string {
    return "custom_table_name"
}
```

3. **迁移 sqlc 后自然消失**：sqlc 中表名直接写在 SQL 文件里，不需要 `TableName()` 方法。

### 3.2 Tag 定制问题

**问题**：Gen 自动生成 `gorm` 和 `json` tag，无法添加自定义 tag（如 `validate:"required"`、`form:"username"`）。

**当前生成结果**：
```go
type User struct {
    UserID   string `gorm:"column:user_id;type:character varying(20);primaryKey" json:"user_id"`
    UserName string `gorm:"column:user_name;type:character varying(100);not null" json:"user_name"`
    // 想添加 validate:"required" 但无法修改生成文件
}
```

**解决方式**：

1. **Gen 配置自定义 Tag 策略**（推荐）：

```go
// scripts/gen_from_db.go
g := gen.NewGenerator(gen.Config{
    OutPath: "internal/dal/query",
    Mode:    gen.WithoutContext,
})

// 自定义 JSON tag 命名策略
g.WithJSONTagNameStrategy(func(colName string) string {
    // 将 user_name 转为 userName
    parts := strings.Split(colName, "_")
    for i := 1; i < len(parts); i++ {
        parts[i] = strings.Title(parts[i])
    }
    return strings.Join(parts, "")
})
```

2. **DTO 层转换**（当前项目已采用）：

项目已有 `internal/converter/` 和 `internal/dto/` 层，生成的 Model 不会直接暴露给 API。Tag 定制问题通过 DTO 层解决。

3. **Post-generation 脚本**（补充方案）：

```bash
# Makefile 中在 gen-db 后执行
gen-db:
	go run scripts/gen_from_db.go
	@go run scripts/postgen_tags.go  # 添加自定义 tag
```

```go
// scripts/postgen_tags.go
// 读取生成文件，为特定字段添加 validate tag
```

**结论**：当前项目的 DTO 层已经隔离了生成 Model 和 API 响应，Tag 问题的影响有限。如果迁移 sqlc，可以在 `sqlc.yaml` 中配置 `emit_json_tags` 和 `overrides` 来完全控制 tag。

---

## 第四部分：综合建议与实施路径

### 路径对比

| | 路径一：渐进式 | 路径二：全面迁移 | 路径三：最小改动 |
|---|---|---|---|
| **数据库层** | 保留 GORM + Gen | 迁移 sqlc | 保留 GORM + Gen |
| **权限系统** | Casbin → 数据库 RBAC | Casbin → 数据库 RBAC | 保留 Casbin |
| **多租户** | 保留 Callback + 增强 CLAUDE.md | 显式 Repository | 保留 Callback + 增强 CLAUDE.md |
| **改动量** | 中 | 大 | 小 |
| **风险** | 中 | 高 | 低 |
| **收益** | 高（解决 Casbin 问题） | 最高（全部解决） | 低（缓解 AI 问题） |

### 路径一：渐进式改进（推荐）

保留 GORM + Gen，去掉 Casbin，增强 AI 可见性。

**步骤**：

1. **创建数据库迁移**（~1 天）
   - 新增 `user_roles` 表（替代 Casbin g 策略）
   - 新增 `role_permissions` 表（替代 Casbin p 策略）
   - `roles` 表添加 `parent_role_id` 列（替代 g2 策略）
   - 编写数据迁移脚本（`casbin_rule` → 新表）

2. **实现权限缓存**（~1 天）
   - `pkg/permission/cache.go`：缓存 API/菜单/按钮权限
   - 启动时加载 + 定时刷新（30 秒 TTL）

3. **替换中间件**（~0.5 天）
   - `CasbinMiddleware` → `RBACMiddleware`（使用 PermissionCache）
   - 移除 `pkg/casbin/` 目录

4. **重写 Service 层**（~2 天）
   - `role_service.go`：权限分配改为操作 `role_permissions` 表
   - `user_role_service.go`：角色绑定改为操作 `user_roles` 表
   - `user_menu_service.go`：菜单查询改为 SQL JOIN

5. **清理**（~0.5 天）
   - 移除 `casbin/v2` 和 `gorm-adapter/v3` 依赖
   - 移除 `casbin_rule` 表
   - 更新 `CLAUDE.md`

6. **增强 CLAUDE.md**（~0.5 天，可在步骤 1 之前做）
   - 添加多租户 Callback 行为说明
   - 列出带/不带 tenant_id 的表
   - 添加常见错误模式警告

**总计：~5 天**

### 路径二：全面迁移

如果同时决定从 GORM 迁移到 sqlc（参见 `docs/go-database-comparison.md`），可以一步到位：

1. 多租户改为显式 Repository 模式（sqlc 天然是显式的）
2. Casbin 替换为数据库 RBAC
3. sqlc 中表名和 tag 完全可控

**总计：~10-15 天**（含 sqlc 迁移）

### 路径三：最小改动

如果不打算大改，只优化 AI 可见性：

1. 增强 `CLAUDE.md`（添加多租户规则、全局表列表）
2. 在 `scopes.go` 中启用 `hasTenantCondition()` 检测，防止双重过滤
3. 为 Casbin 操作封装更清晰的 Repository 方法

**总计：~1 天**

### 最终建议

**短期（1 天内）**：执行路径三的增强 CLAUDE.md 部分，立即改善 AI 编程体验。

**中期（下一个迭代）**：执行路径一，移除 Casbin，用纯数据库 RBAC 替代。这是性价比最高的改进。

**长期（评估后决定）**：如果决定从 GORM 迁移 sqlc，所有问题（隐式多租户、表名映射、Tag 定制）一次性解决。
