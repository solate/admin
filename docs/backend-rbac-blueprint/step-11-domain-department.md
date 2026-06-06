# Step 11: 域 — 部门管理（树形结构）

## 目标

实现部门管理的树形 CRUD：支持无限层级、拖拽排序、部门树查询。

## 前置条件

- Step 09 用户管理完成

## 文件清单

```
internal/
├── handler/department/
│   ├── department.go
│   ├── create.go
│   ├── update.go
│   ├── delete.go
│   └── query.go               # 含树形查询
├── service/department/
│   ├── department.go
│   ├── create.go
│   ├── update.go
│   ├── delete.go
│   ├── query.go               # 含树形构建
│   └── converter.go           # 含树形转换
├── repository/
│   └── department_repo.go     # 含 GetDescendantIDs
└── dto/
    └── department_dto.go      # 含 DepartmentTree
```

## 实现细节

### 1. 树形结构设计

```
departments 表：
  department_id  VARCHAR(20)  PK
  tenant_id      VARCHAR(20)  租户隔离
  parent_id      VARCHAR(20)  父部门（空=根节点）
  department_name VARCHAR(100)
  sort           INT          排序
  leader         VARCHAR(50)  负责人（可选）
  status         SMALLINT     状态
```

**树形查询方式**：一次性查出当前租户所有部门（WHERE tenant_id = ? AND deleted_at = 0），在内存中构建树。

> 不用递归 CTE 查树——部门数量通常 < 100，全量查出后内存建树更高效。

### 2. 内存建树算法

```go
// service/department/converter.go
func BuildTree(depts []*model.Department) []*dto.DepartmentTree {
    // 1. 创建 map[departmentID]*dto.DepartmentTree
    // 2. 遍历，每个节点初始化 Children = []
    // 3. 再遍历，如果 parent_id != ""，加入父节点的 Children
    // 4. 收集所有根节点（parent_id == "" 或 parent_id == "0"）
    // 5. 按 sort 排序
    return roots
}
```

### 3. 部门更新（含移动）

```
PUT /departments/:id
Body: { "name": "xxx", "parent_id": "new_parent_id", "sort": 1 }

校验：
1. 不能把自己设为自己的子部门（循环检测）
2. 新 parent_id 必须是同一租户的部门
3. parent_id 不能是自己下面的子孙部门

循环检测用 GetDescendantIDs：
  如果 new_parent_id 在当前部门的子孙列表中 → 拒绝
```

### 4. 部门删除

```
DELETE /departments/:id

校验：
1. 不能有子部门
2. 不能有用户（CountByDept）
3. 两个条件都满足 → 软删除
```

### 5. 部门下拉列表（不分页）

```
GET /departments/tree    → 返回完整树形结构
GET /departments/list    → 返回扁平列表（用于下拉选择）
```

> 部门数据量小（通常 < 100），不分页，全量返回。

### 6. Repository

```go
// department_repo.go
func (r *DeptRepo) Create(ctx context.Context, dept *model.Department) error
func (r *DeptRepo) Update(ctx context.Context, id string, updates map[string]interface{}) error
func (r *DeptRepo) Delete(ctx context.Context, id string) error
func (r *DeptRepo) GetByID(ctx context.Context, id string) (*model.Department, error)
func (r *DeptRepo) GetAll(ctx context.Context) ([]*model.Department, error)  // 当前租户所有部门
func (r *DeptRepo) GetDescendantIDs(ctx context.Context, deptID string) ([]string, error)
func (r *DeptRepo) CountChildren(ctx context.Context, parentID string) (int64, error)
func (r *DeptRepo) CountUsers(ctx context.Context, deptID string) (int64, error)
```

**GetDescendantIDs**：递归查询所有子孙部门 ID，用于循环检测和后续数据权限。

## 验收标准

- [ ] `POST /departments` 创建部门成功
- [ ] `GET /departments/tree` 返回正确的树形结构
- [ ] `GET /departments/list` 返回扁平列表
- [ ] `PUT /departments/:id` 更新部门成功
- [ ] `PUT /departments/:id` 把自己设为子部门 → 返回 409
- [ ] `DELETE /departments/:id` 有子部门 → 返回 409
- [ ] `DELETE /departments/:id` 有用户 → 返回 409
- [ ] 不同租户的部门互不可见
- [ ] 树形结构层级正确（支持 3+ 层）

## AI 协作提示

```
请按 step-11-domain-department.md 实现部门管理。
关键点：
1. 树形结构在内存中构建（一次全量查询，不递归 CTE）
2. 更新时检查循环引用（GetDescendantIDs）
3. 删除时检查子部门和用户
4. 部门不分页，全量返回
5. 租户隔离（所有查询 WHERE tenant_id = ?）
```
