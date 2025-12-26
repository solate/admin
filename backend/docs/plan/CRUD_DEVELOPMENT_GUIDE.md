# CRUD 快速开发指南

快速实现标准 CRUD 功能，复制模板即可完成开发。

---

## 快速开始（5分钟上手）

### 开发步骤

```
1. DTO 定义    → internal/dto/{resource}_dto.go
2. Repository  → internal/repository/{resource}_repo.go
3. Service     → internal/service/{resource}_service.go
4. Handler     → internal/handler/{resource}_handler.go
5. 路由注册    → internal/router/router.go
6. 注册 Handler → internal/router/app.go
```

---

## 命名规范

| 类型 | 格式 | 示例 |
|-----|------|-----|
| 文件 | `{resource}_xxx.go` | `user_handler.go` |
| Handler | `{Resource}Handler` | `UserHandler` |
| Service | `{Resource}Service` | `UserService` |
| Repository | `{Resource}Repo` | `UserRepo` |
| 请求 DTO | `{Op}{Resource}Request` | `CreateUserRequest` |
| 响应 DTO | `{Resource}Response` | `UserResponse` |
| 列表响应 | `List{Resource}sResponse` | `ListUsersResponse` |

---

## 分页规范（重要）

使用 **指针内嵌** 方式实现分页：

```go
// pkg/pagination/pagination.go
type Request struct {
    Page     int `form:"page" json:"page" binding:"omitempty,min=1"`
    PageSize int `form:"page_size" json:"page_size" binding:"omitempty,min=1,max=100"`
}

type Response struct {
    Page      int   `json:"page"`
    PageSize  int   `json:"page_size"`
    Total     int64 `json:"total"`
    TotalPage int64 `json:"total_page"`
}

func NewResponse(r *Request, total int64) *Response
```

### DTO 定义

```go
// 列表请求 - 指针内嵌 *pagination.Request
type ListUsersRequest struct {
    *pagination.Request `json:",inline"`
    Keyword             string `form:"keyword" binding:"omitempty"`
    Status              int    `form:"status" binding:"omitempty,oneof=1 2"`
}

// 列表响应 - 指针内嵌 *pagination.Response
// 注意：Response 字段必须在 List 之前（保证 JSON 字段顺序）
type ListUsersResponse struct {
    *pagination.Response `json:",inline"`
    List                 []*UserResponse `json:"list"`
}
```

### Service 层使用

```go
func (s *UserService) ListUsers(ctx context.Context, req *dto.ListUsersRequest) (*dto.ListUsersResponse, error) {
    users, total, err := s.userRepo.ListWithFilters(ctx, req.GetOffset(), req.GetLimit(), req.Keyword, req.Status)
    if err != nil {
        return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户列表失败", err)
    }

    userResponses := make([]*dto.UserResponse, len(users))
    for i, user := range users {
        userResponses[i] = s.toUserResponse(ctx, user)
    }

    return &dto.ListUsersResponse{
        Response: pagination.NewResponse(req.Request, total), // 注意：req.Request 是指针，不需要 &
        List:     userResponses,
    }, nil
}
```

---

## 代码模板

### 1. DTO 模板

```go
package dto

import "admin/pkg/pagination"

// Create{Resource}Request 创建请求
type Create{Resource}Request struct {
    Code        string `json:"code" binding:"required,min=2,max=50"`
    Name        string `json:"name" binding:"required,min=2,max=200"`
    Description string `json:"description"`
}

// Update{Resource}Request 更新请求
type Update{Resource}Request struct {
    Name        string `json:"name" binding:"omitempty,min=2,max=200"`
    Description string `json:"description" binding:"omitempty"`
    Status      int    `json:"status" binding:"omitempty,oneof=1 2"`
}

// {Resource}Response 响应
type {Resource}Response struct {
    {Resource}ID  string `json:"{resource}_id"`
    Code          string `json:"code"`
    Name          string `json:"name"`
    Description   string `json:"description"`
    Status        int    `json:"status"`
    CreatedAt     int64  `json:"created_at"`
    UpdatedAt     int64  `json:"updated_at"`
}

// List{Resource}sRequest 列表请求
type List{Resource}sRequest struct {
    *pagination.Request `json:",inline"`
    Keyword             string `form:"keyword" binding:"omitempty"`
    Status              int    `form:"status" binding:"omitempty,oneof=1 2"`
}

// List{Resource}sResponse 列表响应
type List{Resource}sResponse struct {
    *pagination.Response `json:",inline"`
    List                 []*{Resource}Response `json:"list"`
}
```

