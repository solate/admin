# 通用审计日志系统设计

> **设计目标**: 构建一套简单、通用、易用的审计日志系统，支持登录日志和操作日志的统一管理，方便业务代码调用。

---

## 一、设计原则

1. **简单优先**: API 设计简洁，学习成本低
2. **统一管理**: 登录日志和操作日志使用统一的接口
3. **自动采集**: 自动收集请求信息，减少手动代码
4. **异步写入**: 日志写入不阻塞业务逻辑
5. **租户隔离**: 支持多租户数据隔离
6. **敏感脱敏**: 自动脱敏密码等敏感字段

---

## 二、核心概念

### 2.1 日志类型

| 类型 | 说明 | 示例 |
|------|------|------|
| `LOGIN` | 登录日志 | 用户登录系统 |
| `LOGOUT` | 登出日志 | 用户退出系统 |
| `CREATE` | 创建操作 | 新增用户、角色 |
| `UPDATE` | 更新操作 | 修改用户信息 |
| `DELETE` | 删除操作 | 删除角色 |
| `QUERY` | 查询操作 | 查询敏感数据 |
| `EXPORT` | 导出操作 | 导出数据报表 |

### 2.2 日志状态

| 状态 | 值 | 说明 |
|------|-----|------|
| 成功 | `1` | 操作执行成功 |
| 失败 | `2` | 操作执行失败 |

---

## 三、数据库设计

### 3.1 登录日志表 (login_logs)

```sql
CREATE TABLE login_logs (
    log_id VARCHAR(20) PRIMARY KEY,
    tenant_id VARCHAR(20) NOT NULL,
    user_id VARCHAR(20),
    user_name VARCHAR(100),
    nickname VARCHAR(100),
    login_type VARCHAR(20),
    login_ip VARCHAR(50),
    user_agent VARCHAR(500),
    status SMALLINT NOT NULL DEFAULT 1,
    fail_reason VARCHAR(500),
    created_at BIGINT NOT NULL,
    INDEX idx_tenant_user (tenant_id, user_id),
    INDEX idx_tenant_time (tenant_id, created_at DESC)
);

COMMENT ON TABLE login_logs IS '登录日志表';
COMMENT ON COLUMN login_logs.log_id IS '日志ID';
COMMENT ON COLUMN login_logs.tenant_id IS '租户ID';
COMMENT ON COLUMN login_logs.user_id IS '用户ID';
COMMENT ON COLUMN login_logs.user_name IS '用户账号';
COMMENT ON COLUMN login_logs.nickname IS '用户昵称';
COMMENT ON COLUMN login_logs.login_type IS '登录类型: PASSWORD/SSO/OAUTH';
COMMENT ON COLUMN login_logs.login_ip IS '登录IP';
COMMENT ON COLUMN login_logs.user_agent IS '用户代理';
COMMENT ON COLUMN login_logs.status IS '状态: 1成功 2失败';
COMMENT ON COLUMN login_logs.fail_reason IS '失败原因';
```

### 3.2 操作日志表 (operation_logs)

