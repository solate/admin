# 从零设计 Go 定时任务：从单机 cron 到分布式调度（2026）

> 读完你能弄懂：定时任务到底难在哪 → 单机四档执行模型（`time.Ticker` → cron 库 → 持久队列 → Redis 队列）怎么按「可丢 vs 必达」升级 → robfig/cron 与 gocron 到底怎么选 → 时间轮 / 分布式共识 / 一致性哈希这三块底层原理的实现 → 多节点下「只跑一个」怎么做才对 → XXL-Job / Elastic-Job / Temporal 这类重型方案何时才真需要。
>
> 技术栈：Go 1.26 + [robfig/cron](https://github.com/robfig/cron) v3 / [go-co-op/gocron](https://github.com/go-co-op/gocron) v2 / [River](https://riverqueue.com/)（Postgres 队列）/ [asynq](https://github.com/hibiken/asynq)（Redis 队列）。原理层覆盖 Kafka / Netty 时间轮、Raft/ZAB 共识、一致性哈希。基于真实 SaaS 后端项目的生产实践提炼。

---

## 0. 开篇：定时任务到底难在哪

「加个定时任务」听起来是最简单的需求——`cron.AddFunc("0 2 * * *", doJob)` 一行就完事。但真正落到生产，它至少藏着五个层层递进的难点，且大多数人把它们搅成了一锅：

1. **跑几遍？** 单进程里没问题，可一旦服务多副本部署（k8s `replicas: 3`），同一个任务会被 3 个副本各触发一次。发通知的任务就发 3 遍、扣费的扣 3 次。
2. **丢不丢？** 进程崩溃、重启的那段时间，到点的任务错过了怎么办？「下次再跑」行不行，还是「必须补上」？
3. **海量定时器怎么扛？** 几十个周期任务无所谓，可要管几十万个超时定时器（每个连接一个），用什么数据结构？
4. **多节点怎么协调「只跑一个」？** 加把 Redis 锁就够了吗？为什么有人说那不可靠？
5. **什么时候才需要 XXL-Job / Temporal 这种重家伙？** 是不是项目大了就该上？

这篇文章的核心主张是：**这五个问题分属两条完全不同的轴，绝大多数人的困惑来自把它们混成一条「升级链」。** 想清楚了轴，选型就不再纠结。下面先建这套心智模型，再逐层深挖实现。

## 1. 心智模型：两根正交轴 + 三件独立的事

### 1.1 任务调度是三件独立的事

「任务调度」是个过载词，拆开看是三件互不相干的事：

| 关注点 | 回答什么 | 手段 |
|---|---|---|
| **触发** | 什么时候跑 | cron 表达式解析、时间轮、`time.Ticker` |
| **协调** | 谁来跑 / 跑几遍 | 拆调度器单副本、分布式锁、共识选主 |
| **执行保证** | 跑了会不会丢 / 要不要重试 | 持久队列、重试退避、幂等 |

关键认识：**这三件事可以任意组合，不是绑定的。** 「用了 cron 表达式」不代表「解决了多副本」；「多副本只跑一个」也不代表「任务不会丢」。把它们分开，才能对症下药。

### 1.2 两根正交轴

再往上抽象，所有方案落在两根**互相独立**的轴上：

```
可靠性轴（丢不丢）
  可丢 ────────────────────────────────► 必达
  time.Ticker    cron 库      River（持久队列）
  （崩了就丢）   （崩了就丢）  （崩溃恢复、重试、事务性入队）

吞吐/规模轴（一个节点够不够）
  单点够用 ──────────────────────────────► 海量算不完
  单进程调度   拆调度器 replicas:1   分布式调度引擎
                                    （XXL-Job / Elastic-Job）
```

各方案钉在坐标上：

| 方案 | 可靠性轴 | 吞吐/规模轴 | 一句话定位 |
|---|---|---|---|
| `time.Ticker` | 可丢档 | 单点 | 零依赖地板 |
| robfig / gocron | 可丢档 | 单点 | cron 表达式触发 |
| 拆调度器 `replicas:1` | 不变 | 单点（消灭协调问题） | 多副本下的架构首选 |
| River / asynq | 必达档 | 单点~中 | 持久队列，崩溃恢复 |
| XXL-Job / Elastic-Job | 看配置 | 海量分片 | 吞吐轴末端 |
| Temporal | 必达 + 编排 | 另一维度 | 工作流引擎 |

**两个最重要的立场：**

- **两轴正交，绝大多数项目只走可靠性轴。** 你的后台任务从「可丢」变「必达」（上 River），跟「要不要分布式引擎」（吞吐轴）没有半点关系。
- **最常见的过度设计，是把可靠性需求误当规模需求。** 「我的任务不能丢」→ 有人就去上 XXL-Job，其实要的是 River。分布式引擎是吞吐轴的**岔路**，不是升级链的顶点。

### 1.3 先钉死三个高频误解

这三个误解后面各有一章专门深挖，这里先给结论，读的时候心里有数：

1. **「时间轮能解决多副本」——错。** 时间轮是单机内的数据结构（性能优化），和「跑几遍」正交（见第 4 章）。
2. **「一致性哈希保证数据一致 / 不重复」——错。** 它是分片手段，「一致」指的是「节点增减时映射稳定」，不是 CAP 的 consistency，更不保证不重复（见第 7 章）。
3. **「多副本只跑一个 = 加把锁」——不够。** 它本质是分布式共识；Redis 锁只是 best-effort（见第 6 章）。

阅读路径：**第 2-4 章是单机段**（执行模型、库选型、时间轮），**第 5-7 章是多节点段**（重复跑、共识、一致性哈希），**第 8 章是成熟方案**（何时上重型引擎）。

## 2. 单机执行模型：从 ticker 到持久队列的四档

先把视角收到单机（一个进程）。沿可靠性轴，后台任务有四档递进方案，核心区分是**「可丢 vs 必达」**——不是「定时 vs 按需」。

| 任务形态 | 特征 | 典型例子 | 丢了有没有事 | 对应方案 |
|---|---|---|---|---|
| 周期轻量维护 | 定时、轻、无副作用 | 清理过期 session | 无事，下次再跑 | ① `time.Ticker` |
| 周期复杂时间 | 需要 cron 表达式 | 每天凌晨 2 点统计 | 无事，下次再跑 | ② cron 库 |
| 事件触发必达 | 按需、要重试、要唯一 | 审核通过发邮件 | 必须执行完 | ③ River |
| 崩溃恢复 backfill | 「怕漏了补扫」 | 本该实时做、定时兜底 | 这就是 ③ 的信号 | ③ River |

### 2.1 ① 地板档：`time.Ticker`（可丢 + 固定周期）

最轻一档不需要任何库，stdlib 的 `time.Ticker` 塞进一个 goroutine 就够。适用条件三个同时满足：**固定周期**（不需要 cron 表达式）、**丢了无所谓**（崩溃期间没跑，下次启动再跑即可）、**单进程触发**。

```go
// 用 errgroup 管理，让优雅关闭能打断它
g.Go(func() error {
    ticker := time.NewTicker(30 * time.Minute)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            cleanupExpiredSessions(ctx, db) // 业务逻辑，传可取消 ctx
        case <-ctx.Done():
            return nil // 收到关闭信号，退出
        }
    }
})
```

关键点：ticker goroutine **必须监听 `ctx.Done()`**，业务函数传**可取消的 ctx**（别传 `context.Background()`），`defer ticker.Stop()` 释放资源。零依赖、零配置——够用时这是最简的对。

**何时不够**：需要 `0 2 * * *`（每天凌晨 2 点）这类复杂时间规则，ticker 表达不了 → 升到 cron 库。

### 2.2 ② cron 库：复杂时间规则

需要 cron 表达式（每天几点、每周几、每月几号）时，上一个进程内 cron 调度器。库选型是下一章的主题，这里先记住：**它仍是纯内存、崩溃即丢**，只是把「何时触发」从固定间隔升级到了 cron 表达式。可靠性上它和 ticker 同档。

```go
s, _ := gocron.NewScheduler()
s.NewJob(
    gocron.CronJob("0 2 * * *", false),       // 每天凌晨 2 点
    gocron.NewTask(func() { heavyReport(ctx, db) }),
)
s.Start()
```

**何时不够**（升到 River 的信号，满足任一即升）：任务**不能丢**、失败要**自动重试**、要**唯一**（同 key 只入队一次）、要**事务性入队**（业务 commit 了任务才入队）、出现**「backfill / 补扫」需求**。

### 2.3 ③ 必达档：River（Postgres 持久队列）

[River](https://riverqueue.com/) 是 Postgres-backed 的 Go 原生任务队列。与 cron 库的本质区别：**cron 库是调度器（决定何时跑），River 是队列（保证必达）**。任务落在 PG 表里，崩溃重启后继续跑；失败自动重试；支持唯一去重；最关键是**事务性入队**——任务和业务数据在同一个事务里提交。

```go
// 业务层入队：与业务 DB 操作在同一事务
tx.Exec("UPDATE reviews SET status='approved' WHERE id=$1", reviewID)
riverClient.InsertTx(ctx, tx, EmailJobArgs{To: email}, nil)
tx.Commit() // 业务 commit，任务才真入队；业务回滚则任务不入队
```

周期任务也能交给 River，由它的 **leader 副本**（通过 PG advisory lock 选举）自动插入，多副本天然安全：

```go
riverClient, _ := river.NewClient(dbPool, &river.Config{
    PeriodicJobs: []*river.PeriodicJob{
        river.NewPeriodicJob(
            river.PeriodicInterval(30*time.Minute), // 或 river.Cron("0 2 * * *")
            func() (river.JobArgs, *river.InsertOpts) { return CleanupJobArgs{}, nil },
            &river.PeriodicJobOpts{RunOnStart: true},
        ),
    },
})
```

唯一 + 重试：

```go
riverClient.Insert(ctx, ProcessJobArgs{ResourceID: id}, &river.InsertOpts{
    UniqueOpts:  river.UniqueOpts{ByArgs: true}, // 同参数只入队一次
    MaxAttempts: 5,                              // 失败最多重试 5 次，指数退避
})
```

**关键纠偏**：「每小时全扫一遍怕漏的记录再补处理」这种 **backfill cron，本质是缺持久队列的信号**——「实时该做、怕崩溃漏了、定时补」= 队列 + 重试的低配翻版。一旦写出 backfill，别在 cron 外面继续搭，直接上 River。

### 2.4 ④ Redis 侧备选：asynq

若技术栈已重度使用 Redis，且要优先级队列 / 基于 Redis 的限流 / web 监控面板，可选 [asynq](https://github.com/hibiken/asynq)。但它做不到 River 的**事务性入队**（Redis 与业务 DB 是两个存储，无法同事务）。若技术栈里已有 Postgres，River 通常更简——零新基建、事务性入队。

### 2.5 按特征选档 + 升级决策树

| 特征 | ticker | cron 库 | River | asynq |
|---|---|---|---|---|
| 零依赖 | ✅ stdlib | ❌ | ❌ | ❌ 需 Redis |
| cron 表达式 | ❌ | ✅ | ✅ | ✅ |
| 任务持久化 | ❌ | ❌ | ✅ PG 表 | ✅ Redis |
| 失败重试 | ❌ | ❌ | ✅ 自动+退避 | ✅ 自动+退避 |
| 唯一 / 去重 | ❌ | ⚠️ 周期任务天然唯一 | ✅ unique jobs | ✅ |
| 事务性入队 | ❌ | ❌ | ✅ PG 同事务 | ❌ 异存储 |

```
需要后台任务
  ↓
固定周期 + 可丢？ ──是──► ① time.Ticker
  │否
需要 cron 表达式（仍可丢）？ ──是──► ② cron 库
  │否
需要 持久化/重试/唯一/事务性入队/backfill？ ──任一是──► ③ River
                                          └─ 已重度用 Redis + 要优先级/面板 ──► ④ asynq
```

## 3. cron 库层：robfig 与 gocron 怎么选

第 2 章的「② cron 库」到底选哪个？Go 生态两大主流：`robfig/cron` 和 `go-co-op/gocron`。先给结论：**纯库层看，两者差异是「加分项」而非「刚需」；而多副本协调是架构问题（第 5 章），不是选库能解决的。**

### 3.1 能力对比（事实层）

| | robfig/cron v3 | gocron v2 |
|---|---|---|
| 本质 | 极小的纯调度内核（解析表达式 → 到点触发） | 完整调度器框架 |
| 维护状态 | **基本冻结**：v3.0.1 停在 2020 年 | **活跃**：定期发版 |
| 生态地位 | Go 事实标准，K8s 生态在用，久经沙场 | 应用层最流行之一，但无同量级「招牌大型采用」 |
| 调度类型 | 仅 cron 表达式 + `@every` | Duration / Daily / Weekly / Monthly / OneTime / Cron |
| 单任务串行 | 需自己加 wrapper | 内建 `SingletonMode` |
| 全局并发上限 | 无 | 内建 `LimitConcurrentJobs` |
| 生命周期钩子 | 无 | 内建 Before / After / OnError |
| 任务收 context | 无（`func()`） | 有（可被优雅关闭打断） |
| 多副本协调 | 无 | 内建 `Locker` / `Elector`（Redis/etcd，可选模块） |
| API 形态 | `AddFunc(spec, fn)`，极简 | `NewScheduler` + `NewJob` + option，较重 |

**gocron 真正强在哪（不含水分）**：① 还在维护 vs 已冻结；② 调度类型丰富（不用把「每天几点」硬翻译成 cron 串）；③ 内建单任务串行；④ 内建并发上限 / 钩子 / context。这些都是**加分项、非刚需**——不足以成为「为它引一个更重依赖」的唯一理由。

### 3.2 两个必须知道的事实

**事实一：gocron 底层的 cron 解析仍用 robfig 的 parser。** gocron v2 的调度循环是它自研的（基于 `time.Timer` + clock 抽象），但 cron 表达式解析这一步复用了 `robfig/cron/v3` 的 parser。所以它是「换引擎、留油泵」，不是「另起炉灶」——两者不是竞争对手那么简单的关系。

**事实二：gocron 的分布式 Locker 是 best-effort，不是 exactly-once。** 它的 `WithDistributedLocker`（基于 Redis）存在时钟偏移、锁续期竞态的窗口，极端情况仍可能两个副本都跑。**有副作用的任务该幂等还得幂等。** 这背后的原理是第 6 章的主题。

### 3.3 单机选型顺序

1. **只要固定周期** → `time.Ticker`，连 cron 库都不引。
2. **要 cron 表达式、单进程** → robfig（极简、久经沙场）或 gocron（功能全）都行，选熟悉的。gocron 的招牌功能（分布式协调）单机下用不上。
3. **纠结「新库更好」时** → 记住 robfig 冻结不代表差，内核小到「基本没 bug 可出」是一种优点；gocron 功能多、代码面也大。
4. **多副本** → 先看第 5 章（大概率不该靠库解决）。

---

## 4. 深挖①：时间轮——单机海量定时器的数据结构

前面反复说 robfig/gocron「底层是最小堆」。那有没有更快的结构？有——**时间轮（timing wheel）**，Kafka、Netty 都在用。但这里要先钉死一个高频误解：

> **时间轮是单机内的数据结构，解决的是「一个进程里多快能管理海量定时器」这个性能问题。它和「多副本会不会重复跑」完全正交——它不是分布式方案，别把它当协调手段。**

很多人把「时间轮」和「分布式调度」混在一起谈，这是错的。下面讲清它到底是什么、快在哪、以及为什么它跟多副本无关。

### 4.1 问题：海量定时器怎么存

「定时器」的核心操作有三个：**添加**一个「T 秒后触发」的任务、**取消**一个还没触发的任务、每个时刻**找出所有到期**的任务执行。

最直觉的实现是**最小堆/优先队列**（按到期时间排序）：

- 添加：O(log n)——插入并上浮。
- 取消：O(log n)——删除并调整。
- 取最近到期：O(1) 看堆顶，弹出 O(log n)。

n 是几十、几百时，O(log n) 毫无压力——**这正是业务 cron 的量级，最小堆绰绰有余**。但当 n 到几十万、上百万——网络框架里「每个 TCP 连接一个空闲超时」、消息队列里「每条延迟消息一个投递定时器」——每次增删都 O(log n)、且增删极频繁，堆调整的开销就变得可观。时间轮就是为这个量级设计的。

### 4.2 时间轮原理：环形数组 + 指针

时间轮借用钟表的比喻：一个**环形数组**，每个格子（slot/bucket）代表一个时间单位（tick，比如 1 秒），一根**指针**每 tick 前进一格，走到头绕回开头。

```
        slot0  slot1  slot2  slot3  ...  slotN-1
        [·]    [任务] [·]    [任务]       [·]
                              ↑
                            指针（每 tick 前进一格）
每格挂一个链表，存「将在这一格到期的所有任务」
```

- **数组有 N 个格子**，每格挂一个链表，存「将在这一格到期的所有任务」。
- **添加任务**：算出它在几个 tick 后到期，`slot = (当前指针 + delay) % N`，直接挂到那个格子的链表——**O(1)**，不需要和其他任务比较排序。
- **取消任务**：从链表摘掉——**O(1)**（配合一个 `任务 → 节点` 的 map）。
- **推进**：指针每 tick 前进一格，取出当前格链表里的所有任务执行——**均摊 O(1)**，不用扫描整个结构找「谁到期了」。

对比最小堆：时间轮用「**空间换比较**」——不排序，靠「到期时间 → 数组下标」的直接映射定位，把 log n 的比较开销消掉了。代价是精度受 tick 粒度限制（1 秒的轮做不了毫秒级），以及超过一圈的延迟需要额外处理（见下）。

### 4.3 分层时间轮：解决时间跨度 vs 精度

单层时间轮有个矛盾：**格子数有限，覆盖的时间跨度 = 格子数 × tick**。要精度高（tick=1s）又要跨度大（能表达「7 天后」），单层就得开 60 万个格子——浪费。

**分层时间轮（hierarchical timing wheel）**照搬钟表的时/分/秒三针：

- 第一层：秒轮，60 格，tick=1s，覆盖 1 分钟。
- 第二层：分轮，60 格，tick=1min，覆盖 1 小时。
- 第三层：时轮，24 格，tick=1h，覆盖 1 天。

一个「1 小时 23 分 5 秒后」的任务先挂在时轮。当时轮指针推进到它那一格，把它**降级**重新分配到分轮（还剩 23 分 5 秒）；分轮推进到那格再降级到秒轮（还剩 5 秒）；秒轮到点才真正触发。这样**少量格子就能覆盖很大跨度**，且保持底层精度。Kafka 的时间轮就是这种多层结构 + 溢出时动态加更高层。

### 4.4 工业实现

时间轮不是学术玩具，主流基础设施里到处是它：

- **Kafka**：`DelayedOperationPurgatory`（延迟操作管理，如 producer acks 超时、fetch 等待）底层是分层时间轮 + `DelayQueue` 推进。
- **Netty**：`HashedWheelTimer`——单层哈希时间轮，用于海量连接的空闲检测、请求超时。文档明确说它适合「大量超时任务、但精度要求不高」。
- **Linux 内核**：内核定时器（`timer_list`）历史上就是多级时间轮（cascading timing wheel）。
- **Go 生态**：`time.Timer`/`time.Ticker` 底层是运行时的**四叉最小堆**（per-P timer heap），**不是**时间轮——所以基于 `time.Timer` 的调度器（robfig/gocron）底层也是堆。Go 社区有独立时间轮库（如 `RussellLuo/timingwheel`），但标准库没走这条路，因为 Go 程序里定时器数量通常不到需要时间轮的量级。

### 4.5 关键澄清：时间轮 ≠ 多副本协调

这是本章最重要的一句：

> 时间轮解决的是「一个进程内，多快能管理和触发海量定时器」——纯粹的单机性能问题。它和「多个副本会不会把同一个任务跑多遍」毫无关系。

- 你在单机调度器里用时间轮，一样是纯内存、一样在多副本下每个副本各跑一次。
- Kafka 能多副本安全，**靠的不是时间轮**，是它的分区独占 + controller 选主（共识，见第 6 章）；时间轮只是每个 broker 进程内管理延迟操作的数据结构。

所以「用了时间轮就能多副本」这个因果**根本不成立**。时间轮回答「多快」，不回答「跑几遍」。「跑几遍」是第 5、6 章的题目，和数据结构无关。

**对绝大多数业务后台：你不需要关心时间轮。** 了解它是为了看懂 Kafka/Netty 的底层，以及破除上面这个误解。

---

## 5. 从单机到多节点：重复跑与拆调度器

现在进入多节点。这是全篇的枢纽——也是最多人一上来就选错方案的地方。

先说结论：**分水岭不是「副本数」、更不是「新旧」，而是「调度器要不要单独拆成一个服务」。**

### 5.1 多副本下，定时任务会发生什么

多副本几乎从不是为定时任务开的，而是业务服务为了高可用/负载均衡顺带的结果：K8s `replicas: 3`、多台机器、滚动更新时新旧版本并存。

如果你把定时任务代码和业务代码打进**同一个二进制**、再 `replicas: 3` 部署：3 个一模一样的 Pod，每个进程都执行了同一段「注册定时任务」的代码，它们**互不知道对方存在**。到了 `0 2 * * *`，**3 个 Pod 各自触发一次 → 同一个任务被执行 3 遍**。

后果取决于任务性质：

| 任务类型 | 跑 3 遍的后果 |
|---|---|
| 幂等（清理过期 session、`DELETE WHERE expired`） | 没事。3 个删同一批，结果一样，顶多浪费点 DB 查询 |
| 累加/统计（把今日订单数写进报表表） | **数据错**。可能算 3 次，或 3 个并发读到中间态互相覆盖 |
| 有外部副作用（发邮件、短信、扣费、调第三方 API） | **事故**。用户收到 3 封邮件、扣 3 次费 |
| 写文件/生成导出 | **竞争**。3 个同时写同一路径，文件损坏或互相覆盖 |

**关键：cron 库完全不知道自己在多副本里，它没有任何机制阻止这件事。** 这不是库的 bug——它的职责就只是「到点触发」，协调是别人的事。

### 5.2 两条路：解决「多副本跑 N 遍」

要让「到点只跑一次」，有两条**互斥**的路。

**路 A：库层面——上 gocron + Redis Locker（分布式锁的招牌功能）**

cron 和业务仍挤在同一个多副本部署里。3 个 Pod 到点都想跑，但先去 Redis 抢一把锁，只有抢到的那个真跑，另外俩看到锁被占就跳过：

```go
// 选项 1：Locker（基于 Redis/etcd 的分布式锁，任一时刻只一个副本执行）
redisLocker := redislock.NewRedisLocker(rdbClient)
s, _ := gocron.NewScheduler(gocron.WithDistributedLocker(redisLocker))
s.NewJob(
    gocron.CronJob("0 2 * * *", false),
    gocron.NewTask(heavyReport),
    gocron.WithSingletonMode(gocron.LimitModeReschedule), // 拿不到锁→下次再试
)

// 选项 2：Elector（选主，leader 独占全部周期任务）
redisElector := redislock.NewRedisElector(rdbClient)
s, _ := gocron.NewScheduler(gocron.WithDistributedElector(redisElector))
// 整个 scheduler 只在 leader 副本跑
```

- **代价**：引 gocron + 依赖 Redis + 每个任务配 Singleton。
- **且是 best-effort**：锁续期竞态、时钟偏移、主从切换下，极端情况仍可能两个副本都跑，**不保证严格一次**（根因见第 6 章）。有副作用的任务该幂等还得幂等。

**路 B：架构层面——把调度器单拆成 `replicas:1`（首选）**

业务还是 `replicas: 3`（但这些 Pod 里**不注册** cron），另起一个只有 `replicas: 1` 的调度器部署单元。全局只有 1 个进程会触发定时任务，「跑一次」天然成立，**零锁、零 Redis、零分布式库**。`time.Ticker`/robfig 照样够。

落地「调度器单副本」三选一，都比分布式锁简单：

- **配置开关（最贴合单体架构）**：同一个二进制加个 `ENABLE_SCHEDULER` 配置。业务 Deployment 跑 `replicas: N` + `ENABLE_SCHEDULER=false`；另起一个 Deployment 跑 `replicas: 1` + `ENABLE_SCHEDULER=true`，只有它注册 cron。代码一份，部署两份。
- **独立部署**：把调度器拆成单独的小服务/二进制，`replicas: 1`。
- **K8s CronJob**：干脆让 K8s 定时拉起 Job pod，连进程内调度器都不要。

### 5.3 两条路对比：消灭问题 vs 解决问题

| | 路 A：库层面（分布式锁） | 路 B：架构层面（拆调度器，首选） |
|---|---|---|
| 做法 | cron 和业务挤在同一多副本部署，靠 Redis 锁去重 | 调度器单拆 `replicas:1`，业务无状态多副本 |
| 依赖 | gocron + Redis + 每任务 Singleton 配置 | 无（`time.Ticker`/robfig 即可） |
| 保证 | best-effort，极端情况仍可能重复跑 | 全局只有一个进程会触发，天然只跑一次 |
| 复杂度 | 分布式锁的续期/防死锁/时钟问题都要理解 | 一份代码 / 一个部署单元 |

这里有个关键视角差异：**路 A 是「解决共识问题」（花力气让 N 个副本协调成 1 个），路 B 是「消灭共识问题」（让 N=1，问题根本不出现）。** 对中小项目，消灭永远比解决省。分布式锁那套（Redis Locker）解决的正是「我不想拆、就想让挤在一起的多副本协调成一个」这个前提——你一旦选路 B，这个前提不存在，那套东西的价值就归零。

### 5.4 边界：拆调度器无 HA

拆调度器 `replicas:1` 有一个诚实的代价：**没有 HA**。那个进程重启/发布的几秒钟里正好到点，这一次触发会漏。

- 周期维护任务漏一次，下次补上即可——可接受。
- **如果「漏一次都不行」**，那不是 cron 库或分布式锁的问题，是你需要**持久队列**（第 2 章的 River）：任务落库、崩溃恢复、leader 自动接管。这跟选 robfig 还是 gocron 无关。

一句话：多副本先想「能不能拆调度器」，几乎总能；不能拆且任务幂等，才考虑 Redis 锁；要严格一次或不能丢，上持久队列。

---

## 6. 深挖②：分布式共识——「只跑一个」的本质

第 5 章说 Redis 锁是 best-effort、拆调度器是「消灭问题」。这一章回答**为什么**：为什么 Redis 锁不可靠，而 etcd / ZooKeeper / Postgres 这些系统却「天然多副本安全」。

### 6.1 问题的真面目：这是一个共识问题

把「多副本只跑一个」抽象干净：N 个对等进程，同一时刻都想执行同一个任务，我们要让**其中恰好一个**真正执行。

它看起来只是「加把锁」，实则是分布式系统里最难的一类问题——**共识（consensus）**：多个节点要对「谁是那个唯一」达成一致，且这个判断在节点崩溃、网络分区、消息延迟下**依然正确**。自己在业务层用 Redis 搓锁之所以总踩坑，就是因为共识很难做对。

### 6.2 共识算法速览：Paxos / Raft / ZAB

共识算法解决的核心问题：**一组节点如何对某个值达成一致，即使部分节点故障或网络不可靠。**

- **Paxos**（Lamport, 1998）：理论奠基，正确但难懂难实现。
- **Raft**（2014）：为「可理解性」重新设计，拆成 leader 选举 + 日志复制 + 安全性三块。etcd、Consul、TiKV 都用它。
- **ZAB**（ZooKeeper Atomic Broadcast）：ZooKeeper 的协议，思路与 Raft 相近。

共同保证：**在多数派（quorum，> N/2 节点）存活时，系统对「当前值 / 当前 leader」有唯一且线性化的答案。** 关键词是「多数派」——共识不依赖任何单个节点存活，这正是它区别于「单点 Redis 锁」的根本。

对任务调度，我们不需要自己实现共识，只需用它派生出的一个能力：**选主（leader election）**。

### 6.3 「不重复」的三件套：共识 + fencing + 幂等

「让多副本只跑一个」落到工程上，是三件事的叠加，缺一不可：

**① 共识选出唯一 owner。** 通过 Raft/ZAB 或数据库原子操作选出一个 leader（或某任务的锁持有者）。共识保证同一时刻**至多一个** owner。

**② Fencing token 挡住「过期的 owner」。** 这是最容易被忽略、也是 Redis 锁最致命的缺口。设想这个时序：

```
副本 A  ── 拿到锁 ──▶ GC 停顿 10s（或网络分区）…………… 恢复，继续往下写 ✗
                          │
                          ▼ 锁 TTL 到期，自动释放
副本 B          ────────── 拿到锁 ──▶ 开始执行 ✓
```

A 从停顿中恢复时**并不知道自己的锁已过期**，继续写——于是 A 和 B 同时在跑。

Fencing token 的解法：每次授予锁都附带一个**单调递增的序号**（token），B 拿到的 token 比 A 大。下游存储只接受「token ≥ 已见过的最大值」的写入——A 带着旧 token 来写会被拒绝。**没有 fencing token 的锁，本质上都无法防住这种「过期持有者」竞态。**

**③ 幂等兜底。** 即便前两者到位，工程上仍建议任务幂等（`execution_id` 去重表、`INSERT ... ON CONFLICT`）。共识 + fencing 把重复压到极低概率，幂等让「万一重复」也无害。

**一句话：不重复 = 共识（选唯一）+ fencing（挡过期）+ 幂等（兜底）。** 只做第一步（比如裸 Redis 锁），就是 best-effort。

### 6.4 为什么 etcd / ZK / Kafka / Postgres「天然多副本」

理解了三件套，就看清了这些系统的共同套路：**副本自己保持无状态、对等，把「谁来跑 + 状态存哪」外置到一个内部已解决共识的强一致存储。**

| 系统 | 共识/原子原语 | 怎么仲裁出「只有一个」 |
|---|---|---|
| **etcd** | Raft | 内建 leader election；`CAS` + `lease`（带 TTL 的租约，续约由多数派共识保证）+ revision 号做 fencing |
| **ZooKeeper** | ZAB | 临时有序节点（ephemeral sequential）+ watch 实现选主；zxid 单调递增天然是 fencing token |
| **Kafka** | 副本 ISR 日志 + controller | 一个 partition 在一个消费组里**只分配给一个消费者**（单一 owner）；controller 负责分配 |
| **Postgres** | ACID + advisory lock / `SELECT FOR UPDATE SKIP LOCKED` | 单写者串行化：选主用 advisory lock，任务领取用行锁，数据库保证同一行只被一个 worker 抢到 |

共同点一句话：**「只有一个跑」这个不变量，不是靠副本之间互相商量，而是靠一个提供了原子/线性化操作的外部存储来裁决。** 副本只管去问它「我能跑吗」，存储原子地回答，并用序号/revision 挡住过期请求。

这也解释了「天然多副本」往往**同时**给了两个东西：**协调**（谁来跑，靠共识原语）+ **持久化**（状态在外部存储，副本崩了状态还在）。这两个属性的绑定，正是第 2 章 River「多副本天然安全」的来源。

### 6.5 Redis SETNX 为什么不是真共识

回到最初的问题。`SET NX EX` 做锁，缺的正是上面几条：

1. **单点，不是多数派共识。** 单实例 Redis 挂了锁信息全丢；它的「原子」只是单机命令原子，不是分布式共识。
2. **无 fencing token。** `SETNX` 只返回成功/失败，没有单调序号，6.3 那个竞态防不住。
3. **主从切换 split-brain。** 生产 Redis 多是主从 + Sentinel。主写入锁后、还没同步到从就宕机，Sentinel 把从提为主——新主上没这把锁，另一个副本又拿到了，**两个副本同时持锁**。
4. **锁靠时钟 TTL。** 续期依赖客户端时钟和网络，时钟漂移或续期延迟都会让锁提前失效或迟迟不放。

**Redlock 的争议**：Redis 作者提出 Redlock（多个独立 Redis 实例投票）试图补强，但 Martin Kleppmann 有著名批评——Redlock 依赖「时钟不漂移」「进程不长时间停顿」这类不安全假设，且仍无 fencing token，无法在异步网络模型下证明正确。结论：**要严格正确的分布式锁，用带 fencing 的共识系统（etcd/ZK）或数据库，而不是 Redis。**

这就是「为什么 etcd 直接能多副本、Redis 锁却只能 best-effort」的根因——**多数派共识 + fencing vs 单点原子命令，不是一个量级。**

---

## 7. 深挖③：一致性哈希与分片调度

这一章处理另一个高频混淆：**一致性哈希（consistent hashing）不是用来「保证数据一致性」的**。名字里的「一致」是历史误译坑——它保证的是「节点增减时映射关系的**稳定性**」，属于**负载分片**手段，和 CAP 里的 consistency（数据一致/不重复）是两码事。

### 7.1 问题：取模的全量重排

你有 10 万个 job、10 个调度节点，想把 job 分摊下去，每个节点只管约 1 万个（这是**吞吐轴**上的分片需求——单点触发扛不住了）。

最直觉的映射是**取模** `hash(job_id) % N`。问题在 N 一变（加/减节点、或节点挂了）：

```
N=10 → N=9：hash(job) % 10  vs  hash(job) % 9
几乎每个 job 的归属都变了 → 10 万个 job 近乎全量迁移
```

迁移期间大面积重复或漏跑。取模的致命伤就是**扩缩容即全量重排**。

### 7.2 一致性哈希原理：哈希环 + 虚拟节点

一致性哈希把节点和 job 哈希到同一个**环形值域**（如 0 ~ 2³²-1）上：

1. 每个**节点**哈希到环上若干点。
2. 每个 **job** 也哈希到环上一点。
3. job 归属**顺时针方向遇到的第一个节点**。
4. 加/减节点时，只有**环上相邻的那段** job 需要换归属，其余不动。

节点增减只影响相邻约 **1/N** 的 job，而非全量。为避免节点在环上分布不均导致倾斜，每个物理节点放大成成百上千个**虚拟节点**均匀撒在环上，负载就均衡了。

### 7.3 它解决什么、不解决什么

这是本章的核心，也是纠误所在：

- **解决**：分片映射稳定（扩缩容只迁移 1/N）、负载均摊（虚拟节点）、平滑扩缩容。
- **不解决**：① 它**不是**「数据一致性」——名字里的 consistent 指的是映射的一致/稳定，不是 CAP 的 C；② 它**不保证「不重复跑」**——它只决定「哪个节点大概率负责这片」，同一片在成员视图不一致时仍可能被两个节点都认领。

而且一致性哈希有个**隐藏前提**：所有节点必须对「环上现在有哪些节点」有一致的视图。若 A 以为 B 还活着、C 以为 B 已死，两者算出的归属就不同，同一 job 被两个节点认领 → 重复跑。所以**一致性哈希永远不单独用，底下必须垫一个共识/成员管理层**（ZK/etcd/gossip 维护一致的成员视图）。

### 7.4 分片与仲裁：正交对称

把第 6 章（共识）和本章（一致性哈希）摆在一起，它们是一对正交的镜像：

| | 一致性哈希（分片） | 共识 + fencing（仲裁） |
|---|---|---|
| 解决什么 | 把海量 job **分摊**到多节点、扩缩容少迁移 | 保证同一 job/片同一刻**只有一个**在跑 |
| 属于哪根轴 | 吞吐/规模轴 | 可靠性轴（协调面） |
| 能单独用吗 | 不能，需共识垫底提供一致成员视图 | 能，本身就是最终裁决 |
| 叠加使用 | 分片决定「谁大概率负责这片」 | 仲裁决定「这片同一刻谁真在跑」 |

真实的分布式调度引擎两者叠加：一致性哈希把 job 打散（吞吐），共识 + fencing 保证每片不重复（可靠性）。下一章的 Elastic-Job 就是活标本。

---

## 8. 成熟方案：XXL-Job / Elastic-Job / Temporal，何时才真需要

走到这里，可以回答开篇第 5 个痛点了：**这些「重型分布式调度引擎」是不是终点？** 不是——它们坐在**吞吐/规模轴的末端**，不是升级链顶点。最常见的误判，是把**可靠性需求**误当成**规模需求**。

### 8.1 先自查：你的痛点在哪根轴

| 你的真实痛点 | 该走的方向 | 别做的事 |
|---|---|---|
| 任务不能丢、要崩溃恢复、要重试 | **可靠性轴** → River / 持久队列 | 别上分布式调度引擎——那不解决「丢」 |
| 多副本只想让一个跑 | **协调** → 拆调度器 replicas:1 | 别为这个引 ZooKeeper 集群 |
| job 多到/重到一个节点算不完 | **吞吐轴** → 分布式调度引擎 | 这才是引擎的主场 |

只有第三行——**真正的吞吐瓶颈**——才需要下面这些引擎。99% 的中小项目止步于「拆调度器 + River」。

### 8.2 XXL-Job：中心调度 + 执行器分离

- **架构**：一个「调度中心」（管到点触发、HA 靠 DB 行锁/主备）+ 执行器集群（真正跑任务）。元数据存 MySQL，带可视化管理台。
- **路由策略**：轮询 / 随机 / 一致性哈希（同 job 固定打到同一执行器）/ 分片广播。
- ✅ 上手快、有界面、故障转移成熟，国内生态广。⚠️ 调度中心是中心节点（虽可 HA），重依赖 MySQL；偏「定时触发 + 派发」，不做工作流编排。

### 8.3 Elastic-Job：去中心 + ZK 选主 + 一致性哈希分片

- **架构**：去中心化，依赖 **ZooKeeper** 做选主与成员协调（回扣第 6 章共识）+ **一致性哈希分片**（回扣第 7 章）。
- **分片模型**：把一个大 job 切成 N 片，每个节点认领自己那几片并行跑：

```java
// 作业按分片项拆分，每个节点只处理分到的片
switch (shardingContext.getShardingItem()) {
    case 0: processUsers(0, 10000);      break;
    case 1: processUsers(10000, 20000);  break;
    // …ZK 保证分片分配的一致视图，节点增减自动 rebalance
}
```

- ✅ 无中心瓶颈、分片能力强，适合「一个大任务拆多节点并行」。⚠️ 强依赖 ZooKeeper 集群，运维复杂度上一个台阶。

### 8.4 Temporal / Cadence：另一个维度——工作流编排

Temporal（及其前身 Uber Cadence）和上面两者**本质不同**：它不是「定时触发器」，而是 **durable execution 工作流引擎**——把一段可能跑几天、带多步骤和补偿的业务流程，以 event sourcing + 重放的方式持久化，进程崩了能从中断处精确恢复。

它兼跨可靠性与编排两个维度，解决的是「长流程、多步骤、要精确一次语义的编排」，而非「每天 2 点跑个清理」。用它来做简单定时任务是高射炮打蚊子。

### 8.5 横向对比

| 维度 | XXL-Job | Elastic-Job | Temporal | River |
|---|---|---|---|---|
| 范式 | 中心调度+派发 | 去中心分片 | 工作流编排 | 持久队列 |
| 主解决 | 定时触发+可视化 | 海量 job 分片 | 长流程 durable execution | 必达+重试 |
| 主要轴 | 吞吐（中等） | 吞吐（高） | 可靠性+编排 | 可靠性 |
| 协调机制 | DB 行锁/主备 | ZooKeeper 选主 | 内部共识集群 | PG advisory lock |
| 分片 | 路由策略 | 一致性哈希 | 无（按 workflow） | worker 池抢占 |
| 新基建 | MySQL + 调度中心 | ZooKeeper 集群 | Temporal 集群 | 复用 Postgres |
| 运维成本 | 中 | 高 | 高 | 低 |
| 适用规模 | 中大 | 大（分片） | 中大（编排） | 中小 |

### 8.6 钉回两根轴

- XXL-Job / Elastic-Job 坐在**吞吐轴末端**——为「job 多到一个节点扛不住」而生。
- Temporal 在**可靠性 + 编排**维度，是长流程场景的专用解。
- 它们**都不是**「gocron 的下一级」这种线性关系。从 gocron 到它们，你跨的是轴，不是档位。

先确认你真在吞吐轴走到了尽头，再考虑它们。否则「拆调度器 + River」几乎覆盖所有中小项目。

---

## 结语

回到开篇的五个痛点，逐一收束：

1. **多副本会不会跑多遍** → 会。同一二进制 replicas:3，同一任务跑 3 遍（第 5 章）。解法优先是拆调度器 replicas:1，让问题消失。
2. **丢了怎么办** → 分「可丢 vs 必达」（第 2 章）。可丢用 ticker/cron，必达用持久队列 River；backfill cron 是「该上持久队列」的信号。
3. **海量定时器怎么高效** → 时间轮（第 4 章），O(1) 增删。但那是 Kafka/Netty 的量级，业务 cron 的几十个任务用最小堆足矣。
4. **多节点协调靠什么** → 本质是共识（第 6 章）：不重复 = 共识 + fencing + 幂等。Redis 锁缺后两者，只能 best-effort；etcd/ZK/PG 把共识做对了才「天然多副本」。
5. **什么时候上重型方案** → 先分清痛点在哪根轴（第 8 章）。可靠性问题上 River，纯吞吐瓶颈才上分布式引擎，别把可靠性需求误当规模需求。

**一条主线收敛全文**：任务调度是**两根正交轴**——可靠性（可丢→必达）和吞吐/规模（单点→海量）。绝大多数项目只在可靠性轴上走，终点是持久队列；吞吐轴的分布式引擎是岔路而非顶点。**先在地图上找到自己的位置，再选工具**，别被「新库更全」「大厂都用引擎」带着往上爬。

而三个反复出现的误解，一并钉死：**时间轮**是单机性能优化（≠多副本）、**一致性哈希**是分片手段（≠数据一致、≠不重复）、**「只跑一个」**是共识问题（≠加把锁）。看穿这三点，多数调度选型的迷雾就散了。

## 参考

**Go 调度库 / 队列**
- [go-co-op/gocron](https://github.com/go-co-op/gocron)、[gocron-redis-lock](https://github.com/go-co-op/gocron-redis-lock)
- [robfig/cron](https://github.com/robfig/cron)、[netresearch/go-cron](https://github.com/netresearch/go-cron)（robfig 活跃 fork）
- [River（Postgres 队列）](https://riverqueue.com/)、[asynq（Redis 队列）](https://github.com/hibiken/asynq)

**时间轮**
- [Kafka Purgatory / 分层时间轮](https://cwiki.apache.org/confluence/display/KAFKA/Purgatory)
- [Netty HashedWheelTimer](https://netty.io/4.1/api/io/netty/util/HashedWheelTimer.html)
- [Hashed and Hierarchical Timing Wheels（Varghese & Lauck）](https://www.cs.columbia.edu/~nahum/w6998/papers/ton97-timing-wheels.pdf)
- [RussellLuo/timingwheel（Go 实现）](https://github.com/RussellLuo/timingwheel)

**共识与分布式锁**
- [Raft 论文](https://raft.github.io/raft.pdf)、[etcd — Why etcd](https://etcd.io/docs/latest/learning/why/)
- [ZooKeeper recipes（选主、锁）](https://zookeeper.apache.org/doc/current/recipes.html)
- [Martin Kleppmann — How to do distributed locking](https://martin.kleppmann.com/2016/02/08/how-to-do-distributed-locking.html)

**一致性哈希**
- [Consistent Hashing 原始论文（Karger et al., 1997）](https://www.cs.princeton.edu/courses/archive/fall09/cos518/papers/chash.pdf)

**分布式调度引擎 / 工作流**
- [XXL-Job](https://github.com/xuxueli/xxl-job)、[Apache ShardingSphere ElasticJob](https://github.com/apache/shardingsphere-elasticjob)
- [Temporal](https://temporal.io/)、[Cadence](https://github.com/cadence-workflow/cadence)
- [Kubernetes CronJob](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/)

---

*系列教程：[saas-backend](../saas-backend/README.md)*
