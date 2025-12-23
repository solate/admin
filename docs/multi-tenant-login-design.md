# 多租户登录设计方案

## 1. 设计概述

本方案实现了一个用户可以在多个租户中拥有不同角色的多租户 SaaS 平台登录系统。核心特点：

- **用户与租户解耦**：用户表不再包含 `tenant_id` 字段，用户名全局唯一
- **智能租户选择**：根据用户租户数量自动处理登录流程
- **记住上次选择**：自动使用用户上次选择的租户登录

## 2. 数据库设计

### 2.1 用户表 (users)

```sql
CREATE TABLE users (
    user_id VARCHAR(255) PRIMARY KEY,
    user_name VARCHAR(255) NOT NULL UNIQUE,  -- 全局唯一
    password VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL DEFAULT '',
    avatar VARCHAR(255),
    phone VARCHAR(20),
    email VARCHAR(255),
    status INTEGER NOT NULL DEFAULT 1,
    remark TEXT,
    last_login_time BIGINT,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0
);
```

**关键变更**：
- 移除了 `tenant_id` 字段
- 移除了 `role_type` 字段
- `user_name` 改为全局唯一

### 2.2 用户-租户-角色关联表 (user_tenant_role)

```sql
CREATE TABLE user_tenant_role (
    user_tenant_role_id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(36) NOT NULL,
    role_id VARCHAR(255) NOT NULL,
    status INTEGER NOT NULL DEFAULT 1,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0,
    UNIQUE KEY uk_user_tenant_role (user_id, tenant_id, role_id)
);
```

**作用**：一个用户可以在不同租户中拥有不同的角色。

## 3. 登录流程

### 3.1 流程图

```
用户输入账号密码
       │
       ▼
验证账号密码
       │
       ▼
查询用户有权限的租户列表
       │
       ├─────────────┬─────────────┬─────────────┐
       ▼             ▼             ▼             ▼
  无租户权限    只有一个租户   多租户+有效    多租户+无效
       │             ▲          last_tenant_id  last_tenant_id
       ▼             │                │              │
   返回错误      自动进入        自动进入      显示选择界面
                  该租户         该租户
```

### 3.2 API 设计

#### 登录接口

**请求**：
```json
POST /api/v1/auth/login
{
  "username": "admin",
  "password": "password",
  "captcha_id": "xxx",
  "captcha": "1234",
  "last_tenant_id": "tenant-123"  // 可选，上次选择的租户ID
}
```

**响应 - 需要选择租户**：
```json
{
  "code": 0,
  "data": {
    "need_select_tenant": true,
    "user_id": "user-123",
    "tenants": [
      {
        "tenant_id": "tenant-1",
        "tenant_name": "公司A",
        "tenant_code": "company_a",
        "role_type": 2
      },
      {
        "tenant_id": "tenant-2",
        "tenant_name": "公司B",
        "tenant_code": "company_b",
        "role_type": 1
      }
    ]
  }
}
```

**响应 - 直接登录成功**：
```json
{
  "code": 0,
  "data": {
    "need_select_tenant": false,
    "access_token": "xxx",
    "refresh_token": "yyy",
    "expires_in": 3600,
    "user_id": "user-123",
    "current_tenant": {
      "tenant_id": "tenant-1",
      "tenant_name": "公司A",
      "tenant_code": "company_a",
      "role_type": 2
    },
    "phone": "13800000000",
    "email": "admin@example.com"
  }
}
```

#### 选择租户接口

**请求**：
```json
POST /api/v1/auth/select-tenant
Headers: {
  "X-User-ID": "user-123"
}
{
  "tenant_id": "tenant-2"
}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "access_token": "xxx",
    "refresh_token": "yyy",
    "expires_in": 3600,
    "current_tenant": {
      "tenant_id": "tenant-2",
      "tenant_name": "公司B",
      "tenant_code": "company_b",
      "role_type": 1
    }
  }
}
```

## 4. 核心代码实现

