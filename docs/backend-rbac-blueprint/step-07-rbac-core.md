# Step 07: RBAC 核心 — PermissionCache / 中间件 / 角色继承

## 目标

实现 RBAC 权限缓存和检查中间件：从数据库加载权限到内存，支持角色继承（递归 CTE），支持 API/MENU/BUTTON 三种权限类型。

## 前置条件

- Step 06 完成，认证流程可用
- Step 04 完成，permissions 和 role_permissions 表存在

## 文件清单

```
internal/
├── rbac/
│   └── cache.go               # PermissionCache 核心
├── middleware/
│   └── rbac.go                 # RBAC 中间件
└── repository/
    ├── permission_repo.go      # 权限查询
    └── role_permission_repo.go # 角色-权限查询

pkg/constants/
└── system.go                  # 角色编码、权限类型常量
```

## 实现细节

### 1. PermissionCache 结构

```go
type PermissionCache struct {
    mu          sync.RWMutex
    apiPerms    map[string][]APIPermission  // roleID → [{path, method}]
    menuPerms   map[string][]string         // roleID → [menuID]
    buttonPerms map[string][]string         // roleID → [permissionID]
    db          *gorm.DB
    ttl         time.Duration
    done        chan struct{}
}

type APIPermission struct {
    Path   string
    Method string
}
```

### 2. 三重刷新机制

```
启动加载：watchRefresh() 协程首次运行
TTL 定时：每 30 秒 ticker
事件驱动：NotifyRefresh() 在角色/权限变更时调用
```

### 3. 角色继承 — 递归 CTE

```sql
WITH RECURSIVE role_ancestors AS (
    SELECT role_id, role_id AS ancestor_role_id, 0 AS depth
    FROM roles
    WHERE deleted_at = 0

    UNION ALL

    SELECT r.role_id, ra.ancestor_role_id, ra.depth + 1
    FROM roles r
    JOIN role_ancestors ra ON r.parent_role_id = ra.role_id
    WHERE ra.depth < 10 AND r.deleted_at = 0
)
SELECT DISTINCT ra.role_id, p.resource, p.action
FROM role_ancestors ra
JOIN role_permissions rp ON rp.role_id = ra.ancestor_role_id
JOIN permissions p ON p.permission_id = rp.permission_id
WHERE ra.role_id = ANY(?)
  AND p.type = 'API'
  AND p.status = 1
  AND p.deleted_at = 0;
```

> 同一条 CTE，`p.type = 'MENU'` 和 `p.type = 'BUTTON'` 分别查菜单和按钮权限。

### 4. 权限检查方法

```go
// 检查 API 权限
func (c *PermissionCache) CheckAPI(roleIDs []string, path, method string) bool

// 获取用户菜单
func (c *PermissionCache) GetUserMenus(roleIDs []string) []string

// 获取用户按钮权限
func (c *PermissionCache) GetUserButtons(roleIDs []string) []string

// 通知刷新（事件驱动）
func (c *PermissionCache) NotifyRefresh()
```

### 5. RBAC 中间件

```go
func RBAC(cache *rbac.PermissionCache) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := c.Request.Context()

        // 超管绕行 — 一行搞定
        if xcontext.HasRole(ctx, constants.RoleSuperAdmin) {
            c.Next()
            return
        }

        // 获取 roleIDs
        roleIDs := xcontext.GetRoleIDs(ctx)
        if len(roleIDs) == 0 {
            response.Fail(c, xerr.ErrForbidden)
            c.Abort()
            return
        }

        // 检查 API 权限
        path := c.Request.URL.Path
        method := c.Request.Method
        if !cache.CheckAPI(roleIDs, path, method) {
            response.Fail(c, xerr.ErrForbidden)
            c.Abort()
            return
        }

        c.Next()
    }
}
```

### 6. 路径匹配规则

```
精确匹配：/api/v1/users          → 只匹配这个路径
单段通配：/api/v1/users/:id      → 匹配 /api/v1/users/xxx
全局通配：/api/v1/**             → 匹配 /api/v1/ 下所有路径
方法通配：method = "*"           → 匹配所有 HTTP 方法
```

### 7. 常量定义

```go
// pkg/constants/system.go
const (
    RoleSuperAdmin = "super_admin"
    RoleAdmin      = "admin"
    RoleAuditor    = "auditor"
    RoleUser       = "user"

    TypeAPI   = "API"
    TypeMenu  = "MENU"
    TypeButton = "BUTTON"
    TypeData  = "DATA"

    StatusEnabled  = 1
    StatusDisabled = 2
)
```

### 8. App 初始化集成

```go
// router/app.go
func (a *App) initRBAC() {
    a.RBAC = rbac.NewPermissionCache(a.DB, 30*time.Second)
}
```

## 验收标准

- [ ] PermissionCache 启动时成功加载所有权限
- [ ] 普通角色有权限的 API → 请求通过
- [ ] 普通角色无权限的 API → 返回 403
- [ ] 超管（super_admin）→ 所有 API 都通过
- [ ] 无角色的用户 → 返回 403
- [ ] 角色继承：子角色能访问父角色的权限
- [ ] 路径通配符匹配正确（`:id`, `**`）
- [ ] `NotifyRefresh()` 调用后，新权限立即生效
- [ ] 30 秒 TTL 刷新正常工作

## AI 协作提示

```
请按 step-07-rbac-core.md 实现 RBAC 核心。
这是权限系统的心脏，必须确保：
1. PermissionCache 的三重刷新机制（启动/TTL/事件）
2. 递归 CTE 正确解析角色继承（depth < 10 防循环）
3. 超管绕行只有一行代码
4. 路径匹配支持精确/通配/全局三种模式
5. 刷新时不影响并发读（RWMutex）
参考现有 backend/internal/rbac/cache.go 的递归 CTE 实现。
```
