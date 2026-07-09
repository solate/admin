# schema 优先 vs 数据库优先

> 本轮（2026-07）承接 [02-数据库访问层选型调研.md](./02-数据库访问层选型调研.md) 与 [03-迁移工具与数据初始化方案.md](./03-迁移工具与数据初始化方案.md)。02 推荐的是 **schema 优先（Go struct 真相源）** 方向，本轮把它与 **数据库优先（SQL/DB 真相源）** 摊开对比，回答「数据库优先是什么情况、有没有可讨论的地方」。

## 一、背景与本轮问题

02/03 两份文档推荐的路线都假定了 **schema 优先**——真相源是 Go struct，数据库跟着走。但有一个隐藏的事实没点破：

**老项目 A 的 `gen-db`（gorm/gen 读活库反射生成 model）本来就是 database-first（数据库优先）**。所以「数据库优先」对本项目不是假设，而是**现状**。02 推荐的 Go struct 优先反而是要**掉头**。

这引出本轮要回答的核心问题：
1. 数据库优先到底是什么样的工作流？
2. 老项目那一堆问题（3/12/57 三源 drift）是不是 database-first 的锅？
3. 数据库优先有没有合理的使用场景、值得讨论的优势？

本文把两极摊开对比，呈现取舍，**不替用户拍板**（02/03 保持不动，此处作为补充视角）。

## 二、两条路线的本质区别

| 维度 | **schema 优先**（代码 → DB） | **数据库优先**（DB → 代码） |
|---|---|---|
| 真相源 | Go 结构体（手写或 ent schema） | SQL DDL 文件 / 实际数据库结构 |
| 方向 | struct → 生成迁移 SQL → DB 执行 | SQL → 建库/迁移 → 生成 Go model |
| 你只需关注 | **Go**（DB 是派生物、自动跟随） | **SQL**（Go 是派生物、重跑 codegen） |
| 改 schema 后的动作 | 改 struct → 生成迁移 | 改 SQL → apply → **重跑 codegen** |
| codegen 依赖 | DB 状态（迁移 apply 后） | DB 状态（迁移/手建后） |
| 代表工具 | ent、GORM + Atlas（struct 模式） | gorm/gen（现状）、Bob、sqlc |
| 适合的团队偏好 | 想只碰 Go、AI 补全驱动 | 想直接掌控 SQL、DBA 思维 |

**关键差异**：schema 优先让你「活在 Go 里」，数据库优先让你「活在 SQL 里」。前者 Go 是权威、DB 追随；后者 DB 是权威、Go 追随。

## 三、关键澄清：老项目的乱不是 database-first 的错

03 文档揭示的老项目病根——**3 vs 12 vs 57 三源 drift**（migrations / dev_schema.sql / 反射生成的 model 互相打架）——容易让人误以为「数据库优先 = 必然混乱」。必须澄清：

**病根是「3 个真相源并存」，不是 database-first 方向本身。** 如果只保留一个 canonical SQL schema 作为唯一真相源，数据库优先一样干净：

- **正面样本**：把 `dev_schema.sql`（12 表完整版）定为唯一 DDL 真相源，dev 环境跑它建库，`gen-db` 从库生成 model，生产环境用 Atlas 从这份 SQL diff 出版本化迁移 → **单一真相源，无 drift**。
- **老项目的错**：迁移里建 3 表、dev_schema.sql 建 12 表、生产手工建几十表，**三者互相矛盾且都不完整** → drift 是必然。这跟方向无关，是管理失控。

结论：**database-first 只要遵守「单一 canonical schema」原则，不比 schema-first 更容易乱**。两条路线都能干净，也都能搞乱（后者的乱是「struct 改了忘生成迁移」）。

## 四、值得讨论的三个真问题

### 问题 1：database-first 更贴合 PostgreSQL 原生特性（02 约束 5）

02 文档的约束 5 是「**PostgreSQL 优先——不需要跨库抽象，要能用 PG 原生特性（jsonb / array / upsert / RETURNING / CTE）**」。这条约束其实**内含张力**，因为：

**GORM struct tag 表达 PG 高级 DDL 的天花板**：

- **部分索引**（`WHERE` 条件）：`gorm:"index:,where:status='active'"` 能做简单的，复杂表达式写不进去。
- **GIN/GiST 索引**（jsonb/全文搜索）：`gorm:"index:,type:gin"` 勉强能用，`gin(jsonb_path_ops)` 之类的选项没法表达。
- **Check 约束**：`gorm:"check:age > 0"` 基础约束可以，但复杂的多列约束、引用其他表的约束无法表达。
- **Generated column**（PG 12+）：`GENERATED ALWAYS AS (...)` 没有对应的 struct tag。
- **表分区**（range / list / hash）：完全无法用 struct tag 描述。
- **Extension / 自定义类型**（如 `postgis` 几何类型、`ltree` 层次结构）：需要自定义 scanner/valuer，但 DDL 层面的 `CREATE EXTENSION` / `CREATE TYPE` 仍要手写 SQL。

