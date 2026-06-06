# Step 11: 角色管理域

## 目标

实现角色的 CRUD、权限分配、菜单分配、角色继承管理。

## 前置条件

- Step 10 完成

## 文件清单

```
internal/
├── handler/role/
│   ├── handler.go
│   ├── create.go
│   ├── update.go
│   ├── delete.go
│   ├── query.go
│   ├── permission.go          # 分配/查询权限
│   ├── menu.go                # 分配/查询菜单
│   └── dto.go
├── service/role/
│   ├── service.go
│   ├── create.go
│   ├── update.go
│   ├── delete.go
│   ├── query.go
│   ├── permission.go
│   ├── menu.go
│   └── converter.go
├── repository/
│   ├── role_repo.go           # 补充完整
│   ├── role_permission_repo.go
│   └── role_menu_repo.go
```

## 实现规范

### 1. API 设计

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/roles | 角色列表（租户隔离） |
| GET | /api/v1/roles/:id | 角色详情 |
| POST | /api/v1/roles | 创建角色 |
| PUT | /api/v1/roles/:id | 更新角色 |
| DELETE | /api/v1/roles | 批量删除角色 |
| PUT | /api/v1/roles/:id/status | 启用/禁用 |
| PUT | /api/v1/roles/:id/permissions | 分配权限（全量替换） |
| GET | /api/v1/roles/:id/permissions | 获取角色权限 ID 列表 |
| PUT | /api/v1/roles/:id/menus | 分配菜单（全量替换） |
| GET | /api/v1/roles/:id/menus | 获取角色菜单 ID 列表 |

### 2. 业务规则

| 规则 | 说明 |
|------|------|
| role_code 唯一 | 同一租户内唯一（partial index） |
| super_admin | 系统预置，不可修改/删除 |
| 角色继承 | parent_role_id 指向父角色，子角色继承父角色所有权限 |
| 循环检测 | 设置 parent_role_id 时检测循环（递归深度 < 10） |
| 删除检查 | 有用户关联的角色不允许删除 |
| 权限变更后 | 调用 rbacCache.NotifyRefresh() |

### 3. DTO

```go
type CreateRoleRequest struct {
    RoleCode     string `json:"role_code" binding:"required,min=2,max=50"`
    Name         string `json:"name" binding:"required,max=100"`
    Description  string `json:"description"`
    ParentRoleID string `json:"parent_role_id"`
    Sort         int    `json:"sort"`
}

type UpdateRoleRequest struct {
    Name         string `json:"name" binding:"omitempty,max=100"`
    Description  string `json:"description"`
    ParentRoleID string `json:"parent_role_id"`
    Sort         int    `json:"sort"`
    DataScope    int    `json:"data_scope" binding:"omitempty,oneof=1 2 3 4 5"`
    DataScopeDeptIDs []string `json:"data_scope_dept_ids"`
}

type RoleInfo struct {
    RoleID       string `json:"role_id"`
    RoleCode     string `json:"role_code"`
    Name         string `json:"name"`
    Description  string `json:"description"`
    ParentRoleID string `json:"parent_role_id"`
    DataScope    int    `json:"data_scope"`
    Sort         int    `json:"sort"`
    Status       int    `json:"status"`
    UserCount    int64  `json:"user_count"` // 关联用户数
    CreatedAt    int64  `json:"created_at"`
}
```

### 4. 循环检测

```go
// service/role/update.go
func (s *Service) checkCircularInheritance(ctx context.Context, roleID, parentRoleID string) error {
    if parentRoleID == "" {
        return nil
    }
    if roleID == parentRoleID {
        return xerr.New(xerr.CodeParamInvalid, "不能将自己设为父角色")
    }

    // 向上追溯 parent 链，最多 10 层
    visited := map[string]bool{roleID: true}
    current := parentRoleID
    for i := 0; i < 10; i++ {
        if visited[current] {
            return xerr.New(xerr.CodeParamInvalid, "角色继承存在循环")
        }
        visited[current] = true
        role, err := s.roleRepo.GetByID(ctx, current)
        if err != nil || role.ParentRoleID == "" {
            break
        }
        current = role.ParentRoleID
    }
    return nil
}
```

### 5. 权限/菜单分配

```go
// 全量替换模式
func (s *Service) AssignPermissions(ctx context.Context, roleID string, permIDs []string) error {
    tenantID := xcontext.GetTenantID(ctx)

    // 事务内操作
    err := s.db.Transaction(func(tx *gorm.DB) error {
        // 1. 删除旧的
        if err := s.rolePermRepo.DeleteByRole(ctx, tx, roleID, tenantID); err != nil {
            return err
        }
        // 2. 插入新的
        if err := s.rolePermRepo.BatchCreate(ctx, tx, roleID, permIDs, tenantID); err != nil {
            return err
        }
        return nil
    })
    if err != nil {
        return xerr.Wrap(xerr.CodeInternal, "分配权限失败", err)
    }

    // 刷新缓存
    s.rbacCache.NotifyRefresh()
    return nil
}
```

## 验收标准

```bash
# 1. 创建角色
curl -s -X POST http://localhost:8080/api/v1/roles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"role_code":"editor","name":"编辑员"}'
# 期望：成功

# 2. 角色列表
curl -s "http://localhost:8080/api/v1/roles?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN"
# 期望：返回当前租户角色列表，含 user_count

# 3. role_code 唯一
# 同租户内重复 role_code → {"code":40901}

# 4. 角色继承
# 创建子角色 child → parent_role_id = editor
# 给 editor 分配权限 A → child 自动继承 A

# 5. 循环检测
# editor.parent = child → 报错"角色继承存在循环"

# 6. 删除检查
# 有用户关联的角色 → {"code":40900,"message":"角色已关联用户，无法删除"}

# 7. 权限分配后缓存刷新
# 分配后日志出现 "permission cache reloaded"
```

## AI 协作提示

```
请按 step-11-domain-role.md 实现角色管理域。

要点：
1. 完整三层 + converter
2. 租户隔离：角色按 tenant_id 隔离
3. 权限/菜单分配：事务内全量替换
4. 分配后 rbacCache.NotifyRefresh()
5. 继承循环检测（向上追溯最多 10 层）
6. 预置 super_admin 角色不可删除
7. 删除前检查用户关联
8. 列表返回 user_count
```

---

*上一步：[Step 10 - 用户管理](step-10-domain-user.md) | 下一步：[Step 12 - 部门管理](step-12-domain-department.md)*
