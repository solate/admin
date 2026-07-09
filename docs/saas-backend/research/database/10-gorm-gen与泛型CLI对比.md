# gorm/gen vs 泛型 CLI：工具换不换，AI 编程视角

> 2026-07-09。本文回答两个问题：① 项目现用的 `gorm.io/gen`（v0.3.28）有没有新版、
> 要不要换成 GORM 官方新推的**泛型 CLI**（`gorm.io/cli/gorm`）？② **AI 编程场景下，
> 泛型 CLI 到底是有利还是有害？**
>
> 这是**工具层对比**，不重开数据库访问层选型（那是 [[02-数据库访问层选型调研]]）。
> 复用 [[04-schema优先与数据库优先]] 的真相源方向框架、[[08-GORM与原生SQL对比]] 的
> 「AI 手误挡编译期」核心洞察。结论只服务一个决策：**要不要从 gen 换到 CLI。**
>
> 对两个工具零基础也能读——第一节先教「它们分别是什么」，再对比。

## 结论速览

**保持 `gorm/gen`，暂不换泛型 CLI。**

新 CLI 不是「gen 的升级版」，是官方另起的**平行路线**（两者并存、可混用）。它在 AI 友好上
确有真优势（无状态设计、编译期类型更强、官方明确主打 AI），但对**本项目此刻**不划算——
gen 未废弃、DB-first 现状已定、迁移成本实打实，而且有一条反直觉的关键事实：**AI 训练语料里
gen 的写法远多于 CLI，所以「新工具」不等于「AI 写得更对」。**

| 维度 | gorm/gen（现用） | 泛型 CLI（新） | 谁更好 |
|------|-----------------|---------------|--------|
| 编译期类型安全 | ✅ 字段引用类型安全 | ✅ 泛型 + field helper，更强 | CLI 略胜 |
| AI 友好 | ✅ **语料海量、已验证** | ⚠️ 类型更优但**语料少** | 看维度（见第三节） |
| 迁移成本（本项目） | ✅ 零（已建成） | ❌ 重写生成器 + repo + 可能翻转真相源 | gen 胜 |

## 一、先搞懂两个工具分别是什么

### gorm/gen（项目现用，2021 生，泛型出现之前）

`gorm.io/gen` 是 GORM 官方 2021 年推出的代码生成器，那时 Go 还没有泛型。它的做法是：
**反射活库**（读真实数据库的表结构）→ 生成 model 结构体 + **一整套自己的查询层**。

本项目现状（`scripts/gen-from-db/main.go`）就是 DB-first 反射：`db.Migrator().GetTables()`
拿到所有表 → `g.GenerateModel(table)` → 生成到 `internal/dal/model` + `internal/dal/query`。
repo 里持有一个**有状态的 `q` 对象**，查询挂在它身上：

```go
type UserRepo struct {
    db *gorm.DB
    q  *query.Query          // 有状态，query.Use(db) 构造时绑死连接
}
func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{db: db, q: query.Use(db)} }

// 查询：从 r.q.User 起手，WithContext(ctx) 链式，字段引用 r.q.User.Status
r.q.User.WithContext(ctx).
    Where(r.q.User.Status.Eq(int16(status))).
    Order(r.q.User.CreatedAt.Desc()).
    FindByPage(offset, limit)
```

特征：**有状态 `q`**、`WithContext(ctx)` 起手、自成一套 `IXxxDo` 查询接口、DB-first 反射生成。

### 泛型 CLI（官方新推，2024+，基于 Go 泛型）

`gorm.io/cli/gorm` 是 GORM 官方在 Go 有了泛型（1.18+）之后另起的新工具。它**不是 gen 的
升级版**，是**平行的另一条路线**——官方文档专门有 `cli_vs_gen` 页并列对比两者。核心变化：
**无状态泛型入口 `gorm.G[T](db)`**，生成的东西更薄、更贴原生 GORM。它有**两个生成器**：

```bash
go install gorm.io/cli/gorm@latest
gorm gen -i ./examples -o ./generated                 # 默认：严格泛型
gorm gen -i ./examples -o ./generated --typed=false   # Standard API：可混 raw 条件
```