```sql
CREATE TABLE operation_logs (
    log_id VARCHAR(20) PRIMARY KEY,
    tenant_id VARCHAR(20) NOT NULL,
    user_id VARCHAR(20),
    user_name VARCHAR(100),
    nickname VARCHAR(100),
    module VARCHAR(50),
    operation_type VARCHAR(20),
    resource_type VARCHAR(50),
    resource_id VARCHAR(100),
    resource_name VARCHAR(255),
    request_method VARCHAR(10),
    request_path VARCHAR(500),
    request_params TEXT,
    old_value TEXT,
    new_value TEXT,
    status SMALLINT NOT NULL DEFAULT 1,
    error_message TEXT,
    ip_address VARCHAR(50),
    user_agent VARCHAR(500),
    created_at BIGINT NOT NULL,
    INDEX idx_tenant_user (tenant_id, user_id),
    INDEX idx_tenant_time (tenant_id, created_at DESC),
    INDEX idx_module (tenant_id, module, created_at DESC),
    INDEX idx_resource (resource_type, resource_id)
);

COMMENT ON TABLE operation_logs IS '操作日志表';
COMMENT ON COLUMN operation_logs.log_id IS '日志ID';
COMMENT ON COLUMN operation_logs.tenant_id IS '租户ID';
COMMENT ON COLUMN operation_logs.user_id IS '操作用户ID';
COMMENT ON COLUMN operation_logs.user_name IS '操作用户账号';
COMMENT ON COLUMN operation_logs.nickname IS '操作用户昵称';
COMMENT ON COLUMN operation_logs.module IS '功能模块';
COMMENT ON COLUMN operation_logs.operation_type IS '操作类型: LOGIN/LOGOUT/CREATE/UPDATE/DELETE/QUERY/EXPORT';
COMMENT ON COLUMN operation_logs.resource_type IS '资源类型';
COMMENT ON COLUMN operation_logs.resource_id IS '资源ID';
COMMENT ON COLUMN operation_logs.resource_name IS '资源名称';
COMMENT ON COLUMN operation_logs.request_method IS '请求方法';
COMMENT ON COLUMN operation_logs.request_path IS '请求路径';
COMMENT ON COLUMN operation_logs.request_params IS '请求参数(JSON,脱敏)';
COMMENT ON COLUMN operation_logs.old_value IS '变更前值(JSON)';
COMMENT ON COLUMN operation_logs.new_value IS '变更后值(JSON)';
COMMENT ON COLUMN operation_logs.status IS '状态: 1成功 2失败';
COMMENT ON COLUMN operation_logs.error_message IS '错误信息';
```

---

## 四、后端实现设计

### 4.1 目录结构

```
backend/pkg/audit/
├── types.go              # 类型定义
├── context.go            # Context 操作
├── recorder.go           # 核心记录器
├── middleware.go         # Gin 中间件
└── writer.go             # 数据库写入器
```

### 4.2 类型定义 (types.go)

```go
package audit

// OperationType 操作类型常量
const (
    OperationLogin   = "LOGIN"
    OperationLogout  = "LOGOUT"
    OperationCreate  = "CREATE"
    OperationUpdate  = "UPDATE"
    OperationDelete  = "DELETE"
    OperationQuery   = "QUERY"
    OperationExport  = "EXPORT"
)

// LogStatus 日志状态常量
const (
    StatusSuccess = 1
    StatusFailure = 2
)

// LogEntry 日志条目
type LogEntry struct {
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
    RequestMethod string
    RequestPath   string
    RequestParams string
    IPAddress     string
    UserAgent     string
    Status        int16
    ErrorMessage  string
    CreatedAt     int64
}

// LogContext 日志上下文（存储在 request.Context 中）
type LogContext struct {
    Module        string
    OperationType string
    ResourceType  string
    ResourceID    string
    ResourceName  string
    OldValue      any
    NewValue      any
    Status        int16
    ErrorMessage  string
    CreatedAt     int64
}

// Builder 日志构建器（链式调用）
type Builder struct {
    entry *LogEntry
}
```

### 4.3 Context 操作 (context.go)

```go
package audit

import "context"

type contextKey int

const ctxKeyLogContext contextKey = 0

// WithLogContext 将日志上下文存入 context
func WithLogContext(ctx context.Context, lc *LogContext) context.Context {
    return context.WithValue(ctx, ctxKeyLogContext, lc)
}

// GetLogContext 从 context 获取日志上下文
func GetLogContext(ctx context.Context) (*LogContext, bool) {
    lc, ok := ctx.Value(ctxKeyLogContext).(*LogContext)
    return lc, ok
}
```

### 4.4 核心记录器 (recorder.go)

