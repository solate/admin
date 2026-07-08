# 数据库迁移与数据变更规范

> 本规则优先级：**AI 友好度 > 理论完美性**。单一目录、单一规则、机械化 append-only。
> 论证与调研出处见 `docs/saas-backend/research/database/06-GORM与golang-migrate最佳实践.md`。

## 规则 1：只有 migrations/ 一个目录

所有数据库变更（schema + data + fix）都是 migration 文件，放在 `migrations/`。

**AI 的规则只有一条**：改数据库 = 新建编号 migration（`.up.sql` + `.down.sql`）。不做「这该放 schema/seed/correction 哪个目录」的分类判断。

**禁止**：
- 禁止创建 `seeds/` 目录、YAML seed 文件、`scripts/apply_seed.go` 这类翻译层
- 禁止 `scripts/corrections/` 独立订正脚本
- 禁止 `dev_schema.sql` / `dev-reset.sh` 快照建库

## 规则 2：文件名前缀即分类（人和 AI 都靠它识别）

命名：`migrations/{6位序号}_{前缀}_{描述}.up.sql`（+ 对应 `.down.sql`）

- 结构类前缀：`ddl_`（建表/加列/加索引/改约束）
- 数据类前缀：`data_`（INSERT/UPDATE 配置数据）
- 修复类前缀：`fix_`（订正错误，数据或结构均可）

```
migrations/
├── 000001_ddl_init_schema.up.sql       # 结构 - 建表
├── 000002_data_base_config.up.sql      # 数据 - 初始菜单/字典/角色
├── 000003_ddl_add_video_tables.up.sql  # 结构 - 加表
├── 000004_data_video_menus.up.sql      # 数据 - 加视频菜单（依赖 000003）
├── 000005_fix_menu_icon.up.sql         # 修复 - 改错的图标
```

前缀做分类、序号做时间线。**不拆多目录**——拆目录会丢掉 DDL↔data 的交错执行顺序（如「加列→填数据→加 NOT NULL 约束」）。

## 规则 3：data migration 必须幂等

用 `INSERT ... ON CONFLICT (自然键) DO UPDATE SET ...`，可重跑、收敛到声明状态。

```sql
-- ✅ 正确：幂等
INSERT INTO menus (menu_id, menu_code, name, icon, sort, parent_code, created_at, updated_at) VALUES
  ('153547313510393194', 'system', '系统管理', 'settings', 1, '',
   EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000)
ON CONFLICT (menu_code) DO UPDATE SET
  name = EXCLUDED.name, icon = EXCLUDED.icon, sort = EXCLUDED.sort,
  updated_at = EXCLUDED.updated_at;

-- ❌ 错误：裸 INSERT，重跑 duplicate key error
INSERT INTO menus (...) VALUES (...);

-- ❌ 错误：count-then-insert，非幂等，重跑会漏
-- if count > 0 { return }
```

## 规则 4：ID 硬编码在 SQL 里

本项目主键是 **Sonyflake 数字字符串**（`varchar(20)`，如 `'153547313510393194'`，非 UUID、非自增）。

- ✅ ID 直接写进 VALUES；用 Go 的 `idgen.GenerateUUID()` 预生成后粘贴进 SQL
- ❌ 禁止 `gen_random_uuid()`（每环境 ID 不同，关联表外键无法硬编码）
- ❌ 禁止 `idgen.GenerateUUIDs(248)` 批量生成再手动切片（数量一变静默错位）

时间戳统一 `EXTRACT(EPOCH FROM NOW())::BIGINT * 1000`（毫秒 bigint）。

## 规则 5：初始数据一次定义，后续只增量

- `000002_data_base_config.up.sql`：用 `INSERT ON CONFLICT` **一次性幂等定义**全部基础菜单/字典/角色
- 后续版本：新建 `000NNN_data_xxx.up.sql` 只加 genuinely-new 行，或 `000NNN_fix_xxx.up.sql` 显式改/删
- **禁止 `_full` 全量重刷**（content-center 的 `_for_v1_1_0` / `_for_v1_1_0_full` 就是反面教材）

## 规则 6：每个 up 必须配 down

- **结构 down**：`DROP TABLE` / `DROP COLUMN` / `DROP INDEX`
- **数据 down**：`DELETE FROM menus WHERE menu_code IN ('video', 'video_list');` 显式列出本次范围；UPDATE 类改回旧值

生产以 fix-forward 为主（错了写新 migration 改回来），down 主要服务本地 `make reset` 全量重建，但仍必须写（golang-migrate 要求成对）。

## 规则 7：api_resource 从路由自动同步，禁止手写

`api_resources` 表（唯一键 `path + method`，无 `handler` 列）不写任何 data migration。启动时遍历 Gin `Engine.Routes()` upsert（`internal/router/sync_api_resource.go`）。路由注册是唯一真相源。

## 规则 8：Makefile 是唯一入口

`migrate-up` / `migrate-down` / `migrate-reset` / `migrate-create` / `gen-db` / `reset`（= `migrate-reset` + `gen-db`）。

**没有 `seed` target**——数据是 data migration，`migrate-up` 一次跑完 schema + data。禁止绕过 Makefile 的 shell 脚本。

## 规则 9：改表后重跑 gen-db

`migrations/` 是真相源，改表结构后跑 `make gen-db`（gorm/gen 反射活库重新生成 `internal/dal/model` + `internal/dal/query`）。不要手改 `*.gen.go`（会被覆盖）。

## 检查命令

```bash
# migrations 目录不该出现的东西
ls seeds/ scripts/apply_seed.go scripts/dev_schema.sql 2>/dev/null && echo "❌ 存在违规文件"
# data migration 是否幂等（有 INSERT 但无 ON CONFLICT）
grep -rLi "on conflict" $(grep -rl "INSERT INTO" migrations/*data*.up.sql 2>/dev/null) 2>/dev/null
```

---

**最后更新**：2026-07-07（AI 友好度优先，单一 migrations/ 目录 + `ddl_`/`data_`/`fix_` 前缀，data 走 migration 不用 YAML seed）
