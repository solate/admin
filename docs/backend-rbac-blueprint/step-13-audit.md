# Step 13: 操作审计 + 登录日志

## 目标

实现操作审计中间件（记录谁在什么时候对什么资源做了什么）和登录日志（记录登录/登出事件）。

## 前置条件

- Step 06 认证完成
- Step 04 数据模型中有 login_logs 和 operation_logs 表

## 文件清单

```
internal/
├── middleware/
│   └── audit.go               # 审计中间件
├── handler/loginlog/
│   ├── loginlog.go
│   └── query.go               # GET /login-logs
├── handler/operationlog/
│   ├── operationlog.go
│   └── query.go               # GET /operation-logs
├── service/loginlog/
│   ├── loginlog.go
│   └── query.go
├── service/operationlog/
│   ├── operationlog.go
│   └── query.go
├── repository/
│   ├── login_log_repo.go
│   └── operation_log_repo.go
└── dto/
    ├── login_log_dto.go
    └── operation_log_dto.go

pkg/audit/
├── recorder.go                # 审计记录器
└── async_writer.go            # 异步写入（channel + goroutine）
```

## 实现细节

### 1. 审计中间件

```go
func Audit(recorder *audit.Recorder) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Before：记录开始时间
        start := time.Now()

        c.Next()

        // After：记录操作
        recorder.Record(c.Request.Context(), audit.Record{
            UserID:     xcontext.GetUserID(ctx),
            UserName:   xcontext.GetUserName(ctx),
            TenantID:   xcontext.GetTenantID(ctx),
            Method:     c.Request.Method,
            Path:       c.Request.URL.Path,
            Status:     c.Writer.Status(),
            IP:         c.ClientIP(),
            UserAgent:  c.Request.UserAgent(),
            Duration:   time.Since(start).Milliseconds(),
            RequestBody:  truncateBody(c.Request.Body, 4096),
            ResponseCode: getResponseCode(c),
        })
    }
}
```

### 2. 异步写入

```go
// pkg/audit/async_writer.go
type AsyncWriter struct {
    ch  chan *Record
    repo *repository.OperationLogRepo
}

func (w *AsyncWriter) Start() {
    go func() {
        for record := range w.ch {
            // 批量积累后写入（每 100 条或每秒 flush）
            w.repo.BatchCreate(context.Background(), records)
        }
    }()
}

func (w *AsyncWriter) Record(ctx context.Context, record Record) {
    // 非阻塞写入 channel
    select {
    case w.ch <- &record:
    default:
        // channel 满了，丢弃（不阻塞请求）
        log.Warn().Msg("audit channel full, record dropped")
    }
}
```

**设计决策**：
- 审计日志 **异步写入**（channel + goroutine），不阻塞请求
- 请求体截断到 4KB，避免巨大 body 占用存储
- channel 满时丢弃日志（不阻塞业务）
- CopyContext 用于异步 goroutine（保留租户信息）

### 3. 登录日志

```
在 auth service 的 login 方法中记录：
  - login_type: email / phone
  - status: success / failed
  - ip, user_agent, location（可选）
  - fail_reason（如果失败）

在 logout 方法中记录：
  - login_type: logout
```

### 4. 日志查询

```
GET /login-logs?page=1&page_size=10&user_name=xxx&status=1&start_date=xxx&end_date=xxx
GET /operation-logs?page=1&page_size=10&user_name=xxx&method=POST&path=/api/v1/users
```

- 只有 admin 及以上角色可查看
- 租户隔离（WHERE tenant_id = ?）
- 支持时间范围过滤（start_date/end_date，毫秒时间戳）

## 验收标准

- [ ] 每次 API 请求后，operation_logs 表新增一条记录
- [ ] 登录成功后，login_logs 表新增一条记录（status=success）
- [ ] 登录失败后，login_logs 表新增一条记录（status=failed + fail_reason）
- [ ] 审计日志不阻塞请求（异步写入）
- [ ] 审计日志包含正确的用户、租户、路径、方法、状态码
- [ ] `GET /login-logs` 分页查询正确
- [ ] `GET /operation-logs` 分页查询正确
- [ ] 日志查询有租户隔离
- [ ] 请求体截断到 4KB

## AI 协作提示

```
请按 step-13-audit.md 实现审计日志。
关键点：
1. 审计中间件放在 RBAC 之后（只记录已认证的请求）
2. 异步写入用 channel + goroutine，CopyContext 保留租户信息
3. 登录日志在 auth service 中直接写入（不走 channel）
4. 日志查询也要租户隔离
```
