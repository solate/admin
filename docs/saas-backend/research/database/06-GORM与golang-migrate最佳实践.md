# GORM 与 golang-migrate 最佳实践

> 本轮（2026-07）承接 [02-数据库访问层选型调研.md](./02-数据库访问层选型调研.md)、[03-迁移工具与数据初始化方案.md](./03-迁移工具与数据初始化方案.md)、[04-schema优先与数据库优先.md](./04-schema优先与数据库优先.md)、[05-数据库演进与迁移规范.md](./05-数据库演进与迁移规范.md)。前五份是「选型/真相源/演进规范」，本轮回答一个实际问题——**如果用 gorm + golang-migrate，怎么避免老项目（backend-rbac / content-center-backend）的混乱？**

## 一、背景与本轮问题

前五份文档帮你理清了「用什么工具、schema 真相源在哪、演进规范是什么」，但还有一个更现实的问题悬着：

> **"如果还是准备用 gorm 来迁移 golang-migrate 这种模式，我应该怎么来优化原来的问题？现在 content-center-backend 就碰到很多问题，怎么来避免？数据库表和数据变更在现在这种调研了这么多的情况下，怎么避免和演进，让整个项目简单，不用操心框架的东西，可以关注业务开发？"**

我探索了你的两个老项目（backend-rbac、content-center-backend），发现它们的混乱**不是 gorm 或 golang-migrate 本身的问题**，而是**缺少明确约定**，导致：

### backend-rbac 的 3/12/57 schema drift

| 源 | 位置 | 表数量 | 说明 |
|---|---|---|---|
| migrations | `migrations/*.up.sql` | **3 个 CREATE TABLE** | 仅 users / user_roles / role_permissions |
| dev schema | `scripts/dev_schema.sql` | **12 个表** | tenants / users / roles / menus / permissions / depts / positions / ... |
| generated models | `internal/dal/model/*.gen.go` | **57 个模型** | 包含 face/device/video/casbin_rule/api_resources 等，反映真实库 |

**三套路径互不一致**：`migrate up` 建 3 表、`dev-reset.sh` 加载 dev_schema 建 12 表、gen-db 生成 57 model 反映线上真实库。哪个是真相源？没人说得清。

**种子数据冲突**：Go seeder (`scripts/init_data/`) 不在 Makefile、手动触发、未文档化，且与 `migrations/000001` 的 `INSERT users` 硬编码（admin/auditor/uploader）冲突——两套 user seed，一个在迁移、一个在 seeder。

### content-center-backend 的五大痛点（62 个 migration）

1. **api_resource 三源维护**：Go seeder (`scripts/init_data/seeds/api_resource.go`, 629 行) + migration INSERT（`000046`/`000056`/`000062`）+ 独立 SQL 脚本（44KB `insert_api_resources_data.sql`）——同一份数据三处维护，必然 drift。
2. **UUID 手工计数**：seeder `main.go` 注释「6 + 29 menu + 19 dept + 37 position + 52 dict + 105 API = 248」，`idgen.GenerateUUIDs(248)` 后手动推进 `idIndex` 切片——数量一变静默错位，典型维护陷阱。
3. **版本补丁式迁移**：`000056_add_api_resources_for_v1_1_0.up.sql` 插 1 条，后来 `000062_add_api_resources_for_v1_1_0_full.up.sql` 再「full」重做插 5 条——同一份数据两次打补丁。
4. **Casbin 与 api_resource 双源同步**：migration 000007 注释「此表仅用于元数据管理和前端展示，不用于权限检查。权限检查由 Casbin 负责」——元数据一处（api_resource）、鉴权另一处（Casbin policies），手动保持同步。
5. **dev_schema.sql 快照与 migration 并存**：`scripts/dev-reset.sh` 加载 50KB `dev_schema.sql` 快照，本地重置不走 migration 链——两套建库路径，快照渐渐过期。

### 关键洞察

**这些痛全是「流程/约定」问题，不是「工具」问题。** 换 sqlc 一个都解决不了——因为它们发生在迁移和种子数据层，与查询层用什么无关。而且你的新 admin 项目已经架构性消掉一部分痛（Casbin 完全移除，见 `.claude/rules/rbac-multi-tenant.md`），所以 content-center 的「Casbin/api_resource 双源」在新项目根本不存在。

