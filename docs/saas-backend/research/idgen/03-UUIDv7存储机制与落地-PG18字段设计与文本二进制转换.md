# UUIDv7 存储机制与落地：PG18 字段设计与「文本↔16 字节」自动转换

> 本文是 ID 选型系列第三篇，承接 [01](01-ID生成方案选型-雪花vs-UUID.md)（要不要换）与
> [02](02-UUIDv7落地细节-前端影响与生成方式.md)（换的话四个担忧拆解）。
> 01/02 都没展开一个最容易误解的点——**「字符串怎么就变成数据库里的 16 字节？」**
> 本文回答：**PG 的 `uuid` 类型「对外是文本、对内是 16 字节二进制」，转换由 PostgreSQL 自动完成，
> 应用层和前端从头到尾只看到字符串，不需要也不应该手动转换。**
> 附带回答 PG18 升级的许可证 / 版本差异担忧，以及字段该怎么设计、返回怎么返回。
> 结论仍与 01 一致：**UUIDv7，且已于 2026-07 在 backend/ 落地**，本文只把「怎么存、怎么返回」这条链讲透。
> 日常写代码不必读；真要落地 UUIDv7 列时看这里。

## 结论速览

- **PG18 升级无许可证风险**：PostgreSQL 一直是宽松的 PostgreSQL License（BSD/MIT 风格），
  PG18 未变更。它不属于 MongoDB/Redis/Elastic/HashiCorp 那批闭源化项目——由社区基金会维护，
  没有单一公司能改授权。
- **字段设计只有一句**：主键列写原生 `uuid` 类型，**不要写 `varchar(36)`**。`uuid` → 磁盘 16 字节、
  二进制比较；`varchar(36)` → 37 字节、文本比较，是最差组合。
- **「字符串变 16 字节」是 PG 内部自动做的**：INSERT 传入 36 字符文本 → PG 压成 16 字节存磁盘/索引 →
  SELECT 时格式化回 36 字符文本。应用层进出都是字符串，**省空间是 PG 白送的，无需写转换代码**。
- **Go / 前端零心智负担**：GORM 字段保持 `string` + 列类型标 `uuid`，驱动自动 Scan 成字符串；
  API `json` 输出字符串，前端 axios 收到就是 string，落进现有 string 处理路径（呼应 02 第一节）。

## 背景

用户问了三个连着的问题：PG 升 18 有没有坑（许可证 / 版本差异）？要存 UUIDv7 字段怎么设计、
返回怎么返回？以及最核心的一个直觉困惑——**「怎么直接从字符串存在数据库里，就变成了 16 字节数字？」**

这个困惑很典型：以为要在应用层把 `"0192f8c2-..."` 手动转成 16 字节再存。**其实不用**——
`uuid` 是 PG 的内建类型，文本与二进制的互转是类型自带行为。下面逐条拆开。

## 一、PG18 升级：许可证与版本差异

**许可证：没变，可放心。** PostgreSQL 核心始终是 **PostgreSQL License**（一种 BSD/MIT 风格的宽松开源
许可），PG18 没有任何授权变更。担心「开源证书换了」通常源于近年 MongoDB（SSPL）、Redis、Elastic、
HashiCorp 的闭源化事件——**PostgreSQL 不在其列**，它由 PostgreSQL Global Development Group / 社区
基金会维护，无单一商业公司能单方面改授权。这正是它长期作为默认数据库选型的底气之一。

**版本差异 / 升级注意（均为 PG 通用常识，非 18 特有，除标注外）：**

- **跨大版本升级走 `pg_upgrade` 或 dump/restore**，不能直接替换二进制目录。所有 PG 大版本升级都如此。
- **`uuidv7()` 是 PG18 才有的函数**（本文关键）。PG17 及以下没有；SQL 里写 `DEFAULT uuidv7()`
  拿到旧库会直接报错——这就是「绑不绑版本」的分界（见第三节 A/B 路径，详见 02 第三节）。
- **跨数据库引擎不通用**：`uuidv7()` 是 PG 专有函数名，MySQL 8 是 `UUID_TO_BIN()` 一套。若日后可能
  换引擎，就别用库端 `DEFAULT`，改应用层生成（照样可移植，见第三节）。
- 其余 PG18 大特性（异步 I/O、虚拟生成列、B-tree skip scan 等）是增强，不影响本 ID 用法。

## 二、字段怎么设计：列类型写 `uuid`，别写 `varchar`

