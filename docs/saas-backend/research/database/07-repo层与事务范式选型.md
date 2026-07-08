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
