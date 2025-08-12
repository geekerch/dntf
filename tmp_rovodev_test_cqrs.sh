#!/bin/bash

# 測試 CQRS 功能的腳本

BASE_URL="http://localhost:8080"
API_KEY="dev-key-123"

echo "=== 測試 CQRS 功能 ==="
echo

# 測試健康檢查
echo "1. 測試健康檢查..."
curl -s -w "Status: %{http_code}\n" "$BASE_URL/health" | head -2
echo

# 測試傳統 API v1 (非 CQRS)
echo "2. 測試傳統 API v1 - 建立通道..."
RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -X POST "$BASE_URL/api/v1/channels" \
  -d '{
    "channelName": "traditional-email-channel",
    "description": "Traditional API email channel",
    "enabled": true,
    "channelType": "email",
    "commonSettings": {
      "timeout": 30,
      "retryAttempts": 3,
      "retryDelay": 5
    },
    "config": {
      "smtpHost": "smtp.example.com",
      "smtpPort": 587
    },
    "recipients": [
      {
        "name": "Test User",
        "email": "test@example.com",
        "type": "email"
      }
    ],
    "tags": ["traditional", "test"]
  }')

HTTP_STATUS=$(echo $RESPONSE | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
BODY=$(echo $RESPONSE | sed -e 's/HTTPSTATUS:.*//g')

echo "Status: $HTTP_STATUS"
if [ "$HTTP_STATUS" = "201" ]; then
  echo "✓ Traditional API 建立通道成功"
  TRADITIONAL_CHANNEL_ID=$(echo $BODY | jq -r '.channelId // empty')
  echo "Channel ID: $TRADITIONAL_CHANNEL_ID"
else
  echo "✗ Traditional API 建立通道失敗"
  echo "Response: $BODY"
fi
echo

# 測試 CQRS API v2
echo "3. 測試 CQRS API v2 - 建立通道..."
RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -X POST "$BASE_URL/api/v2/channels" \
  -d '{
    "channelName": "cqrs-email-channel",
    "description": "CQRS API email channel",
    "enabled": true,
    "channelType": "email",
    "commonSettings": {
      "timeout": 30,
      "retryAttempts": 3,
      "retryDelay": 5
    },
    "config": {
      "smtpHost": "smtp.example.com",
      "smtpPort": 587
    },
    "recipients": [
      {
        "name": "CQRS User",
        "email": "cqrs@example.com",
        "type": "email"
      }
    ],
    "tags": ["cqrs", "test"]
  }')

HTTP_STATUS=$(echo $RESPONSE | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
BODY=$(echo $RESPONSE | sed -e 's/HTTPSTATUS:.*//g')

echo "Status: $HTTP_STATUS"
if [ "$HTTP_STATUS" = "201" ]; then
  echo "✓ CQRS API 建立通道成功"
  CQRS_CHANNEL_ID=$(echo $BODY | jq -r '.channelId // empty')
  echo "Channel ID: $CQRS_CHANNEL_ID"
  
  # 檢查 CQRS 特有的標頭
  COMMAND_ID=$(curl -s -I -H "X-API-Key: $API_KEY" -H "Content-Type: application/json" \
    -X POST "$BASE_URL/api/v2/channels" \
    -d '{"channelName":"test","channelType":"email","commonSettings":{"timeout":30,"retryAttempts":3,"retryDelay":5},"config":{},"recipients":[],"tags":[]}' \
    | grep -i "x-command-id" | cut -d' ' -f2 | tr -d '\r')
  
  if [ ! -z "$COMMAND_ID" ]; then
    echo "✓ CQRS Command ID 追蹤: $COMMAND_ID"
  fi
else
  echo "✗ CQRS API 建立通道失敗"
  echo "Response: $BODY"
fi
echo

# 測試 CQRS 查詢功能
if [ ! -z "$CQRS_CHANNEL_ID" ]; then
  echo "4. 測試 CQRS 查詢功能..."
  
  # 測試取得單一通道
  echo "4.1 取得單一通道..."
  RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" \
    -H "X-API-Key: $API_KEY" \
    "$BASE_URL/api/v2/channels/$CQRS_CHANNEL_ID")
  
  HTTP_STATUS=$(echo $RESPONSE | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
  
  if [ "$HTTP_STATUS" = "200" ]; then
    echo "✓ CQRS 查詢單一通道成功"
    
    # 檢查查詢 ID 標頭
    QUERY_ID=$(curl -s -I -H "X-API-Key: $API_KEY" \
      "$BASE_URL/api/v2/channels/$CQRS_CHANNEL_ID" \
      | grep -i "x-query-id" | cut -d' ' -f2 | tr -d '\r')
    
    if [ ! -z "$QUERY_ID" ]; then
      echo "✓ CQRS Query ID 追蹤: $QUERY_ID"
    fi
  else
    echo "✗ CQRS 查詢單一通道失敗 (Status: $HTTP_STATUS)"
  fi
  
  # 測試查詢通道列表
  echo "4.2 查詢通道列表..."
  RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" \
    -H "X-API-Key: $API_KEY" \
    "$BASE_URL/api/v2/channels?channelType=email&maxResultCount=10")
  
  HTTP_STATUS=$(echo $RESPONSE | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
  
  if [ "$HTTP_STATUS" = "200" ]; then
    echo "✓ CQRS 查詢通道列表成功"
  else
    echo "✗ CQRS 查詢通道列表失敗 (Status: $HTTP_STATUS)"
  fi
  echo
fi

# 比較傳統 API 和 CQRS API 的回應時間
echo "5. 效能比較測試..."

if [ ! -z "$TRADITIONAL_CHANNEL_ID" ]; then
  echo "5.1 傳統 API 查詢效能..."
  time curl -s -H "X-API-Key: $API_KEY" \
    "$BASE_URL/api/v1/channels/$TRADITIONAL_CHANNEL_ID" > /dev/null
fi

if [ ! -z "$CQRS_CHANNEL_ID" ]; then
  echo "5.2 CQRS API 查詢效能..."
  time curl -s -H "X-API-Key: $API_KEY" \
    "$BASE_URL/api/v2/channels/$CQRS_CHANNEL_ID" > /dev/null
fi
echo

# 測試 CQRS 進階查詢功能
echo "6. 測試 CQRS 進階查詢功能..."

echo "6.1 分頁查詢..."
curl -s -w "Status: %{http_code}\n" \
  -H "X-API-Key: $API_KEY" \
  "$BASE_URL/api/v2/channels?skipCount=0&maxResultCount=5" | head -1

echo "6.2 排序查詢..."
curl -s -w "Status: %{http_code}\n" \
  -H "X-API-Key: $API_KEY" \
  "$BASE_URL/api/v2/channels?sortField=createdAt&sortOrder=desc" | head -1

echo "6.3 篩選查詢..."
curl -s -w "Status: %{http_code}\n" \
  -H "X-API-Key: $API_KEY" \
  "$BASE_URL/api/v2/channels?channelType=email&enabled=true" | head -1
echo

echo "=== CQRS 測試完成 ==="
echo
echo "測試總結："
echo "- 傳統 API (v1): /api/v1/channels"
echo "- CQRS API (v2): /api/v2/channels"
echo "- CQRS 提供命令/查詢 ID 追蹤"
echo "- CQRS 支援進階查詢功能"
echo "- 兩套 API 可以並存使用"