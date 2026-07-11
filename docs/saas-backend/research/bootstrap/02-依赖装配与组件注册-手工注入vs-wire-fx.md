# 依赖装配与组件注册：手工注入 vs wire vs fx

> 2026-07。本文回答一个实际问题——**老项目 B 的 `App.initHandlers()` 是 165 行扁平手工接线，`Handlers` 结构体 40 个字段，某设备服务的构造器单行 28 个位置参数，每加一个 handler 要改三处（struct 字段、构造调用、路由注册）。新项目要不要引入 DI 框架（wire/fx）来消除样板？**
>
> 结论先行（行业共识优先）：**手工构造注入 + 零 DI 框架**，用两招消除老项目样板痛点——**(a) 按域分组的装配函数**（每个域一个 `newXxxModule(deps) → moduleHandlers`，替代扁平 `initHandlers`）；**(b) deps 聚合结构体** 替代 mega-constructor。这**不是本项目独有偏好，而是 Go 社区 greenfield 的成熟共识首选**（[rednafi](https://rednafi.com/go/di-frameworks-bleh/)、[ehewen](https://ehewen.com/en/blog/go-dependency-injection/)、[leapcell](https://leapcell.io/blog/go-dependency-injection-approaches-wire-vs-fx-and-manual-best-practices) 均推 manual injection first）。wire/fx 都留作升级路径：wire 在依赖图 ≥50+ provider 时；fx 在需动态模块拼装或插件化时。本项目 12 域、手工接线量已被分域压到很低，未踩升级触发点。
>
> 关联：生命周期编排见 [01](./01-应用启动与组件生命周期-main组合根与gin-cron编排.md)，cron 依赖复用见 [03](./03-cron任务注册与复用-jobs装配与Server生命周期集成.md)；落地约束参照 [domain-architecture.md](../../../../.claude/rules/domain-architecture.md)、[repo-transaction-convention.md](../../../../.claude/rules/repo-transaction-convention.md)、[service-patterns.md](../../../../.claude/rules/service-patterns.md)。
>
> 日常写代码不必读；想搞明白「为什么不上 wire/fx」时看这里。

## 结论速览

装配样板的痛源不是「没有 DI 框架」，而是**组织方式不对**：

| 痛点 | 老项目 B 做法（反面） | 本项目推荐（手工注入优化版） |
|------|---------------------|------------------------|
| 扁平巨结构 | `Handlers` 40 字段一坨，`initHandlers` 165 行一坨 | **按域分组**：12 个域 × 每域一个 `newXxxModule()` |
| mega-constructor | 某服务构造器 28 个位置参数 | **deps 聚合结构体**：`ModuleDeps{DB, RDB, Log, Audit}` 一把传，服务特有参数再单独传 |
| 重复装配 | cron jobs 在 `jobs.InitXxx()` 里**重建 repo/service 塞包级全局**（双真相源） | cron 复用装配好的 service（见 [03](./03-cron任务注册与复用-jobs装配与Server生命周期集成.md)） |
| 路由注册分散 | handler 定义、装配、路由注册三处分离 | handler 定义 + 路由注册在域子包内聚，装配在模块函数 |

**wire/fx 解决不了这三个痛**——它们解决的是「构造函数参数顺序搞错」「循环依赖检测」，而上面四个痛全是「没按域分组、deps 没聚合、重复装配」的组织问题，换框架一个都解决不了，反而多引魔法 + 调试成本。

被刻意否掉的三个方向：wire（codegen 带来的间接层 vs 省下的手工接线，边际收益不划算）、fx（运行期反射 + 全局容器，与「构造函数即依赖契约、编译期可查」的显式装配取向冲突）、service locator（最差，隐式依赖 + 运行期炸）。

## 一、老项目 B 的装配痛点（反面标本）

老项目 B 的 `internal/router/app.go` 是全项目最大的单文件（582 行），痛点集中在三处：

```go
// ① Handlers：40 个字段的扁平巨结构，加一个域就多一行
type Handlers struct {
    HealthHandler, CaptchaHandler, AuthHandler, UserHandler,
    TenantHandler, RoleHandler, MenuHandler, /* ... 共 40 个 */
}

// ② initHandlers：165 行手工接线，repo → service → handler 全拍平在一个方法里
func (s *App) initHandlers() error {
    userRepo := repository.NewUserRepo(s.DB)
    roleRepo := repository.NewRoleRepo(s.DB)
    // ... ~55 个 NewXxxRepo
    userSvc := service.NewUserService(userRepo, roleRepo, s.Audit /* ... */)
    // ... ~35 个 NewXxxService，且顺序敏感（注释写着"必须先于 X 初始化因为 Y 依赖它"）
    s.Handlers = &Handlers{ UserHandler: handler.NewUserHandler(userSvc), /* ... 40 行 */ }
    return nil
}

// ③ mega-constructor：某设备服务构造器单行 ~28 个位置参数
svc := service.NewXxxDeviceService(a, b, c, d, e, f, g, h, i, j, k, l, m,
    n, o, p, q, r, s, t, u, v, w, x, y, z, aa)  // 传错一个位置，编译过、运行错
```

三个痛点的共性：**装配逻辑没有边界**。55 个 repo、35 个 service、40 个 handler 全挤在一个方法的一个作用域里，谁依赖谁靠人肉排序（注释里写满"must init before"）；服务参数一多就退化成位置参数长列表，传错位置编译器抓不到。

这不是「缺 DI 框架」的表现，是「缺分域组织」的表现。下面看三条出路。

## 二、三方对比：手工注入 vs wire vs fx

| 维度 | 手工注入（本项目选） | google/wire | uber/fx |
|------|-------------------|-------------|---------|
| 注入时机 | 编译期（就是普通 Go 代码） | 编译期（codegen 生成 Go 代码） | **运行期**（反射 + DAG 解析） |
| 编译期安全 | ✅ 全是普通函数调用 | ✅ 生成的也是普通代码 | ❌ 缺依赖要到启动才 panic |
| 启动可读性 | ✅ 装配即代码，点进去就看到 | ⚠️ 要看生成的 `wire_gen.go` | ❌ Provide/Invoke 声明，实际图靠 fx 运行时拼 |
| 样板量 | 中（靠分域 + 聚合 struct 压下来） | 低（wire 生成接线） | 低（声明式） |
| 学习/魔法成本 | ✅ 零，会写 Go 就会 | ⚠️ wire 的 ProviderSet/绑定规则 | ❌ 高：生命周期钩子、参数对象、结果对象、模块 |
| 生命周期管理 | 手工（`Server.Start/Stop`，见 01） | ❌ 不管（只管构造） | ✅ 内置 `OnStart/OnStop` 钩子 |
| 可测试性 | ✅ 直接 `newXxxModule(fakeDeps)` | ✅ 手工替换 provider | ⚠️ 要 `fxtest`，或绕过容器 |
| 与现有 rules 契合 | ✅ 零全局状态、显式依赖流向 | ⚠️ 生成物是全局装配函数 | ❌ 全局 App 容器，违背零全局状态 |
| 可移植性 | ✅ 无依赖 | ⚠️ 需 `wire` 工具链 + go generate | ⚠️ 运行期强依赖 fx |
| 排错难度 | ✅ 栈直达 | ⚠️ 栈经过生成代码 | ❌ DAG 解析错误信息晦涩 |

**各自的「何时选它」**：

- **手工注入** —— 组件数量可控（本项目 12 域）、团队重视「装配即代码、无魔法」、已有零全局状态约束。**本项目全中。**
- **wire** —— 依赖图很深很宽、手工接线确实开始出错、团队接受 codegen 工作流（每次改依赖跑 `go generate`）。典型是 50+ provider 的中大型服务。
- **fx** —— 需要**动态**模块拼装（按 feature flag 组合不同模块）、需要框架级 `OnStart/OnStop` 生命周期钩子、团队已在 fx 生态（如用了 Uber 的其他库）。典型是插件化的大型微服务。

### 「手工注入是 greenfield 首选」是行业共识，不是本项目偏好

这一点值得单独强调：选手工注入**不是因为本项目 rules 这么写**，而是它本就是 Go 社区对新项目的主流建议。多篇成熟讨论都把手工构造注入列为默认起点、DI 框架列为「对象图大到手工重复才上」的后备：

- [rednafi《You probably don't need a DI framework》](https://rednafi.com/go/di-frameworks-bleh/)：DI 作为**技术**很有用，但 DI **框架**在 Go 里往往得不偿失——普通构造函数已经把依赖显式化了。
- [ehewen《DI in Go》](https://ehewen.com/en/blog/go-dependency-injection/) 明确把 Manual Injection 标为 "Recommended First Choice"。
- [softwarepatternslexicon](https://softwarepatternslexicon.com/go/modern-design-patterns-in-go/dependency-injection/)：「Go 的 DI 通常始于普通构造函数 + 窄接口；容器可帮忙处理大对象图，但**不该隐藏启动错误与生命周期归属**。」
- [leapcell《Wire vs fx vs Manual》](https://leapcell.io/blog/go-dependency-injection-approaches-wire-vs-fx-and-manual-best-practices)：系统对比后，把「plain manual」列为常被低估但最透明的方案。

共识的判据高度一致：**对象图手工写不再痛（重复到成百上千行）时才上库**；容器不应吞掉编译期检查与启动错误。本项目 12 个域、分域装配后每处都短，远未到那个临界点——这才是选手工注入的真正理由，rules 只是把这个共识固化成了本项目约定。

## 三、推荐落地形态：按域装配 + deps 聚合（仅骨架，不写进代码）

### 3.1 deps 聚合结构体，干掉 mega-constructor

把「几乎每个域都要的公共依赖」收进一个结构体，一把传：

```go
// internal/server/deps.go（示意，Step 05+ 落地）
// 公共基础设施，所有域装配共享
type ModuleDeps struct {
    DB    *gorm.DB
    RDB   *redis.Client
    Log   *slog.Logger
    Audit *audit.Recorder     // 审计记录器
    JWT   *jwt.Manager
}
```

服务构造器从「28 个位置参数」变成「1 个 deps + 少量本域特有参数」：

```go
// 反面：service.NewXxxDeviceService(a, b, ..., aa)  // 28 个位置参数
// 正面：本域特有的依赖才单独列，公共的走 deps
func NewDeviceService(deps ModuleDeps, cfg DeviceConfig) *DeviceService { ... }
```

位置参数长列表的两个致命伤——传错位置编译器不报错、加参数要改所有调用点——聚合成 struct 后都消失：字段名显式、加字段不破坏现有调用。

### 3.2 按域装配函数，干掉 165 行扁平 initHandlers

每个域一个装配函数，域内 repo→service→handler 的链路自己收敛，对外只暴露一个 handler：

```go
// internal/server/modules.go（示意）
type moduleHandlers struct {
    User *user.Handler
    Role *role.Handler
    // ... 每域一个，但装配细节各自封装
}

func buildModules(deps ModuleDeps) *moduleHandlers {
    return &moduleHandlers{
        User: newUserModule(deps),
        Role: newRoleModule(deps),
        // ... 一域一行，加域只加一行
    }
}

// 单个域的装配：repo → service → handler 链路收敛在域内一处
func newUserModule(deps ModuleDeps) *user.Handler {
    userRepo := repository.NewUserRepo(deps.DB)
    roleRepo := repository.NewRoleRepo(deps.DB)
    svc := user.NewService(deps, userRepo, roleRepo)   // 遵 domain-architecture：service 按域分子包
    return user.NewHandler(svc)
}
```

对比老项目 B 的 165 行大平铺，收益是：
- **加一个域** = 写一个 `newXxxModule` + 在 `buildModules` 加一行，不用在 40 字段结构体里找位置；
- **依赖顺序局部化**：跨域依赖（如 user 需要 role 的 converter）在域函数内显式传，不再靠全局 "must init before" 注释；
- **可测**：`newUserModule(fakeDeps)` 直接单独装配一个域测试。

### 3.3 与 repo/事务约定的关系

装配顺序天然是 **repo → service → handler**（[domain-architecture.md](../../../../.claude/rules/domain-architecture.md) 规则 2 的依赖链）。事务不在装配期处理——[repo-transaction-convention.md](../../../../.claude/rules/repo-transaction-convention.md) 规则 4 明确「事务在 service 层用 `s.db.Transaction` + 闭包内 `NewXxxRepo(tx)` 重建」，所以装配期只需把 base repo（吃 `deps.DB`）注入 service，事务版 repo 在运行期由 service 自己重建。装配层不碰事务，职责干净。

## 四、被否掉的方案

### ① google/wire —— codegen 的间接层不抵边际收益

wire 用 `go generate` 把「provider 函数集合」编译成一个 `wire_gen.go` 装配函数。它确实能省掉手写接线，但：
- **省下的正是我们靠「分域 + 聚合 struct」已经压到很低的那部分**——12 个域、每域一个 `newXxxModule`，手工接线量本来就不大；
- **多一层 codegen 工作流**：每次改依赖要记得 `go generate`，忘了就编译过但装配是旧的；
- **排错栈穿过生成代码**，比纯手写多一跳。

**重估触发**：provider 数量到 50+、或手工接线开始频繁出现"传错依赖"的 bug 时，wire 是比 fx 更稳妥的第一升级选项（它仍是编译期、生成的仍是普通 Go 代码）。

### ② uber/fx —— 运行期容器与显式装配的取向冲突

fx 是运行期 DI 容器：`fx.Provide` 注册构造器、`fx.Invoke` 触发、框架用反射解析依赖 DAG 并管理 `OnStart/OnStop`。否掉的核心理由：
- **运行期反射**：缺依赖、类型不匹配到**启动时**才 panic，丢掉了 Go 最大的优势——编译期安全；
- **全局 App 容器**：`fx.App` 本质是个全局依赖注册表，与「构造函数即依赖契约、装配在组合根、零包级可变状态」的显式装配取向直接冲突（[README 原则 #10 零全局状态](../../README.md) 与此一致，非独创）；
- **DAG 排错难**：依赖图错误的报错信息晦涩，新人难定位；
- **心智税重**：参数对象（`fx.In`）、结果对象（`fx.Out`）、模块（`fx.Module`）、生命周期钩子是一整套要学的概念。

fx 的 `OnStart/OnStop` 生命周期钩子确实是它的亮点，但本项目用 `run.Group` / `Server{Start/Stop}`（见 [01](./01-应用启动与组件生命周期-main组合根与gin-cron编排.md)）已经覆盖了这个需求，且更显式。**重估触发**：需要按 feature flag 动态拼装模块、或组件生命周期复杂到手工 `Start/Stop` 排不清时。

### ③ service locator / 包级全局 —— 最差，明确禁止

「搞个全局 registry，用到啥去里面取」是最省事也最坏的方案：依赖关系彻底隐式，编译器帮不上忙，运行期才炸。老项目 B 的 cron jobs 把 repo/service 塞包级全局（`var aggGroupRepo ...`）就是这个反模式的局部体现——[03](./03-cron任务注册与复用-jobs装配与Server生命周期集成.md) 专门讲怎么改掉它。本项目 rules 全线禁止包级可变全局状态。

## 五、成本与 tradeoff（已知并接受）

- **手工装配的样板不会归零**：12 个 `newXxxModule` 函数要手写，加域要手动加一行。接受——换来的是「装配即代码、点进去就懂、编译期全保障、零框架依赖」，且分域后每处都很短。
- **deps 聚合 struct 的粒度要把握**：塞太多变成新的 god-struct，塞太少又退化回长参数列表。约定：**只放「过半数域都要」的公共依赖**（DB/RDB/Log/Audit/JWT），本域特有的（如某域专用的第三方 client）单独作参数传。
- **跨域 converter 依赖仍需手工传**：如 user 域要用 role 域的 exported converter（[domain-architecture.md](../../../../.claude/rules/domain-architecture.md) 规则 5），在 `newUserModule` 里显式传入。接受——显式优于隐式。
