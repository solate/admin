# 分层对象化与共享并发状态：为什么三层一律 struct，而非纯方法

> 2026-07。本文回答一个实际问题——**handler/service/repo 三层有必要都做成对象（struct）吗，还是包级纯方法就够？老项目里有个「需要加锁」的模块，结果整块退化成了包级纯函数，这种「加锁 → 只能纯方法」的困境怎么破？**
>
> 结论先行（Go 社区封装共识）：**有依赖或有共享可变状态的组件 → struct；纯计算无状态 → 包级函数**。handler/service/repo 三层天然属前者（都持依赖），所以一律 struct。这与「构造注入是 Go DI 的成熟默认」（见 [02](./02-依赖装配与组件注册-手工注入vs-wire-fx.md)，[rednafi: you probably don't need a DI framework](https://rednafi.com/go/di-frameworks-bleh/)、[glukhov: DI patterns & best practices](https://glukhov.org/app-architecture/code-architecture/dependency-injection-in-go/)）同源——struct 是「把依赖显式收进对象」的载体。而「加锁逼出纯方法」是**误判**：两个老项目里所有 `Mutex` 其实都是 struct 字段，没有一个裸包级锁。唯一真正需要「一个进程一份」的，是那个**并发限制器实例本身**——把它做成一个**被注入的 struct**，同一实例的指针同时交给请求路径和 cron 路径即可，不必把整层降级成包级纯函数。（更进一步：若并发限制交给 gocron v2 / River 这类成熟库，连这个 struct 都可能不用自己写，见第四节末。）
>
> 关联：生命周期 [01](./01-应用启动与组件生命周期-main组合根与gin-cron编排.md)、装配 [02](./02-依赖装配与组件注册-手工注入vs-wire-fx.md)、cron [03](./03-cron任务注册与复用-jobs装配与Server生命周期集成.md)；落地约束参照 [repo-transaction-convention.md](../../../../.claude/rules/repo-transaction-convention.md)、[domain-architecture.md](../../../../.claude/rules/domain-architecture.md)、[service-patterns.md](../../../../.claude/rules/service-patterns.md)。
>
> 日常写代码不必读；纠结「这个有共享状态的模块要不要破例写成包级函数」时看这里。

## 结论速览

「有个要加锁的模块 → 整层只能写成纯方法」是把两件独立的事绑死了。拆开看：

| 维度 | struct（对象，本项目选） | 包级纯方法 |
|------|------------------------|-----------|
| 依赖注入 | ✅ 构造函数注入，依赖显式 | ❌ 依赖只能靠包级 `var` + `InitXxx()` 塞 |
| 可测试性 | ✅ `NewXxx(fakeDeps)` 直接造实例测 | ❌ 包级全局，测试间互相污染 |
| 零全局状态 | ✅ 无包级可变 `var`（Go 社区共识：避免全局可变状态） | ❌ 天然要包级 `var` 存依赖/状态 |
| 事务约定 | ✅ receiver 持 `s.db`，`s.db.Transaction`+重建 repo（本项目落地形态见 [repo-transaction](../../../../.claude/rules/repo-transaction-convention.md)） | ❌ 无 receiver，`db` 只能来自包级 `var` |
| 并行测试 | ✅ 各实例独立 | ❌ 共享全局，`t.Parallel()` 打架 |
| 共享并发状态 | ✅ 锁/信号量作字段，**共享同一实例**即可 | ⚠️ 靠包级 `var` + `sync.Once`「碰巧」共享 |

**关键认知：「共享锁」= 「共享同一个 struct 实例」，不等于「必须包级全局」。** 锁作为 struct 字段，装配期只 `New` 一次，把同一个 `*T` 指针注入所有调用点（请求 handler + cron），就得到完全相同的进程级互斥——还顺带拿回上面那一整列 ✅。

被刻意否掉的方向：包级 `var` + `sync.Once` 做「进程单例」（用「单实例 + 共享引用」替代）、以及「有并发状态就该用包级函数」这个前提本身。

## 一、反面标本：那个「要加锁」的模块为什么退化成包级全局

老项目 B 的 40 个 service 里，**37 个是规规矩矩的 struct**，只有 3 个重型后台模块（并发受限的资源处理任务）破了例。它们的骨架长这样（脱敏后）：

```go
// 包级全局并发控制（进程级共享，请求触发 + cron 重试共用）——反面
var (
    procSemaphore chan struct{}   // 进程级并发上限
    procInFlight  sync.Map        // 资源 ID 级去重（作者记忆里的「锁」）
    procOnce      sync.Once
)

// 包级全局依赖，由 InitProc 初始化——反面
var (
    procDB   *gorm.DB
    procRepo *repository.XxxRepo
    procCfg  *config.XxxConfig
)

func InitProc(db *gorm.DB, cfg *config.XxxConfig) {
    procOnce.Do(func() {
        procDB = db
        procRepo = repository.NewXxxRepo(db)
        procSemaphore = make(chan struct{}, cfg.MaxConcurrency)
    })
}

// 纯函数，不是方法——因为它要够到上面那堆包级 var
func StartProc(ctx context.Context, resourceID string) {
    if _, loaded := procInFlight.LoadOrStore(resourceID, struct{}{}); loaded {
        return   // 去重：同一资源正在处理则跳过
    }
    defer procInFlight.Delete(resourceID)
    procSemaphore <- struct{}{}            // 阻塞获取信号量
    defer func() { <-procSemaphore }()
    // ... 用 procRepo / procDB 干活
}
```

为什么会走到这一步？因为这个模块有**两条独立触发路径，必须共享同一个信号量 + 同一个去重集**：

- **请求/事件路径**：上游回调触发 `StartProc(...)`；
- **cron 重试路径**：`internal/jobs/*_retry_worker.go` 里崩溃兜底，也调同一个 `StartProc(...)`（呼应 [03](./03-cron任务注册与复用-jobs装配与Server生命周期集成.md) 的 retry/backfill 兜底）。

两条路径调**同一个包级函数**，于是「自动」共享了同一个 `procSemaphore` 和 `procInFlight`——作者用包级全局 + `sync.Once`，**跳过了「把同一个实例显式传给两个调用点」这段接线**。图的是省事，代价是三样：

1. **包级可变全局**：`var procSemaphore` / `var procRepo` 直接违背 [README #10 零全局状态](../../README.md)。
2. **挡并行测试**：包级状态被所有测试共享，`sync.Once` 还让「重新初始化一个干净实例」变得不可能。
3. **依赖来路发散**：`InitProc` 的参数列表随功能增长越来越长，正是 [03 第二节](./03-cron任务注册与复用-jobs装配与Server生命周期集成.md) 说的「每次要初始化一堆东西」在 service 层的翻版。

**注意**：真正做互斥的是 `chan struct{}` 信号量 + `sync.Map`，不是 `sync.Mutex`。作者记忆里的「锁」其实是「进程级共享的并发状态」。这一点很关键——它根本不要求包级函数，下一节说明。

## 二、认知纠偏：共享锁 = 共享实例，不是包级全局

把上面那坨包级 `var` 原样搬进一个 struct，`StartProc` 从纯函数变成方法，锁的语义**一个字都不变**：

```go
// 正面：并发状态封装进 struct，锁/信号量是字段
type Processor struct {
    sem      chan struct{}   // 进程级并发上限
    inFlight sync.Map        // 资源级去重
    repo     *repository.XxxRepo
    db       *gorm.DB
}

func NewProcessor(deps ModuleDeps, cfg XxxConfig) *Processor {
    return &Processor{
        sem:  make(chan struct{}, cfg.MaxConcurrency),
        repo: repository.NewXxxRepo(deps.DB),
        db:   deps.DB,
    }
}

// 方法，不是纯函数
func (p *Processor) Start(ctx context.Context, resourceID string) {
    if _, loaded := p.inFlight.LoadOrStore(resourceID, struct{}{}); loaded {
        return
    }
    defer p.inFlight.Delete(resourceID)
    p.sem <- struct{}{}
    defer func() { <-p.sem }()
    // ... 用 p.repo / p.db 干活
}
```

进程级互斥的本质是**「所有调用者操作的是同一个 `sem` / `inFlight`」**。包级全局用「同一个包变量」达成这一点；struct 用「**同一个实例的指针**」达成——两者对锁而言完全等价。区别只在：

- 包级全局：谁 import 这个包，谁就碰到同一份状态（隐式、不可控、不可测）；
- 共享实例：装配期 `NewProcessor(...)` **只建一次**，把这**同一个 `*Processor`** 显式注入需要它的每个调用点（显式、可控、可测）。

所以「这个模块需要进程内唯一的并发限制器」这个**真需求**，用「**单个实例 + 共享引用**」满足，而不是「包级 `var` + `sync.Once`」。`sync.Once` 想保证的「只初始化一次」，被「装配期只 `New` 一次」天然替代了（[02](./02-依赖装配与组件注册-手工注入vs-wire-fx.md) 的按域装配就是「每个组件建一次」）。

那「两条触发路径怎么共享」呢？就是普通的依赖注入——同一个 `*Processor` 分别给两边：

```go
// 装配期建一次
proc := NewProcessor(deps, cfg)

// ① 注入请求 handler（事件触发路径）
uploadHandler := NewUploadHandler(proc)

// ② 注入 jobs.Deps（cron 重试路径，见 03）
jobs.Register(cronRunner, jobs.Deps{ Processor: proc, /* ... */ })
```

两边握着同一个指针，`proc.Start(...)` 命中同一个 `sem`/`inFlight`——和包级全局的效果逐字节相同，但没有一个包级 `var`。

> **更进一步**：若后台任务改用 [03](./03-cron任务注册与复用-jobs装配与Server生命周期集成.md) 推荐的 **gocron v2 / River**，这些并发控制（多副本只跑一个、并发上限、任务级去重）都是**库内建能力**（gocron 的 Locker/Singleton/`LimitConcurrentJobs`、River 的 unique jobs / worker concurrency）——**连手写这个 `Processor` struct 都可能不需要了**，任务直接注册给 scheduler/queue，依赖注入的只是 service 方法。这正是老项目手搓那套与成熟方案差距最大的地方。

## 三、正例对照：老项目 A 的 PermissionCache 早就这么做了

不用去别处找范例——老项目 A 自己就有一个「**有锁 + 有后台刷新 goroutine + 仍然是干净 struct + 作为依赖注入**」的现成组件：`PermissionCache`（见 [rbac-multi-tenant.md](../../../../.claude/rules/rbac-multi-tenant.md)）。

```go
type PermissionCache struct {
    mu         sync.RWMutex               // 锁是字段
    apiPerms   map[string][]string        // 受 mu 保护的共享状态
    menuPerms  map[string][]string
    // ...
}

func (c *PermissionCache) Get(...) { c.mu.RLock(); defer c.mu.RUnlock(); /* ... */ }
func (c *PermissionCache) watchRefresh() { /* 后台 goroutine 定时/事件刷新，写时加 c.mu.Lock() */ }
```

它同时具备本文讨论的全部「困难特征」——共享可变状态、读写锁、后台常驻 goroutine、需要进程内唯一——却**依然是一个 struct，构造后作为依赖注入**（JWT 中间件、role/menu service 都持有同一个 `*PermissionCache`）。这直接证伪了「加锁 → 只能纯方法」：同一个团队、同一批约束下，`PermissionCache` 用 struct 把这些全解决了。那 3 个后台模块退化成包级全局，纯粹是当时图省事，不是锁的必然。

## 四、落地形态：装配一次，注入两处（仅骨架）

顺着 [02](./02-依赖装配与组件注册-手工注入vs-wire-fx.md) 的按域装配 + [03](./03-cron任务注册与复用-jobs装配与Server生命周期集成.md) 的 cron 复用 service，共享并发组件的装配就是「多建一个实例、多注入一处」：

```go
// internal/server 装配期（示意，Step 05+ 落地）
func buildModules(deps ModuleDeps) (*moduleHandlers, jobs.Deps) {
    proc := NewProcessor(deps, deps.Cfg.Xxx)   // ★ 进程内唯一，只此一次

    handlers := &moduleHandlers{
        Upload: newUploadModule(deps, proc),   // 事件触发路径拿到 proc
        // ...
    }
    jobDeps := jobs.Deps{
        Processor: proc,                        // cron 重试路径拿到同一个 proc
        // ...
    }
    return handlers, jobDeps
}
```

要点：`proc` 在装配根建一次，谁需要就把这个指针传给谁。没有 `InitProc()`、没有 `sync.Once`、没有包级 `var`。加一个这类组件 = 加一行 `New` + 在用到的调用点多传一个参数。

## 五、成本与 tradeoff（已知并接受）

- **要多写一次「把单例接到两个调用点」的接线**：这正是老项目图省事跳过的那一步。多几行注入代码，换来可测 + 无全局 + 依赖显式，值。这也是 [02 结论](./02-依赖装配与组件注册-手工注入vs-wire-fx.md)「手工装配的样板不会归零，但每处都短」的同一笔账。
- **进程级 vs 多副本**：struct 里的信号量只在**单进程内**限流。多副本部署时「全局并发上限」「跨副本去重」要靠成熟库内建的多副本协调（gocron Locker/Elector、River/asynq 的 DB/Redis 抢占，见 [03](./03-cron任务注册与复用-jobs装配与Server生命周期集成.md) 第四节）——这是分布式的固有成本，手写 struct 信号量不解决它，成熟库才解决。
- **什么时候包级函数才真的可以**：**无状态**的纯工具（`pkg/utils/convert`、`pkg/utils/idgen` 这类无依赖、无共享可变状态的函数）本来就该是包级函数——它们没有依赖要注入、没有状态要隔离，套 struct 反而是过度设计。本文的「一律 struct」只针对 **handler/service/repo 这三层有依赖、有状态的业务组件**。判据一句话：**有依赖或有共享可变状态 → struct；纯计算无状态 → 包级函数。**
