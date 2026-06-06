#!/bin/bash

# ========================================
# 开发环境数据库重置脚本
# 用途：删除并重建数据库，重新执行所有初始化脚本
# 警告：此脚本会删除所有数据，仅用于开发环境！
# ========================================

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 用于存储失败的信息
FAILED_TABLES=()
SQL_ERRORS=()
HAS_ERROR=false

# 配置变量（可以通过环境变量覆盖）
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_NAME="${DB_NAME:-admin}"
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

# 先提取 SQL 文件中所有期望创建的表名
# 匹配格式：CREATE TABLE table_name ( 或 CREATE UNLOGGED TABLE table_name (
EXPECTED_TABLES=$(grep -iE "^[[:space:]]*CREATE[[:space:]]+(UNLOGGED[[:space:]]+)?TABLE[[:space:]]+[a-zA-Z_][a-zA-Z0-9_]*[[:space:]]*\(" "$SQL_FILE" | sed -E 's/^[[:space:]]*CREATE[[:space:]]+(UNLOGGED[[:space:]]+)?TABLE[[:space:]]+([a-zA-Z_][a-zA-Z0-9_]*)[[:space:]]*\(.*/\2/i' | tr '[:upper:]' '[:lower:]' | sort -u)

# 执行 SQL 并捕获错误输出
SQL_OUTPUT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$SQL_FILE" 2>&1)
SQL_EXIT_CODE=$?

if [ $SQL_EXIT_CODE -ne 0 ]; then
    HAS_ERROR=true
    echo -e "${RED}  ✗ SQL 执行过程中出现错误${NC}"
    # 收集错误信息（包含表名）
    while IFS= read -r line; do
        if [[ "$line" =~ ERROR: ]]; then
            SQL_ERRORS+=("$line")
            # 尝试从错误信息中提取表名
            TABLE_NAME=$(echo "$line" | sed -E 's/.*relation\s+"?([a-zA-Z_][a-zA-Z0-9_]*)"?.*/\1/' | head -1)
            if [[ -n "$TABLE_NAME" && "$TABLE_NAME" != "$line" ]]; then
                FAILED_TABLES+=("$TABLE_NAME")
            fi
        fi
    done <<< "$SQL_OUTPUT"
else
    echo "  ✓ 数据库表结构已创建"
fi

echo -e "${GREEN}步骤 5: 验证表创建${NC}"
TABLE_COUNT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
SELECT COUNT(*) FROM information_schema.tables
WHERE table_schema = 'public' AND table_type = 'BASE TABLE';
")
echo "  ✓ 已创建 $TABLE_COUNT 张表"

# 对比期望创建的表和实际创建的表
if [ -n "$EXPECTED_TABLES" ]; then
    ACTUAL_TABLES=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
SELECT table_name FROM information_schema.tables
WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
ORDER BY table_name;" | tr -d ' ')

    # 找出缺失的表
    while IFS= read -r expected_table; do
        if [ -n "$expected_table" ]; then
            if ! echo "$ACTUAL_TABLES" | grep -q "^${expected_table}$"; then
                FAILED_TABLES+=("$expected_table")
                HAS_ERROR=true
            fi
        fi
    done <<< "$EXPECTED_TABLES"
fi

echo ""
echo -e "${GREEN}========================================${NC}"

# 如果有错误，显示失败信息
if [ "$HAS_ERROR" = true ]; then
    echo -e "${RED}数据库重置完成，但有错误！${NC}"
    echo -e "${RED}========================================${NC}"
    echo ""

    # 显示失败的表
    if [ ${#FAILED_TABLES[@]} -gt 0 ]; then
        echo -e "${RED}失败的表：${NC}"
        for table in "${FAILED_TABLES[@]}"; do
            echo -e "  ${RED}✗${NC} $table"
        done
        echo ""
    fi

    # 显示 SQL 错误信息
    if [ ${#SQL_ERRORS[@]} -gt 0 ]; then
        echo -e "${YELLOW}SQL 错误详情：${NC}"
        for error in "${SQL_ERRORS[@]}"; do
            echo -e "  ${YELLOW}!${NC} $error"
        done
        echo ""
    fi

    exit 1
else
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
fi


