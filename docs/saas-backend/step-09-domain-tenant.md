# Step 09: 租户管理域

## 目标

实现租户（Tenant）的 CRUD 管理。租户是多租户系统的顶层实体，只有超管可以管理。

## 前置条件

- Step 08 完成，权限系统可用

## 文件清单

```
internal/
├── handler/tenant/
│   ├── handler.go
│   ├── create.go
│   ├── update.go
│   ├── delete.go
│   ├── query.go               # List + GetByID
│   └── dto.go
├── service/tenant/
│   ├── service.go
│   ├── create.go
│   ├── update.go
│   ├── delete.go
│   ├── query.go
│   └── converter.go
├── repository/
│   └── tenant_repo.go         # 补充 CRUD 方法
```

## 实现规范

### 1. API 设计

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/tenants | 租户列表（分页） |
| GET | /api/v1/tenants/:id | 租户详情 |
| POST | /api/v1/tenants | 创建租户 |
| PUT | /api/v1/tenants/:id | 更新租户 |
| DELETE | /api/v1/tenants/:id | 删除租户（软删除） |
| PUT | /api/v1/tenants/:id/status | 启用/禁用租户 |

### 2. 业务规则

- 只有 `super_admin` 可以管理租户
- 租户表**无 tenant_id 字段**（它自己就是顶层实体）
- `tenant_code` 全局唯一，创建后不可修改
- 禁用租户后，该租户下所有用户无法登录（登录时检查）
- 删除租户前检查是否有活跃用户（有则拒绝删除或提示）
- 列表查询不带 tenant_id 过滤（超管看所有租户）

### 3. DTO

```go
type CreateTenantRequest struct {
    TenantCode   string `json:"tenant_code" binding:"required,min=2,max=50,alphanum"`
    Name         string `json:"name" binding:"required,max=100"`
    Description  string `json:"description" binding:"max=500"`
    ContactName  string `json:"contact_name" binding:"max=50"`
    ContactPhone string `json:"contact_phone" binding:"max=20"`
    MaxUsers     int    `json:"max_users" binding:"omitempty,min=1"`
    ExpiredAt    int64  `json:"expired_at"`  // 0=永不过期
}

type UpdateTenantRequest struct {
    Name         string `json:"name" binding:"omitempty,max=100"`
    Description  string `json:"description" binding:"max=500"`
    ContactName  string `json:"contact_name" binding:"max=50"`
    ContactPhone string `json:"contact_phone" binding:"max=20"`
    MaxUsers     int    `json:"max_users" binding:"omitempty,min=1"`
    ExpiredAt    int64  `json:"expired_at"`
}

type TenantInfo struct {
    TenantID     string `json:"tenant_id"`
    TenantCode   string `json:"tenant_code"`
    Name         string `json:"name"`
    Description  string `json:"description"`
    ContactName  string `json:"contact_name"`
    ContactPhone string `json:"contact_phone"`
    Status       int    `json:"status"`
    MaxUsers     int    `json:"max_users"`
    ExpiredAt    int64  `json:"expired_at"`
    UserCount    int64  `json:"user_count"` // 当前用户数
    CreatedAt    int64  `json:"created_at"`
}
```

### 4. 关键实现点

```go
// service/tenant/create.go
func (s *Service) Create(ctx context.Context, req *CreateTenantRequest) (string, error) {
    // 1. 检查 tenant_code 唯一性
    // 2. 生成 tenant_id
    // 3. 创建租户记录
    // 4. 创建该租户的默认管理员角色（role_code=admin）
    // 5. 返回 tenant_id
}
```

创建租户时自动初始化：
- 默认 `admin` 角色
- 可选：创建该租户的管理员用户

## 验收标准

```bash
# 1. 创建租户
curl -s -X POST http://localhost:8080/api/v1/tenants \
  -H "Authorization: Bearer $TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{"tenant_code":"acme","name":"ACME 公司","max_users":50}'
# 期望：{"code":0,"data":{"tenant_id":"..."}}

# 2. 租户列表
curl -s "http://localhost:8080/api/v1/tenants?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN_ADMIN"
# 期望：分页返回租户列表，含 user_count

# 3. tenant_code 唯一性
# 再次创建相同 tenant_code → {"code":40901,"message":"租户编码已存在"}

# 4. 禁用租户
curl -s -X PUT http://localhost:8080/api/v1/tenants/TENANT_ID/status \
  -H "Authorization: Bearer $TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{"status":2}'
# 期望：成功。该租户用户登录时提示"租户已被禁用"

# 5. 非超管无法操作
# 普通 admin 角色访问 → 403
```

## AI 协作提示

```
请按 step-09-domain-tenant.md 实现租户管理域。

要点：
1. 完整 handler/service/repository 三层
2. 租户表无 tenant_id 过滤（超管看全部）
3. tenant_code 创建后不可修改
4. 创建租户时自动创建 admin 角色
5. 列表查询返回 user_count（关联查询 users 表 COUNT）
6. 双排序：ORDER BY created_at DESC, tenant_id DESC
7. 软删除：deleted_at = 当前毫秒时间戳
```

---

*上一步：[Step 08 - 权限管理](step-08-permission-management.md) | 下一步：[Step 10 - 用户管理](step-10-domain-user.md)*
