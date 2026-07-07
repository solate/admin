# OpenTelemetry 链路追踪集成路径

> 本文是 pkg/xslog 的演进指南：从"单体日志"到"微服务全链路追踪"。
> 前提：你已经用 xslog 完成基础日志封装（ContextExtractor 机制已就位）。

---

## 0. OTel 是什么，解决什么问题

**OpenTelemetry（OTel）** 是一套**可观测性的开放标准 + SDK**。它统一管理三种信号（业界叫"可观测性三大支柱"）：

| 信号 | 是什么 | 例子 |
|---|---|---|
| **Traces（链路）** | 一个请求从进入到返回，经过哪些函数/服务/DB，每步耗时多少 | 就是本文的 trace_id/span_id |
| **Metrics（指标）** | 可聚合的数字，画成监控大盘 | QPS、延迟 P99、错误率、内存 |
| **Logs（日志）** | 结构化日志本身 | 你用 xslog 打的那些 |

**核心价值是"标准"**：你的代码只对接 OTel 一套 API，后端想用 Jaeger、Tempo、Datadog、阿里云 ARMS 都行，换后端不改一行业务代码。**这个"前端统一、后端可换"的思路，和 slog 的 `Handler` 解耦是同一个哲学**——slog 是日志界的这套，OTel 是整个可观测性的这套。

**对系统的好处**：出问题时能快速定位"慢在哪、错在哪"。举例——用户报"下单很慢"：

- **没有 trace**：翻日志，看到 service 层打了几条，但不知道是 DB 慢、Redis 慢、还是外部调用慢。靠猜 + 加日志复现。
- **有 trace**：打开这个请求的调用链，一眼看到 `SELECT orders` 花了 1.8s，其他都是毫秒级。直接定位。

跨服务时价值放大——微服务里一个请求穿 5 个服务，没有 trace 根本没法排查。但**这也说明它的适用边界**：单体单进程时，收益远不如它的接入成本，见下节。

---

## 1. 前置决策：什么时候接 OTel

| 阶段 | 判断 | 做什么 | 不做什么 |
|---|---|---|---|
| **单体 + 单实例 + 未上线/用户少** | 当前项目 | xslog 留 ContextExtractor 口子 | ❌ OTel 是纯负担（要跑 collector、每个依赖加 instrument）— **YAGNI** |
| **单体 + 生产初期，有真实用户** | "某接口偶尔慢"类模糊报障开始出现 | Prometheus metrics（`/metrics` 端点 + QPS/延迟/错误率大盘）| trace 暂缓（性价比：metrics >> trace） |
| **拆了微服务 / 复杂异步调用** | 请求链路跨多个服务/进程 | 上 Trace，TraceExtractor + otelgin + 跨服务传播 | — |

**核心洞察**：单体里请求链路在一个进程，`ERROR` 日志 + 堆栈 + request_id 已经够定位；trace 的价值在**跨服务/进程**时才爆发——那时没 trace 是真的抓瞎。

---

## 1. Level 1：日志-trace 关联（单体 → 微服务初期）

这一步的目标：**日志里带上 trace_id/span_id，能从一条日志跳到 Jaeger 里的完整调用链。** 日志还是 JSON 走 stdout（Loki 采集），只是多了两个字段。

### 1.1 TraceExtractor 实现

```go
// pkg/xtrace/extractor.go（不放 xslog，解耦 OTel 依赖）
package xtrace

import (
    "context"
    "log/slog"
    "go.opentelemetry.io/otel/trace"
)

// TraceExtractor 从 context 提取 trace_id/span_id，供 xslog 的 ContextExtractor 机制注入。
// 如果 ctx 里没有有效 span（启动期、后台任务），返回 nil 不注入任何字段。
func TraceExtractor(ctx context.Context) []slog.Attr {
    sc := trace.SpanContextFromContext(ctx)
    if !sc.IsValid() {
        return nil // 无 span，不注入
    }
    attrs := []slog.Attr{
        slog.String("trace_id", sc.TraceID().String()),
        slog.String("span_id", sc.SpanID().String()),
    }
    // 可选：采样标记（方便日志里直接看出这条 trace 有没有被采样）
    if sc.IsSampled() {
        attrs = append(attrs, slog.Bool("trace_sampled", true))
    }
    return attrs
}
```