### 2. Repository 模板

```go
package repository

import (
    "context"
    "admin/internal/dal/model"
    "admin/internal/dal/query"
    "gorm.io/gorm"
)

type {Resource}Repo struct {
    db *gorm.DB
    q  *query.Query
}

func New{Resource}Repo(db *gorm.DB) *{Resource}Repo {
    return &{Resource}Repo{
        db: db,
        q:  query.Use(db),
    }
}

// Create 创建
func (r *{Resource}Repo) Create(ctx context.Context, {resource} *model.{Resource}) error {
    return r.q.{Resource}.WithContext(ctx).Create({resource})
}

// GetByID 根据 ID 查询
func (r *{Resource}Repo) GetByID(ctx context.Context, {resource}ID string) (*model.{Resource}, error) {
    return r.q.{Resource}.WithContext(ctx).
        Where(r.q.{Resource}.{Resource}ID.Eq({resource}ID)).
        First()
}

// Update 更新
func (r *{Resource}Repo) Update(ctx context.Context, {resource}ID string, updates map[string]interface{}) error {
    _, err := r.q.{Resource}.WithContext(ctx).
        Where(r.q.{Resource}.{Resource}ID.Eq({resource}ID)).
        Updates(updates)
    return err
}

// Delete 删除
func (r *{Resource}Repo) Delete(ctx context.Context, {resource}ID string) error {
    _, err := r.q.{Resource}.WithContext(ctx).
        Where(r.q.{Resource}.{Resource}ID.Eq({resource}ID)).
        Delete()
    return err
}

// ListWithFilters 分页列表查询
func (r *{Resource}Repo) ListWithFilters(ctx context.Context, offset, limit int, keywordFilter string, statusFilter int) ([]*model.{Resource}, int64, error) {
    query := r.q.{Resource}.WithContext(ctx)

    if keywordFilter != "" {
        query = query.Where(r.q.{Resource}.Name.Like("%"+keywordFilter+"%"))
    }
    if statusFilter != 0 {
        query = query.Where(r.q.{Resource}.Status.Eq(int16(statusFilter)))
    }

    total, err := query.Count()
    if err != nil {
        return nil, 0, err
    }

    items, err := query.
        Order(r.q.{Resource}.CreatedAt.Desc()).
        Offset(offset).
        Limit(limit).
        Find()

    return items, total, err
}

// CheckExists 唯一性检查
func (r *{Resource}Repo) CheckExists(ctx context.Context, code string) (bool, error) {
    count, err := r.q.{Resource}.WithContext(ctx).
        Where(r.q.{Resource}.Code.Eq(code)).
        Count()
    return count > 0, err
}

// UpdateStatus 更新状态
func (r *{Resource}Repo) UpdateStatus(ctx context.Context, {resource}ID string, status int) error {
    _, err := r.q.{Resource}.WithContext(ctx).
        Where(r.q.{Resource}.{Resource}ID.Eq({resource}ID)).
        Update(r.q.{Resource}.Status, status)
    return err
}
```

### 3. Service 模板

