#!/bin/bash

# 數據庫初始化腳本

set -e

echo "🗄️ 初始化數據庫..."

# 數據庫配置
DB_HOST="localhost"
DB_PORT="5432"
DB_USER="postgres"
DB_NAME="notification_api"

# 檢查 PostgreSQL 是否可用
if ! pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER > /dev/null 2>&1; then
    echo "❌ 無法連接到 PostgreSQL"
    echo "請確保 PostgreSQL 服務正在運行並且可以連接"
    exit 1
fi

echo "✅ PostgreSQL 連接正常"

# 創建數據庫（如果不存在）
echo "📝 創建數據庫 $DB_NAME（如果不存在）..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -tc "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'" | grep -q 1 || \
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -c "CREATE DATABASE $DB_NAME;"

echo "✅ 數據庫 $DB_NAME 已準備就緒"

# 顯示數據庫信息
echo "📊 數據庫信息："
echo "   - 主機: $DB_HOST:$DB_PORT"
echo "   - 數據庫: $DB_NAME"
echo "   - 用戶: $DB_USER"

echo "🎉 數據庫初始化完成！"