### 4.1 后端核心逻辑

#### DTO 定义 ([`backend/internal/dto/auth.go`](../backend/internal/dto/auth.go))

```go
// LoginRequest 登录请求
type LoginRequest struct {
    UserName     string `json:"username" binding:"required"`
    Password     string `json:"password" binding:"required"`
    CaptchaID    string `json:"captcha_id" binding:"required"`
    Captcha      string `json:"captcha" binding:"required"`
    LastTenantID string `json:"last_tenant_id" binding:"omitempty"` // 上次选择的租户ID
}

// TenantInfo 租户信息
type TenantInfo struct {
    TenantID   string `json:"tenant_id"`
    TenantName string `json:"tenant_name"`
    TenantCode string `json:"tenant_code"`
    RoleType   int32  `json:"role_type"`
}

// LoginResponse 登录响应
type LoginResponse struct {
    // 需要选择租户的情况
    NeedSelectTenant bool         `json:"need_select_tenant"`
    UserID          string       `json:"user_id"`
    Tenants         []TenantInfo `json:"tenants,omitempty"`

    // 直接登录成功的情况
    AccessToken    string       `json:"access_token,omitempty"`
    RefreshToken   string       `json:"refresh_token,omitempty"`
    ExpiresIn      int64        `json:"expires_in,omitempty"`
    CurrentTenant  *TenantInfo  `json:"current_tenant,omitempty"`
    Phone          string       `json:"phone,omitempty"`
    Email          string       `json:"email,omitempty"`
}
```

#### Service 层 ([`backend/internal/service/auth_service.go`](../backend/internal/service/auth_service.go))

```go
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
    // 1. 验证验证码
    // 2. 查询用户（全局查询）
    // 3. 验证密码
    // 4. 查询用户关联的所有租户
    // 5. 构建租户信息列表
    // 6. 智能选择租户逻辑：
    //    - 只有一个租户 → 直接使用
    //    - 多租户 + 有效 last_tenant_id → 使用上次选择的
    //    - 否则 → 返回租户列表让用户选择
}

func (s *AuthService) SelectTenant(ctx context.Context, userID, tenantID string) (*dto.SelectTenantResponse, error) {
    // 1. 验证用户是否有该租户权限
    // 2. 生成该租户的 JWT token
    // 3. 返回 token 和租户信息
}
```

### 4.2 前端核心逻辑

#### Token 管理 ([`frontend/src/utils/token.ts`](../frontend/src/utils/token.ts))

```typescript
// 保存 token 和租户信息
export function saveTokens(data: {
  access_token: string
  refresh_token: string
  user_id: string
  current_tenant?: TenantInfo
}) {
  localStorage.setItem('access_token', data.access_token)
  localStorage.setItem('refresh_token', data.refresh_token)
  localStorage.setItem('user_id', data.user_id)
  if (data.current_tenant) {
    localStorage.setItem('tenant_id', data.current_tenant.tenant_id)
    localStorage.setItem('tenant_info', JSON.stringify(data.current_tenant))
  }
}

// 获取上次选择的租户ID
export function getLastTenantId(): string | null {
  return localStorage.getItem('tenant_id')
}
```

#### 登录页面 ([`frontend/src/views/Login.vue`](../frontend/src/views/Login.vue))

```typescript
// 登录步骤状态
type LoginStep = 'credentials' | 'select_tenant'

// 登录提交
async function onSubmit() {
  const lastTenantId = getLastTenantId() || undefined

  const res = await authApi.login({
    username: form.value.username,
    password: form.value.password,
    captcha_id: captchaId.value,
    captcha: form.value.captcha,
    last_tenant_id: lastTenantId
  })

  // 需要选择租户
  if (res.need_select_tenant) {
    loginStep.value = 'select_tenant'
    pendingUserId.value = res.user_id
    availableTenants.value = res.tenants || []
    return
  }

  // 直接登录成功
  if (res.access_token && res.current_tenant) {
    saveTokens({ access_token: res.access_token, ... })
    router.push('/')
  }
}

// 选择租户
async function selectTenant(tenant: TenantInfo) {
  const res = await authApi.selectTenant(pendingUserId.value, { tenant_id: tenant.tenant_id })
  saveTokens({ access_token: res.access_token, ... })
  router.push('/')
}
```