```go
package audit

import (
    "context"
    "time"

    "admin/pkg/xcontext"
)

// Recorder 审计日志记录器
type Recorder struct {
    writer *Writer
}

// NewRecorder 创建记录器
func NewRecorder(writer *Writer) *Recorder {
    return &Recorder{writer: writer}
}

// Login 记录登录日志（直接写入，不走中间件）
// 注意：登录时通常还没经过认证中间件，需要手动传入完整信息
func (r *Recorder) Login(ctx context.Context, tenantID, userID, userName string, err error) *LoginBuilder {
    return &LoginBuilder{
        recorder: r,
        ctx:      ctx,
        tenantID: tenantID,
        userID:   userID,
        userName: userName,
        err:      err,
    }
}

// LoginBuilder 登录日志构建器（支持链式调用设置 IP 和 UserAgent）
type LoginBuilder struct {
    recorder  *Recorder
    ctx       context.Context
    tenantID  string
    userID    string
    userName  string
    ipAddress string
    userAgent string
    err       error
}

// WithClientInfo 设置客户端信息（IP 和 UserAgent）
func (b *LoginBuilder) WithClientInfo(ip, userAgent string) *LoginBuilder {
    b.ipAddress = ip
    b.userAgent = userAgent
    return b
}

// Write 写入登录日志
func (b *LoginBuilder) Write() {
    entry := &LogEntry{
        TenantID:      b.tenantID,
        UserID:        b.userID,
        UserName:      b.userName,
        Module:        "auth",
        OperationType: OperationLogin,
        IPAddress:     b.ipAddress,
        UserAgent:     b.userAgent,
        Status:        StatusSuccess,
        CreatedAt:     time.Now().UnixMilli(),
    }

    if b.err != nil {
        entry.Status = StatusFailure
        entry.ErrorMessage = b.err.Error()
    }

    // 异步写入
    go b.recorder.writer.Write(b.ctx, entry)
}

// Logout 记录登出日志（直接写入，不走中间件）
func (r *Recorder) Logout(ctx context.Context) *LogoutBuilder {
    return &LogoutBuilder{
        recorder: r,
        ctx:      ctx,
    }
}

// LogoutBuilder 登出日志构建器
type LogoutBuilder struct {
    recorder  *Recorder
    ctx       context.Context
    ipAddress string
    userAgent string
}

// WithClientInfo 设置客户端信息
func (b *LogoutBuilder) WithClientInfo(ip, userAgent string) *LogoutBuilder {
    b.ipAddress = ip
    b.userAgent = userAgent
    return b
}

// Write 写入登出日志
func (b *LogoutBuilder) Write() {
    entry := &LogEntry{
        TenantID:      xcontext.GetTenantID(b.ctx),
        UserID:        xcontext.GetUserID(b.ctx),
        UserName:      xcontext.GetUserName(b.ctx),
        Module:        "auth",
        OperationType: OperationLogout,
        IPAddress:     b.ipAddress,
        UserAgent:     b.userAgent,
        Status:        StatusSuccess,
        CreatedAt:     time.Now().UnixMilli(),
    }

    // 异步写入
    go b.recorder.writer.Write(b.ctx, entry)
}

// Record 记录操作日志（返回 Builder，支持链式调用）
func (r *Recorder) Record(module, operationType string) *Builder {
    return &Builder{
        entry: &LogEntry{
            Module:        module,
            OperationType: operationType,
            Status:        StatusSuccess,
            CreatedAt:     time.Now().UnixMilli(),
        },
    }
}

// RecordCreate 记录创建操作的便捷方法
func (r *Recorder) RecordCreate(module string) *Builder {
    return r.Record(module, OperationCreate)
}

// RecordUpdate 记录更新操作的便捷方法
func (r *Recorder) RecordUpdate(module string) *Builder {
    return r.Record(module, OperationUpdate)
}

// RecordDelete 记录删除操作的便捷方法
func (r *Recorder) RecordDelete(module string) *Builder {
    return r.Record(module, OperationDelete)
}

// RecordQuery 记录查询操作的便捷方法
func (r *Recorder) RecordQuery(module string) *Builder {
    return r.Record(module, OperationQuery)
}

// RecordExport 记录导出操作的便捷方法
func (r *Recorder) RecordExport(module string) *Builder {
    return r.Record(module, OperationExport)
}

// Commit 提交日志（将 LogContext 存入 context）
func (b *Builder) Commit(ctx context.Context) context.Context {
    // 从 context 获取租户和用户信息
    b.entry.TenantID = xcontext.GetTenantID(ctx)
    b.entry.UserID = xcontext.GetUserID(ctx)
    b.entry.UserName = xcontext.GetUserName(ctx)

    lc := &LogContext{
        Module:        b.entry.Module,
        OperationType: b.entry.OperationType,
        ResourceType:  b.entry.ResourceType,
        ResourceID:    b.entry.ResourceID,
        ResourceName:  b.entry.ResourceName,
        OldValue:      b.entry.OldValue,
        NewValue:      b.entry.NewValue,
        Status:        b.entry.Status,
        ErrorMessage:  b.entry.ErrorMessage,
    }

    return WithLogContext(ctx, lc)
}

// WithResource 设置资源信息（链式调用）
func (b *Builder) WithResource(resourceType, resourceID, resourceName string) *Builder {
    b.entry.ResourceType = resourceType
    b.entry.ResourceID = resourceID
    b.entry.ResourceName = resourceName
    return b
}

// WithValue 设置变更值（链式调用）
func (b *Builder) WithValue(oldValue, newValue any) *Builder {
    b.entry.OldValue = oldValue
    b.entry.NewValue = newValue
    return b
}

// WithError 标记操作失败（链式调用）
func (b *Builder) WithError(err error) *Builder {
    if err != nil {
        b.entry.Status = StatusFailure
        b.entry.ErrorMessage = err.Error()
    }
    return b
}
```

