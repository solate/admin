#!/bin/bash

# ========================================
# 开发环境数据库重置脚本
# 用途：删除并重建数据库，重新执行所有初始化脚本
# 警告：此脚本会删除所有数据，仅用于开发环境！
# ========================================

set -e  # 遇到错误立即退出

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 配置变量（可以通过环境变量覆盖）
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_NAME="${DB_NAME:-admin_db}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"

# 脚本目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"
SQL_FILE="$SCRIPT_DIR/dev_schema.sql"

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}开发环境数据库重置${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""
echo "数据库配置："
echo "  主机: $DB_HOST"
echo "  端口: $DB_PORT"
echo "  用户: $DB_USER"
echo "  数据库: $DB_NAME"
echo ""

# 确认操作
echo -e "${RED}警告: 此操作将删除数据库 '$DB_NAME' 及其所有数据！${NC}"
read -p "确认继续？(y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "操作已取消"
    exit 0
fi

# 设置 PGPASSWORD 环境变量以避免密码提示
export PGPASSWORD="$DB_PASSWORD"

echo ""
echo -e "${GREEN}步骤 1: 断开现有数据库连接${NC}"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c "
SELECT pg_terminate_backend(pg_stat_activity.pid)
FROM pg_stat_activity
WHERE pg_stat_activity.datname = '$DB_NAME'
  AND pid <> pg_backend_pid();
" 2>/dev/null || true

echo -e "${GREEN}步骤 2: 删除现有数据库${NC}"
dropdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$DB_NAME" --if-exists
echo "  ✓ 数据库已删除"

echo -e "${GREEN}步骤 3: 创建新数据库${NC}"
createdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$DB_NAME"
echo "  ✓ 数据库已创建"

echo -e "${GREEN}步骤 4: 执行初始化 SQL 脚本${NC}"
if [ ! -f "$SQL_FILE" ]; then
    echo -e "${RED}错误: SQL 文件不存在: $SQL_FILE${NC}"
    exit 1
fi

psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$SQL_FILE"
echo "  ✓ 数据库表结构已创建"

echo -e "${GREEN}步骤 5: 验证表创建${NC}"
TABLE_COUNT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
SELECT COUNT(*) FROM information_schema.tables 
WHERE table_schema = 'public' AND table_type = 'BASE TABLE';
")
echo "  ✓ 已创建 $TABLE_COUNT 张表"

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}数据库重置完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "已创建的表："
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
ORDER BY table_name;
"


