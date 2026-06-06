# Step 13: 菜单管理域

## 目标

实现菜单树的 CRUD 管理，以及按当前用户角色返回可见菜单。菜单是全局表，由超管维护。

## 前置条件

- Step 11 完成（角色可分配菜单）

## 文件清单

```
internal/
├── handler/menu/
│   ├── handler.go
│   ├── create.go
│   ├── update.go
│   ├── delete.go
│   ├── query.go               # Tree（全部） + UserMenus（按角色过滤）
│   └── dto.go
├── service/menu/
│   ├── service.go
│   ├── create.go
│   ├── update.go
│   ├── delete.go
│   ├── query.go
│   └── converter.go
├── repository/
│   └── menu_repo.go
```

## 实现规范

### 1. API 设计

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/menus/tree | 完整菜单树（管理用，super_admin） |
| GET | /api/v1/menus/user | 当前用户可见菜单（按角色过滤） |
| POST | /api/v1/menus | 创建菜单 |
| PUT | /api/v1/menus/:id | 更新菜单 |
| DELETE | /api/v1/menus/:id | 删除菜单 |

### 2. 菜单类型

```
1 = 目录（Directory）：容器，不对应实际页面
2 = 菜单（Menu）：对应前端路由页面
3 = 按钮（Button）：页面内的操作按钮
```

### 3. 用户菜单获取流程

```
1. 超管 → 返回全部菜单（type=1,2）
2. 普通用户 →
   a. 从 xcontext 获取 RoleIDs
   b. 查 role_menus 得到该用户所有角色关联的 menu_ids
   c. 查 menus 表（IN menu_ids, status=1）
   d. 构建树形返回
3. 只返回 type=1（目录）和 type=2（菜单），不返回按钮
4. 按钮权限通过 /menus/buttons 接口单独获取（或合并到 user 菜单中的 buttons 字段）
```

### 4. DTO

```go
type CreateMenuRequest struct {
    ParentID   string `json:"parent_id"`
    Name       string `json:"name" binding:"required,max=100"`
    Path       string `json:"path"`       // 前端路由
    Component  string `json:"component"`  // 前端组件
    Icon       string `json:"icon"`
    Sort       int    `json:"sort"`
    Type       int    `json:"type" binding:"required,oneof=1 2 3"`
    Visible    int    `json:"visible" binding:"omitempty,oneof=1 2"`
    Permission string `json:"permission"` // 权限标识
}

type MenuTree struct {
    MenuID     string      `json:"menu_id"`
    ParentID   string      `json:"parent_id"`
    Name       string      `json:"name"`
    Path       string      `json:"path"`
    Component  string      `json:"component"`
    Icon       string      `json:"icon"`
    Sort       int         `json:"sort"`
    Type       int         `json:"type"`
    Visible    int         `json:"visible"`
    Permission string      `json:"permission"`
    Children   []*MenuTree `json:"children"`
}

// UserMenu 用户菜单（精简，给前端渲染用）
type UserMenu struct {
    MenuID    string      `json:"menu_id"`
    Name      string      `json:"name"`
    Path      string      `json:"path"`
    Component string      `json:"component"`
    Icon      string      `json:"icon"`
    Sort      int         `json:"sort"`
    Children  []*UserMenu `json:"children"`
    Buttons   []string    `json:"buttons,omitempty"` // 该菜单下的按钮权限标识
}
```

### 5. 关键设计

- 菜单是**全局表**（无 tenant_id），所有租户共享菜单定义
- 菜单可见性由 `role_menus` 关联表控制（带 tenant_id）
- 每个租户的每个角色可以独立配置看到哪些菜单
- 树形构建与部门相同：一次性查出 → 内存组装

### 6. 按钮权限

```go
// 获取用户在某菜单下的按钮权限
// 前端用来控制按钮显隐
// 返回格式：["user:create", "user:delete", "user:update"]

func (s *Service) GetUserButtons(ctx context.Context) ([]string, error) {
    if xcontext.HasRole(ctx, "super_admin") {
        // 超管返回所有按钮
        return s.menuRepo.GetAllButtons(ctx)
    }
    roleIDs := xcontext.GetRoleIDs(ctx)
    return s.menuRepo.GetButtonsByRoles(ctx, roleIDs)
}
```

## 验收标准

```bash
# 1. 完整菜单树（超管）
curl -s http://localhost:8080/api/v1/menus/tree \
  -H "Authorization: Bearer $TOKEN_ADMIN"
# 期望：返回完整菜单树

# 2. 用户菜单（按角色过滤）
curl -s http://localhost:8080/api/v1/menus/user \
  -H "Authorization: Bearer $TOKEN_USER"
# 期望：只返回该用户角色关联的菜单

# 3. 超管看所有
curl -s http://localhost:8080/api/v1/menus/user \
  -H "Authorization: Bearer $TOKEN_ADMIN"
# 期望：返回所有菜单（type=1,2）

# 4. 创建菜单
curl -s -X POST http://localhost:8080/api/v1/menus \
  -H "Authorization: Bearer $TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{"name":"系统管理","type":1,"sort":1,"icon":"setting"}'
# 期望：成功

# 5. 删除菜单（有子菜单时拒绝）
# 期望：{"code":40900,"message":"菜单存在子项，无法删除"}
```

## AI 协作提示

```
请按 step-13-domain-menu.md 实现菜单管理域。

要点：
1. 菜单是全局表（无 tenant_id），只有 super_admin 能 CRUD
2. role_menus 关联有 tenant_id，控制可见性
3. /menus/user 接口按当前用户 RoleIDs 过滤
4. 超管返回全部菜单
5. 树形构建：内存组装
6. 按钮权限单独收集返回
7. type: 1=目录 2=菜单 3=按钮
```

---

*上一步：[Step 12 - 部门管理](step-12-domain-department.md) | 下一步：[Step 14 - 数据权限](step-14-data-permission.md)*