**正解**：不换工具，定一套明确约定，把 content-center 的每个坑逐条堵死。

## 二、为什么不选 sqlc（快速否定）

在给出约定前，先快速排除「是否该换 sqlc」这条岔路。

### sqlc 的动态查询硬伤未解

[02-数据库访问层选型调研.md](./02-数据库访问层选型调研.md) 约束 3 已论证：sqlc 对动态 WHERE 的支持仍是 workaround——`sqlc.narg` / `sqlc.slice` 本质是 `COALESCE` hack，无法优雅处理「用户可能传 0 ~ 5 个过滤条件」的常见场景。2026 年这一点未变（[reintech.io 2026 对比](https://reintech.io/blog/sqlc-vs-gorm-vs-sqlx-go-database-libraries-compared-2026) / [glukhov.org 对比](https://www.glukhov.org/app-architecture/data-access/comparing-go-orms-gorm-ent-bun-sqlc/) 仍提这个限制）。

gorm/gen 的链式 `Where().Where()` 或 `Apply(func(dao) dao)` 轻松搞定。这是**查询层的体验差距**。

### content-center 的痛在迁移/seed 层

回看第一节的五大痛点——api_resource 三源、UUID 手工数、版本补丁式迁移、双源同步、快照与 migration 并存——**全都在迁移/种子数据层，与查询层工具零关联**。换成 sqlc 后这些问题一个都不会自动消失，反而查询体验降级。

### sqlc 用户并不难受——只是痛点排序和我们相反

否定 sqlc 时容易陷入一个误区：「写 SQL 那么麻烦，用它的人不难受吗？」——**恰恰相反，sqlc 用户把「必须写 SQL」当收益，不当成本。** 理解这一点，才知道为什么它对我们不划算（而不是简单地「它不好」）。

**两种相反的心智模型**：

- **我们的模型**：SQL 是要被抽象掉的实现细节，想「活在 Go 里」，动态拼条件要顺手。
- **sqlc 用户的模型**：SQL 才是设计语言，Go 只是把结果接住（[SQLAlchemy Core vs ORM 2026](https://thelinuxcode.com/sqlalchemy-core-vs-orm-how-i-choose-the-right-layer-in-2026/) 所谓 "the database is the center of the universe"）。对他们，把脑中现成的 SQL 翻译成 GORM 链式 DSL 才是绕路。

**sqlc 用户真正买单的三个硬收益**（GORM 给不了或给得弱）：

1. **编译期 SQL 校验**：`sqlc generate` 时真的解析 SQL、比对 schema，列名/类型/表写错在生成阶段就报错，到不了运行时（[Bytebase 2025](https://www.bytebase.com/blog/golang-orm-query-builder/)）。GORM 的字符串条件、`Raw()` 往往运行到那行才炸。
2. **零反射、无 ORM 魔法开销**：sqlc 生成的是朴素 `database/sql` 代码。[Cloudflare 的 zero-ORM 实践](https://medium.com/@yashbatra11111/go-postgres-with-sqlc-the-zero-orm-stack-cloudflare-uses-for-99-99-uptime-d4beaddbfdf8)正是冲着「避开 N+1、运行时反射、脆弱迁移」反向选 SQL-first。
3. **所见即所得**：写的 SQL 就是执行的 SQL，调优/看执行计划/DBA review 无需反推 ORM 生成了啥（[leapcell 论 ORM 的 "unwanted magic"](https://www.leapcell.io/blog/embracing-sqlx-for-efficient-and-secure-database-operations-in-go-without-gorm)）。

**为什么这些收益对本项目不划算**：

| sqlc 的收益 | 对本项目是否敏感 |
|---|---|
| 编译期 SQL 校验 | ⚠️ 收益打折——gorm/gen 已生成 `r.q.User.Name` 类型安全字段引用，拼错列名同样编译期报错，已拿到大半 |
| 零反射低延迟 | ❌ 不敏感——教学/中等负载项目，运行时反射非瓶颈 |
| SQL 透明可调优 | ⚠️ 有价值但非刚需——admin 后台查询不复杂 |
| **动态多条件 WHERE** | ❌ **反成致命短板**——admin 列表全是可选过滤，正是 sqlc 死穴、gorm/gen 主场 |

**结论**：sqlc 用「写更多 SQL」换「编译期安全 + 零开销 + 透明」，适合 Cloudflare 那种「延迟敏感、SQL 即设计、宁可多写」的画像；本项目诉求是「少操心框架、专注业务、动态查询顺手」，正好相反。**不是 sqlc 不好，是它的收益我们不敏感、它的短板我们天天踩。**

### 结论

**不换工具，改约定。** 保持 gorm + golang-migrate，用一套清晰的目录规划、Makefile、data migration 模式，把 content-center 的混乱彻底避开。

## 三、核心约定（避坑原则）

每条约定对应 content-center 的一个具体坑：

### 约定 1：单一真相源（禁止 migration / dev_schema / snapshot 三套路径）

**坑**：backend-rbac 有 3 表 migration、12 表 dev_schema、57 model；content-center 有 62 个 migration 但 dev-reset 走快照。

**约定**：**只有一个 canonical schema 定义**，取决于你选的真相源方向（04 文档）：
- **数据库优先**（本文默认）：`migrations/*.up.sql` 唯一真相源，删掉 `dev_schema.sql` / 任何 schema 快照。
- **schema 优先**（如 GORM+Atlas）：Go struct 唯一源，Atlas 自动 diff 生成 migration。

**禁止**：`scripts/dev_schema.sql`、`scripts/dev-reset.sh` 加载快照、任何「不走 migration 的建库捷径」。

**落地**：`make reset` = `migrate-reset`（`drop -f` → `migrate up`）+ `gen-db`，从零重建。

### 约定 2：Schema 与 Data 分文件、同序列（05 文档落地）

**坑**：content-center 把 DDL 与 INSERT 混在同一个 migration（000044/046/056/062），改结构和改数据搅在一起，回滚和 review 都难。

**约定**（关键：分**文件**，不分**目录**）：
- **Schema migration**（`000001_ddl_init_schema.up.sql`）：只写 DDL（`CREATE TABLE` / `ALTER TABLE` / `CREATE INDEX`）。
- **Data migration**（`000002_data_base_config.up.sql`）：只写数据（`INSERT ... ON CONFLICT` / `UPDATE`），**独立成文件，但仍在同一个 `migrations/` 版本序列里**。
- **修复 migration**（`000005_fix_xxx.up.sql`）：数据订正或结构修复，按性质可能是 DDL 也可能是 DML，统一用 `fix_` 前缀。

**命名前缀规范**（人和 AI 都靠它识别）：
- `ddl_` — 结构变更（init / create / alter / drop / add_column 等）
- `data_` — 配置数据（base_config / video_menus / roles 等）
- `fix_` — 错误修复（不论 DDL 还是 data，统一前缀）

**示例**：
```
000001_ddl_init_schema.up.sql      ← DDL
000002_data_base_config.up.sql     ← 初始数据
000003_ddl_add_icon_col.up.sql     ← DDL 加列
000004_data_video_menus.up.sql     ← 数据（依赖 000003）
000005_fix_menu_name.up.sql        ← 修复
```

**为什么不拆出 `seeds/` 目录**：
1. 多一个目录 = AI/人多一个「这改动该放哪」的判断点
2. **DDL 和数据有交错依赖**（数据依赖新列、新列要先 backfill 数据再加约束）——拆目录无法表达这个顺序，单序列自然保证
3. 所有库变更统一走 `migrations/`，规则只有一条——**改库 = 新建编号 migration**，DDL 与 data 靠**文件名前缀**区分

**落地**：不需要 `seed` / `backfill` 独立 target，全部由 `migrate-up` 按序号一次跑完。

### 约定 3：Data migration 幂等 + 内联 ID（禁止手工数 UUID、count-then-insert）

**坑**：content-center 手工数 248 个 UUID（"6+29+19+37+52+105"），数错就错位；backend-rbac 的 menu seeder 用 `Count() > 0` 跳过整个 seed（非幂等）。

**约定**：
- **`INSERT ... ON CONFLICT (unique_key) DO UPDATE`**：以业务自然键（如 `menu_code`）为冲突目标，重跑收敛到声明状态，幂等。
- **ID 内联硬编码**：本项目主键是 Sonyflake 数字字符串（`api_resource_id varchar(20)` 等），由 Go 侧 `idgen.GenerateUUID()` 预生成后**直接写进 SQL 的 VALUES**（如 `'153547313510393194'`）。稳定、跨环境一致、外键引用不会因重跑漂移。
- **禁止 `gen_random_uuid()`**：库内随机生成会让每个环境 ID 不同，关联表无法硬编码外键。
- **禁止手工数量切片**：不写 `GenerateUUIDs(248)` 再手动推进 index——数量一变静默错位。

**禁止**：count-then-insert（`if count > 0 { return }`）——非幂等，重跑会漏。

**落地**：第六节给完整 data migration 模板。

### 约定 4：api_resource 从路由同步（方案 A，03 已定）

**坑**：content-center 的 api_resource 在 Go seeder、migration INSERT、独立 SQL 脚本三处维护。

**约定**：**路由注册是唯一真相源**（03 seed 三层分类的方案 A）。启动时遍历 Gin `Engine.Routes()`，自动 upsert 到 `api_resources` 表。

**禁止**：手写 api_resource INSERT（无论在 migration、seeder、还是独立 SQL）。

**落地**：第七节给 Gin 路由同步示例（`internal/router/sync_api_resource.go`）。

### 约定 5：Makefile 标准化（DB 操作不散落）

**坑**：backend-rbac 的 seeder 不在 Makefile、手动触发、未文档化；content-center 的 `dev-reset.sh` 绕过 migration。

**约定**：**Makefile 暴露全部 DB 操作**，禁止隐藏在 shell 脚本里：
- `migrate-up` / `migrate-down` / `migrate-reset` / `migrate-create`
- `gen-db`（gorm/gen 反射活库生成 model/query）
- `reset`（`migrate-reset` + `gen-db`，一键重建本地）

**注意**：没有独立的 `seed` target——数据是 data migration，`migrate-up` 一次跑完 schema + data（约定 2）。

**禁止**：`scripts/dev-reset.sh` 等绕过 Makefile 的自定义脚本。

**落地**：第五节给可 copy 的 Makefile 模板（对齐 `backend-rbac/Makefile` 现有 target）。

### 约定 6：本地环境一致性（reset 从零重建）

**坑**：content-center 的 `dev-reset.sh` 加载 50KB 快照，渐渐与 migration 链 drift。

**约定**：**`make reset` 必须走 migration from scratch**，不走快照。这保证本地环境与生产的 migration 历史完全一致。

**落地**：删掉 `dev_schema.sql` / `dev-reset.sh`，只留 `make reset` = `migrate-reset`（drop + up）+ `gen-db`。

## 四、目录结构规划

一套把上述 6 条约定物化的目录布局：

```
backend-rbac/                      # 项目实际结构
├── migrations/                    # golang-migrate 唯一源（schema + data 都在此）
│   ├── 000001_ddl_init_schema.up.sql        # DDL - 建表
│   ├── 000001_ddl_init_schema.down.sql
│   ├── 000002_data_base_config.up.sql       # 数据 - 初始菜单/字典/角色（幂等 INSERT ON CONFLICT）
│   ├── 000002_data_base_config.down.sql     # DELETE FROM menus WHERE menu_code IN (...);
│   ├── 000003_ddl_add_video_tables.up.sql   # DDL - 加视频相关表
│   ├── 000003_ddl_add_video_tables.down.sql
│   ├── 000004_data_video_menus.up.sql       # 数据 - 新增视频菜单（依赖 000003 的表）
│   ├── 000004_data_video_menus.down.sql
│   ├── 000005_fix_menu_icon.up.sql          # 修复 - 改错的菜单图标
│   ├── 000005_fix_menu_icon.down.sql
│   └── ...                              # DDL/data/fix 靠前缀分类，序号保证交错依赖顺序（约定 2）
├── scripts/
│   └── gen_from_db.go              # gorm/gen 配置（反射活库生成，输出到 internal/dal/）
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
└── Makefile                        # 全部 DB 操作入口（约定 5）
```

**只有一个 `migrations/` 目录**——schema 和 data 都在里面，靠文件名前缀（`ddl_` / `data_` / `fix_`）区分，不拆多目录。序号是全局递增的单一序列，天然表达「先建表、再插数据、又改表、再补数据」这种 DDL↔data 交错依赖——这正是拆两个目录会丢掉的能力（拆目录后无法让 data 步骤插在两个 DDL 步骤之间）。

**已删除**（相比老项目）：
- ❌ `scripts/dev_schema.sql`（快照，违反约定 1/6）
- ❌ `scripts/dev-reset.sh`（绕过 migration，违反约定 6）
- ❌ `scripts/insert_api_resources_data.sql`（api_resource 手写 SQL，违反约定 4）
- ❌ `seeds/*.yaml` + `scripts/apply_seed.go`（YAML 翻译层，多余——数据直接写 data migration SQL）

## 五、Makefile 标准 target 模板

可直接 copy 的 Makefile 片段（约定 5）：

```makefile
# ── 数据库连接（环境变量覆盖）──
DB_HOST     ?= localhost
DB_PORT     ?= 5432
DB_USER     ?= postgres
DB_PASSWORD ?= postgres
DB_NAME     ?= admin
DB_URL      := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# ── 迁移（golang-migrate CLI）──
.PHONY: migrate-up migrate-down migrate-reset migrate-create
migrate-up:                        ## 应用所有未执行的迁移
	migrate -path ./migrations -database "$(DB_URL)" up

migrate-down:                      ## 回滚最后一个迁移
	migrate -path ./migrations -database "$(DB_URL)" down 1

migrate-reset:                     ## 全部回滚后重新应用（危险，需确认）
	@read -p "确认重置所有迁移? [y/N] " c && [ "$$c" = "y" ] || exit 1
	migrate -path ./migrations -database "$(DB_URL)" drop -f
	$(MAKE) migrate-up

migrate-create:                    ## 新建迁移: make migrate-create NAME=create_orders
	migrate create -ext sql -dir ./migrations -seq $(NAME)

# ── 代码生成 ──
.PHONY: gen-db
gen-db:                            ## 从活库反射生成 model/query
	go run scripts/gen_from_db.go

# ── 本地一键重建（约定 6：从零走 migration，不走快照）──
.PHONY: reset
reset:                             ## migrate-reset（drop + up）→ gen-db
	$(MAKE) migrate-reset
	$(MAKE) gen-db
```

**关键点**：
- `reset` 是本地开发的核心——任何时候库脏了，一条命令从零重建，且**走的就是生产同款 migration 链**（约定 6）。`migrate-reset` 内部 `drop -f` + `migrate-up`，data migration 随 schema 一并跑完。
- **没有 `seed` target**——数据是 data migration，`migrate-up` 一次跑完 schema + data（约定 2）。
- 没有任何 `dev_schema.sql` / `dev-reset.sh`（约定 1/6）。

## 六、Data Migration 幂等模板（SQL + 内联 ID）

数据不走 YAML/Go seeder，直接写成 data migration 的 `.up.sql`——SQL 本身就是声明式、可读、可 diff，不需要 `YAML → struct → ORM` 三层翻译。

### `000002_data_base_config.up.sql`（初始配置数据，幂等）

```sql
-- 初始菜单：ID 用 Sonyflake 数字字符串，内联硬编码（约定 3）
-- 这些 ID 由 idgen.GenerateUUID() 预生成后粘贴进来，跨环境固定不变
INSERT INTO menus (menu_id, menu_code, name, icon, sort, parent_code, created_at, updated_at) VALUES
  ('153547313510393194', 'system',      '系统管理', 'settings', 1, '',       EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000),
  ('153547313510393195', 'system_user', '用户管理', 'users',    1, 'system', EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000),
  ('153547313510393196', 'system_role', '角色管理', 'shield',   2, 'system', EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000)
ON CONFLICT (menu_code) DO UPDATE SET
  name        = EXCLUDED.name,
  icon        = EXCLUDED.icon,
  sort        = EXCLUDED.sort,
  parent_code = EXCLUDED.parent_code,
  updated_at  = EXCLUDED.updated_at;
```

### `000002_data_base_config.down.sql`（回滚：显式列出删除范围）

```sql
DELETE FROM menus WHERE menu_code IN ('system', 'system_user', 'system_role');
```

### 关键点

- **幂等**：`ON CONFLICT (menu_code) DO UPDATE`——以业务自然键 `menu_code` 为冲突目标，`make reset` 或重跑都收敛到声明状态，不会 duplicate key error。
- **ID 内联硬编码**：Sonyflake 数字字符串（`'153547313510393194'`，19 位）直接写进 VALUES。用 Go 的 `idgen.GenerateUUID()` 生成后粘贴——**不用 `gen_random_uuid()`**（每环境不同、外键没法硬编码），**不用 `GenerateUUIDs(248)` 批量数**（数量变了静默错位）。
- **时间戳**：`EXTRACT(EPOCH FROM NOW())::BIGINT * 1000`（毫秒），对齐项目 `autoCreateTime:milli` 约定。
- **down 必写**：`DELETE ... WHERE menu_code IN (...)` 显式列出本 migration 加的行，`migrate down` / `make reset` 时干净回退。

### 后续增量（数据随功能持续增多）

加视频功能时的菜单，是一个**新** migration，不改 `000002`：

```sql
-- 000004_data_video_menus.up.sql
INSERT INTO menus (menu_id, menu_code, name, icon, sort, parent_code, created_at, updated_at) VALUES
  ('160123456789012345', 'video',      '视频管理', 'video', 10, '',      EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000),
  ('160123456789012346', 'video_list', '视频列表', 'list',  1,  'video', EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000)
ON CONFLICT (menu_code) DO UPDATE SET
  name = EXCLUDED.name, icon = EXCLUDED.icon, sort = EXCLUDED.sort, updated_at = EXCLUDED.updated_at;
```

```sql
-- 000004_data_video_menus.down.sql —— 只删本次加的，对应"回退视频功能"
DELETE FROM menus WHERE menu_code IN ('video', 'video_list');
```

**菜单从 5 个涨到 500 个 = 往序列后面 append 几十个 `data_xxx` migration**，每个各自带 down、删各自加的那几条。老 migration 一旦提交不再碰（不可变黄金规则）。

**对比 content-center 的做法**：
- ❌ 老：`GenerateUUIDs(248)` 手工数 + count-then-insert 跳过 + `_for_v1_1_0_full` 全量重刷
- ✅ 新：ID 内联、`ON CONFLICT` 幂等、每次只 append 新 migration 加 genuinely-new 行

> **修复怎么办**：改错了名字/图标 → 新建 `000005_fix_menu_icon.up.sql`（`UPDATE menus SET icon=... WHERE menu_code=...`，down 里改回旧值）；加错了不该加的 → `000006_fix_remove_wrong_menu.up.sql`（`DELETE`，down 里 `INSERT` 回来）。**修复也是新 migration，不改老文件**——生产实践以 fix-forward 为主，down 主要服务本地 `make reset` 全量重建。

## 七、api_resource 路由同步示例

`internal/router/sync_api_resource.go`（约定 4：路由是唯一真相源）：

```go
package router

import (
	"context"

	"admin/internal/dal/model"
	"admin/pkg/utils/idgen"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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
			// 注意：api_resources 表无 handler 列。name/module/description
			// 供前端展示，需在路由注册时带元数据写入（本示例从简，仅 path+method）
		})
	}
	if len(resources) == 0 {
		return nil
	}

	// path+method 是唯一键（uk_api_resources_path_method），冲突时更新元数据（约定 4：幂等）
	return db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "path"}, {Name: "method"}},
		DoUpdates: clause.AssignmentColumns([]string{"description", "module", "updated_at"}),
	}).Create(&resources).Error
}
```

启动时调用（`cmd/server/main.go`，`router.Setup` 后、`server.Start` 前）：

```go
r := gin.New()
router.Setup(r, handlers, cfg, jwtMgr, rbacCache)
if err := router.SyncAPIResources(ctx, db, r); err != nil {
	return fmt.Errorf("sync api resources: %w", err)
}
```

**收益**：新增一个接口 → 重启 → api_resource 自动登记，永远和真实路由一致。彻底消灭 content-center 的「三处维护 + 手写 INSERT + 版本补丁」。

> **进阶**：若要软失效（删接口保留历史绑定），可加一步——先把全表标记 `deleted_at`，再对本次路由 upsert 清空 `deleted_at`。本文从简，按需扩展。

## 八、gorm/gen 配置最佳实践

### database-first vs struct-first

- **database-first**（本文默认，延续老项目习惯）：migration 是真相源，`gen_from_db.go` 反射活库生成 model/query。改表流程：写 migration → `make migrate-up` → `make gen-db`。
- **struct-first**（如 02 的 GORM+Atlas）：Go struct 是真相源。本套约定同样适用，只是「改 struct → Atlas diff」替代「写 migration → gen-db」。

### `scripts/gen_from_db.go` 要点

```go
g := gen.NewGenerator(gen.Config{
	OutPath:      "internal/dal/query",
	ModelPkgPath: "internal/dal/model",
	Mode:         gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
})
db, _ := gorm.Open(postgres.Open(dsn))
g.UseDB(db)

// 排除迁移表；其余全生成
tables, _ := db.Migrator().GetTables()
for _, t := range tables {
	if t == "schema_migrations" {
		continue
	}
	g.GenerateModel(t,
		gen.FieldGORMTag("created_at", func(t field.GormTag) field.GormTag { /* autoCreateTime */ return t }),
		// deleted_at → soft_delete.DeletedAt 等字段映射
	)
}
g.Execute() // 注意：只调一次（老项目有重复调用的 code smell）
```

**避坑**（老项目实证）：backend-rbac / content-center 的 `gen_from_db.go` 都有**重复的 `ApplyBasic`/`Execute` 块**（生成跑两遍，无害但脏）。新项目保证 `Execute()` 只调一次。

### 何时重跑 gen-db

- 每次 `make migrate-up` 改了表结构后
- 首次 clone 项目、`make reset` 之后
- **不要**手改 `internal/dal/model/*.gen.go`（会被下次 gen-db 覆盖）

## 九、与前五份文档的衔接

本文是 02-05 的**实操落地**，逐条对应：

| 文档 | 结论 | 本文如何落地 |
|---|---|---|
| **02** | GORM+gen（database-first）满足 7 约束 | 延续 gorm/gen，`gen_from_db.go` 反射活库（第八节） |
| **03** | seed 三层分类 + 方案 A（路由同步）+ 方案 D（YAML seed） | 方案 A → 第七节 sync_api_resource；方案 D **经 2026-07 讨论改为 data migration**（SQL 在 migrations/ 序列里，第六节），不再使用独立 seed 机制 |
| **04** | database-first / schema-first 两种真相源都合法 | 约定 1 明确「单一真相源」，两方向都适用 |
| **05** | expand-contract、schema/data 分离、backfill 分批 | 约定 2 落地为 `migrations/`（DDL 与 data 分文件同序列，`ddl_` / `data_` 前缀区分），backfill 同样是 migration（如 `000058_data_backfill_device_type.up.sql`） |

**一句话**：02-05 讲「为什么这么选、规范是什么」，本文讲「目录怎么摆、Makefile 怎么写、data migration 怎么写才不踩老项目的坑」。

## 十、承认的 tradeoff（不回避）

- **database-first 需要手动 `make gen-db`**：改表后多一步生成。Atlas 的 struct-first auto-diff 能省这步，但 03 已因「约束 6 之外的概念负担」权衡后选了 golang-migrate。接受这一步换取工具简单、与老项目一致。
- **data migration append-only**：配置数据频繁改会有较多小 migration 文件（`data_video_menus` / `data_xxx_menus` / `fix_yyy`）。但本项目菜单/字典改动低频，可控。换来的是每个变更可独立回滚、序列清晰可追溯——比 content-center 的 `_full` 重刷干净。
- **down 回滚的局限**：数据被后续 migration 改过后，down 可能丢失那些修改。真实场景**生产以 fix-forward 为主**（错了写新 migration 改回来），down 主要服务本地 `make reset`。
- **路由同步依赖注册规范**：api_resource 的 `name`/`module`/`description` 要好看，得在路由注册时带元数据（03 已提）。不带则只有 path/method，前端展示朴素。
- **golang-migrate 的 dirty state**：迁移失败会留 dirty version 需手动 `force` 修（03 第二节详述）。这是 golang-migrate 的固有代价，用它就要接受——好在开发期 `make reset` 一键重建规避了多数场景。
- **对比独立 seed 的 tradeoff**：Supabase 式独立 seed 省了 down SQL 编写（声明式重放），但失去版本序列内的原子回滚和 DDL↔data 交错依赖能力。我们选单序列是 AI 友好度优先——一条规则（改库=新 migration）、一个目录、机械化 append-only。

## 十一、落地检查清单

新项目 / review 时逐条自查：

- [ ] **单一真相源**？没有 `dev_schema.sql` / schema 快照，只有 `migrations/`（或 Go struct + Atlas）。
- [ ] **DDL 与 data 分文件、同序列**？`ddl_` / `data_` / `fix_` 前缀区分，都在 `migrations/` 一个目录，不拆多目录。
- [ ] **data migration 幂等**？用 `INSERT ... ON CONFLICT` / `UPDATE ... WHERE`，ID 内联硬编码，可重跑。
- [ ] **`make reset` 能跑通**？从零 `migrate-reset` + `gen-db`，得到完整可跑的库（不依赖快照）。
- [ ] **api_resource 自动同步**？启动时从路由 upsert，没有手写 api_resource INSERT。
- [ ] **无手工数 UUID**？data migration 的 ID 直接写在 VALUES 旁边（预生成后粘贴），不是 `GenerateUUIDs(248)` 批量数再切片。
- [ ] **初始数据一次定义**？`000002_data_base_config` 幂等声明基础配置，后续只 append genuinely-new（`000NNN_data_xxx`）或显式修复（`fix_yyy`），不写 `_full` 重刷。
- [ ] **Makefile 是唯一入口**？migrate/gen-db/reset 全在 Makefile，没有绕过它的 shell 脚本；**没有 `seed` target**（数据是 data migration）。

## 十二、参考链接

- [golang-migrate/migrate](https://github.com/golang-migrate/migrate)（迁移执行器，仍活跃维护）
- [gorm.io/gen](https://github.com/go-gorm/gen)（类型安全 model/query 生成）
- [GORM — Upsert / On Conflict](https://gorm.io/docs/create.html#Upsert-On-Conflict)（幂等 data migration 的 ON CONFLICT 基础）
- [My favourite way to handle SQL in Golang — sqlc + golang-migrate](https://lucianonooijen.com/blog/best-golang-sql-handling/)（sqlc+migrate 组合的对照视角）
- [Go Database Libraries Compared 2026](https://reintech.io/blog/sqlc-vs-gorm-vs-sqlx-go-database-libraries-compared-2026)（sqlc 动态查询限制 2026 仍在）
- **2026-07 调研来源**（单一 migrations/ 目录决策依据，见第十节 tradeoff）：
  - [Supabase — Seeding your database](https://supabase.com/docs/guides/local-development/seeding-your-database)（seed = 幂等初始化，SQL 而非 YAML）
  - [Prisma — Seeding](https://www.prisma.io/docs/orm/v6/prisma-migrate/workflows/seeding)（`prisma db seed` 命令）
  - [The Vibe Coder's Guide to Supabase Environments](https://supabase.com/blog/the-vibe-coders-guide-to-supabase-environments)（单一 db 根目录作为真相源）
  - [How to not destroy your production database](https://blog.rpanachi.com/how-to-not-destroy-your-production-database)（purist：migration 纯 DDL 观点）
  - [Laravel 11 Migrations Best Practices](https://kevdees.com/laravel-11-migrations-best-practices-for-stability-and-scalability)（schema/data 分离）
- 交叉引用：
  - [02-数据库访问层选型调研.md](./02-数据库访问层选型调研.md)（GORM+gen、7 约束、sqlc 动态查询硬伤）
  - [03-迁移工具与数据初始化方案.md](./03-迁移工具与数据初始化方案.md)（seed 三层分类、方案 A/D）
  - [04-schema优先与数据库优先.md](./04-schema优先与数据库优先.md)（两种真相源方向）
  - [05-数据库演进与迁移规范.md](./05-数据库演进与迁移规范.md)（expand-contract、schema/data 分离、backfill）
