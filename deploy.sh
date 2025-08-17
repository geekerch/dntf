#!/bin/bash

# 部署腳本 - 將專案 build 成 Docker 並部署在本機

set -e

echo "🚀 開始部署 Notification API..."

# 檢查 Docker 是否安裝
if ! command -v docker &> /dev/null; then
    echo "❌ Docker 未安裝，請先安裝 Docker"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose 未安裝，請先安裝 Docker Compose"
    exit 1
fi

# 檢查 .env 文件是否存在
if [ ! -f ".env" ]; then
    echo "❌ .env 文件不存在，請先創建 .env 文件"
    exit 1
fi

# 檢查 NATS credentials 文件是否存在
if [ ! -f "cmd/server/edgesync_shadowagent.creds" ]; then
    echo "❌ NATS credentials 文件不存在: cmd/server/edgesync_shadowagent.creds"
    exit 1
fi

echo "📋 檢查配置文件..."
echo "✅ .env 文件存在"
echo "✅ NATS credentials 文件存在"

# 停止並移除現有容器（如果存在）
echo "🛑 停止現有容器..."
docker-compose -f docker-compose.deploy.yml down --remove-orphans || true

# 清理舊的 Docker 映像（可選）
echo "🧹 清理舊的 Docker 映像..."
docker image prune -f || true

# 構建新的 Docker 映像
echo "🔨 構建 Docker 映像..."
docker-compose -f docker-compose.deploy.yml build --no-cache

# 啟動容器
echo "🚀 啟動容器..."
docker-compose -f docker-compose.deploy.yml up -d

# 等待容器啟動
echo "⏳ 等待容器啟動..."
sleep 10

# 檢查容器狀態
echo "📊 檢查容器狀態..."
docker-compose -f docker-compose.deploy.yml ps

# 檢查健康狀態
echo "🏥 檢查應用健康狀態..."
for i in {1..30}; do
    if curl -f http://localhost:8080/health &>/dev/null; then
        echo "✅ 應用啟動成功！"
        echo "🌐 API 可在以下地址訪問: http://localhost:8080"
        echo "📚 API 文檔: http://localhost:8080/swagger/index.html"
        break
    else
        echo "⏳ 等待應用啟動... ($i/30)"
        sleep 2
    fi
    
    if [ $i -eq 30 ]; then
        echo "❌ 應用啟動超時，請檢查日誌"
        echo "📋 查看日誌命令: docker-compose -f docker-compose.deploy.yml logs -f"
        exit 1
    fi
done

echo ""
echo "🎉 部署完成！"
echo ""
echo "📋 常用命令:"
echo "  查看日誌: docker-compose -f docker-compose.deploy.yml logs -f"
echo "  停止服務: docker-compose -f docker-compose.deploy.yml down"
echo "  重啟服務: docker-compose -f docker-compose.deploy.yml restart"
echo "  查看狀態: docker-compose -f docker-compose.deploy.yml ps"
echo ""