### 4.5 数据库写入器 (writer.go)

```go
package audit

import (
    "admin/internal/dal/model"
    "admin/pkg/idgen"
    "context"
    "encoding/json"

    "gorm.io/gorm"
)

// Writer 日志写入器
type Writer struct {
    db *gorm.DB
}

// NewWriter 创建写入器
func NewWriter(db *gorm.DB) *Writer {
    return &Writer{db: db}
}

// Write 写入日志
func (w *Writer) Write(ctx context.Context, entry *LogEntry) error {
    if entry == nil {
        return nil
    }

    switch entry.OperationType {
    case OperationLogin, OperationLogout:
        return w.writeLoginLog(ctx, entry)
    default:
        return w.writeOperationLog(ctx, entry)
    }
}

func (w *Writer) writeLoginLog(ctx context.Context, entry *LogEntry) error {
    logID, _ := idgen.GenerateUUID()

    loginLog := &model.LoginLog{
        LogID:         logID,
        TenantID:      entry.TenantID,
        UserID:        entry.UserID,
        UserName:      entry.UserName,
        LoginType:     entry.Module,
        LoginIP:       entry.IPAddress,
        LoginLocation: "", // 暂不实现 IP 地址解析
        UserAgent:     entry.UserAgent,
        Status:        entry.Status,
        FailReason:    entry.ErrorMessage,
        CreatedAt:     entry.CreatedAt,
    }

    go w.db.WithContext(ctx).Create(loginLog)
    return nil
}

func (w *Writer) writeOperationLog(ctx context.Context, entry *LogEntry) error {
    logID, _ := idgen.GenerateUUID()

    operationLog := &model.OperationLog{
        LogID:         logID,
        TenantID:      entry.TenantID,
        UserID:        entry.UserID,
        UserName:      entry.UserName,
        Module:        entry.Module,
        OperationType: entry.OperationType,
        ResourceType:  entry.ResourceType,
        ResourceID:    entry.ResourceID,
        ResourceName:  entry.ResourceName,
        RequestMethod: entry.RequestMethod,
        RequestPath:   entry.RequestPath,
        RequestParams: entry.RequestParams,
        OldValue:      toJSON(entry.OldValue),
        NewValue:      toJSON(entry.NewValue),
        Status:        entry.Status,
        ErrorMessage:  entry.ErrorMessage,
        IPAddress:     entry.IPAddress,
        UserAgent:     entry.UserAgent,
        CreatedAt:     entry.CreatedAt,
    }

    go w.db.WithContext(ctx).Create(operationLog)
    return nil
}

func toJSON(v any) string {
    if v == nil {
        return ""
    }
    data, _ := json.Marshal(v)
    return string(data)
}
```

