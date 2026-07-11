# 应用启动与组件生命周期：main 组合根与 gin/cron 编排

> 2026-07。本文回答一个实际问题——**新项目 `backend/` 的一个 `main` 要同时管住 gin HTTP、cron 同步任务、以及未来的后台组件（WS/worker），怎么组织才既统一在一处、又不重蹈老项目「god-App + 15 步 init + 局部变量吞掉优雅关闭」的覆辙？**
>
> 结论先行（成熟方案优先）：**`run()` 组合根 + 长驻组件编排 + `signal.NotifyContext` 优雅退出**。基础设施（config/log/db/redis）在 `run()` 里顺序创建、`defer` 逆序清理；长驻组件（HTTP、cron、将来 WS/worker）的启停用 **[`oklog/run` run.Group](https://github.com/oklog/run)**（Grafana/Prometheus 采用的 actor 模式）或 `errgroup.WithContext` 编排——**HTTP+scheduler 组合已足够匹配 run.Group**，不是「组件 ≥3 才值得」。`backend/` 现有的裸 `select` + `Server{Start/Stop}` 是最小可行版，但成熟方案是 run.Group（见第三节对比）。
>
> 关联：`cmd/server/main.go`、`internal/server/server.go`、[README](../../README.md) 核心设计原则 #1 组合根 / #9 优雅退出 / #10 零全局状态、[step-01-project-scaffolding](../../step-01-project-scaffolding.md)、[step-02-config-logger-db](../../step-02-config-logger-db.md)。依赖装配见 [02](./02-依赖装配与组件注册-手工注入vs-wire-fx.md)，cron 细节见 [03](./03-cron任务注册与复用-jobs装配与Server生命周期集成.md)。
>
> 日常写代码不必读；想搞明白「一个 main 怎么编排多个长驻组件、run.Group vs errgroup vs 裸 select 怎么选」时看这里。

## 结论速览

一个进程内有三类东西，生命周期各不相同，别混为一谈：

| 类别 | 例子 | 谁持有 | 怎么清理 |
|------|------|--------|----------|
| 基础设施（被动资源） | config、log、db pool、redis client | `run()` 局部变量 | `defer` 逆序 Close |
| 可启停组件（主动 goroutine） | HTTP server、cron、WS hub | `Server` 结构体字段 | `Server.Stop(ctx)` 逆序优雅关闭 |
| 请求级 / 任务级 | 每个 HTTP handler、每个 cron job | 不持有，随调用产生消亡 | ctx 取消传播 |

**组合根在 `main`（的 `run()`），编排交给一个 group orchestrator**。`main` 只做两件事：把资源按序创建好，然后把每个长驻组件（HTTP、scheduler、worker、信号）注册成一个 actor，任一退出即触发全体优雅关闭。这条边界是本文的全部要点。

编排机制的取舍（详见第二、四节）：
- **单组件（只有 HTTP）**：`signal.NotifyContext` + `select{srvErr / ctx.Done}` 足矣，零依赖、最直白。
- **多组件（HTTP + scheduler/worker，本项目的实际方向）**：**`oklog/run` 的 run.Group 是行业标准默认**（Grafana/Prometheus/Thanos/Cortex 都用它），`errgroup.WithContext` 是等价备选。两者都是「任一 actor 退出 → 触发全体 interrupt」的 actor 模型。
- **否掉 fx**：运行期反射容器，与显式装配/零全局状态取向冲突（见 [02](./02-依赖装配与组件注册-手工注入vs-wire-fx.md)）。

## 一、现状：`backend/` 的组合根已经做对了大半

新项目 `cmd/server/main.go` 的骨架（现有代码，本文只是把它的设计讲清楚，不改）：

```go
func main() {
    // main 只负责退出码：run 返回 error → 打日志 + 非零码退出，让容器/systemd 感知重启
    if err := run(); err != nil {
        slog.Error("server exited with error", slog.Any("err", err))
        os.Exit(1)
    }
}

func run() error {
    cfg, err := config.InitConfig()          // 1. 配置
    log := xslog.New(...)                      // 2. 日志
    db, err := xgorm.New(...)                  // 3. DB
    defer xgorm.Close(db)                      //    逆序清理
    rdb, err := xredis.New(...)                // 4. Redis
    defer rdb.Close()

    srv, err := server.New(server.Options{Config: cfg, DB: db, RDB: rdb, Log: log})  // 5. 组装

    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    srvErr := make(chan error, 1)              // 缓冲 1，Start 失败时 goroutine 不泄漏
    go func() {
        if err := srv.Start(); err != nil { srvErr <- err }
    }()

    select {                                   // 二选一：出错 或 收到信号
    case err := <-srvErr:
        return fmt.Errorf("start server: %w", err)
    case <-ctx.Done():
        stop()
        log.Info("shutdown signal received")
    }

    shutdownCtx, cancel := context.WithTimeout(context.Background(),
        time.Duration(cfg.Server.GracefulTimeout)*time.Second)
    defer cancel()
    srv.Stop(shutdownCtx)                      // 带超时的优雅关闭
    return nil
}
```

这套骨架的四个正确决策，逐条对应 README 原则：

1. **`main` 只管退出码，逻辑全在 `run() error`**（原则 #1）。`os.Exit(1)` 只出现在 `main`，因为 `os.Exit` 会跳过 `defer`；把逻辑放进 `run` 才能让所有 `defer` 逆序执行。`xslog` 刻意不提供 `Fatal`，就是逼所有失败走 `return err` 这一条路。
2. **基础设施用 `defer` 逆序清理**（原则 #10 零全局状态）。db、redis 是 `run` 的局部变量，不是全局单例，不是 `App` 字段——谁创建谁 `defer` 关，顺序天然正确（redis 先关、db 后关，与创建相反）。
3. **`signal.NotifyContext` + `select`**（原则 #9 优雅退出）。信号变成 `ctx.Done()`，和「server 出错」`srvErr` 一起塞进一个 `select`——无论哪个先到都走同一段收尾逻辑。
4. **带超时的 `srv.Stop(shutdownCtx)`**。`GracefulTimeout` 秒内排空在途请求，到点强制退出。**这正是老项目 B 缺失的东西**（见第三节）。

`server.go` 里，`Server` 只持可启停组件 + 它们需要的依赖，并留好了 cron 的钩子：

```go
type Server struct {
    httpSrv *http.Server
    db      *gorm.DB
    rdb     *redis.Client
    log     *slog.Logger
    // Step 05 起追加：cronRunner *cron.Runner
}

func (s *Server) Start() error {
    // Step 05 起追加：go s.cronRunner.Start(ctx)
    if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        return err
    }
    return nil
}

func (s *Server) Stop(ctx context.Context) {
    // Step 05 起逆序追加：s.cronRunner.Stop()
    if err := s.httpSrv.Shutdown(ctx); err != nil {
        s.log.Error("server shutdown error", slog.Any("err", err))
    }
}
```

**cron/WS 未来就顺着 `// Step 05` 这两行注释长出来**：字段加一个、`Start()` 里拉起、`Stop()` 里逆序停。`main.go` 一个字都不用改——这就是「统一在一个 main 下管理」的正解：不是把所有东西堆进 main，而是 main 只认 `Server.Start/Stop` 两个动词，组件在 `Server` 内部各自封装。

## 二、协调机制对比：按组件数分档选

「多个组件跑在一个进程、怎么统一启停」在 Go 生态有几种主流答案。逐维度对比：

| 方案 | 启动错误传播 | 关闭顺序控制 | HTTP 优雅关闭 | 心智负担 | 额外依赖 |
|------|------------|------------|--------------|---------|---------|
| 老项目 B：`signal.Notify` 裸 channel + `App.Close()` | ❌ HTTP 起在 `go` 里，`os.Exit` 直接吞错 | ⚠️ 手工 `Close()`，但 HTTP 不在其中 | ❌ **没有**（见第三节） | 低 | 无 |
| **本项目：`signal.NotifyContext` + `select{srvErr/ctx.Done}`** | ✅ `srvErr` channel 回传 | ✅ `Stop()` 内显式逆序 | ✅ `Shutdown(ctx)` 带超时 | 低 | 无（纯标准库） |
| `errgroup.WithContext` | ✅ 任一 goroutine 返回 err 即 cancel | ⚠️ 靠 ctx 广播，顺序不显式 | ✅ 可做 | 中 | `golang.org/x/sync` |
| `oklog/run` run.Group（actor 模式） | ✅ 任一 actor 退出即触发全体 | ✅ 每 actor 配对 `execute/interrupt` | ✅ 可做 | 中高 | `github.com/oklog/run` |

关键判断，按组件数分档：

- **只有 HTTP 一个长驻组件**：`select{srvErr / ctx.Done}` + `Stop()` 已经够，零依赖最直白，没必要引 orchestrator。
- **HTTP + scheduler/worker（本项目实际方向）**：一旦有第二个对等长驻组件，「任一失败拉全体优雅退出」就成了刚需——这正是 `oklog/run` run.Group 的本职。此时 run.Group **不是过度设计，而是行业标准默认**：Grafana、Prometheus、Thanos、Cortex 的 `func main` 都用它编排一堆长驻组件（见 [gopheracademy run.Group](https://blog.gopheracademy.com/advent-2017/run-group/)、[oklog/run](https://github.com/oklog/run)）。它把「怎么跑」`execute()` 和「怎么被打断」`interrupt()` 成对绑定，`g.Run()` 在任一 actor 返回时调用其余所有 actor 的 interrupt——比手写「N 个 channel + 大 select + 逐个 Stop」更标准、更不易漏。

`signal.NotifyContext`（Go 1.16+）在两档里都保留：它把信号变成可传播、可被下游 ctx 感知的 `context`（run.Group 里则用 `run.SignalHandler` 这个现成 actor 承接），配合 `defer stop()` 还能在退出后恢复默认信号行为（再按一次 Ctrl-C 能强杀卡住的进程）。

> 关于 README 原则 #1/#9/#10：它们（组合根在 main、优雅退出、零全局状态）本身与上述行业实践一致——run.Group 正是「组合根在 main 里编排长驻组件」的成熟实现。本文把它们当**与行业一致的既有约定**引用，而非「因为规定所以照做」。

## 三、组件归属对比：god-App vs run() + Server 两层

老项目（A/B）用一个 `App` god-struct 把**所有**长生命周期对象拍平塞进去：

```go
// 老项目 B 现状（反面）：14 字段的 App，什么都往里塞
type App struct {
    Config, Router, DB, Redis, JWT, Enforcer, RSACipher,
    Cron, StorageProvider, WSServer, UploadSem, Handlers // ...共 14 个
}
func NewApp() (*App, error) {  // ~15 步编号 init：1 config → 2 log → ... → 9 router
    // 每步一个 init 方法，任一失败 return wrap error
}
```

问题不在「编号 init 可读」（这点其实不坏），而在**边界糊了**：db/redis 这种被动资源、cron/WS 这种主动 goroutine、Handlers 这种请求级装配，全挤在一个结构体里，谁该 `defer` 关、谁该 `Stop`、关闭顺序如何，全靠人肉记忆。最典型的塌方就是下一节的 HTTP 优雅关闭。

本项目的两层划分：

| 层 | 持有什么 | 生命周期动作 | 谁写 |
|----|---------|------------|------|
| `run()`（组合根） | config/log/db/redis 等被动资源 | 顺序创建 + `defer` 逆序 Close | `cmd/server/main.go` |
| `Server`（编排器） | http/cron/WS 等主动组件 + 其依赖 | `Start()` 拉起 / `Stop(ctx)` 逆序优雅关闭 | `internal/server/server.go` |

**边界规则一句话**：需要 `Close()`/连接池的被动资源留在 `run()`（defer 最适合）；有自己 goroutine、需要「优雅」关闭的主动组件进 `Server`。db 不进 `Server`——它没有「优雅关闭」语义，`Close()` 一把梭即可，放 `run()` 的 defer 里最省心。

## 四、成熟编排方案：run.Group / errgroup / 裸 select

「一个 `main` 编排 N 个长驻组件」是个有成熟解的问题，三种主流方案按组件数与场景选：

| 方案 | 错误传播 | interrupt/关闭 | 社区采用 | 依赖 | 适用 |
|------|---------|--------------|---------|------|------|
| 裸 `select{srvErr/ctx.Done}` + `Stop()` | 手写 channel 回传 | 手写逐个 `Stop()` | 教程常见 | 无 | **单**长驻组件（仅 HTTP） |
| **`oklog/run` run.Group** | ✅ 任一 actor 返回即触发全体 interrupt | ✅ 每 actor 配对 `execute/interrupt` | ✅ Grafana/Prometheus/Thanos/Cortex | `github.com/oklog/run`（极小，单文件） | **HTTP + scheduler/worker（本项目）** |
| `errgroup.WithContext` | ✅ 任一返回 err 即 cancel ctx | ⚠️ 靠 ctx 广播，interrupt 不成对、顺序不显式 | ✅ 广泛 | `golang.org/x/sync` | 一批**会自然结束**的并发任务 |

**推荐：组件 ≥2 用 run.Group。** 本项目 HTTP + scheduler（cron/worker，见 [03](./03-cron任务注册与复用-jobs装配与Server生命周期集成.md)）恰好落在这。run.Group 相比 errgroup 的关键优势是 `interrupt(error)` 与 `execute()` **成对**——长驻服务的「如何优雅打断」是显式写出来的（`httpSrv.Shutdown`、`scheduler.Shutdown`），而不是像 errgroup 那样全靠每个 goroutine 自己监听 ctx。这对「一批需要被信号打断的常驻服务」更贴切；errgroup 更适合「一批会跑完的任务」。

**何时 run.Group 也不必要**：整个进程只有 HTTP 一个长驻组件、且短期不会加第二个——裸 `select` 版足够，别为「将来可能」预付依赖。判据就是组件数：第二个长驻组件出现之日，就是换 run.Group 之时。

**单副本 vs 多副本**：编排机制（select / run.Group / errgroup）都是**进程内**的，与部署几个副本无关——它只管「本进程内这几个组件怎么一起启停」。多副本下「同一任务只让一个副本跑」是**调度层**的事（gocron 的 Locker、River 的 leader 选举，见 [03](./03-cron任务注册与复用-jobs装配与Server生命周期集成.md)），不在生命周期层。所以本文结论对单副本、多副本都成立：单副本时 scheduler 直接跑，多副本时 scheduler 照跑、由调度库内建的协调决定哪个副本实际执行。

### 被否的重方案：go-kratos 式两层 `internal/app/` 编排

kratos 用 `internal/app` 做应用装配 + `kratos.App` 做生命周期，是为「多传输层（HTTP+gRPC）+ 服务注册发现 + 配置中心」的微服务全家桶设计的。本项目是单体 admin BFF，`run() + Server` 一层足矣。引入 kratos 的 `App`/`Server`/`Transport` 抽象，是在没有对应问题时先付抽象税。

## 五、成本与 tradeoff（已知并接受）

- **run.Group 每个 actor 是一个 `execute/interrupt` 对**：HTTP 的 `ListenAndServe()` 阻塞在 `execute` 里，对应的 `interrupt` 是 `httpSrv.Shutdown(ctx)`；scheduler 的 `execute` 拉起调度后阻塞等 ctx，`interrupt` 是 `scheduler.Shutdown()`（drain 在途任务）。关闭顺序不靠手写逐个 `Stop`，而是 group 在任一 actor 退出时并发调用所有 interrupt。scheduler/worker 的选型与集成见 [03](./03-cron任务注册与复用-jobs装配与Server生命周期集成.md)。
- **裸 `select` 版只处理「首个」事件**（单组件档的已知限制）：收到信号后若关闭期间又崩了另一个组件，不会再被 `select` 捕获——这正是多组件该上 run.Group 的理由之一（group 统一收敛所有 actor 的退出）。
- **反面备忘（老项目 B 的坑，新项目勿犯）**：老项目里按需拉起的 worker 用 `go func()` + `context.Background()`，`Background()` 永不取消，SIGTERM 时这些 goroutine 完全不受控、直接被进程消灭。**本项目所有后台 goroutine 必须接 `Server` 传下来的可取消 ctx**，让优雅关闭能穿透到最底层任务。