**为什么不放 xslog 里？** 保持 xslog 零 OTel 依赖，可跨项目复制。`pkg/xtrace` 是可选的 OTel 集成层。

### 1.2 注册（main.go）

```go
import (
    "admin/pkg/xslog"
    "admin/pkg/xcontext"
    "admin/pkg/xtrace"
)

log := xslog.New(xslog.Config{
    Level: cfg.Log.Level,
    ContextExtractors: []xslog.ContextExtractor{
        xcontext.LogExtractor, // request_id / tenant_id
        xtrace.TraceExtractor, // trace_id / span_id ← 加这一行
    },
})
```

完事。业务代码依然 `log.InfoContext(ctx, "...")`，日志里自动多出 `trace_id` 和 `span_id`。

### 1.3 request_id vs trace_id：为什么两个 ID 都要留

接了 trace 之后常有人问：有了 trace_id/span_id，request_id 是不是就多余了？**不是——两者互补，都该留。** 上面的注册里两个 extractor 并存正是这个原因：一条日志同时带 `request_id` 和 `trace_id`。

| 维度 | request_id | trace_id / span_id |
|---|---|---|
| **依赖** | 中间件生成的 UUID，零依赖，永远在 | 只在有活跃 span 时存在（要 OTel SDK 起了 span） |
| **无 span 场景** | 照样有（启动期、cron、后台任务、消息消费） | 没有（`TraceExtractor` 里 `!sc.IsValid()` 返回 nil） |
| **面向客户端** | 回写 `X-Request-ID` 响应头，用户报障直接甩给你 | 通常内部使用，一般不回给客户端 |
| **采样** | 日志全量走 Loki，每条都带 | 生产 1% 采样（§4），99% 请求在 Jaeger 查不到 |
| **范围** | 一个 HTTP 请求（单服务视角） | 整条链路跨所有服务（trace）/ 链路里的一段（span） |
| **落地** | 日志（Loki） | Jaeger / Tempo |

**排障动线（两个 ID 接力）**：客户端给你 **request_id** → grep 日志找到那条 → 日志里带着 **trace_id** → 跳 Jaeger 看完整链路和耗时。前者是"客户端能给的入口 + 全量兜底"，后者是"看链路在哪慢 / 哪错"。

**对当前项目**：还没接 OTel，压根没有 trace_id，request_id 是唯一的关联 ID——这也是它零基础设施依赖的价值：接不接 OTel 都在。

### 1.4 otelgin 中间件（span 起点）

```go
import "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

r := gin.New()
r.Use(otelgin.Middleware("admin-api")) // ← service name，span 从这里开始
r.Use(RequestIDMiddleware())
r.Use(AccessLogMiddleware(log))
```

`otelgin.Middleware` 在每个请求入口 start 一个 span，塞进 ctx。之后 `SpanContextFromContext(ctx)` 才能捞到。

**中间件顺序**：otelgin 必须在最前面（它要包住整个请求链路）；RequestID 在它后面（request_id 不影响 trace）。

### 1.5 业务代码手动加 span（关键路径埋点）

otelgin 只记录 HTTP 请求级的 span（一个请求一个）。如果你想看"请求里的哪个步骤慢"，需要手动加子 span：

```go
import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("admin-service") // package 级变量，复用

func (s *UserService) CreateUser(ctx context.Context, req *dto.CreateUserReq) error {
    ctx, span := tracer.Start(ctx, "UserService.CreateUser") // 起子 span
    defer span.End() // defer 保证 span 一定关闭
    
    log.InfoContext(ctx, "开始创建用户", "name", req.Name) // trace_id 自动带
    
    if err := s.repo.Create(ctx, user); err != nil {
        span.RecordError(err) // 把 error 记到 span（Jaeger 里会标红）
        log.ErrorContext(ctx, "创建用户失败", xslog.Err(err))
        return err
    }
    
    span.SetAttributes(attribute.String("user_id", user.UserID)) // 可选：记关键业务字段
    return nil
}
```

**什么地方该加 span？** 关键路径（写 DB、调外部 API、复杂计算），不是每个函数都加（太细粒度会淹没）。

### 1.6 Exporter 配置（Jaeger 本地验证）

Trace 要发到一个后端才能看。本地开发用 Jaeger all-in-one 最简单：

