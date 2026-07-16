# Step 03: HTTP 框架 — Gin 封装 / 中间件 / 响应 / 错误码

## 目标

引入 Gin 框架，实现统一响应格式、业务错误码、基础中间件链、健康检查端点。

## 前置条件

- Step 02 完成，配置/日志/数据库/Redis 均可初始化

## 文件清单

```
pkg/
├── xerr/
│   ├── codes.go               # 错误码常量定义
│   └── error.go               # BizError struct + New/Wrap/Is
├── response/
│   └── response.go            # OK / Fail / Page 统一响应

internal/
├── server/
│   └── server.go              # HTTP Server 封装（启动 + 优雅关闭）
├── router/
│   ├── router.go              # Setup() 路由注册
│   └── handlers.go            # Handlers 聚合结构体
├── middleware/
│   ├── request_id.go          # X-Request-ID 生成
│   ├── logger.go              # 请求日志（slog）
│   ├── recovery.go            # panic 恢复
│   └── cors.go                # CORS 跨域
├── handler/health/
│   └── health.go              # 健康检查 handler
└── dto/
    └── common.go              # PageRequest / IDRequest / IDsRequest

cmd/server/main.go             # 更新：创建 Gin Engine + 启动 Server
```

## 实现规范

### 1. 错误码体系 — xerr

```go
// pkg/xerr/codes.go
package xerr

// 错误码规则：
// - 0 = 成功
// - 5 位数字：前 3 位 HTTP 状态码 + 后 2 位业务序号
// - 例：40101 = 401 未认证的第 01 号错误

const (
    CodeSuccess = 0

    // 400xx 参数错误
    CodeParamInvalid  = 40001 // 参数校验失败
    CodeParamMissing  = 40002 // 缺少必要参数

    // 401xx 认证错误
    CodeUnauthorized   = 40100 // 未登录/Token 无效
    CodeTokenExpired   = 40101 // Token 过期
    CodeTokenRevoked   = 40102 // Token 已吊销

    // 403xx 权限错误
    CodeForbidden      = 40300 // 无权限
    CodeTenantDisabled = 40301 // 租户已禁用

    // 404xx 未找到
    CodeNotFound       = 40400 // 资源不存在
    CodeRouteNotFound  = 40401 // 路由不存在

    // 409xx 冲突
    CodeConflict       = 40900 // 数据冲突
    CodeDuplicate      = 40901 // 重复数据

    // 500xx 内部错误
    CodeInternal       = 50000 // 服务器内部错误
    CodeDBError        = 50001 // 数据库错误
    CodeRedisError     = 50002 // Redis 错误
)
```

```go
// pkg/xerr/error.go
package xerr

import "fmt"

// BizError 业务错误
type BizError struct {
    Code    int    // 业务错误码
    Message string // 用户可见的错误信息
    Err     error  // 原始错误（日志用，不暴露给前端）
}

func (e *BizError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
    }
    return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *BizError) Unwrap() error { return e.Err }

// New 创建业务错误（无原始错误）
func New(code int, message string) *BizError {
    return &BizError{Code: code, Message: message}
}

// Wrap 包装原始错误为业务错误
func Wrap(code int, message string, err error) *BizError {
    return &BizError{Code: code, Message: message, Err: err}
}

// Is 判断 error 是否为指定错误码
func Is(err error, code int) bool {
    if e, ok := err.(*BizError); ok {
        return e.Code == code
    }
    return false
}
```

### 2. 统一响应 — response

```go
// pkg/response/response.go
package response

import (
    "errors"
    "net/http"
    "admin/pkg/xerr"
    "github.com/gin-gonic/gin"
)

type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

type PageData struct {
    List     interface{} `json:"list"`
    Total    int64       `json:"total"`
    Page     int         `json:"page"`
    PageSize int         `json:"page_size"`
}

// OK 成功响应
func OK(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, Response{
        Code:    xerr.CodeSuccess,
        Message: "success",
        Data:    data,
    })
}

// OKWithMessage 成功响应（自定义消息）
func OKWithMessage(c *gin.Context, message string) {
    c.JSON(http.StatusOK, Response{
        Code:    xerr.CodeSuccess,
        Message: message,
    })
}

// Page 分页响应
func Page(c *gin.Context, list interface{}, total int64, page, pageSize int) {
    c.JSON(http.StatusOK, Response{
        Code:    xerr.CodeSuccess,
        Message: "success",
        Data: PageData{
            List:     list,
            Total:    total,
            Page:     page,
            PageSize: pageSize,
        },
    })
}

// Fail 失败响应（从 error 提取错误码）
func Fail(c *gin.Context, err error) {
    var bizErr *xerr.BizError
    if errors.As(err, &bizErr) {
        c.JSON(http.StatusOK, Response{
            Code:    bizErr.Code,
            Message: bizErr.Message,
        })
        return
    }
    // 未知错误，不暴露细节
    c.JSON(http.StatusOK, Response{
        Code:    xerr.CodeInternal,
        Message: "服务器内部错误",
    })
}
```