```go
package service

import (
    "context"
    "admin/internal/dal/model"
    "admin/internal/dto"
    "admin/internal/repository"
    "admin/pkg/xerr"
    "gorm.io/gorm"
)

type {Resource}Service struct {
    Repo *repository.{Resource}Repo
}

func New{Resource}Service(repo *repository.{Resource}Repo) *{Resource}Service {
    return &{Resource}Service{Repo: repo}
}

// Create{Resource} 创建
func (s *{Resource}Service) Create{Resource}(ctx context.Context, req *dto.Create{Resource}Request) (*dto.{Resource}Response, error) {
    // 唯一性验证
    exists, err := s.Repo.CheckExists(ctx, req.Code)
    if err != nil {
        return nil, xerr.Wrap(err, "检查唯一性失败")
    }
    if exists {
        return nil, xerr.New("编码已存在")
    }

    // 创建
    {resource} := &model.{Resource}{
        Code:        req.Code,
        Name:        req.Name,
        Description: &req.Description,
        Status:      1,
    }

    if err := s.Repo.Create(ctx, {resource}); err != nil {
        return nil, xerr.Wrap(err, "创建失败")
    }

    return s.to{Resource}Response({resource}), nil
}

// Get{Resource}ByID 详情
func (s *{Resource}Service) Get{Resource}ByID(ctx context.Context, {resource}ID string) (*dto.{Resource}Response, error) {
    {resource}, err := s.Repo.GetByID(ctx, {resource}ID)
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, xerr.New("记录不存在")
        }
        return nil, xerr.Wrap(err, "查询失败")
    }
    return s.to{Resource}Response({resource}), nil
}

// List{Resource}s 列表
func (s *{Resource}Service) List{Resource}s(ctx context.Context, req *dto.List{Resource}sRequest) (*dto.List{Resource}sResponse, error) {
    items, total, err := s.Repo.ListWithFilters(ctx, req.GetOffset(), req.GetLimit(), req.Keyword, req.Status)
    if err != nil {
        return nil, xerr.Wrap(err, "查询列表失败")
    }

    list := make([]*dto.{Resource}Response, len(items))
    for i, item := range items {
        list[i] = s.to{Resource}Response(item)
    }

    return &dto.List{Resource}sResponse{
        Response: pagination.NewResponse(req.Request, total),
        List:     list,
    }, nil
}

// Update{Resource} 更新
func (s *{Resource}Service) Update{Resource}(ctx context.Context, {resource}ID string, req *dto.Update{Resource}Request) (*dto.{Resource}Response, error) {
    // 检查是否存在
    old, err := s.Repo.GetByID(ctx, {resource}ID)
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, xerr.New("记录不存在")
        }
        return nil, xerr.Wrap(err, "查询失败")
    }

    // 构建更新数据
    updates := make(map[string]interface{})
    if req.Name != "" {
        updates["name"] = req.Name
    }
    if req.Description != "" {
        updates["description"] = &req.Description
    }
    if req.Status != 0 {
        updates["status"] = req.Status
    }

    if err := s.Repo.Update(ctx, {resource}ID, updates); err != nil {
        return nil, xerr.Wrap(err, "更新失败")
    }

    updated, _ := s.Repo.GetByID(ctx, {resource}ID)
    return s.to{Resource}Response(updated), nil
}

// Delete{Resource} 删除
func (s *{Resource}Service) Delete{Resource}(ctx context.Context, {resource}ID string) error {
    if _, err := s.Repo.GetByID(ctx, {resource}ID); err != nil {
        if err == gorm.ErrRecordNotFound {
            return xerr.New("记录不存在")
        }
        return xerr.Wrap(err, "查询失败")
    }

    if err := s.Repo.Delete(ctx, {resource}ID); err != nil {
        return xerr.Wrap(err, "删除失败")
    }
    return nil
}

// Update{Resource}Status 更新状态
func (s *{Resource}Service) Update{Resource}Status(ctx context.Context, {resource}ID string, status int) error {
    if err := s.Repo.UpdateStatus(ctx, {resource}ID, status); err != nil {
        return xerr.Wrap(err, "更新状态失败")
    }
    return nil
}

// to{Resource}Response 转换
func (s *{Resource}Service) to{Resource}Response({resource} *model.{Resource}) *dto.{Resource}Response {
    resp := &dto.{Resource}Response{
        {Resource}ID: {resource}.{Resource}ID,
        Code:        {resource}.Code,
        Name:        {resource}.Name,
        Status:      int({resource}.Status),
        CreatedAt:   {resource}.CreatedAt.UnixMilli(),
        UpdatedAt:   {resource}.UpdatedAt.UnixMilli(),
    }
    if {resource}.Description != nil {
        resp.Description = *{resource}.Description
    }
    return resp
}
```

### 4. Handler 模板

