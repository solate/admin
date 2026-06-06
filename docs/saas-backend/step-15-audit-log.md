# Step 15: 审计日志 — 操作日志 / 登录日志

## 目标

实现操作审计和登录日志记录，满足合规要求。操作日志通过中间件自动采集，登录日志在认证流程中记录。

## 前置条件

- Step 06 完成（登录流程可记录登录日志）
- Step 03 完成（中间件链可插入审计中间件）

## 文件清单

```
internal/
├── middleware/
│   └── audit.go               # 审计中间件（自动记录操作日志）
├── handler/audit/
│   ├── handler.go
│   └── query.go               # 查询操作日志 / 登录日志
│   └── dto.go
├── service/audit/
│   ├── service.go
│   ├── recorder.go            # Recorder：异步写入日志
│   └── query.go
├── repository/
│   ├── operation_log_repo.go
│   └── login_log_repo.go
```

## 实现规范

### 1. 审计中间件 — 自动采集

```go
// internal/middleware/audit.go
package middleware

import (
    "bytes"
    "io"
    "time"
    "admin/internal/service/audit"
    "admin/pkg/xcontext"
    "github.com/gin-gonic/gin"
)

func Audit(recorder *audit.Recorder) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 只记录写操作（POST/PUT/DELETE）
        if c.Request.Method == "GET" {
            c.Next()
            return
        }

        start := time.Now()

        // 读取 body（需要 body reader 复用）
        var bodyBytes []byte
        if c.Request.Body != nil {
            bodyBytes, _ = io.ReadAll(c.Request.Body)
            c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
        }

        // 执行请求
        c.Next()

        // 异步记录（不阻塞响应）
        ctx := xcontext.CopyContext(c.Request.Context())
        go recorder.Record(ctx, &audit.OperationEntry{
            UserID:      xcontext.GetUserID(c.Request.Context()),
            UserName:    xcontext.GetUserName(c.Request.Context()),
            TenantID:    xcontext.GetTenantID(c.Request.Context()),
            Method:      c.Request.Method,
            Path:        c.Request.URL.Path,
            IP:          c.ClientIP(),
            UserAgent:   c.Request.UserAgent(),
            RequestBody: string(bodyBytes),
            Status:      c.Writer.Status(),
            Duration:    int(time.Since(start).Milliseconds()),
        })
    }
}
```

**关键设计**：
- 只记录写操作（GET 不记录，避免日志膨胀）
- 异步写入（`go recorder.Record()`），不影响响应延迟
- 使用 `CopyContext` 确保 goroutine 中能获取认证信息
- Request Body 脱敏：密码字段替换为 `***`

### 2. Recorder — 异步批量写入

```go
// internal/service/audit/recorder.go
package audit

import (
    "context"
    "sync"
    "time"
    "admin/internal/repository"
    "github.com/rs/zerolog"
)

type Recorder struct {
    opLogRepo    *repository.OperationLogRepo
    loginLogRepo *repository.LoginLogRepo
    log          zerolog.Logger
    buffer       []*OperationEntry
    mu           sync.Mutex
    flushSize    int
    flushInterval time.Duration
}

type OperationEntry struct {
    UserID      string
    UserName    string
    TenantID    string
    Module      string
    Action      string
    Method      string
    Path        string
    IP          string
    UserAgent   string
    RequestBody string
    Status      int
    Duration    int
}

func NewRecorder(opLogRepo *repository.OperationLogRepo, loginLogRepo *repository.LoginLogRepo, log zerolog.Logger) *Recorder {
    r := &Recorder{
        opLogRepo:     opLogRepo,
        loginLogRepo:  loginLogRepo,
        log:           log,
        buffer:        make([]*OperationEntry, 0, 100),
        flushSize:     50,    // 攒 50 条批量写入
        flushInterval: 5 * time.Second, // 或每 5 秒刷一次
    }
    go r.flushLoop()
    return r
}

// Record 记录操作日志（异步缓冲）
func (r *Recorder) Record(ctx context.Context, entry *OperationEntry) {
    // 自动推断 module 和 action
    entry.Module, entry.Action = inferModuleAction(entry.Method, entry.Path)

    // 脱敏
    entry.RequestBody = sanitizeBody(entry.RequestBody)

    r.mu.Lock()
    r.buffer = append(r.buffer, entry)
    shouldFlush := len(r.buffer) >= r.flushSize
    r.mu.Unlock()

    if shouldFlush {
        r.flush()
    }
}

// RecordLogin 记录登录日志（立即写入，不缓冲）
func (r *Recorder) RecordLogin(ctx context.Context, entry *LoginEntry) {
    if err := r.loginLogRepo.Create(ctx, entry.ToModel()); err != nil {
        r.log.Error().Err(err).Msg("record login log failed")
    }
}

func (r *Recorder) flush() {
    r.mu.Lock()
    if len(r.buffer) == 0 {
        r.mu.Unlock()
        return
    }
    batch := r.buffer
    r.buffer = make([]*OperationEntry, 0, 100)
    r.mu.Unlock()

    // 批量写入数据库
    if err := r.opLogRepo.BatchCreate(context.Background(), batch); err != nil {
        r.log.Error().Err(err).Int("count", len(batch)).Msg("flush operation logs failed")
    }
}

func (r *Recorder) flushLoop() {
    ticker := time.NewTicker(r.flushInterval)
    defer ticker.Stop()
    for range ticker.C {
        r.flush()
    }
}
```

