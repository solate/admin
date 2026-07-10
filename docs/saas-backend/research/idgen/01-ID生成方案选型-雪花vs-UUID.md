# ID 生成方案选型：雪花算法 vs UUID

> 本轮（2026-07）为 idgen 选型所做的留档。**站在 2026、新项目无历史包袱、部署库已确认
> PostgreSQL 18，主键该继续用雪花，还是换 UUIDv7？**
> 结论：**UUIDv7**（PG18 原生支持、索引友好、零机器协调、库内反而更省）。
> **已于 2026-07 在 backend/ 落地**：主键改用 PG18 原生 `uuid` 列 + `DEFAULT uuidv7()`（库端生成），
> `backend/pkg/utils/idgen` 换成 `uuid.NewV7()` 的应用层备用封装（实现在 `idgen.go`）。
> 日常写代码不必读；纠结「要不要换 ID 方案」时看这里。

## 结论速览

- **PG18 已确认，UUIDv7 是 2026 的推荐主键方向**：时间有序 → B-tree 索引友好、零机器协调、
  PG 原生 `uuidv7()` 可由数据库直接生成。这也是当前社区主流建议（「别再用随机 v4，需要分布式友好键就上 v7」）。
- **已于 2026-07 在 backend/ 落地**（`backend-rbac/` 不在本轮范围）。切换窗口正好最佳（业务表几乎为空、
  config_id 无外键、idgen 零业务调用），一次把列类型 + 迁移 + idgen 封装一起换掉，改动面极小。
- **「值不值」的关键纠正**：不是「花复杂度只换省一半长度」。真实的账是——**PG18 原生 `uuid` 类型
  存 UUIDv7，在磁盘和索引上（16 字节二进制整数比较）比现状 `varchar(20)` 雪花串（~21 字节文本比较）
  还更小更快**，同时删掉机器 ID 协调 / init panic / 时钟这一整套复杂度。代价只有前端 JSON 传输 /
  日志里每个 ID 多十几字节——对后台管理系统可忽略。「36 字符」只是它序列化给前端的样子，不是库内的样子。

## 背景与约束

老项目用 sony/sonyflake，主键是 64 位整数转成的数字字符串（`varchar(20)`）。新 backend 起初原样移植，
本轮（2026-07）切换为 UUIDv7。项目现有硬约束（`.claude/rules/dto-id-type.md`、`migration-data-convention.md`）：

1. **所有 ID 字段用 `string`**，禁止 `int64`（切换后仍成立：PG 原生 `uuid` 列由 gen 映射成 Go `string`）。
2. 主键现为 **PG18 原生 `uuid` + `DEFAULT uuidv7()`**（库端生成，插入不传 ID）；config_id 无外键，
   故 seed 不再硬编码 ID（对 `migration-data-convention` 规则 4 的合理豁免）。
3. 前端「超过安全整数会截断，故转 string」——这条约定的根因见第一节；换 UUIDv7 后 ID 天生是 string，隐患消失。

切换前探查确认锁定程度极低：仅 1 张演示表、2 行 seed、idgen 包零业务调用，业务目录全空——
**此刻切换成本最低**，越往后表越多成本越高（详见第六节）。

## 一、核心约束：JavaScript 的 2^53 精度上限（用户亲历）

这是选型里最实际的一条，也是「前端说 ID 太长、改成字符串」的技术根因。

**JavaScript 的 `Number` 是 IEEE 754 双精度浮点**，能精确表示的最大整数是
`Number.MAX_SAFE_INTEGER = 2^53 − 1 = 9007199254740991`（16 位）。而 sonyflake ID 是 64 位、
18-19 位十进制数，**远超 2^53**。后果是 JS 解析时**静默丢精度、不报错**：

```js
JSON.parse('{"id": 186429408913933113}').id
// → 186429408913933110   ← 末尾被改，ID 悄悄错掉
```

最坑在于无异常：ID 看着还在、只是尾数错，前端拿错 ID 去查详情/删除 → 404 或删错行。
标准解法就是用户当年的做法——**后端把 ID 序列化成 string**，JS 收到 `"186429408913933113"`
按字符串处理，一位不差。

**这条对选型的意义**（注意：这里比的是「前端 JSON 传输」这一层，不是数据库存储层，见第二节澄清）：

| | 雪花（数字） | UUID |
|---|---|---|
| 前端 JS 安全 | **必须转 string**，否则丢精度 | 天生字符串，无此问题 |
| JSON 传输长度 | 19 字符 | 36 字符 |
| 是否需前端配合 | 需后端统一转 string（或前端上 json-bigint） | 无需 |

