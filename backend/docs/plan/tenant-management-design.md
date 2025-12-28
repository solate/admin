# 租户管理系统设计

## 设计原则

- **租户隔离**：数据完全隔离，通过 `tenant_id` 区分
- **默认租户**：`tenant_id` 为 `"00000000000000000000"`（20 个零），用于超级管理员
- **租户状态**：支持启用/禁用/过期
- **配额管理**：限制租户的资源使用

---

## 数据模型

### 租户表 (tenants)

```sql
CREATE TABLE tenants (
    tenant_id VARCHAR(20) PRIMARY KEY,
    tenant_code VARCHAR(50) NOT NULL UNIQUE,
    tenant_name VARCHAR(100) NOT NULL,
    logo VARCHAR(255),                       -- 租户Logo
    domain VARCHAR(100),                     -- 自定义域名
    status TINYINT DEFAULT 1,                -- 1:正常 0:禁用 2:过期
    expire_at BIGINT,                        -- 过期时间
    max_users INT DEFAULT 100,               -- 最大用户数
    max_storage BIGINT DEFAULT 10737418240,  -- 最大存储(10GB)
    created_at BIGINT,
    updated_at BIGINT,
    deleted_at BIGINT,
    INDEX idx_status (status, expire_at)
);
```

### 租户配置表 (tenant_configs)

```sql
CREATE TABLE tenant_configs (
    config_id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(20) NOT NULL,
    config_key VARCHAR(50) NOT NULL,
    config_value TEXT,
    description VARCHAR(255),
    created_at BIGINT,
    updated_at BIGINT,
    UNIQUE KEY uk_tenant_key (tenant_id, config_key)
);
```

### 租户配额表 (tenant_quotas)

```sql
CREATE TABLE tenant_quotas (
    quota_id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(20) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,      -- users, storage, api_calls
    used_count BIGINT DEFAULT 0,
    max_count BIGINT NOT NULL,
    reset_at BIGINT,                         -- 重置时间（用于周期性配额）
    created_at BIGINT,
    updated_at BIGINT,
    UNIQUE KEY uk_tenant_resource (tenant_id, resource_type)
);
```

---

## 业务逻辑

### 1. 创建租户

```go
func (s *TenantService) Create(ctx context.Context, req *CreateTenantRequest) error {
    // 验证编码唯一性
    existing, _ := s.tenantRepo.GetByCode(ctx, req.TenantCode)
    if existing != nil {
        return errors.New("租户编码已存在")
    }

    tenant := &Tenant{
        TenantID:    uuid.New().String(),
        TenantCode:  req.TenantCode,
        TenantName:  req.TenantName,
        Logo:        req.Logo,
        Domain:      req.Domain,
        Status:      TenantStatusEnabled,
        ExpireAt:    req.ExpireAt,
        MaxUsers:    req.MaxUsers,
        MaxStorage:  req.MaxStorage,
    }

    return s.tenantRepo.Create(ctx, tenant)
}
```

### 2. 更新租户信息

```go
func (s *TenantService) Update(ctx context.Context, tenantID string, req *UpdateTenantRequest) error {
    tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
    if err != nil {
        return errors.New("租户不存在")
    }

    if req.TenantName != nil {
        tenant.TenantName = *req.TenantName
    }
    if req.Logo != nil {
        tenant.Logo = *req.Logo
    }
    if req.Domain != nil {
        // 验证域名唯一性
        if existing, _ := s.tenantRepo.GetByDomain(ctx, *req.Domain); existing != nil && existing.TenantID != tenantID {
            return errors.New("域名已被使用")
        }
        tenant.Domain = *req.Domain
    }

    return s.tenantRepo.Update(ctx, tenant)
}
```

### 3. 启用/禁用租户

```go
func (s *TenantService) SetStatus(ctx context.Context, tenantID string, status int) error {
    tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
    if err != nil {
        return err
    }

    tenant.Status = status
    return s.tenantRepo.Update(ctx, tenant)
}
```

### 4. 配置管理

```go
// 设置配置
func (s *TenantService) SetConfig(ctx context.Context, tenantID, key, value, description string) error {
    config, _ := s.configRepo.GetByTenantAndKey(ctx, tenantID, key)
    if config != nil {
        config.ConfigValue = value
        config.Description = description
        return s.configRepo.Update(ctx, config)
    }

    config = &TenantConfig{
        ConfigID:    uuid.New().String(),
        TenantID:    tenantID,
        ConfigKey:   key,
        ConfigValue: value,
        Description: description,
    }
    return s.configRepo.Create(ctx, config)
}

// 获取配置
func (s *TenantService) GetConfig(ctx context.Context, tenantID, key string) (string, error) {
    config, err := s.configRepo.GetByTenantAndKey(ctx, tenantID, key)
    if err != nil {
        return "", err
    }
    return config.ConfigValue, nil
}

// 获取所有配置
func (s *TenantService) GetAllConfigs(ctx context.Context, tenantID string) (map[string]string, error) {
    configs, _ := s.configRepo.ListByTenant(ctx, tenantID)

    result := make(map[string]string)
    for _, c := range configs {
        result[c.ConfigKey] = c.ConfigValue
    }
    return result, nil
}
```