```bash
# 启动 Jaeger
docker run -d --name jaeger \
  -p 16686:16686 \
  -p 14268:14268 \
  jaegertracing/all-in-one:latest

# 浏览器打开 http://localhost:16686 就能看 UI
```

main.go 配置 exporter：

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func initTracer() func() {
    exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(
        jaeger.WithEndpoint("http://localhost:14268/api/traces"),
    ))
    if err != nil {
        log.Fatal("failed to create jaeger exporter", xslog.Err(err))
    }
    
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(resource.NewWithAttributes(
            semconv.ServiceNameKey.String("admin-api"), // service name
        )),
    )
    otel.SetTracerProvider(tp)
    
    return func() { _ = tp.Shutdown(context.Background()) }
}

func main() {
    shutdown := initTracer()
    defer shutdown()
    // ... 启动 HTTP server
}
```

打几个请求，去 Jaeger UI 搜 `admin-api`，就能看到完整调用链。

---

## 2. 微服务场景：跨服务 trace 传播

单体里 ctx 在一个进程透传就行；微服务要跨网络，trace context 必须通过 **HTTP header** 或 **消息队列 header** 传递。

### 2.1 服务 A 调用服务 B（HTTP）

#### 服务 A（调用方）

```go
import "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

client := &http.Client{
    Transport: otelhttp.NewTransport(http.DefaultTransport), // 包一层
}

req, _ := http.NewRequestWithContext(ctx, "POST", "http://service-b/api/users", body)
resp, _ := client.Do(req) // otelhttp 自动在 header 里注入 traceparent
```

`otelhttp.NewTransport` 会在请求发出前，把当前 span 的 trace context 塞进 `traceparent` header（W3C 标准格式）。

#### 服务 B（被调方）

服务 B 的 Gin 用了 `otelgin.Middleware`，它会自动从 `traceparent` header 提取 trace context，继续这条链路（不是新起一条）。

**不需要手动传递**——只要两边都用 `otelgin` + `otelhttp`，trace 自动接上。

### 2.2 消息队列（Kafka/RabbitMQ）

HTTP 有 header 天然传 context，消息队列需要手动注入 message header。

#### 生产者（发消息）

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/propagation"
    "github.com/segmentio/kafka-go"
)

// 把 trace context 注入到一个 map
carrier := propagation.MapCarrier{}
otel.GetTextMapPropagator().Inject(ctx, carrier)

msg := kafka.Message{
    Key:   []byte(userID),
    Value: body,
    Headers: []kafka.Header{
        {Key: "traceparent", Value: []byte(carrier.Get("traceparent"))}, // 注入
    },
}
writer.WriteMessages(ctx, msg)
```

#### 消费者（收消息）

```go
// 从 message headers 提取 trace context
headers := make(map[string]string)
for _, h := range msg.Headers {
    headers[h.Key] = string(h.Value)
}
ctx := otel.GetTextMapPropagator().Extract(context.Background(), propagation.MapCarrier(headers))

// 后续用这个 ctx 处理，trace 就接上了
processMessage(ctx, msg.Value)
```

---

## 3. Context 传播的坑（trace 断链头号原因）

Trace 最常见的 bug 不是"没加 span"，而是 **ctx 没传下去，span 拿不到**。

### 3.1 异步 goroutine

```go
// ❌ 错误：新起 goroutine 用 context.Background()
func (s *Service) HandleRequest(ctx context.Context, req *Request) {
    go func() {
        ctx := context.Background() // trace 断了！
        s.ProcessAsync(ctx, req)
    }()
}

// ✅ 正确方式 1：直接传入父 ctx（如果父请求返回不会 cancel）
go func(ctx context.Context) {
    s.ProcessAsync(ctx, req)
}(ctx)

// ✅ 正确方式 2：detach span context（父请求返回会 cancel，但 trace 要保留）
go func() {
    // 从父 ctx 拷贝 span context，但用新的 Background 做底
    ctx := trace.ContextWithSpanContext(context.Background(), trace.SpanContextFromContext(ctx))
    s.ProcessAsync(ctx, req)
}()
```

**本项目已有 `xcontext.CopyContext`** 拷贝 TenantID/UserID 等字段，trace 场景同理——需要显式拷贝 span context。可以扩展它：

