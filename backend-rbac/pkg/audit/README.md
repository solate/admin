# 审计日志系统

## 概述

审计日志系统用于记录系统中的关键操作，包括用户认证（登录/登出）和业务操作（创建/更新/删除/查询/导出）。

**核心设计**：Service 层记录审计日志 + 返回 DTO

## 设计原则

### 为什么选择 Service 层记录？

根据业界实践和架构设计原则：

1. **业务逻辑归属**：审计日志是业务逻辑的一部分，应与业务操作在同一层
2. **多入口点一致性**：Service 可能被 HTTP、CLI、gRPC 等调用，在 Service 记录可保证一致性
3. **接口清晰**：Service 返回 DTO，Handler 只负责 HTTP 处理
4. **代码简洁**：避免 Handler 层充斥重复的日志记录代码

### 业界实践

参考来源：
- [Building an Audit Log System for a Go Application](https://medium.com/@alameerashraf/building-an-audit-log-system-for-a-go-application-ce131dc21394)
- [Implementing Audit Log using GORM in Go](https://aayushacharya.com.np/blog/audit-log-gorm/)
- [WorkOS Audit Trail](https://pkg.go.dev/github.com/workos/workos-go/pkg/audittrail)

主流方案：
1. **GORM Callback**：数据层自动记录（适合字段级审计）
2. **Service 层记录**：业务层记录（推荐，适合业务级审计）← 本项目采用
3. **Handler 层记录**：较少使用，通常导致代码冗余

## 架构流程

```
HTTP Request
    ↓
┌─────────────────────────────────────────────────────┐
│ Middleware (AuditMiddleware)                        │
│ 提取 IP、UserAgent、Method、Path、Params            │
│ 存入 context.Context                                │
└─────────────────────────────────────────────────────┘
    ↓
┌─────────────────────────────────────────────────────┐
│ Handler 层                                          │
│ 1. 参数绑定 (BindJSON/BindQuery)                    │
│ 2. 调用 Service 执行业务                            │
│ 3. 返回 HTTP 响应                                   │
└─────────────────────────────────────────────────────┘
    ↓
┌─────────────────────────────────────────────────────┐
│ Service 层                                          │
│ 1. 执行业务逻辑                                      │
│ 2. 记录审计日志 recorder.Log(...)                   │
│ 3. 返回 DTO                                         │
└─────────────────────────────────────────────────────┘
    ↓
┌─────────────────────────────────────────────────────┐
│ audit.Recorder.Log()                                │
│ 1. 应用 LogOption 构建 LogEntry                     │
│ 2. 从 context 自动提取用户/HTTP 信息                │
│ 3. 异步写入数据库（xcontext.CopyContext）           │
└─────────────────────────────────────────────────────┘
```

## 关键设计点

### 1. 信息自动提取

`Recorder.Log()` 从 context 自动提取以下信息：

```go
// 用户信息（从 xcontext 提取）
entry.TenantID = xcontext.GetTenantID(ctx)
entry.UserID = xcontext.GetUserID(ctx)
entry.UserName = xcontext.GetUserName(ctx)

// 客户端信息（从 audit context 提取）
if clientInfo := GetClientInfo(ctx); clientInfo != nil {
    entry.IPAddress = clientInfo.IP
    entry.UserAgent = clientInfo.UserAgent
}

// 请求信息（从 audit context 提取）
if reqInfo := GetRequestInfo(ctx); reqInfo != nil {
    entry.RequestMethod = reqInfo.Method
    entry.RequestPath = reqInfo.Path
    entry.RequestParams = reqInfo.Params
}
```

### 2. 异步写入

使用 `xcontext.CopyContext(ctx)` 异步写入，避免被请求取消影响，同时保留租户和用户信息：

```go
func (r *Recorder) Log(ctx context.Context, opts ...LogOption) {
    entry := &LogEntry{...}

    // 提取 context 信息到 entry
    fillFromContext(ctx, entry)

    // 异步写入，保留租户和用户信息供 GORM 租户插件使用
    go r.db.Write(xcontext.CopyContext(ctx), entry)
}
```

**为什么用 `xcontext.CopyContext(ctx)`**：
- 所有需要的信息已经提取到 `entry` 中
- 写入操作不应受原请求生命周期影响
- 需要保留租户和用户信息，供 GORM 租户插件在写入时使用

### 3. Service 层设计规范

Service 层注入 recorder，在业务操作完成后记录审计日志，返回 DTO：

```go
// === Service 层：注入 recorder，返回 DTO ===

type UserService struct {
    userRepo *repository.UserRepo
    recorder *audit.Recorder  // 注入 recorder
}

func (s *UserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
    // 1. 执行业务逻辑
    user := &model.User{...}
    if err := s.userRepo.Create(ctx, user); err != nil {
        // 2. 记录失败日志
        s.recorder.Log(ctx,
            audit.WithCreate(audit.ModuleUser),
            audit.WithError(err),
        )
        return nil, err
    }

    // 3. 记录成功日志
    s.recorder.Log(ctx,
        audit.WithCreate(audit.ModuleUser),
        audit.WithResource(audit.ResourceUser, user.UserID, user.UserName),
        audit.WithValue(nil, user),
    )

    // 4. 返回 DTO
    return s.toUserResponse(ctx, user), nil
}

func (s *UserService) UpdateUser(ctx context.Context, userID string, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
    // 1. 获取旧数据
    oldUser, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        return nil, err
    }

    // 2. 执行更新
    if err := s.userRepo.Update(ctx, userID, updates); err != nil {
        s.recorder.Log(ctx,
            audit.WithUpdate(audit.ModuleUser),
            audit.WithError(err),
        )
        return nil, err
    }

    // 3. 获取新数据
    newUser, _ := s.userRepo.GetByID(ctx, userID)

    // 4. 记录变更日志
    s.recorder.Log(ctx,
        audit.WithUpdate(audit.ModuleUser),
        audit.WithResource(audit.ResourceUser, newUser.UserID, newUser.UserName),
        audit.WithValue(oldUser, newUser),
    )

    return s.toUserResponse(ctx, newUser), nil
}

// === Handler 层：简洁，只负责 HTTP 处理 ===

func (h *UserHandler) CreateUser(c *gin.Context) {
    var req dto.CreateUserRequest
    if err := c.BindJSON(&req); err != nil {
        response.Error(c, err)
        return
    }

    resp, err := h.userService.CreateUser(c.Request.Context(), &req)
    if err != nil {
        response.Error(c, err)
        return
    }

    response.Success(c, resp)
}
```

**关键点**：
- Service 返回 `*dto.UserResponse`，接口清晰
- Service 注入 `*audit.Recorder`，在业务操作完成后记录
- Handler 层简洁，只负责参数绑定和响应
- 审计日志与业务逻辑在同一处，易于维护

### 4. Defer 模式（推荐）

为了避免在每个成功/失败分支重复编写审计日志代码，推荐使用 `defer` 模式统一处理：

```go
// === Service 层：使用 defer 模式统一处理审计日志 ===

func (s *TenantService) CreateTenant(ctx context.Context, req *dto.TenantCreateRequest) (resp *dto.TenantResponse, err error) {
    var tenant *model.Tenant

    defer func() {
        if err != nil {
            // 失败时记录错误
            s.recorder.Log(ctx,
                audit.WithCreate(audit.ModuleTenant),
                audit.WithError(err),
            )
        } else if tenant != nil {
            // 成功时记录资源
            s.recorder.Log(ctx,
                audit.WithCreate(audit.ModuleTenant),
                audit.WithResource(audit.ResourceTenant, tenant.TenantID, tenant.Name),
                audit.WithValue(nil, tenant),
            )
        }
    }()

    // 检查租户编码是否已存在
    exists, err := s.tenantRepo.CheckExists(ctx, req.Code)
    if err != nil {
        return nil, xerr.Wrap(xerr.ErrInternal.Code, "检查租户编码失败", err)
    }
    if exists {
        return nil, xerr.New(xerr.ErrConflict.Code, "租户编码已存在")
    }

    // 生成租户ID
    tenantID, err := idgen.GenerateUUID()
    if err != nil {
        return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成租户ID失败", err)
    }

    // 构建租户模型
    tenant = &model.Tenant{
        TenantID:    tenantID,
        TenantCode:  req.Code,
        Name:        req.Name,
        Description: req.Description,
        Status:      1,
    }

    // 创建租户
    if err := s.tenantRepo.Create(ctx, tenant); err != nil {
        return nil, xerr.Wrap(xerr.ErrInternal.Code, "创建租户失败", err)
    }

    return s.toTenantResponse(tenant), nil
}

// 更新操作示例
func (s *TenantService) UpdateTenant(ctx context.Context, tenantID string, req *dto.TenantUpdateRequest) (resp *dto.TenantResponse, err error) {
    var oldTenant, newTenant *model.Tenant

    defer func() {
        if err != nil {
            s.recorder.Log(ctx,
                audit.WithUpdate(audit.ModuleTenant),
                audit.WithError(err),
            )
        } else if newTenant != nil {
            s.recorder.Log(ctx,
                audit.WithUpdate(audit.ModuleTenant),
                audit.WithResource(audit.ResourceTenant, newTenant.TenantID, newTenant.Name),
                audit.WithValue(oldTenant, newTenant),
            )
        }
    }()

    // 获取旧租户信息
    oldTenant, err = s.tenantRepo.GetByID(ctx, tenantID)
    if err != nil {
        return nil, err
    }

    // ... 执行更新操作 ...

    // 获取更新后的租户信息
    newTenant, err = s.tenantRepo.GetByID(ctx, tenantID)
    if err != nil {
        return nil, err
    }

    return s.toTenantResponse(newTenant), nil
}
```

**Defer 模式的优势**：

1. **消除重复代码**：不需要在成功和失败分支分别编写 `recorder.Log()`
2. **不易遗漏**：新增错误返回点时，defer 会自动记录，不会忘记
3. **代码简洁**：业务逻辑更清晰，不被审计日志代码干扰
4. **集中管理**：审计日志逻辑集中在一处，易于维护和修改

**关键技术点**：

- **命名返回值**：`(resp *dto.TenantResponse, err error)` - defer 函数能访问返回值
- **局部变量**：`var tenant *model.Tenant` - defer 函数能访问业务数据
- **延迟执行**：函数返回时自动执行，无论成功或失败都会记录
- **条件判断**：通过 `if err != nil` 区分成功/失败场景

## API 使用

### 核心方法

```go
// Log - 通用日志记录
func (r *Recorder) Log(ctx context.Context, opts ...LogOption)

// Login - 记录登录
func (r *Recorder) Login(ctx context.Context, tenantID, userID, userName string, err error)

// Logout - 记录登出
func (r *Recorder) Logout(ctx context.Context)

// RecordCreate - 记录创建操作
func (r *Recorder) RecordCreate(ctx context.Context, module, resourceType, resourceID, resourceName string, newValue any)

// RecordUpdate - 记录更新操作
func (r *Recorder) RecordUpdate(ctx context.Context, module, resourceType, resourceID, resourceName string, oldValue, newValue any)

// RecordDelete - 记录删除操作
func (r *Recorder) RecordDelete(ctx context.Context, module, resourceType, resourceID, resourceName string, oldValue any)

// RecordQuery - 记录查询操作
func (r *Recorder) RecordQuery(ctx context.Context, module, resourceType string)

// RecordExport - 记录导出操作
func (r *Recorder) RecordExport(ctx context.Context, module, resourceType, resourceID, resourceName string)
```

### LogOption 选项

```go
// 操作类型（组合了 Module + OperationType）
WithCreate(module string)      // 创建
WithUpdate(module string)      // 更新
WithDelete(module string)      // 删除
WithQuery(module string)       // 查询
WithExport(module string)      // 导出

// 认证相关
WithLogin()                    // 登录（Module 设置为 LoginTypePassword）
WithLogout()                   // 登出

// 资源和值
WithResource(type, id, name string)
WithValue(oldValue, newValue any)
WithError(err error)

// 用户和租户
WithUser(tenantID, userID, userName string)  // 设置用户信息（用于登录等场景）
WithTenantID(tenantID string)                // 设置租户ID（用于跨租户操作）

// 基础选项
WithModule(module string)
WithOperation(op string)
WithClient(ip, userAgent string)  // 可选，中间件已经自动提取
```

### 使用示例

```go
// 创建操作
h.recorder.Log(ctx,
    audit.WithCreate(audit.ModuleUser),
    audit.WithResource(audit.ResourceUser, user.ID, user.Name),
    audit.WithValue(nil, user),
)

// 更新操作
h.recorder.Log(ctx,
    audit.WithUpdate(audit.ModuleUser),
    audit.WithResource(audit.ResourceUser, user.ID, user.Name),
    audit.WithValue(oldUser, newUser),
)

// 删除操作
h.recorder.Log(ctx,
    audit.WithDelete(audit.ModuleRole),
    audit.WithResource(audit.ResourceRole, roleID, roleName),
    audit.WithValue(role, nil),
)

// 错误记录
h.recorder.Log(ctx,
    audit.WithCreate(audit.ModuleTenant),
    audit.WithError(err),
)
```

## 中间件配置

```go
// internal/router/router.go
import "admin/pkg/audit"

authorized := v1.Group("")
authorized.Use(middleware.AuthMiddleware())
authorized.Use(audit.AuditMiddleware()) // 提取 HTTP 信息
{
    // ... 业务路由 ...
}
```

中间件自动提取：
- `ClientInfo`：IP、UserAgent
- `RequestInfo`：Method、Path、Params（含脱敏）

## 数据结构

### LogEntry

```go
type LogEntry struct {
    // 业务信息（由 LogOption 设置）
    TenantID      string
    UserID        string
    UserName      string
    Module        string
    OperationType string
    ResourceType  string
    ResourceID    string
    ResourceName  string
    OldValue      any
    NewValue      any
    Status        int16
    ErrorMessage  string

    // HTTP 信息（从 context 自动提取）
    RequestMethod string
    RequestPath   string
    RequestParams string
    IPAddress     string
    UserAgent     string

    CreatedAt int64
}
```

### 常量定义

```go
// 操作类型
const (
    OperationLogin  = "LOGIN"
    OperationLogout = "LOGOUT"
    OperationCreate = "CREATE"
    OperationUpdate = "UPDATE"
    OperationDelete = "DELETE"
    OperationQuery  = "QUERY"
    OperationExport = "EXPORT"
)

// 模块类型
const (
    ModuleUser       = "user"
    ModuleRole       = "role"
    ModuleMenu       = "menu"
    ModuleTenant     = "tenant"
    ModuleDepartment = "department"
    ModulePosition   = "position"
    ModuleDict       = "dict"
    ModuleSystem     = "system"
)

// 登录类型（用于登录日志的 login_type 字段）
const (
    LoginTypePassword = "PASSWORD" // 密码登录
    LoginTypeSSO      = "SSO"      // 单点登录
    LoginTypeOAuth    = "OAUTH"    // 第三方登录
)

// 日志状态
const (
    StatusSuccess = 1
    StatusFailure = 2
)

// 资源类型
const (
    ResourceTenant     = "tenant"
    ResourceUser       = "user"
    ResourceRole       = "role"
    ResourceMenu       = "menu"
    ResourceDepartment = "department"
    ResourcePosition   = "position"
    ResourceDict       = "dict"
    ResourceDictItem   = "dict_item"
)
```

## 包结构

```
backend/pkg/audit/
├── types.go       - 常量、LogEntry 定义
├── context.go     - ClientInfo、RequestInfo、context 辅助函数
├── middleware.go  - AuditMiddleware（提取 HTTP 信息）
├── recorder.go    - Recorder、Log/Logout/Login/Record* 方法
├── options.go     - LogOption 类型、With 系列函数
├── writer.go      - DB 写入器（异步写入）
└── README.md      - 本文档
```

## 注意事项

1. **Service 层记录**：Service 注入 `*audit.Recorder`，在业务操作完成后记录审计日志
2. **返回 DTO**：Service 返回 DTO 而非 model，保持接口清晰
3. **使用 Defer 模式**：推荐使用 defer 统一处理成功/失败场景的审计日志，避免重复代码
4. **异步写入**：使用 `xcontext.CopyContext(ctx)` 异步写入，保留租户和用户信息供 GORM 租户插件使用
5. **参数脱敏**：中间件自动对 password、secret 等敏感字段脱敏
6. **登录类型**：`WithLogin()` 会将 Module 设置为 `LoginTypePassword`，用于记录登录方式
7. **跨租户操作**：创建租户等跨租户操作时，使用 `WithTenantID()` 设置租户ID
8. **多入口支持**：Service 层记录确保 HTTP、CLI、gRPC 等不同入口点的审计一致性
