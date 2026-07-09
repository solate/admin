# 从零设计 Go 数据库层：GORM+Gen+golang-migrate 实战与规范落地（2026）

> 本文是一篇完整实战文档。读完你能独立完成：从 0 搭建数据库访问层 → 想清楚"用什么工具、为什么不用 X" → 制定防 drift 的迁移规范 → 编写幂等 data migration → 配置代码生成 → 封装连接与 SQL 日志 → 定义事务约定 → 掌握零停机演进策略。
>
> 技术栈：Go 1.26 + [GORM](https://gorm.io) v1.31 + [gorm/gen](https://github.com/go-gorm/gen) v0.3 + [driver/postgres](https://github.com/go-gorm/postgres) v1.6 + [golang-migrate](https://github.com/golang-migrate/migrate)（CLI）+ PostgreSQL。基于真实 SaaS 后端项目 `backend/` 的生产实现。

---

## 0. 开篇：数据库层到底难在哪

数据库访问层看起来是"连上库、跑 SQL、拿结果"的基础设施，真做起来会踩到一串问题：

1. **多环境 schema 同步**——本地、测试、生产三套库，怎么保证结构一致？改了表结构，三个环境都要手动跑一遍 ALTER 吗？
2. **初始数据不散落**——菜单、字典、权限配置，是写在 migration 的 INSERT 里，还是独立的 seed 脚本，还是代码里硬编码？三个位置都有一点，改起来漏改是常态。
3. **历史数据回填**——上线半年后加了新字段，几百万行老数据的默认值怎么填？一条 `UPDATE` 全表扫描锁库几分钟？
4. **ID 生成策略**——自增 ID、UUID、Sonyflake、数据库 sequence，各有什么坑？跨环境数据迁移时 ID 冲突怎么办？
5. **跨 repo 事务**——一个业务操作要改三张表（三个 repository），事务怎么写？每个 repo 都传 `tx` 参数，还是有更干净的约定？

本文围绕这五个问题，从选型推理、反面镜鉴到真实落地代码，把数据库层的设计思路和实践约定系统梳理一遍。所有代码基于真实的 SaaS 后端项目（`backend/`），migrations、Makefile、gen-db 脚本、连接封装全部可跑。

**为什么选 GORM+gen+golang-migrate**：2026 年 Go 数据库访问层的主流方案是 GORM（ORM 生态最广）+ gorm/gen（类型安全代码生成）+ golang-migrate（迁移执行器）。虽然 sqlc、ent、Bob + Atlas 各有优势，但在「动态查询友好 + 团队熟悉度 + 与老项目一致」这三个约束下，GORM 方案是最小阻力路径。本文也会诚实讲出它的 tradeoff（database-first 需要手动 gen-db、golang-migrate 的 dirty state 风险、重建派事务的啰嗦）。

---

## 1. 选型推理：为什么是 GORM+gen+golang-migrate

### 1.1 七条约束（驱动决策）

在选型窗口期（2026-07），我们列出了七条硬约束——每条都来自实际痛点或团队现状，不是凭空想象：

| 约束 | 来源 | 对工具的影响 |
|------|------|-------------|
| ① 比 ent 轻 | ent 的 schema-as-code + 生成的代码量与概念负担被否掉 | 排除 ent |
| ② 不写裸 SQL | sqlc 那种「手写 SQL 文件」的心智负担太高，尤其多条件 WHERE | 排除 sqlc（部分） |
| ③ 动态多条件 WHERE 要好写 | 列表查询的可选过滤（name/status/时间范围）是高频场景 | **核心痛点**，淘汰 sqlc |
| ④ AI 生成友好 | 类型可发现、API 规律性强，AI 补全/生成不易错 | GORM 语料最多 |
| ⑤ PostgreSQL 优先 | 不需要跨库抽象，要能用 PG 原生特性（jsonb/array/upsert/RETURNING/CTE） | 不强求 ORM 抽象层完美 |
| ⑥ 迁移自动化 | 不想手写手维护 migration SQL，改 schema 后迁移应能自动 diff 生成 | 初版约束，后放宽 |
| ⑦ 生成的 model 是确定类型 | model 结构体字段类型确定、编译期可查，不接受 `map[string]interface{}` | 所有代码生成方案都满足 |

这七条是漏斗：每条约束淘汰一批候选，最后留下的就是答案。

### 1.2 快速否定 sqlc（约束 ③ 的硬伤）

**sqlc 的动态查询硬伤未解**。这不是错觉、是设计限制。sqlc 从**静态** SQL 生成代码，**天生不支持运行时拼条件**。官方 [Discussion #364](https://github.com/sqlc-dev/sqlc/discussions/364) 讨论多年，社区共识仍是「动态查询请另配 squirrel」——即 **sqlc + squirrel 两套并存**，复杂度反而上升，且 AI 要在两种范式间跳。

同样一个「可选 name / status / 时间范围过滤 + 分页」的列表查询，GORM gen 的写法：

```go
query := r.q.User.WithContext(ctx).Where(r.q.User.TenantID.Eq(tenantID))
if name != "" {
    query = query.Where(r.q.User.Name.Eq(name))
}
if status != "" {
    query = query.Where(r.q.User.Status.Eq(status))
}
if !start.IsZero() {
    query = query.Where(r.q.User.CreatedAt.Gte(start))
}
users, count, err := query.FindByPage(offset, limit)
```

sqlc 做不到上面任何一种；要么给每种过滤组合写一条静态查询（组合爆炸），要么 `COALESCE(sqlc.narg('name'), name)` 这类丑陋 hack，要么外挂 squirrel。**这就是它写着难受的根因。**

> sqlc 用户并不难受——只是痛点排序和我们相反。sqlc 用「写更多 SQL」换「编译期安全 + 零开销 + 透明」，适合 Cloudflare 那种「延迟敏感、SQL 即设计、宁可多写」的画像；本项目诉求是「少操心框架、专注业务、动态查询顺手」，正好相反。不是 sqlc 不好，是它的收益我们不敏感、它的短板我们天天踩。

### 1.2.1 更进一步：连 ORM 都不要，AI 直接写原生 SQL 行不行

sqlc 是「少框架」方向的一个极端，还有更极端的一步——既然本项目代码全程由 AI 写、没人手写 SQL，那干脆去掉 GORM，直接用 Go 原生 `database/sql` + 手写 model/repo，技术栈少一个组件、不是更「简单」？

这个推理藏了个错误前提：**它假设选型是为了「让写代码的人省力」**。AI 全程写代码时，选型标准变了——不是「哪个初写更快」（AI 写哪个都快），而是**「哪个把 AI 的概率性手误挡在编译期、哪个改列时不用 AI 手动同步散落各处的列顺序」**。

**编译期 vs 运行期**：AI 会拼错字段名、用错类型，频率不比人低（它是概率生成，甚至更高）。区别只在错误何时暴露——

```go
// GORM + gen：拼错字段名 → 编译不过，当场红
r.q.User.Staus.Eq(1)      // ❌ compile error: q.User has no field Staus
r.q.User.Status.Eq("1")   // ❌ compile error: Status is int, not string

// 原生 SQL：拼错藏在字符串里 → 编译通过，跑到那行才炸
db.QueryRowContext(ctx, "SELECT staus FROM users WHERE id=$1", id)  // ✅ 编译通过
                                                                    // 💥 运行时 pq: column "staus" does not exist
```

更隐蔽的是 `Scan()` 按**位置**匹配：列顺序和 struct 字段顺序错位，可能**不报错**，只是把 `email` 塞进了 `phone` 字段（静默数据错乱），比直接崩还难查。

**改一列的爆炸半径**：这是原生 SQL 最大的长期税。给 `users` 加个 `phone` 列——

| 方案 | 要改几处 |
|---|---|
| GORM + gen | **2 处**：写 migration + `make gen-db`（model/query 自动重新生成） |
| 原生 SQL | **N+2 处**：migration + 手改 model + 逐一同步 N 条查询的 `SELECT` 列表和 `Scan()` 参数——漏一个就运行时错位 |

`make gen-db` 从活库反射一条命令收敛，AI 不需要记住「哪些地方引用了这张表」；原生 SQL 则要 AI 人肉找全所有引用点。本项目 12 个域、几十上百个 repo 方法，这个差距是天天付的。

> 核心洞察：**AI 写代码不是「随便选哪个都行」，恰恰相反——因为 AI 会犯概率性手误，更需要一层编译期类型系统当安全网。GORM+gen 的类型安全在 AI 时代不是「锦上添花」，是「刚需」。**

**但也别把 ORM 当教条**——正解是 **90/10 混合**：90% 的标准 CRUD 走 gen 的类型安全 API，10% 真正复杂的查询（递归 CTE 部门树、窗口函数报表）用 GORM 的 `Raw().Scan()` 直接落裸 SQL。要 SQL 表达力时随时能落到裸 SQL，不必为这 10% 把 90% 的类型安全全丢掉。

**什么时候原生 SQL 才更优**：项目 < 5 表、纯 CRUD、无动态查询；或 schema 已 100% 冻结；或**实测**（不是「觉得反射慢」，先 profile）GORM 是性能瓶颈。本项目三条都不沾，全部指向保留 GORM。完整五场景对比见 [08-GORM与原生SQL对比](../saas-backend/research/database/08-GORM与原生SQL对比.md)。

### 1.3 不选 ent（约束 ① 的概念负担）

ent 是 2026 年「最一站式」的方案——schema-as-code + 内建 Atlas 迁移引擎（同源团队 Ariga）+ 类型安全查询 + Edges/Hooks 全家桶。但正是这个「全」带来了约束 ① 的问题：

- **schema DSL 是新学习曲线**：不是写 Go struct，是写 ent 的 schema 描述（`field.String("name").NotEmpty()`），团队需要统一掌握这套 DSL。
- **生成代码量大**：一张表生成十几个文件（create/update/query builders、edges、predicates），虽然不手改它们，但 IDE 索引慢、Git diff 大。
- **概念多**：Edges（关联）、Hooks（生命周期）、Mixin（复用）、Privacy（权限）——对本项目是「功能过剩」。

老项目 A 已用 GORM+gen 跑通全套多租户 RBAC，团队熟悉、AI 语料多。新项目选 ent 等于推翻重来，学习成本 > 收益。

### 1.3.1 补充实证：Ent 除了「重」是不是全面更好

§1.3 是内部推理。但 Ent 有两张牌前面没接住：`tx.Client()` 让事务里无需重建 repo、schema-diff 自动生成迁移——**除了「重」，它是不是其他方面全面更好？** 这里补两个新论点 + 三条外部实证，给出诚实回答。

**论点一：「重」不是可以单独摘掉的缺点，它就是三件事本身。** 想「只要 Ent 的 `tx.Client()` 和 auto-diff，把重去掉」，前提不成立——因为「重」具体就是：① schema 必须单包集中（违反本项目 `domain-architecture.md` 的 12 域子包架构）；② 真相源被迫翻成 schema-first（本项目是 database-first，见 §1.6.1）；③ 图 DSL 学习成本（Edge O2O/O2M/M2M、Mixin、Hook、Privacy）。这三条不是「重之外的缺点」——**它们就是重**。拿不掉。

**论点二：Ent 的迁移优势只在 DDL，而 data migration 才是本项目的大头。** Ent 的 auto-diff 强在「建表/加列」，但本项目迁移工作量一大半是 **data migration**（菜单/字典/权限的幂等 `INSERT ON CONFLICT`，见第 6 章）——这一侧 Ent **零自动化**，同样手写 SQL，甚至更绕。「Ent 迁移更强」的直觉在占比更大的 data 侧不成立。

**三条外部实证**（2026 web 调研，非本项目自说自话）：

1. **公开从 GORM 转 Ent 的团队，逃离的是「裸 GORM」**——`AutoMigrate` + `struct` + 无类型查询 + 无版本化迁移（Nitric 原话「dissatisfied with the lack of typing」）。而本项目起点根本不是裸 GORM：gorm/gen 解决了类型、手写 golang-migrate 解决了版本化迁移。**他们要翻的两座山，本项目已经在山那边了。**
2. **Ent 官方博客自己承认**：auto-migration 随项目变大不够用，2022 才补上版本化迁移（"as their project grows, they may find that they need more control over the migration process"）。**Ent 绕几年才到的终点，本项目一步到位。**
3. **2026 最全的几篇对比（Encore、Glukhov、dasroot）没有一篇给出「必选 X」结论**，全是「看场景」；且 Encore 2026 点名 Ent 的短板：**编译慢、单体化、微服务里难拆**。

**什么时候该反选 Ent**：全新项目 + 图密集域（社交关系、组织树深度遍历、好友的好友）+ 接受单包 schema + 团队愿学图 DSL——**四条全中**才划算。本项目一条都不占。完整实证与引用见 [09-Ent与GORM迁移工具链对比](../saas-backend/research/database/09-Ent与GORM迁移工具链对比.md)。

### 1.4 不选 Atlas（约束 ⑥ 放宽后非必需）

初版约束 ⑥「迁移自动化」时，Atlas 是唯一解——它能读 GORM struct 或 SQL schema，自动 diff 出版本化迁移，还能 `migrate lint` 检测破坏性变更。但随着调研深入，发现两个关键事实：

1. **老项目的乱不是 golang-migrate 的锅**（详见第 2 章），是「三源 drift + 缺约定」。golang-migrate 只是 SQL 执行器，配上明确约定（单一真相源 + Makefile 标准化 + 幂等 data migration），它一样干净。
2. **Atlas 是额外的二进制 + CI 环节 + 学习成本**。对教学/中等规模项目，手写 migration SQL（或从模板 copy）+ golang-migrate 执行，已经够简单可控。

**两者的定位差一层**：golang-migrate 是「迁移执行器」——只负责按版本号顺序跑你手写的 SQL、记录跑到哪一版，SQL 内容全靠手写。Atlas 是「schema 引擎」——理解 schema 的**状态**，你声明期望结构（或直接读 GORM struct），它对比当前 DB 自动 diff 出迁移。一句话：**golang-migrate 帮你「跑」迁移，Atlas 帮你「写」迁移**。

| 维度 | golang-migrate（本项目） | Atlas |
|---|---|---|
| 迁移 SQL 谁写 | 全手写 up + down | 自动 diff 生成 |
| drift 防护 | ❌ 无（与 model/库无校验） | ✅ `migrate lint` + schema diff |
| 失败后状态 | ⚠️ 留 **dirty version**，需手工修 | ✅ pre-migration 校验，失败处理更稳 |
| 学习成本 | 极低（一条命令跑 SQL） | 中（HCL/diff/两种工作流） |

**要重点认识 golang-migrate 的 dirty state 坑**（也是它最大的固有代价）：迁移执行**中途失败**时（比如一条 `ALTER` 撞上锁超时），golang-migrate 会把 `schema_migrations` 表标记为 **dirty**，且**不会自动回滚**——因为它不在事务里做预校验。此后所有 `migrate up/down` 都被拒绝，必须人工 `migrate force <version>` 手动确认版本、清 dirty 标记，才能继续。这在生产/CI 里是个经典的「深夜惊魂」。Atlas 正是靠 lint 预检 + 更完善的失败处理来改进这一点。

**我们怎么规避**：开发期 `make reset` 一键 drop + 重跑，dirty 了直接重建，不用手动 force；生产则靠「迁移前在预发库充分测试 + 小步提交（一个 migration 只做一件事，失败面小）+ DDL 头部加 `lock_timeout`（见第 10 章）」三条兜底。接受 dirty state 这个代价，换 golang-migrate 的极简。

Atlas 的价值在「频繁改 schema + 多人协作 + 需要自动 lint」的场景最大化。本项目 schema 变更低频、团队小、可以在 PR review 时人工检查 migration——**暂不引入 Atlas，保持工具链简单**。若未来 schema 演进频繁到手写吃力，再上 Atlas 也不迟（它能读现有 migrations/，平滑接入）。

> 承认 tradeoff：不用 Atlas 意味着放弃自动 diff、自动 lint、防 drift 的机制保护。换来的是少一个工具、少一层抽象、golang-migrate 的简单性。这是「简单 vs 自动化」的天平，我们选了前者——对当前规模合理，对未来可逆。

### 1.5 最终选择：GORM + gorm/gen + golang-migrate

| 维度 | GORM + gen + golang-migrate |
|------|----------------------------|
| 动态 WHERE | ✅ 链式条件，约束 ③ 满足 |
| 确定类型 model | ✅ gen 生成，约束 ⑦ 满足 |
| PG 原生特性 | ✅ 支持（经 ORM 抽象层），约束 ⑤ 部分满足 |
| AI 生成友好 | ✅ 语料最多、老项目已验证，约束 ④ 满足 |
| 轻量度 | 中（GORM 反射运行时，但比 ent 轻），约束 ① 满足 |
| 团队/仓库一致性 | ✅ 与 老项目 A 完全一致 |
| 迁移成本 | 低：复刻老项目体验，补约定即可 |
| 生态成熟度 | ✅ 最成熟稳定 |

这套组合的核心优势是**最小阻力路径**：老项目用了、团队熟、AI 熟、满足全部 7 条约束（约束 ⑥ 放宽后 golang-migrate 够用）。不是「最先进」或「最自动化」，而是「最适合当前」。

### 1.6 备选路线：Bob + Atlas（若愿为 PG 契合付学习成本）

如果本 step 想借机做一次面向 PostgreSQL 的技术升级、且能接受仓库里两种范式共存，2026 年最「新」的正解是 **Bob + Atlas**。

Bob（`stephenafamo/bob`）是 database-first 的代码生成方案，两个特点贴合前 5 条约束：一是 `Apply()` mods 让**动态查询成为一等公民**（和 gorm/gen 链式条件同级，不像 sqlc 要外挂 squirrel）；二是**按 PG spec 定制**，jsonb / array / upsert / RETURNING / CTE 贴原生，不经 ORM 抽象层损耗。配 Atlas 管 SQL/HCL schema 与自动迁移，闭环成立。

| 维度 | GORM+gen+golang-migrate（本项目） | Bob + Atlas |
|---|---|---|
| 真相源方向 | 数据库优先（SQL migration） | 数据库优先（SQL/HCL schema） |
| 动态 WHERE | ✅ 链式条件 | ✅ `Apply()` mods |
| PG 原生特性 | ✅ 经 ORM 抽象层 | ✅✅ 按 PG spec 定制，贴原生 |
| 迁移自动化 | ❌ 手写 SQL | ✅ Atlas 自动 diff |
| 与 老项目 A 一致 | ✅ 完全一致 | ❌ 引入第二种范式 |
| 生态成熟度 | ✅ 最成熟 | 活跃但年轻（2024+ 崛起） |

**为什么本项目仍不选它**：Bob 引入与 老项目 A 不同的第二种范式（仓库风格分裂）、生态年轻语料少、database-first 心智 + Bob API + Atlas 三件套都要学。**求稳、团队一致的诉求压过「更贴 PG」的收益**。但这条路线本身没错——若你的项目 PG 高级特性用得重（分区、GIN、生成列），它比 GORM 方案更契合。

### 1.6.1 为什么不选 schema-first（Go struct 真相源）

选型还有一条正交的轴：**真相源方向**。

- **schema-first**（代码→DB）：Go struct 是权威、DB 追随。改 struct → 生成迁移 → apply，全程在 Go 里。代表：ent、GORM + Atlas（struct 模式）。
- **database-first**（DB→代码）：SQL 是权威、Go 追随。改 SQL → apply → 重跑 gen-db。代表：本项目、Bob、sqlc。

本项目选 **database-first**，除了「与老项目 `gen-db` 一致」，还有一条实打实的技术理由——**PG 满血**。GORM struct tag 表达 PG 高级 DDL 有天花板：

| PG 特性 | struct tag 能表达吗 |
|---|---|
| 部分索引（`WHERE` 条件） | ⚠️ 简单的能，复杂表达式写不进 |
| GIN/GiST 索引（`jsonb_path_ops` 等选项） | ⚠️ 勉强，选项没法表达 |
| 生成列（`GENERATED ALWAYS AS`） | ❌ 无对应 tag |
| 表分区（range/list/hash） | ❌ 完全无法 |
| Extension / 自定义类型 | ❌ DDL 层面要手写 SQL |

真要「用满 PG」，schema-first 会反复撞天花板、最后靠手改生成的迁移 SQL 打补丁——反而破坏了「只碰 Go」的初衷。**database-first 直接写 SQL DDL，能写什么就能用什么**。

**代价**（诚实说）：每次改 schema 要多跑一步 `make gen-db`（Go 是下游、不能「改 struct 即触发一切」）。这一步我们用 Makefile 收敛（见第 5 章），接受它换 SQL 的完全掌控。

> 两条路线同等合法，取舍见 [04-schema优先与数据库优先](../saas-backend/research/database/04-schema优先与数据库优先.md)。一句话：schema-first =「我是 Go 开发者，DB 只是存储细节」；database-first =「我是设计 schema 的人，Go 只是实现层」。

---

## 2. 反面镜鉴：老项目的三大病根

选型时，我们探索了两个老项目（老项目 A、老项目 B），发现它们的混乱**不是 GORM 或 golang-migrate 的问题**，而是**缺少明确约定**。把这些病根列出来，作为反面教材——新项目要做的是「定约定堵漏洞」，不是「换工具」。

### 2.1 老项目 A 的 3/12/57 schema drift

| 源 | 位置 | 表数量 | 说明 |
|---|---|---|---|
| migrations | `migrations/*.up.sql` | **3 个 CREATE TABLE** | 仅 users / user_roles / role_permissions |
| dev schema | `scripts/dev_schema.sql` | **12 个表** | tenants / users / roles / menus / permissions / depts / positions / ... |
| generated models | `internal/dal/model/*.gen.go` | **57 个模型** | 包含 resource/asset/content/casbin_rule/api_resources 等，反映真实库 |

**三套路径互不一致**：`migrate up` 建 3 表、`dev-reset.sh` 加载 dev_schema 建 12 表、gen-db 生成 57 model 反映线上真实库。哪个是真相源？没人说得清。

**种子数据冲突**：Go seeder (`scripts/init_data/`) 不在 Makefile、手动触发、未文档化，且与 `migrations/000001` 的 `INSERT users` 硬编码（admin/auditor/uploader）冲突——两套 user seed，一个在迁移、一个在 seeder。

**根因**：golang-migrate 只是「SQL 执行器」，不理解 schema 的**状态**；gen-db 又是纯 database-first（读活库反射）。两者之间**没有任何一致性校验**，drift 是这套组合的必然结果——但不是工具的锅，是管理失控。

### 2.2 老项目 B 的五大痛点（62 个 migration）

1. **api_resource 三源维护**：Go seeder (`scripts/init_data/seeds/api_resource.go`, 629 行) + migration INSERT（`000046`/`000056`/`000062`）+ 独立 SQL 脚本（44KB `insert_api_resources_data.sql`）——同一份数据三处维护，必然 drift。
2. **UUID 手工计数**：seeder `main.go` 注释「6 + 29 menu + 19 dept + 37 position + 52 dict + 105 API = 248」，`idgen.GenerateUUIDs(248)` 后手动推进 `idIndex` 切片——数量一变静默错位，典型维护陷阱。
3. **版本补丁式迁移**：`000056_add_api_resources_for_v1_1_0.up.sql` 插 1 条，后来 `000062_add_api_resources_for_v1_1_0_full.up.sql` 再「full」重做插 5 条——同一份数据两次打补丁。
4. **Casbin 与 api_resource 双源同步**：migration 000007 注释「此表仅用于元数据管理和前端展示，不用于权限检查。权限检查由 Casbin 负责」——元数据一处（api_resource）、鉴权另一处（Casbin policies），手动保持同步。
5. **dev_schema.sql 快照与 migration 并存**：`scripts/dev-reset.sh` 加载 50KB `dev_schema.sql` 快照，本地重置不走 migration 链——两套建库路径，快照渐渐过期。

### 2.3 关键洞察

**这些痛全是「流程/约定」问题，不是「工具」问题。** 换 sqlc 一个都解决不了——因为它们发生在迁移和种子数据层，与查询层用什么无关。而且新项目已经架构性消掉一部分痛（Casbin 完全移除，见 `.claude/rules/rbac-multi-tenant.md`），所以 老项目 B 的「Casbin/api_resource 双源」在新项目根本不存在。

**正解**：不换工具，定一套明确约定，把 老项目 B 的每个坑逐条堵死。

---

## 3. 六条核心约定（从病根推导出的避坑原则）

每条约定对应老项目的一个具体坑。这六条构成了本项目数据库层的**约定大于配置**铁律。

### 约定 1：单一真相源（禁止 migration / dev_schema / snapshot 三套路径）

**坑**：老项目 A 有 3 表 migration、12 表 dev_schema、57 model；老项目 B 有 62 个 migration 但 dev-reset 走快照。

**约定**：**只有一个 canonical schema 定义**。本项目选**数据库优先**（database-first）：`migrations/*.up.sql` 唯一真相源。

**禁止**：`scripts/dev_schema.sql`、`scripts/dev-reset.sh` 加载快照、任何「不走 migration 的建库捷径」。

**落地**：`make reset` = `migrate-reset`（`drop -f` → `migrate up`）+ `gen-db`，从零重建。

### 约定 2：Schema 与 Data 分文件、同序列（05 文档落地）

**坑**：老项目 B 把 DDL 与 INSERT 混在同一个 migration（000044/046/056/062），改结构和改数据搅在一起，回滚和 review 都难。

**约定**（关键：分**文件**，不分**目录**）：
- **Schema migration**（`000001_ddl_init_schema.up.sql`）：只写 DDL（`CREATE TABLE` / `ALTER TABLE` / `CREATE INDEX`）。
- **Data migration**（`000002_data_base_config.up.sql`）：只写数据（`INSERT ... ON CONFLICT` / `UPDATE`），**独立成文件，但仍在同一个 `migrations/` 版本序列里**。
- **修复 migration**（`000005_fix_xxx.up.sql`）：数据订正或结构修复，按性质可能是 DDL 也可能是 DML，统一用 `fix_` 前缀。

**命名前缀规范**（人和 AI 都靠它识别）：
- `ddl_` — 结构变更（init / create / alter / drop / add_column 等）
- `data_` — 配置数据（base_config / content_menus / roles 等）
- `fix_` — 错误修复（不论 DDL 还是 data，统一前缀）

**示例**：
```
000001_ddl_init_schema.up.sql      ← DDL
000002_data_base_config.up.sql     ← 初始数据
000003_ddl_add_icon_col.up.sql     ← DDL 加列
000004_data_content_menus.up.sql     ← 数据（依赖 000003）
000005_fix_menu_name.up.sql        ← 修复
```

**为什么不拆出 `seeds/` 目录**：
1. 多一个目录 = AI/人多一个「这改动该放哪」的判断点
2. **DDL 和数据有交错依赖**（数据依赖新列、新列要先 backfill 数据再加约束）——拆目录无法表达这个顺序，单序列自然保证
3. 所有库变更统一走 `migrations/`，规则只有一条——**改库 = 新建编号 migration**，DDL 与 data 靠**文件名前缀**区分

**落地**：不需要 `seed` / `backfill` 独立 target，全部由 `migrate-up` 按序号一次跑完。

### 约定 3：Data migration 幂等 + 内联 ID（禁止手工数 UUID、count-then-insert）

**坑**：老项目 B 手工数 248 个 UUID（"6+29+19+37+52+105"），数错就错位；老项目 A 的 menu seeder 用 `Count() > 0` 跳过整个 seed（非幂等）。

**约定**：
- **`INSERT ... ON CONFLICT (unique_key) DO UPDATE`**：以业务自然键（如 `menu_code`）为冲突目标，重跑收敛到声明状态，幂等。
- **ID 内联硬编码**：本项目主键是 Sonyflake 数字字符串（`config_id varchar(20)` 等），由 Go 侧 `idgen.GenerateUUID()` 预生成后**直接写进 SQL 的 VALUES**（如 `'153547313510393194'`）。稳定、跨环境一致、外键引用不会因重跑漂移。
- **禁止 `gen_random_uuid()`**：库内随机生成会让每个环境 ID 不同，关联表无法硬编码外键。
- **禁止手工数量切片**：不写 `GenerateUUIDs(248)` 再手动推进 index——数量一变静默错位。

**禁止**：count-then-insert（`if count > 0 { return }`）——非幂等，重跑会漏。

**落地**：第 6 节给完整 data migration 模板。

### 约定 4：api_resource 从路由同步（方案 A，03 已定）

**坑**：老项目 B 的 api_resource 在 Go seeder、migration INSERT、独立 SQL 脚本三处维护。

**约定**：**路由注册是唯一真相源**（03 seed 三层分类的方案 A）。启动时遍历 Gin `Engine.Routes()`，自动 upsert 到 `api_resources` 表。

**禁止**：手写 api_resource INSERT（无论在 migration、seeder、还是独立 SQL）。

**先分清「初始数据」的三层**——老项目乱在没按「数据的变化节奏」分类，一股脑要么塞 migration、要么临时写脚本：

| 层 | 典型数据 | 变化节奏 | 真相源 | 推荐机制 |
|---|---|---|---|---|
| ① schema | 表结构、索引、约束 | 随迭代 | migration | DDL migration（约定 2） |
| ② 框架种子 | 菜单、字典、权限点 | 跟着代码走 | 代码/配置 | **data migration**（约定 3，幂等 INSERT ON CONFLICT） |
| ③ api_resource | API 清单 | 跟着路由走 | **Gin 路由** | **启动时路由同步**（本约定） |
| ④ 运行时数据 | role_permissions 绑定、租户业务数据 | 运行时人操作 | 数据库 | 管理界面，**不 seed** |

**为什么第 ② 层用 data migration，不用独立 seeder / YAML seed**：
- **纯 Go seeder 的致命缺陷**：不记录「跑过哪个版本」、多环境不一致（`git checkout v1.1 && make seed` 跑的是 v1.1 代码声明的菜单，但库里可能还留着 v1.2 的记录，upsert 语义不会删多余的）。老项目 A 更用了 `Count() > 0` 跳过整个 seed——非幂等，加了新菜单也不会补。
- **YAML seed + migration 触发**（曾考虑的方案 D）：能版本化，但多一层 `YAML → struct → ORM` 翻译，还要写 `apply_seed.go` 应用器——对本项目是过度设计。
- **结论**：菜单/字典直接写成 data migration（SQL 本身就声明式、可 diff、随版本序列走），一条规则「改库 = 新 migration」到底，见约定 2/3。

**为什么第 ③ 层特殊、单拎出来走路由同步**：api_resource 的真相源是**路由注册代码本身**，不是配置数据。程序里所有 API 本就要注册进 Gin 路由，那份 path/method 清单已经是最权威、最不会漏的 API 列表。启动时遍历一次 upsert，永远和真实路由一致——彻底消灭「加接口忘补权限」和 老项目 B 的三源维护。

**落地代码**（`internal/router/sync_api_resource.go`，本项目暂未实现，留作权限模块 Step 07/08 落地点）：

```go
// SyncAPIResources 遍历已注册路由，upsert 进 api_resources 表。
// 启动时调用一次 —— 新增接口自动登记，无需手写 INSERT（约定 4）。
func SyncAPIResources(ctx context.Context, db *gorm.DB, engine *gin.Engine) error {
	routes := engine.Routes()
	resources := make([]*model.APIResource, 0, len(routes))
	ids := idgen.GenerateUUIDs(len(routes))

	for i, r := range routes {
		resources = append(resources, &model.APIResource{
			APIResourceID: ids[i],
			Path:          r.Path,   // 如 /api/v1/users/:id
			Method:        r.Method, // GET / POST / ...
		})
	}
	if len(resources) == 0 {
		return nil
	}

	// path+method 唯一键，冲突时更新元数据（幂等）
	return db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "path"}, {Name: "method"}},
		DoUpdates: clause.AssignmentColumns([]string{"description", "module", "updated_at"}),
	}).Create(&resources).Error
}
```

启动时调用（`cmd/server/main.go`，`router.Setup` 后）：

```go
router.Setup(r, handlers, cfg, jwtMgr, rbacCache)
if err := router.SyncAPIResources(ctx, db, r); err != nil {
	return fmt.Errorf("sync api resources: %w", err)
}
```

> 完整的 seed 三层分类与方案 A/B/C/D 演化见 [03-迁移工具与数据初始化方案](../saas-backend/research/database/03-迁移工具与数据初始化方案.md)。

### 约定 5：Makefile 标准化（DB 操作不散落）

**坑**：老项目 A 的 seeder 不在 Makefile、手动触发、未文档化；老项目 B 的 `dev-reset.sh` 绕过 migration。

**约定**：**Makefile 暴露全部 DB 操作**，禁止隐藏在 shell 脚本里：
- `migrate-up` / `migrate-down` / `migrate-reset` / `migrate-create`
- `gen-db`（gorm/gen 反射活库生成 model/query）
- `reset`（`migrate-reset` + `gen-db`，一键重建本地）

**注意**：没有独立的 `seed` target——数据是 data migration，`migrate-up` 一次跑完 schema + data（约定 2）。

**禁止**：`scripts/dev-reset.sh` 等绕过 Makefile 的自定义脚本。

**落地**：第 5 节给可 copy 的 Makefile 模板（对齐 `backend/Makefile` 现有 target）。

### 约定 6：本地环境一致性（reset 从零重建）

**坑**：老项目 B 的 `dev-reset.sh` 加载 50KB 快照，渐渐与 migration 链 drift。

**约定**：**`make reset` 必须走 migration from scratch**，不走快照。这保证本地环境与生产的 migration 历史完全一致。

**落地**：删掉 `dev_schema.sql` / `dev-reset.sh`，只留 `make reset` = `migrate-reset`（drop + up）+ `gen-db`。

---

## 4. 目录结构规划

一套把上述 6 条约定物化的目录布局：

```
backend/                      # 项目实际结构
├── migrations/               # golang-migrate 唯一源（schema + data 都在此）
│   ├── 000001_ddl_init_schema.up.sql        # DDL - 建表
│   ├── 000001_ddl_init_schema.down.sql
│   ├── 000002_data_base_config.up.sql       # 数据 - 初始配置（幂等 INSERT ON CONFLICT）
│   ├── 000002_data_base_config.down.sql     # DELETE FROM system_config WHERE ...
│   └── ...                              # DDL/data/fix 靠前缀分类，序号保证交错依赖顺序（约定 2）
├── scripts/
│   └── gen-from-db/
│       └── main.go             # gorm/gen 配置（反射活库生成，输出到 internal/dal/）
├── internal/
│   ├── dal/
│   │   ├── model/                  # gorm/gen 生成的 model（勿手改）
│   │   └── query/                  # gorm/gen 生成的类型安全 query
│   ├── repository/                 # 手写 repo 层（调用 dal/query）
│   ├── service/{domain}/           # 业务逻辑 + converter
│   ├── handler/{domain}/           # HTTP handler
│   └── router/
│       ├── router.go               # 路由注册
│       └── sync_api_resource.go    # 启动时从路由同步 api_resource（约定 4）
├── pkg/
│   └── database/
│       ├── config.go               # 连接参数结构体
│       ├── database.go             # New(*Config) 连接封装
│       └── gorm_logger.go          # GORM SQL 日志桥接到 slog
└── Makefile                        # 全部 DB 操作入口（约定 5）
```

**只有一个 `migrations/` 目录**——schema 和 data 都在里面，靠文件名前缀（`ddl_` / `data_` / `fix_`）区分，不拆多目录。序号是全局递增的单一序列，天然表达「先建表、再插数据、又改表、再补数据」这种 DDL↔data 交错依赖——这正是拆两个目录会丢掉的能力。

**已删除**（相比老项目）：
- ❌ `scripts/dev_schema.sql`（快照，违反约定 1/6）
- ❌ `scripts/dev-reset.sh`（绕过 migration，违反约定 6）
- ❌ `scripts/insert_api_resources_data.sql`（api_resource 手写 SQL，违反约定 4）
- ❌ `seeds/*.yaml` + `scripts/apply_seed.go`（YAML 翻译层，多余——数据直接写 data migration SQL）

---

## 5. Makefile 标准 target 模板

DB 操作全部收敛到 Makefile（约定 5）。以下是本项目 `backend/Makefile` 的真实片段：

```makefile
# ── 数据库连接(可被环境变量覆盖,默认对齐 config/config.yaml) ──
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= postgres
DB_PASSWORD ?= postgres
DB_NAME ?= app_dev
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
MIGRATION_DIR := ./migrations

# dev: 迁移 → 生成 DAL → 启动(对齐 migration-data-convention 约定的一键流程)
dev: migrate-up gen-db
	go run ./cmd/server

# ── 数据库迁移(golang-migrate CLI) ──

## migrate-up: 执行全部未应用的迁移
migrate-up:
	@if command -v migrate > /dev/null; then \
		echo "⬆️  执行数据库迁移..."; \
		migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" up; \
		echo "✅ 迁移完成"; \
	else \
		echo "❌ golang-migrate 未安装: brew install golang-migrate"; \
		exit 1; \
	fi

## migrate-down: 回滚一步
migrate-down:
	@migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" down 1

## migrate-create: 新建迁移文件对 (make migrate-create NAME=ddl_add_xxx)
migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "❌ 请指定名称: make migrate-create NAME=ddl_add_user_table"; \
		exit 1; \
	fi
	@migrate create -ext sql -dir $(MIGRATION_DIR) -seq $(NAME)

## migrate-reset: 全量回滚后重跑(危险,仅本地)
migrate-reset:
	@echo "⚠️  即将完全重置数据库 $(DB_NAME)! 按回车继续,Ctrl+C 取消..."
	@read confirm
	@migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" down -all || true
	@migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" up
	@echo "✅ 数据库重置完成"

# ── GORM Gen 代码生成(从活库反射) ──

## gen-db: 从数据库生成 model + query 到 internal/dal
gen-db:
	go run ./scripts/gen-from-db/

## reset: 一键重建本地(重置迁移 + 重新生成 DAL)
reset: migrate-reset gen-db
```

**关键点**：
- `DB_URL` 用 `?=` 允许环境变量覆盖，默认对齐 `config/config.yaml`。CI 里 `DB_HOST=ci-db make migrate-up` 即可换库。
- `migrate-create NAME=ddl_add_xxx` 里 NAME 就带前缀（约定 2），生成 `000003_ddl_add_xxx.up.sql` + `.down.sql` 一对空文件。
- `reset` 是本地开发的核心——任何时候库脏了，一条命令从零重建，且**走的就是生产同款 migration 链**（约定 6）。`migrate-reset` 内部 `down -all` + `up`，data migration 随 schema 一并跑完。
- **没有 `seed` target**——数据是 data migration，`migrate-up` 一次跑完 schema + data（约定 2）。
- **没有任何 `dev_schema.sql` / `dev-reset.sh`**（约定 1/6）。

---

## 6. migration 编写规范（真实代码）

### 6.1 DDL migration：只写结构

`migrations/000001_ddl_init_schema.up.sql`（本项目真实文件）：

```sql
-- 000001 结构 - 初始化 schema
-- 建一张基础设施级配置表，用于验证 DAL 管线（时间戳毫秒 bigint + soft_delete）。
-- 业务域表（tenants/users/roles...）由后续 step-04 定义，不在此处。

CREATE TABLE IF NOT EXISTS system_config (
    config_id    VARCHAR(20)  PRIMARY KEY,          -- Sonyflake 数字字符串主键
    config_key   VARCHAR(128) NOT NULL,             -- 配置键
    config_value TEXT         NOT NULL DEFAULT '',   -- 配置值
    remark       VARCHAR(255) NOT NULL DEFAULT '',   -- 备注
    created_at   BIGINT       NOT NULL DEFAULT 0,    -- 创建时间(毫秒)
    updated_at   BIGINT       NOT NULL DEFAULT 0,    -- 更新时间(毫秒)
    deleted_at   BIGINT       NOT NULL DEFAULT 0,    -- 软删除时间(毫秒,0=未删)
    UNIQUE (config_key, deleted_at)                  -- 键唯一(软删维度)
);

CREATE INDEX IF NOT EXISTS idx_system_config_key ON system_config (config_key);
```

对应的 `000001_ddl_init_schema.down.sql`：

```sql
DROP TABLE IF EXISTS system_config;
```

**几个项目级约定**（写进了每张表）：

1. **主键 `VARCHAR(20)` + Sonyflake 数字字符串**：不是自增、不是 UUID。Sonyflake 生成 19 位数字字符串（如 `153547313510393194`），全局唯一、趋势递增、跨环境一致。存 `VARCHAR(20)` 而非 `BIGINT`，是为了前端 JS `Number` 精度安全（超过 2^53 的整数 JS 会丢精度）。
2. **时间戳 `BIGINT` 毫秒**：`created_at` / `updated_at` / `deleted_at` 统一 `BIGINT`，存毫秒级时间戳（如 `1720425600000`）。不用 `TIMESTAMPTZ`，是为了跨语言/跨时区一致（前端直接拿毫秒数 new Date）。
3. **软删 `deleted_at BIGINT DEFAULT 0`**：0 = 未删，非 0 = 删除时刻的毫秒时间戳。配合 GORM 的 `soft_delete` 插件（见第 7 节）。
4. **唯一约束带软删维度**：`UNIQUE (config_key, deleted_at)` 而非 `UNIQUE (config_key)`——软删后 `deleted_at` 变成时间戳，同一个 key 可以重新插入，不会撞唯一约束。

### 6.2 Data migration：幂等 + 内联 ID

`migrations/000002_data_base_config.up.sql`（本项目真实文件）：

```sql
-- 000002 数据 - 基础配置初始化（幂等）
-- 演示 data migration 规范：INSERT ... ON CONFLICT DO UPDATE，ID 硬编码，时间戳毫秒。
-- 后续新增配置项：新建 000NNN_data_xxx.up.sql 只加新行，不改此文件。

INSERT INTO system_config (config_id, config_key, config_value, remark, created_at, updated_at, deleted_at) VALUES
  ('153547313510393201', 'site_name', 'Admin', '站点名称',
   EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, 0),
  ('153547313510393202', 'site_status', 'active', '站点状态',
   EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, 0)
ON CONFLICT (config_key, deleted_at) DO UPDATE SET
  config_value = EXCLUDED.config_value,
  remark = EXCLUDED.remark,
  updated_at = EXCLUDED.updated_at;
```

对应的 `000002_data_base_config.down.sql`（回滚：显式列出删除范围）：

```sql
DELETE FROM system_config WHERE config_key IN ('site_name', 'site_status');
```

**关键点**：

- **幂等**：`ON CONFLICT (config_key, deleted_at) DO UPDATE`——以唯一约束为冲突目标，`make reset` 或重跑都收敛到声明状态，不会 duplicate key error。冲突目标必须和表的 `UNIQUE` 约束列一致。
- **ID 内联硬编码**：`'153547313510393201'` 直接写进 VALUES。用 Go 的 `idgen.GenerateUUID()` 预生成后粘贴——**不用 `gen_random_uuid()`**（每环境不同、外键没法硬编码），**不用 `GenerateUUIDs(248)` 批量数**（数量变了静默错位）。
- **时间戳**：`EXTRACT(EPOCH FROM NOW())::BIGINT * 1000`（毫秒 bigint），对齐第 6.1 节的时间戳约定。
- **down 必写**：`DELETE ... WHERE config_key IN (...)` 显式列出本 migration 加的行，`migrate down` / `make reset` 时干净回退。

### 6.3 后续增量：只 append，不改老文件

加新配置项时，是一个**新** migration，不改 `000002`：

```sql
-- 000004_data_more_config.up.sql
INSERT INTO system_config (config_id, config_key, config_value, remark, created_at, updated_at, deleted_at) VALUES
  ('160123456789012345', 'upload_max_size', '10485760', '上传大小上限(字节)',
   EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, 0)
ON CONFLICT (config_key, deleted_at) DO UPDATE SET
  config_value = EXCLUDED.config_value, remark = EXCLUDED.remark, updated_at = EXCLUDED.updated_at;
```

**配置项从 2 个涨到 200 个 = 往序列后面 append 几十个 `data_xxx` migration**，每个各自带 down、删各自加的那几条。老 migration 一旦提交不再碰（不可变黄金规则）。

> **改错了怎么办**：改错了值/备注 → 新建 `000005_fix_config_value.up.sql`（`UPDATE ... WHERE config_key=...`，down 里改回旧值）；加错了不该加的 → `000006_fix_remove_wrong.up.sql`（`DELETE`，down 里 `INSERT` 回来）。**修复也是新 migration，不改老文件**——生产实践以 fix-forward 为主，down 主要服务本地 `make reset` 全量重建。

---

## 7. GORM Gen 代码生成（database-first）

### 7.1 工作流

本项目走 **database-first**：migrations 是真相源，改表流程是——

```
写 migration → make migrate-up（apply 到库）→ make gen-db（反射活库生成 model/query）
```

`scripts/gen-from-db/main.go`（本项目真实文件）：

```go
// gen-from-db 是 GORM Gen 的代码生成入口(database-first)。
//
// 用法:make gen-db(等价 go run ./scripts/gen-from-db/)。
// 流程:读 internal/config → 连活库 → 反射所有表 → 生成 internal/dal/model + internal/dal/query。
package main

import (
	"fmt"
	"log"
	"strings"

	"admin/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// codegen 专用连接:与运行时连接分离,开启预编译、跳过默认事务(仅反射建表用不到事务)。
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	genCfg := gen.Config{
		OutPath:           "./internal/dal/query",
		OutFile:           "gen.go",
		ModelPkgPath:      "./internal/dal/model",
		Mode:              gen.WithDefaultQuery | gen.WithQueryInterface,
		FieldNullable:     true,
		FieldCoverable:    false,
		FieldSignable:     false,
		FieldWithIndexTag: true,
		FieldWithTypeTag:  true,
	}
	genCfg.WithImportPkgPath("gorm.io/plugin/soft_delete")

	g := gen.NewGenerator(genCfg)
	g.UseDB(db)

	tables, err := db.Migrator().GetTables()
	if err != nil {
		log.Fatalf("读取数据库表失败: %v", err)
	}

	// schema_migrations 是 golang-migrate 的版本记账表,不生成 model。
	excludeTables := map[string]bool{
		"schema_migrations": true,
	}

	var models []any
	for _, table := range tables {
		if excludeTables[strings.ToLower(table)] {
			continue
		}
		opts := []gen.ModelOpt{
			gen.FieldGORMTag("created_at", func(tag field.GormTag) field.GormTag {
				tag.Set("autoCreateTime", "milli")
				return tag
			}),
			gen.FieldGORMTag("updated_at", func(tag field.GormTag) field.GormTag {
				tag.Set("autoUpdateTime", "milli")
				return tag
			}),
			gen.FieldType("deleted_at", "soft_delete.DeletedAt"),
			gen.FieldGORMTag("deleted_at", func(tag field.GormTag) field.GormTag {
				tag.Set("softDelete", "milli")
				return tag
			}),
		}
		models = append(models, g.GenerateModel(table, opts...))
	}

	g.ApplyBasic(models...)
	g.Execute()

	fmt.Printf("✅ 代码生成完成:%d 张表 → internal/dal/model + internal/dal/query\n", len(models))
}
```

### 7.2 三个关键配置

1. **排除 `schema_migrations`**：这是 golang-migrate 的版本记账表，不是业务表，不该生成 model。用 `excludeTables` map 跳过。
2. **时间戳字段自动打 tag**：`created_at` → `autoCreateTime:milli`、`updated_at` → `autoUpdateTime:milli`。这样 GORM 在 Create/Update 时自动填毫秒时间戳，业务代码不用手动 set。对齐第 6.1 节的时间戳约定。
3. **软删字段映射为 `soft_delete.DeletedAt`**：`deleted_at` 字段类型改成 `soft_delete.DeletedAt`（milli 模式）。GORM 的 soft_delete 插件会自动在查询时加 `WHERE deleted_at = 0`、删除时 `UPDATE deleted_at = <毫秒时间戳>`。

### 7.3 生成物示例

`internal/dal/model/system_config.gen.go`（gen 生成，**勿手改**）：

```go
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"gorm.io/plugin/soft_delete"
)

const TableNameSystemConfig = "system_config"

// SystemConfig mapped from table <system_config>
type SystemConfig struct {
	ConfigID    string                `gorm:"column:config_id;type:character varying(20);primaryKey" json:"config_id"`
	ConfigKey   string                `gorm:"column:config_key;type:character varying(128);not null;index:idx_system_config_key,priority:1" json:"config_key"`
	ConfigValue string                `gorm:"column:config_value;type:text;not null" json:"config_value"`
	Remark      string                `gorm:"column:remark;type:character varying(255);not null" json:"remark"`
	CreatedAt   int64                 `gorm:"column:created_at;type:bigint;not null;autoCreateTime:milli" json:"created_at"`
	UpdatedAt   int64                 `gorm:"column:updated_at;type:bigint;not null;autoUpdateTime:milli" json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;softDelete:milli" json:"deleted_at"`
}

func (*SystemConfig) TableName() string {
	return TableNameSystemConfig
}
```

**何时重跑 gen-db**：
- 每次 `make migrate-up` 改了表结构后
- 首次 clone 项目、`make reset` 之后
- **不要**手改 `internal/dal/model/*.gen.go`（会被下次 gen-db 覆盖）

### 7.4 要不要换 GORM 官方泛型 CLI

用了 gen，很自然会问一句：GORM 官方在 2024 年起推了一个新的**泛型 CLI**（`gorm.io/cli/gorm`），
本项目的 `gorm/gen v0.3.28` 是不是该升级过去？

先厘清定位：**新 CLI 不是 gen 的升级版，是官方并行的另一条路线**——两者可共存、可混用，
官方甚至专门有一页 `cli_vs_gen` 讲区别。CLI 确实带来几条 AI 友好的真实优势：无状态的
`gorm.G[T](db)` 入口（AI 不用理解 repo 实例的生命周期与重建）、更强的编译期类型
（官方原话 focus on type safety，if it compiles, it works，正好接上 [08 号文](../saas-backend/research/database/08-GORM与原生SQL对比.md)
「类型系统是 AI 安全网」的论点）、以及一等公民的强类型关联操作。

**但有一条反直觉的杀手事实**：

> **AI 训练语料里 gen 的写法远多于 CLI，「新工具」≠「AI 写得更对」。语料丰富度是
> AI 生成质量的第一性原理，这一条排在类型优势之前。**

官方那句「focus on type safety... for the AI coding」要拆成两个轴看：**类型系统轴上成立**
（编译期能拦更多错），**语料丰富度轴上（当前）不成立**——2026 的 AI 写 CLI 的正确率
短期内不一定比写 gen 高，因为它见过的 CLI 代码太少，容易生成混合两种 API 的错误代码、
或猜错 `gorm.G[T]` 的泛型约束。

| 维度 | 泛型 CLI | gorm/gen（本项目） |
|------|---------|------------------|
| 编译期类型安全 | ✅ 更强（if it compiles, it works） | ✅ 够用 |
| 状态管理心智 | ✅ 无状态 `gorm.G[T](db)` | 有状态 `r.q.User`（需理解重建） |
| **AI 语料丰富度** | ❌ 年轻、语料稀缺（**关键**） | ✅ 语料最多 |
| DSL 简单度 | ⚠️ 多一套 SQL 模板 DSL | ✅ 链式条件直白 |
| 类型安全泄漏 | ⚠️ `--typed=false` 给 AI 退回裸字符串的口子 | ✅ 无此模式 |
| 真相源方向 | model-first（读 Go struct） | ✅ DB-first（与本项目一致） |

**结论**：保持 `gorm/gen v0.3.28`——gen 未废弃、DB-first 已定、语料占优、迁移成本高、
与老项目一致。CLI 的类型优势是真的，但骗不过「AI 语料现实」：此刻让 AI 写它，不一定更对。
flip 条件——等 CLI 生态与语料再成熟一两年（2028+？）的**全新项目**再考虑。完整评估见
[10-gorm-gen与泛型CLI对比](../saas-backend/research/database/10-gorm-gen与泛型CLI对比.md)。

---

## 8. 连接封装与 SQL 日志（pkg/database）

### 8.1 Config：纯连接参数

`pkg/database/config.go`（本项目真实文件）：

```go
package database

// Config 数据库连接参数（纯连接配置，不含日志字段）。
// 由 internal/config.DatabaseConfig 逐字段映射而来（见 cmd/server/main.go）。
// SQL 日志的可见性由 slog 的 cfg.Log.Level 统一控制，不在此单独配置。
type Config struct {
	Host            string // 数据库主机地址
	Port            int    // 端口
	User            string // 用户名
	Password        string // 密码
	DBName          string // 库名
	SSLMode         string // SSL 模式：disable / require / verify-full 等
	MaxIdleConns    int    // 连接池最大空闲连接数
	MaxOpenConns    int    // 连接池最大打开连接数
	ConnMaxLifetime int    // 连接最大存活时长（秒）
}
```

**设计要点**：`pkg/database` 是可复用封装包，只吃纯连接参数，不 import 项目的 `internal/config`。`internal/config.DatabaseConfig` 在 `cmd/server/main.go` 里逐字段映射进来——这样 `pkg/database` 能整目录 copy 到其他项目。SQL 日志级别不在这里配，交给 slog 统一管（见 8.3）。

### 8.2 New：连接 + 连接池

`pkg/database/database.go`：

```go
package database

import (
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func New(cfg Config, log *slog.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newGormLogger(log),
	})
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	}

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
```

**要点**：
- 连接池三参数（`MaxIdleConns` / `MaxOpenConns` / `ConnMaxLifetime`）从 config 注入，不写死。默认值在 `config/config.yaml`：`max_idle_conns: 10`、`max_open_conns: 100`、`conn_max_lifetime: 3600`（秒）。
- `ConnMaxLifetime` 用 `int` 秒 + 代码里 `* time.Second`（对齐项目「时长字段用 int 秒」的约定，见配置加载 blog）。
- `Ping()` 兜底——连不上直接 fail-fast，不让服务带病启动。

### 8.3 SQL 日志：桥接到 slog

`pkg/database/gorm_logger.go`——把 GORM 的 SQL 日志桥接到项目统一的 slog：

```go
// gormSlogger 将 GORM 的 SQL 日志桥接到项目的 slog。
// 行为对齐 GORM 原生 logger：正常 SQL→Info、慢查询→Warn、出错→Error；
// 忽略 ErrRecordNotFound（正常业务未命中，不当错误刷屏）。
type gormSlogger struct {
	log           *slog.Logger
	slowThreshold time.Duration
}

// Trace 每条 SQL 执行后由 GORM 回调，按原生三分支语义分级输出到 slog。
func (l *gormSlogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	attrs := []slog.Attr{
		slog.String("sql", sql),
		slog.Int64("rows", rows),
		slog.Duration("elapsed", elapsed),
	}
	switch {
	// 出错 → Error；忽略 RecordNotFound（正常业务未命中，不当错误刷屏）
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		l.log.LogAttrs(ctx, slog.LevelError, "sql error", append(attrs, slog.Any("err", err))...)
	// 慢查询 → Warn
	case l.slowThreshold > 0 && elapsed > l.slowThreshold:
		l.log.LogAttrs(ctx, slog.LevelWarn, "slow sql", append(attrs, slog.Duration("threshold", l.slowThreshold))...)
	// 正常 → Info
	default:
		l.log.LogAttrs(ctx, slog.LevelInfo, "sql", attrs...)
	}
}
```

**设计要点**：
- **三分支分级**：出错 → Error、慢查询（>200ms）→ Warn、正常 → Info，对齐 GORM 原生 logger 语义。
- **忽略 `ErrRecordNotFound`**：`First()` 没查到是正常业务分支，不当错误刷屏。
- **级别门禁交给 slog**：`LogMode` 直接返回自身，不在适配器里做级别过滤——slog handler 已按 `cfg.Log.Level` 过滤。这样 SQL 日志的开关和其他日志统一由一个配置控制。
- **全程 `*Context` 变体**：把 `ctx` 透传给 slog，请求级字段（request_id 等）由中间件注入 ctx，SQL 日志自动带上——不用在这里关心具体字段。

---

## 9. Repository 与事务约定（重建派）

### 9.1 结论：重建派

跨 repo 事务在 Go 生态有三种范式（重建派 / 显式派 / ctx 派），社区**没有多数派**。本项目选**重建派**——与 老项目 B 生产验证过的写法一致，团队心智统一、零迁移成本。

- repo 是结构体，构造吃 `*gorm.DB`，内部 `query.Use(db)` 得到 `q`
- 不定义 interface（不 mock，靠集成测试）
- 所有事务：service 层 `s.db.Transaction(func(tx *gorm.DB))` + 闭包内 `NewXxxRepo(tx)` 重建，调 repo 封装好的方法。repo 内部不开事务

#### 9.1.1 三范式对比：为什么选重建派

三种范式的本质是「连接从哪来」的三个答案，都能让 repo 用上事务连接，没有谁对谁错：

| 范式 | 连接从哪来 | 事务里代价 | repo 方法签名 |
|------|-----------|-----------|--------------|
| **重建派**（本项目） | 结构体字段，构造时绑死 | 事务闭包开头 `NewXxxRepo(tx)` 重建 | 干净，无 q |
| 显式派 | 方法参数传入 | 无重建 | 每个方法挂 `q *query.Query` |
| ctx 派（Transactor） | `Conn(ctx)` 运行时从 ctx 取 | 无重建、无传参 | 干净，但连接藏 ctx |

- **为什么否 ctx 派**：把连接塞进 ctx（`Conn(ctx)` 运行时取），好处是零重建零传参，但**连接藏在 ctx 里不可见、易误用**——service 层忘传或传错 ctx，操作会静默跑在非事务连接上，回滚时不回滚。本项目曾写过 `pkg/database/transactor.go` 试这条路，因这个隐患删掉了。
- **为什么否显式派**：每个 repo 方法都挂一个 `q *query.Query` 参数，签名全部要改，啰嗦且传递链长。
- **为什么选重建派**：它把成本摊在 **service 事务闭包那几行 `new`**（少、集中、可见），而不是摊在每个 repo 方法签名上，也不藏进 ctx。加上与 老项目 B 生产写法一致、团队心智统一，综合最优。

### 9.2 repo 构造吃 db，方法签名不带 q

```go
type SystemConfigRepo struct {
	db *gorm.DB
	q  *query.Query
}

func NewSystemConfigRepo(db *gorm.DB) *SystemConfigRepo {
	return &SystemConfigRepo{db: db, q: query.Use(db)}
}

func (r *SystemConfigRepo) GetByKey(ctx context.Context, key string) (*model.SystemConfig, error) {
	return r.q.SystemConfig.WithContext(ctx).Where(r.q.SystemConfig.ConfigKey.Eq(key)).First()
}
```

**为什么必须「重建」**：repo 构造时 `q` 就绑死在结构体字段上（`&Repo{q: query.Use(db)}`）。`q` 是常量，连接换不了。所以事务要用 `tx` 连接，**只能拿 `tx` 重新造一个 repo 实例**——换连接 = 换实例。这是「连接存在结构体里」这个选择的必然结果，不是缺陷。

### 9.3 事务在 service 层重建 repo

单 repo 多步（先删后插）：

```go
func (s *SystemConfigService) ReplaceConfigs(ctx context.Context, configs []*model.SystemConfig) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		txRepo := repository.NewSystemConfigRepo(tx)   // 用 tx 重建
		if err := txRepo.DeleteAll(ctx); err != nil {
			return err
		}
		return txRepo.BatchCreate(ctx, configs)
	})
}
```

跨多 repo 事务（核心）：

```go
func (s *CustomResourceService) DeleteCategory(ctx context.Context, categoryID string, personIDs []string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		txPersonRepo := repository.NewCustomResourcePersonRepo(tx)
		txImageRepo := repository.NewCustomResourceImageRepo(tx)
		if len(personIDs) > 0 {
			if err := txImageRepo.BatchDeleteByPersonIDs(ctx, personIDs); err != nil {
				return xerr.Wrap(xerr.ErrInternal.Code, "删除照片失败", err)
			}
		}
		return txPersonRepo.BatchDeleteByIDs(ctx, personIDs)
	})
}
```

### 9.4 两条最易踩的坑

1. **事务闭包内禁用 `s.xxxRepo`**：闭包内全部用 `tx` 重建的 repo。混用 `s.imageRepo`（base 版，持的是非事务 db）会让该操作**静默跑在事务外**，回滚时不回滚，数据不一致。**约定：事务版 repo 变量名以 `tx` 开头**，一眼能看出。
2. **事务闭包内 error 必须透传**：任何一步 error 必须 `return`，中途吞掉会导致本该回滚的事务被提交。用 `xerr.Wrap` 包装（前端才不会收到裸的 `pq: foreign_key_violation`）。

> 完整论证（为什么选重建派、否掉 ctx 派/显式派、社区无多数派）见 `docs/saas-backend/research/database/07-repo层与事务范式选型.md` 和 `.claude/rules/repo-transaction-convention.md`。

### 9.5 WithTx 变体：流式克隆，还是独立构造函数？

有个常见的「改进」提案——给 repo 加一个 `WithTx` 方法，事务里用它派生事务版 repo，替代直接调构造函数：

```go
// 提案：流式克隆
func (r *SystemConfigRepo) WithTx(tx *gorm.DB) *SystemConfigRepo {
    return &SystemConfigRepo{db: tx, q: query.Use(tx)}
}
// service 里：s.configRepo.WithTx(tx)   vs   现状：NewSystemConfigRepo(tx)
```

**先破一个误解**：`WithTx` **并没有「避免重建」**——它的方法体 `return &SystemConfigRepo{...}` 本身就是一次重建，只是把 `NewXxxRepo(tx)` 换了个名字叫 `r.WithTx(tx)`。`query.Use` 在构造时就把连接绑死进 `q` 字段，事务要用 `tx` 这个不同连接，就必然要一个绑 `tx` 的新实例——**重建是「把 q 缓存进结构体」的必然结果，任何写法都绕不开**。所以这是**同一范式内的平级选择**，不是升级。

网上给 `WithTx` 的三条「优势」，逐条对本项目核验：

| 论点 | 是否成立 |
|---|---|
| ① 构造细节收拢进 repo（未来加 cache 字段只改 WithTx） | 真实，但**本项目用不上**——约定钉死 repo 构造只吃 `*gorm.DB`，签名不漂移 |
| ② 更符合 DDD「工作单元」语义 | 修辞——WithTx 没实现 Unit of Work，事务边界已经是 `s.db.Transaction` 闭包本身；且与本项目「套 repo 只为收敛 CRUD，不追 DDD 纯度」的哲学冲突 |
| ③ repo 层划清边界、可换 ORM、可 Mock | 跑题——这是在讲「要不要 repo 层」；本项目主动放弃了 interface/mock（不写 interface、靠集成测试），边界本就是漏的 |

**反向成本**也实打实：破坏与 老项目 B 的一致性（老项目用 `NewXxxRepo(tx)`，而「与老项目一致」正是选重建派的原始理由）、每个 repo 多一个样板方法、字段名若叫 `query` 还会遮蔽 `query` 包（现状用 `q` 正是躲这个）。

**结论**：`WithTx` 不是错，是**风格偏好**。对本项目，它的收益（构造收拢）因约定而用不上，代价（一致性、样板、命名坑）却实打实——**现状的 `NewXxxRepo(tx)` 就是更合适的选择**。完整评估见 [07-repo层与事务范式选型](../saas-backend/research/database/07-repo层与事务范式选型.md) 第五节。

---


## 10. 演进策略：字段和数据怎么改才不出事

前面九章解决的是「从零建库、灌初始数据、生成代码、写事务」。但项目上线跑久了，真正天天碰到、也最容易出事故的是**演进**——加个 `NOT NULL` 字段、重命名一列、几百万行历史数据怎么回填。这一章讲规范，业界已经很成熟，核心就一句话：

> **任何破坏性变更，都拆成若干个「向后兼容的小步」，每步单独部署、可回滚——绝不「一步到位」。**

### 10.1 核心模式：Expand-Contract（扩展-收缩）

灰度/滚动发布期间，**新旧两版应用代码会同时连着同一个库**。若一步 `ALTER TABLE users RENAME COLUMN email TO email_address`，灰度期老实例还在执行 `SELECT email` → 立即报「列不存在」，整个发布窗口都在故障。

正解是拆成三个各自向后兼容的阶段，每阶段一次独立部署：

| 阶段 | 动作 | 数据库状态 | 应用行为 |
|---|---|---|---|
| **① Expand（扩展）** | 加新结构，**旧结构保留** | 新旧列并存 | 老代码照常用旧列，无感知 |
| **② Migrate（过渡）** | 双写 + backfill 历史数据 | 新旧列都有完整数据 | 新代码同时写新旧、读时以新为准 |
| **③ Contract（收缩）** | 删旧结构 | 只剩新列 | 所有实例已只用新列，安全删旧 |

关键：**相邻两阶段之间，新旧代码可以任意混跑**——这正是滚动发布与随时回滚的安全保证。落到本项目的 migration 序列上，就是三个（组）编号 migration：`ddl_add_email_address` → `data_backfill_email` → `ddl_drop_email`。

### 10.2 Schema 演进 vs Data 演进：必须分离

上面的「改列名」混杂了两件事：① 加列（schema migration），② 把几百万行 `email` 拷到 `email_address`（data backfill）。业界铁律：**这俩必须分开**，因为性质完全相反。

| 维度 | Schema Migration | Data Backfill |
|---|---|---|
| 内容 | `ADD COLUMN` / `ADD INDEX` | 填默认值、算派生字段、清洗 |
| 事务 | 在事务内（DDL 失败整体回滚） | **事务外**（大表长事务会锁表） |
| 耗时 | 毫秒~秒（PG 多数 DDL 是元数据操作） | 分钟~小时（正比于数据量） |
| 幂等 | 不强制（只跑一次） | **必须幂等**（中断重跑不重复/遗漏） |
| 工具 | golang-migrate 自动跑 `.sql` | 独立 Go 脚本 `scripts/backfill_*.go` |

**反例**（data 混进 schema migration）：`ALTER TABLE ... ADD COLUMN` 后面直接跟一条全表 `UPDATE`——百万行 UPDATE 扫全表要几十秒，这段时间表被长事务锁住，所有 INSERT/UPDATE 卡死，且跑到一半崩了没法幂等重跑。

**正解**：migration 只加列（秒完），backfill 用独立脚本分批 + 限速 + 幂等：

```go
// scripts/backfill_user_status/main.go —— 慢但安全
func main() {
	var lastID int64
	for {
		// 每批 1000 行，WHERE ... IS NULL 保证幂等（已填过的跳过）
		affected := db.Exec(`
			UPDATE users SET status = 'active'
			WHERE status IS NULL AND id > ? ORDER BY id LIMIT 1000
		`, lastID).RowsAffected
		if affected == 0 {
			break
		}
		time.Sleep(100 * time.Millisecond) // 限速，避免压垮库
	}
}
```

> 这和第 6 章的 data migration 是两回事：第 6 章的 `data_` migration 是**配置数据**（菜单/字典，声明式、幂等、进版本序列）；这里的 backfill 是**业务历史数据的一次性转换**（过程式、跑完即弃、独立脚本）。两者都遵守「data 与 schema 分离」。

### 10.3 PostgreSQL 安全 DDL 清单

并非所有 DDL 都秒完。常见操作的锁表风险与安全写法：

| 操作 | 风险 | 安全做法 |
|---|---|---|
| `ADD COLUMN`（无默认/非空） | ✅ 不锁，仅元数据 | 直接执行 |
| `CREATE INDEX` | ❌ 锁表，阻塞 DML | `CREATE INDEX CONCURRENTLY` |
| `ADD CONSTRAINT`（CHECK/FK） | ❌ 验证阶段全表扫描 | `... NOT VALID` → 后台 `VALIDATE CONSTRAINT` |
| `SET NOT NULL` | ❌ 全表扫描验证 | 加 `CHECK(col IS NOT NULL) NOT VALID` → `VALIDATE` → backfill → `SET NOT NULL` |
| `ALTER COLUMN TYPE` | ❌ 全表重写 | 走 expand-contract（加新列→backfill→删旧列） |

迁移文件头部建议加兜底：

```sql
SET lock_timeout = '2s';        -- 拿锁超 2 秒放弃，避免排队雪崩
SET statement_timeout = '30s';  -- 单条 SQL 超 30 秒强制中断
```

### 10.4 按规模裁剪，不做教条

expand-contract 把一次改动拆成 3 次部署、3 个 PR，代价不小。**本项目（教学/中等负载，表长期 < 10 万行）绝大多数场景用不到**——停服几秒直接 `ALTER` 反而更省事。这一节的价值是**知道天花板在哪**：等哪天某张表真涨到千万行，不至于一条 `ALTER` 锁库几小时才反应过来。

> 完整规范（Fowler 演化式数据库四原则、gh-ost/pg_repack 大表工具、更详细的 backfill 模板）见 `docs/saas-backend/research/database/05-数据库演进与迁移规范.md`。

## 11. 承认的 tradeoff（不回避）

没有免费的方案，这套组合的代价都摆在台面上：

- **database-first 要手动 `make gen-db`**：改表后多一步生成 model。Atlas 的 struct-first auto-diff 能省这步，但要多学一套工具、多一层概念。接受这一步，换工具链简单、与老项目一致。
- **data migration append-only，小文件会变多**：菜单/字典频繁改会攒下一堆 `data_xxx` / `fix_yyy` migration。但本项目这类改动低频，可控；换来的是每个变更可独立回滚、序列清晰可追溯——比 老项目 B 的 `_full` 全量重刷干净得多。
- **golang-migrate 的 dirty state**：迁移中途失败会把 `schema_migrations` 标记为 dirty，此后所有迁移被拒，需手动 `migrate force <version>` 修。这是 golang-migrate 的固有代价——好在开发期 `make reset` 一键重建规避了多数场景，生产则靠迁移前充分测试 + 小步提交。
- **重建派事务闭包啰嗦**：跨 repo 事务开头有一堆 `txXxxRepo := NewXxxRepo(tx)`。已知并接受——换来的是 repo 方法签名干净、连接来源显式可见（不藏 ctx）。
- **暂不上 Atlas 的代价**：放弃了 `migrate lint`（CI 自动拦截破坏性/锁表 DDL）和自动 diff。当前靠人工 review + 本文的安全 DDL 清单兜底；等 schema 变更频繁、团队变大，再引入 Atlas 也不迟——两者迁移文件格式兼容，迁移平滑。

## 12. 参考链接

### 本项目研究文档（选型/规范论证的完整出处）

- [02-数据库访问层选型调研](../saas-backend/research/database/02-数据库访问层选型调研.md)（7 条约束、sqlc 动态查询硬伤、Bob 对比）
- [03-迁移工具与数据初始化方案](../saas-backend/research/database/03-迁移工具与数据初始化方案.md)（golang-migrate vs Atlas、seed 三层分类、api_resource 路由同步）
- [04-schema优先与数据库优先](../saas-backend/research/database/04-schema优先与数据库优先.md)（两种真相源方向、PG 满血论证）
- [05-数据库演进与迁移规范](../saas-backend/research/database/05-数据库演进与迁移规范.md)（expand-contract、安全 DDL、Fowler 四原则）
- [06-GORM与golang-migrate最佳实践](../saas-backend/research/database/06-GORM与golang-migrate最佳实践.md)（六条约定、老项目病根实证）
- [07-repo层与事务范式选型](../saas-backend/research/database/07-repo层与事务范式选型.md)（重建派论证）
- [08-GORM与原生SQL对比](../saas-backend/research/database/08-GORM与原生SQL对比.md)（AI 时代为何仍要 ORM、编译期类型安全是刚需、90/10 混合边界）
- [09-Ent与GORM迁移工具链对比](../saas-backend/research/database/09-Ent与GORM迁移工具链对比.md)（「重」不可分离、data migration 才是大头、三条第三方实证）
- [10-gorm-gen与泛型CLI对比](../saas-backend/research/database/10-gorm-gen与泛型CLI对比.md)（gen vs 官方泛型 CLI、语料丰富度第一性原理）

### 项目落地约定（`.claude/rules`）

- `migration-data-convention.md`（单一 migrations/ 目录、`ddl_`/`data_`/`fix_` 前缀、幂等 + 内联 ID）
- `repo-transaction-convention.md`（重建派、事务在 service 层重建 repo）
- `junction-table-design.md`（关联表独立主键 + 唯一约束）
- `list-query-ordering.md`（分页双排序字段兜底）

### 官方文档与经典文章

- [gorm.io/gen](https://github.com/go-gorm/gen)（类型安全 model/query 生成）
- [golang-migrate/migrate](https://github.com/golang-migrate/migrate)（迁移执行器）
- [GORM — Upsert / On Conflict](https://gorm.io/docs/create.html#Upsert-On-Conflict)（幂等 data migration 基础）
- [sqlc Discussion #364 — Support dynamic queries](https://github.com/sqlc-dev/sqlc/discussions/364)（动态查询硬伤）
- [Evolutionary Database Design — Martin Fowler](https://www.martinfowler.com/articles/evodb.html)（演化式数据库四原则）
- [Parallel Change — Martin Fowler](https://martinfowler.com/bliki/ParallelChange.html)（expand-contract 模式命名）
- [The "dirty secret" of golang-migrate — Atlas 博客](https://atlasgo.io/blog/2025/04/06/golang-migrate-dirty-secret)（dirty state 痛点）
- [gorm.io/cli — GORM 官方泛型 CLI](https://gorm.io/cli)（`gorm.G[T]` 无状态泛型、字段助手 + SQL 模板；含 `cli_vs_gen` 对照页）
- [entgo.io — Versioned Migrations](https://entgo.io/docs/versioned-migrations/)（Ent 从 auto 到版本化迁移的演进，2022 补齐）

---

> 姊妹篇：[从零设计 Go 结构化日志](./从零设计Go结构化日志-slog封装与Gin集成.md)、[从零设计 Go 配置加载](./从零设计Go配置加载-viper封装与约定式设计.md)。三篇共同构成本项目基础设施层（日志 / 配置 / 数据库）的设计实录。