```go
// xcontext/copy.go
func CopyContextWithTrace(parent context.Context) context.Context {
    ctx := context.Background()
    // 拷贝已有的业务字段
    ctx = context.WithValue(ctx, tenantIDKey, parent.Value(tenantIDKey))
    ctx = context.WithValue(ctx, userIDKey, parent.Value(userIDKey))
    // ... 其他字段
    
    // 拷贝 span context（trace 不断）
    sc := trace.SpanContextFromContext(parent)
    if sc.IsValid() {
        ctx = trace.ContextWithSpanContext(ctx, sc)
    }
    return ctx
}
```

### 3.2 DB 查询、Redis 调用必须用 Context 变体

GORM / Redis 都有带 ctx 和不带 ctx 两套 API：

```go
// ❌ 错误：不带 ctx，otelgorm 不知道当前 span
db.Create(user)

// ✅ 正确：带 ctx，otelgorm 自动关联 span
db.WithContext(ctx).Create(user)
```

```go
// ❌ 错误
rdb.Get("key")

// ✅ 正确
rdb.Get(ctx, "key")
```

如果用不带 ctx 的版本，链路里就看不到 DB/Redis 的耗时，等于盲区。

---

## 4. 采样策略（控制成本）

全量采集 trace 成本爆炸（存储 + 性能），生产必须采样。但 1% 随机采样又怕漏掉错误请求。

### 4.1 Head-based sampling（SDK 端决定）

采样决策在 span 开始时做，简单但有缺陷：

```go
tp := sdktrace.NewTracerProvider(
    sdktrace.WithSampler(sdktrace.TraceIDRatioBased(0.01)), // 1% 采样
    sdktrace.WithBatcher(exporter),
    ...
)
```

**缺陷**：span 开始时不知道后面会不会出错，所以错误请求也可能被丢（采样决策已做完）。

### 4.2 Tail-based sampling（collector 端决定，推荐）

Collector 收到**完整 trace**（所有 span 都到齐）后，再根据规则决定留哪条：

- 所有错误 trace 保留
- 慢请求（> 500ms）保留
- 正常请求 1% 采样

配置在 OTel Collector 的 `config.yaml`：

```yaml
receivers:
  otlp:
    protocols:
      http:

processors:
  tail_sampling:
    decision_wait: 10s  # 等 10s 收集完整 trace
    policies:
      - name: errors
        type: status_code
        status_code: {status_codes: [ERROR]}
      - name: slow
        type: latency
        latency: {threshold_ms: 500}
      - name: randomized
        type: probabilistic
        probabilistic: {sampling_percentage: 1}

exporters:
  jaeger:
    endpoint: jaeger:14250

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [tail_sampling]
      exporters: [jaeger]
```

**生产推荐 tail-based**，但需要独立部署 OTel Collector（应用直接发 OTLP 到 collector，不直连 Jaeger）。

---

## 5. Level 2：日志作为 OTel 信号（OTLP 直发）

Level 1 是"日志带 trace_id"（日志还是 JSON 走 stdout → Loki）；Level 2 是"日志本身变成 OTel log 信号"，通过 OTLP 直发 collector，和 trace/metric 走同一管道。

### 5.1 多路分发（stdout + OTLP 双写）

通常你还想保留 stdout（Loki 继续采集），所以需要**多路分发**——一条日志同时写两个后端。

标准库没有 `MultiHandler`（提案 golang/go#65954 还在讨论），需要自己写 10 行或用 [`slogmulti`](https://github.com/samber/slog-multi) 库：

```go
// 内部实现（或用 slogmulti 库）
type multiHandler struct {
    handlers []slog.Handler
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
    for _, h := range m.handlers {
        if err := h.Handle(ctx, r); err != nil {
            return err // 任一失败就失败
        }
    }
    return nil
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
    for _, h := range m.handlers {
        if h.Enabled(ctx, level) {
            return true // 任一启用就启用
        }
    }
    return false
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
    newHandlers := make([]slog.Handler, len(m.handlers))
    for i, h := range m.handlers {
        newHandlers[i] = h.WithAttrs(attrs)
    }
    return &multiHandler{handlers: newHandlers}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
    newHandlers := make([]slog.Handler, len(m.handlers))
    for i, h := range m.handlers {
        newHandlers[i] = h.WithGroup(name)
    }
    return &multiHandler{handlers: newHandlers}
}
```

