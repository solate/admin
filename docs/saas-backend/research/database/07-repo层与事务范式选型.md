# repo 层与事务范式选型论证

> 2026-07-08。本文是 `.claude/rules/repo-transaction-convention.md` 的背景论证，
> 记录「为什么选重建派」的完整推理，避免下次重新纠结、重新搜索。
> 日常写代码不必读，rule 里的硬约束就够；想搞明白「为什么」时看这里。

## 结论速览

本项目采用**重建派**：repo 结构体构造吃 `*gorm.DB`、内部 `query.Use(db)`，
跨 repo 事务在 service 层 `s.db.Transaction` + 闭包内 `NewXxxRepo(tx)` 重建。
理由是与 content-center-backend 生产验证过的写法一致，团队心智统一、零迁移成本。

## 一、为什么事务里必须「重建」repo

repo 构造时 `q` 就绑死在结构体字段上（`&Repo{q: query.Use(db)}`）。`q` 是常量，
连接换不了。所以事务要用 `tx` 连接，**只能拿 `tx` 重新造一个 repo 实例**——
换连接 = 换实例。这是「重建」的根因，不是缺陷，是「连接存在结构体里」这个选择的必然结果。

三种范式本质是「连接从哪来」的三个答案，都能让 repo 用上事务连接，没有谁对谁错：

| 范式 | 连接从哪来 | 事务里代价 | repo 方法签名 |
|------|-----------|-----------|--------------|
| 重建派（本项目） | 结构体字段，构造时绑死 | 事务闭包开头 `NewXxxRepo(tx)` 重建 | 干净，无 q |
| 显式派 | 方法参数传入 | 无重建 | 每个方法挂 `q *query.Query` |
| ctx 派（Transactor） | `Conn(ctx)` 运行时从 ctx 取 | 无重建、无传参 | 干净，但连接藏 ctx |

「不用重建」不是某一派独享——显式派和 ctx 派也不用。选重建派是因为**它把成本
摊在 service 事务闭包那几行 new**（少、集中、可见），而不是摊在每个 repo 方法签名上，
也不是藏进 ctx（不可见、易误用）。

## 二、为什么 Ent 不用重建，GORM Gen 也能不重建

Ent 里 `tx.Client()` 一次调用，返回的 client 上**所有实体**自动绑 tx，所以不用
手动重建每个 repo。本质是：**Ent 直接用生成的 client 当数据访问层，没在上面再套
自己的 repo 结构体**。是「自套 repo 结构体 + 把 q 存进去」制造了重建需求，不是 GORM 的锅。

GORM Gen 的 `*query.Query` **就等于 Ent 的 client**——`tx *query.Query` 身上
`tx.CustomFacePerson`、`tx.CustomFaceImage` 全自动绑好事务连接。所以如果直接用
`query.Query`（不套 repo），也能得到 Ent 式「零重建」体验（rule 规则 4 单 repo 场景就是这个原理）。

**本项目仍选择套 repo 层**，理由只有一个：**收敛重复的 CRUD 查询**（很多 service
用一样的增删改查）。这是唯一动机——不是为了 mock（不写 interface），不是为了给裸
ORM 套皮（Gen 已经够）。既然套了 repo 且把 q 存进结构体，就接受重建这个代价。

## 三、社区没有多数派（别再去搜「大家怎么做」）

Go 生态这件事是分裂的，三派都有大量真实项目，没有权威共识：

- **套 repo 结构体持 db**：Clean Architecture / DDD 模板的默认（content-center、
  各种 go-clean-arch 脚手架）
- **不写 repo，service 直接用 ORM**：小到中型项目常见，r/golang 上「GORM 本身
  就是 repository」呼声很高
- **包级函数 / 直接用生成 client**：sqlc、Ent 用户的常态

所以「工具生成的访问层够不够格当 repo」才是决定性问题：用原始 GORM 的人多半要包
repo（API 太裸），用 Ent 的人多半不包（client 已够）。GORM Gen 的处境更接近 Ent。
本项目选择包一层，纯粹为了 CRUD 复用，不是因为社区主流。

## 四、被否掉的方案

- **ctx 派（Transactor，连接塞 ctx）**：曾写过 `pkg/database/transactor.go`，
  否掉原因是连接藏在 ctx 不可见、易误用；已删除。
- **显式派（每个 repo 方法加 `q` 参数）**：否掉原因是方法签名全部要挂 q，啰嗦。
- **不写 repo（service 直接用 query.Query）**：否掉原因是放弃了 CRUD 复用，
  而复用正是本项目建 repo 的唯一动机。

**重建派的代价**：事务闭包开头有一堆 `txXxxRepo := NewXxxRepo(tx)`。已知并接受。

## 五、WithTx 变体评估（2026-07-09 补）

有人会提议把「独立构造函数 `NewXxxRepo(tx)`」换成「repo 上挂一个 `WithTx(tx)`
流式克隆方法」，并认为这样「更优雅、能避免重复 newRepo」。本节记录完整评估，
结论先行：**WithTx 是重建派内部的一个平级子写法，不是升级；对本项目收益用不上、
代价实打实。原来的 `NewXxxRepo(tx)` 就是很好的做法。**

### 背景：两种写法长什么样

