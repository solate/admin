# 审计日志系统

## 概述

审计日志系统用于记录系统中的关键操作，包括用户认证（登录/登出）和业务操作（创建/更新/删除/查询/导出）。

**核心设计**：Handler 层记录 + Service 层保持纯净

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
│ 1. 调用 Service 执行业务                            │
│ 2. 根据业务结果记录操作日志                         │
│    recorder.Log(ctx, WithCreate(...), ...)          │
└─────────────────────────────────────────────────────┘
    ↓
┌─────────────────────────────────────────────────────┐
│ audit.Recorder.Log()                                │
│ 1. 应用 LogOption 构建 LogEntry                     │
│ 2. 从 context 自动提取用户/HTTP 信息                │
│ 3. 异步写入数据库（context.Background()）           │
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

使用独立的 `context.Background()` 异步写入，避免被请求取消影响：

```go
func (r *Recorder) Log(ctx context.Context, opts ...LogOption) {
    entry := &LogEntry{...}

    // 提取 context 信息到 entry
    fillFromContext(ctx, entry)

    // 异步写入，所有信息已在 entry 中
    go r.db.Write(context.Background(), entry)
}
```

**为什么用 `context.Background()`**：
- 所有需要的信息已经提取到 `entry` 中
- 写入操作不应受原请求生命周期影响
- 不需要复制 context 中的值

### 3. Service 层保持纯净

Service 层不依赖 audit 包，返回业务数据供 Handler 记录：

```go
// Service 返回业务数据
func (s *UserService) Update(ctx context.Context, userID string, req *dto.UpdateUserRequest) (*model.User, *model.User, error) {
    oldUser, _ := s.userRepo.GetByID(ctx, userID)
    newUser := &model.User{...}
    s.userRepo.Update(ctx, userID, newUser)
    return oldUser, newUser, nil
}

// Handler 记录日志
func (h *UserHandler) Update(c *gin.Context) {
    oldUser, newUser, err := h.service.Update(c.Request.Context(), userID, &req)
    if err != nil {
        h.recorder.Log(c.Request.Context(), audit.WithUpdate(audit.ModuleUser), audit.WithError(err))
        return
    }
    h.recorder.Log(c.Request.Context(),
        audit.WithUpdate(audit.ModuleUser),
        audit.WithResource(audit.ResourceUser, userID, newUser.UserName),
        audit.WithValue(oldUser, newUser),
    )
}
```

## API 使用

### 核心方法

```go
// Log - 通用日志记录
func (r *Recorder) Log(ctx context.Context, opts ...LogOption)

// Login - 记录登录
func (r *Recorder) Login(ctx context.Context, tenantID, userID, userName string, err error)

// Logout - 记录登出
func (r *Recorder) Logout(ctx context.Context)
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
WithLogin()                    // 登录
WithLogout()                   // 登出

// 资源和值
WithResource(type, id, name string)
WithValue(oldValue, newValue any)
WithError(err error)

// 基础选项
WithModule(module string)
WithOperation(op string)
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
    ModuleAuth       = "auth"
    ModuleUser       = "user"
    ModuleRole       = "role"
    ModuleMenu       = "menu"
    ModuleTenant     = "tenant"
    ModuleDepartment = "department"
    ModulePosition   = "position"
    ModuleDict       = "dict"
    ModuleSystem     = "system"
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

1. **Service 层纯净**：Service 不依赖 audit 包，返回数据供 Handler 记录
2. **异步写入**：使用 `context.Background()` 异步写入，不影响业务
3. **参数脱敏**：中间件自动对 password、secret 等敏感字段脱敏
4. **信息已提取**：写入时所有信息已在 entry 中，无需传递 context 值