### 4.6 Gin 中间件 (middleware.go)

```go
package audit

import (
    "admin/pkg/bodyreader"
    "admin/pkg/database"
    "admin/pkg/xcontext"
    "context"

    "github.com/gin-gonic/gin"
)

// 客户端信息 key（用于在 gin.Context 中存储）
const (
    ctxKeyClientIP    = "client_ip"
    ctxKeyUserAgent   = "user_agent"
    ctxKeyRequestParams = "request_params"
)

// AuditMiddleware 审计日志中间件
// 功能：
// 1. 提前提取请求参数（防止 Body 被读取后无法恢复）
// 2. 提取并存储客户端信息（IP、UserAgent）
// 3. 处理请求后，自动填充日志条目并写入
func AuditMiddleware(writer *Writer) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 提前提取并存储请求参数（带脱敏）
        requestParams := extractParams(c)
        c.Set(ctxKeyRequestParams, requestParams)

        // 2. 提前提取并存储客户端信息
        c.Set(ctxKeyClientIP, c.ClientIP())
        c.Set(ctxKeyUserAgent, c.Request.UserAgent())

        // 3. 处理请求
        c.Next()

        // 4. 检查是否有日志上下文
        lc, exists := GetLogContext(c.Request.Context())
        if !exists {
            return
        }

        // 5. 跳过登录/登出（由 Recorder 直接写入）
        if lc.OperationType == OperationLogin || lc.OperationType == OperationLogout {
            return
        }

        // 6. 构建日志条目（中间件自动填充所有 HTTP 相关信息）
        entry := &LogEntry{
            // 用户信息（从 context 获取，由 AuthMiddleware 注入）
            TenantID: xcontext.GetTenantID(c.Request.Context()),
            UserID:   xcontext.GetUserID(c.Request.Context()),
            UserName: xcontext.GetUserName(c.Request.Context()),

            // 日志上下文信息（由业务代码设置）
            Module:        lc.Module,
            OperationType: lc.OperationType,
            ResourceType:  lc.ResourceType,
            ResourceID:    lc.ResourceID,
            ResourceName:  lc.ResourceName,
            OldValue:      lc.OldValue,
            NewValue:      lc.NewValue,
            Status:        lc.Status,
            ErrorMessage:  lc.ErrorMessage,
            CreatedAt:     lc.CreatedAt,

            // HTTP 请求信息（由中间件自动填充）
            RequestMethod: c.Request.Method,
            RequestPath:   c.Request.URL.Path,
            RequestParams: requestParams,
            IPAddress:     c.GetString(ctxKeyClientIP),
            UserAgent:     c.GetString(ctxKeyUserAgent),
        }

        // 7. 检查响应错误（优先使用业务代码设置的错误）
        if len(c.Errors) > 0 && entry.Status == StatusSuccess {
            entry.Status = StatusFailure
            entry.ErrorMessage = c.Errors.Last().Error()
        }

        // 8. 异步写入（使用独立 context，避免随请求取消）
        ctx := database.SkipTenantCheck(context.Background())
        go writer.Write(ctx, entry)
    }
}

// extractParams 提取请求参数（带脱敏）
func extractParams(c *gin.Context) string {
    if c.Request.Method == "GET" {
        return c.Request.URL.RawQuery
    }
    bodyStr, restoredBody := bodyreader.ReadBodyString(c.Request.Body)
    if restoredBody != nil {
        c.Request.Body = restoredBody
    }
    return bodyreader.SanitizeParams(bodyStr)
}

// GetClientIP 从 gin.Context 获取客户端 IP（供登录/登出使用）
func GetClientIP(c *gin.Context) string {
    if ip, exists := c.Get(ctxKeyClientIP); exists {
        return ip.(string)
    }
    return c.ClientIP()
}

// GetUserAgent 从 gin.Context 获取 UserAgent（供登录/登出使用）
func GetUserAgent(c *gin.Context) string {
    if ua, exists := c.Get(ctxKeyUserAgent); exists {
        return ua.(string)
    }
    return c.Request.UserAgent()
}
```