这意味着：如果你真要「用满 PG」，schema-first（GORM struct）会反复撞天花板，最后靠**手改生成的迁移 SQL 打补丁**（正是你说的「偶尔字段不太一样要手改」）——这就破坏了「只碰 Go、DB 自动跟」的初衷。

**直接写 SQL DDL = 100% PG 满血**：

```sql
CREATE TABLE events (
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    data JSONB NOT NULL,
    tags TEXT[],
    search_vector tsvector GENERATED ALWAYS AS (to_tsvector('english', data->>'title')) STORED,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    CHECK (jsonb_typeof(data) = 'object')
) PARTITION BY RANGE (created_at);

CREATE INDEX idx_events_data_gin ON events USING GIN (data jsonb_path_ops);
CREATE INDEX idx_events_search ON events USING GIN (search_vector);
CREATE INDEX idx_events_tags ON events USING GIN (tags);
CREATE INDEX idx_active_events ON events (event_id) WHERE (data->>'status' = 'active');
```

上面这段 DDL **在 GORM struct 里根本表达不出来**。database-first 直接写、直接 apply，不用跟 ORM 抽象层搏斗。

**结论**：就「PG 优先」这条约束，**database-first 反而更强**——它不经过 ORM 的抽象层损耗，SQL 能写什么就能用什么。

### 问题 2：Atlas 是双向的——数据库优先也能自动 diff 迁移

很多人以为 Atlas 只配 Go struct 用，实际上 **Atlas 的原生模式就是 SQL/HCL schema 当真相源**（declarative 工作流）。这打开了一条优雅的中间路线：

**中间路线：SQL schema + Atlas + gorm/gen**

1. 维护一份 **SQL schema 文件**（或 Atlas HCL）作为唯一真相源，里面是你期望的完整 DDL。
2. `atlas migrate diff --to file://schema.sql` → Atlas 自动对比当前 DB，**生成版本化迁移 SQL**（你不用手写 ALTER）。
3. `atlas migrate apply` → 执行迁移，DB 更新到期望状态。
4. `make gen-db` → gorm/gen 从更新后的 DB 生成 model。

**收益**：
- 既保留你熟悉的 **gorm/gen 数据库优先流**（Go 是生成物、不需要手写 model）
- 又补齐了 **自动迁移 + 防 drift + lint**（不用手写 ALTER、Atlas 检测漂移、CI lint 拦截破坏性变更）
- 还能 **100% 用满 PG**（SQL 是你自己写的，没有 struct tag 天花板）

对比 02 推荐的 GORM + Atlas（struct 模式），这条路线只是把真相源从「Go struct」换成「SQL schema 文件」，**Atlas 和 gen 的部分完全一样**。你甚至可以把这两条路线看成 Atlas 的「两种配置」——一个吃 struct、一个吃 SQL，中间引擎是同一个。

### 问题 3：代价——Go 变成下游，「只碰 Go」的梦破了

database-first 的核心代价是：**每次改 schema 都要重跑 codegen，Go 变成下游**。

- schema 优先：改 struct → 生成迁移 → apply，**全程在 Go 里**，AI 补全 Go 代码即可驱动整个流程。
- 数据库优先：改 SQL → apply → **重跑 `gen-db`**，生成的 model 才更新。AI 补全 SQL（或你手写 SQL）→ 再跑一步生成。

这破坏了 ent 那种「只关注 Go、DB 自动跟、AI 一把梭」的体验——后者正是 schema-first 的核心卖点。如果你欣赏的就是这点，那 database-first 直接 pass。

但如果你本来就习惯「先设计 schema、再写代码」的 DBA 思维，或者你的团队里有专门的 DBA 负责 DDL 设计，database-first 反而更自然——**SQL 是你的设计语言，Go 只是实现细节**。

## 五、取舍矩阵

| 你更看重 | 推荐方向 | 代表路线 | 核心优势 | 核心代价 |
|---|---|---|---|---|
| 只碰 Go / DB 自动跟 / AI 补全驱动 / 概念统一 | **schema 优先** | ent / GORM + Atlas（struct） | 全程在 Go、AI 友好、单一心智模型 | struct tag 表达 PG 高级 DDL 有天花板、遇到需手改迁移 |
| SQL 完全掌控 / 用满 PG / DBA 思维 / 贴合现有 gorm/gen | **数据库优先** | SQL schema + Atlas + gorm/gen / Bob + Atlas | 100% PG 满血、SQL 即设计、不绕 ORM 抽象 | 每次改 schema 要重跑 codegen、Go 是下游、AI 补全 SQL 没有补全 Go 顺 |