```sql
-- 方式 A：库端自动生成（绑 PG18，因为 uuidv7() 仅 PG18+）
CREATE TABLE users (
    user_id  uuid PRIMARY KEY DEFAULT uuidv7(),
    tenant_id uuid NOT NULL,
    ...
);

-- 方式 B：应用层生成后 INSERT（不绑版本，uuid 列类型 PG13+ 都支持）
CREATE TABLE users (
    user_id  uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    ...
);
```

关键就一句：**列类型写 `uuid`，不写 `varchar(36)`**。

| 列类型 | 磁盘/行 | 索引比较 | 评价 |
|---|---|---|---|
| `uuid`（推荐） | **16 字节**（128 位整数） | 二进制整数比较，快 | ✅ 存 UUID 的正确姿势 |
| `varchar(36)` | ~37 字节 | 文本逐字节比较 | ❌ 最差组合，浪费一倍空间 |
| （对照）现状 `varchar(20)` 雪花串 | ~21 字节 | 文本逐字节比较 | 现状 |

hex 字符串每字节只编码 4 bit 信息，36 字符文本存 128 bit 数据要占 ~37 字节，浪费近一倍——
所以**在 PG 里存 UUID 一律用原生 `uuid` 类型**（02 第二节已量化，此处给字段设计落点）。

### PG18 以下会不会出问题？——列仍是 `uuid`，只去掉 DEFAULT，绝不退 `varchar`

一个高频误区：以为「PG17 及以下用不了 UUIDv7，就得把字段退回 `varchar(36)`」。**不对。**
先把两个东西分清——它们的版本要求完全不同：

| 东西 | 起始版本 | PG17 及以下 |
|---|---|---|
| `uuid` **列类型** | 远早于 18（PG 很早就内建） | ✅ 正常，照样 16 字节二进制存储 |
| `gen_random_uuid()`（v4 生成函数） | PG13+ | ✅ 有 |
| `uuidv7()` **生成函数** | **仅 PG18+** | ❌ 报错 `function uuidv7() does not exist` |

所以在旧库上**唯一会炸的是这一行**：

```sql
-- ❌ PG17 报错：function uuidv7() does not exist
user_id uuid PRIMARY KEY DEFAULT uuidv7()
```

**炸的是 `DEFAULT uuidv7()` 这个默认值表达式，不是 `uuid` 列类型本身。** 降级写法只需去掉
DEFAULT、改由应用层生成——列类型一直是 `uuid`：

```sql
-- ✅ PG13+ 都能跑：列仍是原生 uuid，只是不让库自动生成
user_id uuid PRIMARY KEY
```

```go
// 应用层生成后赋值（google/uuid v1.6.0+），即 02 第三节的「路径 B」
id, _ := uuid.NewV7()
user.UserID = id.String()
```

这样磁盘仍是 16 字节、二进制比较，该省的空间一点没丢；ID 由 Go 生成、不依赖 `uuidv7()` 函数，
**不绑 PG 版本**（呼应 02 第三节「应用层生成不绑版本」、用户「换低版本 DB 也不影响」的直觉正确）。

> 结论：**降级 PG 版本 → 把 `DEFAULT uuidv7()` 去掉、改应用层生成即可，列类型始终是 `uuid`，
> 永远不用退回 `varchar`。** `varchar(36)` 是最差组合（37B + 文本比较），只有极端异构库
> （某个根本不支持 `uuid` 类型的数据库）才被迫用——PostgreSQL 任何在用版本都不属于此列。

## 三、核心：字符串怎么「变成」16 字节？——PG 自动转，应用层无感

这是最容易卡住的直觉。`uuid` 类型的行为可以一句话概括：

> **对外（输入 / 输出）永远是 36 字符文本；对内（磁盘 / 索引）才是 16 字节二进制。
> 两者互转由 PostgreSQL 自动完成，应用层完全碰不到 16 字节。**

数据流：

```
应用层 INSERT 传入字符串  "0192f8c2-7c3e-7xxx-xxxx-xxxxxxxxxxxx"
        │  PG 解析该文本，压缩成 16 字节二进制
        ▼
磁盘 / 索引：  16 字节  （128 位整数，二进制比较）
        │  SELECT 时 PG 把 16 字节格式化回 36 字符文本
        ▼
应用层收到字符串  "0192f8c2-7c3e-7xxx-xxxx-xxxxxxxxxxxx"
```

要点：