---

## 五、业务代码使用

### 5.1 初始化

```go
// 在 main.go 或 router 初始化时
writer := audit.NewWriter(db)
recorder := audit.NewRecorder(writer)
router.Use(audit.AuditMiddleware(writer))
```

### 5.2 记录登录/登出

```go
func (s *AuthService) Login(ctx context.Context, c *gin.Context, req *LoginRequest) (*LoginResponse, error) {
    // ... 验证用户名密码 ...
    user, passwordErr := s.validateUser(req)

    if passwordErr != nil {
        // 记录登录失败
        s.recorder.Login(ctx, tenantID, userID, userName, passwordErr).
            WithClientInfo(audit.GetClientIP(c), audit.GetUserAgent(c)).
            Write()
        return nil, passwordErr
    }

    // 记录登录成功
    s.recorder.Login(ctx, tenantID, userID, userName, nil).
        WithClientInfo(audit.GetClientIP(c), audit.GetUserAgent(c)).
        Write()

    return resp, nil
}

func (s *AuthService) Logout(ctx context.Context, c *gin.Context) error {
    // ... 登出业务逻辑 ...

    // 记录登出（从 context 获取用户信息）
    s.recorder.Logout(ctx).
        WithClientInfo(audit.GetClientIP(c), audit.GetUserAgent(c)).
        Write()

    return nil
}
```

**说明**：
- 登录时还没认证，需要手动传入 `tenantID, userID, userName`
- 登出时已认证，从 context 自动获取用户信息
- 使用 `WithClientInfo()` 链式调用设置客户端信息
- 调用 `Write()` 触发异步写入

### 5.3 记录操作日志

操作日志走中间件，**业务代码无需关心 IP、UserAgent、请求路径等 HTTP 信息**，中间件会自动填充。

#### 方式一：链式调用（推荐）

```go
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) error {
    // ... 创建用户逻辑 ...
    user, err := s.repo.Create(ctx, req)

    // 如果出错，记录失败日志
    if err != nil {
        s.recorder.RecordCreate("user").WithError(err).Commit(ctx)
        return err
    }

    // 记录成功日志（只需要指定业务相关信息）
    ctx = s.recorder.RecordCreate("user").
        WithResource("user", user.UserID, user.UserName).
        WithValue(nil, user).
        Commit(ctx)

    return nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *UpdateUserRequest) error {
    // 获取旧数据
    oldUser, _ := s.repo.GetByID(ctx, req.UserID)

    // ... 更新逻辑 ...
    newUser, err := s.repo.Update(ctx, req)

    // 记录操作日志（无论成功失败）
    builder := s.recorder.RecordUpdate("user").
        WithResource("user", req.UserID, req.UserName).
        WithValue(oldUser, newUser)

    if err != nil {
        builder.WithError(err).Commit(ctx)
        return err
    }

    builder.Commit(ctx)
    return nil
}
```

#### 方式二：简洁写法

```go
func (s *RoleService) DeleteRole(ctx context.Context, roleID string) error {
    role, _ := s.repo.GetByID(ctx, roleID)

    // ... 删除逻辑 ...

    // 记录删除（一行代码）
    ctx = s.recorder.RecordDelete("role").
        WithResource("role", roleID, role.RoleName).
        Commit(ctx)

    return nil
}
```

### 5.4 使用示例总结

