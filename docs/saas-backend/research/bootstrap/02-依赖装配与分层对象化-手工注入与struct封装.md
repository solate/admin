# 依赖装配与分层对象化：手工注入与 struct 封装

> 2026-07。本文合并回答两个纠缠在一起的问题——**(1) 装配：新项目要不要引 DI 框架（wire/fx）来消除手工接线的样板？(2) 分层：handler/service/repo 三层有必要都做成 struct（对象）吗，包级纯方法不够用吗？** 老项目里有个「需要加锁」的模块，结果整块退化成了包级纯函数，这个「加锁 → 只能纯方法」的困境，正是把这两个问题搅在一起造成的。
>
> 结论先行（更简 + 行业共识优先）：**手工构造注入，零 DI 框架**——这本就是 Go 社区对 greenfield 的成熟共识首选（[rednafi](https://rednafi.com/go/di-frameworks-bleh/)、[ehewen](https://ehewen.com/en/blog/go-dependency-injection/) 的 "Manual Injection Recommended First Choice"），不是本项目独有偏好。装配**从平铺起步**，量长大了再抽「按域 module 函数 + deps 聚合 struct」——**是重构手段，不是 day-1 就搭的框架**。分层上**有依赖或有共享可变状态 → struct；纯计算无状态 → 包级函数**，三层都持依赖故一律 struct。而「加锁逼出纯方法」是**误判**：共享锁 = **共享同一个被注入的实例**，不等于「必须包级全局」。
>
> 关联：生命周期编排见 [01](./01-应用启动与组件生命周期-main组合根与gin-cron编排.md)，后台任务见 [cron 系列](../cron/02-单机执行模型-从ticker到持久队列的四档.md)；落地约束参照 [domain-architecture.md](../../../../.claude/rules/domain-architecture.md)、[repo-transaction-convention.md](../../../../.claude/rules/repo-transaction-convention.md)、[service-patterns.md](../../../../.claude/rules/service-patterns.md)、[rbac-multi-tenant.md](../../../../.claude/rules/rbac-multi-tenant.md)。
>
> 日常写代码不必读；想搞明白「为什么不上 wire/fx」「有共享状态的模块要不要破例写成包级函数」时看这里。

## 结论速览

装配与分层是同一件事的两面——**都在回答「依赖怎么进到代码里」**。一张表说清两个问题的取舍：

| 问题 | 更简的对（本项目选） | 更重/更差的错 |
|------|---------------------|--------------|
| 装配用什么 | 手工构造注入（普通 Go 代码，编译期安全） | wire（codegen 间接层）/ fx（运行期反射容器） |
| 装配怎么起步 | 平铺，痛了再抽 module 函数 | day-1 预造 `buildModules` 框架 |
| 三层用什么形态 | struct（持依赖，构造注入） | 包级纯方法 + `InitXxx()` 塞全局 var |
| 共享并发状态 | 注入**同一个实例**的指针 | 包级 `var` + `sync.Once`「碰巧」共享 |

**两个问题的共同答案是「显式注入实例」**：装配是「把依赖显式传进构造函数」，分层是「用 struct 承接这些依赖」，共享并发状态是「注入同一个实例而非包级全局」。DI 框架、包级函数、包级全局，都是在不同位置放弃了「显式」。

被刻意否掉的方向：wire（codegen 边际收益不划算）、fx（运行期容器 + 全局注册表，丢编译期安全）、service locator / 包级全局（最差，隐式依赖运行期才炸）、「有并发状态就该用包级函数」这个前提本身。

## 一、装配：手工注入 vs wire vs fx

老项目 B 的 `internal/router/app.go` 是全项目最大的单文件（582 行），痛点集中在三处，但**没一个是「缺 DI 框架」造成的**：

> **反面标本（老项目 B，一句带过）**：① `Handlers` 是 40 字段的扁平巨结构；② `initHandlers` 是 165 行手工接线，55 个 repo + 35 个 service + 40 个 handler 全拍平在一个方法里，谁依赖谁靠注释里的 "must init before X"；③ 某设备服务构造器单行 28 个位置参数，传错一个位置编译过、运行错。三个痛点的共性是**装配逻辑没有边界**——是「缺分域组织」，不是「缺框架」。

三方对比，看清楚框架解决的到底是不是这些痛：

| 维度 | 手工注入（本项目选） | google/wire | uber/fx |
|------|-------------------|-------------|---------|
| 注入时机 | 编译期（普通 Go 代码） | 编译期（codegen） | **运行期**（反射 + DAG） |
| 编译期安全 | ✅ 全是普通函数调用 | ✅ 生成的也是普通代码 | ❌ 缺依赖启动才 panic |
| 启动可读性 | ✅ 装配即代码，点进去就看到 | ⚠️ 要看生成的 `wire_gen.go` | ❌ 靠运行时拼 DAG |
| 学习/魔法成本 | ✅ 零，会写 Go 就会 | ⚠️ ProviderSet/绑定规则 | ❌ 高：生命周期钩子/参数对象/模块 |
| 可测试性 | ✅ 直接 `newXxxModule(fakeDeps)` | ✅ 手工替换 provider | ⚠️ 要 `fxtest` |
| 排错难度 | ✅ 栈直达 | ⚠️ 栈经过生成代码 | ❌ DAG 错误信息晦涩 |
| 可移植性 | ✅ 无依赖 | ⚠️ 需 `wire` 工具链 | ⚠️ 运行期强依赖 fx |

**「手工注入是 greenfield 首选」是行业共识，不是本项目偏好。** 多篇成熟讨论都把它列为默认起点、DI 框架列为「对象图大到手工重复才上」的后备：

- [rednafi《You probably don't need a DI framework》](https://rednafi.com/go/di-frameworks-bleh/)：DI 作为**技术**很有用，但 DI **框架**在 Go 里往往得不偿失——普通构造函数已把依赖显式化了。
- [ehewen《DI in Go》](https://ehewen.com/en/blog/go-dependency-injection/)：明确把 Manual Injection 标为 "Recommended First Choice"。
- [leapcell《Wire vs fx vs Manual》](https://leapcell.io/blog/go-dependency-injection-approaches-wire-vs-fx-and-manual-best-practices)：系统对比后把「plain manual」列为常被低估但最透明的方案。
- [softwarepatternslexicon](https://softwarepatternslexicon.com/go/modern-design-patterns-in-go/dependency-injection/)：Go 的 DI 通常始于普通构造函数 + 窄接口；容器**不该隐藏启动错误与生命周期归属**。

**各自的「何时选它」**：手工注入——组件数量可控（本项目 12 域）、重视「装配即代码、无魔法」；**wire**——依赖图很深很宽、手工接线确实开始出错、团队接受 codegen 工作流（典型 50+ provider）；**fx**——需要按 feature flag **动态**拼装模块、需框架级 `OnStart/OnStop` 钩子（典型插件化大型微服务）。判据高度一致：**对象图手工写不再痛时才上库**。本项目 12 域、分域后每处都短，远未到临界点。

## 二、装配怎么组织：平铺起步，长大了再抽

手工注入不代表「一上来就搭装配框架」。**最简的正确形态是平铺**：在组合根（`main`/`server`）里顺着 repo → service → handler 建下来，一条链看得清清楚楚。

```go
// 最简平铺：域少时这就够了，不需要任何"装配框架"
func buildHandlers(deps Deps) *Handlers {
    userRepo := repository.NewUserRepo(deps.DB)
    roleRepo := repository.NewRoleRepo(deps.DB)
    userSvc := user.NewService(deps.DB, userRepo, roleRepo)
    roleSvc := role.NewService(deps.DB, roleRepo)
    return &Handlers{
        User: user.NewHandler(userSvc),
        Role: role.NewHandler(roleSvc),
    }
}
```

**只有当平铺开始痛**（一个函数几十行、跨域依赖靠位置记忆、加一个域要在多处改），才抽两样东西——注意这是**重构手段，不是 day-1 就搭的框架**：

**(a) 按域 module 函数**：每个域一个 `newXxxModule`，域内 repo→service→handler 链路自己收敛：

```go
func newUserModule(deps Deps) *user.Handler {
    userRepo := repository.NewUserRepo(deps.DB)
    roleRepo := repository.NewRoleRepo(deps.DB)
    svc := user.NewService(deps.DB, userRepo, roleRepo)   // 按域分子包，见 domain-architecture
    return user.NewHandler(svc)
}
// buildHandlers 里就变成一域一行：User: newUserModule(deps)
```

**(b) deps 聚合结构体**，干掉「28 个位置参数」这种 mega-constructor：

```go
// 只放"过半数域都要"的公共依赖；本域特有的单独作参数传
type Deps struct {
    DB    *gorm.DB
    RDB   *redis.Client
    Log   *slog.Logger
    Audit *audit.Recorder
    JWT   *jwt.Manager
}
// 服务构造器：func NewDeviceService(deps Deps, cfg DeviceConfig) *DeviceService
```

位置参数长列表的两个致命伤——传错位置编译器不报错、加参数要改所有调用点——聚合成 struct 后都消失。

**关键是时序**：day-1 平铺 → 痛了抽 module 函数 → 公共依赖多了抽 deps struct。**不要反过来先造框架**——那是老项目 B「先有 40 字段 Handlers 再往里塞」的另一种翻版，只是把「太乱」换成了「太重」。

### 与 repo/事务约定的关系

装配顺序天然是 **repo → service → handler**（[domain-architecture.md](../../../../.claude/rules/domain-architecture.md) 的依赖链）。**事务不在装配期处理**——[repo-transaction-convention.md](../../../../.claude/rules/repo-transaction-convention.md) 明确「事务在 service 层用 `s.db.Transaction` + 闭包内 `NewXxxRepo(tx)` 重建」。所以装配期只把 base repo（吃 `deps.DB`）注入 service，事务版 repo 在运行期由 service 自己重建。装配层不碰事务，职责干净。

### AI 协作场景的调整：约定前置，别等"痛了再抽"

上面"day-1 平铺 → 痛了再抽"是 **greenfield 探索期**的路径——它依赖一个隐含前提：**有人能在混乱长出来时整体识别"该抽什么、按什么维度抽"并重构**。人类做得到，但 **AI 增量写代码做不到**：AI 靠规则与既有约定的一致性逐次生成，没有"等乱到一定程度我自己回头识别该抽哪一刀"的全局重构能力。若架构约定不前置，AI 只会写出"当下能跑"的代码，方向逐次发散；等规模长大再让 AI 重构，它已无法还原"这 50 个 handler 本该按什么逻辑分组"。

**关键认知：本项目已过探索期。** 架构维度已经定死在 [domain-architecture.md](../../../../.claude/rules/domain-architecture.md)——12 个业务域、按域分子包（`internal/handler/{domain}/`、`internal/service/{domain}/`）、依赖链 `router → handler/{domain} → service/{domain} → repository`、装配在 `internal/router/app.go` 的 `initHandlers`。这套约定就是那"再抽"的结果，已经落定，**不必也不该让 AI 从平铺重走一遍**。

所以 AI 增量写代码时，**从第一行起就按约定写**，不存在"先平铺、等以后再抽"：

- 新功能先判断归属——是新域还是现有域扩展；新域直接建 `internal/handler/{domain}/` + `internal/service/{domain}/` 子包，**一步到位是分域结构**；
- 构造函数遵守"只接收本域实际需要的依赖"（[domain-architecture.md](../../../../.claude/rules/domain-architecture.md) 规则 3），不塞用不到的东西；
- 装配逻辑写进 `initHandlers`，一域一段，不建全局变量、不留"待抽"的临时平铺。

**分工**：本文档（`research/bootstrap/`）是**给人看的论证**——为什么选手工注入、为什么否掉 wire/fx、tradeoff 在哪；`.claude/rules/` 是**给 AI 看的执行合约**——具体建哪个包、构造函数怎么签名、装配写在哪。人做架构决策时读本文，AI 增量写代码时严格照 rules 执行。"先平铺后重构"是探索期给人的方法论；**约定既已定型，就以约定为 day-1 起点**，这是 AI 友好度对本节的唯一修正。

## 三、分层：为什么三层一律 struct，而非纯方法

第二个问题——「handler/service/repo 有必要都做成 struct 吗，包级纯方法不够吗？」——常被一个误判绑架：**「有个模块要加锁 → 整层只能写成纯方法」**。这是把两件独立的事焊死了。拆开看：

| 维度 | struct（本项目选） | 包级纯方法 |
|------|------------------|-----------|
| 依赖注入 | ✅ 构造函数注入，依赖显式 | ❌ 只能靠包级 `var` + `InitXxx()` 塞 |
| 可测试性 | ✅ `NewXxx(fakeDeps)` 直接造实例测 | ❌ 包级全局，测试间互相污染 |
| 无全局状态 | ✅ 无包级可变 `var` | ❌ 天然要包级 `var` 存依赖 |
| 并行测试 | ✅ 各实例独立 | ❌ 共享全局，`t.Parallel()` 打架 |
| 共享并发状态 | ✅ 锁/信号量作字段，共享同一实例 | ⚠️ 靠包级 `var` +`sync.Once`「碰巧」共享 |

三层（handler/service/repo）都持依赖，天然属「有依赖」的一类，所以一律 struct。这与第一节「手工注入」同源——**struct 就是「把依赖显式收进对象」的载体**。真正可以用包级函数的，只有**无状态纯工具**（`pkg/utils/convert`、`idgen` 这类无依赖、无共享可变状态的函数）——它们没有依赖要注入、没有状态要隔离，套 struct 反而过度设计。

**判据一句话：有依赖或有共享可变状态 → struct；纯计算无状态 → 包级函数。**

## 四、认知纠偏：共享锁 = 共享实例，不是包级全局

那个「要加锁所以退化成包级函数」的模块，病根不在锁，在**跳过了「把同一个实例传给两个调用点」这段接线**。

> **反面标本（老项目 B，一句带过）**：某重型后台模块有两条触发路径（请求回调 + cron 重试）要共享同一个信号量 + 同一个去重集。作者用**包级 `var procSemaphore chan struct{}` + `var procInFlight sync.Map` + `sync.Once`** 让两条路径「自动」共享，`StartProc` 于是只能是纯函数（要够到那堆包级 var）。图省事，代价是包级可变全局、挡并行测试、`InitProc` 参数越滚越长。注意：真正做互斥的是 `chan`+`sync.Map`，**不是 `sync.Mutex`**——作者记忆里的「锁」其实是「进程级共享的并发状态」。

把那坨包级 `var` 原样搬进 struct，锁语义**一个字都不变**：

```go
type Processor struct {
    sem      chan struct{}   // 进程级并发上限
    inFlight sync.Map        // 资源级去重
    repo     *repository.XxxRepo
}

func NewProcessor(deps Deps, cfg XxxConfig) *Processor {
    return &Processor{sem: make(chan struct{}, cfg.MaxConcurrency), repo: repository.NewXxxRepo(deps.DB)}
}

func (p *Processor) Start(ctx context.Context, resourceID string) {
    if _, loaded := p.inFlight.LoadOrStore(resourceID, struct{}{}); loaded {
        return
    }
    defer p.inFlight.Delete(resourceID)
    p.sem <- struct{}{}
    defer func() { <-p.sem }()
    // ... 用 p.repo 干活
}
```

进程级互斥的本质是**「所有调用者操作同一个 `sem`/`inFlight`」**。包级全局用「同一个包变量」达成；struct 用「**同一个实例的指针**」达成——对锁而言完全等价。区别只在：

- 包级全局：谁 import 谁碰到同一份状态（隐式、不可控、不可测）；
- 共享实例：装配期 `NewProcessor(...)` **只建一次**，把这**同一个 `*Processor`** 显式注入每个调用点（显式、可控、可测）。

`sync.Once` 想保证的「只初始化一次」，被「装配期只 `New` 一次」天然替代。两条触发路径怎么共享？就是普通依赖注入——同一个 `*Processor` 分别给两边：

```go
proc := NewProcessor(deps, cfg)          // 装配期建一次
uploadHandler := NewUploadHandler(proc)  // ① 请求路径
scheduler.RegisterRetry(proc)            // ② cron 重试路径（见 03）
```

两边握同一个指针，命中同一个 `sem`/`inFlight`——与包级全局逐字节等效，但没有一个包级 `var`。

> **更进一步**：若后台任务改用 [cron 系列](../cron/02-单机执行模型-从ticker到持久队列的四档.md) 的 **gocron v2 / River**，「多副本只跑一个、并发上限、任务去重」都是**库内建能力**（gocron 的 Locker/Singleton/`LimitConcurrentJobs`、River 的 unique jobs / worker concurrency）——**连手写这个 `Processor` struct 都可能不需要**，任务直接注册给 scheduler/queue。这正是老项目手搓那套与成熟方案差距最大处。

## 五、正例对照：老项目 A 的 PermissionCache 早就这么做了

不用去别处找范例——老项目 A 自己就有一个「**有锁 + 有后台刷新 goroutine + 仍是干净 struct + 作为依赖注入**」的现成组件：`PermissionCache`（见 [rbac-multi-tenant.md](../../../../.claude/rules/rbac-multi-tenant.md)）。

```go
type PermissionCache struct {
    mu         sync.RWMutex               // 锁是字段
    apiPerms   map[string][]string        // 受 mu 保护的共享状态
    // ...
}

func (c *PermissionCache) Get(...) { c.mu.RLock(); defer c.mu.RUnlock(); /* ... */ }
func (c *PermissionCache) watchRefresh() { /* 后台 goroutine，写时 c.mu.Lock() */ }
```

它同时具备本文讨论的全部「困难特征」——共享可变状态、读写锁、后台常驻 goroutine、需要进程内唯一——却**依然是一个 struct，构造后作为依赖注入**（JWT 中间件、role/menu service 都持同一个 `*PermissionCache`）。这直接证伪「加锁 → 只能纯方法」：同团队、同约束下，`PermissionCache` 用 struct 把这些全解决了。那几个后台模块退化成包级全局，纯粹是当时图省事，不是锁的必然。

## 六、成本与 tradeoff（已知并接受）

- **手工装配的样板不会归零**：域多了要手写 `newXxxModule`、加域手动加一行。接受——换来「装配即代码、点进去就懂、编译期全保障、零框架依赖」。但**不预付**：day-1 平铺，痛了再抽。
- **deps 聚合 struct 的粒度要把握**：塞太多变成新 god-struct，塞太少退回长参数列表。约定：**只放「过半数域都要」的公共依赖**（DB/RDB/Log/Audit/JWT），本域特有的单独作参数传。
- **多写一次「把单例接到两个调用点」的接线**：这正是老项目图省事跳过的那步。多几行注入，换可测 + 无全局 + 依赖显式，值。
- **进程级 vs 多副本**：struct 里的信号量只在**单进程内**限流。多副本要靠拆调度器或成熟库内建的协调（见 [cron/05 从单机到多节点](../cron/05-从单机到多节点-重复跑与拆调度器.md)）——分布式的固有成本，手写 struct 信号量不解决它。