**生成器 1 — field helpers（从 model struct 生成类型安全字段）：**

```go
// 无状态 gorm.G[User]，字段用 generated.User.X，ctx 在末尾 .Find(ctx)
users, err := gorm.G[User](db).Where(generated.User.Age.Gt(18)).Find(ctx)

gorm.G[User](db).Where(generated.User.Name.Eq("alice")).
    Set(generated.User.Name.Set("jinzhu"), generated.User.Count.Incr(1)).
    Update(ctx)
```

**生成器 2 — interface-driven（在 Go 接口里用 SQL 模板注释，生成类型安全方法）：**

```go
type Query[T any] interface {
    // SELECT * FROM @@table WHERE id=@id       ← SQL 写在注释里，@@table/@id 是模板变量
    GetByID(id int) (T, error)
}
u, err := generated.Query[User](db).GetByID(ctx, 123)
```

**关联是一等公民（强类型 Create/Unlink/Delete）：**

```go
gorm.G[User](db).Where(generated.User.ID.Eq(1)).
    Set(generated.User.Pets.Create(generated.Pet.Name.Set("fido"))).
    Update(ctx)
```

特征：**无状态 `gorm.G[T]`**、`ctx` 在末尾、model-first（从 Go struct/interface 生成，不反射活库）、
SQL 模板接口、关联操作强类型。

### 概念对照表

| 维度 | gorm/gen（现用） | 泛型 CLI（新） |
|------|-----------------|---------------|
| 诞生年代 | 2021（泛型前） | 2024+（基于泛型） |
| 定位关系 | 老牌、维护中（v0.3.28 未废弃） | 平行新路线，非升级版 |
| 真相源方向 | **DB-first**（反射活库） | **model-first**（读 Go struct/interface） |
| 查询入口 | 有状态 `r.q.User` | 无状态 `gorm.G[User](db)` |
| ctx 位置 | `WithContext(ctx)` 起手 | `.Find(ctx)` 末尾 |
| 生成物 | model + 完整查询层 `IXxxDo` | field helpers + interface SQL 模板 |
| 返回风格 | `First() (*User, error)` | 泛型直接返回 `(T, error)` |
| 关联 | 用 GORM 原生 Preload/Association | 一等公民强类型 Create/Unlink/Delete |
| 错误处理 | 链尾方法返回 error | 操作方法直接返回 error |

## 二、同一个查询，两种写法

用本项目真实场景对照。注意 gen 的 `WithContext` 起手 vs CLI 的 `ctx` 末尾。

### GetByID

```go
// gen（现状）
func (r *UserRepo) GetByID(ctx context.Context, userID string) (*model.User, error) {
    return r.q.User.WithContext(ctx).
        Where(r.q.User.TenantID.Eq(tenantID)).
        Where(r.q.User.UserID.Eq(userID)).
        First()
}

// 泛型 CLI
user, err := gorm.G[User](db).
    Where(generated.User.TenantID.Eq(tenantID), generated.User.UserID.Eq(userID)).
    First(ctx)
```

### 动态过滤 + 分页 List

```go
// gen（现状）
query := r.q.User.WithContext(ctx).Where(r.q.User.TenantID.Eq(tenantID))
if nickname != "" {
    query = query.Where(r.q.User.Nickname.Like("%" + nickname + "%"))
}
if status != 0 {
    query = query.Where(r.q.User.Status.Eq(int16(status)))   // 类型跟着列走
}
total, _ := query.Count()
users, err := query.Order(r.q.User.CreatedAt.Desc()).Offset(offset).Limit(limit).Find()

// 泛型 CLI
q := gorm.G[User](db).Where(generated.User.TenantID.Eq(tenantID))
if nickname != "" {
    q = q.Where(generated.User.Nickname.Like("%" + nickname + "%"))
}
if status != 0 {
    q = q.Where(generated.User.Status.Eq(int16(status)))
}
users, err := q.Order(generated.User.CreatedAt.Desc()).Offset(offset).Limit(limit).Find(ctx)
```

### Update

```go
// gen（现状）
_, err := r.q.User.WithContext(ctx).
    Where(r.q.User.UserID.Eq(userID)).
    Update(r.q.User.Status, status)

// 泛型 CLI
_, err := gorm.G[User](db).
    Where(generated.User.UserID.Eq(userID)).
    Set(generated.User.Status.Set(status)).
    Update(ctx)
```