**设计决策**：
- HTTP 状态码统一 200，业务错误码在 JSON body 中
- 好处：前端统一处理逻辑，不需要区分 HTTP 状态码
- `Fail` 会自动判断是否为 `BizError`，未知错误不暴露内部细节

### 3. 中间件链

#### RequestID

```go
// internal/middleware/request_id.go
// 为每个请求生成唯一 ID，放入 Header 和 Context
// X-Request-ID: <uuid>
// 如果请求已带 X-Request-ID 则复用
```

#### Logger

```go
// internal/middleware/logger.go
// 记录请求日志：method, path, status, latency, request_id, client_ip
// 使用 slog（*Context 变体 + slog.Attr 强类型字段），一行写完
// 跳过 /health 路径（避免刷屏）
```

#### Recovery

```go
// internal/middleware/recovery.go
// 捕获 panic，记录堆栈，返回 500
// 不暴露堆栈给客户端
// 日志中记录完整 panic 信息
```

#### CORS

```go
// internal/middleware/cors.go
// 允许的 Origin：从配置读取，或开发模式允许 *
// 允许的 Methods：GET, POST, PUT, DELETE, OPTIONS
// 允许的 Headers：Authorization, Content-Type, X-Request-ID
// Max-Age: 86400（24小时缓存预检）
```

### 4. Server 封装

```go
// internal/server/server.go
package server

import (
    "context"
    "fmt"
    "log/slog"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"

    "admin/internal/config"
    "admin/internal/router"
)

// New 用 Options 聚合依赖（Config/DB/RDB/Log）而非位置参数——加字段不破坏调用点，
// 取舍见 bootstrap doc 02。engine 在 New 内部构建：router.Setup 注册中间件与路由。
func New(opts Options) (*Server, error) {
    gin.SetMode(opts.Config.Server.Mode)
    engine := gin.New()
    router.Setup(engine, opts.Config, opts.Log)

    httpSrv := &http.Server{
        Addr:         fmt.Sprintf(":%d", opts.Config.Server.Port),
        Handler:      engine,
        ReadTimeout:  time.Duration(opts.Config.Server.ReadTimeout) * time.Second,
        WriteTimeout: time.Duration(opts.Config.Server.WriteTimeout) * time.Second,
        IdleTimeout:  60 * time.Second,
    }
    return &Server{
        httpSrv:         httpSrv,
        db:              opts.DB,
        rdb:             opts.RDB,
        log:             opts.Log,
        gracefulTimeout: time.Duration(opts.Config.Server.GracefulTimeout) * time.Second,
    }, nil
}

// Run 用 errgroup 编排长驻组件（HTTP + 其带超时优雅关闭；Step 05 起加 scheduler）。
// 完整实现（两个 g.Go、ErrServerClosed 过滤成 nil、g.Wait 收敛）与 bootstrap doc 01
// 第一节逐字一致，此处不复制——doc 01 是生命周期编排的唯一权威版本。
func (s *Server) Run(ctx context.Context) error { /* errgroup 编排，见 bootstrap doc 01 */ }
```

> `Options`/`Server` 字段与 `Run(ctx)` 完整代码见 [bootstrap doc 01](./research/bootstrap/01-应用启动与组件生命周期-main组合根与gin-cron编排.md) 第一节。`main.go` 只认 `Run(ctx)` 一个动词，`Start()/Stop()` 双方法已被 errgroup 编排取代。

### 5. Router Setup

```go
// internal/router/router.go
package router

import (
    "admin/internal/middleware"
    "admin/pkg/response"
    "admin/pkg/xerr"
    "github.com/gin-gonic/gin"
    "log/slog"
)

func Setup(engine *gin.Engine, handlers *Handlers, log *slog.Logger) {
    // 全局中间件
    engine.Use(
        middleware.RequestID(),
        middleware.Logger(log),
        middleware.Recovery(log),
        middleware.CORS(),
    )

    // 404 处理
    engine.NoRoute(func(c *gin.Context) {
        response.Fail(c, xerr.New(xerr.CodeRouteNotFound, "路由不存在"))
    })

    // 公共路由（无需认证）
    public := engine.Group("/api/v1")
    {
        public.GET("/health", handlers.Health.Check)
    }

    // 认证路由（需要 JWT）
    // auth := engine.Group("/api/v1")
    // auth.Use(middleware.JWTAuth(jwtMgr))
    // {
    //     // Step 06+ 填充
    // }
}
```