### 5. 配额检查

```go
// 检查用户配额
func (s *TenantService) CheckUserQuota(ctx context.Context, tenantID string) error {
    quota, _ := s.quotaRepo.GetByTenantAndResource(ctx, tenantID, "users")

    // 未设置配额，使用默认值
    if quota == nil {
        tenant, _ := s.tenantRepo.GetByID(ctx, tenantID)
        if tenant.MaxUsers > 0 {
            used, _ := s.userRepo.CountByTenant(ctx, tenantID)
            if used >= tenant.MaxUsers {
                return errors.New("用户数量已达上限")
            }
        }
        return nil
    }

    if quota.UsedCount >= quota.MaxCount {
        return errors.New("用户数量已达上限")
    }

    return nil
}

// 增加配额使用
func (s *TenantService) IncrementQuota(ctx context.Context, tenantID, resource string, count int64) error {
    quota, _ := s.quotaRepo.GetByTenantAndResource(ctx, tenantID, resource)
    if quota == nil {
        return nil
    }

    quota.UsedCount += count
    return s.quotaRepo.Update(ctx, quota)
}
```

### 6. 过期检查（定时任务）

```go
func (s *TenantService) CheckExpired() {
    now := time.Now().UnixMilli()

    // 查找过期租户
    tenants, _ := s.tenantRepo.ListExpired(context.Background(), now)

    for _, tenant := range tenants {
        if tenant.Status == TenantStatusEnabled {
            tenant.Status = TenantStatusExpired
            s.tenantRepo.Update(context.Background(), tenant)
        }
    }
}
```

---

## API 设计

### 超管接口

```
POST   /api/v1/system/tenants              创建租户
GET    /api/v1/system/tenants              获取租户列表
GET    /api/v1/system/tenants/:id          获取租户详情
PUT    /api/v1/system/tenants/:id          更新租户
DELETE /api/v1/system/tenants/:id          删除租户
POST   /api/v1/system/tenants/:id/enable   启用租户
POST   /api/v1/system/tenants/:id/disable  禁用租户
```

### 租户配置接口

```
GET    /api/v1/system/tenants/:id/configs  获取租户配置
PUT    /api/v1/system/tenants/:id/configs  更新租户配置
```

### 租户自身接口

```
GET    /api/v1/tenant/info                 获取租户信息
PUT    /api/v1/tenant/info                 更新租户信息
GET    /api/v1/tenant/configs              获取配置
GET    /api/v1/tenant/quota                获取配额使用
```

---

## Repository 层

```go
// 根据编码查询
func (r *TenantRepo) GetByCode(ctx context.Context, code string) (*Tenant, error)

// 根据域名查询
func (r *TenantRepo) GetByDomain(ctx context.Context, domain string) (*Tenant, error)

// 查询过期租户
func (r *TenantRepo) ListExpired(ctx context.Context, timestamp int64) ([]*Tenant, error)

// 配置相关
func (r *ConfigRepo) GetByTenantAndKey(ctx context.Context, tenantID, key string) (*TenantConfig, error)
func (r *ConfigRepo) ListByTenant(ctx context.Context, tenantID string) ([]*TenantConfig, error)

// 配额相关
func (r *QuotaRepo) GetByTenantAndResource(ctx context.Context, tenantID, resource string) (*TenantQuota, error)
```

---

## 租户上下文

```go
// 租户上下文（从请求注入）
type TenantContext struct {
    TenantID   string
    TenantCode string
}

// 从Context获取租户ID
func GetTenantID(ctx context.Context) string {
    if tid, ok := ctx.Value("tenant_id").(string); ok {
        return tid
    }
    return ""
}

// 从Context获取租户Code
func GetTenantCode(ctx context.Context) string {
    if tc, ok := ctx.Value("tenant_code").(string); ok {
        return tc
    }
    return ""
}
```

---

## 常量定义

```go
package constants

const (
    // 租户状态
    TenantStatusEnabled = 1  // 正常
    TenantStatusDisabled = 0 // 禁用
    TenantStatusExpired = 2  // 过期

    // 默认租户
    DefaultTenantID   = "00000000000000000000" // 默认租户ID（20个零）
    DefaultTenantCode = "default"              // 默认租户code

    // 默认配额
    DefaultMaxUsers   = 100
    DefaultMaxStorage = 10 * 1024 * 1024 * 1024 // 10GB

    // 配额资源类型
    QuotaResourceUsers     = "users"
    QuotaResourceStorage   = "storage"
    QuotaResourceAPICalls  = "api_calls"
)
```
