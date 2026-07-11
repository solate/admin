# 后台任务执行模型选型：进程内调度器、durable queue 与生命周期集成

> 2026-07。本文回答用户的原始痛点——**「gin 和 cron 同步任务都在一个 main 里，一个 main 管多个部分；现在初始化 handler、cron 调用太复杂，每次要初始化一堆东西」**——但站在成熟方案的视角重新展开：新项目的后台任务（定时清理/同步、崩溃重试、事件触发的重活）到底该用什么执行模型？进程内 cron？还是持久化 job queue？怎么随进程优雅启停、怎么在多副本下不重复跑？
>
> 结论先行（成熟方案优先，按需分档，不押单一栈）：
> - **纯定时、可丢的任务**（清理、拉取声明式配置）→ 进程内调度器，选 **[go-co-op/gocron v2](https://github.com/go-co-op/gocron)**（内建 distributed locker/elector、singleton、并发上限），**不要**再用 `robfig/cron` + 手搓 Redis 锁。
> - **要持久化 / 重试 / 唯一 / 崩溃恢复的任务** → durable job queue：Postgres 已在栈内 → **[River](https://riverqueue.com/)**（事务性入队、leader 插周期任务、unique jobs、自动重试）；已重度用 Redis 且要优先级/限流/Web 面板 → **[asynq](https://github.com/hibiken/asynq)**。
> - **执行器（scheduler / worker）作为一个 actor 交给 `run.Group` 编排**（见 [01](./01-应用启动与组件生命周期-main组合根与gin-cron编排.md)），随进程优雅启停。
>
> 关联：生命周期编排见 [01](./01-应用启动与组件生命周期-main组合根与gin-cron编排.md)，依赖装配见 [02](./02-依赖装配与组件注册-手工注入vs-wire-fx.md)；反面标本见老项目 B 的手搓并发控制；落地约束参照 [logging-style.md](../../../../.claude/rules/logging-style.md)、[service-patterns.md](../../../../.claude/rules/service-patterns.md)。
>
> 日常写代码不必读；选后台任务栈、或纠结「这个任务要不要持久化」时看这里。

## 结论速览

后台任务不是「一种东西」，按**可靠性需求**分三档，成熟方案各有归属：

| 档位 | 任务特征 | 成熟方案 | 多副本协调 |
|------|---------|---------|-----------|
| ① 纯定时、丢了无所谓 | 周期清理、拉声明式配置、健康自检 | **gocron v2**（进程内调度器） | 库内建 locker/elector/singleton |
| ② 要持久化/重试/唯一 | 事件触发的重活、发信、数据处理、必须完成的作业 | **River**（PG 栈内）/ asynq（Redis） | DB/Redis 抢占，天然安全 |
| ③ 崩溃恢复 backfill | 「本该实时做、怕漏了每小时补扫」 | **这就是 ② 的信号** → 直接上 River | 同 ② |

一句话：**「定时」和「可靠执行」是两件事**。进程内调度器只解决「按点触发」；一旦任务「必须完成、失败要重试、重复不能跑两次」，那是 job queue 的活，手搓 cron + 锁 + backfill 等于在重新发明一个残缺的持久队列。

被本轮明确纠偏的老做法：`robfig/cron` + 手搓 Redis `SetNX` 锁 + `sync.Map` 去重 + hourly backfill 兜底——**每一件成熟库都已内建**，见第三节。

## 一、先分清两个正交的问题

「后台任务怎么做」其实是两个独立问题，老项目把它们搅在了一起：

1. **触发（when）**：什么时候跑？固定间隔 / cron 表达式（定时），还是某个业务事件发生时（事件驱动）。
2. **执行保证（how reliably）**：跑一次就行、丢了无所谓？还是必须成功、失败要重试、进程崩了重启要接着做、同一任务不能并发跑两份？

进程内调度器（gocron / robfig-cron）只回答第 1 个问题的「定时」半边。第 2 个问题——持久化、重试、唯一性、崩溃恢复——**它一概不管**。老项目的痛，本质是用只回答问题 1 的工具，去硬扛问题 2 的需求（于是手搓锁、手搓去重、手搓 backfill）。成熟方案的分档就是按问题 2 的强度选工具。

## 二、成熟方案三档：gocron v2 / River / asynq 横向对比

### 2.1 档位 ① 进程内调度器：gocron v2

**场景**：纯定时、可丢、无需崩溃恢复。典型例子——每 30 分钟清理过期临时文件、每天凌晨拉一次声明式配置、每 5 分钟健康自检。

**[go-co-op/gocron v2](https://github.com/go-co-op/gocron)**：robfig/cron 的活跃维护分支，**2026 年的正当默认进程内调度器**。相比 robfig/cron 多了生产必需的四样：

| 能力 | gocron v2 | robfig/cron | 老项目 B 手搓 |
|------|-----------|-------------|-------------|
| 多副本只跑一个 | ✅ 内建 Locker/Elector（Redis/etcd） | ❌ 无，全副本都跑 | ⚠️ 手搓 Redis `SetNX` |
| 任务级去重 | ✅ Singleton mode | ❌ 无 | ⚠️ 手搓 `sync.Map` |
| 并发上限 | ✅ `LimitConcurrentJobs(N)` | ❌ 无 | ⚠️ 手搓 `chan struct{}` 信号量 |
| 优雅关闭 | ✅ `Shutdown()` drain | ⚠️ `Stop()` 但不 drain | ✅ `Stop()` + `<-ctx.Done()` |

**用法骨架**（Redis locker，多副本下任务只跑一个）：

```go
import (
    "github.com/go-co-op/gocron/v2"
    gocronredis "github.com/go-co-op/gocron-redis-lock"
)

locker, _ := gocronredis.NewRedisLocker(redisClient, gocronredis.WithTries(1))
s, _ := gocron.NewScheduler(
    gocron.WithDistributedLocker(locker),  // 多副本只选举一个跑
)
s.NewJob(gocron.DurationJob(30*time.Minute), gocron.NewTask(cleanupTempFiles))
s.Start()
defer s.Shutdown()
```

**何时够**：任务失败不影响正确性（下次重跑就行），进程崩了任务丢了无所谓。**何时不够**：一旦出现「本该跑但漏了怎么办」「失败了要重试」「同一事件不能重复处理」——这些是持久队列的地盘，继续看 ②。

---

### 2.2 档位 ② 持久化队列（durable job queue）

**场景**：任务必须完成、失败自动重试、进程崩了重启要接着做、同一作业不能跑两次（唯一性）。典型——事件触发的发信/媒体处理、订单超时取消、数据管道 ETL。

两个成熟选项，按已有基础设施选：

#### River（[riverqueue.com](https://riverqueue.com/)）—— Postgres 栈内首选

- **[事务性入队](https://riverqueue.com/docs/transactional-enqueueing)**：`tx.Exec(...); river.InsertTx(tx, job)` 原子，业务失败则任务不入队，避免「业务回滚、任务已入队」的不一致。
- **[周期任务](https://riverqueue.com/docs/periodic-jobs)**：leader 节点按 cron 表达式插入，多副本安全（选主）。
- **[unique jobs](https://riverqueue.com/docs/unique-jobs)**：按 args/kind/period/queue/state 去重，防「同一订单取消任务跑两次」。
- **自动重试 + 指数退避**、优先级队列、并发限制（per-worker-type）、dead letter。
- **零新基建**：Postgres 本来就在栈内，River 只是一批表 + 一套 Go client。多副本下 leader 选举靠 PG advisory lock，天然安全。

**判据**：Postgres 已在栈内 + 任务需持久化/重试/唯一 → River 是自然选择。

#### asynq（[github.com/hibiken/asynq](https://github.com/hibiken/asynq)）—— Redis 栈优先

- Redis-backed、at-least-once、优先级队列（P0–P9）、周期任务、unique tasks（TTL 去重）、rate limiting、retries + 死信。
- **[asynqmon](https://github.com/hibiken/asynq#web-ui) Web 面板**：实时查看队列状态、重试失败任务、手工触发，运维友好。
- 多副本安全：Redis 单点抢占（Lua 脚本），天然不重复消费。

**判据**：已重度依赖 Redis + 要优先级/限流/面板 + 不想给 PG 加负载 → asynq。

#### 对比（River vs asynq）

| 维度 | River | asynq |
|------|-------|-------|
| 存储 | Postgres（栈内，事务性入队） | Redis（需独立维护） |
| 周期任务 | ✅ leader 插入 | ✅ 调度器插入 |
| 唯一性 | ✅ args/kind/state 组合 | ✅ TTL 去重 |
| 优先级 | ❌ Pro 版才有 | ✅ P0–P9 |
| 限流 | ❌ | ✅ rate limiting |
| Web UI | ❌ | ✅ asynqmon |
| 并发控制 | ✅ worker-type 级 | ✅ queue 级 |
| 社区 | 新（2023+）但快速增长 | 成熟（2020+） |

River 优势：Postgres 栈内 + 事务性入队避免一致性坑；asynq 优势：成熟度 + 优先级/限流 + Web 面板。**对本项目**：PG 已在栈内，除非明确要 asynq 的优先级/限流/面板，否则 River 是更自然的起点（零新基建）。

---

### 2.3 档位 ③「崩溃恢复 backfill cron」= 档位 ② 的信号

老项目 B 有个模式：每小时跑一个 cron，扫「状态=处理中但超时未完成」的记录，重新触发处理——这叫 backfill / retry worker，名义上是「兜底」。**实质是在用 cron 模拟持久队列的崩溃恢复**：

```go
// 老项目 B 的 hourly backfill cron（反面）
func RetryStuckJobs() {
    jobs := repo.FindStuckJobs(ctx)  // 状态异常的记录
    for _, j := range jobs {
        StartProc(ctx, j.ID)  // 重新触发
    }
}
```

三个问题：
1. **轮询低效**：hourly 扫全表，99% 时候查不出东西，空转。
2. **重试策略硬编码**：何时算「stuck」、重试几次、退避间隔全在业务代码里，改一次全改。
3. **持久队列的残缺实现**：这套逻辑本质是在手搓「任务入队→失败重试→超时死信」，正是 River/asynq 的一等公民能力。

**决策规则**：一旦你写出了「每小时扫未完成记录重试」的 cron，就是档位 ② 的需求——直接上 River 或 asynq，别手搓。队列的 `RetryPolicy{MaxRetries, Backoff}` 替代整个 backfill cron；任务失败自动重试、超时自动死信，不需要轮询。

## 三、老项目 B 手搓并发控制的代价（反面标本）

老项目 B 的某几个重型后台模块（资源并发受限的处理任务）用了包级全局 + 手搓并发控制。现在看，**它手搓的每一件成熟库都已内建**：

| 老项目手搓的 | 成熟库哪层内建 |
|------------|--------------|
| Redis `SetNX` 锁（多副本只跑一个） | gocron v2 Locker/Elector、River/asynq 天然多副本安全 |
| `sync.Map` 去重（同资源不并发） | gocron v2 Singleton、River/asynq unique jobs |
| `chan struct{}` 信号量（进程内并发上限） | gocron v2 `LimitConcurrentJobs`、River/asynq worker concurrency |
| Hourly backfill cron（崩溃恢复） | River/asynq 自动重试 + 死信，不需要轮询 |

三个代价：
1. **包级可变全局**：`var procSemaphore chan struct{}`、`var procInFlight sync.Map` 违背零全局状态，挡并行测试。
2. **依赖来路发散**：`InitProc(db, redis, cfg, ...)` 参数随任务增长膨胀（正是用户说的「每次要初始化一堆东西」在 service 层的翻版）。
3. **重复造轮子且造得残缺**：backfill 只兜底「超时未完成」，无法兜底「进程崩溃时正在跑的任务」，也没有指数退避、死信、可观测——这些 River/asynq 免费送。

**为什么当时这么做**？因为选了 `robfig/cron`（只解决「按点触发」），但任务需求其实落在档位 ②（要持久化/重试/唯一）。于是在 cron 上层手搓一套简陋的执行保证——这就是用错工具的典型后果。新项目按档位选工具就不会走到这一步。

## 四、本项目推荐（基于已确认边界）

**已知约束**：部署形态暂未定，单副本/多副本都要兼顾；Postgres 已在栈内；Redis 也在栈内但更希望作缓存而非基建核心。

分档推荐（**不押单一栈**，按任务特征匹配）：

| 任务类型 | 推荐栈 | 单副本路径 | 多副本路径 |
|---------|--------|-----------|-----------|
| 纯定时、可丢（清理/健康检查） | **gocron v2** | 直接跑，无需 locker | 加 Redis Locker/Elector |
| 持久化/重试/唯一/崩溃恢复 | **River**（首选）或 asynq | 直接跑 | 天然多副本安全（PG 选主） |
| 要优先级/限流/Web 面板 | **asynq** | 直接跑 | 天然多副本安全（Redis 抢占） |

**默认起手**：**gocron v2**（覆盖大部分定时任务，无新基建）。一旦任务出现以下信号，升级到 River：
- 「这个任务失败了必须重试，不能丢」
- 「同一事件（订单 ID / 用户 ID）只能处理一次」
- 「进程崩了重启要接着做，不能从头来」
- 「写了个 hourly backfill cron 来补漏」 ← **这就是信号**

asynq 作为「重度 Redis 用户 + 要 Web 面板 + 要优先级队列」的备选。PG 栈内、任务量不大 → River 是自然选择。

**装配形态**（承接 [02](./02-依赖装配与组件注册-手工注入vs-wire-fx.md) 按域装配）：

```go
// internal/server 装配期（示意）
func buildModules(deps ModuleDeps) (*moduleHandlers, *gocron.Scheduler, error) {
    // gocron 调度器：装配期建一次
    locker, _ := gocronredis.NewRedisLocker(deps.RDB)
    scheduler, _ := gocron.NewScheduler(gocron.WithDistributedLocker(locker))
    
    // 注册任务（复用装配好的 service，不重建）
    cleanupSvc := cleanup.NewService(deps)
    scheduler.NewJob(gocron.DurationJob(30*time.Minute), gocron.NewTask(cleanupSvc.Run))
    
    // 若有 River worker，同样装配一次，见下节
    // riverWorkers := river.NewWorkers(); riverWorkers.Add(...)
    
    handlers := &moduleHandlers{ /* ... */ }
    return handlers, scheduler, nil
}
```

任务依赖**复用装配好的 service**（不在 jobs 里重建），与请求 handler 共享同一套 service 实例——呼应 [02](./02-依赖装配与组件注册-手工注入vs-wire-fx.md) 消除双真相源。

## 五、生命周期集成：scheduler/worker 作为一个 actor

承接 [01](./01-应用启动与组件生命周期-main组合根与gin-cron编排.md) 的 `run.Group` 编排，scheduler 或 worker 是一个独立 actor（HTTP 是另一个）：

```go
// cmd/server/main.go 伪代码（承接 01 的 run.Group 修正）
func run() error {
    // ...装配（buildModules 得到 scheduler / riverWorkers）
    
    var g run.Group
    
    // Actor 1: HTTP server
    g.Add(func() error {
        return httpSrv.ListenAndServe()
    }, func(error) {
        httpSrv.Shutdown(context.Background())
    })
    
    // Actor 2: gocron scheduler
    g.Add(func() error {
        scheduler.Start()
        <-ctx.Done()  // 等信号
        return nil
    }, func(error) {
        scheduler.Shutdown()  // drain 在途任务
    })
    
    // Actor 3（可选）: River workers
    // g.Add(func() error { return riverClient.Start(ctx) }, ...)
    
    // 信号 actor
    g.Add(run.SignalHandler(ctx, syscall.SIGINT, syscall.SIGTERM))
    
    return g.Run()  // 任一退出→全体 interrupt
}
```

**为什么要 run.Group**（承接 doc 01）：HTTP + scheduler 两个长驻组件，任一失败要拉全体优雅退出。run.Group 的 actor 模式正是这个——比手写 `select{srvErr/ctx.Done}` + 逐个 `Stop()` 更标准。gocron/River 的 `Shutdown()` / `Stop(ctx)` 都有 drain 语义，正好对应 run.Group 的 interrupt 函数。

**单副本 vs 多副本**：run.Group 是进程内编排，与副本数无关。多副本协调在调度器层（gocron locker / River leader 选举 / asynq Redis 抢占），不在生命周期层。

## 六、成本与 tradeoff（已知并接受）

- **gocron v2 相比 robfig/cron 多一层依赖（Locker）**：多副本下需 Redis/etcd locker；单副本下可不加。接受——换来的是不手搓 `SetNX`、不手搓 `sync.Map`、不手搓信号量，边际成本低。
- **River/asynq 引入「队列」心智模型**：与 cron「到点就跑」不同，队列是「入队→worker 拉取→重试」，开发者要理解 at-least-once 语义。接受——这是持久化任务的固有复杂度，手搓 backfill cron 也绕不开，成熟库只是把它显式化。
- **River 给 Postgres 加负载**：周期任务插入、worker 轮询都是 SQL。本项目任务量预期不大（清理/同步/发信级别），可接受；若任务吞吐到 1000+ jobs/s，考虑 asynq（Redis 扛得更高）。
- **多一个执行模型要学**：团队要懂两套（gocron 定时 + River 队列）。接受——分档选工具的代价；好处是不再手搓轮子、不再写 backfill cron。
- **「万一只有一两个定时任务」**：gocron v2 仍是最轻的（一个 `NewScheduler` + `NewJob`，10 行代码）；River 只在真有持久化需求时才引入，不预设。判据明确：出现 backfill cron = 上 River。

**不接受的代价（明确规避）**：手搓 Redis 锁 + `sync.Map` + 信号量 + backfill cron = 重新发明持久队列的残缺版（老项目的坑），既难测又难维护，成熟库已解决。
