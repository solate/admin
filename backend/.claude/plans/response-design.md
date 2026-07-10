# Response 封装与 Trace/Request ID 设计方案

## 背景

参考 backend-rbac 成熟实现，为当前 SaaS backend 设计统一响应封装。核心问题：
1. Response envelope 结构（永远 200 + body code）
2. **Trace ID vs Request ID 策略**（是否暴露、如何生成、与 OTel 集成）
3. 错误码体系（xerr.AppError + 分段 codes）

## 一、Response Envelope 设计（已定）

采用**国内后台惯例**：永远 HTTP 200，业务码在 body。

```go
type Response struct {
    Code      int    `json:"code"`
    Message   string `json:"message"`
    Data      any    `json:"data,omitempty"`
    RequestID string `json:"request_id"` // 见第二节讨论
}
```

### API 设计
```go
response.Success(c, data)                    // {200, "success", data, rid}
response.Error(c, err)                        // 自动识别 *xerr.AppError
response.ErrorWithMessage(c, code, message)   // 显式指定
```

### 与 xerr 集成
- `response.Error` 做 type assertion：`err.(*xerr.AppError)` → 用其 Code/Message
- fallback：非 AppError → `ErrInternal.Code`(1000) + `err.Error()` 作为 message

## 二、Trace ID vs Request ID 策略（需决策）

### 核心问题

| 方案 | request_id 来源 | 优点 | 缺点 |
|------|----------------|------|------|
| **A. 纯 Request ID**（参考实现） | 取 `X-Request-ID` header，无则 `uuid.New()` | 简单，即刻可用 | 未来接 OTel 后两套 ID（trace_id + request_id），日志关联复杂 |
| **B. Trace ID 优先** | 优先取 `traceparent` 的 trace-id，无则 uuid | 一套 ID 串全链路，OTel 无缝 | 需解析 W3C traceparent，trace-id 是 32 字符 hex（比 uuid 长） |
| **C. 不暴露 ID** | 不进 response body，只在日志/header 里 | body 最干净，trace 属观测性非业务 | 前端无法主动报 ID 排障（需教育用户"打开 devtools 看 header"） |

### 参考实现的做法（方案 A）

```go
// middleware
requestID := c.GetHeader("X-Request-ID")
if requestID == "" {
    requestID = uuid.New().String()
}
c.Set("X-Request-ID", requestID)
c.Writer.Header().Set("X-Request-ID", requestID)

// response.go
RequestID: getRequestID(c) // 从 gin context 取
```

每个 response body 都带 `request_id` 字段。日志 middleware 每行日志都打 `request_id`。

### 推荐方案：**B. Trace ID 优先（为 OTel 预留）**

```go
// middleware/request_id.go
func RequestIDMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        var requestID string
        
        // 1. 优先取 W3C traceparent（OTel 标准）
        if traceparent := c.GetHeader("traceparent"); traceparent != "" {
            // 格式：00-{trace-id}-{span-id}-{flags}
            parts := strings.Split(traceparent, "-")
            if len(parts) == 4 {
                requestID = parts[1] // trace-id (32 hex)
            }
        }
        
        // 2. 兜底取 X-Request-ID（兼容旧客户端/网关）
        if requestID == "" {
            requestID = c.GetHeader("X-Request-ID")
        }
        
        // 3. 最终生成 uuid
        if requestID == "" {
            requestID = uuid.New().String()
        }
        
        c.Set("request_id", requestID)
        c.Writer.Header().Set("X-Request-ID", requestID)
        c.Next()
    }
}
```

#### 决策点 1：是否暴露到 response body？

| 选项 | 说明 |
|------|------|
| **暴露**（参考实现） | `Response` 结构体有 `RequestID` 字段，每个响应都带。前端能直接拿到，用户截图/复制即可报障。 |
| **不暴露** | 只通过 response header `X-Request-ID` 返回，body 不带。需要用户/前端主动看 header 或前端拦截器提取。 |

**我的建议**：
- **初期暴露**（response body 带 `request_id`）—— 降低接入成本，前端/用户直接能看到
- **长期可选**：配置开关 `expose_request_id_in_body`，生产环境关掉（只留 header），减少 body 字节

暴露的理由：国内 toB 后台，运营/客户反馈 bug 时很少会 F12 看 header，直接截图 response 最快。

#### 决策点 2：trace-id 长度问题

W3C trace-id 是 **32 字符 hex**（如 `4bf92f3577b34da6a3ce929d0e0e4736`），比 UUID 的 36 字符短，但没有 `-` 分隔，可读性略差。

**处理**：
- 字段名保持 `request_id`（业务语义），但值可能是 trace-id
- Swagger example 用 `4bf92f3577b34da6a3ce929d0e0e4736`（trace 格式）或 `550e8400-e29b-41d4-a716-446655440000`（uuid 格式）都行，前端按字符串处理即可