**设计决策**：
- 操作日志：缓冲批量写入（减少 DB 压力）
- 登录日志：立即写入（安全审计需要实时性）
- 脱敏：密码、Token 等敏感字段替换

### 3. 查询接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/operation-logs | 操作日志列表（分页） |
| GET | /api/v1/login-logs | 登录日志列表（分页） |

```go
type ListOperationLogRequest struct {
    dto.PageRequest
    UserName  string `form:"user_name"`
    Module    string `form:"module"`
    StartTime int64  `form:"start_time"` // 毫秒时间戳
    EndTime   int64  `form:"end_time"`
}

type OperationLogInfo struct {
    LogID       string `json:"log_id"`
    UserName    string `json:"user_name"`
    Module      string `json:"module"`
    Action      string `json:"action"`
    Method      string `json:"method"`
    Path        string `json:"path"`
    IP          string `json:"ip"`
    Status      int    `json:"status"`
    Duration    int    `json:"duration"`
    CreatedAt   int64  `json:"created_at"`
}
```

### 4. Module/Action 推断

```go
// 从 HTTP Method + Path 自动推断模块和操作
func inferModuleAction(method, path string) (module, action string) {
    // /api/v1/users → module=user
    // /api/v1/roles/:id/permissions → module=role

    // POST → 创建, PUT → 更新, DELETE → 删除
    switch method {
    case "POST":
        action = "create"
    case "PUT":
        action = "update"
    case "DELETE":
        action = "delete"
    }

    // 从 path 提取 module
    // /api/v1/{module}/... → module
    parts := strings.Split(strings.TrimPrefix(path, "/api/v1/"), "/")
    if len(parts) > 0 {
        module = parts[0]
    }
    return
}
```

### 5. Body 脱敏

```go
func sanitizeBody(body string) string {
    // 限制长度
    if len(body) > 2000 {
        body = body[:2000] + "...[truncated]"
    }
    // 替换敏感字段
    // password, token, secret → "***"
    // 用正则或 JSON 解析后替换
    return body
}
```

## 验收标准

```bash
# 1. 操作日志自动记录
# 执行 POST/PUT/DELETE 操作后
curl -s "http://localhost:8080/api/v1/operation-logs?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN"
# 期望：包含刚才的操作记录

# 2. GET 不记录
# 执行 GET 请求后查日志 → 不包含 GET 记录

# 3. 登录日志
curl -s "http://localhost:8080/api/v1/login-logs?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN"
# 期望：包含登录成功/失败记录

# 4. 脱敏
# 创建用户时 body 中有 password → 日志中显示 ***

# 5. 性能
# 审计中间件不影响接口响应时间（异步写入）

# 6. 租户隔离
# 只能看到本租户的日志
```

## AI 协作提示

```
请按 step-15-audit-log.md 实现审计日志。

要点：
1. internal/middleware/audit.go — 自动采集写操作
2. service/audit/recorder.go — 异步批量写入（操作日志） + 立即写入（登录日志）
3. CopyContext 确保 goroutine 中有认证信息
4. Body 脱敏：密码/token 替换为 ***
5. module/action 从 method+path 自动推断
6. 查询接口支持时间范围 + user_name + module 筛选
7. 只记录 POST/PUT/DELETE，不记录 GET
8. 在 router.go 的 auth 路由组添加 Audit 中间件（在 RBAC 之后）
```

---

*上一步：[Step 14 - 数据权限](step-14-data-permission.md) | 下一步：[Step 16 - Swagger 文档](step-16-swagger.md)*
