# 日志审计系统设计

## 设计原则

- **分表存储**：登录日志、操作日志、数据变更日志分开存储
- **租户隔离**：每个租户只能查看自己的日志
- **异步写入**：日志写入不影响业务性能
- **敏感数据脱敏**：密码等敏感信息不记录

---

## 数据模型

### 登录日志表 (login_logs)

```sql
CREATE TABLE login_logs (
    log_id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36),
    username VARCHAR(50),
    login_type VARCHAR(20),                  -- PASSWORD, SSO, OAUTH
    login_ip VARCHAR(50),
    login_location VARCHAR(100),             -- IP解析的地理位置
    user_agent VARCHAR(255),
    status TINYINT,                          -- 1:成功 0:失败
    fail_reason VARCHAR(255),                -- 失败原因
    created_at BIGINT,
    INDEX idx_tenant_user (tenant_id, user_id),
    INDEX idx_tenant_time (tenant_id, created_at)
);
```

### 操作日志表 (operation_logs)

```sql
CREATE TABLE operation_logs (
    log_id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36),
    username VARCHAR(50),
    user_real_name VARCHAR(100),             -- 真实姓名
    module VARCHAR(50),                      -- 模块名
    operation_type VARCHAR(20),              -- CREATE, UPDATE, DELETE, QUERY
    resource_type VARCHAR(50),               -- 资源类型
    resource_id VARCHAR(255),                -- 资源ID
    resource_name VARCHAR(255),              -- 资源名称
    request_method VARCHAR(10),              -- GET, POST, PUT, DELETE
    request_path VARCHAR(500),               -- 请求路径
    request_params TEXT,                     -- 请求参数（脱敏）
    old_value TEXT,                          -- 旧值（JSON）
    new_value TEXT,                          -- 新值（JSON）
    status TINYINT,                          -- 1:成功 2:失败
    error_message TEXT,                      -- 错误信息
    ip_address VARCHAR(50),
    user_agent TEXT,
    created_at BIGINT,
    INDEX idx_tenant_user (tenant_id, user_id),
    INDEX idx_tenant_time (tenant_id, created_at),
    INDEX idx_module (tenant_id, module, created_at),
    INDEX idx_resource (resource_type, resource_id)
);
```

### 数据变更日志表 (data_change_logs)

```sql
CREATE TABLE data_change_logs (
    log_id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36),
    username VARCHAR(50),
    table_name VARCHAR(50),                  -- 表名
    record_id VARCHAR(36),                   -- 记录ID
    operation VARCHAR(20),                   -- INSERT, UPDATE, DELETE
    old_value JSON,                          -- 旧值
    new_value JSON,                          -- 新值
    changed_fields TEXT,                     -- 变更字段列表
    ip_address VARCHAR(50),
    created_at BIGINT,
    INDEX idx_tenant_table (tenant_id, table_name),
    INDEX idx_record (table_name, record_id),
    INDEX idx_tenant_time (tenant_id, created_at)
);
```

---

## pkg/operationlog 封装

### 目录结构

```
pkg/operationlog/
├── context.go          # 日志上下文
├── helper.go           # 便捷函数
├── logger.go           # 写入器（根据类型分发）
└── writers/
    ├── writer.go       # Writer 接口
    ├── login.go        # 登录日志写入
    └── operation.go    # 操作日志写入
```

### context.go（保持不变）

```go
package operationlog

type LogContext struct {
    TenantID      string
    Module        string
    OperationType string   // LOGIN, LOGOUT, CREATE, UPDATE, DELETE, QUERY
    ResourceType  string
    ResourceID    string
    ResourceName  string
    OldValue      any
    NewValue      any
    Status        int16   // 1:成功 2:失败
    ErrorMessage  string
    CreatedAt     int64
}

// WithLogContext 存入 context
func WithLogContext(ctx context.Context, lc *LogContext) context.Context

// GetLogContext 获取 LogContext
func GetLogContext(ctx context.Context) (*LogContext, bool)
```

### logger.go（修改：根据类型分发）