## 三、xerr 错误码体系（照搬参考实现）

### 结构

```go
// pkg/xerr/errors.go
type AppError struct {
    Code    int
    Message string
    Err     error `json:"-"` // 内部 cause，不暴露给前端
}

func New(code int, message string) *AppError
func Wrap(code int, message string, err error) *AppError
```

### 预定义错误（pkg/xerr/codes.go）

分段规则（100 一段）：

| 范围 | 用途 | 示例 |
|------|------|------|
| 200 | 成功 | `ErrSuccess` |
| 1000-1999 | 通用错误 | `ErrInternal`(1000), `ErrInvalidParams`(1001), `ErrUnauthorized`(1003), `ErrForbidden`(1004) |
| 2000-2099 | 数据库错误 | `ErrRecordNotFound`(2000), `ErrDbRecordExist`(2001) |
| 2100-2199 | 认证错误 | `ErrInvalidCredentials`(2100), `ErrTokenExpired`(2101) |
| 2200-2299 | 租户错误 | `ErrTenantNotFound`(2201), `ErrTenantDisabled`(2204) |
| 2300+ | 业务模块错误 | 角色/菜单/部门... 每个模块 100 个号段 |

当前 backend 初期先定义 **1000-2299**（通用 + DB + 认证 + 租户），业务模块错误按需扩展。

## 四、Middleware 日志集成（与 xslog 适配）

参考实现用 zerolog，我们用 slog。关键点：
1. RequestIDMiddleware 设置 `request_id` 到 gin context
2. LoggerMiddleware 消费 `c.Errors`，每行日志带 `request_id`
3. 每个请求一条访问日志（method/path/status/duration）+ N 条错误日志（如果有）

```go
// middleware/logger.go (slog 版)
func LoggerMiddleware(log *slog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        
        c.Next()
        
        duration := time.Since(start)
        
        // 访问日志
        log.Info("request",
            slog.String("request_id", getRequestID(c)),
            slog.String("method", c.Request.Method),
            slog.String("path", path),
            slog.Int("status", c.Writer.Status()),
            slog.Duration("duration", duration),
            slog.String("client_ip", c.ClientIP()),
        )
        
        // 错误日志（如果有）
        if len(c.Errors) > 0 {
            for _, e := range c.Errors {
                log.Error("request error",
                    slog.String("request_id", getRequestID(c)),
                    slog.Any("error", e.Err),
                )
            }
        }
    }
}
```

## 五、实施清单

### 阶段 1：基础设施（本次）

- [ ] `go get github.com/gin-gonic/gin`，把 google/uuid 提为 direct 依赖
- [ ] 实现 `pkg/xerr`（errors.go + codes.go，定义 1000-2299）
- [ ] 实现 `pkg/response`（Response 结构 + Success/Error 函数）
- [ ] 实现 `internal/middleware/request_id.go`（trace-aware 生成）
- [ ] 实现 `internal/middleware/logger.go`（slog 版，消费 c.Errors）
- [ ] 实现 `internal/middleware/recovery.go`（panic → xerr.ErrInternal）
- [ ] 实现 `internal/router/router.go`（Setup 函数，注册 middleware）
- [ ] 更新 `internal/server/server.go`（从 net/http 切换到 gin）

### 阶段 2：健康检查 handler（验证）

- [ ] `internal/handler/health/health.go`（GET /health → `response.Success`）
- [ ] 注册路由，测试 response envelope + request_id

### 阶段 3：Swagger 集成（可选，后续）

- [ ] 安装 swag (`go install github.com/swaggo/swag/cmd/swag@latest`)
- [ ] Response 结构加 swagger 注解
- [ ] Handler 加 `@Success 200 {object} response.Response{data=...}`

## 六、待决策问题（请确认）

### Q1: request_id 是否暴露到 response body？
- **选项 A**（推荐）：暴露（`Response` 结构体有 `RequestID string` 字段）
- **选项 B**：不暴露（只通过 `X-Request-ID` response header 返回）

### Q2: trace-id 生成策略
- **选项 A**（推荐）：优先取 `traceparent` trace-id，兜底取 `X-Request-ID`，最终 uuid
- **选项 B**：只取 `X-Request-ID`，不管 traceparent（与参考实现一致，但未来接 OTel 时要重构）

### Q3: 错误码初始范围
- **选项 A**（推荐）：1000-2299（通用 + DB + 认证 + 租户），够初期用，业务模块按需扩展
- **选项 B**：直接照搬参考实现全部错误码（2000-2699，包含角色/菜单/部门/岗位），即使暂时用不上

---

**我的推荐配置**：
- Q1: **选项 A**（暴露 request_id 到 body）
- Q2: **选项 A**（trace-aware 生成，为 OTel 预留）
- Q3: **选项 A**（1000-2299 初始集，按需扩展）

请确认后我立即实施。
