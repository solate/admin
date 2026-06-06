#!/bin/bash

# 数据库初始化脚本

set -e

# 配置
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-admin_db}"

echo "🔧 Initializing database: $DB_NAME"

# 创建数据库（如果不存在）
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -tc "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'" | grep -q 1 || \
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -c "CREATE DATABASE $DB_NAME"

echo "✅ Database $DB_NAME is ready"

# 运行迁移
echo "🚀 Running migrations..."

# 使用 golang-migrate 运行迁移
if command -v migrate &> /dev/null; then
    migrate -path ./migrations -database "postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable" up
    echo "✅ Migrations completed"
else
    echo "⚠️  golang-migrate not found. Install it with:"
    echo "   brew install golang-migrate  # macOS"
    echo "   Or manually run migrations/*.up.sql"
    
    # 手动运行迁移文件
    echo "🔄 Running migrations manually..."
    for file in migrations/*.up.sql; do
        echo "  Running: $file"
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$file"
    done
    echo "✅ Migrations completed manually"
fi

echo "🎉 Database initialization completed!"

