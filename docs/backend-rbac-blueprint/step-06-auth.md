# Step 06: 认证 — 登录 / Token 刷新 / 租户切换

## 目标

实现完整的认证流程：邮箱/手机登录、JWT 签发、Token 刷新、租户切换、登出。

## 前置条件

- Step 05 完成，JWT 和 xcontext 可用

## 文件清单

```
internal/
├── handler/auth/
│   ├── auth.go                # Handler struct + 构造函数
│   ├── login.go               # POST /auth/login
│   ├── refresh.go             # POST /auth/refresh
│   ├── logout.go              # POST /auth/logout
│   └── switch_tenant.go       # POST /auth/switch-tenant
├── service/auth/
│   ├── auth.go                # Service struct + 构造函数
│   ├── login.go               # 登录逻辑
│   ├── refresh.go             # 刷新逻辑 + 租户切换
│   └── logout.go              # 登出逻辑
├── repository/
│   ├── user_repo.go           # 用户查询（登录时用）
│   ├── user_role_repo.go      # 用户角色查询
│   ├── role_repo.go           # 角色查询
│   └── tenant_repo.go         # 租户查询
└── dto/
    └── auth_dto.go            # LoginRequest, TokenResponse, etc.
```

## 实现细节

### 1. 登录流程

```
POST /api/v1/auth/login
Body: { "email": "admin@test.com", "password": "xxx" }

流程：
1. 查用户（WHERE email = ? AND deleted_at = 0）—— 跨租户查询
2. 验证密码（bcrypt）
3. 查用户状态（status == 1）
4. 查用户的租户（tenant_id → tenants 表）—— 验证租户状态
5. 查用户在租户中的角色（user_roles WHERE user_id = ? AND tenant_id = ?）
6. 如果没有角色 → 返回 ErrUserNoRoles
7. 查角色详情（获取 role_code）
8. GenerateTokenPair(tenantID, tenantCode, userID, userName, roleCodes, roleIDs)
9. 存储 Refresh Token 到 Redis（user:sessions:{userID} SET，用于多设备管理）
10. 返回 { access_token, refresh_token, expires_in }
```

**关键决策**：
- email 全局唯一（跨租户），所以登录时先查用户再取租户
- 用户可能属于多个租户，登录时用主租户（users.tenant_id）
- 角色查询必须带 `tenant_id`，确保只查当前租户的角色

### 2. Token 刷新

```
POST /api/v1/auth/refresh
Body: { "refresh_token": "xxx" }

流程：
1. VerifyRefreshToken
2. 检查是否在黑名单
3. 生成新的 Token Pair（相同 Claims，新 TokenID）
4. 旧 Refresh Token 加入黑名单
5. 更新 Redis 中的 session 记录
6. 返回新的 Token Pair
```

### 3. 租户切换

```
POST /api/v1/auth/switch-tenant
Body: { "tenant_id": "target_tenant_id" }

流程：
1. 超管（HasRole("super_admin")）→ 可以切换到任意租户
2. 普通用户 → 检查 user_roles 中是否有该租户的角色
3. 查询目标租户的角色
4. 生成新的 Token Pair（目标租户的 tenantID + 新角色）
5. 返回新的 Token Pair
```

### 4. 登出

```
POST /api/v1/auth/logout

流程：
1. 提取当前 TokenID
2. Access Token 加入黑名单
3. Refresh Token 加入黑名单
4. 从 Redis session SET 中移除
```

### 5. Repository 层关键方法

```go
// user_repo.go
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error)
// 不加 tenant_id 条件（email 全局唯一，跨租户查询）

// user_role_repo.go
func (r *UserRoleRepo) GetUserRoleIDs(ctx context.Context, userID, tenantID string) ([]string, error)
// WHERE user_id = ? AND tenant_id = ?

// role_repo.go
func (r *RoleRepo) GetByIDs(ctx context.Context, ids []string) ([]*model.Role, error)
// WHERE role_id IN (?) —— 不加 tenant_id（角色 ID 全局唯一）

// tenant_repo.go
func (r *TenantRepo) GetByID(ctx context.Context, id string) (*model.Tenant, error)
```

### 6. DTO

```go
type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

type TokenResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int64  `json:"expires_in"`
    TokenType    string `json:"token_type"`
}

type RefreshRequest struct {
    RefreshToken string `json:"refresh_token" binding:"required"`
}

type SwitchTenantRequest struct {
    TenantID string `json:"tenant_id" binding:"required"`
}
```

## 验收标准

- [ ] `POST /auth/login` 正确邮箱+密码 → 返回 200 + Token Pair
- [ ] 错误密码 → 返回 401
- [ ] 用户被禁用 → 返回 403
- [ ] 租户被禁用 → 返回 403
- [ ] 用户无角色 → 返回 403 + 特定错误码
- [ ] `POST /auth/refresh` 有效 refresh_token → 返回新 Token Pair
- [ ] `POST /auth/refresh` 过期 token → 返回 401
- [ ] `POST /auth/logout` → 旧 Token 立即失效
- [ ] `POST /auth/switch-tenant` 超管 → 可以切换到任意租户
- [ ] `POST /auth/switch-tenant` 普通用户 → 只能切换有角色的租户
- [ ] 登录后用 access_token 调用 `/auth/logout` → 成功
- [ ] 登出后再用该 token 调接口 → 返回 401

## AI 协作提示

```
请按 step-06-auth.md 实现认证功能。
这是第一个完整的域（auth），包含 handler → service → repository 全链路。
注意以下要点：
1. 域子包组织：handler/auth/, service/auth/
2. 方法单文件：login.go, refresh.go, logout.go 各一个文件
3. Repository 集中式：repository/ 下扁平文件
4. 登录流程的完整步骤（查用户→验证→查角色→签发）
5. 租户切换的超管绕行逻辑
先写 repository → service → handler，从底向上。
```
