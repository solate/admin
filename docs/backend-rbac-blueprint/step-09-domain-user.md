# Step 09: 域 — 用户管理 CRUD

## 目标

实现用户管理的完整 CRUD + 角色分配 + 个人信息修改。

## 前置条件

- Step 08 完成，租户 CRUD 模式已建立

## 文件清单

```
internal/
├── handler/user/
│   ├── user.go                # Handler + RegisterRoutes
│   ├── create.go              # POST /users（管理员创建）
│   ├── update.go              # PUT /users/:id
│   ├── delete.go              # DELETE /users/:id（批量）
│   ├── query.go               # GET /users + GET /users/:id
│   ├── profile.go             # GET/PUT /users/profile（个人信息）
│   └── role.go                # POST /users/:id/roles（分配角色）
├── service/user/
│   ├── user.go                # Service struct
│   ├── create.go
│   ├── update.go
│   ├── delete.go
│   ├── query.go
│   ├── profile.go
│   ├── role.go                # 角色分配逻辑
│   └── converter.go           # modelToUserInfo
├── repository/
│   ├── user_repo.go           # 用户查询
│   └── user_role_repo.go      # 用户角色关联
└── dto/
    └── user_dto.go
```

## 实现细节

### 1. 用户创建流程

```
POST /users
Body: CreateUserRequest

Service 逻辑：
1. 校验 email 全局唯一（跨租户）
2. 校验 phone 全局唯一（如果提供了）
3. 生成 user_id（idgen）
4. 密码 bcrypt 加密
5. 从 xcontext 获取 tenant_id，赋值给 user
6. 插入 users 表
7. 如果指定了角色，插入 user_roles 表
```

### 2. 用户列表（租户隔离）

```
GET /users?page=1&page_size=10&keyword=xxx&status=1&department_id=xxx

Repository 逻辑：
1. WHERE tenant_id = xcontext.GetTenantID(ctx)  ← 租户隔离
2. AND deleted_at = 0
3. AND (user_name LIKE ? OR email LIKE ?)       ← keyword 搜索
4. AND status = ?                                ← 可选过滤
5. AND department_id = ?                         ← 可选过滤
6. ORDER BY created_at DESC, user_id DESC        ← 双排序
7. OFFSET/LIMIT 分页
8. 返回 (list, count)
```

### 3. 角色分配

```
POST /users/:id/roles
Body: { "role_codes": ["admin", "auditor"] }

Service 逻辑：
1. 从 xcontext 获取 tenant_id
2. 校验角色 codes 存在于当前租户
3. 删除旧的 user_roles（WHERE user_id = ? AND tenant_id = ?）
4. 插入新的 user_roles
5. 事务保证原子性（db.Transaction）
```

### 4. 个人信息修改

```
GET /users/profile    → 从 JWT 获取 userID，返回当前用户信息
PUT /users/profile    → 修改 user_name, avatar, phone
POST /users/password  → 修改密码（需验证旧密码）
```

### 5. Repository 关键方法

```go
// user_repo.go
func (r *UserRepo) Create(ctx context.Context, user *model.User) error
func (r *UserRepo) Update(ctx context.Context, id string, updates map[string]interface{}) error
func (r *UserRepo) Delete(ctx context.Context, ids []string) error  // 批量软删除
func (r *UserRepo) GetByID(ctx context.Context, id string) (*model.User, error)
func (r *UserRepo) List(ctx context.Context, req *dto.UserListRequest) ([]*model.User, int64, error)
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error)
func (r *UserRepo) CountByDept(ctx context.Context, deptID string) (int64, error)

// user_role_repo.go
func (r *UserRoleRepo) GetUserRoleIDs(ctx context.Context, userID, tenantID string) ([]string, error)
func (r *UserRoleRepo) AssignRoles(ctx context.Context, userID string, roleIDs []string, tenantID string) error
func (r *UserRoleRepo) DeleteByUser(ctx context.Context, userID, tenantID string) error
```

### 6. 租户隔离规则

```
所有 Repository 方法（除了 GetByEmail 和 GetByPhone）都必须加：
  WHERE tenant_id = xcontext.GetTenantID(ctx)

例外（跨租户查询）：
  - GetByEmail（登录时用，此时还没有租户上下文）
  - GetByPhone（同上）
  - CountByDept（检查时已明确 deptID）
```

## 验收标准

- [ ] `POST /users` 创建用户成功，密码正确加密
- [ ] `POST /users` email 重复 → 返回 409
- [ ] `GET /users` 返回当前租户的用户列表
- [ ] `GET /users` 不同租户的用户互不可见
- [ ] `GET /users/:id` 返回用户详情（含角色信息）
- [ ] `PUT /users/:id` 更新用户信息成功
- [ ] `DELETE /users/:id` 软删除成功
- [ ] `POST /users/:id/roles` 角色分配成功
- [ ] `GET /users/profile` 返回当前用户信息
- [ ] `PUT /users/profile` 修改个人信息成功
- [ ] 列表双排序、分页正确

## AI 协作提示

```
请按 step-09-domain-user.md 实现用户管理。
遵循 step-08 建立的域模式（handler/service/repository/dto）。
重点：
1. 用户创建时自动赋值 tenant_id（从 xcontext）
2. 所有查询（除 GetByEmail）必须加租户隔离
3. 角色分配用事务（db.Transaction）
4. Converter 接收 []*model.User，内部构建 map
5. 密码用 bcrypt，不存储明文
```