```go
package operationlog

type Logger struct {
    loginWriter    writers.Writer
    operationWriter writers.Writer
}

func NewLogger(db *gorm.DB) *Logger {
    return &Logger{
        loginWriter:    writers.NewLoginWriter(db),
        operationWriter: writers.NewOperationWriter(db),
    }
}

// Write 根据 OperationType 分发到不同的 Writer
func (l *Logger) Write(ctx context.Context, entry *LogEntry) error {
    lc := entry.LogContext

    switch lc.OperationType {
    case "LOGIN", "LOGOUT":
        return l.loginWriter.Write(ctx, entry)
    default:
        return l.operationWriter.Write(ctx, entry)
    }
}
```

### writers/writer.go

```go
package writers

import "admin/pkg/operationlog"

type Writer interface {
    Write(ctx context.Context, entry *operationlog.LogEntry) error
}
```

### writers/login.go

```go
package writers

type LoginWriter struct {
    db *gorm.DB
}

func NewLoginWriter(db *gorm.DB) *LoginWriter {
    return &LoginWriter{db: db}
}

func (w *LoginWriter) Write(ctx context.Context, entry *operationlog.LogEntry) error {
    lc := entry.LogContext

    log := &model.LoginLog{
        LogID:        idgen.MustGenerateUUID(),
        TenantID:     entry.TenantID,
        UserID:       entry.UserID,
        Username:     entry.UserName,
        LoginType:    lc.Module,           // 从 Module 获取登录类型
        LoginIP:      entry.IPAddress,
        UserAgent:    entry.UserAgent,
        Status:       lc.Status,
        FailReason:   &lc.ErrorMessage,
        CreatedAt:    lc.CreatedAt,
    }

    go w.db.Create(log)
    return nil
}
```

### writers/operation.go

```go
package writers

type OperationWriter struct {
    db *gorm.DB
}

func NewOperationWriter(db *gorm.DB) *OperationWriter {
    return &OperationWriter{db: db}
}

func (w *OperationWriter) Write(ctx context.Context, entry *operationlog.LogEntry) error {
    lc := entry.LogContext

    log := &model.OperationLog{
        LogID:         idgen.MustGenerateUUID(),
        TenantID:      entry.TenantID,
        Module:        lc.Module,
        OperationType: lc.OperationType,
        ResourceType:  &lc.ResourceType,
        ResourceID:    &lc.ResourceID,
        ResourceName:  &lc.ResourceName,
        UserID:        entry.UserID,
        UserName:      entry.UserName,
        RequestMethod: &entry.RequestMethod,
        RequestPath:   &entry.RequestPath,
        RequestParams: &entry.RequestParams,
        OldValue:      serializeJSON(lc.OldValue),
        NewValue:      serializeJSON(lc.NewValue),
        Status:        lc.Status,
        ErrorMessage:  &lc.ErrorMessage,
        IPAddress:     &entry.IPAddress,
        UserAgent:     &entry.UserAgent,
        CreatedAt:     lc.CreatedAt,
    }

    go w.db.Create(log)
    return nil
}
```

---

## 中间件

### OperationLogMiddleware（保持不变）

```go
func OperationLogMiddleware(logger *operationlog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        requestParams := extractRequestParams(c)
        clientInfo := useragent.GetClientInfo(c.Request)
        c.Set("client_info", clientInfo)
        c.Set("request_params", requestParams)
        c.Next()

        lc, exists := operationlog.GetLogContext(c.Request.Context())
        if !exists {
            return
        }

        entry := &operationlog.LogEntry{
            TenantID:      xcontext.GetTenantID(c.Request.Context()),
            UserID:        xcontext.GetUserID(c.Request.Context()),
            UserName:      xcontext.GetUserName(c.Request.Context()),
            RequestMethod: c.Request.Method,
            RequestPath:   c.Request.URL.Path,
            RequestParams: requestParams,
            IPAddress:     clientInfo.IP,
            UserAgent:     clientInfo.UserAgent,
            LogContext:    lc,
        }

        updateLogStatusFromResponse(c, lc)
        logger.Write(c.Request.Context(), entry)
    }
}
```

---

## 业务代码使用

### 记录登录

