#!/bin/bash

# Notification API 管理腳本

COMPOSE_FILE="docker-compose.deploy.yml"

show_help() {
    echo "Notification API 管理腳本"
    echo ""
    echo "用法: $0 [命令]"
    echo ""
    echo "命令:"
    echo "  start     啟動服務"
    echo "  stop      停止服務"
    echo "  restart   重啟服務"
    echo "  status    查看狀態"
    echo "  logs      查看日誌"
    echo "  deploy    重新部署"
    echo "  health    健康檢查"
    echo "  shell     進入容器"
    echo "  clean     清理資源"
    echo "  help      顯示幫助"
    echo ""
}

case "$1" in
    start)
        echo "🚀 啟動 Notification API..."
        docker-compose -f $COMPOSE_FILE up -d
        ;;
    stop)
        echo "🛑 停止 Notification API..."
        docker-compose -f $COMPOSE_FILE down
        ;;
    restart)
        echo "🔄 重啟 Notification API..."
        docker-compose -f $COMPOSE_FILE restart
        ;;
    status)
        echo "📊 查看服務狀態..."
        docker-compose -f $COMPOSE_FILE ps
        echo ""
        docker stats notification-api --no-stream
        ;;
    logs)
        echo "📋 查看服務日誌..."
        docker-compose -f $COMPOSE_FILE logs -f
        ;;
    deploy)
        echo "🔨 重新部署..."
        docker-compose -f $COMPOSE_FILE down
        docker-compose -f $COMPOSE_FILE up -d --build
        ;;
    health)
        echo "🏥 健康檢查..."
        echo "容器狀態:"
        docker-compose -f $COMPOSE_FILE ps
        echo ""
        echo "API測試:"
        curl -s http://localhost:8080/api/v1/channels | head -100 || echo "API連接失敗"
        ;;
    shell)
        echo "🐚 進入容器..."
        docker exec -it notification-api sh
        ;;
    clean)
        echo "🧹 清理資源..."
        docker-compose -f $COMPOSE_FILE down --remove-orphans
        docker image prune -f
        docker volume prune -f
        ;;
    help|*)
        show_help
        ;;
esac