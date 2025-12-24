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
| 查询详情 | `Get{Resource}` | `Get{Resource}ByID` | `GetByID` |
| 查询列表 | `List{Resource}s` | `List{Resource}s` | `ListWithFilters` |
| 更新 | `Update{Resource}` | `Update{Resource}` | `Update` |
| 删除 | `Delete{Resource}` | `Delete{Resource}` | `Delete` |
| 状态更新 | `Update{Resource}Status` | `Update{Resource}Status` | `UpdateStatus` |

**说明：**
- Service 层的查询详情方法命名为 `Get{Resource}ByID`，更明确地表达通过 ID 查询的语义
- Handler 层保持 `Get{Resource}` 简洁命名
- Repository 层统一使用 `GetByID` 命名

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

    resp, err := h.Service.Get{Resource}ByID(c.Request.Context(), {resource}ID)
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

    if err := h.Service.Update{Resource}Status(c.Request.Context(), {resource}ID, status); err != nil {
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

// Get{Resource}ByID 获取资源详情
func (s *{Resource}Service) Get{Resource}ByID(ctx context.Context, {resource}ID string) (*dto.{Resource}Response, error) {
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
    {resource}s, total, err := s.Repo.ListWithFilters(ctx, req.GetOffset(), req.GetLimit(), req.Keyword, req.Status)
    if err != nil {
        return nil, xerr.Wrap(err, "查询xxx列表失败")
    }

    items := make([]*dto.{Resource}Response, 0, len({resource}s))
    for _, {resource} := range {resource}s {
        items = append(items, s.to{Resource}Response({resource}))
    }

    return &dto.List{Resource}sResponse{
        Response: pagination.NewResponse(items, &req.Request, total),
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

// Update{Resource}Status 更新状态
func (s *{Resource}Service) Update{Resource}Status(ctx context.Context, {resource}ID string, status int) error {
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

    "your-project/internal/dal/model"
    "your-project/internal/dal/query"
)

// {Resource}Repo 资源数据访问层
type {Resource}Repo struct {
    db *gorm.DB
    q  *query.Query
}

// New{Resource}Repo 创建资源数据访问层（内部初始化 query）
func New{Resource}Repo(db *gorm.DB) *{Resource}Repo {
    return &{Resource}Repo{
        db: db,
        q:  query.Use(db),
    }
}

// Create 创建资源
func (r *{Resource}Repo) Create(ctx context.Context, {resource} *model.{Resource}) error {
    return r.q.{Resource}.WithContext(ctx).Create({resource})
}

// GetByID 根据 ID 查询资源
func (r *{Resource}Repo) GetByID(ctx context.Context, {resource}ID string) (*model.{Resource}, error) {
    return r.q.{Resource}.WithContext(ctx).
        Where(r.q.{Resource}.{Resource}ID.Eq({resource}ID)).
        First()
}

// Update 更新资源
func (r *{Resource}Repo) Update(ctx context.Context, {resource}ID string, updates map[string]interface{}) error {
    _, err := r.q.{Resource}.WithContext(ctx).
        Where(r.q.{Resource}.{Resource}ID.Eq({resource}ID)).
        Updates(updates)
    return err
}

// Delete 删除资源（软删除）
func (r *{Resource}Repo) Delete(ctx context.Context, {resource}ID string) error {
    _, err := r.q.{Resource}.WithContext(ctx).
        Where(r.q.{Resource}.{Resource}ID.Eq({resource}ID)).
        Delete()
    return err
}

// ListWithFilters 条件查询资源列表
func (r *{Resource}Repo) ListWithFilters(ctx context.Context, offset, limit int, keywordFilter string, statusFilter int) ([]*model.{Resource}, int64, error) {
    query := r.q.{Resource}.WithContext(ctx)

    // 构建查询条件
    if keywordFilter != "" {
        query = query.Where(r.q.{Resource}.Name.Like("%" + keywordFilter + "%"))
    }
    if statusFilter != 0 {
        query = query.Where(r.q.{Resource}.Status.Eq(int16(statusFilter)))
    }

    // 获取总数
    total, err := query.Count()
    if err != nil {
        return nil, 0, err
    }

    // 分页查询
    {resource}s, err := query.
        Order(r.q.{Resource}.CreatedAt.Desc()).
        Offset(offset).
        Limit(limit).
        Find()

    return {resource}s, total, nil
}

// CheckExists 检查唯一性（具体参数根据业务需求定义）
func (r *{Resource}Repo) CheckExists(ctx context.Context, uniqueValue string) (bool, error) {
    count, err := r.q.{Resource}.WithContext(ctx).
        Where(r.q.{Resource}.UniqueField.Eq(uniqueValue)).
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
    // 可选字段（非指针，通过空值判断是否更新）
    RequiredField string `json:"required_field" binding:"omitempty"`
    OptionalField string `json:"optional_field" binding:"omitempty"`
    Status        int    `json:"status" binding:"omitempty,oneof=1 2"`
}

// List{Resource}sRequest 资源列表请求
type List{Resource}sRequest struct {
    pagination.Request

    // 筛选条件
    Keyword string `json:"keyword" form:"keyword" binding:"omitempty"`        // 关键词搜索
    Status  int    `json:"status" form:"status" binding:"omitempty,oneof=1 2"` // 状态筛选（非指针类型）
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
    Response pagination.Response[*{Resource}Response]
}
```

---

## 数据类型规范

### 状态字段类型

| 层级 | 类型 | 说明 |
|-----|------|-----|
| 数据库 | `SMALLINT` | 节省空间，状态值范围小 |
| Model | `int16` | 对应数据库 SMALLINT |
| DTO Request | `int` | 简化 JSON 绑定，使用 validation tag |
| DTO Response | `int` | JSON 序列化 |

**示例：**
```sql
-- 数据库表定义
status SMALLINT NOT NULL DEFAULT 1
```

```go
// Model 定义
type User struct {
    Status int16 `gorm:"column:status;type:smallint;not null;default:1"`
}

// DTO Request 定义
type UpdateUserRequest struct {
    Status int `json:"status" binding:"omitempty,oneof=1 2"`
}

// DTO Response 定义
type UserResponse struct {
    Status int `json:"status"`
}
```

### 指针字段使用规则

**UpdateRequest 中的状态字段不需要使用指针**，原因：
- 状态值已规范（1=启用，2=禁用），零值不会出现
- 使用 `omitempty` tag 可省略该字段
- 在 Service 层通过 `req.Status != 0` 判断是否需要更新

**其他可选字段仍使用指针**，用于区分"不更新"和"更新为零值"：

```go
type UpdateUserRequest struct {
    Password string `json:"password" binding:"omitempty"`         // 密码（非指针，空字符串不更新）
    Phone    string `json:"phone" binding:"omitempty"`            // 手机号（非指针，空字符串不更新）
    Email    string `json:"email" binding:"omitempty"`            // 邮箱（非指针，空字符串不更新）
    Status   int    `json:"status" binding:"omitempty,oneof=1 2"` // 状态不用指针
    Name     string `json:"name" binding:"omitempty"`             // 姓名（非指针，空字符串不更新）
    Remark   string `json:"remark" binding:"omitempty"`           // 备注（非指针，空字符串不更新）
}
```

**Service 层更新判断：**

```go
// 准备更新数据
updates := make(map[string]interface{})
if req.Password != "" {
    updates["password"] = hashedPassword
}
if req.Phone != "" {
    updates["phone"] = &req.Phone
}
if req.Email != "" {
    updates["email"] = &req.Email
}
if req.Status != 0 {
    updates["status"] = req.Status
}
```

---

## 分页规范

### 使用泛型分页

项目使用泛型分页，确保类型安全：

```go
// pkg/pagination/pagination.go
type Request struct {
    Page     int `form:"page" json:"page" binding:"omitempty,min=1"`
    PageSize int `form:"page_size" json:"page_size" binding:"omitempty,min=1,max=100"`
}

type Response[T any] struct {
    List      []T   `json:"list"`
    Page      int   `json:"page"`
    PageSize  int   `json:"page_size"`
    Total     int64 `json:"total"`
    TotalPage int64 `json:"total_page"`
}

func NewResponse[T any](list []T, r *Request, total int64) Response[T]
```

### DTO 定义

```go
// 列表请求 - 内嵌 pagination.Request
type ListUsersRequest struct {
    pagination.Request
    UserName string `form:"username" binding:"omitempty"`         // 用户名模糊搜索
    Status   int    `form:"status" binding:"omitempty,oneof=1 2"` // 状态筛选（非指针类型）
}

// 列表响应 - 使用泛型 Response
type ListUsersResponse struct {
    Response pagination.Response[*UserResponse]
}
```

**注意：**
- 使用 `pagination.Request` 而非 `pagination.PageRequest`
- 状态字段使用 `int` 而非 `*int`（零值用于表示"不筛选"）
- 响应结构体字段名为 `Response` 而非内嵌

### Service 层使用

```go
func (s *UserService) ListUsers(ctx context.Context, req *dto.ListUsersRequest) (*dto.ListUsersResponse, error) {
    users, total, err := s.userRepo.ListWithFilters(ctx, req.GetOffset(), req.GetLimit(), req.UserName, req.Status)
    if err != nil {
        return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户列表失败", err)
    }

    userResponses := make([]*dto.UserResponse, len(users))
    for i, user := range users {
        userResponses[i] = s.toUserResponse(ctx, user)
    }

    return &dto.ListUsersResponse{
        Response: pagination.NewResponse(userResponses, &req.Request, total),
    }, nil
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

在 Service 层的关键操作（创建、更新、删除）中记录操作日志。

**重要：资源类型应使用常量定义，避免硬编码字符串。**

资源类型常量定义在 `pkg/constants/operation_log.go` 中：

```go
// 资源类型常量（用于操作日志记录）
const (
    ResourceTypeUser       = "user"       // 用户资源
    ResourceTypeRole       = "role"       // 角色资源
    ResourceTypePermission = "permission" // 权限资源
    ResourceTypeTenant     = "tenant"     // 租户资源
    ResourceTypeMenu       = "menu"       // 菜单资源
    ResourceTypeDict       = "dict"       // 字典资源
)
```

**使用示例（以租户管理为例）：**

```go
import (
    "admin/pkg/constants"
    "admin/pkg/operationlog"
)

// 创建操作
func (s *TenantService) CreateTenant(ctx context.Context, req *dto.CreateTenantRequest) (*dto.TenantResponse, error) {
    // ... 业务逻辑 ...

    // 创建成功后记录操作日志
    ctx = operationlog.RecordCreate(ctx, constants.ModuleTenant, constants.ResourceTypeTenant, tenant.TenantID, tenant.Name, tenant)

    return s.toTenantResponse(tenant), nil
}

// 更新操作
func (s *TenantService) UpdateTenant(ctx context.Context, tenantID string, req *dto.UpdateTenantRequest) (*dto.TenantResponse, error) {
    // 获取旧值用于日志
    oldTenant, err := s.Repo.GetByID(ctx, tenantID)
    // ... 更新逻辑 ...

    // 获取更新后的数据
    updatedTenant, err := s.Repo.GetByID(ctx, tenantID)

    // 记录操作日志
    ctx = operationlog.RecordUpdate(ctx, constants.ModuleTenant, constants.ResourceTypeTenant, updatedTenant.TenantID, updatedTenant.Name, oldTenant, updatedTenant)

    return s.toTenantResponse(updatedTenant), nil
}

// 删除操作
func (s *TenantService) DeleteTenant(ctx context.Context, tenantID string) error {
    // 获取删除前的数据用于日志
    tenant, err := s.Repo.GetByID(ctx, tenantID)
    // ... 删除逻辑 ...

    // 记录操作日志（注意：删除操作不需要返回 ctx）
    operationlog.RecordDelete(ctx, constants.ModuleTenant, constants.ResourceTypeTenant, tenant.TenantID, tenant.Name, tenant)

    return nil
}
```

**操作日志参数说明：**
- `ctx`: 上下文
- `module`: 模块常量（如 `constants.ModuleTenant`）
- `resourceType`: 资源类型常量（如 `constants.ResourceTypeTenant`）
- `resourceID`: 资源 ID
- `resourceName`: 资源名称（用于日志展示）
- `oldData`: 旧数据（更新/删除时需要）
- `newData`: 新数据（创建/更新时需要）

**新增资源类型常量时：**

如果需要为新的资源类型添加常量，请在 `pkg/constants/operation_log.go` 中的 `ResourceType` 常量组中添加：

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