> 关键轴：雪花被迫转 string 后，「短」在前端传输层只剩「19 vs 36 字符」的差距，而它的机器 ID
> 协调成本、碰撞隐患一分没省。用户这段经历实际是在**削弱雪花的核心卖点**。

Vue 层面补充：Vue 本身不碰此问题（不解析 JSON），风险在 `axios`/`fetch` 拿到响应后的
`JSON.parse` 那一步就丢精度。后端转 string 比前端全局上 `json-bigint` 简单得多。

## 二、成本到底在哪：数据库存储层对比（纠正「长度」直觉）

第一节的「19 vs 36 字符」**只在前端 JSON 传输层成立**。主键真正的成本大头在**数据库怎么存、
索引怎么比**，而这一层结论相反——PG18 原生 `uuid` 反而比现状 `varchar(20)` 雪花串更省更快：

| 方案 | PG 列类型 | 磁盘/索引占用 | 索引比较方式 |
|---|---|---|---|
| 现状：雪花转 string | `varchar(20)` | ~21 字节（含长度头） | 文本逐字节比较 |
| **UUIDv7 原生** | `uuid`（PG 原生类型） | **16 字节**（128 位整数） | 二进制整数比较，更快 |
| （对照）bigint | `bigint` | 8 字节 | 整数比较 |

原生 `uuid` 是 16 字节二进制，不是 36 字符文本——`varchar(36)` 存字符串才是 36 字节，
**PG 里存 UUID 应当用原生 `uuid` 类型，而非 `text`/`varchar`**。

> 一句话：**「36 字符」只是 UUID 序列化成 JSON 给前端时的样子，不是它在库里的样子。**
> 库内 16 字节 < 现状 varchar 雪花串 ~21 字节。所以命题不是「花复杂度换省一半长度」——
> 库内更小更快，「长」只体现在前端 JSON 传输 / 日志字节上。

## 三、UUID 版本澄清（RFC 9562，2024-05 发布）

一次性定义了 v6/v7/v8，**目前没有比 v8 更新的版本**。要点：

- **v7**：时间有序（48 位毫秒时间戳 + 随机）。**做主键就是它**。相对 v4 的核心改进：有序键落在
  B-tree 最右端（像自增一样），不像 v4 全随机那样频繁页分裂、拖慢写入、膨胀索引。已公开基准显示
  v7 主键索引比 v4 小约 **26–27%**、有序扫描快数倍；离 bigint 的差距只剩 16 vs 8 字节。
- **v8**：**自定义/实验版**，RFC 留给厂商自定义布局用。**不是 v7 的升级**，是「更自由」，
  非主键场景。
- 所以论主键，**v7 是终点**，不存在「更新更好」的版本。用户「记得有 v8 或更新」的印象即 v8，
  但它不是要走的方向。

## 四、PostgreSQL 原生支持：本项目已确认 PG18

- **PG 18（2025-09-25 正式发布）**：内置 `uuidv7()` 函数，无需扩展，可直接
  `id uuid PRIMARY KEY DEFAULT uuidv7()`，**数据库自己生成有序 UUID**。
- **本项目部署库已确认为 PG18**，可走这条最干净的原生路径——无需应用层生成、无需扩展。
- （背景）PG 17 及以下无内置 `uuidv7()`，需应用层生成或装 `pg_uuidv7` 扩展；
  原生 `gen_random_uuid()`（v4）在 PG13+ 内置。本项目不受此限。
- 依赖注意：现有 `google/uuid v1.3.0` **不支持 v7**，若走应用层生成需升到 v1.6.0+；它目前只被
  `internal/middleware/requestid.go` 用于 HTTP request-id（v4，那个场景没问题），未被 idgen 引用。
  走 PG18 `DEFAULT uuidv7()` 数据库生成路径则连这个升级都不需要。

## 五、换成 UUIDv7 的好处 / 坏处

**好处**
- **生成零协调**：砍掉 machine ID 整套逻辑 + 「300 台机器 50% 碰撞」隐患 + `MACHINE_ID` 兜底
  （见 `machine_id.go`）。任意机器/数据库直接生成不撞——这正是雪花最费劲、UUIDv7 碾压的点。
- **无需应用层生成**：PG18 下 `DEFAULT uuidv7()` 数据库直接生成，Go 侧连 idgen 包都可省掉。
- **时间有序 → B-tree 索引友好**：v7 相对 v4 的核心改进，索引小 ~26–27%、写入不页分裂。
- **库内更省**：原生 `uuid` 16 字节 < 现状 `varchar(20)` 雪花串 ~21 字节（见第二节）。
- **全局唯一**：跨库/跨服务合并数据不撞。
- **前端天然安全**：本就是字符串，无 2^53 精度问题。

