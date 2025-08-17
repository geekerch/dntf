#!/bin/bash

# 部署腳本 - 將 Notification API 部署到 Docker

set -e

echo "🚀 開始部署 Notification API..."

# 檢查 Docker 是否運行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker 未運行，請先啟動 Docker"
    exit 1
fi

# 檢查必要的服務是否運行
echo "🔍 檢查依賴服務..."

# 檢查 PostgreSQL
if ! nc -z localhost 5432 2>/dev/null; then
    echo "❌ PostgreSQL (端口 5432) 未運行"
    echo "請確保 PostgreSQL 服務已啟動"
    exit 1
fi
echo "✅ PostgreSQL 服務正常"

# 檢查 NATS
if ! nc -z localhost 4222 2>/dev/null; then
    echo "❌ NATS (端口 4222) 未運行"
    echo "請確保 NATS 服務已啟動"
    exit 1
fi
echo "✅ NATS 服務正常"

# 停止並移除現有容器（如果存在）
echo "🛑 停止現有容器..."
docker-compose -f docker-compose.standalone.yml down 2>/dev/null || true

# 構建新的映像
echo "🔨 構建 Docker 映像..."
docker-compose -f docker-compose.standalone.yml build --no-cache

# 啟動服務
echo "🚀 啟動服務..."
docker-compose -f docker-compose.standalone.yml up -d

# 等待服務啟動
echo "⏳ 等待服務啟動..."
sleep 10

# 檢查服務狀態
echo "🔍 檢查服務狀態..."
if docker-compose -f docker-compose.standalone.yml ps | grep -q "Up"; then
    echo "✅ 服務啟動成功！"
    
    # 檢查健康狀態
    echo "🏥 檢查健康狀態..."
    for i in {1..30}; do
        if curl -s http://localhost:8080/health > /dev/null; then
            echo "✅ 健康檢查通過！"
            break
        fi
        echo "⏳ 等待健康檢查... ($i/30)"
        sleep 2
    done
    
    echo ""
    echo "🎉 部署完成！"
    echo "📊 服務信息："
    echo "   - API 端點: http://localhost:8080"
    echo "   - Swagger 文檔: http://localhost:8080/swagger/index.html"
    echo "   - 健康檢查: http://localhost:8080/health"
    echo ""
    echo "📋 管理命令："
    echo "   - 查看日誌: docker-compose -f docker-compose.standalone.yml logs -f"
    echo "   - 停止服務: docker-compose -f docker-compose.standalone.yml down"
    echo "   - 重啟服務: docker-compose -f docker-compose.standalone.yml restart"
    
else
    echo "❌ 服務啟動失敗"
    echo "查看日誌："
    docker-compose -f docker-compose.standalone.yml logs
    exit 1
fi