```go
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
    // ... 验证逻辑 ...

    // 记录登录日志
    ctx = operationlog.RecordLogin(ctx, user.UserID, user.UserName)
    if err != nil {
        operationlog.RecordError(ctx, err)
    }

    return resp, nil
}
```

### 记录业务操作

```go
func (s *UserService) UpdateUser(ctx context.Context, req *UpdateUserRequest) error {
    // 获取旧数据
    oldUser, _ := s.userRepo.GetByID(ctx, req.UserID)

    // 更新
    user := &User{UserName: req.UserName}
    s.userRepo.Update(ctx, user)

    // 记录操作日志
    ctx = operationlog.RecordUpdate(ctx, "user", "user", req.UserID, req.UserName, oldUser, user)

    return nil
}
```

---

## API 设计

### 登录日志接口

```
GET    /api/v1/logs/login                 获取登录日志列表
GET    /api/v1/logs/login/stats           登录统计
```

### 操作日志接口

```
GET    /api/v1/logs/operation             获取操作日志列表
GET    /api/v1/logs/operation/stats       操作统计
```

### 数据变更接口

```
GET    /api/v1/logs/data-change           获取数据变更日志
GET    /api/v1/logs/data-change/:table/:id  获取记录变更历史
```

---

## Repository 层

```go
// LoginLogRepo
func (r *LoginLogRepo) List(ctx context.Context, req *ListRequest) ([]*LoginLog, int64, error)
func (r *LoginLogRepo) StatsByDate(ctx context.Context, tenantID string, days int) ([]*StatItem, error)

// OperationLogRepo
func (r *OperationLogRepo) List(ctx context.Context, req *ListRequest) ([]*OperationLog, int64, error)
func (r *OperationLogRepo) ListByUser(ctx context.Context, userID string, limit int) ([]*OperationLog, error)

// DataChangeLogRepo
func (r *DataChangeLogRepo) List(ctx context.Context, req *ListRequest) ([]*DataChangeLog, int64, error)
func (r *DataChangeLogRepo) GetByRecord(ctx context.Context, tableName, recordID string) ([]*DataChangeLog, error)
```

---

## 合并查询（用户活动时间线）

```go
func (s *LogService) GetUserTimeline(ctx context.Context, userID string, limit int) ([]*ActivityItem, error) {
    var items []*ActivityItem

    // 并行查询
    var wg sync.WaitGroup
    var mu sync.Mutex

    wg.Add(2)

    go func() {
        defer wg.Done()
        logs, _ := s.loginLogRepo.ListByUser(ctx, userID, limit/2)
        mu.Lock()
        for _, log := range logs {
            items = append(items, &ActivityItem{
                Type:      "login",
                Timestamp: log.CreatedAt,
                Content:   fmt.Sprintf("从 %s 登录", log.LoginIP),
            })
        }
        mu.Unlock()
    }()

    go func() {
        defer wg.Done()
        logs, _ := s.operationLogRepo.ListByUser(ctx, userID, limit/2)
        mu.Lock()
        for _, log := range logs {
            items = append(items, &ActivityItem{
                Type:      "operation",
                Timestamp: log.CreatedAt,
                Content:   fmt.Sprintf("%s %s", log.Module, log.OperationType),
            })
        }
        mu.Unlock()
    }()

    wg.Wait()

    // 按时间排序
    sort.Slice(items, func(i, j int) bool {
        return items[i].Timestamp > items[j].Timestamp
    })

    if len(items) > limit {
        items = items[:limit]
    }

    return items, nil
}
```

---

## 常量定义

```go
package constants

const (
    // 登录类型
    LoginTypePassword = "PASSWORD"
    LoginTypeSSO      = "SSO"
    LoginTypeOAuth    = "OAUTH"

    // 操作类型
    OperTypeLogin   = "LOGIN"
    OperTypeLogout  = "LOGOUT"
    OperTypeCreate  = "CREATE"
    OperTypeUpdate  = "UPDATE"
    OperTypeDelete  = "DELETE"
    OperTypeQuery   = "QUERY"
    OperTypeExport  = "EXPORT"

    // 日志保留天数
    LoginLogRetentionDays      = 90
    OperationLogRetentionDays  = 180
    DataChangeLogRetentionDays = 365
)
```
