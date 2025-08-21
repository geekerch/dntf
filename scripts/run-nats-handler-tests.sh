#!/bin/bash

# NATS Handler 測試運行腳本
# 此腳本用於運行 internal/presentation/nats/handlers 的所有測試

set -e

echo "🧪 開始運行 NATS Handler 測試..."

# 設置測試環境變數
export GO_ENV=test
export LOG_LEVEL=error

# 檢查必要的依賴
echo "📦 檢查測試依賴..."

# 檢查 Go 是否安裝
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安裝，請先安裝 Go"
    exit 1
fi

# 檢查 NATS Server 是否可用（用於集成測試）
if ! command -v nats-server &> /dev/null; then
    echo "⚠️  NATS Server 未安裝，某些集成測試可能會失敗"
    echo "   可以通過以下命令安裝: go install github.com/nats-io/nats-server/v2@latest"
fi

# 進入項目根目錄
cd "$(dirname "$0")/.."

echo "📁 當前目錄: $(pwd)"

# 下載測試依賴
echo "📥 下載測試依賴..."
go mod download
go mod tidy

# 運行測試
echo "🏃 運行 NATS Handler 測試..."

# 設置測試標誌
TEST_FLAGS="-v -race -timeout=30s"
COVERAGE_FLAGS="-coverprofile=coverage.out -covermode=atomic"

# 測試目標
TEST_PACKAGE="./internal/presentation/nats/handlers"

echo "📋 測試配置:"
echo "  - 包: $TEST_PACKAGE"
echo "  - 標誌: $TEST_FLAGS"
echo "  - 覆蓋率: $COVERAGE_FLAGS"
echo ""

# 運行具體的測試
echo "🔍 運行 Channel Handler 測試..."
go test $TEST_FLAGS $TEST_PACKAGE -run "TestChannelNATSHandler.*" || echo "❌ Channel Handler 測試失敗"

echo ""
echo "🔍 運行 Template Handler 測試..."
go test $TEST_FLAGS $TEST_PACKAGE -run "TestTemplateNATSHandler.*" || echo "❌ Template Handler 測試失敗"

echo ""
echo "🔍 運行 Message Handler 測試..."
go test $TEST_FLAGS $TEST_PACKAGE -run "TestMessageNATSHandler.*" || echo "❌ Message Handler 測試失敗"

echo ""
echo "🔍 運行整合測試..."
go test $TEST_FLAGS $TEST_PACKAGE -run "TestChannelTemplateIntegration.*|TestChannelOldSystemSync.*|TestSMTPEmailSending.*" || echo "❌ 整合測試失敗"

echo ""
echo "📊 運行完整測試套件並生成覆蓋率報告..."
go test $TEST_FLAGS $COVERAGE_FLAGS $TEST_PACKAGE

# 生成覆蓋率報告
if [ -f coverage.out ]; then
    echo ""
    echo "📈 生成覆蓋率報告..."
    go tool cover -html=coverage.out -o coverage.html
    echo "✅ 覆蓋率報告已生成: coverage.html"
    
    # 顯示覆蓋率統計
    echo ""
    echo "📊 覆蓋率統計:"
    go tool cover -func=coverage.out | tail -1
fi

echo ""
echo "🎉 NATS Handler 測試完成！"

# 清理
echo "🧹 清理測試文件..."
rm -f coverage.out

echo "✅ 測試運行完成"