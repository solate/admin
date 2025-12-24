# 后端 CRUD 开发规范

本文档定义了后端增删改查功能的标准化实现规范，确保所有模块保持一致的代码结构和风格。

## 目录

- [架构概述](#架构概述)
- [目录结构](#目录结构)
- [分层设计](#分层设计)
- [命名规范](#命名规范)
- [代码模板](#代码模板)
  - [Handler 层](#handler-层)
  - [Service 层](#service-层)
  - [Repository 层](#repository-层)
  - [DTO 定义](#dto-定义)
- [路由规范](#路由规范)
- [错误处理](#错误处理)
- [最佳实践](#最佳实践)

---

## 架构概述

项目采用经典的三层架构：

```
┌─────────────────────────────────────────────────────────────┐
│                         Handler Layer                        │
│                    (路由处理、参数绑定)                        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                         Service Layer                        │
│              (业务逻辑、数据验证、事务管理)                     │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                       Repository Layer                       │
│                  (数据访问、数据库操作)                        │
└─────────────────────────────────────────────────────────────┘
```

### 职责划分

| 层级 | 职责 | 特点 |
|-----|------|-----|
| **Handler** | 处理 HTTP 请求、参数绑定、调用 Service、返回响应 | 薄层，不含业务逻辑 |
| **Service** | 实现业务逻辑、数据验证、事务管理、调用 Repository | 核心层，业务逻辑集中 |
| **Repository** | 数据库 CRUD 操作、查询构建 | 数据访问层 |

---

## 目录结构

```
backend/
├── internal/
│   ├── handler/          # Handler 层
│   │   └── {resource}_handler.go
│   ├── service/          # Service 层
│   │   └── {resource}_service.go
│   ├── repository/       # Repository 层
│   │   └── {resource}_repo.go
│   ├── dal/
│   │   └── model/        # 数据模型 (GORM Gen 生成)
│   └── dto/              # DTO 定义
│       └── {resource}_dto.go
└── internal/
    └── router/
        └── router.go     # 路由注册
```

---

## 分层设计

### 1. Handler 层

**职责：**
- HTTP 请求处理
- 参数绑定和验证
- 调用 Service 层方法
- 统一响应格式

**特点：**
- 使用依赖注入接收 Service
- 不包含业务逻辑
- 统一错误处理

### 2. Service 层

**职责：**
- 实现核心业务逻辑
- 数据验证（唯一性、业务规则）
- 数据转换（Model ↔ DTO）
- 事务管理
- 操作日志记录

**特点：**
- 使用依赖注入接收 Repository
- 错误使用 `xerr.Wrap()` 包装
- 提供私有转换方法

### 3. Repository 层

**职责：**
- 数据库 CRUD 操作
- 构建查询条件
- 分页查询

**特点：**
- 组合 `*gorm.DB` 和 `*query.Query`
- 使用 GORM Gen 生成的类型安全查询
- 支持条件筛选

---

## 命名规范

### 文件命名

| 层级 | 格式 | 示例 |
|-----|------|-----|
| Handler | `{resource}_handler.go` | `user_handler.go` |
| Service | `{resource}_service.go` | `user_service.go` |
| Repository | `{resource}_repo.go` | `user_repo.go` |
| DTO | `{resource}_dto.go` | `user_dto.go` |

### 结构体命名

| 类型 | 格式 | 示例 |
|-----|------|-----|
| Handler | `{Resource}Handler` | `UserHandler` |
| Service | `{Resource}Service` | `UserService` |
| Repository | `{Resource}Repo` | `UserRepo` |
| 请求 DTO | `{Operation}{Resource}Request` | `CreateUserRequest` |
| 响应 DTO | `{Resource}Response` | `UserResponse` |
| 列表响应 | `List{Resource}sResponse` | `ListUsersResponse` |

### 方法命名

| 操作 | Handler 方法 | Service 方法 | Repository 方法 |
|-----|-------------|-------------|----------------|
| 创建 | `Create{Resource}` | `Create{Resource}` | `Create` |
| 查询详情 | `Get{Resource}` | `Get{Resource}` | `GetByID` |
| 查询列表 | `List{Resource}s` | `List{Resource}s` | `ListWithFilters` |
| 更新 | `Update{Resource}` | `Update{Resource}` | `Update` |
| 删除 | `Delete{Resource}` | `Delete{Resource}` | `Delete` |
| 状态更新 | `Update{Resource}Status` | `UpdateStatus` | `UpdateStatus` |

### 路由命名

| 操作 | HTTP 方法 | 路径格式 | 示例 |
|-----|-----------|---------|-----|
| 创建 | POST | `/{resources}` | `POST /users` |
| 查询详情 | GET | `/{resources}/:id` | `GET /users/:user_id` |
| 查询列表 | GET | `/{resources}` | `GET /users` |
| 更新 | PUT | `/{resources}/:id` | `PUT /users/:user_id` |
| 删除 | DELETE | `/{resources}/:id` | `DELETE /users/:user_id` |
| 状态更新 | PUT | `/{resources}/:id/status/:status` | `PUT /users/:user_id/status/:status` |

### 路径参数命名

- 资源 ID 使用 `{resource}_id` 格式
- 示例：`:user_id`, `:role_id`, `:permission_id`

---

## 代码模板

### Handler 层

```go
package handler

import (
    "github.com/gin-gonic/gin"

    "your-project/internal/dto"
    "your-project/internal/service"
    "your-project/pkg/response"
)

// {Resource}Handler 资源处理器
type {Resource}Handler struct {
    Service *service.{Resource}Service
}

// New{Resource}Handler 创建资源处理器
func New{Resource}Handler(service *service.{Resource}Service) *{Resource}Handler {
    return &{Resource}Handler{
        Service: service,
    }
}

// Create{Resource} 创建资源
// @Summary 创建资源
// @Description 创建新资源
// @Tags {Resource}
// @Accept json
// @Produce json
// @Param request body dto.Create{Resource}Request true "创建请求"
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

// Get{Resource} 获取资源详情
// @Summary 获取资源详情
// @Description 根据 ID 获取资源详情
// @Tags {Resource}
// @Accept json
// @Produce json
// @Param {resource}_id path string true "资源ID"
// @Success 200 {object} dto.{Resource}Response
// @Router /{resources}/:{resource}_id [get]
func (h *{Resource}Handler) Get{Resource}(c *gin.Context) {
    {resource}ID := c.Param("{resource}_id")

    resp, err := h.Service.Get{Resource}(c.Request.Context(), {resource}ID)
    if err != nil {
        response.Error(c, err)
        return
    }

    response.Success(c, resp)
}

// List{Resource}s 获取资源列表
// @Summary 获取资源列表
// @Description 获取资源列表，支持分页和筛选
// @Tags {Resource}
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
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

// Update{Resource} 更新资源
// @Summary 更新资源
// @Description 更新资源信息
// @Tags {Resource}
// @Accept json
// @Produce json
// @Param {resource}_id path string true "资源ID"
// @Param request body dto.Update{Resource}Request true "更新请求"
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

// Delete{Resource} 删除资源
// @Summary 删除资源
// @Description 删除资源（软删除）
// @Tags {Resource}
// @Accept json
// @Produce json
// @Param {resource}_id path string true "资源ID"
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

// Update{Resource}Status 更新资源状态
// @Summary 更新资源状态
// @Description 更新资源状态
// @Tags {Resource}
// @Accept json
// @Produce json
// @Param {resource}_id path string true "资源ID"
// @Param status path int true "状态"
// @Success 200 {object} response.Response
// @Router /{resources}/:{resource}_id/status/:status [put]
func (h *{Resource}Handler) Update{Resource}Status(c *gin.Context) {
    {resource}ID := c.Param("{resource}_id")
    status := c.Param("status")

    if err := h.Service.UpdateStatus(c.Request.Context(), {resource}ID, status); err != nil {
        response.Error(c, err)
        return
    }

    response.Success(c, nil)
}
```

### Service 层

```go
package service

import (
    "context"
    "fmt"

    "gorm.io/gen"
    "gorm.io/gorm"

    "your-project/internal/dal/model"
    "your-project/internal/dto"
    "your-project/internal/repository"
    "your-project/pkg/xerr"
)

// {Resource}Service 资源服务
type {Resource}Service struct {
    Repo *repository.{Resource}Repo
}

// New{Resource}Service 创建资源服务
func New{Resource}Service(repo *repository.{Resource}Repo) *{Resource}Service {
    return &{Resource}Service{
        Repo: repo,
    }
}

// Create{Resource} 创建资源
func (s *{Resource}Service) Create{Resource}(ctx context.Context, req *dto.Create{Resource}Request) (*dto.{Resource}Response, error) {
    // 业务验证
    exists, err := s.Repo.CheckExists(ctx, gen.Cond{
        "{unique_field} = ?", req.UniqueField,
    })
    if err != nil {
        return nil, xerr.Wrap(err, "检查唯一性失败")
    }
    if exists {
        return nil, xerr.New("xxx已存在")
    }

    // 构建模型
    {resource} := &model.{Resource}{
        Field1: req.Field1,
        Field2: req.Field2,
        // 设置默认值
    }

    // 创建
    if err := s.Repo.Create(ctx, {resource}); err != nil {
        return nil, xerr.Wrap(err, "创建xxx失败")
    }

    return s.to{Resource}Response({resource}), nil
}

// Get{Resource} 获取资源详情
func (s *{Resource}Service) Get{Resource}(ctx context.Context, {resource}ID string) (*dto.{Resource}Response, error) {
    {resource}, err := s.Repo.GetByID(ctx, {resource}ID)
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, xerr.New("xxx不存在")
        }
        return nil, xerr.Wrap(err, "查询xxx失败")
    }

    return s.to{Resource}Response({resource}), nil
}

// List{Resource}s 获取资源列表
func (s *{Resource}Service) List{Resource}s(ctx context.Context, req *dto.List{Resource}sRequest) (*dto.List{Resource}sResponse, error) {
    {resource}s, total, err := s.Repo.ListWithFilters(ctx, req)
    if err != nil {
        return nil, xerr.Wrap(err, "查询xxx列表失败")
    }

    items := make([]*dto.{Resource}Response, 0, len({resource}s))
    for _, {resource} := range {resource}s {
        items = append(items, s.to{Resource}Response({resource}))
    }

    return &dto.List{Resource}sResponse{
        PageResponse: dto.PageResponse{
            Page:     req.Page,
            PageSize: req.PageSize,
            Total:    total,
        },
        Items: items,
    }, nil
}

// Update{Resource} 更新资源
func (s *{Resource}Service) Update{Resource}(ctx context.Context, {resource}ID string, req *dto.Update{Resource}Request) (*dto.{Resource}Response, error) {
    // 检查是否存在
    {resource}, err := s.Repo.GetByID(ctx, {resource}ID)
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, xerr.New("xxx不存在")
        }
        return nil, xerr.Wrap(err, "查询xxx失败")
    }

    // 唯一性验证（如果更新了唯一字段）
    if req.UniqueField != nil && *req.UniqueField != {resource}.UniqueField {
        exists, err := s.Repo.CheckExists(ctx, gen.Cond{
            "{unique_field} = ?", *req.UniqueField,
            "{resource}_id != ?", {resource}ID,
        })
        if err != nil {
            return nil, xerr.Wrap(err, "检查唯一性失败")
        }
        if exists {
            return nil, xerr.New("xxx已存在")
        }
    }

    // 构建更新数据
    updates := make(map[string]interface{})
    if req.Field1 != nil {
        updates["field1"] = *req.Field1
    }
    // ... 其他字段

    // 更新
    if err := s.Repo.Update(ctx, {resource}ID, updates); err != nil {
        return nil, xerr.Wrap(err, "更新xxx失败")
    }

    // 返回更新后的数据
    updated{Resource}, err := s.Repo.GetByID(ctx, {resource}ID)
    if err != nil {
        return nil, xerr.Wrap(err, "查询更新后的数据失败")
    }

    return s.to{Resource}Response(updated{Resource}), nil
}

// Delete{Resource} 删除资源
func (s *{Resource}Service) Delete{Resource}(ctx context.Context, {resource}ID string) error {
    // 检查是否存在
    _, err := s.Repo.GetByID(ctx, {resource}ID)
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return xerr.New("xxx不存在")
        }
        return xerr.Wrap(err, "查询xxx失败")
    }

    // 删除
    if err := s.Repo.Delete(ctx, {resource}ID); err != nil {
        return xerr.Wrap(err, "删除xxx失败")
    }

    return nil
}

// UpdateStatus 更新状态
func (s *{Resource}Service) UpdateStatus(ctx context.Context, {resource}ID, status string) error {
    if err := s.Repo.UpdateStatus(ctx, {resource}ID, status); err != nil {
        return xerr.Wrap(err, "更新状态失败")
    }
    return nil
}

// to{Resource}Response 模型转换为响应
func (s *{Resource}Service) to{Resource}Response({resource} *model.{Resource}) *dto.{Resource}Response {
    resp := &dto.{Resource}Response{
        {Resource}ID: {resource}.{Resource}ID,
        Field1:      {resource}.Field1,
        // ... 其他字段
    }

    // 处理可选字段
    if {resource}.OptionalField != nil {
        resp.OptionalField = *{resource}.OptionalField
    } else {
        resp.OptionalField = "" // 或零值
    }

    return resp
}
```

### Repository 层

```go
package repository

import (
    "context"

    "gorm.io/gen"
    "gorm.io/gorm"
    "gorm.io/gen/field"

    "your-project/internal/dal/model"
    "your-project/internal/dal/query"
)

// {Resource}Repo 资源数据访问层
type {Resource}Repo struct {
    db    *gorm.DB
    query *query.Query
}

// New{Resource}Repo 创建资源数据访问层
func New{Resource}Repo(db *gorm.DB, q *query.Query) *{Resource}Repo {
    return &{Resource}Repo{
        db:    db,
        query: q,
    }
}

// Create 创建资源
func (r *{Resource}Repo) Create(ctx context.Context, {resource} *model.{Resource}) error {
    return r.query.{Resource}.WithContext(ctx).Create({resource})
}

// GetByID 根据 ID 查询资源
func (r *{Resource}Repo) GetByID(ctx context.Context, {resource}ID string) (*model.{Resource}, error) {
    return r.query.{Resource}.WithContext(ctx).
        Where(r.query.{Resource}.{Resource}ID.Eq({resource}ID)).
        First()
}

// Update 更新资源
func (r *{Resource}Repo) Update(ctx context.Context, {resource}ID string, updates map[string]interface{}) error {
    _, err := r.query.{Resource}.WithContext(ctx).
        Where(r.query.{Resource}.{Resource}ID.Eq({resource}ID)).
        Updates(updates)
    return err
}

// Delete 删除资源（软删除）
func (r *{Resource}Repo) Delete(ctx context.Context, {resource}ID string) error {
    _, err := r.query.{Resource}.WithContext(ctx).
        Where(r.query.{Resource}.{Resource}ID.Eq({resource}ID)).
        Delete()
    return err
}

// ListWithFilters 条件查询资源列表
func (r *{Resource}Repo) ListWithFilters(ctx context.Context, req *dto.List{Resource}sRequest) ([]*model.{Resource}, int64, error) {
    q := r.query.{Resource}.WithContext(ctx)

    // 构建查询条件
    if req.Keyword != "" {
        q = q.Where(r.query.{Resource}.Name.Like("%" + req.Keyword + "%"))
    }
    if req.Status != nil {
        q = q.Where(r.query.{Resource}.Status.Eq(*req.Status))
    }

    // 获取总数
    total, err := q.Count()
    if err != nil {
        return nil, 0, err
    }

    // 分页查询
    {resource}s, err := q.
        Order(r.query.{Resource}.CreatedAt.Desc()).
        Limit(req.PageSize).
        Offset(req.GetOffset()).
        Find()

    return {resource}s, total, nil
}

// CheckExists 检查唯一性
func (r *{Resource}Repo) CheckExists(ctx context.Context, conds ...gen.Condition) (bool, error) {
    count, err := r.query.{Resource}.WithContext(ctx).
        Where(conds...).
        Count()
    return count > 0, err
}

// UpdateStatus 更新状态
func (r *{Resource}Repo) UpdateStatus(ctx context.Context, {resource}ID, status string) error {
    _, err := r.query.{Resource}.WithContext(ctx).
        Where(r.query.{Resource}.{Resource}ID.Eq({resource}ID)).
        Update(r.query.{Resource}.Status, status)
    return err
}
```

### DTO 定义

```go
package dto

import "your-project/pkg/pagination"

// Create{Resource}Request 创建资源请求
type Create{Resource}Request struct {
    // 必填字段
    RequiredField string `json:"required_field" binding:"required"`

    // 可选字段
    OptionalField string `json:"optional_field"`

    // 枚举字段
    Status int `json:"status" binding:"omitempty,oneof=1 2"`
}

// Update{Resource}Request 更新资源请求
type Update{Resource}Request struct {
    // 所有字段都是指针，表示可选更新
    RequiredField *string `json:"required_field"`
    OptionalField *string `json:"optional_field"`
    Status        *int    `json:"status" binding:"omitempty,oneof=1 2"`
}

// List{Resource}sRequest 资源列表请求
type List{Resource}sRequest struct {
    pagination.PageRequest

    // 筛选条件
    Keyword string `json:"keyword" form:"keyword"`        // 关键词搜索
    Status  *int   `json:"status" form:"status"`          // 状态筛选
}

// {Resource}Response 资源响应
type {Resource}Response struct {
    {Resource}ID  string `json:"{resource}_id"`
    RequiredField string `json:"required_field"`
    OptionalField string `json:"optional_field"`
    Status        int    `json:"status"`
    CreatedAt     int64  `json:"created_at"`
    UpdatedAt     int64  `json:"updated_at"`
}

// List{Resource}sResponse 资源列表响应
type List{Resource}sResponse struct {
    PageResponse
    Items []*{Resource}Response `json:"items"`
}
```

---

## 路由规范

### 路由注册

```go
// 在 router.go 中注册路由

// 资源接口（需要认证）
{resource}Group := authenticated.Group("/{resources}")
{
    {resource}Group.POST("", app.Handlers.{Resource}Handler.Create{Resource})                        // 创建
    {resource}Group.GET("", app.Handlers.{Resource}Handler.List{Resource}s)                          // 列表
    {resource}Group.GET("/:{resource}_id", app.Handlers.{Resource}Handler.Get{Resource})             // 详情
    {resource}Group.PUT("/:{resource}_id", app.Handlers.{Resource}Handler.Update{Resource})          // 更新
    {resource}Group.DELETE("/:{resource}_id", app.Handlers.{Resource}Handler.Delete{Resource})       // 删除
    {resource}Group.PUT("/:{resource}_id/status/:status", app.Handlers.{Resource}Handler.Update{Resource}Status) // 状态
}
```

### RESTful 设计

| 操作 | HTTP 方法 | 路径 | 说明 |
|-----|-----------|------|-----|
| 创建 | POST | `/{resources}` | 创建新资源 |
| 查询详情 | GET | `/{resources}/:id` | 获取单个资源 |
| 查询列表 | GET | `/{resources}` | 获取资源列表，支持分页 |
| 更新 | PUT | `/{resources}/:id` | 更新资源 |
| 删除 | DELETE | `/{resources}/:id` | 软删除资源 |
| 状态更新 | PUT | `/{resources}/:id/status/:status` | 更新资源状态 |

---

## 错误处理

### 统一错误处理

```go
// Handler 层统一使用
if err != nil {
    response.Error(c, err)
    return
}

// Service 层使用 xerr 包装错误
if err != nil {
    return nil, xerr.Wrap(err, "操作失败")
}

// 返回业务错误
return nil, xerr.New("xxx已存在")
```

### 常见错误处理

```go
// 记录不存在
if err == gorm.ErrRecordNotFound {
    return nil, xerr.New("xxx不存在")
}

// 唯一性验证
if exists {
    return nil, xerr.New("xxx已存在")
}

// 参数验证
if err := c.BindJSON(&req); err != nil {
    response.Error(c, err)
    return
}
```

---

## 最佳实践

### 1. 分层职责清晰

- **Handler** 只负责 HTTP 层面的处理，不包含业务逻辑
- **Service** 包含所有业务逻辑，是核心层
- **Repository** 只负责数据访问，不包含业务逻辑

### 2. 使用依赖注入

```go
// 构造函数注入依赖
func New{Resource}Handler(service *service.{Resource}Service) *{Resource}Handler {
    return &{Resource}Handler{Service: service}
}
```

### 3. 错误包装保持错误链

```go
// 使用 xerr.Wrap 保留原始错误信息
return nil, xerr.Wrap(err, "创建xxx失败")
```

### 4. DTO 和 Model 分离

- DTO 用于 API 交互
- Model 用于数据库操作
- Service 层负责 DTO ↔ Model 转换

### 5. 可选字段使用指针

```go
// 更新请求中的可选字段使用指针
type Update{Resource}Request struct {
    Name *string `json:"name"`
    Age  *int    `json:"age"`
}
```

### 6. 分组路由清晰

```go
// 公开接口
public := r.Group("/api/v1")
{
    public.POST("/login", ...)
}

// 认证接口
authenticated := r.Group("/api/v1")
authenticated.Use(middleware.Auth())
{
    authenticated.GET("/users", ...)
}

// 管理员接口
admin := r.Group("/api/v1")
admin.Use(middleware.Auth(), middleware.Admin())
{
    admin.DELETE("/users/:id", ...)
}
```

### 7. 使用 GORM Gen 类型安全查询

```go
// 使用生成的类型安全查询
user, err := r.query.User.Where(
    r.query.User.UserID.Eq(userID),
).First()
```

### 8. 分页查询统一处理

```go
// 使用 pagination 包提供的 PageRequest
type List{Resource}sRequest struct {
    pagination.PageRequest
    // 其他筛选条件
}

// Repository 中使用 req.GetOffset() 和 req.GetLimit()
```

### 9. 软删除优先

```go
// 优先使用软删除
func (r *{Resource}Repo) Delete(ctx context.Context, id string) error {
    _, err := r.query.{Resource}.Where(
        r.query.{Resource}.ID.Eq(id),
    ).Delete()
    return err
}
```

### 10. 操作日志记录

```go
// 在 Service 层记录关键操作
// 可通过中间件自动记录
```

---

## 开发检查清单

开发新的 CRUD 功能时，确保完成以下项目：

- [ ] 创建 DTO 定义（Request/Response）
- [ ] 实现 Handler 层
- [ ] 实现 Service 层（业务逻辑）
- [ ] 实现 Repository 层（数据访问）
- [ ] 注册路由
- [ ] 添加必要的验证规则
- [ ] 实现唯一性检查
- [ ] 实现分页查询
- [ ] 添加错误处理
- [ ] 添加 API 注释（Swagger）
- [ ] 编写单元测试
- [ ] 测试所有接口

---

## 示例参考

完整的实现示例可参考：
- [user_handler.go](../internal/handler/user_handler.go)
- [user_service.go](../internal/service/user_service.go)
- [user_repo.go](../internal/repository/user_repo.go)
- [user_dto.go](../internal/dto/user_dto.go)
