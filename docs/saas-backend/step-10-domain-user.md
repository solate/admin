# Step 10: 用户管理域

## 目标

实现用户的 CRUD、角色分配、密码重置、状态管理。用户是多租户系统的核心实体，所有操作带租户隔离。

## 前置条件

- Step 09 完成，租户管理可用

## 文件清单

```
internal/
├── handler/user/
│   ├── handler.go
│   ├── create.go
│   ├── update.go
│   ├── delete.go
│   ├── query.go               # List + GetByID
│   ├── status.go              # 启用/禁用
│   ├── password.go            # 重置密码
│   ├── role.go                # 分配角色
│   └── dto.go
├── service/user/
│   ├── service.go
│   ├── create.go
│   ├── update.go
│   ├── delete.go
│   ├── query.go
│   ├── status.go
│   ├── password.go
│   ├── role.go
│   └── converter.go
├── repository/
│   ├── user_repo.go           # 补充完整 CRUD
│   └── user_role_repo.go      # 用户-角色关联
```

## 实现规范

### 1. API 设计

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/users | 用户列表（分页，租户隔离） |
| GET | /api/v1/users/:id | 用户详情 |
| POST | /api/v1/users | 创建用户 |
| PUT | /api/v1/users/:id | 更新用户信息 |
| DELETE | /api/v1/users | 批量删除用户（软删除） |
| PUT | /api/v1/users/status | 批量启用/禁用 |
| PUT | /api/v1/users/:id/password | 重置密码 |
| PUT | /api/v1/users/:id/roles | 分配角色 |

### 2. 租户隔离规则

```go
// 所有用户查询带 tenant_id
func (r *UserRepo) List(ctx context.Context, ...) {
    tenantID := xcontext.GetTenantID(ctx)
    query.Where(u.TenantID.Eq(tenantID))
}

// 创建用户自动赋 tenant_id
func (r *UserRepo) Create(ctx context.Context, user *model.User) {
    user.TenantID = xcontext.GetTenantID(ctx)
}
```

### 3. DTO

```go
type CreateUserRequest struct {
    Email        string   `json:"email" binding:"required,email"`
    UserName     string   `json:"user_name" binding:"required,min=2,max=50"`
    RealName     string   `json:"real_name"`
    Phone        string   `json:"phone"`
    Password     string   `json:"password" binding:"required,min=6"`
    DepartmentID string   `json:"department_id"`
    PositionID   string   `json:"position_id"`
    RoleIDs      []string `json:"role_ids"` // 创建时可直接分配角色
}

type UpdateUserRequest struct {
    UserName     string `json:"user_name" binding:"omitempty,min=2,max=50"`
    RealName     string `json:"real_name"`
    Phone        string `json:"phone"`
    Avatar       string `json:"avatar"`
    DepartmentID string `json:"department_id"`
    PositionID   string `json:"position_id"`
}

type ResetPasswordRequest struct {
    Password string `json:"password" binding:"required,min=6"`
}

type AssignRolesRequest struct {
    RoleIDs []string `json:"role_ids" binding:"required"`
}

type UserInfo struct {
    UserID       string     `json:"user_id"`
    Email        string     `json:"email"`
    UserName     string     `json:"user_name"`
    RealName     string     `json:"real_name"`
    Phone        string     `json:"phone"`
    Avatar       string     `json:"avatar"`
    Status       int        `json:"status"`
    DepartmentID string     `json:"department_id"`
    Department   string     `json:"department"`   // 部门名称
    PositionID   string     `json:"position_id"`
    Position     string     `json:"position"`     // 岗位名称
    Roles        []RoleItem `json:"roles"`
    LastLoginAt  int64      `json:"last_login_at"`
    CreatedAt    int64      `json:"created_at"`
}

type RoleItem struct {
    RoleID   string `json:"role_id"`
    RoleCode string `json:"role_code"`
    Name     string `json:"name"`
}

// 列表请求（支持搜索和筛选）
type ListUserRequest struct {
    dto.PageRequest
    Keyword      string `form:"keyword"`       // 模糊搜索：user_name / email / phone
    Status       int    `form:"status"`        // 1=启用 2=禁用
    DepartmentID string `form:"department_id"` // 按部门筛选
}
```