**坏处**
- **前端 JSON / 日志字节变长**：36 字符 vs 雪花转 string 后的 19 字符。**注意这只是传输/日志层**——
  数据库存储反而更小（第二节）。URL、日志行、JSON 体积略增，对后台管理系统可忽略。
- **非数字**：若前端/对接方期望数字 ID 要改（但本项目 ID 本就是 string，影响小）。
- **迁移成本**：表类型、硬编码 ID、seed 需改（本项目此刻极低：1 DDL 行 + 2 seed 行 + `make gen-db`，见第六节）。

**关于「err / panic / 时钟 / 机器」**（回答「google 这种是否可以不处理错误」）：
- 应用层用 `google/uuid` 时 `uuid.NewV7()` 返回 `(UUID, error)`，但该 err **仅在读系统随机源失败时**
  出现（几乎不可能），**没有雪花的「机器 ID 协调 / 时钟回拨 / 174 年上限」这些概念**。
- 走 PG18 `DEFAULT uuidv7()` 由数据库生成则更彻底：**Go 侧连生成代码都不写**，自然没有
  `init()` / panic / `machine_id.go` 那一整套——现状 idgen 里为雪花兜底的复杂度整体消失。

## 六、当前锁定程度（支撑「新项目可随便改」）

探查结论：整仓库依赖「数字字符串主键」的地方仅——
- `migrations/000001_ddl_init_schema.up.sql` 的 `config_id VARCHAR(20) PRIMARY KEY` 一行（演示表）。
- `migrations/000002_data_base_config.up.sql` 的 2 行硬编码数字 ID。
- `backend/pkg/utils/idgen/` 包自身，**无任何业务调用点**。
- `internal/dal/model/system_config.gen.go` 的 `ConfigID string`（database-first 反射生成，
  改 DDL 后 `make gen-db` 自动跟着变，手工改动为零）。

service/handler/repository/dto/rbac 目录全空，业务代码尚不存在。**切换窗口现在最佳**，越往后越贵。

## 七、决策与待办

- **推荐方向**：UUIDv7。PG18 已确认，可走 `DEFAULT uuidv7()` 原生路径——索引友好、零机器协调、
  库内更省、删掉 idgen 一整套复杂度。这是 2026 的主流建议。
- **已于 2026-07 在 backend/ 落地**（`backend-rbac/` 不在本轮范围）：
  1. 主键列类型改为 **PG18 原生 `uuid` + `DEFAULT uuidv7()`**（库内 16 字节、库自生成、Go 侧 idgen 降级为备用封装）。
  2. 改动面：`000001` 建表列类型、`000002` seed 不再硬编码 ID（由 DB 生成）、重跑 `make gen-db`（model 自动重生成）；
     前端/DTO 仍按 string 传输（本就 string，无需前端改）。
  3. `backend/pkg/utils/idgen` 改为 `google/uuid` v1.6.0 的 `uuid.NewV7()` 备用封装（主路径是 DB 生成，
     仅在应用层需预生成 ID 的场景才调用）。

---

## 参考

- [RFC 9562 — UUIDs (IETF, 2024-05)](https://datatracker.ietf.org/doc/rfc9562/)
- [PostgreSQL 18 Release Notes](https://www.postgresql.org/docs/current/release-18.html)
- [Postgres 18 UUIDv7 主键实践指南（含 26–27% 索引缩小基准）— nerdleveltech](https://nerdleveltech.com/postgres-18-uuidv7-primary-keys)
- [PG18 UUIDv4 vs v7 深度对比 — credativ](https://www.credativ.de/en/blog/postgresql-en/a-deeper-look-at-old-uuidv4-vs-new-uuidv7-in-postgresql-18/)
- [Stop Using Random UUIDs as Primary Keys — devops-daily](https://devops-daily.com/posts/postgres-18-uuidv7-primary-keys)
- [PG 存 UUID 用原生 16 字节类型而非 varchar（2026）— techearl](https://techearl.com/store-uuid-postgresql)
- [UUID vs CHAR/VARCHAR 存储对比 — StackOverflow](https://stackoverflow.com/questions/32189129/performance-difference-between-uuid-char-and-varchar-in-postgresql-table)
- [google/uuid #148 — UUIDv7 单调性 / NewV7 行为](https://github.com/google/uuid/issues/148)
- [sony/sonyflake](https://github.com/sony/sonyflake)
- [Number.MAX_SAFE_INTEGER — MDN](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Number/MAX_SAFE_INTEGER)

---

**最后更新**：2026-07-10