**诚实对照**：两者动态 WHERE 都是链式 `if` 拼条件，**难度基本一样**。CLI 少了一个有状态 `q`
对象、`gorm.G[User](db)` 一行自解释、错误直接返回；gen 的 `WithContext` 起手更啰嗦，但项目
已经写了 13 个 repo、全团队和 AI 都熟。差别是**风格**，不是能力级差。

## 三、AI 编程视角：泛型 CLI 是利还是弊

这是全文最关键的一节，直接回答你的追问。而且 GORM 官方恰好把新 CLI 明确定位成
**「更适合 AI coding」**——这个官方主张必须诚实评估，不能照单全收。

### 官方立场

[GORM CLI 官网](https://gorm.io/cli/) 原话（粗体是我加的）：

> "Everything compiles to plain Go that works with gorm.io/gorm, with a focus on **type safety—if
> it compiles, it works**—making the generated APIs more predictable and **developer-friendly for
> the AI coding**."

官方明确主打两个点：① 类型安全（编译即正确）；② AI 友好。那我们逐条验证。

### 利（真优势，诚实列出）

**① 无状态 `gorm.G[T](db)`，AI 不需理解生命周期。**

gen 的 `r.q.User` 是有状态对象，`query.Use(db)` 构造时绑死连接。AI 要理解「q 在 repo
结构体里、事务时要重建」（[[07-repo层与事务范式选型]] 的重建派）。CLI 的 `gorm.G[User](db)`
**每次从头调、无状态、一行自解释**，AI 不需要理解对象复用和连接管理，少一层心智负担。

**② 编译期类型安全 +「if it compiles, it works」，承接 [[08]] 核心洞察。**

[[08-GORM与原生SQL对比]] 第一节的核心洞察是：**AI 会犯概率性手误（字段名、类型），编译期
类型系统是安全网**。CLI 的泛型 `gorm.G[User]` + field helper `generated.User.Age.Gt(18)`
在类型安全上比 gen 的 `r.q.User.Age.Gt(18)` **稍强一点**（泛型返回值直接是 `T`，不经过
`IXxxDo` 接口一层；Update 用 `Set()` 强类型而非 gen 的 `Update(column, value)` 两参数）。
这条官方主张**在类型系统维度成立**。

**③ 关联/更新强类型，AI 少写错列名和 FK。**

CLI 把关联操作做成一等公民：`generated.User.Pets.Create(generated.Pet.Name.Set("fido"))`，
整条链都是强类型。gen 用原生 GORM `Association("Pets").Append(&pet)`，`"Pets"` 是字符串，
AI 容易拼错。这条 CLI 真赢。

**④ SQL 模板接口：AI 在注释里写 SQL，生成器兜底类型。**

CLI 的 interface-driven 查询让 AI 在 Go 注释里写 SQL（`// SELECT * FROM @@table WHERE id=@id`），
生成器把模板变量 `@id` 绑定到方法参数、`@@table` 替换成表名，最后生成类型安全方法。这比裸 SQL
字符串安全（AI 手误会被模板解析器抓住），比 gen 的「全用 DSL」多了一个「写 SQL」的选项。

### 弊（不回避，关键平衡点）

**① 生态年轻、语料少——「新 ≠ AI 更会写」（最重要的反直觉）。**

这是最容易被忽略、但对 AI 编程最致命的一条：**AI 训练数据里 gen 的 `r.q.User` 写法远多于
CLI 的 `gorm.G[T]`**。CLI 是 2024+ 才推的，公开代码库、StackOverflow、GitHub issue 里的
语料相比 gen（2021 生、生产用了 5 年）**少一个数量级**。

这意味着：即使 CLI 的类型系统理论上更优、更「适合 AI」，但**2026 的 AI 写 CLI 的正确率
短期内不一定比写 gen 高**——因为它见过的 CLI 代码太少，容易生成混合两种 API 的错误代码，
或者猜错 `gorm.G[T]` 的泛型约束。**「新工具」≠「AI 写得更对」。** 语料丰富度是 AI 生成
质量的**第一性原理**，这条在类型优势之前。

**② 两套生成器 + SQL 模板 DSL，是新的心智负担。**

CLI 有两个生成器（field helpers + interface SQL 模板），后者的 `@@table/@column/@param`
模板语法是一套新 DSL。AI 要学模板约定（`@@` 是表/列变量、`@` 是参数绑定、如何处理 JOIN），
可能生成错模板（比如 `@id` 写成 `@userId` 导致绑定失败）。gen 只有一套查询 DSL，AI 犯错面更窄。

**③ `--typed=false` 混用模式给了 AI「退回 raw string」的口子。**

CLI 的 Standard API（`--typed=false`）允许 `Where("name = ?", "x")` 和
`Where(generated.User.Age.Gt(18))` 混用。这个「灵活性」对 AI 是双刃剑——它可能图省事
直接生成 raw string 条件，反而削弱了 CLI 主打的类型安全卖点。gen 没有这个口子（要 raw 就
显式 `r.db.Raw()`，不会静默混进查询链）。

**④ model-first vs 本项目 DB-first 的方向冲突。**

CLI 是 model-first（从 Go struct 生成），本项目是 **DB-first**（`scripts/gen-from-db/main.go`
反射活库，[[04-schema优先与数据库优先]] 已定）。换 CLI 要么改成 model-first（推翻 04 的结论），
要么用 CLI 但反其道而行、仍然写个脚本反射 DB 生成 model struct 再喂给 CLI——都是额外摩擦。

### 结论：官方「更适合 AI」的主张，要拆开看

| 维度 | gen | CLI | 谁对 AI 更友好 |
|------|-----|-----|---------------|
| 编译期类型安全 | ✅ 强 | ✅ 更强（泛型 + 强类型关联） | CLI 胜 |
| **语料丰富度**（AI 训练数据） | ✅ **海量、5 年生产验证** | ❌ **少、2024+ 新物** | **gen 胜**（关键） |
| 状态管理心智 | ⚠️ 有状态 `q` | ✅ 无状态 `gorm.G[T]` | CLI 胜 |
| DSL 复杂度 | ✅ 单一查询 DSL | ⚠️ 两套生成器 + SQL 模板 | gen 胜 |
| 类型安全漏口 | ✅ 无（要 raw 显式） | ⚠️ `--typed=false` 混用 | gen 胜 |

**一句话**：官方主张在**类型系统维度成立**（CLI 编译期更强、无状态更简洁），但在
**语料丰富度维度当前不成立**。对 2026 的 AI 编程，gen 的「海量训练语料 + 已验证」
比 CLI 的「理论更优类型」**当前更稳**。CLI 真正对 AI 更友好，要等它语料积累起来（2028+？）。

**新工具的类型优势是真的，但 AI 写它的正确率短期内不一定更高**——这是对你追问最诚实的答案。

## 四、要不要从 gen 换到 CLI

**换的收益：**

- 无状态 `gorm.G[T]` 更简洁，少一层 `q` 生命周期心智
- 关联操作一等公民、强类型（gen 要退回原生 GORM 的字符串 Association）
- 官方未来主推方向——gen 进入维护状态，CLI 是新重心（长期看会积累语料、补生态）

**换的成本：**

- 重写 `scripts/gen-from-db/main.go` 生成器 + 重搭生成流程（`go generate` vs 现有 `make gen-db`）
- **真相源方向可能被迫从 DB-first 翻转到 model-first**（撞 [[04-schema优先与数据库优先]] 已定的方向）
- 重写现有 repo 的查询 API（backend-rbac 13 个 repo，backend 待建）
- `gorm.io/cli` 生态年轻、坑未知（见第五节）
- 与 content-center-backend / backend-rbac 不一致（老项目都是 gen）
- **AI 生成正确率短期内可能下降**（第三节：语料少）

**网上现状：没有「大家都转过去了」。** CLI 是 2024+ 新物，社区仍以 gen 为绝对主流，属
**早期采用阶段**。官方在推、尝鲜者在试，但远没到「gen 已过时、都该换 CLI」的程度。别被
「新工具」的光环误导——gen 仍是 2026 GORM 代码生成的稳健默认。

## 五、坑与注意事项

你明确问「有没有坑」，逐条列（都是换 CLI 才会踩的）：

1. **model-first vs DB-first 方向冲突（最大的坑）**：CLI 从 model struct 生成，本项目
   现状是反射活库。换 CLI 约等于改真相源方向，不是「换个生成器」这么简单——要么推翻 [[04]]，
   要么额外写反射脚本喂 struct 给 CLI。
2. **SQL 模板 DSL 学习曲线 + 生成时机**：interface-driven 的 `@@table/@column/@param` 是新语法；
   `go generate` 触发方式和现有 `make gen-db`（`go run ./scripts/gen-from-db/`）流程不同，要重搭。
3. **`--typed=false` 的类型安全漏口**：图灵活开了 Standard API，就等于给「raw string 条件」开门，
   削弱 CLI 主打的类型安全（第三节弊 ③）。要用就锁定默认严格泛型模式。
4. **需要 Go 1.18+ 泛型**：项目 Go 1.26，无问题；但作为前提列出（老项目若卡在 1.17 以下不能用）。
5. **生态年轻、插件兼容边界**：官方说泛型 API 可与传统 API、GORM 插件混用，但**边界要实测**——
   尤其本项目重度依赖的 `dbresolver`（读写分离）、`soft_delete`（软删除）在 CLI 泛型链下的行为。
6. **项目定制需重新验证等价物**：现有 gen 生成器注入了 `created_at/updated_at` 毫秒时间戳 tag、
   `deleted_at → soft_delete.DeletedAt` 类型、多租户字段——这些**定制在 CLI 下的等价写法要逐一确认**，
   不是开箱即得。

## 六、承认的 tradeoff（不回避）

**CLI 真正赢的地方（不否认）：**

- 无状态 `gorm.G[T]` 设计更干净、心智更轻
- 编译期类型更强（泛型直返、强类型关联、Set() 更新）
- 关联操作一等公民
- 官方未来方向，长期会补齐生态与语料
- 类型维度对 AI 更友好（「if it compiles, it works」）

**gen 真正赢的地方：**

- **海量 AI 训练语料 + 5 年生产验证**（对 AI 编程是第一性优势）
- DB-first 贴合本项目现状（[[04]] 已定）
- 已建成、零迁移成本、与老项目一致
- 单一查询 DSL、无 `--typed=false` 漏口、犯错面窄

**什么场景该换 CLI：** 全新项目 + 接受 model-first + 愿吃早期生态坑 + 重度依赖关联操作 +
团队欣赏无状态泛型风格。四五条同时成立，CLI 是不错的现代选择。

**对本项目：不换。** gen 未废弃、DB-first 已定、语料更利 AI、迁移成本高、与老项目一致。
等 CLI 生态和语料再成熟一两年（有更多生产案例、AI 语料补上来），下一个全新项目再考虑。

**结论不变：保持 `gorm/gen v0.3.28`。** 它是 2026 GORM 代码生成的稳健默认——不是因为它更「新」，
而是因为对「AI 主力写代码 + 已建成 + DB-first」的本项目，它此刻最稳。CLI 的类型优势是真的，
但「新工具」骗不过「AI 语料现实」：**AI 写它，短期内不一定更对。**

## 七、参考链接

- [GORM CLI 官网（AI 定位原话）](https://gorm.io/cli/)
- [GORM CLI vs Gen 官方对比](http://gorm.io/cli/cli_vs_gen.html)
- [GORM CLI — Field Helpers](https://gorm.io/cli/field_helpers.html)
- [GORM CLI — SQL Templates](http://gorm.io/cli/sql_templates.html)
- [GORM CLI — Workflow（两个生成器）](https://gorm.io/cli/workflow.html)
- [GORM — The Generics Way（泛型 API）](http://gorm.io/docs/the_generics_way.html)
- [go-gorm/cli — GitHub](https://github.com/go-gorm/cli)
- [go-gorm/gen — GitHub（现用工具）](https://github.com/go-gorm/gen)
- 交叉引用：[[02-数据库访问层选型调研]]、[[04-schema优先与数据库优先]]、[[08-GORM与原生SQL对比]]
