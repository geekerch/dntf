#!/bin/bash

# Test script for custom-email plugin

echo "🧪 Testing Custom Email Plugin"
echo "================================"

# Test 1: Load plugin from file
echo "📁 Test 1: Loading plugin from file..."
curl -X POST http://localhost:8080/api/v1/plugins/load-file \
  -H "Content-Type: application/json" \
  -d '{"file_path": "./plugins/custom-email/plugin.go"}' \
  2>/dev/null | jq '.'

echo ""

# Test 2: List all plugins
echo "📋 Test 2: Listing all plugins..."
curl -X GET http://localhost:8080/api/v1/plugins \
  2>/dev/null | jq '.'

echo ""

# Test 3: Get specific plugin status
echo "🔍 Test 3: Getting custom-email plugin status..."
curl -X GET http://localhost:8080/api/v1/plugins/custom-email \
  2>/dev/null | jq '.'

echo ""

# Test 4: Test plugin configuration validation
echo "⚙️ Test 4: Testing configuration validation..."
echo "This would require creating a channel with custom-email type"
echo "You can test this through the channel creation API:"
echo ""
echo "curl -X POST http://localhost:8080/api/v1/channels \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{"
echo "    \"channelName\": \"Test Custom Email\","
echo "    \"description\": \"Test custom email channel\","
echo "    \"enabled\": true,"
echo "    \"channelType\": \"custom-email\","
echo "    \"commonSettings\": {"
echo "      \"timeout\": 30000,"
echo "      \"retryAttempts\": 3,"
echo "      \"retryDelay\": 1000"
echo "    },"
echo "    \"config\": {"
echo "      \"smtp_host\": \"smtp.gmail.com\","
echo "      \"smtp_port\": 587,"
echo "      \"username\": \"test@gmail.com\","
echo "      \"password\": \"your-password\","
echo "      \"from_email\": \"test@gmail.com\","
echo "      \"from_name\": \"Test Sender\","
echo "      \"use_tls\": true"
echo "    },"
echo "    \"recipients\": ["
echo "      {"
echo "        \"name\": \"Test User\","
echo "        \"target\": \"user@example.com\","
echo "        \"type\": \"email\""
echo "      }"
echo "    ]"
echo "  }'"

echo ""
echo "✅ Plugin testing script completed!"
echo "Note: Make sure the server is running before executing this script"