```go
package handler

import (
    "github.com/gin-gonic/gin"
    "admin/internal/dto"
    "admin/internal/service"
    "admin/pkg/response"
)

type {Resource}Handler struct {
    Service *service.{Resource}Service
}

func New{Resource}Handler(service *service.{Resource}Service) *{Resource}Handler {
    return &{Resource}Handler{Service: service}
}

// Create{Resource} 创建
// @Summary 创建{Resource}
// @Tags {Resource}
// @Accept json
// @Produce json
// @Param request body dto.Create{Resource}Request true "请求"
// @Success 200 {object} dto.{Resource}Response
// @Router /{resources} [post]
func (h *{Resource}Handler) Create{Resource}(c *gin.Context) {
    var req dto.Create{Resource}Request
    if err := c.BindJSON(&req); err != nil {
        response.Error(c, err)
        return
    }

    resp, err := h.Service.Create{Resource}(c.Request.Context(), &req)
    if err != nil {
        response.Error(c, err)
        return
    }

    response.Success(c, resp)
}

// Get{Resource} 详情
// @Summary 获取{Resource}详情
// @Tags {Resource}
// @Param {resource}_id path string true "ID"
// @Success 200 {object} dto.{Resource}Response
// @Router /{resources}/:{resource}_id [get]
func (h *{Resource}Handler) Get{Resource}(c *gin.Context) {
    {resource}ID := c.Param("{resource}_id")

    resp, err := h.Service.Get{Resource}ByID(c.Request.Context(), {resource}ID)
    if err != nil {
        response.Error(c, err)
        return
    }

    response.Success(c, resp)
}

// List{Resource}s 列表
// @Summary 获取{Resource}列表
// @Tags {Resource}
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param keyword query string false "关键词"
// @Param status query int false "状态"
// @Success 200 {object} dto.List{Resource}sResponse
// @Router /{resources} [get]
func (h *{Resource}Handler) List{Resource}s(c *gin.Context) {
    var req dto.List{Resource}sRequest
    if err := c.ShouldBindQuery(&req); err != nil {
        response.Error(c, err)
        return
    }

    resp, err := h.Service.List{Resource}s(c.Request.Context(), &req)
    if err != nil {
        response.Error(c, err)
        return
    }

    response.Success(c, resp)
}

// Update{Resource} 更新
// @Summary 更新{Resource}
// @Tags {Resource}
// @Param {resource}_id path string true "ID"
// @Param request body dto.Update{Resource}Request true "请求"
// @Success 200 {object} dto.{Resource}Response
// @Router /{resources}/:{resource}_id [put]
func (h *{Resource}Handler) Update{Resource}(c *gin.Context) {
    {resource}ID := c.Param("{resource}_id")
    var req dto.Update{Resource}Request
    if err := c.BindJSON(&req); err != nil {
        response.Error(c, err)
        return
    }

    resp, err := h.Service.Update{Resource}(c.Request.Context(), {resource}ID, &req)
    if err != nil {
        response.Error(c, err)
        return
    }

    response.Success(c, resp)
}

// Delete{Resource} 删除
// @Summary 删除{Resource}
// @Tags {Resource}
// @Param {resource}_id path string true "ID"
// @Success 200 {object} response.Response
// @Router /{resources}/:{resource}_id [delete]
func (h *{Resource}Handler) Delete{Resource}(c *gin.Context) {
    {resource}ID := c.Param("{resource}_id")

    if err := h.Service.Delete{Resource}(c.Request.Context(), {resource}ID); err != nil {
        response.Error(c, err)
        return
    }

    response.Success(c, nil)
}

// Update{Resource}Status 更新状态
// @Summary 更新{Resource}状态
// @Tags {Resource}
// @Param {resource}_id path string true "ID"
// @Param status path int true "状态"
// @Success 200 {object} response.Response
// @Router /{resources}/:{resource}_id/status/:status [put]
func (h *{Resource}Handler) Update{Resource}Status(c *gin.Context) {
    {resource}ID := c.Param("{resource}_id")
    status := c.Param("status")

    if err := h.Service.Update{Resource}Status(c.Request.Context(), {resource}ID, status); err != nil {
        response.Error(c, err)
        return
    }

    response.Success(c, nil)
}
```

### 5. 路由注册模板