- **你 INSERT 的是字符串**（`'0192f8c2-...'`），不是自己转好的字节。
- **你 SELECT 出来的也是字符串**，格式完全一样。
- **「16 字节」只发生在 PG 内部**，是它对 `uuid` 类型的存储优化。省空间、加快索引比较，都是
  **PG 白送的，不用你写一行转换代码**。

和现状对比看得最清楚：

| | 现状 `varchar(20)` 雪花串 | 换 `uuid` 存 v7 |
|---|---|---|
| 应用层传入 | 文本 | 文本 |
| **磁盘存储** | **文本原样（~21B）** | **二进制 16B（PG 自动压）** |
| 应用层读出 | 文本 | 文本 |

也就是说：**唯一变化是「中间那层存储从文本变二进制」，进出口都还是字符串。** 你现在担心的
「手动把字符串转 16 位数字」这一步根本不存在——那是类型内部的事。

> 补充：如果列错用了 `varchar(36)`，PG 就**不做**这层压缩，老老实实按文本存 37 字节。所以「省空间」
> 的前提是列类型必须是 `uuid`（呼应第二节）。

## 四、Go / GORM 落地与返回

因为 `uuid` 对外是文本，Go 侧**保持 ID 为 `string`** 即可，与项目现有约束
（[`.claude/rules/dto-id-type.md`](../../../.claude/rules/dto-id-type.md)：所有 ID 用 string）天然兼容。

```go
// GORM 模型：字段 string，列类型标 uuid
type User struct {
    // 方式 A：库端生成，加 default:uuidv7()（绑 PG18）
    UserID string `gorm:"column:user_id;type:uuid;default:uuidv7()" json:"user_id"`
    // 方式 B：去掉 default，在 service 层或 BeforeCreate 钩子用 uuid.NewV7() 赋值（不绑版本）
}
```

- **数据库 → Go**：pgx / lib/pq 驱动把 `uuid` 列读成 36 字符字符串，直接 Scan 进 `UserID string`。
  Go 侧不接触 16 字节。
- **Go → 前端**：`json:"user_id"` 输出 `"0192f8c2-..."`；前端 axios 拿到即 string，落进现有 string
  处理路径，**零改动**（详见 02 第一节：两个前端都全程按 string 处理 ID）。
- **与现有赋值习惯对比**：现状 service 层 `idgen.GenerateUUID()` 赋值 → 换 v7 方式 B 就是
  `uuid.NewV7()` 赋值，习惯一致；方式 A 连赋值都省，由库填。（生成方式 A/B 的取舍见 02 第三、四节。）

**database-first 注意**：本项目 model 由 `make gen-db` 反射活库生成
（[migration-data-convention.md](../../../.claude/rules/migration-data-convention.md) 规则 9）。
改 DDL 列类型为 `uuid` 后重跑 `make gen-db`，`*.gen.go` 里字段会自动更新，不要手改。

## 五、为什么 v7 不用处理错误？要不要封装成包？

用户的疑问：现状 idgen 里 sonyflake 的 `NextID()` 要处理 error、还有 init panic，换 v7 是不是也一样？

### 为什么可以不处理 error

`google/uuid`（v1.6.0+）生成 v7 的签名是 `func NewV7() (UUID, error)`，但那个 error
**只在读系统随机源（`crypto/rand`）失败时**才返回——在正常运行的操作系统上几乎不可能发生。
这与现状 idgen 里 sonyflake 的判断是**同一类**：唯一的 error 实际不可达，所以可以用
「无 error 的对外 API + 内部 panic 兜底」这种 fail-fast 写法。

换句话说，**换 v7 没有改变现状 idgen 的错误哲学**，只是把「实际不可达的那个 error」从
「雪花的机器 ID 协调 / 时钟回拨 / 174 年上限」换成了「v7 的随机源读取失败」——而后者比前者
更不可能触发。库甚至自带官方助手把这件事写死：

```go
// uuid.Must：出错即 panic，正是「不处理 error、失败即 fail-fast」的官方写法
id := uuid.Must(uuid.NewV7()).String()
```

**顺带回答「有序性靠不靠谱」**：`google/uuid` v1.6.0 起，对同一毫秒内的多次生成用
**mutex 保证单调递增**（monotonicity），批量生成的 ID 严格有序，不需要自己加锁——
这正是做主键要的性质，比现状还省心。

### 要不要封装成包？——看走哪条生成路径

- **走方式 A（库 `DEFAULT uuidv7()` 自动生成）**：Go 侧根本不生成 ID，**不需要任何包**，
  现状 `idgen` 整个删掉即可。