```go
// 现状：独立构造函数
func (s *Service) Transfer(ctx context.Context) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        userRepo := repository.NewUserRepo(tx)   // 直接调构造函数
        orderRepo := repository.NewOrderRepo(tx)
        // ...
    })
}

// 提案：WithTx 流式克隆
func (r *UserRepo) WithTx(tx *gorm.DB) *UserRepo {
    return &UserRepo{db: tx, q: query.Use(tx)}   // 内部仍是重建一个实例
}
func (s *Service) Transfer(ctx context.Context) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        userRepo := s.userRepo.WithTx(tx)        // 从已有 repo 克隆
        orderRepo := s.orderRepo.WithTx(tx)
        // ...
    })
}
```

**关键事实**：`WithTx` 内部 `return &UserRepo{...}` 本身就是一次 newRepo。它没有
「避免重建」——`query.Query` 构造时把连接绑死，事务要用 `tx` 就必须有一个绑 `tx`
的新 `query` 实例 = 新 repo 实例。这是「把 q 缓存进结构体」的必然结果，`WithTx`
只是把 `NewUserRepo(tx)` 改名叫 `r.WithTx(tx)`，实例照样重建。所谓「避免重复
newRepo」是误解。

### 三个论点逐条评估

**论点 1「不是不重建，而是由谁重建」（构造细节收拢）**——真实，但对本项目≈0。

WithTx 确实把「重建时带哪些字段」收进 repo 内部：未来 `UserRepo` 加个 `cache`
字段，只改 `WithTx` 一处，事务调用点无感知；而 `NewUserRepo(tx)` 若签名变了，
每个事务闭包要跟着改。这是真实的 DIP 收益——**但对本项目接近于零**：

1. 本项目 rule（规则 1）已硬性约定 repo 构造**只吃 `*gorm.DB`**，构造签名被钉死，
   「未来加 cache 字段」的前提在约定下不会发生。
2. 就算真加了，`NewUserRepo(db, cache)` 的签名变化会让**所有非事务调用点**都要改，
   WithTx 只省了「事务子集」那部分，耦合没消除、只是挪位置。

**论点 2「DDD 聚合根 / 工作单元」**——修辞灌水，且与本项目哲学冲突。

WithTx 并没有实现 Unit of Work——真正的 UoW 要追踪脏对象、协调提交时机；WithTx
只是换了个绑 `tx` 的连接。而且本项目的事务边界**已经是** `s.db.Transaction(func(tx))`
闭包本身，不需要 WithTx 来「更符合业务语义」。更关键：本文第二节白纸黑字写了
**「套 repo 的唯一动机是收敛 CRUD，不是为了 mock、不是为了 DDD 纯度」**——拿 DDD
聚合根论证 WithTx，正好和自己刻意选的哲学打架。

**论点 3「Repo 层划清边界 / 可换 ORM / 可 Mock」**——跑题，且不适用。

这条根本不是在讲 WithTx，是在讲「要不要有 repo 层」。而且它列的三个好处，本项目
主动放弃了两个：① 可 Mock/测试解耦——本文明确「不写 interface，靠集成测试」，
没有接口，Mock 无从谈起；② 可换 sqlx/原生 SQL——同理没有接口边界，换 ORM 照样
改所有 repo，这层「边界」是漏的。只有「CRUD 复用」这一条成立，而这恰恰是现状
`NewXxxRepo` 已经在做的，和 WithTx 无关。

### 反向成本（WithTx 实打实的代价）

- **破坏与 content-center-backend 一致性**：选重建派的**原始理由**（第零节）就是
  「与老项目一致、团队心智统一、零迁移成本」，content-center 用的是 `NewXxxRepo(tx)`。
  改用 WithTx 恰恰破坏这条一致性理由。
- **字段名 `query` 遮蔽包名**：提案里 `type UserRepo struct { query *query.Query }`
  字段叫 `query`、包也叫 `query`，方法内想引用 `query.Xxx` 包会被字段遮蔽。现状用
  `q` 正是为了躲这个。
- **每个 repo 多一个样板方法**：N 个 repo 各加一个 `WithTx`；而 `NewXxxRepo(tx)`
  是本来就有的构造函数，零额外代码。

### 逐维度对比

| 维度 | 现状 `NewXxxRepo(tx)` | 提案 `WithTx(tx)` |
|------|----------------------|-------------------|
| 性能 | 相同（缓存 q，方法零 Use） | 相同 |
| 避免重建 | 做不到（本质绕不开） | **也做不到**（内部就是重建） |
| 构造细节收拢 | 分散在调用点 | 收进 repo（约定下用不上） |
| 与 content-center 一致 | ✅ | ❌ 破坏 |
| 样板代码 | 复用已有构造函数 | 每个 repo 多一个方法 |
| 字段命名 | `q`（躲开包名） | `query`（遮蔽包名，需小心） |

### 结论

WithTx 不是错，它是**同一范式（重建派）内的平移**——功能等价，纯风格偏好。拿
「每个 repo 多写一个 WithTx + 字段命名要小心 + 破坏老项目一致性」换「事务闭包里
`s.userRepo.WithTx(tx)` 读起来顺一点」，不划算。**保持 `NewXxxRepo(tx)`。**
