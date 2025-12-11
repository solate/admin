# Admin Backend - 后端

> 基于 Go + Gin + GORM/Gen + PostgreSQL 的管理后台骨架

## 🚀 快速开始

### 前置要求

- Go 1.24+
- PostgreSQL 15+（本地或容器）
- `golang-migrate`（数据库迁移）
- `swag`（Swagger 文档生成）
- 可选：`golangci-lint`（代码检查）

安装工具示例：

```bash
# 安装迁移和文档工具
brew install golang-migrate            # macOS（或参考官方安装）
go install github.com/swaggo/swag/cmd/swag@latest

# 可选安装代码检查工具
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```


## 💡 开发模式

本项目采用 **Database First** 开发模式：

1. 编写 SQL Migration → migrations/*.sql
2. 执行迁移 → `make migrate-up`
3. 生成代码 → `make gen-db` 从数据库生成 Model 和 Query
4. 业务开发 → 在 Service/Handler 中使用生成的代码

环境配置与覆盖：

- 配置文件位于 `config/`（`config.yaml` + `config.{env}.yaml`）
- 通过环境变量 `APP_ENV` 指定环境（`dev`/`prod`），或使用 `GIN_MODE`（`release`→`prod`）
- 支持环境变量覆盖配置，如 `DATABASE_HOST`、`DATABASE_PORT`

## 📦 技术栈

- **Web 框架**: Gin
- **ORM**: GORM + GORM/Gen（代码生成）
- **数据库**: PostgreSQL
- **迁移工具**: golang-migrate
- **认证**: JWT
- **权限**: Casbin
- **日志**: Zerolog
- **配置**: Viper

## 🛠️ 常用命令

```bash
# 开发
make dev              # 一键开发（migrate + gen-db + run）
make run              # 只运行应用（不迁移不生成）
make init             # 首次初始化项目

# 数据库迁移
make migrate-up       # 执行迁移
make migrate-down     # 回滚一步
make migrate-reset    # 完全重置数据库（谨慎使用）
make migrate-create NAME=xxx  # 创建新迁移

# 代码生成
make gen-db           # 从数据库生成 Model + Query 代码

# 其他
make help             # 查看所有命令
make fmt              # 格式化代码
make test             # 运行测试
make lint             # 运行代码检查
make swagger          # 生成Swagger文档
```

环境变量（可覆盖默认）

```bash
# 数据库（供 make/migrate/gen-db 使用）
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=root
export DB_PASSWORD=root
export DB_NAME=admin_db
```

生成代码说明：

- `scripts/gen_from_db.go` 读取数据库结构，生成 `internal/dal/query/gen.go`
- 默认排除 `schema_migrations` 表；如需调整，修改脚本内的排除列表
- 仅用于 Query 层生成，Model 可按需扩展在 `internal/dal/model`


## 📁 项目结构

```
backend/
├── cmd/
│   └── server/
│       └── main.go
├── config/
│   ├── config.yaml
│   ├── config.dev.yaml
│   └── config.prod.yaml
├── internal/
│   ├── constants/
│   │   └── status.go
│   ├── handler/
│   │   └── health_handler.go
│   ├── middleware/
│   │   ├── cors.go
│   │   ├── logger.go
│   │   ├── rate_limit.go
│   │   ├── recovery.go
│   │   └── request_id.go
│   └── router/
│       └── router.go
├── migrations/
│   ├── 000001_create_users.up.sql
│   └── 000001_create_users.down.sql
├── pkg/
│   ├── common/
│   │   ├── common.go
│   │   └── page.go
│   ├── config/
│   │   ├── config.go
│   │   ├── viper.go
│   │   └── viper_test.go
│   ├── database/
│   │   └── postgres.go
│   ├── errors/
│   │   ├── codes.go
│   │   ├── errors.go
│   │   └── http.go
│   ├── idgen/
│   │   ├── idgen.go
│   │   ├── idgen_test.go
│   │   └── machine_id.go
│   ├── logger/
│   │   ├── logger.go
│   │   └── logger_test.go
│   ├── response/
│   │   └── response.go
│   └── validator/
│       └── validator.go
├── scripts/
│   ├── dev-reset.sh
│   ├── dev_schema.sql
│   ├── gen_from_db.go
│   └── init_db.sh
├── .gitignore
├── Makefile
├── go.mod
├── go.sum
└── README.md
```