- **走方式 B（应用层生成）**：**建议保留现有 `idgen` 包的对外 API 不变**，只换内部实现：

  ```go
  package idgen

  import "github.com/google/uuid"

  // GenerateUUID 对外签名不变（仍返回 string、无 error），调用方（service 层）零改动。
  func GenerateUUID() string {
      return uuid.Must(uuid.NewV7()).String()
  }
  ```

  这样做的价值：**调用方一行都不用改**（还是 `idgen.GenerateUUID()`），同时
  `machine_id.go`（3 级机器 ID fallback）、`init()` 的 panic 分支、`epoch` 那一整套
  sonyflake 兜底代码**全部删除**——包从「需要理解机器 ID / 时钟 / init」退化成一个单行函数。

> 一句话：**封不封装取决于生成路径。走库自动生成就不用包；走应用层就复用现有 idgen 的壳、
> 只换芯（sonyflake → `uuid.Must(NewV7())`），对外无感、内部大幅变简。**

### 但如果是全新起步——直接裸调，别造包（诚实结论）

上面「保留 idgen 壳」是**因为本项目已经有这个包和 `GenerateUUID()` 调用约定**，换芯比删包全局改
调用点省事——这是「顺水推舟」，不是「一行值得封装」。要分清：

- **全新起步 / 没有历史调用点 → 直接裸调 `google/uuid`，不必造包**。`uuid` 库本身就是那层封装，
  在用到的地方写 `uuid.Must(uuid.NewV7()).String()` 即可，再包一层 `idgen` 是多余（YAGNI）。
- **单论「一行」不值得封装**：坦白说，封装的价值是「收口 + 命名 + 兼容既有调用」，**不是「藏复杂度」**
  ——复杂度已经随 sonyflake 一起消失了。sonyflake 值得封装是因为它是「一套机制」（机器 ID / 时钟 /
  init panic）；v7 塌缩成一行后，这个理由不复存在。
- **两派做法都存在、都不算错**：裸调派（到处直接 `uuid.NewString()` / `NewV7()`）与薄封装派
  （`idgen` / 可注入生成器）。薄封装最硬的理由是「测试时替换成确定性 ID」，但本项目 rule 明确
  **不定义 interface、不 mock、靠集成测试**（[`repo-transaction-convention.md`](../../../.claude/rules/repo-transaction-convention.md)），
  这条理由在本项目被削弱。

> 一句话（收束）：**别为「一行」造包。全新起步直接裸调 `google/uuid` 即可；本项目留着 idgen 只是
> 兼容既有调用的顺水推舟，不是「一行值得封装」。走库端 `DEFAULT uuidv7()` 那条路则连这一行都没有。**

## 六、决策与落地范围

- **方案**：UUIDv7（与 01/02 一致）。字段用原生 `uuid`、进出走字符串、GORM 字段保持 string。
- **已于 2026-07 在 backend/ 落地**：`000001` 建表列 `config_id` 改为 `uuid PRIMARY KEY DEFAULT uuidv7()`
  （库端生成）、`000002` 种子去掉硬编码 config_id 由 `DEFAULT` 填、`make gen-db` 重生成 model
  （`ConfigID string`，靠 gen 的 `uuid→string` 类型映射）；`backend/pkg/utils/idgen` 换成
  `uuid.NewV7()` 应用层备用封装。前端 / DTO 仍按 string，无需前端改。
- **范围仅 backend/**：`backend-rbac/` 不在本轮范围（独立 idgen、多处调用点，未动）。

---

## 参考

- [PostgreSQL License（官方）](https://www.postgresql.org/about/licence/)
- [PostgreSQL 18 Release Notes](https://www.postgresql.org/docs/current/release-18.html)
- [PostgreSQL 文档：UUID 类型](https://www.postgresql.org/docs/current/datatype-uuid.html)
- [PG 存 UUID 用原生 16 字节类型而非 varchar（2026）— techearl](https://techearl.com/store-uuid-postgresql)
- [PostgreSQL: 存 UUID 用 uuid 而非 text/varchar — StackOverflow](https://stackoverflow.com/questions/33836749/postgresql-using-uuid-vs-text-as-primary-key)
- [UUID vs CHAR/VARCHAR 存储对比 — StackOverflow](https://stackoverflow.com/questions/32189129/performance-difference-between-uuid-char-and-varchar-in-postgresql-table)
- [Postgres 18 UUIDv7 主键实践指南 — nerdleveltech](https://nerdleveltech.com/postgres-18-uuidv7-primary-keys)

---

**最后更新**：2026-07-10
