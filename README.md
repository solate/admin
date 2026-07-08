# admin

一个完整的后端管理系统

## 📦 技术栈

- **Web 框架**: Gin
- **ORM**: GORM + GORM/Gen（代码生成）
- **数据库**: PostgreSQL
- **迁移工具**: golang-migrate
- **认证**: JWT
- **权限**: 纯数据库 RBAC（PermissionCache 内存缓存，无 Casbin）
- **日志**: slog（标准库）+ pkg/xslog 封装
- **配置**: Viper（pkg/xviper 泛型封装 + internal/config 类型化）

## 🤖 开发原则：AI 友好优先

本项目后续开发主要由 AI 辅助完成。定架构、定规范时，**AI 友好度 > 理论完美性**：

- **单一目录、单一规则**：避免多目录让 AI 每次判断「这改动该放哪」。如数据库变更统一走 `migrations/` 一个目录，规则只有一条——改库 = 新建编号 migration。
- **机械化 append-only**：AI 新建文件、不改已应用文件，自动满足 migration 不可变黄金规则。
- **原生格式优先**：能用 SQL 就不用 YAML→struct→ORM 翻译层，减少 AI 出错的中间环节。
- **文件名即分类**：靠命名前缀（`ddl_` / `data_` / `fix_`）区分，人和 AI 都一眼可辨，不靠目录拆分。

## 📐 开发规范（AI 每次会话自动加载）

所有可执行铁律见 [`.claude/rules/`](.claude/rules/)，其中数据库相关：

- [migration-data-convention.md](.claude/rules/migration-data-convention.md) — 迁移与数据变更（单目录、命名前缀、幂等 data migration、ID 内联）
- [junction-table-design.md](.claude/rules/junction-table-design.md) — 关联表设计
- [list-query-ordering.md](.claude/rules/list-query-ordering.md) — 列表查询双排序字段
- [dto-id-type.md](.claude/rules/dto-id-type.md) — DTO ID 用 string

论证与调研（人查阅，不塞给 AI 每次读）见 [docs/saas-backend/research/database/](docs/saas-backend/research/database/)，
其中 [06-GORM与golang-migrate最佳实践.md](docs/saas-backend/research/database/06-GORM与golang-migrate最佳实践.md)
记录了迁移/数据管理方案的三轮讨论、成熟项目调研与 tradeoff。
