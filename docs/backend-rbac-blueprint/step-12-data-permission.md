# Step 12: 数据权限 — data_scope 五级

## 目标

在 RBAC 基础上扩展数据权限：角色可以设定数据可见范围（全部/自定义/本部门/本部门及下级/仅本人），查询时自动注入 WHERE 条件。

## 前置条件

- Step 07 RBAC 核心完成
- Step 11 部门管理完成（需要 GetDescendantIDs）

## 文件清单

```
pkg/
├── xcontext/
│   └── datascope.go           # DataScope 结构 + Get/Set

internal/
├── rbac/
│   └── cache.go               # 扩展：加载 data_scope 到缓存
├── middleware/
│   └── auth.go                # 扩展：解析 data_scope 存入 context
├── repository/
│   └── datascope.go           # ApplyDataScope 通用函数
└── dto/
    └── role_dto.go            # 扩展：DataScope 字段

pkg/constants/
└── system.go                  # DataScopeAll=1 等常量
```

## 实现细节

### 1. 五级数据范围

```go
const (
    DataScopeAll        = 1 // 全部数据
    DataScopeCustom     = 2 // 自定义部门
    DataScopeDept       = 3 // 本部门
    DataScopeDeptAndSub = 4 // 本部门及下级
    DataScopeSelf       = 5 // 仅本人
)
```

### 2. DataScope 结构

```go
// pkg/xcontext/datascope.go
type DataScope struct {
    Scope       int32    // 1-5
    DeptIDs     []string // 有效的部门 ID 列表（已展开子部门）
    UserID      string   // 当前用户 ID（仅本人模式用）
    CustomDepts []string // 自定义部门列表（JSONB 解析结果）
}
```

### 3. Auth 中间件扩展

```
在 Auth 中间件中，JWT 验证通过后：

1. 获取 roleIDs
2. 从 PermissionCache 查询角色的 data_scope
3. 如果 data_scope = 4（本部门及下级）→ 调用 GetDescendantIDs 展开子部门
4. 如果 data_scope = 2（自定义）→ 解析 data_scope_dept_ids JSONB
5. 构建 DataScope 结构 → 存入 xcontext
6. 超管跳过以上步骤
```

### 4. PermissionCache 扩展

```go
type PermissionCache struct {
    // ... 原有字段
    dataScopes map[string]*DataScopeInfo // roleID → {scope, deptIDs}
}

type DataScopeInfo struct {
    Scope         int32
    CustomDeptIDs []string // JSONB 解析结果
}
```

在 `Refresh()` 方法中增加一条查询，从 roles 表读取 `data_scope` 和 `data_scope_dept_ids`。

### 5. ApplyDataScope 通用函数

```go
// repository/datascope.go
func ApplyDataScope(query *gorm.DB, scope *xcontext.DataScope, deptCol string, creatorCol string) *gorm.DB {
    if scope == nil {
        return query // 超管或不适用数据权限
    }
    switch scope.Scope {
    case constants.DataScopeAll:
        // 不加任何过滤
    case constants.DataScopeCustom:
        allDepts := append(scope.DeptIDs, scope.CustomDepts...)
        query = query.Where(deptCol+" IN ?", unique(allDepts))
    case constants.DataScopeDept:
        query = query.Where(deptCol+" IN ?", scope.DeptIDs)
    case constants.DataScopeDeptAndSub:
        query = query.Where(deptCol+" IN ?", scope.DeptIDs)
    case constants.DataScopeSelf:
        query = query.Where(creatorCol+" = ?", scope.UserID)
    }
    return query
}
```

### 6. Repository 中使用

```go
// user_repo.go 的 List 方法
func (r *UserRepo) List(ctx context.Context, req *dto.UserListRequest) ([]*model.User, int64, error) {
    query := r.q.User.WithContext(ctx).
        Where(r.q.User.TenantID.Eq(xcontext.GetTenantID(ctx))).
        Where(r.q.User.DeletedAt.Eq(0))

    // 应用数据权限
    scope := xcontext.GetDataScope(ctx)
    query = ApplyDataScope(query, scope, "department_id", "creator_id")

    // ... 其他过滤条件
}
```

### 7. 多角色合并策略

```
取最大权限（值越小权限越大）：
  1(全部) > 2(自定义) > 4(部门+子) > 3(本部门) > 5(仅本人)

多个 custom 角色时，合并部门列表并去重。
超管角色直接跳过数据权限。
```

### 8. 角色管理 UI 变更

角色创建/编辑时增加数据权限选项：

```
角色编辑页面新增：
  数据权限范围：[下拉框：全部/自定义/本部门/本部门及下级/仅本人]
  自定义部门：  [多选部门树，仅 data_scope=2 时显示]
```

## 验收标准

- [ ] 角色 data_scope 字段可正确读写
- [ ] data_scope=1（全部）→ 查询无额外过滤
- [ ] data_scope=3（本部门）→ `WHERE dept_id = ?`
- [ ] data_scope=4（本部门及下级）→ `WHERE dept_id IN (展开的子部门)`
- [ ] data_scope=5（仅本人）→ `WHERE creator_id = ?`
- [ ] data_scope=2（自定义）→ `WHERE dept_id IN (JSONB 解析)`
- [ ] 超管不受数据权限限制
- [ ] 多角色合并取最大权限
- [ ] PermissionCache 正确加载 data_scope
- [ ] 前端角色编辑页面可选择数据权限

## AI 协作提示

```
请按 step-12-data-permission.md 实现数据权限。
参考 docs/rbac-analysis/06-data-permission-rbac.md 的设计分析。
关键点：
1. data_scope 是角色的属性（静态），不是 ABAC
2. ApplyDataScope 是通用函数，所有需要数据权限的 Repository 调用它
3. 多角色合并取最大权限
4. 超管完全跳过数据权限检查
5. 不需要新表，只需 roles 表的 data_scope 和 data_scope_dept_ids 字段
```
