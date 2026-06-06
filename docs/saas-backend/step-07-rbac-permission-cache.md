# Step 07: RBAC 核心 — PermissionCache / 角色继承 / RBAC 中间件

## 目标

实现纯数据库 RBAC 权限系统：内存缓存权限数据、角色继承（递归 CTE）、RBAC 中间件拦截无权请求。

## 前置条件

- Step 06 完成，登录流程可用，JWT Claims 包含 RoleIDs

## 文件清单

```
internal/
├── rbac/
│   ├── cache.go               # PermissionCache 内存缓存
│   ├── loader.go              # 从数据库加载权限数据
│   └── types.go               # 类型定义
├── middleware/
│   └── rbac.go                # RBAC 权限检查中间件
├── repository/
│   ├── permission_repo.go     # 权限查询
│   └── role_repo.go           # 角色查询（含继承）
```

## 实现规范

### 1. 权限缓存结构

```go
// internal/rbac/types.go
package rbac

// Permission 权限定义
type Permission struct {
    PermissionID string
    Name         string
    Type         string // API / MENU / BUTTON
    Resource     string // API: "/api/v1/users" or MENU: "system:user"
    Action       string // GET / POST / PUT / DELETE
}

// RolePermissions 角色拥有的权限集合
type RolePermissions struct {
    APIs    map[string]bool // key: "METHOD:PATH" → true
    Menus   map[string]bool // key: menu_resource → true
    Buttons map[string]bool // key: button_resource → true
}
```

### 2. PermissionCache

```go
// internal/rbac/cache.go
package rbac

import (
    "context"
    "sync"
    "time"
    "github.com/rs/zerolog"
)

type PermissionCache struct {
    mu       sync.RWMutex
    // key: roleID → 该角色拥有的权限（含继承）
    rolePerms map[string]*RolePermissions

    loader   *Loader
    log      zerolog.Logger

    // 刷新机制
    refreshCh chan struct{}
    ttl       time.Duration
}

func NewPermissionCache(loader *Loader, log zerolog.Logger, ttl time.Duration) *PermissionCache {
    pc := &PermissionCache{
        rolePerms: make(map[string]*RolePermissions),
        loader:    loader,
        log:       log,
        refreshCh: make(chan struct{}, 1),
        ttl:       ttl,
    }
    // 首次加载
    if err := pc.reload(context.Background()); err != nil {
        log.Error().Err(err).Msg("initial permission cache load failed")
    }
    // 启动后台刷新
    go pc.watchRefresh()
    return pc
}

// NotifyRefresh 通知缓存刷新（非阻塞）
func (pc *PermissionCache) NotifyRefresh() {
    select {
    case pc.refreshCh <- struct{}{}:
    default:
    }
}

// CheckAPI 检查角色是否有某 API 权限
func (pc *PermissionCache) CheckAPI(roleIDs []string, method, path string) bool {
    pc.mu.RLock()
    defer pc.mu.RUnlock()

    key := method + ":" + path
    for _, roleID := range roleIDs {
        if perms, ok := pc.rolePerms[roleID]; ok {
            if perms.APIs[key] {
                return true
            }
        }
    }
    return false
}

// GetMenus 获取角色能看到的菜单列表
func (pc *PermissionCache) GetMenus(roleIDs []string) []string {
    pc.mu.RLock()
    defer pc.mu.RUnlock()

    menus := make(map[string]bool)
    for _, roleID := range roleIDs {
        if perms, ok := pc.rolePerms[roleID]; ok {
            for m := range perms.Menus {
                menus[m] = true
            }
        }
    }
    result := make([]string, 0, len(menus))
    for m := range menus {
        result = append(result, m)
    }
    return result
}

// reload 从数据库重新加载所有权限数据
func (pc *PermissionCache) reload(ctx context.Context) error {
    data, err := pc.loader.LoadAll(ctx)
    if err != nil {
        return err
    }
    pc.mu.Lock()
    pc.rolePerms = data
    pc.mu.Unlock()
    pc.log.Info().Int("roles", len(data)).Msg("permission cache reloaded")
    return nil
}

// watchRefresh 后台协程：事件驱动 + TTL 定时器
func (pc *PermissionCache) watchRefresh() {
    ticker := time.NewTicker(pc.ttl)
    defer ticker.Stop()
    for {
        select {
        case <-pc.refreshCh:
            _ = pc.reload(context.Background())
        case <-ticker.C:
            _ = pc.reload(context.Background())
        }
    }
}
```

**缓存刷新策略**：
- **事件驱动**：权限变更时调用 `NotifyRefresh()`，立即生效
- **TTL 兜底**：默认 5 分钟自动刷新，防止事件丢失
- **非阻塞通知**：`refreshCh` 容量为 1，重复通知不阻塞

### 3. Loader — 从数据库加载

