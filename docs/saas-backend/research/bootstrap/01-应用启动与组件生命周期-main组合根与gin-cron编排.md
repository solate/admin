# 应用启动与组件生命周期：main 组合根与 gin/cron 编排

> 2026-07。本文回答一个实际问题——**新项目 `backend/` 的一个 `main` 要同时管住 gin HTTP、cron 同步任务、以及未来的后台组件（WS/worker），怎么组织才既统一在一处、又不重蹈老项目「god-App + 15 步 init + 局部变量吞掉优雅关闭」的覆辙？**
>
> 结论先行（更简设计优先）：**`run()` 组合根 + `signal.NotifyContext` 优雅退出 + 按组件数选编排机制**。基础设施（config/log/db/redis）在 `run()` 里顺序创建、`defer` 逆序清理；长驻组件（HTTP、cron、将来 WS/worker）的启停按数量分档——**单组件用裸 `select`（零依赖、最直白），两个及以上用 [`errgroup.WithContext`](https://pkg.go.dev/golang.org/x/sync/errgroup)（头号推荐，唯一依赖 `x/sync`）**，只有当 actor 众多、需要显式 interrupt 排序时才升级到 [`oklog/run`](https://github.com/oklog/run)。`backend/` 现有代码已是 **`errgroup.WithContext` + `Server.Run(ctx)`**——因为 scheduler（Step 05）是已知的第二个长驻组件，day-1 直接用 errgroup、避免日后从 select 改写（这属**提前采用**，单组件下 errgroup 并不比裸 `go func` 简单，理由与代价见第二节末「单组件下 errgroup 并不更简单」小节）。
>
> 关联：`cmd/server/main.go`、`internal/server/server.go`、[README](../../README.md) 核心设计原则 #1 组合根 / #9 优雅退出 / #10 零全局状态、[step-01-project-scaffolding](../../step-01-project-scaffolding.md)、[step-02-config-logger-db](../../step-02-config-logger-db.md)。依赖装配与分层见 [02](./02-依赖装配与分层对象化-手工注入与struct封装.md)，后台任务执行模型见 [cron/02](../cron/02-单机执行模型-从ticker到持久队列的四档.md)。
>
> 日常写代码不必读；想搞明白「一个 main 怎么编排多个长驻组件、select vs errgroup vs run.Group 怎么选」时看这里。

## 结论速览

一个进程内有三类东西，生命周期各不相同，别混为一谈：

| 类别 | 例子 | 谁持有 | 怎么清理 |
|------|------|--------|----------|
| 基础设施（被动资源） | config、log、db pool、redis client | `run()` 局部变量 | `defer` 逆序 Close |
| 可启停组件（主动 goroutine） | HTTP server、cron、WS hub | 编排器（`errgroup` 各一个 goroutine） | ctx 取消 + 各自带超时优雅关闭 |
| 请求级 / 任务级 | 每个 HTTP handler、每个 cron job | 不持有，随调用产生消亡 | ctx 取消传播 |

**组合根在 `main`（的 `run()`），编排按组件数选最简的那档**。`main` 只做两件事：把资源按序创建好，然后把每个长驻组件（HTTP、scheduler、worker）跑起来、任一退出即触发全体优雅关闭。这条边界是本文的全部要点。

编排机制按组件数分档（详见第二、四节）：

- **单组件（只有 HTTP）**：`signal.NotifyContext` + `select{srvErr / ctx.Done}` 足矣，零依赖、最直白——多数 Gin boilerplate/教程的做法。
- **两个及以上（HTTP + scheduler/worker，本项目的实际方向）**：**`errgroup.WithContext` 是头号推荐**。一个依赖（`golang.org/x/sync`）、一个心智模型（「任一 goroutine 返回非 nil → cancel 共享 ctx → 其余组件收到 ctx.Done 各自收尾」），没有额外词汇表。**`backend/` 现状即此档**——scheduler（Step 05）是已知的第二个长驻组件，故 day-1 直接用 errgroup（提前采用，见第二节末小节）。
- **升级项 `oklog/run`**：actor 众多、且需要精确控制 interrupt 触发/排序时才值得（Grafana/Prometheus 级），本项目短期用不上。
- **否掉 fx**：运行期反射容器，与显式装配/零全局状态取向冲突（见 [02](./02-依赖装配与分层对象化-手工注入与struct封装.md)）。

## 一、现状：`backend/` 的组合根已经做对了大半

新项目 `cmd/server/main.go` 的骨架（现有代码，本文只是把它的设计讲清楚）。`main` 只管退出码，`run()` 分两层——① 基础设施顺序创建 + `defer` 逆序 Close，② 长驻组件交 `srv.Run(ctx)` 编排：

```go
func main() {
    // main 只负责退出码：run 返回 error → 打日志 + 非零码退出，让容器/systemd 感知重启
    if err := run(); err != nil {
        slog.Error("server exited with error", slog.Any("err", err))
        os.Exit(1)
    }
}

func run() error {
    // ===== 第一层：基础设施（被动资源）—— defer 逆序 Close =====
    cfg, err := config.InitConfig()            // 1. 配置
    log := xslog.New(...)                       // 2. 日志
    db, err := xgorm.New(xgorm.Config(cfg.Database), log)  // 3. DB
    defer xgorm.Close(db)                       //    逆序清理
    rdb, err := xredis.New(xredis.Config(cfg.Redis))       // 4. Redis
    defer rdb.Close()

    // ===== 第二层：长驻组件（主动 goroutine）—— Server 的 errgroup 编排 =====
    srv, err := server.New(server.Options{Config: cfg, DB: db, RDB: rdb, Log: log})  // 5. 组装
    if err != nil {
        return fmt.Errorf("init server: %w", err)
    }

    // signal.NotifyContext 把 SIGINT/SIGTERM 变成可传播的 ctx
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    // Run 阻塞直到收到信号或任一组件出错，内部 errgroup 统一优雅关闭
    if err := srv.Run(ctx); err != nil {
        return fmt.Errorf("run server: %w", err)
    }
    log.Info("server exited")
    return nil
}
```

这套骨架的四个正确决策，逐条对应 README 原则：

1. **`main` 只管退出码，逻辑全在 `run() error`**（原则 #1）。`os.Exit(1)` 只出现在 `main`，因为 `os.Exit` 会跳过 `defer`；把逻辑放进 `run` 才能让所有 `defer` 逆序执行。`xslog` 刻意不提供 `Fatal`，就是逼所有失败走 `return err` 这一条路。
2. **基础设施用 `defer` 逆序清理**（原则 #10 零全局状态）。db、redis 是 `run` 的局部变量，不是全局单例，不是 `App` 字段——谁创建谁 `defer` 关，顺序天然正确（redis 先关、db 后关，与创建相反）。
3. **`signal.NotifyContext` 把信号变成可传播的 ctx**（原则 #9 优雅退出）。信号变成 `ctx.Done()`，传进 `srv.Run(ctx)`——「收到信号」与「组件出错」两个退出源统一成同一个 ctx 取消；`defer stop()` 在退出后恢复默认信号行为（再按一次 Ctrl-C 能强杀卡住的进程）。
4. **优雅关闭收敛在 `srv.Run(ctx)` 里**。组件的带超时 `Shutdown`（`GracefulTimeout` 秒内排空在途请求）由 Server 内部的 errgroup 编排，`main` 不碰细节。**这正是老项目 B 缺失的东西**（见下方 callout）。

> **反面 callout（老项目 B god-App）**：老项目用一个 14 字段的 `App` god-struct 把 config/db/redis/cron/WS/handlers **全拍平塞进去**，`NewApp()` 里 ~15 步编号 init。问题不在「编号 init」（这点不坏），而在**边界糊了**：被动资源（db/redis）、主动 goroutine（cron/WS）、请求级装配（handlers）挤在一个结构体里，谁 `defer` 关、谁 `Stop`、什么顺序全靠人肉记忆——最典型的塌方是 HTTP **根本没有优雅关闭**（`ListenAndServe` 起在 `go` 里，`os.Exit` 直接吞掉在途请求）。本项目的 `run() + 编排器` 两层划分就是为了根治这个：被动资源留 `run()`（`defer` 最省心），主动组件交编排器（统一优雅关闭）。

`server.go` 里，`Server` 只持可启停组件 + 它们需要的依赖，并留好了 cron 的钩子（`// Step 05` 注释标出了未来长出来的位置）。编排全部收敛在 `Run(ctx)`：

```go
type Server struct {
    httpSrv         *http.Server
    db              *gorm.DB
    rdb             *redis.Client
    log             *slog.Logger
    gracefulTimeout time.Duration
    // Step 05 起追加：scheduler *scheduler.Scheduler
}

// Run 启动全部长驻组件并阻塞，直到 ctx 被取消（收到信号）或任一组件返回非 nil error。
func (s *Server) Run(ctx context.Context) error {
    g, ctx := errgroup.WithContext(ctx)   // 任一 g.Go 返回 err → cancel 这个 ctx

    // 组件 1：HTTP server —— 阻塞跑 ListenAndServe。
    g.Go(func() error {
        // 已知坑：正常 Shutdown 会让 ListenAndServe 返回 http.ErrServerClosed，
        // 必须过滤成 nil，否则会被 errgroup 当成「出错」而误触发全体退出。
        if err := s.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            return err
        }
        return nil
    })

    // 组件 1 的优雅关闭：等 ctx 取消后带超时 Shutdown，排空在途请求。
    g.Go(func() error {
        <-ctx.Done()
        shutdownCtx, cancel := context.WithTimeout(context.Background(), s.gracefulTimeout)
        defer cancel()
        return s.httpSrv.Shutdown(shutdownCtx)
    })

    // Step 05 起追加：组件 2 scheduler —— 再多注册一个 g.Go（start → <-ctx.Done() → Shutdown）

    return g.Wait()   // 收集第一个非 nil error（信号触发的正常退出返回 nil）
}
```

**关键设计点**：`main.go` 只认 `Server.Run(ctx)` 一个动词，编排细节封装在 `Server` 内部。加第 N 个长驻组件（Step 05 的 scheduler）只需在 `Run` 里**多注册一个 `g.Go`**，`main.go` 一行不改——这正是 day-1 就用 errgroup 换来的线性扩展（提前采用的取舍见第二节末小节）。

## 二、协调机制对比：按组件数分档选

「多个组件跑在一个进程、怎么统一启停」在 Go 生态有几种主流答案。逐维度对比：

| 方案 | 启动错误传播 | 关闭顺序控制 | HTTP 优雅关闭 | 心智负担 | 额外依赖 |
|------|------------|------------|--------------|---------|---------|
| 老项目 B：`signal.Notify` 裸 channel + `App.Close()` | ❌ HTTP 起在 `go` 里，`os.Exit` 直接吞错 | ⚠️ 手工 `Close()`，但 HTTP 不在其中 | ❌ **没有** | 低 | 无 |
| 单组件档（教学参照）：`signal.NotifyContext` + `select{srvErr/ctx.Done}` | ✅ `srvErr` channel 回传 | ✅ `Stop()` 内显式逆序 | ✅ `Shutdown(ctx)` 带超时 | 低 | 无（纯标准库） |
| **`errgroup.WithContext`（头号推荐，≥2 组件，`backend/` 现状）** | ✅ 任一 goroutine 返回 err 即 cancel 共享 ctx | ✅ 各组件监听 ctx.Done 各自带超时收尾 | ✅ `Shutdown(ctx)` | 中（一个概念） | `golang.org/x/sync` |
| `oklog/run` run.Group（actor 模式） | ✅ 任一 actor 退出即触发全体 | ✅ 每 actor 配对 `execute/interrupt` | ✅ 可做 | 中高（execute/interrupt 词汇表） | `github.com/oklog/run` |

关键判断，按组件数分档：

- **只有 HTTP 一个长驻组件**：`select{srvErr / ctx.Done}` + `Stop()` 已经够，零依赖最直白，没必要引 orchestrator。多数 Gin boilerplate/教程止步于此——**若确定长期只有 HTTP，这就是终点**。
- **HTTP + scheduler/worker（本项目实际方向）**：一旦有第二个对等长驻组件，「任一失败拉全体优雅退出」就成了刚需——`errgroup.WithContext` 是最简的成熟答案。它只多引一个 `golang.org/x/sync`（准标准库），心智模型就一句话：**每个组件一个 `g.Go(func)`，任一返回非 nil error → 共享 ctx 被 cancel → 其余组件从 `ctx.Done()` 收到信号、各自带超时收尾；`g.Wait()` 收集第一个 error**。不需要学 `execute`/`interrupt` 成对绑定那套词汇。

`signal.NotifyContext`（Go 1.16+）在所有档里都保留：它把信号变成可传播、可被下游 ctx 感知的 `context`，配合 `defer stop()` 还能在退出后恢复默认信号行为（再按一次 Ctrl-C 能强杀卡住的进程）。errgroup 的用法就是把 `signal.NotifyContext` 得到的 ctx 传进 `errgroup.WithContext`，信号与「组件出错」这两个退出源就统一成同一个 ctx 取消。

> 关于 README 原则 #1/#9/#10：它们（组合根在 main、优雅退出、零全局状态）本身与上述行业实践一致——errgroup 编排正是「组合根在 main 里编排长驻组件」的最简成熟实现。本文把它们当**与行业一致的既有约定**引用，而非「因为规定所以照做」。

### 单组件下 errgroup 并不更简单——为什么仍提前采用

先承认一件事，免得读代码时犯嘀咕：**当前只有 HTTP 一个长驻组件，errgroup 版并不比裸 `go func` 简单，反而更重。** 裸写法就三行——

```go
go func() { _ = srv.ListenAndServe() }()   // 后台跑
<-ctx.Done()                                // 等信号
_ = srv.Shutdown(shutdownCtx)               // 优雅关闭
```

errgroup 版要把它拆成**两个 `g.Go`**（一个 `ListenAndServe`、一个 `<-ctx.Done()` 后 `Shutdown`），还多背一个「`ErrServerClosed` 必须过滤成 nil，否则正常关闭被误判为出错」的坑。对单组件而言，这是**净增负担**——多几行、多一个概念、多一个坑。直觉上「不如原来的 `go func` 简单」，这个判断是对的。

那为什么代码 day-1 就用了 errgroup？因为**它省的不是当前这一个组件，是第二个**。其他项目也都这么分档：

| 组件数 | 主流做法 | 代表 |
|--------|---------|------|
| 1（仅 HTTP） | `go func` + `<-ctx.Done()` + `Shutdown` | 多数 Gin boilerplate、教程 |
| 2–N（HTTP + gRPC/worker/scheduler） | **`errgroup.WithContext`** | Go 服务类项目的标准答案 |
| actor 众多、需精确关闭排序 | `oklog/run` | Grafana、Prometheus、Thanos |

本项目 scheduler（Step 05）是**已知的第二个长驻组件**（`server.go` 已留 `// Step 05 起追加：组件 2 scheduler` 钩子）。已知马上要迈进「2–N」档，就没必要先写裸 `go func`、到 Step 05 再整段改写成 errgroup——**day-1 直接用 errgroup，Step 05 只多注册一个 `g.Go`，`main.go` 零改动**。这与 [02](./02-依赖装配与分层对象化-手工注入与struct封装.md) 里「`server.New` 用 `Options` struct 而非位置参数」是同一取舍逻辑：为**已知会增长**的结构预付一点当下成本，换掉将来的返工。

代价（已知并接受）：当前单组件版比裸 `go func` 多几行、多一个 `ErrServerClosed` 过滤坑。收益：Step 05 加 scheduler 时零改写、线性扩展。若哪天确定「这进程永远只有 HTTP」，那 errgroup 就是过度设计，退回裸 `go func` 更诚实——判据始终是**第二个长驻组件会不会出现**，本项目答案是「Step 05 就来」。

## 三、组件归属：run() + Server 两层

本项目的两层划分（对照上一节 callout 里老项目 god-App 的边界糊）：

| 层 | 持有什么 | 生命周期动作 | 谁写 |
|----|---------|------------|------|
| `run()`（组合根） | config/log/db/redis 等被动资源 | 顺序创建 + `defer` 逆序 Close | `cmd/server/main.go` |
| 编排器（`errgroup`） | http/cron/WS 等主动组件 + 其依赖 | 各 `g.Go()` 拉起 / ctx 取消触发各自优雅关闭 | `internal/server/server.go` |

**边界规则一句话**：需要 `Close()`/连接池的被动资源留在 `run()`（defer 最适合）；有自己 goroutine、需要「优雅」关闭的主动组件交编排器。db 不进编排器——它没有「优雅关闭」语义，`Close()` 一把梭即可，放 `run()` 的 defer 里最省心。

### 为什么结论速览是三类，这里只有两层

结论速览表分**三类**，本节却只有**两层**——不是漏了，是第三类**按设计就不归组合根管**。

| 类别 | 谁持有 | 在 main.go 里吗 |
|------|--------|----------------|
| ① 基础设施（被动资源） | `run()` 局部变量 | ✅ 在，`defer` 清理 |
| ② 可启停组件（主动 goroutine） | 编排器（`Server.Run` 的 errgroup） | ✅ 在，但被 `srv.Run(ctx)` 一层收进 `Server`，main 只认这一个动词 |
| ③ 请求级 / 任务级（handler / job） | **无人持有**，随一次调用产生消亡 | ❌ **不在**，也不该在 |

关键在「谁持有」：handler 来一个请求生一个、处理完即消亡；job 到点触发一次、跑完即消亡。它们不是 `run()` 的局部变量、也不注册进 errgroup，生命周期完全跟着「一次调用」走，靠 **ctx 取消传播**收尾（errgroup 的可取消 ctx 一路穿透到 handler/job，见第五节反面备忘）。所以它天然不出现在组合根里——代码在 `internal/handler/{domain}/` 与各 job 函数里，main 碰不到。

**反过来说**：若哪天把 handler 也塞进某个长驻结构体让 main「持有」，那正是本文开头 callout 批评老项目 god-App 的病根——把请求级装配和被动资源、主动 goroutine 挤进同一个 struct，边界就糊了。第三类**不出现在组合根，本身就是边界清晰的体现**，注释里只写两层（① defer + ② 交编排器）是对的，不是缺了一层。

## 四、成熟编排方案：select / errgroup / run.Group

「一个 `main` 编排 N 个长驻组件」是个有成熟解的问题，三种主流方案按组件数与场景选：

| 方案 | 错误传播 | 关闭机制 | 社区采用 | 依赖 | 适用 |
|------|---------|--------------|---------|------|------|
| 裸 `select{srvErr/ctx.Done}` + `Stop()` | 手写 channel 回传 | 手写逐个 `Stop()` | 教程常见 | 无 | **单**长驻组件（仅 HTTP），教学参照 |
| **`errgroup.WithContext`（头号推荐）** | ✅ 任一返回 err 即 cancel 共享 ctx | ✅ 各组件监听 `ctx.Done()` 自行带超时收尾 | ✅ 广泛（准标准库） | `golang.org/x/sync` | **HTTP + scheduler/worker（本项目 2–N 组件）** |
| `oklog/run` run.Group | ✅ 任一 actor 返回即触发全体 interrupt | ✅ 每 actor 配对 `execute/interrupt` | ✅ Grafana/Prometheus/Thanos/Cortex | `github.com/oklog/run` | actor 众多、需精确 interrupt 排序 |

**推荐：组件 ≥2 用 `errgroup.WithContext`。** 本项目 HTTP + scheduler（cron/worker，见 [cron/02](../cron/02-单机执行模型-从ticker到持久队列的四档.md)）恰好落在这。errgroup 的骨架（含 HTTP 优雅退出的已知小坑）：

```go
func run() error {
    // ...基础设施创建 + defer 逆序 Close（同第一节）
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    g, ctx := errgroup.WithContext(ctx)   // 任一 g.Go 返回 err → cancel 这个 ctx

    // 组件 1：HTTP server
    g.Go(func() error {
        // ⚠️ 已知坑：正常 Shutdown 会让 ListenAndServe 返回 http.ErrServerClosed，
        //    必须过滤成 nil，否则会被 errgroup 当成「出错」误触发全体退出。
        if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            return err
        }
        return nil
    })
    g.Go(func() error {                    // HTTP 的优雅关闭：等 ctx 取消再 Shutdown
        <-ctx.Done()
        shutdownCtx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
        defer cancel()
        return httpSrv.Shutdown(shutdownCtx)
    })

    // 组件 2：scheduler（cron/worker，见 03）
    g.Go(func() error {
        scheduler.Start()
        <-ctx.Done()
        scheduler.Shutdown()               // drain 在途任务
        return nil
    })

    return g.Wait()                        // 收集第一个非 nil error（信号退出则为 nil）
}
```

errgroup 相比裸 `select` 的关键收益：**第二个组件出现后，「任一失败拉全体优雅退出」不用再手写「N 个 channel + 大 select + 逐个 Stop」**——`g.Go` 注册、`ctx` 广播、`g.Wait` 收敛，三件事各归其位。相比 run.Group 的收益：**少一层 `execute/interrupt` 词汇表**，HTTP 的优雅关闭直接写成「等 `ctx.Done()` 再 `Shutdown`」，是普通 Go 代码，不用理解 actor 模型。代价是那个 `ErrServerClosed` 过滤要记得写（run.Group 里 interrupt 与 execute 分离，天然没这个坑）——一行的事，注释标清即可。

**何时 errgroup 也不必要**：整个进程只有 HTTP 一个长驻组件、且短期不会加第二个——裸 `select` 版足够，别为「将来可能」预付依赖。判据就是组件数：第二个长驻组件出现之日，就是换 errgroup 之时。

### 升级路径：什么时候才轮到 run.Group

run.Group 不是「更高级所以更好」，它解决的是 errgroup **不擅长**的场景：**actor 数量多（5+）、且需要精确控制「谁的 interrupt 先触发、按什么顺序」**。Grafana/Prometheus/Thanos/Cortex 的 `func main` 里编排一堆长驻组件（HTTP + gRPC + 多个 reloader + 多个 exporter…），run.Group 的 `execute()`/`interrupt(error)` **成对绑定**让「如何优雅打断每个组件」显式且集中，比 errgroup 里散落在各 goroutine 的 `<-ctx.Done()` 更好审计。

机制差一句话：**errgroup 把「优雅关闭」写成普通代码**（另起 `g.Go` 等 `<-ctx.Done()` 再 `Shutdown`，代价是记住 `ErrServerClosed` 要过滤成 nil）；**run.Group 把它写成配对的 `interrupt(error)`**（execute/interrupt 成对绑定，天然没有 `ErrServerClosed` 坑，代价是先理解这套 actor 词汇表）。组件少时前者更划算，组件多、要审计「每个组件怎么被打断」时后者更集中。

判据：**当你发现 errgroup 里的 `<-ctx.Done()` 收尾逻辑开始分散、且需要保证关闭顺序（「A 必须在 B 之前关」）时，才升级到 run.Group。** 本项目 HTTP + scheduler 两个组件、关闭无严格排序要求——**「A 必须在 B 之前关」这个需求现在用不到**，远未到升级点。

### 被否的重方案：go-kratos 式两层 `internal/app/` 编排

kratos 用 `internal/app` 做应用装配 + `kratos.App` 做生命周期，是为「多传输层（HTTP+gRPC）+ 服务注册发现 + 配置中心」的微服务全家桶设计的。本项目是单体 admin BFF，`run() + Server` 一层足矣。引入 kratos 的 `App`/`Server`/`Transport` 抽象，是在没有对应问题时先付抽象税。

**单副本 vs 多副本**：编排机制（select / errgroup / run.Group）都是**进程内**的，与部署几个副本无关——它只管「本进程内这几个组件怎么一起启停」。多副本下「同一任务只让一个副本跑」是**调度层**的事（拆调度器 / gocron 的 Locker / River 的 leader 选举，见 [cron/05](../cron/05-从单机到多节点-重复跑与拆调度器.md)），不在生命周期层。所以本文结论对单副本、多副本都成立：单副本时 scheduler 直接跑，多副本时 scheduler 照跑、由调度库内建的协调决定哪个副本实际执行。

## 五、成本与 tradeoff（已知并接受）

- **errgroup 的 HTTP 优雅退出要拆成两个 `g.Go`**：一个跑 `ListenAndServe`、一个等 `ctx.Done()` 后 `Shutdown`。比裸 `select` 多几行，但换来「加第 N 个组件只是多一个 `g.Go`」的线性扩展。`ErrServerClosed` 必须过滤成 nil（否则正常关闭被误判为出错）——这是 errgroup 编排 HTTP 的唯一已知坑，注释标清。
- **裸 `select` 版只处理「首个」事件**（单组件档的已知限制）：收到信号后若关闭期间又崩了另一个组件，不会再被 `select` 捕获——这正是多组件该上 errgroup 的理由之一（`g.Wait` 统一收敛所有 goroutine 的退出）。
- **反面备忘（老项目 B 的坑，新项目勿犯）**：老项目里按需拉起的 worker 用 `go func()` + `context.Background()`，`Background()` 永不取消，SIGTERM 时这些 goroutine 完全不受控、直接被进程消灭。**本项目所有后台 goroutine 必须接 `errgroup.WithContext` 传下来的可取消 ctx**，让优雅关闭能穿透到最底层任务。scheduler/worker 的选型与集成见 [cron/02](../cron/02-单机执行模型-从ticker到持久队列的四档.md)。
