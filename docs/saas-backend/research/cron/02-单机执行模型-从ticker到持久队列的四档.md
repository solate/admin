# 单机执行模型：从 ticker 到持久队列的四档

> 2026-07。本文回答「后台任务在单机/单副本视角下该怎么执行」——**周期清理、定时拉配置、按需触发的处理、失败重试、崩溃恢复……不同任务特征该选什么执行方式，从零依赖到成熟队列怎么升级？**
>
> 结论先行：**沿「可靠性轴」按「可丢 vs 必达」分四档**——① **`time.Ticker` in errgroup goroutine**（stdlib 地板，零依赖，覆盖周期 + 可丢）→ ② **cron 表达式调度器**（要 `0 2 * * *` 这类复杂时间；库层选 robfig 还是 gocron 见 [03](./03-cron库选型-robfig与gocron对比.md)）→ ③ **[River](https://riverqueue.com/)**（要持久化/重试/唯一/崩溃恢复，PG 栈内、事务性入队）→ ④ [asynq](https://github.com/hibiken/asynq)（Redis 侧备选，优先级队列/限流/面板）。**「定时」与「可靠执行」是两件事**——前者是调度器的活，后者需要持久队列。**「崩溃恢复 backfill cron」= 穷人版持久队列的信号**，一旦出现就直接上 River。
>
> 关联：本文是「可靠性轴」的完整展开（全景与两根轴见 [01](./01-任务调度全景-两根正交轴与知识地图.md)）；cron 库对比见 [03](./03-cron库选型-robfig与gocron对比.md)；**多副本下「只跑一个」的协调**是另一件事，见 [05](./05-从单机到多节点-重复跑与拆调度器.md)；生命周期编排见 [../bootstrap/01](../bootstrap/01-应用启动与组件生命周期-main组合根与gin-cron编排.md)（scheduler 作 errgroup 的一个 goroutine）；落地约束参照 [service-patterns.md](../../../../.claude/rules/service-patterns.md)、[repo-transaction-convention.md](../../../../.claude/rules/repo-transaction-convention.md)。
>
> 日常写代码不必读；要加后台任务、纠结「是不是该上调度库 / 要不要持久队列」时看这里。

## 结论速览

「后台任务」是个过载词，至少涵盖四种不同形态，对执行方式的要求完全不同：

| 任务形态 | 特征 | 典型例子 | 丢了有没有事 | 对应方案档 |
|---------|------|---------|------------|-----------|
| 周期轻量维护 | 定时、轻、无副作用 | 清理过期 session、拉声明式配置 | 无事，下次再跑 | **① `time.Ticker`**（stdlib，零依赖） |
| 周期重活 | 定时、耗时、要 cron 表达式 | 每晚统计、每小时同步 | 有事，但补跑可接受 | **② cron 调度器**（robfig / gocron） |
| 事件触发必达 | 按需、要重试、要唯一 | 审核通过发邮件、数据处理 | 必须执行完 | **③ River**（PG 持久队列） |
| 崩溃恢复 backfill | 「怕漏了补扫」 | 本该实时做、每小时兜底 | 这就是 ③ 的信号 | **直接上 River**，别手搓 |

**核心区分是「可丢 vs 必达」，而非「定时 vs 按需」**：定时只决定触发方式，是否丢得起决定是进程内调度器还是持久队列。

> **一个提醒**：上表刻意不含「多副本」列。多副本下「同一任务被多个进程各跑一遍」是一个**独立于可靠性轴**的协调问题，不该和「可丢 vs 必达」混在一起谈。它的完整讨论（跑 N 遍的后果、拆调度器 vs 分布式锁）在 [05](./05-从单机到多节点-重复跑与拆调度器.md)。本篇只站单机/单逻辑调度器视角。

四档递进（详见下文各节）：

- **① stdlib ticker**（零依赖地板）：周期 + 可丢 → `time.NewTicker` 塞进 errgroup 的一个 goroutine，~10 行，cron 表达式都不需要。
- **② cron 调度器**（首次依赖）：ticker 的升级触发条件是**需要 cron 表达式**（`0 2 * * *` 这类复杂时间）。库选 robfig 薄封装还是 gocron，见 [03](./03-cron库选型-robfig与gocron对比.md)——本篇不重复库对比。
- **③ River**（持久队列，Postgres 栈内）：五个升级触发条件——任务需要**持久化**（崩溃不丢）、**重试**（失败自动再跑）、**唯一**（同 key 只入队一次）、**事务性入队**（业务 commit 了任务才入队）、**backfill/补扫**（这正是持久队列的本职）。PG 已在栈内、无新基建。
- **④ asynq**（Redis 侧备选）：若已重度用 Redis + 要优先级队列/限流/面板，选它；否则 PG 栈的 River 更简。

## 一、地板档：stdlib `time.Ticker`（周期 + 可丢）

最轻一档不需要任何调度库——stdlib 的 `time.Ticker` 塞进 errgroup 的一个 goroutine 就够（见 [../bootstrap/01](../bootstrap/01-应用启动与组件生命周期-main组合根与gin-cron编排.md)）。

**适用条件（两个同时满足）**：
- **固定周期**（每 N 分钟 / 每 N 小时，不需要 cron 表达式）；
- **丢了无所谓**（进程崩溃期间没跑，下次启动再跑即可）。

（多副本下是否会重复跑，是另一层问题，见 [05](./05-从单机到多节点-重复跑与拆调度器.md)。）

典型例子：每 30 分钟清理过期 session、每 5 分钟拉远程声明式配置更新本地缓存、每小时打一次健康自检。

**骨架（约 10 行）**：

```go
// 在 run() 的 errgroup 里注册一个 ticker goroutine
g.Go(func() error {
    ticker := time.NewTicker(30 * time.Minute)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            cleanupExpiredSessions(ctx, db)  // 业务逻辑，传可取消 ctx
        case <-ctx.Done():
            log.Info("ticker stopped")
            return nil
        }
    }
})
```

**关键点**：
- ticker goroutine **必须监听 `ctx.Done()`**（[../bootstrap/01](../bootstrap/01-应用启动与组件生命周期-main组合根与gin-cron编排.md) 第五节反面备忘），让优雅关闭能打断它。
- 业务函数传可取消 `ctx`，任务跑到一半收到 `ctx.Done()` 能及时退出（别传 `context.Background()`）。
- `defer ticker.Stop()` 释放 ticker 的 channel。

这就是全部——零依赖、零配置、零并发控制框架，够用时**这是最简的对**。

**何时不够**（升级到第二档的触发条件）：
- **需要 cron 表达式**：`0 2 * * *`（每天凌晨 2 点）、`0 0 * * 1`（每周一零点）这类复杂时间规则，ticker 做不到 → 上 cron 调度器（robfig / gocron，见 [03](./03-cron库选型-robfig与gocron对比.md)）。
- （**多副本只让一个跑**不是升级到 gocron 的理由——那是架构问题，优先拆调度器，见 [05](./05-从单机到多节点-重复跑与拆调度器.md)。）

## 二、第二档：cron 表达式调度器（robfig / gocron）

当固定间隔不够、要 `0 2 * * *` 这类复杂时间规则时，引入一个进程内 cron 调度器。

**适用条件**：
- 需要 **cron 表达式**（复杂时间规则）——这是升级自 ticker 的主因；
- 要限制**并发执行任务数**、要**单任务串行**（上轮没跑完不重入）、要任务执行**前后钩子** / 监控。

典型例子：每天凌晨 2 点统计报表、每小时同步外部 API。

**库怎么选**：单副本下 `robfig/cron` 薄封装就够；gocron v2 功能更全（Duration/Daily/Weekly 等调度类型、SingletonMode、LimitConcurrentJobs、钩子）但更重。两者的逐项对比、底层实现、成熟度、以及「gocron 的分布式协调是不是刚需」——全部在 [03](./03-cron库选型-robfig与gocron对比.md)，本篇不重复。

**骨架（以 gocron 单副本为例）**：

```go
s, err := gocron.NewScheduler()
s.NewJob(
    gocron.DurationJob(30*time.Minute),   // 或 gocron.CronJob("0 2 * * *", ...)
    gocron.NewTask(func() {
        cleanupExpiredSessions(ctx, db)
    }),
)
s.Start()  // 后台起调度循环

// 集成进 errgroup（见 ../bootstrap/01）
g.Go(func() error {
    <-ctx.Done()
    s.Shutdown()   // 等在途任务 drain 完
    return nil
})
```

> **本档只讲单副本触发。** gocron 的 `WithDistributedLocker` / `WithDistributedElector`（多副本下只让一个跑）不在这里展开——多副本协调优先用「拆调度器 replicas:1」在架构层解决，Locker/Elector 只是「坚持不拆」时的次选，完整讨论见 [05](./05-从单机到多节点-重复跑与拆调度器.md)。

**何时还不够**（升级到第三档 River 的五个触发条件）：

1. **任务要持久化**（进程崩溃、重启期间不能丢）——进程内调度器纯内存，崩了就丢。
2. **要失败自动重试**（带退避策略）——任务里 `return err` 不会重跑，得业务层自己搓重试循环。
3. **要保证唯一**（同 key 任务只入队一次）——周期任务天然唯一，但按需触发的任务（API 回调拉起）做不到去重。
4. **要事务性入队**（业务逻辑 commit 了任务才真入队）——进程内调度器做不到，入队即跑。
5. **出现「backfill / 补扫」需求**——这正是持久队列的信号，别在调度器外面搭 backfill cron，直接上 River。

## 三、第三档：River（持久队列，Postgres 栈内）

[River](https://riverqueue.com/) 是 Postgres-backed job queue，2024 年开源、Go 原生、生产级稳定。与 cron 调度器的本质区别：**调度器决定何时跑（纯内存），River 是队列（保证必达）**。

**适用条件（满足任一即选 River）**：
- 任务**不能丢**（崩溃、重启后要继续跑）；
- 失败要**自动重试**（带指数退避）；
- 要**唯一**（同 key 只入队一次）；
- 要**事务性入队**（与业务 DB 操作在同一事务）；
- 有**「backfill cron」需求**——每小时全扫一遍「怕漏的」记录再补处理，这就是队列 + 重试的穷人版。

典型例子：审核通过发邮件、媒体处理、数据同步、任何「不能因进程崩溃而漏掉」的异步任务。

**为什么选 River 而非 asynq**：
- **栈内零新基建**：PG 已在栈内，River 复用它，不需要单独运维 Redis（asynq 强依赖 Redis）。
- **事务性入队**：`tx.Exec("UPDATE ..."); riverClient.InsertTx(ctx, tx, job)` 在同一事务，业务回滚则任务不入队——asynq 做不到（Redis 与 PG 是两个存储）。
- **周期任务内建**：River 的 periodic jobs 由 leader 自动插入、去重、多副本安全——不需要再跑一个 gocron。
- **Go 原生**：River 是纯 Go、用 PG 的 `FOR UPDATE SKIP LOCKED` 抢任务，无 Lua 脚本、无 Redis 特有语义，排错简单。

asynq 的优势（**只在以下场景选它**）：已重度用 Redis + 要优先级队列 / 要基于 Redis 的限流 / 要 asynqmon web 面板。否则 PG 栈的 River 更简。

**骨架（按需入队 + worker 消费）**：

```go
// 1. 定义 job 类型
type EmailJobArgs struct {
    UserID string `json:"user_id"`
    To     string `json:"to"`
}
func (EmailJobArgs) Kind() string { return "email" }

// 2. 注册 worker
type EmailWorker struct {
    river.WorkerDefaults[EmailJobArgs]
    mailer *Mailer
}
func (w *EmailWorker) Work(ctx context.Context, job *river.Job[EmailJobArgs]) error {
    return w.mailer.Send(ctx, job.Args.To, "审核通过")
}

// 3. 启动 River client + workers
workers := river.NewWorkers()
river.AddWorker(workers, &EmailWorker{mailer: mailer})
riverClient, _ := river.NewClient(riverpgxv5.New(dbPool), &river.Config{Workers: workers})
riverClient.Start(ctx)  // 后台起 worker 池

// 4. 业务层入队（事务性）
tx.Exec("UPDATE reviews SET status='approved' WHERE id=$1", reviewID)
riverClient.InsertTx(ctx, tx, EmailJobArgs{UserID: userID, To: email}, nil)
tx.Commit()  // 业务 commit，任务才真入队

// 5. 集成进 errgroup（见 ../bootstrap/01）
g.Go(func() error {
    <-ctx.Done()
    riverClient.Stop(ctx)   // drain 在途任务
    return nil
})
```

**周期任务（替掉进程内调度器）**：

```go
riverClient, _ := river.NewClient(db, &river.Config{
    PeriodicJobs: []*river.PeriodicJob{
        river.NewPeriodicJob(
            river.PeriodicInterval(30*time.Minute),  // 或 river.Cron("0 2 * * *")
            func() (river.JobArgs, *river.InsertOpts) {
                return CleanupJobArgs{}, nil
            },
            &river.PeriodicJobOpts{RunOnStart: true},
        ),
    },
})
```

周期任务由 River 的 **leader 副本**（通过 PG advisory lock 选举）自动插入，多副本天然安全、无需手搓锁。这里的「leader 选举」为什么能保证多副本安全，其共识原理见 [06](./06-分布式共识-选主仲裁与不重复的本质.md)。

**唯一 + 重试**：

```go
riverClient.Insert(ctx, ProcessJobArgs{ResourceID: resourceID}, &river.InsertOpts{
    UniqueOpts: river.UniqueOpts{
        ByArgs: true,  // 同 ResourceID 只入队一次（已在队列 / 最近 24h 执行过 → 跳过）
    },
    MaxAttempts: 5,    // 失败最多重试 5 次，指数退避
})
```

**单副本 vs 多副本**：
- **单副本**：River 照跑，worker 池处理任务，周期任务正常插入。
- **多副本**：River 自动选 leader（PG advisory lock），leader 独占插周期任务；worker 池在每个副本都跑、通过 `FOR UPDATE SKIP LOCKED` 抢任务，天然负载均衡、不重复处理。这是「持久队列如何天然多副本安全」的范例——它把协调外置给了 PG 的强一致原语（见 [06](./06-分布式共识-选主仲裁与不重复的本质.md)）。

**成本**：River 在 `pgx` 连接池上跑（复用 gorm 底层的 `*sql.DB` 不行，需要单独建一个 `pgxpool.Pool`），约占 2–5 个 PG 连接（leader + job 通知 listener + worker 池）。排查时看 `river_job` / `river_leader` / `river_migration` 三张表。

## 四、决策表：按任务特征选档

| 特征 | ticker | cron 调度器 | River | asynq |
|------|--------|-----------|-------|-------|
| 零依赖 | ✅ stdlib | ❌ 需 robfig/gocron | ❌ 需 River | ❌ 需 asynq+Redis |
| 固定周期 | ✅ | ✅ | ✅ | ✅ |
| cron 表达式 | ❌ | ✅ | ✅ | ✅ |
| 任务持久化 | ❌ | ❌ | ✅ PG 表 | ✅ Redis |
| 失败重试 | ❌ | ❌（业务层搓） | ✅ 自动 + 退避 | ✅ 自动 + 退避 |
| 唯一 / 去重 | ❌ | ⚠️ 周期任务天然唯一 | ✅ unique jobs | ✅ unique tasks |
| 事务性入队 | ❌ | ❌ | ✅（PG 同事务） | ❌（Redis 异存储） |
| 并发控制 | 业务层搓 | ✅（gocron `LimitConcurrentJobs`） | ✅ worker 池 + `MaxWorkers` | ✅ worker 池 |
| 优先级队列 | ❌ | ❌ | ⚠️ 有 priority 但无严格保证 | ✅ |
| 多副本协调 | 见 [05](./05-从单机到多节点-重复跑与拆调度器.md) | 见 [05](./05-从单机到多节点-重复跑与拆调度器.md) | ✅ 内建 leader 选举 | ✅ Redis 抢占 |
| 成本 | 0 | ~10 行配置 | 2–5 PG 连接 + 3 张表 | Redis 运维 + 连接 |

**升级路径流程图**：

```
需要后台任务
  ↓
固定周期 + 可丢？
  ├─ 是 → ① time.Ticker（地板，零依赖）
  └─ 否 → 需要 cron 表达式？
            ├─ 是（且可丢） → ② cron 调度器（robfig / gocron，见 03）
            └─ 需要持久化 / 重试 / 唯一 / 事务性入队 / backfill？
                      ├─ 任一是 → ③ River（PG 栈内）
                      └─ 已重度用 Redis + 要优先级/限流 → ④ asynq
（多副本下「只跑一个」是正交问题，见 05，不在本图）
```

**关键纠偏**：「hourly backfill cron」不是优化手段，是缺持久队列的信号——「实时该做、怕崩溃漏了、定时补」= 队列 + 重试的低配翻版。一旦写出 backfill，直接上 River，别在进程内调度器外面继续搭。

## 五、成本与 tradeoff（已知并接受）

- **ticker 不能表达复杂时间**：`time.NewTicker(30*time.Minute)` 做不到「每天凌晨 2 点」「每月 1 号」，需要这些上 cron 调度器。但**不要为「将来可能」预付依赖**——当前只有固定周期任务时，ticker 就够。
- **cron 调度器不保证任务必达**：进程内、纯内存、崩溃即丢。适合「丢了下次再跑也行」的周期维护，不适合「必须执行完」的关键任务。后者用 River。
- **River 增加 PG 表与连接**：3 张系统表（`river_job` / `river_leader` / `river_migration`）+ 2–5 个常驻连接。接受——换来事务性入队 + 自动重试 + 多副本安全 + 零新基建（PG 已在栈内）。
- **asynq 的 Redis 依赖**：若 PG 已够、没优先级队列刚需，River 更简（栈内、事务性入队）。只在已重度用 Redis 时选 asynq。

## 参考

- [go-co-op/gocron](https://github.com/go-co-op/gocron)
- [River](https://riverqueue.com/)
- [asynq](https://github.com/hibiken/asynq)

**交叉引用**：全景与两根轴 [01](./01-任务调度全景-两根正交轴与知识地图.md)；cron 库对比 [03](./03-cron库选型-robfig与gocron对比.md)；多副本协调 [05](./05-从单机到多节点-重复跑与拆调度器.md)；共识原理 [06](./06-分布式共识-选主仲裁与不重复的本质.md)；生命周期编排 [../bootstrap/01](../bootstrap/01-应用启动与组件生命周期-main组合根与gin-cron编排.md)；落地约束 [service-patterns.md](../../../../.claude/rules/service-patterns.md)、[repo-transaction-convention.md](../../../../.claude/rules/repo-transaction-convention.md)。

---

**最后更新**：2026-07-15（从 bootstrap/03 迁入并重定位为「可靠性轴」执行模型；抽走多副本协调段→05、gocron 库对比段→03）