组装：

```go
import "go.opentelemetry.io/contrib/bridges/otelslog"

opts := &slog.HandlerOptions{Level: parseLevel(cfg.Level), AddSource: cfg.AddSource}

baseHandlers := []slog.Handler{
    slog.NewJSONHandler(os.Stdout, opts),  // Loki 采集（保留）
    otelslog.NewHandler("admin-api"),      // OTLP collector
}

inner := &multiHandler{handlers: baseHandlers}
handler := xslog.newContextHandler(inner, extractors) // contextHandler 包在最外层
log := slog.New(handler)
```

**分层**：`contextHandler` 包 `multiHandler`，`multiHandler` 包两个 base handler。注入逻辑（request_id/trace_id）只在 contextHandler 一处，下游两个 handler 都拿到完整字段。

### 5.2 OTLP Exporter 配置

```go
import (
    "go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
    "go.opentelemetry.io/otel/sdk/log"
)

exporter, err := otlploghttp.New(context.Background(),
    otlploghttp.WithEndpoint("otel-collector:4318"),
    otlploghttp.WithInsecure(), // 生产用 TLS
)
if err != nil {
    log.Fatal("failed to create otlp log exporter", xslog.Err(err))
}

loggerProvider := log.NewLoggerProvider(
    log.WithProcessor(log.NewBatchProcessor(exporter)),
)
// 配置给 otelslog.NewHandler 使用
```

---

## 6. 完整 instrument 清单

微服务场景下，**每个依赖都要单独接 instrument**，漏一个就是链路盲区：

| 依赖 | 库 | 作用 |
|---|---|---|
| **Gin** | `otelgin.Middleware` | HTTP 入口 span（请求级） |
| **GORM** | [`otelgorm`](https://github.com/go-gorm/opentelemetry) 插件 | DB 查询 span |
| **Redis** | `otelredis` hook（或 [`go-redis/extra/redisotel`](https://github.com/redis/go-redis)） | Redis 命令 span |
| **HTTP client** | `otelhttp.NewTransport` | 出站 HTTP span + 自动传播 traceparent |
| **Kafka** | 手动注入 header（见 §2.2） | 消息队列传播 |
| **gRPC** | `otelgrpc.UnaryClientInterceptor` | gRPC 调用 span + 传播 |

每个库的接入方式略有不同，但模式一致：
1. 包一层（middleware / interceptor / hook）
2. 自动 start span
3. HTTP/gRPC 自动传播 trace context，消息队列手动注入

---

## 7. 演进时间线总结

| 阶段 | 信号 | 后端 | xslog 改动 | 业务改动 |
|---|---|---|---|---|
| **单体开发期** | 日志（stdout JSON） | Loki | 无（已就位） | 无 |
| **生产初期** | + Metrics | Prometheus | `/metrics` 端点（不在 xslog） | 埋点（可选） |
| **开始排查慢查询** | + Trace（Level 1） | Jaeger | 注册 `TraceExtractor` | otelgin 中间件 + 关键路径加 span |
| **微服务拆分** | Trace 跨服务 | Tempo / 阿里云 ARMS | 无（extractor 已就位） | otelhttp 包 client + 消息队列注入 header |
| **统一可观测性平台** | Log as signal（Level 2） | OTLP Collector | `multiHandler` 双写 | 无（透明） |

**核心洞察**：xslog 的 `ContextExtractor` 机制让 trace 集成在**业务层接入**（中间件 + instrument），**日志层零改动**（只注册一个 extractor）。这正是依赖倒置的收益。

---

## 8. 参考

- [OpenTelemetry Go SDK 官方文档](https://opentelemetry.io/docs/instrumentation/go/)
- [otelgin middleware](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin)
- [otelslog bridge](https://pkg.go.dev/go.opentelemetry.io/contrib/bridges/otelslog)
- [otelgorm 插件](https://github.com/go-gorm/opentelemetry)
- [Jaeger Quick Start](https://www.jaegertracing.io/docs/latest/getting-started/)
- [OTel Collector Tail Sampling Processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/tailsamplingprocessor)
- [W3C Trace Context 规范](https://www.w3.org/TR/trace-context/)（`traceparent` header 格式）
- [slogmulti 库（多路分发）](https://github.com/samber/slog-multi)