```go
// ===== 登录/登出（不走中间件，需要手动设置客户端信息） =====

// 登录成功
recorder.Login(ctx, tenantID, userID, userName, nil).
    WithClientInfo(ip, userAgent).
    Write()

// 登录失败
recorder.Login(ctx, tenantID, userID, userName, err).
    WithClientInfo(ip, userAgent).
    Write()

// 登出（从 context 自动获取用户信息）
recorder.Logout(ctx).
    WithClientInfo(ip, userAgent).
    Write()

// ===== 操作日志（走中间件，自动填充 HTTP 信息） =====

// 创建资源
ctx = recorder.RecordCreate("user").
    WithResource("user", userID, userName).
    WithValue(nil, newUser).
    Commit(ctx)

// 更新资源
ctx = recorder.RecordUpdate("user").
    WithResource("user", userID, userName).
    WithValue(oldUser, newUser).
    Commit(ctx)

// 删除资源
ctx = recorder.RecordDelete("role").
    WithResource("role", roleID, roleName).
    Commit(ctx)

// 查询敏感数据
ctx = recorder.RecordQuery("user").Commit(ctx)

// 导出数据
ctx = recorder.RecordExport("report").Commit(ctx)

// 记录失败
ctx = recorder.RecordUpdate("user").
    WithResource("user", userID, userName).
    WithError(err).
    Commit(ctx)
```

---

## 六、API 设计

### 6.1 登录日志接口

```
GET    /api/v1/audit/login/logs          分页查询登录日志
GET    /api/v1/audit/login/stats         登录统计（按日期、按用户）
```

### 6.2 操作日志接口

```
GET    /api/v1/audit/operation/logs      分页查询操作日志
GET    /api/v1/audit/operation/stats     操作统计（按模块、按类型）
GET    /api/v1/audit/operation/user/:id  查询用户操作历史
```

---

## 七、查询服务设计

```go
// internal/service/audit_service.go
type AuditService struct {
    repo *repository.AuditLogRepo
}

// ListLoginLogs 查询登录日志
func (s *AuditService) ListLoginLogs(ctx context.Context, req *ListLoginLogsRequest) (*ListLoginLogsResponse, error)

// ListOperationLogs 查询操作日志
func (s *AuditService) ListOperationLogs(ctx context.Context, req *ListOperationLogsRequest) (*ListOperationLogsResponse, error)

// GetLoginStats 登录统计
func (s *AuditService) GetLoginStats(ctx context.Context, req *StatsRequest) (*StatsResponse, error)

// GetOperationStats 操作统计
func (s *AuditService) GetOperationStats(ctx context.Context, req *StatsRequest) (*StatsResponse, error)

// GetUserActivity 获取用户活动时间线（合并登录和操作日志）
func (s *AuditService) GetUserActivity(ctx context.Context, userID string, limit int) ([]*ActivityItem, error)
```

---

## 八、前端设计

### 8.1 API 封装

```typescript
// src/api/audit.ts
export const auditApi = {
  // 登录日志
  getLoginLogs: (params: ListParams) => request.get('/audit/login/logs', { params }),
  getLoginStats: (params: StatsParams) => request.get('/audit/login/stats', { params }),

  // 操作日志
  getOperationLogs: (params: ListParams) => request.get('/audit/operation/logs', { params }),
  getOperationStats: (params: StatsParams) => request.get('/audit/operation/stats', { params }),
  getUserActivity: (userId: string, limit: number) => request.get(`/audit/operation/user/${userId}`, { params: { limit } }),
}
```

### 8.2 页面路由

```
/audit/login                    # 登录日志页面
/audit/operation                # 操作日志页面
/audit/stats                    # 统计分析页面
```

### 8.3 页面组件

```
views/audit/
├── login/
│   └── index.vue               # 登录日志列表
├── operation/
│   └── index.vue               # 操作日志列表
└── stats/
    └── index.vue               # 统计分析（图表展示）
```

---

## 九、与其他方案对比

| 特性 | 本方案 | 原方案 |
|------|--------|--------|
| API 复杂度 | 简单（链式调用） | 中等（多个函数） |
| 学习成本 | 低 | 中 |
| 代码侵入 | 低 | 中 |
| 扩展性 | 好 | 好 |
| 类型安全 | 是 | 是 |

---

## 十、实施步骤

1. 创建数据库迁移文件
2. 实现 pkg/audit 包
3. 更新中间件配置
4. 更新业务代码
5. 实现查询接口和前端页面
