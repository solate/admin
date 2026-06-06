# Step 08: 域 — 租户管理 CRUD

## 目标

实现租户管理的完整 CRUD：创建、更新、删除（软删除）、查询（详情+列表）。

这是第一个完整的业务域，建立 handler → service → repository 的标准模式。

## 前置条件

- Step 07 完成，RBAC 中间件就绪

## 文件清单

```
internal/
├── handler/tenant/
│   ├── tenant.go              # Handler struct + NewHandler + RegisterRoutes
│   ├── create.go              # POST /tenants
│   ├── update.go              # PUT /tenants/:id
│   ├── delete.go              # DELETE /tenants/:id
│   └── query.go               # GET /tenants + GET /tenants/:id
├── service/tenant/
│   ├── tenant.go              # Service struct + NewService
│   ├── create.go
│   ├── update.go
│   ├── delete.go
│   ├── query.go
│   └── converter.go           # modelToTenantInfo (unexported)
├── repository/
│   └── tenant_repo.go         # TenantRepo
└── dto/
    └── tenant_dto.go          # CreateTenantRequest, UpdateTenantRequest, TenantInfo, TenantListRequest
```

## 实现细节

### 1. 标准域模式（所有域都遵循）

```
每个域 4 个操作文件：
  create.go  — 创建逻辑
  update.go  — 更新逻辑
  delete.go  — 删除逻辑（软删除）
  query.go   — 查询逻辑（详情 + 列表）

每个域 1 个 converter 文件：
  converter.go — model → DTO 转换

Handler 方法签名统一：
  func (h *Handler) Create(c *gin.Context)
  func (h *Handler) Update(c *gin.Context)
  func (h *Handler) Delete(c *gin.Context)
  func (h *Handler) GetByID(c *gin.Context)
  func (h *Handler) List(c *gin.Context)
```

### 2. Repository 层

```go
type TenantRepo struct {
    q *query.Query
}

func NewTenantRepo(db *gorm.DB) *TenantRepo {
    return &TenantRepo{q: query.Use(db)}
}

// 方法：
func (r *TenantRepo) Create(ctx context.Context, tenant *model.Tenant) error
func (r *TenantRepo) Update(ctx context.Context, id string, updates map[string]interface{}) error
func (r *TenantRepo) Delete(ctx context.Context, id string) error  // 软删除
func (r *TenantRepo) GetByID(ctx context.Context, id string) (*model.Tenant, error)
func (r *TenantRepo) GetByCode(ctx context.Context, code string) (*model.Tenant, error)
func (r *TenantRepo) List(ctx context.Context, req *dto.TenantListRequest) ([]*model.Tenant, int64, error)
```

**Repository 规则**：
- 使用 `r.q.Tenant.WithContext(ctx)` （GORM Gen 生成）
- 列表查询双排序：`.Order(r.q.Tenant.CreatedAt.Desc()).Order(r.q.Tenant.TenantID.Desc())`
- 软删除：`updates["deleted_at"] = time.Now().UnixMilli()`
- 不加 tenant_id 过滤（tenants 是跨租户的）

### 3. Service 层

```go
type Service struct {
    repo *repository.TenantRepo
}

func NewService(db *gorm.DB) *Service {
    return &Service{repo: repository.NewTenantRepo(db)}
}
```

**Service 规则**：
- 错误用 `xerr.Wrap` 包装
- 不在循环中做数据库操作
- 列表返回 `(list []*dto.XxxInfo, total int64, error)`
- Converter 接收 `[]*model.Xxx`，内部构建 map

### 4. Handler 层

```go
type Handler struct {
    svc *tenant.Service
}

func NewHandler(db *gorm.DB) *Handler {
    return &Handler{svc: tenant.NewService(db)}
}
```

**Handler 规则**：
- 参数绑定：`ShouldBindJSON` / `ShouldBindQuery` / `ShouldBindUri`
- 调用 Service，处理错误返回
- 不包含业务逻辑
- Swagger 注解写在 handler 方法上

### 5. 租户特殊规则

- `tenant_code` 全局唯一，创建时校验
- 删除租户前检查是否有用户（`CountUsers`）
- **租户是跨租户资源**：超管可以管理所有租户
- 不加 tenant_id 过滤

### 6. DTO 示例

```go
type CreateTenantRequest struct {
    TenantCode   string `json:"tenant_code" binding:"required"`
    Name         string `json:"name" binding:"required"`
    Description  string `json:"description"`
    ContactName  string `json:"contact_name"`
    ContactPhone string `json:"contact_phone"`
}

type TenantInfo struct {
    TenantID     string `json:"tenant_id" example:"123456789012345678"`
    TenantCode   string `json:"tenant_code" example:"default"`
    Name         string `json:"name" example:"默认租户"`
    Status       int    `json:"status" example:"1"`
    CreatedAt    int64  `json:"created_at" example:"1717027200000"`
}

type TenantListRequest struct {
    Page     int    `form:"page"`
    PageSize int    `form:"page_size"`
    Keyword  string `form:"keyword"`
    Status   *int   `form:"status"`
}
```

## 验收标准

- [ ] `POST /tenants` 创建租户成功
- [ ] `POST /tenants` tenant_code 重复 → 返回 409
- [ ] `PUT /tenants/:id` 更新租户成功
- [ ] `DELETE /tenants/:id` 软删除成功
- [ ] `DELETE /tenants/:id` 租户下有用户 → 返回 409
- [ ] `GET /tenants/:id` 返回租户详情
- [ ] `GET /tenants` 分页列表正确（含 keyword 搜索）
- [ ] 列表按 `created_at DESC, id DESC` 排序
- [ ] 非超管无法访问租户管理接口 → 403

## AI 协作提示

```
请按 step-08-domain-tenant.md 实现租户管理 CRUD。
这是第一个完整域，建立标准模式，后续域都遵循相同模式：
1. handler/tenant/ — 4 个操作文件 + 1 个主文件
2. service/tenant/ — 4 个操作文件 + 1 个主文件 + 1 个 converter
3. repository/ — tenant_repo.go 一个文件
4. dto/ — tenant_dto.go 一个文件

重点：
- Repository 用 GORM Gen 生成的 query
- 列表双排序字段
- 软删除（更新 deleted_at 字段）
- 错误用 xerr 包装
```