**一句话总结**：
- schema 优先 = 「我是 Go 开发者，数据库只是存储细节」
- 数据库优先 = 「我是设计 schema 的人（或 DBA），Go 只是实现层」

两者**同等合法**，取决于你的团队角色分工、技能偏好、对「AI 补全驱动」的依赖度。

## 六、与 02 的关系（说明，不改 02）

02 文档把「Go struct 优先」讲得像唯一正解，实际上是因为当时只讨论了 schema-first 阵营（ent / GORM+Atlas struct / Bob 都被归为这一侧）。**本文补充另一极：数据库优先是同等合法的另一条路线，在 PG 维度甚至更优。**

两条路线在 02 的 7 条约束上各有千秋：
- **约束 1（比 ent 轻）**：database-first（SQL + Atlas + gorm/gen）三件套概念量 ≈ ent，但 SQL 是熟悉的、不是 ent schema DSL。
- **约束 5（PG 优先）**：database-first 完胜（100% 满血 vs struct tag 天花板）。
- **约束 6（迁移自动化）**：两者都能用 Atlas 自动 diff，打平。
- **约束 4（AI 友好）**：schema-first 胜（AI 补全 Go struct 比补全 SQL DDL 顺）。

**关键洞察**：两条路线的迁移引擎可以是同一个（Atlas），**差别只在真相源方向**——struct 还是 SQL。02 选了前者，本文呈现后者的完整情况，由你根据团队实际定夺。

## 七、承认的 tradeoff（不回避）

- **database-first 每次改 schema 要重跑 codegen**，多一步手动操作（或 Makefile target），不如 schema-first 的「改 struct 即触发一切」丝滑。
- **SQL DDL 文件需要人工维护合并**（多人协作改 schema 时），不如 Go struct 走 Git 合并那么顺（虽然 SQL 也能合并，但冲突解决要懂 DDL 语义）。
- **AI 补全 SQL 没有补全 Go struct 自然**（当前 2026 的 LLM 在 Go 代码补全上训练更充分），schema-first 在「AI 驱动开发」上体验更好。
- **中间路线仍是三件套**（SQL 文件 + Atlas + gorm/gen），不如 ent 的「一个工具链」概念统一——但如果本来就熟 SQL + GORM，接受度会高。

## 八、落地要点（下一步参考，本轮不实施）

**若选数据库优先（SQL schema + Atlas + gorm/gen）**：

- **唯一真相源**：一份 `schema/schema.sql`（或 Atlas HCL）定义完整期望 schema，包含所有表、索引、约束、扩展。
- **Atlas 配置**：`atlas.hcl` 的 `data "external_schema"` 指向这份 SQL 文件（或 HCL），`atlas migrate diff` 对比 dev DB 生成迁移。
- **dev 环境**：`make reset-db` 直接跑 `schema.sql` 建库（快速重置），或走迁移。
- **生产/CI**：走 Atlas 生成的版本化迁移（`atlas migrate apply`）。
- **codegen**：迁移 apply 后跑 `make gen-db`（gorm/gen 从更新后的 DB 生成 model）。
- **Makefile 流程**：`dev` target = `reset-db` → `gen-db` → `run`；`migrate` = `atlas migrate diff` → review → `apply` → `gen-db`。

**若选 schema 优先（ent 或 GORM + Atlas struct）**：参考 02 文档。

## 九、参考链接

- [Atlas — SQL Schema as the desired state](https://atlasgo.io/atlas-schema/sql)（SQL 文件当真相源）
- [Atlas — HCL Schema](https://atlasgo.io/atlas-schema/hcl)（HCL 定义 schema）
- [Atlas — Declarative vs Versioned Migrations](https://atlasgo.io/concepts/declarative-vs-versioned)（两种工作流）
- [Atlas — Introduction to Versioned Migrations](https://www.atlasgo.io/versioned/intro)
- [gorm.io/gen — GitHub](https://github.com/go-gorm/gen)（gorm/gen 代码生成）
- [GORM Gen Guides — Generate Model](https://gorm.io/gen/gen_model.html)（从 DB 生成 model）
- 交叉引用：
  - [02-数据库访问层选型调研.md](./02-数据库访问层选型调研.md)（schema 优先方向、7 条约束）
  - [03-迁移工具与数据初始化方案.md](./03-迁移工具与数据初始化方案.md)（Atlas 与 ent 同源、seed 三层分类）
