# 从零设计 Go 结构化日志：slog 自定义 Handler 与 context 注入实战（2026）

> 本文是一篇完整的实现复盘。读完你能独立完成：从 0 设计一套基于标准库 `log/slog` 的日志包 → 用自定义 `slog.Handler` 实现请求级字段（request_id / tenant_id）自动注入 → 处理级别解析、source 裁剪、敏感字段脱敏这些工程细节 → 干净地接入 Gin。
>
> 更重要的是，本文记录了**封装过程中真实踩过的坑**和**每个设计决策背后的取舍**——不是"应该怎么写"的教条，而是"为什么最后写成这样"的推演。
>
> 选型理由见 [日志库选型调研](../saas-backend/research/logging/01-日志库选型调研.md)——结论是 2026 年新项目用标准库 [`log/slog`](https://pkg.go.dev/log/slog)。对应源码：[`backend/pkg/xslog/`](../../backend/pkg/xslog)。

---

## 0. 开篇：为什么是 slog（而不是 zerolog / zap）

Go 1.21（2023-08）把结构化日志收进了标准库——`log/slog`。到 2026 年，生态已经收敛：新项目的应用代码默认写 `*slog.Logger`。

slog 不是最快的（zerolog 在跑分上快约 4×，但 25ns 和 101ns 对 Web 后端都毫无意义），它的价值在三件事：

1. **标准库**——零外部依赖，永远跟着 Go 发版。
2. **`slog.Handler` 解耦**——你的调用代码和编码引擎分开，未来想换 OTel / 更快后端，只改初始化那一行，调用点一行不动。
3. **生态对齐**——`otelslog`、`sloglint`、charm/log、各观测平台都以 slog 为前端标准。

那更快的 zerolog、更能组合的 zap 呢？对一个 Web 后端，它们都不是更好的选择：

- **性能不是决策点**：zerolog / phuslu 原生 ~25ns、slog ~101ns，看着差 4×，但 Web 后端的瓶颈是 DB / Redis / 网络（毫秒级），日志编码的纳秒差距对总延迟毫无意义。
- **zap**：`zapcore` 可组合性最强（采样、多路由），但要多扛一个依赖、API 更冗长，对管理后台属过度设计。
- **zerolog**：流式 `.Str().Msg()` 很顺手，但**一旦用了它的原生 API 就锁死后端**——想换观测后端得重写调用点。更关键的是 **OTel 时代它没有官方 OTLP bridge**（slog 桥接还慢 46×），这在可观测性统一化的趋势下是硬伤。

一句话：极致性能对 Web 后端是伪需求，而 slog 的"标准库 + 后端可换 + OTel 官方支持"才是长期正确的下注。完整的 benchmark 与成熟项目横评见 [日志库选型调研](../saas-backend/research/logging/01-日志库选型调研.md)。

选定 slog，下一个问题是：为什么不直接 `slog.New(slog.NewJSONHandler(os.Stdout, nil))` 裸用，非要封一层 `pkg/xslog`？因为裸 slog 缺三样生产必需的能力：

1. **请求级字段自动注入**（核心）：request_id / tenant_id 要每条日志手写，繁琐且易漏——漏一条，排障时链路就断了。这是 xslog 真正不可替代的价值（§7）。
2. **工程细节**：source 裁成 `dir/file:line`、敏感字段脱敏、级别字符串解析——每个都是生产必选，裸用得每个项目重写一遍。
3. **可复制**：封装一次（约 150 行），跨项目直接 copy。zerolog 的"简单"是把这些复杂度藏进了库里，代价是换不了后端；xslog 把复杂度收在自己手里（约 150 行），换来完全的控制权和可移植性。

本文目标：把 slog 封装成一个**生产可用、可跨项目复制**的日志包 `pkg/xslog`，核心能力是**请求级字段自动注入**，并交代清楚每个决策的来龙去脉。

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

**关键认知**：`Logger` 只是门面，真正干活的是 `Handler`。换 Handler = 换整个日志行为，调用代码不变。这就是"后端可换"的本质，也是本文所有扩展的落点——**我们扩展 Handler，从不包装 Logger**（`*slog.Logger` 是 struct 不是 interface，包了就丢方法）。

### 1.2 三种调用方式

```go
// ① key-value（最简洁，但弱类型，易写错：奇数个参数会静默产出脏日志）
slog.Info("user login", "user_id", "u1", "role", "admin")

// ② 强类型 Attr（推荐，编译器帮你兜底）
slog.Info("user login", slog.String("user_id", "u1"), slog.String("role", "admin"))

// ③ context 变体（带 ctx，用于 trace 关联和请求级字段注入——本文的重点）
slog.InfoContext(ctx, "user login", slog.String("user_id", "u1"))
```

> ⚠️ 方式 ① 是双刃剑：方便，但 `slog.Info("x", "k")`（漏了 value）能编译通过、运行时静默错误。生产代码用 **②**，并用 [`sloglint`](https://github.com/go-simpler/sloglint) 静态检查强制。
>
> **本文的封装会让你养成全程用 ③（context 变体）的习惯**——因为只有 `ctx` 传进来，自动注入的 request_id 才能生效（见 §7）。

### 1.3 级别

slog 内置四个级别（值是 `int`，可自定义中间级别）：

```go
slog.LevelDebug = -4
slog.LevelInfo  = 0   // ← 零值就是 Info，这个细节后面 parseLevel 会用到
slog.LevelWarn  = +4
slog.LevelError = +8
```

**注意：slog 没有 `Fatal` 级别**（也没有 `os.Exit`）。这是刻意的设计哲学——日志和进程控制分离。我们曾经给它补了个 `Fatal`，后来又删掉了，§6 会完整讲这个来回。

---

## 2. 设计配置：只暴露该调的旋钮

设计从"该暴露哪些参数"开始。原则：**只暴露你真的会在不同环境（开发/测试/生产）调的参数，其余写死或用零值兜底。**

```go
// pkg/xslog/config.go

// Format 指定日志输出格式。
type Format = string

const (
	FormatJSON Format = "json" // 生产环境默认，结构化便于解析
	FormatText Format = "text" // 本地开发，可读性好
)

// Config 是构造 logger 的配置。字段值来自 internal/config.LogConfig。
// 零值可用（Level=info、Format 空串按 JSON 处理、Output=os.Stdout）。
type Config struct {
	Level  string    // 级别 debug/info/warn/error，空或未知值 → info
	Format Format    // 输出格式 json(默认) / text
	Output io.Writer // 输出目标，默认 os.Stdout（测试可注入 bytes.Buffer）
	// AddSource 为 true 时在日志中记录调用位置（裁成 dir/file:line）。
	AddSource bool
	// ContextExtractors 是自定义的 context 字段提取器（见 §7）。
	ContextExtractors []ContextExtractor
	// ReplaceAttr 是用户自定义的字段改写函数（如脱敏，见 §5）。
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
}
```

几个决策交代一下：

**为什么 `Output` 进 Config、却不暴露"文件路径 / 轮转"？** 输出目标做成 `io.Writer` 有实际价值——**单测注入 `bytes.Buffer` 就能断言输出内容**，不必捕获 stdout。但"写哪个文件、怎么轮转"不暴露：云原生部署里日志轮转是容器运行时 / systemd-journald 的职责，应用只管往 stdout 吐 JSON。真要落文件，外部重定向即可。

**为什么 `Format` 用类型别名 `= string` 而不是定义类型？** 定义类型（`type Format string`）能让编译器挡住拼错的值，但代价是从配置读来的 string 要显式转换。考虑到 Format 本来就是从配置字符串来的、且入口只有 `New` 一处，这里主动放弃了那层类型安全换取调用便利——**这是一个有意识的取舍，不是疏忽**。

**配置文件长这样**（viper 反序列化）：

```yaml
# config/config.yaml
log:
  level: debug       # 开发 debug / 生产 info
  format: text       # 开发 text / 生产 json
  add_source: true
```

校验用白名单（不引第三方校验库，对齐项目"零反射"哲学）：

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

## 3. 构造函数 New：把复杂度收敛成一次调用

封装的目标：**调用方一次 `New(cfg)` 拿到配好的 `*slog.Logger`，不碰 Handler 细节。**

```go
// pkg/xslog/logger.go
func New(cfg Config) *slog.Logger {
	// 填充默认值
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}
	// 构造底层 handler（JSON 或 Text）
	opts := &slog.HandlerOptions{
		Level:       parseLevel(cfg.Level),
		AddSource:   cfg.AddSource,
		ReplaceAttr: chainReplaceAttr(cfg.ReplaceAttr),
	}
	var base slog.Handler
	switch cfg.Format {
	case FormatText:
		base = slog.NewTextHandler(cfg.Output, opts)
	default: // "json"/"" 都走 JSONHandler，兑现"零值默认 JSON"契约
		base = slog.NewJSONHandler(cfg.Output, opts)
	}
	// 用 contextHandler 包裹，注入 context 字段（§7 核心）
	handler := newContextHandler(base, cfg.ContextExtractors)
	return slog.New(handler)
}
```

三个组成部分：`parseLevel`（§4）、`chainReplaceAttr`（§5）、`newContextHandler`（§7）。下面逐个拆。

### 踩坑记录：默认格式到底是 JSON 还是 Text？

最初这里写的是 `if cfg.Format == FormatJSON { JSON } else { Text }`——意思是**零值空串会走 Text**。但 config.go 的注释白纸黑字写着"零值默认 JSON"。**代码和文档直接打架**：生产环境如果漏配 `format`，本该拿到机器可解析的 JSON，结果吐出一堆 Text，日志采集器直接懵。

修法是改成 `switch` + `default: JSON`：只有显式 `format: text` 才走 Text，其余（`json`、空串、甚至将来的未知值）一律 JSON。**默认值要向"最安全的生产行为"兜底**，而不是让零值落到一个开发态格式上。

---

## 4. parseLevel：一个"看似多余"的转换，和一个真实的坑

配置里的级别是字符串（`"debug"`），而 `slog.HandlerOptions.Level` 要的是 `slog.Leveler`。中间必须有人转。第一版是手写 switch：

```go
// 第一版——能跑，但有坑
func parseLevel(s string) slog.Level {
    switch s {
    case "debug": return slog.LevelDebug
    case "warn":  return slog.LevelWarn
    case "error": return slog.LevelError
    default:      return slog.LevelInfo
    }
}
```

### 踩坑记录：大小写与空格让 `ERROR` 静默降级成 info

这个 switch 精确匹配小写。配置里要是写成 `level: ERROR`、`level: Debug`、或者不小心带了空格 `level: " warn"`，**全部 fallthrough 到 `default`，静默变成 info**。你以为线上只收 error，结果 info 哗哗地打——排障时才发现级别根本没生效。这种"配置写了但没起作用、还不报错"的坑最咬人。

### 修法：直接复用 `slog.Level.UnmarshalText`

slog 的 `slog.Level` 自带 `UnmarshalText`，**天生大小写不敏感**（内部 `ToUpper`），还顺带支持 `"info+2"` 这种偏移语法。与其自己维护一张会漂移的映射表，不如直接用它：

```go
// parseLevel 把配置字符串映射到 slog.Level（实现 slog.Leveler）。
// 直接复用 slog.Level.UnmarshalText：大小写不敏感（DEBUG/Debug 均可），
// 语义与 slog 自身完全一致，无需自维护映射表。
// 先 TrimSpace 兜住配置里的首尾空白（UnmarshalText 本身不容忍空白）。
// 空值、纯空白或非法值一律回落到 Info（UnmarshalText 失败时不改动目标变量）。
func parseLevel(s string) slog.Level {
    var lv slog.Level
    if err := lv.UnmarshalText([]byte(strings.TrimSpace(s))); err != nil {
        return slog.LevelInfo
    }
    return lv
}
```

两个细节值得记：

1. **`TrimSpace` 是 `UnmarshalText` 唯一没覆盖的短板**——它对大小写宽容，但对首尾空白严格。补一个 `TrimSpace` 就把两者的长处都拿到了。
2. **失败回落 Info 是白送的**：`UnmarshalText` 解析失败时不改动目标变量，而 `slog.LevelInfo` 恰好是 `slog.Level` 的零值（`0`）。所以 `var lv slog.Level` 声明出来就是 Info，非法值直接返回它即可，语义天然吻合"默认 info"的契约。

### 岔路：为什么不用 `*slog.LevelVar`？

`*slog.LevelVar` 是 slog 提供的**可运行时并发安全改级别**的类型。第一版封装曾用它，注释还写着"留将来热改级别的口子"。但复盘时发现——**这个口子从来没接通**：

```go
opts := &slog.HandlerOptions{
    Level: parseLevel(cfg.Level), // ← 就算返回 LevelVar 指针，传进去后也没人持有
}
```

`LevelVar` 的唯一价值是"构造后还能 `.Set()` 动态改"，但要用这个能力，你必须**持有那个指针**（存进某个 struct + 暴露 `SetLevel` 方法）。这里造完就丢给 `HandlerOptions` 了，没人存、没有 setter——所谓"运行时改级别"从未存在。它提供的动态能力实际是**零**，纯粹比 `slog.Level` 多一次 `new()` 分配和一层指针。这是典型的 **YAGNI**：为一个想象中的需求预留了用不上的复杂度。

**将来真要动态调级别怎么办？** 不是退回 `LevelVar` 就完事，而要把口子**真正接通**：持有 `*slog.LevelVar`、暴露 `SetLevel(slog.Level)`、接一个 `/admin/loglevel` 端点。在有这个明确需求之前，不预留半成品。

---

## 5. source 裁剪与脱敏：两个 ReplaceAttr 如何并存

`slog.HandlerOptions.ReplaceAttr` 是个"每条日志的每个字段都会过一遍"的钩子。我们有两个需求都要用它：

- **source 裁剪**（内置）：`AddSource: true` 时 slog 默认输出 `{"source":{"function":...,"file":"/abs/build/path/...","line":42}}`——又冗长又泄露构建机绝对路径。
- **敏感字段脱敏**（用户可选）：把 `password`/`token` 的值替换成 `[REDACTED]`。

但 `ReplaceAttr` 只有一个。第一版把它写死成 source 裁剪，用户就没法再传脱敏；反过来开放给用户，又丢了 source 裁剪。**解法是串联**：

```go
// chainReplaceAttr 把内置的 shortenSource 与用户自定义的 ReplaceAttr 串联：
// source 裁剪永远生效（只动 SourceKey），用户函数随后叠加。二者作用的 key
// 不重叠，互不干扰。user 为 nil 时只跑 shortenSource。
func chainReplaceAttr(user func(groups []string, a slog.Attr) slog.Attr) func(groups []string, a slog.Attr) slog.Attr {
    return func(groups []string, a slog.Attr) slog.Attr {
        a = shortenSource(groups, a)
        if user != nil {
            a = user(groups, a)
        }
        return a
    }
}
```

### source 裁剪：裁成 `dir/file:line`，不是纯文件名

```go
func shortenSource(_ []string, a slog.Attr) slog.Attr {
    if a.Key != slog.SourceKey {
        return a
    }
    src, ok := a.Value.Any().(*slog.Source)
    if !ok || src == nil {
        return a
    }
    a.Value = slog.StringValue(fmt.Sprintf("%s:%d", shortFile(src.File), src.Line))
    return a
}

// shortFile 取路径末尾两段 "dir/file"。不足两段时原样返回。
func shortFile(file string) string {
    if i := strings.LastIndexByte(file, '/'); i >= 0 {
        if j := strings.LastIndexByte(file[:i], '/'); j >= 0 {
            return file[j+1:]
        }
    }
    return file
}
```

**为什么保留一级目录，而不是裁成纯文件名？** 这是踩过的教训。最初裁成 `logger.go:88`，看着够用，但真实项目里 `service.go`、`handler.go`、`query.go` 会在十几个域子包下重复出现（`user/service.go`、`role/service.go`…）。日志里只剩 `service.go:42` 根本不知道是哪个包，定位反而更慢。裁成 `user/service.go:42` 才有意义——这也是 zap 的 `ShortCallerEncoder` 的做法，保留一级目录消歧、又不泄露完整路径。

两个 slog 细节：

1. **为什么用 `strings` 而非 `filepath`**：`slog.Source.File` 由 runtime 填充，分隔符**恒为 `/`**（与运行平台无关，Windows 上也是 `/`）。`filepath.Base` 在不同 OS 行为不一致，`strings.LastIndexByte` 才是跨平台正确写法。
2. **为什么重写整个 `a.Value` 而非原地改 `src.File`**：`slog.Value` 为零分配被设计成**不透明类型**，取内容只能 `.Any()` 断言成 `*slog.Source`（见 [golang/go#59280](https://github.com/golang/go/issues/59280)）。成熟写法是把 `a.Value` 重写成 `slog.StringValue`。

### 脱敏：一行 group-kind 防呆

```go
func RedactReplaceAttr(sensitiveKeys ...string) func(groups []string, a slog.Attr) slog.Attr {
    lowered := make([]string, len(sensitiveKeys))
    for i, k := range sensitiveKeys {
        lowered[i] = strings.ToLower(k)
    }
    return func(_ []string, a slog.Attr) slog.Attr {
        if a.Value.Kind() == slog.KindGroup { // 跳过 group，避免整组结构被字符串替换
            return a
        }
        if slices.Contains(lowered, strings.ToLower(a.Key)) {
            return slog.String(a.Key, "[REDACTED]")
        }
        return a
    }
}
```

那行 `KindGroup` 判断是防呆：`ReplaceAttr` 会对**所有** attr 触发，包括 group 类型的属性。万一某个 group 名恰好命中敏感词，不加这行判断会把整个 group 用 `slog.String` 替换掉、破坏结构。现实里几乎不会发生，但一行成本、无副作用，值得加。

---

## 6. Fatal 的删除来回：一个"方便"如何违背设计哲学

这段是本文最想讲的一个决策——因为它经历了完整的"加进来 → 用着 → 想通了 → 删掉"。

### 为什么当初加了 Fatal

slog 没有 `Fatal`（zerolog 有 `.Fatal()`）。启动阶段"连库失败就退出"很常见，于是封装了一个：

```go
// 曾经的 xslog.Fatal——记一条 Error 再 os.Exit(1)
func Fatal(log *slog.Logger, msg string, args ...any) {
    var pcs [1]uintptr
    runtime.Callers(2, pcs[:]) // 让 source 指向调用方而非 Fatal 自己
    r := slog.NewRecord(time.Now(), slog.LevelError, msg, pcs[0])
    r.Add(args...)
    _ = log.Handler().Handle(context.Background(), r)
    os.Exit(1)
}
```

用起来很顺：`if err != nil { xslog.Fatal(log, "connect db failed", xslog.Err(err)) }`。

### 为什么又删了

深入对比 slog 官方设计后，发现这个"方便"违背了它的核心哲学。slog 作者刻意不提供 Fatal，理由有四条，条条都戳中：

1. **日志库的职责是"记录"，不是"控制流程"**。什么算"致命"是业务决策（启动失败要退，但请求失败只记 error），不该藏在日志调用里。`log.Fatal` 让代码看起来只是记日志，实际会中断流程——隐式副作用，读代码的人容易漏。

2. **`os.Exit` 跳过所有 `defer`**。这是最实际的坑。项目里 `main` 已经 `defer database.Close(db)`，但如果在 goroutine 里调 `xslog.Fatal`，`os.Exit` 会让**所有 defer 全部不执行**——数据库连接、临时文件、分布式锁统统不清理。

3. **测试不友好**。库自己调 `os.Exit`，单测根本没法跑（进程直接退）。要测得起子进程捕获退出码，复杂度爆炸。

4. **关注点分离**。"日志只记录，调用方决定退不退"才是成熟做法。这正是 Go 1.21 引入 slog 时刻意区别于老 `log` 包（有 `log.Fatal`）的改进——官方用"不提供"表明立场。

### 替代方案：`run() error` 模式

删掉 Fatal，`main` 改造成标准的 `run() error`——所有资源用 `defer` 清理，任何一步失败 `return error`，由 `main` 统一打日志 + `os.Exit`：

```go
func main() {
    // main 只负责退出码：run 返回 error 就打日志后非零退出。
    if err := run(); err != nil {
        slog.Error("server exited with error", slog.Any("err", err))
        os.Exit(1) // ← 此时 run 内的 defer 已全部执行
    }
}

func run() error {
    cfg, err := config.InitConfig()
    if err != nil {
        return fmt.Errorf("load config: %w", err)
    }
    log := xslog.New(xslog.Config{
        Level: cfg.Log.Level, Format: cfg.Log.Format, AddSource: cfg.Log.AddSource,
    }).With("service", cfg.App.Name, "env", cfg.App.Env) // 静态字段由调用方加

    db, err := database.New(/* ... */, log)
    if err != nil {
        return fmt.Errorf("connect database: %w", err)
    }
    defer database.Close(db) // ← run 返回时一定执行

    // 原本 goroutine 里调 Fatal 会跳过上面的 defer，改成把 error 送回主流程：
    srvErr := make(chan error, 1)
    go func() {
        if err := srv.Start(); err != nil {
            srvErr <- err
        }
    }()
    select {
    case err := <-srvErr:
        return fmt.Errorf("start server: %w", err)
    case <-ctx.Done():
        log.Info("shutdown signal received")
    }
    srv.Stop(shutdownCtx)
    return nil // ← defer 自动清理，main 看到 nil 正常退出
}
```

**收益**：`os.Exit` 只在 `main`、`run()` 已返回之后调用，defer 保证执行；错误处理显式（每处 `return fmt.Errorf`）；`run()` 可被单测直接调用断言返回的 error。多写几行 `return`，换来的是对齐 slog 官方哲学、日志包保持"只记录"的纯粹。

> 注意上面那行 `.With("service", ...)`：**静态业务字段（service/env）也不该进日志包的 Config**。它们只是普通 attr，由调用方在 main 里 `.With(...)` 挂上。日志包只管机制，不预设业务字段名——否则 Config 会无限膨胀成"全项目日志字段集合"。

---

## 7. 核心：让请求级字段自动注入（本文的重头戏）

**这是 slog 相对 zerolog 最值得学的一招，也是整个 xslog 封装真正的价值所在。**

前面 §2-6 那些（级别解析、source 裁剪、脱敏）其实裸用 slog 也能写，只是每个项目重复一遍。**真正让这个包不可替代的，是 context 字段自动注入**——中间件解析出 request_id/tenant_id 后，业务代码每条日志都要手写 `slog.String("request_id", xcontext.GetRequestID(ctx))`，繁琐且易漏。漏一条，排障时链路就断了。

slog 的 Handler 模型能彻底解决：**写一个自定义 Handler，在每条日志写入前从 `context.Context` 自动读取字段注入进去。一处实现，全局所有日志自动带，调用方零负担。**

### 7.1 自定义 contextHandler

自定义 Handler 是 slog **官方唯一推荐的扩展方式**。它是官方博客《A Guide to Writing slog Handlers》里说的 **wrapping handler**：内部持有另一个 Handler，做完自己的事再委托下去。

```go
// pkg/xslog/handler.go
type contextHandler struct {
    inner      slog.Handler
    extractors []ContextExtractor
}

func newContextHandler(inner slog.Handler, extractors []ContextExtractor) *contextHandler {
    all := make([]ContextExtractor, 0, len(extractors)+1)
    all = append(all, contextFieldsExtractor) // 内置的 WithField(s) 提取器排最前
    all = append(all, extractors...)
    return &contextHandler{inner: inner, extractors: all}
}

// Enabled 直接委托底层。
func (h *contextHandler) Enabled(ctx context.Context, level slog.Level) bool {
    return h.inner.Enabled(ctx, level)
}

// Handle 是核心：依次运行所有提取器，把 Attr 追加到 Record，再交给底层输出。
func (h *contextHandler) Handle(ctx context.Context, rec slog.Record) error {
    for _, extract := range h.extractors {
        if attrs := extract(ctx); len(attrs) > 0 {
            rec.AddAttrs(attrs...)
        }
    }
    return h.inner.Handle(ctx, rec)
}

// WithAttrs / WithGroup 透传给底层，并保持 contextHandler 包装（关键，见下）。
func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
    return &contextHandler{inner: h.inner.WithAttrs(attrs), extractors: h.extractors}
}
func (h *contextHandler) WithGroup(name string) slog.Handler {
    return &contextHandler{inner: h.inner.WithGroup(name), extractors: h.extractors}
}
```

四个方法的写法都不是随意的：

| 方法 | 为什么这么写 |
|---|---|
| `Enabled` | 不改级别逻辑，**原样透传** inner。不关心的方法就转发，是 wrapping handler 的标准写法 |
| `Handle` | **唯一干活的地方**。只有它同时拿得到 ctx（字段来源）和 Record（往哪加字段）。这是 slog 里唯一能"根据 ctx 动态改日志内容"的位置 |
| `WithAttrs` | 必须 `inner.WithAttrs(...)` 后**再用 contextHandler 包回去** |
| `WithGroup` | 同上，重新包装自己 |

### 7.2 最高发的回归 bug：WithAttrs/WithGroup 忘了重新包装

如果图省事把 `WithAttrs` 写成 `return h.inner.WithAttrs(attrs)`，返回的是**裸的底层 handler**，`contextHandler` 这层被剥掉——之后 `logger.With(...)` 出来的子 logger 打日志**不再自动注入 ctx 字段**。这是自定义 handler 最高发的回归 bug。必须"委托底层 + 重新包装自己"两步都做。官方 guide 反复强调这一点。

### 7.3 已知限制：WithGroup 下注入字段会被嵌进 group

这是 wrapping handler 注入 ctx 字段的一个已知陷阱，值得写下来。`Handle` 里注入的字段是通过 `rec.AddAttrs` 追加的，会落在**当前 group 作用域**里。如果调用链上先 `logger.WithGroup("g")` 再打日志，注入的 request_id/tenant_id 会被嵌进 `"g"` 对象而非顶层：

```json
{"msg":"...", "g": {"request_id":"req-123", "biz_field":1}}   // 预期 request_id 在顶层
```

彻底解法是在 handler 里记录 group 深度、把身份字段固定注入到最外层——复杂度不低。项目里几乎不用 `WithGroup`，所以选择**记一笔约定**（"需要顶层身份字段的 logger 不要套 WithGroup"）而非过度设计。这也是一种取舍：明确一个边界，比为小概率场景增加一堆代码更划算。

### 7.4 字段从哪来：两个来源

注入的字段来自两处，都实现同一个 `ContextExtractor` 签名：

```go
// config.go —— 日志包只认这个签名，不认具体字段名
type ContextExtractor func(ctx context.Context) []slog.Attr
```

**来源一：WithField / WithFields（内置）**——临时字段，存进 ctx 的私有 key：

```go
// pkg/xslog/context.go
type ctxKey struct{}
var fieldsKey = ctxKey{}

func WithField(ctx context.Context, key string, val any) context.Context {
    return WithFields(ctx, slog.Any(key, val))
}

func WithFields(ctx context.Context, attrs ...slog.Attr) context.Context {
    if len(attrs) == 0 {
        return ctx
    }
    existing := fieldsFromContext(ctx)
    // 复制一份，避免多个子 ctx 共享同一底层数组产生数据竞争。
    merged := make([]slog.Attr, 0, len(existing)+len(attrs))
    merged = append(merged, existing...)
    merged = append(merged, attrs...)
    return context.WithValue(ctx, fieldsKey, merged)
}
```

> 那次 `make` 复制不是多余的：`append` 可能复用底层数组，多个子 ctx 从同一父 ctx 派生时会踩到同一段内存，产生数据竞争。复制一份切断共享，是并发安全的必要代价。

**来源二：ContextExtractor（用户注册）**——系统身份字段，由 `xcontext` 包提供 extractor：

```go
// pkg/xcontext：身份字段的唯一真相源，顺带提供一个日志 extractor
func LogExtractor(ctx context.Context) []slog.Attr {
    return []slog.Attr{
        slog.String("request_id", GetRequestID(ctx)),
        slog.String("tenant_id", GetTenantID(ctx)),
    }
}
```

### 7.5 为什么要 ContextExtractor 这层抽象（依赖倒置）

最直白的写法是让 `contextHandler` 直接读 `requestIDKey`/`tenantIDKey`。但那样**日志包就得 import 定义这些 key 的包（xcontext）**，字段一多、要接 OTel，这个 import 越来越重。

`ContextExtractor` 把"从 ctx 提取字段"抽象成一个函数签名，日志包只认签名、不认具体字段。由 **xcontext 提供 extractor、组装层（main）把两者粘起来**——日志包不 import xcontext，xcontext 也不 import 日志包：

```go
// main：组装层做胶水
log := xslog.New(xslog.Config{
    ContextExtractors: []xslog.ContextExtractor{
        xcontext.LogExtractor, // request_id/tenant_id
        // otelExtractor,      // 将来接 OTel：再注册一个读 trace_id/span_id 的，日志包不改一行
    },
})
```

这层抽象**不是 slog 官方 API**，是依赖倒置的工程设计。好处是日志包彻底不依赖业务字段、可跨项目直接复制；代价是每条日志多一次 extractor 遍历（可忽略）。

### 7.6 收益：调用代码变干净

装上后，只要用 context 变体，字段自动出现：

```go
// service 层——不用手写 request_id！只要 ctx 是请求传下来的，自动带
func (s *UserService) Get(ctx context.Context, id string) (*User, error) {
    user, err := s.repo.Get(ctx, id)
    if err != nil {
        s.log.ErrorContext(ctx, "get user failed", slog.String("user_id", id), xslog.Err(err))
        return nil, err
    }
    return user, nil
}
```

GORM 的 SQL 日志（用 `DebugContext(ctx, ...)` 桥接）也自动带 request_id——一条慢查询能直接关联到哪个请求。

---

## 8. 接入 Gin：中间件三件套

自动注入的前提是"中间件先把 request_id 写进 ctx"。Gin 侧要三个中间件配合。

### 8.1 关键前置：用 gin.New() 不用 gin.Default()

`gin.Default()` 会装 gin 自带的 Logger 中间件，往 stdout 写 `| 200 | 1.2ms | /api/x` 这种人读格式，和你的结构化访问日志重复且冲突。从裸 `gin.New()` 开始，装自己的：

```go
gin.SetMode(cfg.Server.Mode)
r := gin.New()
r.Use(RequestID())   // 1. 最先：生成/透传 request_id，写进 ctx
r.Use(AccessLog(log))// 2. 依赖 1：记 start → c.Next() → 算 duration
r.Use(Recovery(log)) // 3. 依赖 1：panic 时记 stack
```

**RequestID 必须最先**——后面所有中间件的日志都要带 request_id，得先有人把它放进 ctx。

### 8.2 RequestID：桥接 gin.Context 与标准 context

```go
func RequestID() gin.HandlerFunc {
    return func(c *gin.Context) {
        rid := c.GetHeader("X-Request-ID")
        if rid == "" {
            rid = randomID(8) // 生产可换 google/uuid
        }
        // 关键：写进标准 context，slog 的 contextHandler 才读得到
        ctx := xcontext.WithRequestID(c.Request.Context(), rid)
        c.Request = c.Request.WithContext(ctx)
        c.Header("X-Request-ID", rid) // 回写响应头，客户端可凭此查日志
        c.Next()
    }
}
```

> 想让前端拿到 `X-Request-ID`，CORS 要加 `Access-Control-Expose-Headers: X-Request-ID`，否则跨域下自定义响应头被浏览器挡掉。

### 8.3 AccessLog / Recovery

两个中间件都用 `InfoContext(c.Request.Context(), ...)` / `ErrorContext(...)`——因为 request_id 已经在 ctx 里，**它们不用再手写 request_id，contextHandler 自动带上**：

```go
func AccessLog(log *slog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        log.InfoContext(c.Request.Context(), "HTTP request",
            slog.String("method", c.Request.Method),
            slog.String("path", c.Request.URL.Path),
            slog.Int("status", c.Writer.Status()),
            slog.Duration("duration", time.Since(start)),
        )
    }
}

func Recovery(log *slog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if r := recover(); r != nil {
                log.ErrorContext(c.Request.Context(), "panic recovered",
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

对比 §7 之前的写法，这里每条日志少了一行 `slog.String("request_id", ...)`——这就是自动注入的红利。

---

## 9. 测试：bytes.Buffer + 官方 slogtest

日志最稳的测法：把 `Config.Output` 换成 `bytes.Buffer`，断言输出内容。这也是 §2 把 `Output` 放进 Config 的回报：

```go
func newTestLogger(level string) (*slog.Logger, *bytes.Buffer) {
    var buf bytes.Buffer
    l := xslog.New(xslog.Config{Level: level, Format: "json", AddSource: true, Output: &buf})
    return l, &buf
}
```

除了功能用例（级别过滤、context 注入、脱敏、source 裁剪与脱敏并存），**一定要跑官方 `slogtest.TestHandler`**——它用一批标准用例验证你的自定义 handler 符合 slog.Handler 规范（比如空 attr 处理、group 语义），是自定义 handler 的合规底线：

```go
func TestHandlerCompliance(t *testing.T) {
    var buf bytes.Buffer
    logger := xslog.New(xslog.Config{Format: "json", Output: &buf})
    results := func() []map[string]any { /* 解析 buf 每行 JSON */ }
    if err := slogtest.TestHandler(logger.Handler(), results); err != nil {
        t.Fatal(err)
    }
}
```

### 9.1 用 sloglint 把「最佳实践」变成 CI 卡点

`slogtest` 保证 handler *实现* 合规，但挡不住 *调用方* 写错——§1.2 那个 `slog.Info("x", "k")`（漏了 value）能编译通过、运行时才静默出脏日志。这类问题要靠**静态检查**在提交前拦下。工具是 [`sloglint`](https://github.com/go-simpler/sloglint)，通过 [`golangci-lint`](https://golangci-lint.run) 跑起来：

```yaml
# backend/.golangci.yml
version: "2"          # golangci-lint v2 必须显式声明版本（v1 无此字段）

run:
  timeout: 5m

linters:
  enable:
    - govet
    - staticcheck
    - sloglint        # ← 强制 slog 最佳实践
  settings:
    sloglint:
      no-mixed-args: true   # 禁止 key-value 混用，强制强类型 Attr
      context-only: true    # 强制用 InfoContext 等 context 变体
```

两条规则正好对应本文两个核心约定：

- **`no-mixed-args`** 兜住 §1.2 的坑：`log.Info("msg", "k")`（奇数参数）直接报错，逼你写 `slog.String("k", v)`。
- **`context-only`** 兑现 §7 的前提：不带 ctx 的 `log.Info(...)` 报错，逼你全程用 `InfoContext`——只有 ctx 传进来，request_id 才注入得进去，将来接 OTel 的 trace_id 也走同一条路（§10）。这条让「养成用 context 变体的习惯」从口头建议变成合并前的硬门槛。

**这里要分清哪些是官方的、哪些不是**，否则容易误以为「Go 要求你这么做」：

| 工具 | 出身 | 性质 |
|---|---|---|
| `testing/slogtest` | **Go 标准库**（官方） | 官方合规测试，写自定义 handler 必跑 |
| `golangci-lint` | 社区项目（非 Go 官方） | Go 生态**事实标准**的 lint 聚合器：一次运行几十个 linter |
| `sloglint` | 第三方（go-simpler） | 挂在 golangci-lint 下的一个 slog 专用 linter |

换句话说：**`slogtest` 是官方要求的合规底线；`golangci-lint` + `sloglint` 是社区工具，不是 Go 强制的**。但 golangci-lint 已是 Go 项目近乎默认的 CI 环节，`sloglint` 也是目前落地 slog 规范最省事的方案，所以本项目采用——它约束的是「团队怎么调用 slog」，不是「slog 本身要不要这样」。

> 注意 golangci-lint **v2 与 v1 配置不兼容**：v2 必须写 `version: "2"`，且 `gofmt` 这类改成了 formatter（放独立的 `formatters` 段，不再列在 `linters` 里）。踩过一次「unsupported version」和「gofmt is a formatter」的报错，记一笔。Makefile 里 `lint: golangci-lint run ./...` 直接跑这份配置。

---

## 10. 接入 OpenTelemetry：何时接、怎么接

（选型理由——为什么 slog、为什么不 zerolog/zap、为什么封装——见开篇 §0。本节只谈 OTel：何时该接、真到那步三个库各是什么处境、以及链路追踪的真正难点。）

### 什么阶段不需要 / 什么阶段该接 OTel

| 阶段 | 判断 | 做什么 | YAGNI 原则 |
|---|---|---|---|
| **单体 + 单实例 + 未上线/用户少** | 当前项目 | xslog 留 ContextExtractor 口子 | ❌ OTel 是纯负担（collector、instrument）|
| **单体 + 生产，有真实用户** | "某接口偶尔慢" | Prometheus metrics（QPS/延迟/错误率大盘）| trace 暂缓，性价比：metrics >> trace |
| **拆了微服务 / 复杂异步调用** | 跨服务/进程链路 | Trace（TraceExtractor + otelgin + 跨服务传播） | 这时没 trace 真的抓瞎 |

**核心洞察**：单体里请求链路在一个进程，`ERROR` 日志 + 堆栈 + request_id 已够定位；trace 的价值在**跨服务/进程**时才爆发。

### 到了微服务这步：三个库的 OTel 适配现实

真到了接 OTel 这步，三个库的现实约束分两层：

**Level 1：日志-trace 关联（把 trace_id 打进日志）**
- zerolog：手动 hook 从 span 取 trace_id，每加信号回中间件改。能做，但不如 ContextExtractor 干净。
- zap：`otelzap` 桥接能自动带，比 zerolog 好。
- slog xslog：`ContextExtractor` 注册一行搞定，信号可扩展（request_id / trace_id / span_id 全走同一机制）。

这层差异不大，都是"能做，工作量有别"。

**Level 2：日志作为 OTel 信号走 OTLP 管道 — 这层是分水岭**
- **zerolog：没有官方 OTLP log bridge**。要么自己写 hook 发 collector（维护序列化 + 批量 + 重试），要么就是那个慢 46× 的 slog 桥接陷阱。这是 zerolog 在 OTel 时代最大的短板。
- zap：`otelzap` 主要解决"日志带 trace_id"，OTLP log 直发靠社区桥，不如 slog 官方。
- **slog：`otelslog` 是 OTel 官方维护的 bridge**，一等公民。换 inner handler 就接上，调用点不动。

**结论**：微服务 + 统一可观测性平台时，slog 的官方桥是三者里最顺的。zerolog 的流式 API 和极致性能为"单机极速打印"优化，不为"标准化信号管道"设计。

### 链路追踪的真正难点

加 span 本身很简单（`tracer.Start` + `defer span.End()`）。难点在：

1. **Context 传播**：trace 断链头号原因。ctx 必须全程透传，一处 `context.Background()` 就断（常见于异步 goroutine、消息队列消费）。
2. **跨服务边界传递**：HTTP header（`traceparent` W3C 标准）、消息队列 message header 手动注入。
3. **采样策略**：全量采集成本爆炸，1% 采样怕漏错误请求，需要 tail-based sampling（配置复杂）。

**核心难点是 1 和 2**（做不好就没意义），3 是成本控制。完整集成路径（TraceExtractor 实现、otelgin 中间件、多路分发、微服务传播、采样策略）见 [OpenTelemetry 集成路径](../saas-backend/research/logging/06-OpenTelemetry集成路径.md)。

**前提**：全程用 `*Context` 变体（`InfoContext` 等），且 ctx 里有活跃 span。这也是本文反复强调"用 context 变体"的原因——现在养成习惯，将来接 OTel 零成本。

---

## 11. 总结：踩坑清单

| 坑 / 决策 | 表现 | 最终做法 |
|---|---|---|
| 默认格式落到 Text | 生产漏配 format 吐出非 JSON，采集器懵 | `switch` + `default: JSON`，向生产行为兜底 |
| parseLevel 大小写 | `level: ERROR` 静默降级成 info | `TrimSpace` + `slog.Level.UnmarshalText`，大小写不敏感 |
| `*slog.LevelVar` 空留 | 造了指针没人持有，"热改级别"从未接通 | 退回 `slog.Level`，别预留半成品口子（YAGNI） |
| source 裁成纯文件名 | 十几个 `service.go` 无法区分 | 裁成 `dir/file:line`，保留一级目录消歧 |
| ReplaceAttr 只有一个 | source 裁剪与脱敏二选一 | `chainReplaceAttr` 串联，key 不重叠 |
| 补 Fatal 图方便 | `os.Exit` 跳过 defer、测试不友好 | 删掉，改 `run() error` 模式，日志只记录不控流程 |
| 静态字段进 Config | Config 膨胀成"全项目字段集合" | service/env 由调用方 `.With(...)`，包只管机制 |
| WithAttrs 忘重新包装 | `logger.With(...)` 子 logger 丢注入 | 委托底层 + 重新包 contextHandler，两步都做 |
| WithGroup 下注入 | 身份字段被嵌进 group 对象 | 记一笔约定，不为小概率场景过度设计 |
| 直接读 ctx key | 日志包 import xcontext，耦合业务 | `ContextExtractor` 依赖倒置，包不认字段名 |

一条主线贯穿所有决策：**日志包只管"机制"（怎么注入、怎么裁 source、怎么脱敏、怎么定级别），不管"业务内容"（带哪些字段、字段叫什么、什么算致命）。** 守住这条边界，包就能保持职责单一、可跨项目复制。

---

## 12. 参考

- [日志库选型调研（本文选型依据）](../saas-backend/research/logging/01-日志库选型调研.md)
- [Structured Logging with slog — The Go Blog](https://go.dev/blog/slog)
- [A Guide to Writing slog Handlers — Go 官方](https://github.com/golang/example/blob/master/slog-handler-guide/README.md)
- [`log/slog` — pkg.go.dev](https://pkg.go.dev/log/slog)
- [go-simpler/sloglint — slog 调用静态检查](https://github.com/go-simpler/sloglint)
- [otelslog — OTel bridge](https://pkg.go.dev/go.opentelemetry.io/contrib/bridges/otelslog)

---

*系列教程：[saas-backend](../saas-backend/README.md)*