### 4. 关键业务规则

| 规则 | 说明 |
|------|------|
| email 唯一 | 同一租户内 email 唯一（partial index WHERE deleted_at=0） |
| 创建用户 | 密码 bcrypt 加密后存储 |
| 批量删除 | 不能删除自己；验证 IDs 都属于当前租户 |
| 重置密码 | 管理员重置后设置 must_change_password=1 |
| 分配角色 | 全量替换：先删 user_roles → 再插入新的 |
| 列表查询 | 关联查 department.name + position.name + roles |
| 用户数限制 | 创建前检查：当前租户用户数 < tenant.max_users |

### 5. 避免 N+1 查询

```go
// service/user/query.go
func (s *Service) List(ctx context.Context, req *ListUserRequest) ([]UserInfo, int64, error) {
    // 1. 分页查 users 列表
    users, total, err := s.userRepo.List(ctx, ...)

    // 2. 批量查关联数据（避免 N+1）
    userIDs := extractUserIDs(users)
    deptIDs := extractDeptIDs(users)
    posIDs := extractPositionIDs(users)

    // 3. 批量查 user_roles
    userRolesMap := s.userRoleRepo.BatchGetByUserIDs(ctx, userIDs)

    // 4. 批量查 departments
    deptMap := s.deptRepo.BatchGetByIDs(ctx, deptIDs)

    // 5. 批量查 positions
    posMap := s.posRepo.BatchGetByIDs(ctx, posIDs)

    // 6. converter 组装（接收 list，内部构建 map）
    return convertToUserInfoList(users, userRolesMap, deptMap, posMap), total, nil
}
```

### 6. Converter 规范

```go
// service/user/converter.go
package user

// converter 接收 list，不接收 map
// 内部自行构建 map
func convertToUserInfoList(
    users []*model.User,
    roles []*model.UserRole,   // 注意：传 list
    depts []*model.Department,
    positions []*model.Position,
) []UserInfo {
    // 内部构建 map
    roleMap := buildRoleMap(roles)
    deptMap := buildDeptMap(depts)
    posMap := buildPositionMap(positions)
    // 组装 UserInfo
    ...
}
```

## 验收标准

```bash
# 1. 创建用户
curl -s -X POST http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"email":"test@acme.com","user_name":"测试用户","password":"Test123456","role_ids":["ROLE_ID"]}'
# 期望：{"code":0,"data":{"user_id":"..."}}

# 2. 用户列表（带搜索）
curl -s "http://localhost:8080/api/v1/users?page=1&page_size=10&keyword=test" \
  -H "Authorization: Bearer $TOKEN"
# 期望：分页返回，包含角色/部门/岗位信息

# 3. 租户隔离
# 租户 A 的 Token 看不到租户 B 的用户

# 4. 批量删除
curl -s -X DELETE http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"ids":["USER_ID"]}'
# 期望：成功（软删除）

# 5. 用户数限制
# 超过 tenant.max_users 时创建用户 → {"code":40900,"message":"用户数已达上限"}

# 6. 分配角色
curl -s -X PUT http://localhost:8080/api/v1/users/USER_ID/roles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"role_ids":["ROLE_1","ROLE_2"]}'
# 期望：成功
```

## AI 协作提示

```
请按 step-10-domain-user.md 实现用户管理域。

要点：
1. 完整 handler/service/repository 三层 + converter
2. 所有查询带租户隔离 WHERE tenant_id = xcontext.GetTenantID(ctx)
3. 列表查询避免 N+1：批量查 roles/departments/positions
4. converter 接收 list，内部构建 map
5. 创建用户检查 email 唯一 + 用户数上限
6. 分配角色全量替换
7. 密码 bcrypt 加密
8. 双排序字段
9. 列表支持 keyword / status / department_id 筛选
```

---

*上一步：[Step 09 - 租户管理](step-09-domain-tenant.md) | 下一步：[Step 11 - 角色管理](step-11-domain-role.md)*
