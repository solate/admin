# Step 14: 数据权限 — data_scope 五级过滤

## 目标

实现行级数据权限：根据角色的 data_scope 设置，自动过滤查询结果，确保用户只能看到自己权限范围内的数据。

## 前置条件

- Step 12 完成（部门树可用）
- Step 11 完成（角色有 data_scope 字段）

## 文件清单

```
internal/
├── rbac/
│   └── datascope.go           # 数据权限过滤器
├── repository/
│   └── department_repo.go     # 补充：GetChildIDs（递归子部门）
```

## 实现规范

### 1. 五级数据权限

| data_scope | 含义 | SQL 条件 |
|------------|------|----------|
| 1 | 全部数据 | 不加条件（仅限本租户） |
| 2 | 自定义部门 | `WHERE department_id IN (自定义 dept_ids)` |
| 3 | 本部门及以下 | `WHERE department_id IN (本部门 + 所有子部门)` |
| 4 | 仅本部门 | `WHERE department_id = 当前用户部门` |
| 5 | 仅本人 | `WHERE user_id = 当前用户 ID`（或 `created_by`） |

### 2. 数据权限过滤器

```go
// internal/rbac/datascope.go
package rbac

import (
    "context"
    "admin/pkg/xcontext"
)

// DataScope 数据权限级别
const (
    DataScopeAll         = 1 // 全部
    DataScopeCustom      = 2 // 自定义部门
    DataScopeDeptBelow   = 3 // 本部门及以下
    DataScopeDeptOnly    = 4 // 仅本部门
    DataScopeSelf        = 5 // 仅本人
)

// ScopeResult 数据权限计算结果
type ScopeResult struct {
    Scope     int      // 数据权限级别
    DeptIDs   []string // 允许的部门 ID 列表（scope=2,3,4 时有效）
    UserID    string   // 当前用户 ID（scope=5 时有效）
}

// CalcDataScope 计算当前用户的数据权限
// 多角色取最大权限（数值最小 = 权限最大）
func CalcDataScope(ctx context.Context, roles []*RoleDataScope, childDeptFn func(deptID string) []string) *ScopeResult {
    // 超管直接全部
    if xcontext.HasRole(ctx, "super_admin") {
        return &ScopeResult{Scope: DataScopeAll}
    }

    userID := xcontext.GetUserID(ctx)
    userDeptID := xcontext.GetDepartmentID(ctx) // 需要在 Claims 中携带

    // 多角色取最宽权限
    bestScope := DataScopeSelf
    var customDeptIDs []string

    for _, role := range roles {
        if role.DataScope < bestScope {
            bestScope = role.DataScope
        }
        if role.DataScope == DataScopeCustom {
            customDeptIDs = append(customDeptIDs, role.DeptIDs...)
        }
    }

    switch bestScope {
    case DataScopeAll:
        return &ScopeResult{Scope: DataScopeAll}
    case DataScopeCustom:
        return &ScopeResult{Scope: DataScopeCustom, DeptIDs: unique(customDeptIDs)}
    case DataScopeDeptBelow:
        deptIDs := childDeptFn(userDeptID) // 本部门 + 所有子部门
        deptIDs = append(deptIDs, userDeptID)
        return &ScopeResult{Scope: DataScopeDeptBelow, DeptIDs: deptIDs}
    case DataScopeDeptOnly:
        return &ScopeResult{Scope: DataScopeDeptOnly, DeptIDs: []string{userDeptID}}
    default:
        return &ScopeResult{Scope: DataScopeSelf, UserID: userID}
    }
}

type RoleDataScope struct {
    RoleID    string
    DataScope int
    DeptIDs   []string
}
```

### 3. Repository 中使用

```go
// 在需要数据权限的 Repository 方法中：
func (r *UserRepo) ListWithScope(ctx context.Context, scope *rbac.ScopeResult, offset, limit int) ([]*model.User, int64, error) {
    u := r.q.User
    q := u.WithContext(ctx).
        Where(u.TenantID.Eq(xcontext.GetTenantID(ctx))).
        Where(u.DeletedAt.Eq(0))

    // 应用数据权限
    switch scope.Scope {
    case rbac.DataScopeAll:
        // 不加条件
    case rbac.DataScopeCustom, rbac.DataScopeDeptBelow, rbac.DataScopeDeptOnly:
        q = q.Where(u.DepartmentID.In(scope.DeptIDs...))
    case rbac.DataScopeSelf:
        q = q.Where(u.UserID.Eq(scope.UserID))
    }

    total, _ := q.Count()
    users, err := q.Order(u.CreatedAt.Desc()).Order(u.UserID.Desc()).
        Offset(offset).Limit(limit).Find()
    return users, total, err
}
```

### 4. Service 层调用

```go
// service/user/query.go
func (s *Service) List(ctx context.Context, req *ListUserRequest) ([]UserInfo, int64, error) {
    // 1. 计算数据权限
    roles := s.getUserRoleScopes(ctx)
    scope := rbac.CalcDataScope(ctx, roles, s.deptRepo.GetChildIDs)

    // 2. 带数据权限查询
    users, total, err := s.userRepo.ListWithScope(ctx, scope, req.GetOffset(), req.GetPageSize())
    // ...
}
```

### 5. 哪些接口需要数据权限

| 模块 | 接口 | 过滤字段 |
|------|------|----------|
| 用户列表 | GET /users | department_id |
| 操作日志 | GET /operation-logs | user_id / department_id |
| 审批流程 | 未来扩展 | 提交人 department_id |

**不需要数据权限的**：
- 角色管理（按租户隔离即可）
- 部门管理（管理员需看全局）
- 菜单管理（全局表）
- 权限管理（全局表）

### 6. department_id 在 xcontext 中

需要在 JWT Claims 中额外携带 `department_id`，或在 Auth 中间件中从数据库查一次。

**推荐方案**：在 Claims 中携带，避免每次请求查库。

```go
// 更新 Claims 结构
type Claims struct {
    // ... 现有字段
    DepartmentID string `json:"dept_id"` // 新增
}
```

**注意**：用户切换部门后需要重新登录获取新 Token。

## 验收标准

```bash
# 1. data_scope=1（全部）
# 超管或 scope=1 的角色 → 看到本租户所有用户

# 2. data_scope=4（仅本部门）
# 用户属于"技术部" → 只看到"技术部"的用户

# 3. data_scope=3（本部门及以下）
# 用户属于"技术部" → 看到"技术部" + "后端组" + "前端组"

# 4. data_scope=5（仅本人）
# 只看到自己

# 5. 多角色取最宽
# 用户有两个角色：scope=4 + scope=1 → 取 scope=1（全部）

# 6. 性能
# GetChildIDs 做缓存（部门数据变动少），避免每次递归查库
```

## AI 协作提示

```
请按 step-14-data-permission.md 实现数据权限。

要点：
1. internal/rbac/datascope.go — 五级权限计算
2. 多角色取最宽权限（数值最小）
3. Repository 提供 ListWithScope 方法
4. Service 层调用 CalcDataScope → 传给 Repository
5. Claims 中增加 department_id
6. GetChildIDs 查部门子树（可缓存）
7. 超管直接 DataScopeAll
```

---

*上一步：[Step 13 - 菜单管理](step-13-domain-menu.md) | 下一步：[Step 15 - 审计日志](step-15-audit-log.md)*
