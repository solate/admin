.PHONY: build run dev clean test lint fmt migrate-up migrate-down migrate-create migrate-reset gen-db reset

# ── 数据库连接(可被环境变量覆盖,默认对齐 config/config.yaml) ──
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= postgres
DB_PASSWORD ?= postgres
DB_NAME ?= admin_dev
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
MIGRATION_DIR := ./migrations

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

# dev: 迁移 → 生成 DAL → 启动(对齐 migration-data-convention 约定的一键流程)
dev: migrate-up gen-db
	go run ./cmd/server

clean:
	rm -rf bin/ tmp/

test:
	go test ./... -v -count=1

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

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
	@if command -v migrate > /dev/null; then \
		echo "⬇️  回滚一步..."; \
		migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" down 1; \
		echo "✅ 回滚完成"; \
	else \
		echo "❌ golang-migrate 未安装: brew install golang-migrate"; \
		exit 1; \
	fi

## migrate-create: 新建迁移文件对 (make migrate-create NAME=ddl_add_xxx)
migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "❌ 请指定名称: make migrate-create NAME=ddl_add_user_table"; \
		exit 1; \
	fi
	@if command -v migrate > /dev/null; then \
		migrate create -ext sql -dir $(MIGRATION_DIR) -seq $(NAME); \
	else \
		echo "❌ golang-migrate 未安装: brew install golang-migrate"; \
		exit 1; \
	fi

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
	@echo "🔧 从数据库生成 DAL (internal/dal/model + internal/dal/query)..."
	go run ./scripts/gen-from-db/
	@echo "✅ 生成完成"

## reset: 一键重建本地(重置迁移 + 重新生成 DAL)
reset: migrate-reset gen-db