## 5. 场景处理

| 场景 | 处理方式 | 用户体验 |
|------|----------|----------|
| 只有 1 个租户 | 直接返回 token | ⭐⭐⭐⭐⭐ 无感知 |
| 多租户 + 有有效 last_tenant_id | 自动进入上次租户 | ⭐⭐⭐⭐⭐ 无感知 |
| 多租户 + 无有效 last_tenant_id | 展示选择界面 | ⭐⭐⭐ 合理 |
| 无租户权限 | 返回错误 | - |

## 6. 安全性考虑

| 担心 | 解决方案 |
|------|----------|
| 租户列表泄露 | 只返回当前用户有权限的租户 |
| 扫描租户ID | 使用 UUID，增加猜解难度 |
| 跨租户访问 | Token 绑定 tenant_id，后端严格校验 |
| 选择租户时的身份验证 | 使用 X-User-ID header + 临时会话 |

## 7. 文件修改清单

### 后端

| 文件 | 修改内容 |
|------|----------|
| [`backend/scripts/dev_schema.sql`](../backend/scripts/dev_schema.sql) | 用户表移除 tenant_id、role_type |
| [`backend/internal/dto/auth.go`](../backend/internal/dto/auth.go) | 新增 DTO 定义 |
| [`backend/internal/dto/user.go`](../backend/internal/dto/user.go) | CreateUserRequest 新增 Name 字段 |
| [`backend/internal/service/auth_service.go`](../backend/internal/service/auth_service.go) | 实现智能登录逻辑 |
| [`backend/internal/service/user_service.go`](../backend/internal/service/user_service.go) | 适配新架构 |
| [`backend/internal/handler/auth_handler.go`](../backend/internal/handler/auth_handler.go) | 新增 SelectTenant 接口 |
| [`backend/internal/repository/user_repo.go`](../backend/internal/repository/user_repo.go) | 移除租户相关方法 |
| [`backend/internal/repository/tenant_repo.go`](../backend/internal/repository/tenant_repo.go) | 新增 GetByIDs 方法 |
| [`backend/internal/router/app.go`](../backend/internal/router/app.go) | 更新依赖注入 |
| [`backend/scripts/init_db/init_db.go`](../backend/scripts/init_db/init_db.go) | 适配初始化逻辑 |

### 前端

| 文件 | 修改内容 |
|------|----------|
| [`frontend/src/api/auth.ts`](../frontend/src/api/auth.ts) | 更新 API 类型定义 |
| [`frontend/src/utils/token.ts`](../frontend/src/utils/token.ts) | 支持租户信息存储 |
| [`frontend/src/views/Login.vue`](../frontend/src/views/Login.vue) | 实现租户选择界面 |

## 8. 部署步骤

1. **更新数据库 Schema**
   ```bash
   psql -U root -d admin -f backend/scripts/dev_schema.sql
   ```

2. **重新生成 Model**（如果使用 gorm-gen）
   ```bash
   cd backend && go generate ./...
   ```

3. **运行初始化脚本**
   ```bash
   go run scripts/init_db/init_db.go
   ```

4. **启动后端服务**
   ```bash
   cd backend && go run cmd/server/main.go
   ```

5. **启动前端服务**
   ```bash
   cd frontend && npm run dev
   ```

## 9. 测试验证

1. **单租户登录**：创建只有单个租户权限的用户，验证自动登录
2. **多租户登录（有记录）**：登录后选择租户，退出再登录验证自动进入上次租户
3. **多租户登录（无记录）**：新用户多租户登录，验证显示租户选择界面
4. **租户切换**：验证切换租户后权限正确变更