```go
// internal/router/handlers.go
package router

import "admin/internal/handler/health"

// Handlers 聚合所有 handler，由 main.go 构造
type Handlers struct {
    Health *health.Handler
    // 后续步骤添加更多 handler
}
```

### 6. Health Handler

```go
// internal/handler/health/health.go
package health

import (
    "admin/pkg/response"
    "github.com/gin-gonic/gin"
)

type Handler struct{}

func NewHandler() *Handler { return &Handler{} }

func (h *Handler) Check(c *gin.Context) {
    response.OK(c, gin.H{"status": "ok"})
}
```

### 7. 通用 DTO

```go
// internal/dto/common.go
package dto

// PageRequest 分页请求
type PageRequest struct {
    Page     int `form:"page" binding:"omitempty,min=1"`
    PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

func (p *PageRequest) GetPage() int {
    if p.Page <= 0 {
        return 1
    }
    return p.Page
}

func (p *PageRequest) GetPageSize() int {
    if p.PageSize <= 0 {
        return 10
    }
    return p.PageSize
}

func (p *PageRequest) GetOffset() int {
    return (p.GetPage() - 1) * p.GetPageSize()
}

// IDRequest URI 中的单个 ID
type IDRequest struct {
    ID string `uri:"id" binding:"required"`
}

// IDsRequest 批量 ID
type IDsRequest struct {
    IDs []string `json:"ids" binding:"required,min=1"`
}

// StatusRequest 批量状态变更
type StatusRequest struct {
    IDs    []string `json:"ids" binding:"required,min=1"`
    Status int      `json:"status" binding:"required,oneof=1 2"`
}
```

### 8. main.go 最终更新

Step 03 的变化只在 `server.New()` 内部（engine 组装移进去）——`main` 的 `run()` 组合根骨架 Step 01 就定死了，本步一行不改：`signal.NotifyContext` 得到可传播 ctx，`srv.Run(ctx)` 阻塞到收到信号或组件出错，内部 errgroup 统一优雅关闭。

```go
func run() error {
    // ... 第一层：Step 02 的 config/log/db/redis 初始化 + defer 逆序 Close ...

    // 第二层：组装 Server（engine + 中间件 + 路由都在 New 内部，见上方第 4 节）
    srv, err := server.New(server.Options{Config: cfg, DB: db, RDB: rdb, Log: log})
    if err != nil {
        return fmt.Errorf("init server: %w", err)
    }

    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    if err := srv.Run(ctx); err != nil {   // errgroup 编排：两个 g.Go + g.Wait 收敛
        return fmt.Errorf("run server: %w", err)
    }
    log.Info("server exited")
    return nil
}
```

> `run()` / `Run(ctx)` 的完整代码、errgroup 的两个 `g.Go`（`ListenAndServe` + `ErrServerClosed` 过滤、`<-ctx.Done()` 后带超时 `Shutdown`）、以及「为什么单组件也用 errgroup」的取舍，全部在 [bootstrap doc 01](./research/bootstrap/01-应用启动与组件生命周期-main组合根与gin-cron编排.md) 第一、二节——那里是生命周期编排的唯一权威版本，本步不复制。

## 依赖引入

```bash
go get github.com/gin-gonic/gin
go get github.com/google/uuid  # RequestID 生成
```

## 验收标准

```bash
# 1. 编译通过
go build ./...

# 2. 健康检查
curl -s http://localhost:8080/api/v1/health
# 期望：{"code":0,"message":"success","data":{"status":"ok"}}

# 3. 请求带 Request-ID
curl -s -D - http://localhost:8080/api/v1/health | grep X-Request-ID
# 期望：X-Request-ID: <uuid>

# 4. 404 处理
curl -s http://localhost:8080/api/v1/notexist
# 期望：{"code":40401,"message":"路由不存在"}

# 5. CORS 头
curl -s -D - -X OPTIONS http://localhost:8080/api/v1/health | grep Access-Control
# 期望：包含 Access-Control-Allow-Origin

# 6. 请求日志
# 启动后发请求，观察控制台日志包含：method, path, status, latency, request_id

# 7. Panic 恢复（临时测试）
# 添加一个 panic 路由测试，确认返回 500 且不暴露堆栈
```

---

*上一步：[Step 02 - 基础设施](step-02-config-logger-db.md) | 下一步：[Step 04 - 数据模型](step-04-database-schema.md)*
