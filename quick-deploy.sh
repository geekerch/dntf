#!/bin/bash

# 快速部署腳本

echo "🚀 快速部署 Notification API..."

# 停止現有容器
docker-compose -f docker-compose.deploy.yml down 2>/dev/null || true

# 構建並啟動
docker-compose -f docker-compose.deploy.yml up -d --build

echo "✅ 部署完成！"
echo "🌐 API: http://localhost:8080"
echo "📋 查看日誌: docker-compose -f docker-compose.deploy.yml logs -f"