```go
// internal/router/router.go

// 在 authenticated 组中添加
{resource} := authenticated.Group("/{resources}")
{
    {resource}.POST("", app.Handlers.{Resource}Handler.Create{Resource})
    {resource}.GET("", app.Handlers.{Resource}Handler.List{Resource}s)
    {resource}.GET("/:{resource}_id", app.Handlers.{Resource}Handler.Get{Resource})
    {resource}.PUT("/:{resource}_id", app.Handlers.{Resource}Handler.Update{Resource})
    {resource}.DELETE("/:{resource}_id", app.Handlers.{Resource}Handler.Delete{Resource})
    {resource}.PUT("/:{resource}_id/status/:status", app.Handlers.{Resource}Handler.Update{Resource}Status)
}
```

### 6. Handler 注册

```go
// internal/router/app.go

// Handlers 结构体添加字段
type Handlers struct {
    // ...
    {Resource}Handler *handler.{Resource}Handler
}

// initHandlers 方法中添加初始化
func (app *App) initHandlers() {
    // ...
    app.Handlers.{Resource}Handler = handler.New{Resource}Handler(app.Services.{Resource}Service)
}
```

### 7. Service 注册

```go
// internal/router/app.go

// Services 结构体添加字段
type Services struct {
    // ...
    {Resource}Service *service.{Resource}Service
}

// initServices 方法中添加初始化
func (app *App) initServices() {
    // ...
    app.Services.{Resource}Service = service.New{Resource}Service(
        app.Repositories.{Resource}Repo,
    )
}
```

### 8. Repository 注册

```go
// internal/router/app.go

// Repositories 结构体添加字段
type Repositories struct {
    // ...
    {Resource}Repo *repository.{Resource}Repo
}

// initRepositories 方法中添加初始化
func (app *App) initRepositories() {
    // ...
    app.Repositories.{Resource}Repo = repository.New{Resource}Repo(app.db)
}
```

---

## 状态字段规范

| 层级 | 类型 | 说明 |
|-----|------|-----|
| 数据库 | `SMALLINT` | 节省空间 |
| Model | `int16` | 对应数据库 |
| DTO | `int` | 简化 JSON 绑定 |

```go
// Model
Status int16 `gorm:"type:smallint;not null;default:1"`

// DTO Request
Status int `json:"status" binding:"omitempty,oneof=1 2"`

// DTO Response
Status int `json:"status"`
```

---

## 操作日志规范

常量定义在 `pkg/constants/operation_log.go`：

```go
// 模块常量
const (
    ModuleUser         = "user"
    ModuleRole         = "role"
    ModuleTenant       = "tenant"
    ModuleTenantMember = "tenant_member"
)

// 资源类型常量
const (
    ResourceTypeUser         = "user"
    ResourceTypeRole         = "role"
    ResourceTypeTenant       = "tenant"
    ResourceTypeTenantMember = "tenant_member"
)
```

Service 中使用：

```go
import (
    "admin/pkg/constants"
    "admin/pkg/operationlog"
)

// 创建
ctx = operationlog.RecordCreate(ctx, constants.ModuleTenant, constants.ResourceTypeTenant, tenant.TenantID, tenant.Name, tenant)

// 更新
ctx = operationlog.RecordUpdate(ctx, constants.ModuleTenant, constants.ResourceTypeTenant, updated.TenantID, updated.Name, oldTenant, updatedTenant)

// 删除（不需要返回 ctx）
operationlog.RecordDelete(ctx, constants.ModuleTenant, constants.ResourceTypeTenant, tenant.TenantID, tenant.Name, tenant)
```

---

## 常见问题

### Q: 分页为什么要用指针内嵌？
A: Swagger 不支持 Go 泛型，指针内嵌可实现字段扁平化且兼容文档生成。

### Q: Response 字段为什么要放在 List 之前？
A: 保证 JSON 序列化时字段顺序正确。

### Q: req.Request 为什么不需要 &？
A: 因为 Request 是指针类型，`req.Request` 本身就是指针。

### Q: 状态字段为什么不用指针？
A: 状态值已规范（1/2），零值不会出现，用 `omitempty` 即可。

---

## 参考示例

- [user_dto.go](../internal/dto/user_dto.go)
- [user_repo.go](../internal/repository/user_repo.go)
- [user_service.go](../internal/service/user_service.go)
- [user_handler.go](../internal/handler/user_handler.go)
