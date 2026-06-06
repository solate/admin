.PHONY: help build run test clean migrate-up migrate-down migrate-create migrate-reset gen-db init dev lint fmt swagger

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
APP_NAME := admin-backend
MAIN_PATH := ./cmd/server
BUILD_DIR := ./bin
MIGRATION_DIR := ./migrations

# 数据库配置（可通过环境变量覆盖）
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= postgres
DB_PASSWORD ?= postgres
DB_NAME ?= admin

# 数据库连接字符串
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

## help: 显示帮助信息
help:
	@echo "==================================="
	@echo "Admin Backend - 可用命令"
	@echo "==================================="
	@echo ""
	@echo "🚀 开发命令:"
	@echo "  make dev         - 开发流程（migrate + gen-db + run）"
	@echo "  make init        - 首次初始化项目"
	@echo "  make run         - 仅运行应用"
	@echo ""
	@echo "📦 构建命令:"
	@echo "  make build       - 编译应用"
	@echo "  make clean       - 清理编译文件"
	@echo ""
	@echo "🗄️  数据库迁移:"
	@echo "  make migrate-up       - 执行数据库迁移"
	@echo "  make migrate-down     - 回滚数据库迁移（一步）"
	@echo "  make migrate-reset    - 完全重置数据库（谨慎使用）"
	@echo "  make migrate-create NAME=xxx - 创建新迁移文件"
	@echo ""
	@echo "⚙️  代码生成:"
	@echo "  make gen-db         - 从数据库生成 Model + 查询代码"
	@echo ""
	@echo "📚 文档:"
	@echo "  make swagger        - 生成Swagger文档"
	@echo ""
	@echo "🧪 测试和质量:"
	@echo "  make test           - 运行测试"
	@echo "  make lint           - 运行代码检查"
	@echo "  make fmt            - 格式化代码"
	@echo ""

## dev: 开发流程（migrate + gen-db + run）
dev: migrate-up gen-db
	@echo "🚀 启动应用..."
	@echo ""
	go run $(MAIN_PATH)/main.go

## init: 首次初始化项目
init:
	@echo "======================================"
	@echo "📦 初始化 Admin Backend 项目"
	@echo "======================================"
	@echo ""
	@echo "1️⃣  安装 Go 依赖..."
	go mod download
	@echo ""
	@echo "2️⃣  检查数据库连接..."
	@echo "   数据库: $(DB_HOST):$(DB_PORT)/$(DB_NAME)"
	@echo "   用户: $(DB_USER)"
	@echo ""
	@echo "💡 提示: 确保 PostgreSQL 正在运行（Docker 或本地）"
	@echo ""
	@echo "3️⃣  执行数据库迁移..."
	@$(MAKE) migrate-up
	@echo ""
	@echo "4️⃣  生成 gorm/gen 代码..."
	@$(MAKE) gen-db
	@echo ""
	@echo "======================================"
	@echo "✅ 初始化完成！"
	@echo "======================================"
	@echo ""
	@echo "🚀 运行 'make dev' 启动开发服务器"
	@echo ""

## build: 编译应用
build:
	@echo "🔨 编译应用..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "✅ 编译完成: $(BUILD_DIR)/$(APP_NAME)"

## run: 运行应用
run:
	@echo "🚀 运行应用..."
	go run $(MAIN_PATH)/main.go

## test: 运行测试
test:
	@echo "🧪 运行测试..."
	go test -v -cover ./...

## clean: 清理编译文件
clean:
	@echo "🧹 清理编译文件..."
	rm -rf $(BUILD_DIR)
	rm -rf tmp/
	@echo "✅ 清理完成"

## gen-db: 从数据库生成 Model 和查询代码
gen-db:
	@echo "======================================"
	@echo "⚙️  从数据库生成代码（贫血模型）"
	@echo "======================================"
	@echo ""
	@echo "📍 数据源: PostgreSQL 数据库表"
	@echo "📍 生成目录: internal/dal/model/ + internal/dal/query/"
	@echo ""
	@go run scripts/gen_from_db.go
	@echo ""

## migrate-up: 执行数据库迁移
migrate-up:
	@echo "⬆️  执行数据库迁移..."
	@if command -v migrate > /dev/null; then \
		migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" up; \
		echo "✅ 迁移完成"; \
	else \
		echo ""; \
		echo "❌ golang-migrate 未安装"; \
		echo ""; \
		echo "安装方法:"; \
		echo "  macOS:   brew install golang-migrate"; \
		echo "  Linux:   wget https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz"; \
		echo "  Windows: scoop install migrate"; \
		echo ""; \
		echo "或访问: https://github.com/golang-migrate/migrate"; \
		echo ""; \
		exit 1; \
	fi

## migrate-down: 回滚数据库迁移（一步）
migrate-down:
	@echo "⬇️  回滚数据库迁移（一步）..."
	@if command -v migrate > /dev/null; then \
		migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" down 1; \
		echo "✅ 回滚完成"; \
	else \
		echo "❌ golang-migrate未安装"; \
		exit 1; \
	fi

## migrate-reset: 完全重置数据库（危险操作）
migrate-reset:
	@echo "⚠️  警告：即将完全重置数据库！"
	@echo "按 Ctrl+C 取消，或按回车继续..."
	@read confirm
	@echo "⬇️  回滚所有迁移..."
	@migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" down -all || true
	@echo "⬆️  重新执行迁移..."
	@migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" up
	@echo "✅ 数据库重置完成"

## migrate-create: 创建新的迁移文件 (使用方法: make migrate-create NAME=add_user_table)
migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "❌ 请指定迁移名称: make migrate-create NAME=your_migration_name"; \
		exit 1; \
	fi
	@if command -v migrate > /dev/null; then \
		migrate create -ext sql -dir $(MIGRATION_DIR) -seq $(NAME); \
	else \
		echo "❌ golang-migrate未安装"; \
		exit 1; \
	fi
## lint: 运行代码检查
lint:
	@echo "🔍 运行代码检查..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "❌ golangci-lint未安装"; \
		echo "安装: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

## fmt: 格式化代码
fmt:
	@echo "📝 格式化代码..."
	gofmt -s -w .
	go mod tidy
	@echo "✅ 格式化完成"

## swagger: 生成Swagger文档
swagger:
	@echo "📚 生成Swagger文档..."
	@if command -v swag > /dev/null; then \
		swag init -g cmd/server/main.go -o docs; \
	else \
		echo "❌ swag未安装"; \
		echo "安装: go install github.com/swaggo/swag/cmd/swag@latest"; \
		exit 1; \
	fi

## install-tools: 安装开发工具
install-tools:
	@echo "🔧 安装开发工具..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "✅ 工具安装完成"
