# 从零设计 Go 结构化日志：slog 封装 + Gin 集成实战（2026）

> 本文是一篇完整学习文档。读完你能独立完成：从 0 设计一套日志方案 → 封装成可复用包 → 接入项目 → 配置参数 → 桥接 GORM → 接入 Gin（含需要哪些额外配置）→ 让请求级字段（request_id / tenant_id）自动注入每一条日志。
>
> 选型理由见 [日志库选型调研](../saas-backend/research/logging/01-日志库选型调研.md)——结论是 2026 年新项目用标准库 [`log/slog`](https://pkg.go.dev/log/slog)。本文是那篇调研的"教学续作"：调研告诉你**为什么**，本文教你**怎么做**。

---

## 0. 开篇：为什么是 slog

Go 1.21（2023-08）把结构化日志收进了标准库——`log/slog`。到 2026 年，生态已经收敛：新项目的应用代码默认写 `*slog.Logger`。

slog 不是最快的（zerolog 在跑分上快约 4×，但 25ns 和 101ns 对 Web 后端都毫无意义），它的价值在三件事：

1. **标准库**——零外部依赖，永远跟着 Go 发版。
2. **`slog.Handler` 解耦**——你的调用代码和编码引擎分开，未来想换 OTel / 更快后端，只改初始化那一行，调用点一行不动。
3. **生态对齐**——`otelslog`、`sloglint`、charm/log、各观测平台都以 slog 为前端标准。

本文目标：把 slog 从零封装成一个**生产可用的日志层**，并干净地接入 Gin。所有代码基于真实的 SaaS 后端结构（`pkg/logger` + `internal/config` + `cmd/server/main.go` + Gin 中间件）。

---

## 1. 先搞懂 slog 三件套

写封装前必须理解三个核心概念，否则只会复制粘贴。

### 1.1 Logger / Record / Handler

```
你的代码                slog 内部                     输出
─────────              ──────────                   ─────
logger.Info(...)  ──►  构建 Record（级别/时间/msg/字段）
                        │
                        ▼
                    Handler.Enabled(ctx, level)  ──► 级别过滤
                        │ 通过
                        ▼
                    Handler.Handle(ctx, record)  ──► 编码 + 写出
```

- **`Logger`**：你日常打交道的对象，提供 `Info/Warn/Error/Debug` 等方法。它内部持有一个 Handler。
- **`Record`**：一条日志的不可变快照——时间、级别、消息、调用位置、若干 `Attr`（字段）。
- **`Handler`**：决定"这条日志要不要记、记成什么格式、写到哪里"。`JSONHandler`、`TextHandler` 是标准库提供的两个实现。

**关键认知**：`Logger` 只是门面，真正干活的是 `Handler`。换 Handler = 换整个日志行为，调用代码不变。这就是"后端可换"的本质。

### 1.2 三种调用方式

```go
// ① key-value（最简洁，但弱类型，易写错：奇数个参数会静默产出脏日志）
slog.Info("user login", "user_id", "u1", "role", "admin")

// ② 强类型 Attr（推荐，编译器帮你兜底）
slog.Info("user login", slog.String("user_id", "u1"), slog.String("role", "admin"))

// ③ context 变体（带 ctx，用于 trace 关联和请求级字段注入——后面重点讲）
slog.InfoContext(ctx, "user login", slog.String("user_id", "u1"))
```

> ⚠️ 方式 ① 是双刃剑：方便，但 `slog.Info("x", "k")`（漏了 value）能编译通过、运行时静默错误。生产代码用 **②**，并用 [`sloglint`](https://github.com/go-simpler/sloglint) 静态检查强制。

### 1.3 级别

slog 内置四个级别（值是 `int`，可自定义中间级别）：

```go
slog.LevelDebug = -4
slog.LevelInfo  = 0
slog.LevelWarn  = +4
slog.LevelError = +8
```

**注意：slog 没有 `Fatal` 级别**（也没有 `os.Exit`）。这是设计哲学——日志和进程控制分离。启动失败的 `Fatal` 语义，用标准库 `log.Fatal` 或自己写个 `Fatal(log, msg)` 辅助函数（见 §4.4）。

---

## 2. 第一步：设计配置参数（需要配置什么）

设计从"该暴露哪些旋钮"开始。原则：**只暴露你真的会在不同环境（开发/测试/生产）调的参数，其余写死。**

### 2.1 我的 Config 长这样

```go
// pkg/logger/config.go
package logger

// Config 是日志层的全部配置。字段值来自 internal/config.LogConfig。
type Config struct {
	Level     string // debug / info / warn / error
	Format    string // json（默认，生产）/ text（开发，logfmt 风格）
	AddSource bool   // 是否记录调用位置（file:line）——生产建议开
}
```

### 2.2 每个字段为什么这样选

| 字段 | 为什么留 | 为什么不留别的 |
|---|---|---|
| `Level` | 开发要 debug、生产要 info——必须可调 | — |
| `Format` | 生产收 JSON（机器解析）、终端看 text（人读） | **不暴露"输出路径"**：写 stdout，由容器/系统（systemd、Docker、journald）接管轮转与收集，比应用内 lumberjack 轮转更标准。需要文件轮转时再外挂 |
| `AddSource` | 出问题时 `file:line` 是定位神器；但有几 ns 开销和文件路径泄露，做成开关 | **不暴露 `TimeFormat`**：统一 RFC3339 纳秒，不给人乱改的机会 |

> 经验：配置项越少越好。每多一个旋钮，就多一种"配错了排查半天"的可能。YAGNI——等你真的需要文件轮转、采样、Hook 时再加。

### 2.3 配置文件长这样

```yaml
# config/config.yaml
log:
  level: debug       # 开发 debug / 生产 info
  format: text       # 开发 text / 生产 json
  add_source: true
```

对应项目配置层（viper 反序列化）：

```go
// internal/config/config.go
type LogConfig struct {
	Level     string `mapstructure:"level"`
	Format    string `mapstructure:"format"`
	AddSource bool   `mapstructure:"add_source"`
}
```

校验用白名单（不用反射、不引第三方校验库，对齐项目"零反射"哲学）：

```go
func validate(c *Config) error {
	switch c.Log.Level {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("log.level invalid: %q", c.Log.Level)
	}
	switch c.Log.Format {
	case "", "json", "text":
	default:
		return fmt.Errorf("log.format invalid: %q", c.Log.Format)
	}
	return nil
}
```

---

## 3. 第二步：封装 `pkg/logger`（怎么封装最好）

这是全文的核心。封装的目标：**给项目一个统一的日志入口，把 slog 的初始化复杂度藏起来，调用方只拿到一个配好的 `*slog.Logger`。**

### 3.1 完整实现

```go
// pkg/logger/logger.go
package logger

import (
	"io"
	"os"
	"path/filepath"

	"log/slog"
)

// New 根据配置构建一个生产可用的 *slog.Logger。
// 设计要点：
//   - 级别用 *slog.LevelVar，支持运行时并发安全地调整（热更新级别）
//   - AddSource 记录调用位置；用 ReplaceAttr 把长路径裁成 base name
//   - Format 选 JSONHandler（生产）或 TextHandler（开发）
func New(cfg Config) *slog.Logger {
	// 1. 级别：LevelVar 实现 slog.Leveler，可运行时 Set
	level := new(slog.LevelVar)
	switch cfg.Level {
	case "debug":
		level.Set(slog.LevelDebug)
	case "warn":
		level.Set(slog.LevelWarn)
	case "error":
		level.Set(slog.LevelError)
	default:
		level.Set(slog.LevelInfo)
	}

	// 2. 输出：统一 stdout，交给容器/系统接管轮转
	var out io.Writer = os.Stdout

	// 3. Handler 选项
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
		// 把 source 的完整路径裁成文件名，避免泄露部署路径
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				if src, ok := a.Value.Any().(*slog.Source); ok {
					src.File = filepath.Base(src.File)
				}
			}
			return a
		},
	}

	// 4. 选 Handler
	var handler slog.Handler
	switch cfg.Format {
	case "text":
		handler = slog.NewTextHandler(out, opts)
	default: // "json" 及未识别值都走 JSON（生产默认）
		handler = slog.NewJSONHandler(out, opts)
	}

	return slog.New(handler)
}
```

### 3.2 关键设计决策逐条讲

**(a) 为什么用 `*slog.LevelVar` 而不是 `slog.LevelInfo` 常量？**

`LevelVar` 是并发安全的（底层 `atomic.Int64`）。把它交给 `HandlerOptions.Level` 后，你手里还捏着这个指针，随时 `level.Set(slog.LevelDebug)` 就能**热更新级别**——比如收到 SIGHUP 调到 debug 排障、完事调回 info。如果传常量，改级别就得重建整个 logger。

```go
type App struct {
	level *slog.LevelVar // 持有，供热更新
	log   *slog.Logger
}
// 排障时：app.level.Set(slog.LevelDebug)  —— 全局立即生效
```

**(b) 为什么返回 `*slog.Logger` 走依赖注入，不用 `slog.SetDefault`？**

`slog.SetDefault(logger)` 设全局默认后，包级 `slog.Info(...)` 处处可用，方便。但它和项目的"依赖显式传递、不用全局状态"哲学冲突（与 config 层一致）。**推荐 DI**：main 创建 logger，按需传给 db/server/handler。好处是可测试（测试时传一个写 `bytes.Buffer` 的 logger 断言输出）。

`slog.SetDefault` 不是禁用——它能让那些**你控制不了初始化的第三方库**（用包级 `slog.Info` 的）也走你的 Handler。所以策略是：**DI 为主，`SetDefault` 兜底**：

```go
// main.go 里：DI 传给项目各层，同时设个默认兜第三方库
log := logger.New(cfg)
slog.SetDefault(log) // 可选：让第三方库的包级 slog 调用也走你的 Handler
```

**(c) 关于"console 美化输出"**

slog 没有原生彩色 console writer（`TextHandler` 是 `key=value` 纯文本，无颜色）。开发期想要彩色对齐 zerolog ConsoleWriter 的体验，两个选择：

1. 用 [`charmbracelet/log`](https://github.com/charmbracelet/log) 做 slog 后端（它实现了 `slog.Handler`），终端输出很漂亮。代价：多一个依赖。
2. 自己写个几十行的 color Handler。

教学项目保持零依赖，用 `TextHandler`（开发）+ `JSONHandler`（生产）就够。要彩色时换 charm/log 后端，**调用代码一行不改**——这就是 Handler 解耦的红利。

### 3.3 补 Fatal（slog 没有）

```go
// pkg/logger/logger.go
import "os"

// Fatal 记一条 Error 后立即退出。用于不可恢复的启动错误。
// 放在 logger 包里，避免各处重复 os.Exit。
func Fatal(log *slog.Logger, msg string, args ...any) {
	log.Error(msg, args...)
	os.Exit(1)
}
```

### 3.4 单元测试

测试日志最稳的办法：把 Handler 的 `io.Writer` 换成 `bytes.Buffer`，断言输出内容。

```go
// pkg/logger/logger_test.go
func TestNew_JSONLevelFilter(t *testing.T) {
	var buf bytes.Buffer
	// 直接构造一个写 buffer 的 logger（测试用，绕过 New 的 os.Stdout）
	log := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelWarn, // 只记 Warn 及以上
	}))

	log.Info("should be filtered") // 低于 Warn，被丢弃
	log.Warn("kept", slog.String("k", "v"))

	out := buf.String()
	if strings.Contains(out, "should be filtered") {
		t.Fatal("info leaked through warn level")
	}
	if !strings.Contains(out, "kept") || !strings.Contains(out, `"k":"v"`) {
		t.Fatalf("warn record missing/malformed: %s", out)
	}
}
```

> 测试 `New()` 本身时，用 `os.Pipe` 捕获 stdout（参考 `backend-rbac/pkg/utils/logger/logger_test.go` 的写法）。

---

## 4. 第三步：接入项目（如何加入项目）

封装好后，要在 main.go 里按正确顺序组装。核心顺序：**config → logger → db → redis → server**。logger 必须早于 db，因为 GORM 桥接需要它（见 §5）。

```go
// cmd/server/main.go
package main

import (
	"log"
	"os/signal"
	"syscall"

	"admin/internal/config"
	"admin/internal/server"
	"admin/pkg/database"
	"admin/pkg/logger"
	"admin/pkg/rdb"
)

func main() {
	// 1. 配置
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatal("load config: " + err.Error()) // 此时 logger 还没建，用 stdlib log
	}

	// 2. 日志（最先，后面所有基础设施都依赖它）
	log2 := logger.New(logger.Config{
		Level:     cfg.Log.Level,
		Format:    cfg.Log.Format,
		AddSource: cfg.Log.AddSource,
	})

	// 3. 数据库（把 logger 传进去，桥接 SQL 日志）
	db, err := database.New(database.Config{ /* ... */ }, log2)
	if err != nil {
		logger.Fatal(log2, "connect database failed", slog.Any("err", err))
	}
	defer database.Close(db)

	// 4. Redis
	rdbClient, err := rdb.New(rdb.Config{ /* ... */ })
	if err != nil {
		logger.Fatal(log2, "connect redis failed", slog.Any("err", err))
	}
	defer rdbClient.Close()

	log2.Info("all infrastructure initialized", slog.Int("port", cfg.Server.Port))

	// 5. HTTP 服务器（logger 通过 Options 传入，供 Gin 中间件用）
	srv, err := server.New(server.Options{Config: cfg, DB: db, RDB: rdbClient, Log: log2})
	if err != nil {
		logger.Fatal(log2, "init server failed", slog.Any("err", err))
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		if err := srv.Start(); err != nil {
			logger.Fatal(log2, "start server failed", slog.Any("err", err))
		}
	}()
	<-ctx.Done()
	srv.Stop(shutdownCtx)
}
```

**组装顺序的约束**：logger 在 db 之前——因为 `database.New` 要把 logger 传给 GORM 桥接。config 在最前——logger 的 level/format 来自 config。配置加载失败用 stdlib `log.Fatal`（此时 logger 没建）。

---

## 5. 第四步：桥接 GORM（让 SQL 日志进入你的日志流）

GORM 用自己的 `gormlogger.Interface` 输出 SQL 日志。我们要实现这个接口，把 SQL 执行轨迹导到 slog。结构照搬项目现有的 `gormZerologger`，只把 zerolog 调用换成 slog 的 **context 变体**（这样 SQL 日志也能带 request_id——见 §7）。

```go
// pkg/database/database.go（节选：GORM 适配器）
package database

import (
	"context"
	"fmt"
	"time"

	"log/slog"
	gormlogger "gorm.io/gorm/logger"
)

type gormSlogger struct {
	log      *slog.Logger
	logLevel gormlogger.LogLevel
}

func newGormLogger(log *slog.Logger) gormlogger.Interface {
	return &gormSlogger{log: log, logLevel: gormlogger.Info}
}

func (l *gormSlogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return &gormSlogger{log: l.log, logLevel: level} // 返回新实例，符合 GORM 约定
}

func (l *gormSlogger) Info(ctx context.Context, msg string, args ...any) {
	l.log.InfoContext(ctx, fmt.Sprintf(msg, args...))
}

func (l *gormSlogger) Warn(ctx context.Context, msg string, args ...any) {
	l.log.WarnContext(ctx, fmt.Sprintf(msg, args...))
}

func (l *gormSlogger) Error(ctx context.Context, msg string, args ...any) {
	l.log.ErrorContext(ctx, fmt.Sprintf(msg, args...))
}

// Trace 是核心：每条 SQL 执行后回调。正常走 Debug，出错升级 Error。
func (l *gormSlogger) Trace(ctx context.Context, begin time.Time,
	fc func() (sql string, rowsAffected int64), err error) {

	if l.logLevel <= gormlogger.Silent {
		return
	}
	elapsed := time.Since(begin)
	sql, rows := fc()

	attrs := []slog.Attr{
		slog.String("sql", sql),
		slog.Int64("rows", rows),
		slog.Duration("elapsed", elapsed),
	}

	if err != nil {
		l.log.ErrorContext(ctx, "sql error", append(attrs, slog.Any("err", err))...)
		return
	}
	l.log.DebugContext(ctx, "sql", attrs...) // 慢查询可在这里按 elapsed 阈值升级到 Warn
}
```

接入点（`database.New` 里）：

```go
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
	Logger: newGormLogger(log),
})
```

**为什么用 `*Context` 变体？** 因为 §7 会把 request_id 放进 `context.Context`。GORM 调用 `Trace(ctx, ...)` 时传进来的 `ctx` 就是请求的 ctx——用 `DebugContext(ctx, ...)` 后，request_id 会自动出现在 SQL 日志里。这是 slog context 模型相对 zerolog 的一个实打实优势。

---

## 6. 第五步：Gin 集成（重点：是否需要额外配置）

**直接回答你的问题：需要，但只有四项，且都有明确理由。** 下面逐项讲。

### 6.1 必须的额外配置（四项）

```go
// internal/server/server.go 或 internal/router/router.go
import "github.com/gin-gonic/gin"

func newEngine(cfg *config.Config, log *slog.Logger) *gin.Engine {
	// ① 用 gin.New()，不要用 gin.Default()
	//    gin.Default() 会装 gin 自带的 Logger + Recovery 中间件，往 os.Stdout 写
	//    非结构化日志——和你的 slog 中间件重复且格式冲突。从裸 engine 开始，装自己的。
	gin.SetMode(cfg.Server.Mode) // ② 把 gin 模式（debug/release/test）交给配置
	r := gin.New()

	// ③ 处理 gin.DefaultWriter：gin 的启动路由表、debug 输出走这个 writer
	//    两种做法任选：
	//    (a) 重定向到 slog（推荐，统一日志流）
	gin.DefaultWriter = newSlogWriter(log)
	//    (b) 或干脆丢弃（你的中间件已接管访问日志）
	// gin.DefaultWriter = io.Discard

	// ④ 中间件链（顺序很重要，见 6.2）
	r.Use(RequestID(log))
	r.Use(AccessLog(log))
	r.Use(Recovery(log))
	r.Use(CORS()) // 内部把 Access-Control-Expose-Headers 设为含 X-Request-ID

	return r
}
```

**(a) `gin.New()` 而非 `gin.Default()`** —— 避免重复日志。`gin.Default()` 装的 Logger 写的是 `| 200 |  1.234ms |  /api/x` 这种人读格式，和你的结构化访问日志冲突。自己装。

**(b) `gin.SetMode(cfg.Server.Mode)`** —— release 模式下 gin 不打印启动路由表、关掉一些 debug 输出。模式应由配置驱动。

**(c) `gin.DefaultWriter` 重定向** —— gin 自己的输出（路由表、警告）默认写 stdout。重定向到 slog 让**所有**日志走同一个出口。适配器很简单：

```go
// pkg/logger/gin_writer.go
type slogWriter struct{ log *slog.Logger }

func newSlogWriter(log *slog.Logger) *slogWriter { return &slogWriter{log: log} }

func (w *slogWriter) Write(p []byte) (int, error) {
	msg := strings.TrimSpace(string(p))
	if msg != "" {
		w.log.Info(msg, slog.String("source", "gin")) // 标记来自 gin 自身
	}
	return len(p), nil
}
```

**(d) CORS 暴露 `X-Request-ID`** —— 你会把 request_id 回写响应头（见 §7.1），但默认 CORS 不暴露自定义头，前端拿不到。要加上：

```go
// internal/middleware/cors.go
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID") // ← 关键
		// ...其余 CORS 头
	}
}
```

### 6.2 中间件链顺序（为什么这样排）

```go
r.Use(RequestID(log))   // 1. 最先：生成/透传 request_id，塞进 ctx
r.Use(AccessLog(log))   // 2. 依赖 1 的 request_id；记录 start，c.Next()，算 duration
r.Use(Recovery(log))    // 3. 依赖 1 的 request_id；panic 时记 stack
r.Use(CORS())           // 4. OPTIONS 预检放行
// 鉴权组：
auth.Use(AuthMiddleware())
auth.Use(RBACMiddleware())
```

**为什么 RequestID 必须在最前**：后面所有中间件（AccessLog、Recovery）的日志都要带 request_id，得先有人把它放进 context。否则后面 `FromContext` 取不到。

**为什么 AccessLog 在 Recovery 前**：AccessLog 的 `start := time.Now()` 在 Recovery 之前记录，`c.Next()` 之后算 duration——这样即使 handler panic 被 Recovery 兜住，duration 也正确包含了 panic+recover 的时间。

---

## 7. 中间件三件套（slog 版，原创代码）

下面三个中间件是 slog 版的完整实现。字段集对齐项目成熟范式（`backend-rbac`），但用 slog context 模型重写。

### 7.1 RequestID 中间件

```go
// internal/middleware/request_id.go
package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

type ctxKey struct{} // request_id 的 context key

// RequestID 透传上游 X-Request-ID（网关已生成则复用），否则生成；塞进 ctx + 响应头。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = randomID(8) // 生产可换成 google/uuid
		}

		// 塞进 gin.Context（给同请求其他中间件用）
		c.Set("request_id", rid)
		// 塞进标准 context（给 slog 的 context 模型用——见 §8）
		ctx := context.WithValue(c.Request.Context(), ctxKey{}, rid)
		c.Request = c.Request.WithContext(ctx)
		// 回写响应头，客户端可凭此查日志
		c.Header("X-Request-ID", rid)

		c.Next()
	}
}

func randomID(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
```

### 7.2 AccessLog 中间件

```go
// internal/middleware/access_log.go
package middleware

import (
	"time"

	"log/slog"

	"github.com/gin-gonic/gin"
)

// AccessLog 记录每条 HTTP 请求的结构化访问日志。
func AccessLog(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next() // 执行后续中间件 + handler

		log.InfoContext(c.Request.Context(), "HTTP Request",
			slog.String("request_id", c.GetString("request_id")),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("query", c.Request.URL.RawQuery),
			slog.String("ip", c.ClientIP()),
			slog.Int("status", c.Writer.Status()),
			slog.Int("size", c.Writer.Size()),
			slog.Duration("duration", time.Since(start)),
			slog.String("user_agent", c.Request.UserAgent()),
		)
	}
}
```

> 想记请求 body（带密码脱敏）？参考 `backend-rbac/pkg/utils/bodyreader/bodyreader.go` 的 `SanitizeParams`：在 `c.Next()` 前读 body、脱敏、再把可重放的 `io.NopCloser` 塞回 `c.Request.Body`。逻辑与日志库无关，本文不展开。

### 7.3 Recovery 中间件

```go
// internal/middleware/recovery.go
package middleware

import (
	"runtime/debug"

	"log/slog"

	"github.com/gin-gonic/gin"
)

// Recovery 兜住 panic，记结构化日志（含 stack），返回统一 500。
func Recovery(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.ErrorContext(c.Request.Context(), "panic recovered",
					slog.String("request_id", c.GetString("request_id")),
					slog.String("path", c.Request.URL.Path),
					slog.Any("error", r),
					slog.String("stack", string(debug.Stack())),
				)
				c.AbortWithStatusJSON(500, gin.H{"code": 500, "message": "internal error"})
			}
		}()
		c.Next()
	}
}
```

到这里，你已经有一套完整的 Gin + slog 访问日志、错误恢复、请求追踪了。

---

## 8. 进阶：让请求级字段自动注入（slog 的正确姿势）

**这是 slog 相对 zerolog 最值得学的一招。**

注意上面三个中间件，每条日志都手写 `slog.String("request_id", c.GetString("request_id"))`。这有两个问题：

1. **繁琐**：service 层、GORM 层的日志想带 request_id，得每次手动取手动加（`backend-rbac` 就是这么干的——到处 `log.Error().Str("tenant_id", ...)`）。
2. **易漏**：开发者忘了加，那条日志就没 request_id，排障时断链。

slog 的 context 模型能彻底解决：**写一个自定义 Handler，在每条日志写入前，从 `context.Context` 自动读取 request_id（以及 tenant_id、user_id）注入进去。一处实现，全局所有日志自动带，调用方完全不用关心。**

### 8.1 自定义 context 注入 Handler

```go
// pkg/logger/context_handler.go
package logger

import (
	"context"

	"log/slog"
)

// ContextHandler 包裹任意 Handler，在 Handle 时从 ctx 自动注入请求级字段。
type ContextHandler struct {
	slog.Handler
}

func NewContextHandler(h slog.Handler) *ContextHandler {
	return &ContextHandler{Handler: h}
}

// Handle 是核心：每条日志都经过这里。ctx 携带了 request_id/tenant_id。
func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	// 从 ctx 取请求级字段，追加到这条 record
	if rid, ok := ctx.Value(requestIDKey).(string); ok && rid != "" {
		r.AddAttrs(slog.String("request_id", rid))
	}
	if tid, ok := ctx.Value(tenantIDKey).(string); ok && tid != "" {
		r.AddAttrs(slog.String("tenant_id", tid))
	}
	return h.Handler.Handle(ctx, r)
}

// ⚠️ 必须重写 WithAttrs / WithGroup，否则 logger.With(...) 生成子 logger 时
//    会丢失这层包裹（返回的是内层 Handler，不再自动注入）。
func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextHandler{Handler: h.Handler.WithAttrs(attrs)}
}
func (h *ContextHandler) WithGroup(name string) slog.Handler {
	return &ContextHandler{Handler: h.Handler.WithGroup(name)}
}
```

> 这里的 `requestIDKey` / `tenantIDKey` 要和中间件往 ctx 塞值时用的 key 一致（建议放一个共享的 `pkg/xcontext` 包统一管理，参考 `backend-rbac/pkg/xcontext`）。

### 8.2 装进 logger

```go
// pkg/logger/logger.go 的 New() 末尾改成：
func New(cfg Config) *slog.Logger {
	// ...（前面选 handler 的逻辑不变）
	return slog.New(NewContextHandler(handler))
}
```

### 8.3 收益：调用代码变干净

装上后，**只要用 context 变体**，request_id 自动出现，不用手写：

```go
// service 层
func (s *UserService) Get(ctx context.Context, id string) (*User, error) {
	user, err := s.repo.Get(ctx, id)
	if err != nil {
		// 不用手写 request_id！只要 ctx 是请求传下来的，自动带
		s.log.ErrorContext(ctx, "get user failed",
			slog.String("user_id", id),
			slog.Any("err", err),
		)
		return nil, err
	}
	return user, nil
}
```

GORM 的 SQL 日志（§5 用的 `DebugContext(ctx, ...)`）也自动带上 request_id——一条慢查询能直接关联到哪个请求。

中间件里那几行 `slog.String("request_id", ...)` 现在**可以删掉**了（保留也无害，会重复，所以删掉更干净）。

> **对比 zerolog**：zerolog 要达到同样效果得用 `log.Ctx(ctx)` + 把 logger 存进 context，且 zerolog 的 context 桥接性能差（调研里那个 46× 慢点）。slog 的 Handler 拦截模型是原生为这个场景设计的。

---

## 9. 一句话前瞻：接 OpenTelemetry

等你要做分布式追踪，把 §3.1 里的 Handler 换成 OTel 的即可：

```go
import "go.opentelemetry.io/contrib/bridges/otelslog"

log := slog.New(otelslog.NewLogger()) // 日志成为 OTel 信号，自动带 trace_id/span_id
```

**前提**：全程用 `*Context` 变体（`InfoContext` 等），且 ctx 里有活跃 span。这也是本文从 §5 起就强调"用 context 变体"的原因——你现在养成习惯，将来接 OTel 零改动。

---

## 10. 总结 + 常见坑

### 设计要点回顾

1. **封装藏复杂度**：`pkg/logger.New(cfg)` 返回配好的 `*slog.Logger`，调用方不碰 Handler 细节。
2. **配置最小化**：只暴露 `Level/Format/AddSource`，输出走 stdout 交给容器。
3. **级别用 LevelVar**：留运行时热更新的口子。
4. **DI 为主 + SetDefault 兜底**：项目内显式传 logger，第三方库走默认。
5. **context 变体 + 自定义 Handler**：请求级字段自动注入，调用方零负担。

### 五个常见坑

| 坑 | 表现 | 解法 |
|---|---|---|
| key-value 奇数个 | `slog.Info("x", "k")` 静默脏日志 | 用强类型 `slog.Attr`；CI 加 `sloglint` |
| 忘 `.Msg()` | （这是 zerolog 的坑，slog 没有） | slog 用 `slog.Info(msg, ...)`，不存在此问题——slog 的一个优点 |
| 无 Fatal 级别 | 启动失败想 exit 但 slog 没有 | stdlib `log.Fatal` 或 `logger.Fatal()` 辅助（§3.3） |
| `gin.Default()` 重复日志 | 访问日志打两遍、格式不一 | 用 `gin.New()` + 自己的中间件 |
| 忘用 `*Context` 变体 | request_id / trace_id 丢失，自动注入失效 | 全程用 `InfoContext` 等；`sloglint` 可强制 context-only |
| 自定义 Handler 没重写 `WithAttrs` | `logger.With(...)` 后子 logger 丢注入 | 必须重写 `WithAttrs/WithGroup` 重新包裹（§8.1） |

---

## 11. 参考链接

- [日志库选型调研（本文选型依据）](../saas-backend/research/logging/01-日志库选型调研.md)
- [Structured Logging with slog — The Go Blog](https://go.dev/blog/slog)
- [`log/slog` — pkg.go.dev](https://pkg.go.dev/log/slog)
- [Choosing a Go Logging Library in 2026 — Dash0](https://www.dash0.com/guides/golang-logging-libraries)
- [Logging in Go with slog: A Practitioner's Guide — Dash0](https://www.dash0.com/guides/logging-in-go-with-slog)
- [go-simpler/sloglint — slog 调用静态检查](https://github.com/go-simpler/sloglint)
- [otelslog — OTel bridge](https://pkg.go.dev/go.opentelemetry.io/contrib/bridges/otelslog)
- [charmbracelet/log — 彩色终端 slog 后端](https://github.com/charmbracelet/log)
- [Gin 官方文档](https://gin-gonic.com/docs/)
- [zerolog context 模型 vs slog（社区对比）](https://www.reddit.com/r/golang/comments/1665chi/which_logger_do_you_use_benchmarking_the_most/)

---

*上一篇：[日志库选型调研](../saas-backend/research/logging/01-日志库选型调研.md) | 系列教程：[saas-backend](../saas-backend/README.md)*
