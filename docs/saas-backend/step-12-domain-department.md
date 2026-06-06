# Step 12: 部门管理域

## 目标

实现部门树形结构的 CRUD 管理。部门是组织架构的基础，用于数据权限和用户归属。

## 前置条件

- Step 10 完成（用户有 department_id 字段）

## 文件清单

```
internal/
├── handler/department/
│   ├── handler.go
│   ├── create.go
│   ├── update.go
│   ├── delete.go
│   ├── query.go               # Tree + GetByID
│   └── dto.go
├── service/department/
│   ├── service.go
│   ├── create.go
│   ├── update.go
│   ├── delete.go
│   ├── query.go
│   └── converter.go
├── repository/
│   └── department_repo.go
```

## 实现规范

### 1. API 设计

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/departments/tree | 部门树（完整树形） |
| GET | /api/v1/departments/:id | 部门详情 |
| POST | /api/v1/departments | 创建部门 |
| PUT | /api/v1/departments/:id | 更新部门 |
| DELETE | /api/v1/departments/:id | 删除部门 |

### 2. 树形数据处理

```go
// 树形构建在 Service 层完成（非 SQL 递归）
// 1. 一次性查出当前租户所有部门（通常不超过几百条）
// 2. 在内存中构建树

func buildTree(departments []*model.Department) []*DepartmentTree {
    // 构建 parentID → children 映射
    childMap := make(map[string][]*DepartmentTree)
    nodeMap := make(map[string]*DepartmentTree)

    for _, d := range departments {
        node := convertToTree(d)
        nodeMap[d.DepartmentID] = node
        childMap[d.ParentID] = append(childMap[d.ParentID], node)
    }

    // 根节点：parent_id = ""
    roots := childMap[""]

    // 递归填充 children
    var fillChildren func(nodes []*DepartmentTree)
    fillChildren = func(nodes []*DepartmentTree) {
        for _, n := range nodes {
            n.Children = childMap[n.DepartmentID]
            if n.Children == nil {
                n.Children = []*DepartmentTree{} // 空数组而非 nil
            }
            fillChildren(n.Children)
        }
    }
    fillChildren(roots)

    return roots
}
```

### 3. DTO

```go
type CreateDepartmentRequest struct {
    ParentID string `json:"parent_id"` // 空=顶级部门
    Name     string `json:"name" binding:"required,max=100"`
    LeaderID string `json:"leader_id"` // 部门负责人 user_id
    Sort     int    `json:"sort"`
}

type UpdateDepartmentRequest struct {
    ParentID string `json:"parent_id"`
    Name     string `json:"name" binding:"omitempty,max=100"`
    LeaderID string `json:"leader_id"`
    Sort     int    `json:"sort"`
    Status   int    `json:"status" binding:"omitempty,oneof=1 2"`
}

type DepartmentTree struct {
    DepartmentID string            `json:"department_id"`
    ParentID     string            `json:"parent_id"`
    Name         string            `json:"name"`
    LeaderID     string            `json:"leader_id"`
    LeaderName   string            `json:"leader_name"`
    Sort         int               `json:"sort"`
    Status       int               `json:"status"`
    UserCount    int64             `json:"user_count"`
    Children     []*DepartmentTree `json:"children"`
}
```

### 4. 业务规则

| 规则 | 说明 |
|------|------|
| 树深度限制 | 最多 5 级（创建时检查） |
| 删除检查 | 有子部门或有关联用户时不允许删除 |
| 移动部门 | 修改 parent_id 时检测循环 |
| 禁用传播 | 禁用父部门时，子部门也禁用 |
| 排序 | 同级内按 sort ASC 排序 |

### 5. 循环检测

```go
func (s *Service) checkCircular(ctx context.Context, deptID, newParentID string) error {
    if newParentID == "" || deptID == newParentID {
        if deptID == newParentID {
            return xerr.New(xerr.CodeParamInvalid, "不能将自己设为父部门")
        }
        return nil
    }
    // 查 newParentID 的所有祖先，确保 deptID 不在其中
    // ...
}
```

## 验收标准

```bash
# 1. 创建部门（顶级）
curl -s -X POST http://localhost:8080/api/v1/departments \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"技术部","sort":1}'
# 期望：成功

# 2. 创建子部门
curl -s -X POST http://localhost:8080/api/v1/departments \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"后端组","parent_id":"PARENT_ID","sort":1}'
# 期望：成功

# 3. 获取部门树
curl -s http://localhost:8080/api/v1/departments/tree \
  -H "Authorization: Bearer $TOKEN"
# 期望：返回嵌套树形结构，children 为空数组而非 null

# 4. 删除检查
# 有子部门的部门 → 不允许删除
# 有用户的部门 → 不允许删除

# 5. 租户隔离
# 不同租户看到不同的部门树
```

## AI 协作提示

```
请按 step-12-domain-department.md 实现部门管理域。

要点：
1. 完整三层
2. 树形构建在 Service 层内存处理（不用 SQL 递归）
3. 一次性查全部部门 → 内存构建树
4. children 为空时返回 []（不是 null）
5. 删除前检查子部门 + 关联用户
6. 循环检测
7. 排序：同级内 sort ASC, department_id ASC
8. 租户隔离
```

---

*上一步：[Step 11 - 角色管理](step-11-domain-role.md) | 下一步：[Step 13 - 菜单管理](step-13-domain-menu.md)*