```go
// internal/rbac/loader.go
package rbac

import (
    "context"
    "admin/internal/query"
    "gorm.io/gorm"
)

type Loader struct {
    db *gorm.DB
    q  *query.Query
}

func NewLoader(db *gorm.DB) *Loader {
    return &Loader{db: db, q: query.Use(db)}
}

// LoadAll 加载所有角色的权限（含继承）
// 不带 tenant_id 过滤：缓存是全局的，RBAC 中间件用 RoleIDs 匹配
func (l *Loader) LoadAll(ctx context.Context) (map[string]*RolePermissions, error) {
    // 1. 查所有 role_permissions 关联
    // 2. 查所有 permissions 定义
    // 3. 构建 role_id → RolePermissions 映射
    // 4. 处理角色继承（递归 CTE 或代码递归）：
    //    子角色继承父角色的所有权限
    //    遍历 roles 表的 parent_role_id，向上追溯

    // SQL: 递归 CTE 查角色祖先链
    // WITH RECURSIVE role_ancestors AS (
    //     SELECT role_id, parent_role_id, 0 AS depth FROM roles WHERE deleted_at = 0
    //     UNION ALL
    //     SELECT ra.role_id, r.parent_role_id, ra.depth + 1
    //     FROM role_ancestors ra
    //     JOIN roles r ON ra.parent_role_id = r.role_id
    //     WHERE r.deleted_at = 0 AND ra.depth < 10
    // )
    // SELECT role_id, parent_role_id FROM role_ancestors;

    // 通过祖先链，子角色继承所有祖先角色的权限
    return nil, nil // 具体实现
}
```

### 4. RBAC 中间件

```go
// internal/middleware/rbac.go
package middleware

import (
    "admin/internal/rbac"
    "admin/pkg/response"
    "admin/pkg/xcontext"
    "admin/pkg/xerr"
    "github.com/gin-gonic/gin"
)

func RBAC(cache *rbac.PermissionCache) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := c.Request.Context()

        // 超管绕行
        if xcontext.HasRole(ctx, "super_admin") {
            c.Next()
            return
        }

        // 检查 API 权限
        roleIDs := xcontext.GetRoleIDs(ctx)
        method := c.Request.Method
        path := c.FullPath() // Gin 注册的路由模板，如 /api/v1/users/:id

        if !cache.CheckAPI(roleIDs, method, path) {
            response.Fail(c, xerr.New(xerr.CodeForbidden, "无操作权限"))
            c.Abort()
            return
        }

        c.Next()
    }
}
```

**关键设计决策**：

| 决策 | 选择 | 原因 |
|------|------|------|
| 超管绕行 | `HasRole("super_admin")` 直接放行 | 简单可靠，一行判断 |
| 匹配路径 | `c.FullPath()` 路由模板 | `/users/:id` 而非 `/users/123`，避免路径参数干扰 |
| 缓存粒度 | 按 roleID 索引 | 同一角色跨租户权限一致（permissions 是全局表） |
| 继承方向 | 子角色继承父角色 | 子角色 = 父角色权限 + 自身额外权限 |

### 5. 路由注册更新

```go
// 在 router.go 的 auth 路由组添加 RBAC 中间件
auth := engine.Group("/api/v1")
auth.Use(middleware.JWTAuth(jwtMgr))
auth.Use(middleware.RBAC(rbacCache))
{
    // 所有需要鉴权的路由
}
```

### 6. 权限变更后刷新

当以下操作发生时，Service 层调用 `rbacCache.NotifyRefresh()`：
- 角色权限分配/取消
- 角色创建/删除
- 角色继承关系变更

```go
// 在 service 层操作权限后
s.rbacCache.NotifyRefresh()
```

## 验收标准

```bash
# 1. 编译通过
go build ./...

# 2. 缓存加载
# 启动后日志包含："permission cache reloaded" + roles 数量

# 3. 超管访问任意接口
TOKEN_ADMIN=...  # super_admin 角色
curl -s http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $TOKEN_ADMIN"
# 期望：200（或因尚未实现 users 接口而 404，但不是 403）

# 4. 普通用户无权限
TOKEN_USER=...  # 无 API 权限的角色
curl -s http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $TOKEN_USER"
# 期望：{"code":40300,"message":"无操作权限"}

# 5. 权限分配后生效
# 给角色分配 GET:/api/v1/users 权限 → NotifyRefresh()
# 再次请求 → 200

# 6. 角色继承测试
# 父角色有 A 权限，子角色能继承 A 权限
```

## AI 协作提示

```
请按 step-07-rbac-permission-cache.md 实现 RBAC 权限系统。

要点：
1. internal/rbac/types.go — Permission 和 RolePermissions 类型
2. internal/rbac/cache.go — PermissionCache 内存缓存 + 读写锁 + 双刷新机制
3. internal/rbac/loader.go — 从数据库加载，递归 CTE 处理角色继承
4. internal/middleware/rbac.go — 超管绕行 + CheckAPI 权限检查
5. 缓存 key 是 roleID，值是该角色（含继承）的所有权限
6. CheckAPI 用 "METHOD:FullPath" 匹配
7. NotifyRefresh() 非阻塞刷新
8. 在 router.go auth 组添加 RBAC 中间件
9. 更新 main.go 构造 PermissionCache 并注入
```

---

*上一步：[Step 06 - 登录认证](step-06-auth-login.md) | 下一步：[Step 08 - 权限管理接口](step-08-permission-management.md)*
