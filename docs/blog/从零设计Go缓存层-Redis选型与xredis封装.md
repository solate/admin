# 从零设计 Go 缓存层：Redis 选型与 xredis 封装实战（2026）

> 本文是一篇完整实战文档。读完你能独立完成：想清楚「这个项目到底要不要引入 Redis」→ 用一套决策框架判断「什么状态放哪里」→ 选定客户端库与封装形态 → 落地连接层 `pkg/xredis`（可跑 + 带单测）→ 规划业务缓存层 `pkg/cache`（黑名单 / 验证码 / 限流的家）。
>
> 技术栈：Go 1.26 + [go-redis/v9](https://github.com/redis/go-redis) + [miniredis](https://github.com/alicebob/miniredis)（单测）+ Redis 7.x + PostgreSQL（真相源）。基于真实 SaaS 后端项目 `backend/` 的实现——连接层是已跑代码，业务缓存层是本文给出的设计示范（明确标注区分）。

---

## 0. 开篇：缓存层到底难在哪

Redis 看起来是「装个服务、`SET`/`GET` 一下」的基础设施，真要在一个项目里落地，绕不开一串更上游的问题：

1. **凭什么加？**——以前很多项目不用 Redis 也跑得好好的，现在多引入一个组件（要部署、监控、备份、考虑它挂了怎么办），payoff 到底在哪？
2. **不能自己封吗？**——用 `sync.Map` + TTL 写个进程内 `pkg/cache` 不行吗，为什么要外部组件？
3. **单入口就没事了吧？**——如果是 BFF 结构、大家都走单一入口，甚至只跑一个实例，多副本一致性问题不就没了，Redis 是不是就没意义了？
4. **封装返回什么？**——连接封装该返回接口还是具体类型？要不要做成全局单例？
5. **业务代码放哪？**——JWT 黑名单、验证码、限流这些逻辑，写进连接封装包，还是散落在各个 service 里裸调 Redis？

本文按这五个问题的顺序展开：先回答「要不要用」（第 1-2 章，决策框架），再回答「怎么选、怎么封」（第 3-5 章，选型 + 连接层），最后回答「业务代码放哪」（第 6 章，分层）。所有结论都基于真实项目，诚实区分「已跑代码」和「设计示范」。

**一句话结论先摆这里**：本项目瞄准 k3s 多副本部署，Redis 是**必选项**——不是因为「大家都用」，而是多副本 + 进程会重启这两根轴一旦成立，JWT 黑名单 / 验证码 / 限流就没有进程内的正确实现。连接层用 go-redis + 返回原生 `*redis.Client`（无接口、无单例），业务层收拢到 `pkg/cache`，两层各司其职。

---

## 1. 要不要用 Redis：先回答「凭什么加」

增加基础设施必须有明确 payoff，否则就是过度设计。这一章不灌「Redis 很强你就该用」的鸡汤，而是给一把能自己判断的尺子。

### 1.1 根本轴：单副本 vs 多副本

Redis 的必要性几乎完全由部署形态决定。老项目不用 Redis 能活、新项目必须用，分水岭就是这一个轴：

| | 单副本（老项目常态） | 多副本（k8s/k3s 标配） |
|---|---|---|
| **进程内状态是否安全** | ✅ 唯一进程，内存就是真相 | ❌ 每个 pod 独立进程，状态分裂 |
| **JWT 黑名单存 `sync.Map`** | ✅ 撤销立即生效 | ❌ pod-A 撤销，pod-B 仍认 token 有效（**安全漏洞**） |
| **验证码存内存** | ✅ 生成→校验同进程 | ❌ 生成在 pod-A、校验落 pod-B → 验证码不存在（**登录失败**） |
| **限流计数器** | ✅ 本进程准确 | ❌ 每 pod 独立计数 → 10 req/min × 3 pods = 实际 30（**限流失效**） |
| **分布式锁** | ✅ `sync.Mutex` 即可 | ❌ 每 pod 内部锁，跨 pod 无互斥（**定时任务重复执行**） |

这不是「Redis 比内存好」，而是**多副本下进程内状态根本做不到跨副本一致**。老项目不用 Redis，是因为它跑单副本 + 小流量；本项目一开始就瞄准 k3s 多副本（`config.go` 注释提到 k3s、`go.mod` 引入 dbresolver 读写分离、`pkg/xredis` 预留集群迁移路径），前提变了，必须的组件也变了。

### 1.2 决策框架：什么状态放哪里

成熟团队真正依据的不是「要不要用 Redis」，而是 [12-Factor App](https://12factor.net/processes) 的「无状态进程」原则下的一棵决策树——进程随时可被销毁重建，任何状态都要推到外部后端。问题于是变成「外部存储选谁」：

```
需要存储一份状态
  │
  ├─ Q1: 需要跨副本即时一致？（一个 pod 写，另一个 pod 立刻要读到）
  │
  ├─ 否 ──→ 单副本？ ├─ 是 → 进程内 sync.Map 可以（但不面向扩容）
  │                  └─ 否 → 大概率 Q1 答错了，重想
  │
  └─ 是 ──→ Q2: 需要持久化（重启不丢）？
             ├─ 是 ──→ 数据库（订单/用户/权限配置——真相源永远在 DB）
             └─ 否 ──→ Q3: 读写特征？
                       ├─ 极高读(>1000:1) + 容忍秒级延迟 + DB 是真相源
                       │      → 进程内缓存 + 短 TTL + 变更广播刷新（如 PermissionCache）
                       └─ 其他 → Redis（黑名单/验证码/限流/锁/会话）
```

「**跨副本即时一致**」是第一个、也是最重要的分叉。大部分「需要立即一致但不需持久化」的状态都落在 Redis——这正是它在后端架构里的经典定位：DB 与进程内存之间，那一层共享、快速、带 TTL 的存储。

### 1.3 成熟团队的收敛答案

把「大家实际都怎么干」摊开，选择高度收敛：

- **会话 / Token 几乎都进 Redis**：Spring Session、Laravel、Django、Rails 的生产 session store 默认或首选就是 Redis。**即使用无状态 JWT，登出黑名单 / refresh token 轮换仍要 Redis**——JWT 的「无状态」恰恰意味着它没法主动撤销，黑名单是补这个洞的，而黑名单必须跨副本共享。
- **验证码 / 短信码存 Redis**：`SETEX` 带 TTL，过期自动清理。
- **限流用 Redis**：固定窗口 `INCR`+`EXPIRE`、滑动窗口 `ZSET`、令牌桶 Lua 脚本。网关层（Kong、Sentinel）的分布式限流后端基本都是 Redis。
- **分布式锁用 Redis**：`SET NX EX` 或 Redlock（Redisson 是 Java 生态事实标准）。

### 1.4 那自己封装进程内 cache 呢

回到最实在的一问：不引入 Redis，自己用 `sync.Map` + TTL 写个 `pkg/cache` 不行吗？单副本时完全行；多副本时你会逐个撞上下面这些墙，而每堵墙 Redis 都已经砌好了门：

| 问题 | 自封装进程内 cache | Redis 现成能力 |
|---|---|---|
| **跨副本状态分裂** | 黑名单只在 pod-A，pod-B 不知情 → 安全漏洞 | 所有副本读写同一份 |
| **TTL 要手写** | 验证码过期得自起定时器扫描删除 | `SETEX` / `EXPIRE` 服务端自动过期 |
| **原子操作缺失** | 限流 `get→+1→set` 非原子，race | `INCR` 单命令原子；Lua 脚本 |
| **内存不可控** | 无淘汰策略 → OOM | `maxmemory` + `allkeys-lru` |
| **刷新协调困难** | 权限变更如何通知其他 pod？自写 pub/sub | Redis Pub/Sub 现成 |
| **无持久化** | 重启全丢，冷启动打爆 DB | RDB / AOF 可选持久化 |

**核心结论**：多副本下自封装 cache，为了正确工作迟早要补上「跨副本同步 + TTL + 原子 + 淘汰 + 失效广播」——**补齐就是一个自制的、没经过生产打磨的 Redis**。这笔账怎么算都是直接用 Redis 划算。唯一例外是决策树右下角那类（读极多 + DB 真相源 + 容忍秒级延迟，如 PermissionCache），可进程内缓存，但「变更后刷新」仍要借 Redis Pub/Sub 协调。

> 完整论证（决策树各分支、成熟框架清单、逐场景归属）见 [02-为什么需要redis-进程内缓存vs集中式缓存](../saas-backend/research/redis/02-为什么需要redis-进程内缓存vs集中式缓存.md)。

## 2. 单入口 / 单实例，还需要吗

读到这里你可能有个反问，这也是本项目真实纠结过的问题：

> 如果是 BFF 结构、大家都跟 BFF 通讯，那 Redis 是不是就没意义了？你说的都是为了多副本状态相同，那只有一个 BFF 是不是就没事了？

这个直觉有一半是对的，但结论下过头了。拆成「成立的内核」+「两个漏洞」。

### 2.1 先诚实让步：单进程确实能砍掉一部分

若**真的永远单进程**，第 1 章的「多副本状态分裂」论证确实塌了，而且部分场景反而更简单：

- **分布式锁 → `sync.Mutex`**：单进程内 goroutine 之间用标准库互斥锁天然成立，没有网络往返，也不用碰 Redlock 那套有争议的边界（时钟漂移、GC 停顿导致锁失效）。单进程下 `sync.Mutex` 严格优于 Redis 锁。
- **权限缓存刷新 → 直接刷本进程内存**：省掉 Redis Pub/Sub 跨副本广播，权限改了直接重建本进程的 map。
- **限流计数 → 进程内 counter 本就正确**：没有「副本数乘法失效」，本地计数器就是准的。

**这部分明确让给你：单实例确实能砍掉「锁 / 权限刷新协调 / 限流」这三类的 Redis 需求。** 问题只在于——「单一入口」真的等于「单一进程且永不重启」吗？

### 2.2 漏洞 1：BFF 是架构角色，不是副本数

「BFF」（Backend For Frontend）描述的是**职责**——为某个前端做接口聚合、裁剪、协议适配的那一层，讲的是流量拓扑与分工，不是**部署了几个实例**。「大家都跟 BFF 通讯」是说所有前端流量经过 BFF 这个**逻辑角色**，不是说这个角色只由**一个进程**承载。生产里几乎总是：一个逻辑 BFF，背后 N 个副本，前面挂负载均衡。

BFF 会多副本，理由和任何在线服务一样：高可用（一个 pod 崩了其他顶上）、滚动发布（零停机部署本身就要求发布期同时存在新旧实例）、横向扩容。而且有个**反噬点**：BFF 是所有前端流量的必经单一入口，这恰恰是整个系统里**最不该做成单点故障（SPOF）**的位置——把最关键的咽喉做成单实例，等于赌它永不出问题。所以生产里 BFF 不但会多副本，还是优先级最高的多副本对象。

### 2.3 漏洞 2：重启轴（与副本数正交）

做最大的让步：**假设你铁了心就跑一个 BFF 实例，永不扩容。** Redis 还是有意义——因为你只消除了「副本轴」，还有一根正交的轴：**进程会重启**。

单实例不等于「永远活着的同一个进程」。发布新版本、崩溃（panic/OOM）、节点驱逐 / 重调度、机器维护——每一次重启，进程内存里的状态全部清零：

- **JWT 黑名单**：用户登出拉黑 token → 两分钟后发个版重启 → 黑名单空了 → 「已登出」的 token 在自然过期前**重新生效**。这是发生在「发个版」这种日常动作上的**安全窗口**。
- **验证码**：发出后（存内存，5 分钟有效）恰逢滚动发布 → 验证码没了 → 用户输入正确也校验不过。发布期在途验证码全灭。

这是两根独立的轴：

| 轴 | 方向 | 问什么 | 单实例能消除吗 |
|---|---|---|---|
| **副本轴**（第 1 章主题） | 横向：同一时刻多个进程 | 状态在 pod 之间一致吗？ | ✅ 能 |
| **重启轴**（本章补充） | 纵向：同一角色跨时间的进程生命周期 | 状态在进程重启后还在吗？ | ❌ 不能 |

**Redis 的意义从来不只绑定在「多副本一致」这一个理由上**，它同时是「跨重启存活的快状态存储」。你砍掉副本轴，重启轴还在——靠持久性活着的黑名单、在途验证码，单实例一根汗毛都没动到它。

> 完整辩论（BFF 定义、SPOF、两轴表格、逐场景重判）见 [03-单bff单副本还需要redis吗](../saas-backend/research/redis/03-单bff单副本还需要redis吗.md)。


## 3. 客户端选型：go-redis + 具体类型 + 无单例

确定了「要用 Redis」，接下来是「用哪个库、封装返回什么」。这一层的坑我们在老项目 B 上实打实踩过。

### 3.1 2026 Go Redis 客户端全景

| 库 | 维护方 | 确定类型 | API 手感 | 一句话 |
|---|---|---|---|---|
| **go-redis/v9** | 官方 | ✅ `*redis.Client` | 链式 `.Err()/.Result()` | 最成熟、语料最多、老项目在用 |
| **rueidis** | 官方 | ✅ | command-builder，较 verbose | 高性能新锐（auto-pipeline + 客户端缓存） |
| **redigo** | 社区 | ❌ 无类型安全 | `Do("SET", k, v)` 字符串命令 | 老牌，2026 新代码不推荐 |

库的选择没有悬念——go-redis/v9 已是依赖、生态最广、AI 补全最不易错。rueidis 的性能红利单节点后台吃不到，不值新依赖 + 陌生 API 的代价。**真正要决策的是封装的返回类型：接口还是具体类型？** 这才是老项目痛点的根。

### 3.2 反面镜鉴：老项目的 UniversalClient 单例三宗罪

老项目 `legacy-cms/pkg/xredis` 一开始就返回 `redis.UniversalClient` 接口 + 全局 `sync.Once` 单例，跑通了业务但用起来别扭：

```go
// 反面教材（节选）
var (
    client redis.UniversalClient // 全局单例
    once   sync.Once
)
func Connect(cfg Config) (redis.UniversalClient, error) {
    once.Do(func() { client = newClient(cfg) }) // 写包级全局
    // ... Ping
}
func GetRedis() redis.UniversalClient { return client } // 取包级全局
```

1. **接口没带来任何好处**：全项目只跑单节点，消费方为了用 `redis.Nil` 仍要直接 `import go-redis`，接口没能隐藏底层库，白套一层抽象。
2. **全局单例难测、难多实例**：`Connect` 写全局、`GetRedis()` 取全局，无法并行测试、无法多目标 Redis，第二次 `Connect` 换 Config 还静默无效（`once` 已消费）。
3. **集群支持是假的**：cluster 分支只塞单个 addr，无法真正多节点。所谓「为集群准备的接口」，连集群都没真正支持。

诊断：错不在选了 go-redis，而在**为不存在的集群需求预付了接口 + 单例的复杂度，两个成本都白交了**。

### 3.3 选型矩阵与结论

| 维度 | **具体 `*redis.Client`（选）** | `UniversalClient` 接口 | rueidis |
|---|---|---|---|
| 单节点复杂度 | ✅ 最低 | 多一层接口无收益 | 新依赖 + 陌生 API |
| 上集群成本 | ⚠️ 改封装 1 处 | ✅ 改配置零改（限单 key） | 需重学 API |
| 与 `pkg/xgorm` 一致 | ✅ 同构（都返回原生类型） | ❌ 风格分裂 | ❌ |
| 可测试性 | ✅ 无单例，可并行 | ❌ 全局单例难并行 | ✅ |
| AI 生成友好 | ✅ 类型确定、语料最多 | ✅ 但接口跳转绕 | ⚠️ 语料少 |

**结论：go-redis/v9 + 返回具体 `*redis.Client` + 无接口无单例。** 直接否掉老项目两个痛点，也与 `pkg/xgorm` 返回原生 `*gorm.DB` 的既有风格同构——同一个仓库，连接层封装的心智统一。

> 完整选型（rueidis 深度对比、UniversalClient 到底买到什么）见 [01-redis-客户端封装选型调研](../saas-backend/research/redis/01-redis-客户端封装选型调研.md)。

## 4. 连接封装 pkg/xredis（真实可跑代码）

选型定了（go-redis + 具体类型 + 无单例），落到代码就是一个极薄的连接层：只管「建连 + 探活 + 关闭」，不碰任何业务语义。下面是 `pkg/xredis` 的全部真实代码。

### 4.1 Config：纯连接参数

```go
package xredis

// Config Redis 连接参数。无 mapstructure tag，由 main 从 internal/config 逐字段映射。
type Config struct {
	Addr         string // 地址 host:port
	Password     string // 密码（建议 APP_REDIS_PASSWORD 注入）
	DB           int    // 库编号
	PoolSize     int    // 连接池大小，0 用 go-redis 默认
	MinIdleConns int    // 最小空闲连接，0 用默认
	MaxRetries   int    // 命令重试次数，0 用默认
	DialTimeout  int    // 建连超时(秒)，0 用默认
	ReadTimeout  int    // 读超时(秒)，0 用默认
	WriteTimeout int    // 写超时(秒)，0 用默认
}
```

刻意不挂 `mapstructure` tag——配置解析是 `internal/config` 的事，`pkg/xredis` 作为可 copy 的 x 包不该知道项目怎么加载配置。字段由 `main` 从 `internal/config.RedisConfig` 逐字段映射进来。所有可选参数**为 0 时回退 go-redis 默认**，配置文件里不写就是「用官方推荐值」。

### 4.2 New：建连 + 探活 fail-fast

```go
// New 创建单节点 Redis 客户端并探活。
// 返回 go-redis 原生 *redis.Client（不套接口、不做单例），与 pkg/xgorm 返回 *gorm.DB 风格一致。
// 后期若上集群，只需把 redis.NewClient 换成 redis.NewUniversalClient 并调整返回类型，改动收敛在此一处。
func New(cfg Config) (*redis.Client, error) {
	client := redis.NewClient(buildOptions(cfg))

	// 构造即探活：ping 失败立刻关闭并返回 error（fail-fast）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}

// buildOptions 把项目 Config 映射为 go-redis 的 *redis.Options。
// 连接池/超时等可选参数仅在 >0 时覆盖，为 0 保留 go-redis 默认值——这段纯映射逻辑可零依赖单测。
func buildOptions(cfg Config) *redis.Options {
	opts := &redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	}
	if cfg.PoolSize > 0 {
		opts.PoolSize = cfg.PoolSize
	}
	if cfg.MinIdleConns > 0 {
		opts.MinIdleConns = cfg.MinIdleConns
	}
	if cfg.MaxRetries > 0 {
		opts.MaxRetries = cfg.MaxRetries
	}
	if cfg.DialTimeout > 0 {
		opts.DialTimeout = time.Duration(cfg.DialTimeout) * time.Second
	}
	if cfg.ReadTimeout > 0 {
		opts.ReadTimeout = time.Duration(cfg.ReadTimeout) * time.Second
	}
	if cfg.WriteTimeout > 0 {
		opts.WriteTimeout = time.Duration(cfg.WriteTimeout) * time.Second
	}
	return opts
}
```

两个设计点：**构造即探活**——`New` 返回时 client 一定是连通的，连不上直接 fail-fast，不把「连接是否可用」的判断推给调用方；**`buildOptions` 抽成纯函数**——它是本包唯一的自有逻辑（`Config → *redis.Options` 映射），抽出来才能零依赖、零网络地单测。

### 4.3 为什么不包装 Get/Set

`New` 返回原生 `*redis.Client` 后就撒手，不再封装任何 `Get`/`Set`/`Incr` 方法。原因：

- **go-redis 的 API 已经足够好**——链式 `.Set(ctx,k,v,ttl).Err()` / `.Get(ctx,k).Result()`，再包一层只是徒增维护面且拦住了 go-redis 的新命令。
- **业务语义不属于连接层**——「黑名单」「验证码」这些概念一旦进 `xredis`，它就绑死了本项目、不再可 copy（违反 reusable-package.md）。这些属于下一章的 `pkg/cache`。
- **集群迁移路径收敛在一处**——哪天要上集群，只需把 `New` 里的 `redis.NewClient` 换成 `redis.NewUniversalClient`、返回类型改成 `redis.UniversalClient`，调用方因命令 API 相同而基本不动。

### 4.4 miniredis 单测（自包含）

x 包必须带自包含单测（reusable-package.md 规则 3）。Redis 客户端的测试用 [miniredis](https://github.com/alicebob/miniredis)——纯 Go 内存 Redis，test-only、不进生产二进制、不依赖外部服务：

```go
func TestNew_SetGet(t *testing.T) {
	mr := miniredis.RunT(t)

	client, err := New(Config{Addr: mr.Addr()})
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Set(ctx, "k", "v", time.Minute).Err(); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	got, err := client.Get(ctx, "k").Result()
	if err != nil || got != "v" {
		t.Fatalf("Get() = %q, err = %v", got, err)
	}
}
```

三个测试覆盖全部公开行为：`TestNew_SetGet`（连通 + 往返）、`TestNew_ConnRefused`（连不上 fail-fast）、`TestBuildOptions`（映射逻辑：>0 覆盖、=0 保留默认）。`go test ./pkg/xredis/` 全绿、不需要本地跑 Redis。

## 5. 服务端版本：Redis 7.x vs 8.x 何时升

上一章选的是**客户端库**（go-redis），这一节谈**服务端版本**——本地装 7.x 还是追最新 8.x？go-redis v9 同时支持 7.x / 8.x，服务端选哪个版本，`pkg/xredis` 都不用改一行，所以这是一个正交的、可以独立决定的问题。

**结论先行：本项目当前阶段保持 7.x，没必要升 8.x，收益接近零。**

Redis 8.0（2025-03 GA）的三类主要变化，逐条对到本项目（管理后台）场景：

| 8.x 变化 | 是什么 | 对本项目的意义 |
|---|---|---|
| **模块入内核 + 向量检索** | RedisJSON / RediSearch / RedisTimeSeries 打进核心，新增 Vector Sets（冲 AI/RAG） | ❌ 无关。后台用 Redis 只做 session/黑名单/验证码/缓存，纯单 key KV，一个都用不上 |
| **性能提升** | 部分场景吞吐更高、延迟更低 | ⚠️ 吃不到。单节点低吞吐后台，7.x 早已绰绰有余 |
| **许可证 SSPL → AGPLv3** | 社区版改用 AGPLv3 | ⚠️ 是约束不是收益，但**大概率无影响**：只在「改 Redis 源码并作为网络服务对外提供」时才触发传染；直接跑官方二进制当缓存不构成 AGPL 义务 |

**决策原则：版本对齐 > 版本领先。**

真正该做的不是追新版本号，而是**让本地版本 = 部署/生产环境版本**，避免「本地正常、线上行为不一致」——这比追新重要得多。全新起、无历史包袱直接用 8.x 当默认也行（它是当前正统），但**不为升级而升级**。何时才值得升 8.x：奔着功能去——哪天后台要加向量检索 / 全文搜索 / 时序，那才是升级时机。

> 完整版本分析（8.0 GA 时间线、AGPLv3 传染边界、向量能力定位）见 [01-redis-客户端封装选型调研](../saas-backend/research/redis/01-redis-客户端封装选型调研.md)。

## 6. 业务缓存层 pkg/cache（分层设计）

到这里连接层（xredis）已经就位，但它**刻意不懂任何业务**——没有 `Revoke`、没有 `Captcha`、没有 `RateLimit`。那这些业务代码放哪？答案不是散在各个 service 里裸调 `rdb.Set(...)`，而是收拢到一个专门的业务缓存包 `pkg/cache`。

> 说明：`pkg/xredis` 是当前**已跑的真实代码**；本章 `pkg/cache` 是**规划中的设计示范**——业务用法（黑名单/验证码/限流）目前还是空目录，下面的代码展示的是落地时该长的样子，而不是仓库现状。

### 6.1 为什么要两层

分层的理由，一句话：**xredis 要可复用，pkg/cache 要懂业务，两个诉求互斥，只能拆开。**

- `pkg/xredis`：连接层。按 `reusable-package.md` 的规矩，它必须自包含、可整目录 copy 到别的项目——所以它**绝不能**出现 `blacklist:` 这种业务 key 前缀。一旦掺进业务语义，它就不再可复用了。
- `pkg/cache`：业务缓存层。黑名单的 key 约定、验证码的 TTL、限流的 Lua 脚本——这些是本项目的业务知识，是它们的家。

这不是我临时造的分层：`domain-architecture.md` 规则 6 的业务包清单里**早已列了 `cache`**（`audit, cache, config, constants, response, xcontext, xerr`），只是还没填代码。位置一直预留在那。

### 6.2 三层依赖，与数据库层同构

```
service / middleware  →  pkg/cache  →  *redis.Client（xredis 建连后注入）
```

对照数据库层的 `service → repository → *gorm.DB`，两者结构完全一致：中间那层（cache / repository）收敛对底层客户端的所有调用，上层只认业务方法（`IsRevoked` / `FindByID`），不碰 `rdb.Get` / `db.Where`。构造函数注入也沿用项目既有约定——每个能力一个小结构体，持有注入的 `*redis.Client`。

### 6.3 TokenBlacklist：登出与撤销

```go
// pkg/cache/blacklist.go —— 设计示范
package cache

import (
    "context"
    "time"

    "github.com/redis/go-redis/v9"
)

const blacklistPrefix = "blacklist:" // key 约定收敛在一处

type TokenBlacklist struct {
    rdb *redis.Client
}

func NewTokenBlacklist(rdb *redis.Client) *TokenBlacklist {
    return &TokenBlacklist{rdb: rdb}
}

// Revoke 拉黑一个 token，TTL 设为该 token 的剩余有效期——
// 到期后 key 自动消失，黑名单不会无限膨胀。
func (b *TokenBlacklist) Revoke(ctx context.Context, tokenID string, ttl time.Duration) error {
    return b.rdb.Set(ctx, blacklistPrefix+tokenID, 1, ttl).Err()
}

// IsRevoked 每个请求都会调（middleware 里），用 EXISTS 判断。
func (b *TokenBlacklist) IsRevoked(ctx context.Context, tokenID string) (bool, error) {
    n, err := b.rdb.Exists(ctx, blacklistPrefix+tokenID).Result()
    return n > 0, err
}
```

TTL = token 剩余有效期是关键设计：既保证撤销在 token 自然过期前始终生效，又让 key 自动回收，无需清理任务。

### 6.4 Captcha：一次性校验

```go
// pkg/cache/captcha.go —— 设计示范
const captchaPrefix = "captcha:"

func (c *Captcha) Save(ctx context.Context, key, code string) error {
    return c.rdb.Set(ctx, captchaPrefix+key, code, 5*time.Minute).Err() // SETEX 语义
}

// Verify 校验后立即删除——验证码一次性，防止重放。
func (c *Captcha) Verify(ctx context.Context, key, input string) (bool, error) {
    k := captchaPrefix + key
    code, err := c.rdb.Get(ctx, k).Result()
    if err == redis.Nil {
        return false, nil // 不存在或已过期
    }
    if err != nil {
        return false, err
    }
    c.rdb.Del(ctx, k) // 用过即焚
    return code == input, nil
}
```

### 6.5 RateLimiter：原子计数

```go
// pkg/cache/ratelimit.go —— 设计示范
// 固定窗口：第一次 INCR 得 1 时设过期，后续只 INCR。
// 生产建议用 Lua 脚本把 INCR+EXPIRE 合成一条原子命令，避免两条命令之间进程崩溃留下永不过期的 key。
func (r *RateLimiter) Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
    n, err := r.rdb.Incr(ctx, "rate:"+key).Result()
    if err != nil {
        return false, err
    }
    if n == 1 {
        r.rdb.Expire(ctx, "rate:"+key, window)
    }
    return n <= limit, nil
}
```

### 6.6 为什么不散在 service 里裸调

如果不收拢，黑名单的 `blacklist:` 前缀、验证码 5min TTL、限流的 Lua 原子性会散落在各个 service。而这些用法是**跨域共享**的：黑名单被 auth 域（登出时写）和 middleware（每请求读）两处用，验证码被 captcha 域和 auth 域都碰。约定一旦散落就是 drift 的温床——某处写 `blacklist:`、另一处写 `blacklisted:`，某处 TTL 5min、另一处 10min。收拢到 `pkg/cache` 一处，key 约定、TTL 语义、原子性都只有一个定义点，跨域复用、改一处即全改。这与 `service-patterns.md` 想避免的散落是同一个道理。
