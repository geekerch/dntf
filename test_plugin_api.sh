#!/bin/bash

# Test script for plugin API

echo "🧪 Testing Plugin API"
echo "====================="

# Start server in background (if not already running)
# go run cmd/server/main.go &
# SERVER_PID=$!
# sleep 5

# Test plugin source code
PLUGIN_SOURCE='package main

import (
	"fmt"
	"time"
)

type PluginInfo struct {
	Name        string
	Version     string
	Description string
	Author      string
	LoadedAt    time.Time
}

type Plugin interface {
	GetInfo() PluginInfo
	GetChannelType() ChannelTypeDefinition
	Initialize(config map[string]interface{}) error
	Cleanup() error
}

type ChannelTypeDefinition interface {
	GetName() string
	GetDisplayName() string
	GetDescription() string
	ValidateConfig(config map[string]interface{}) error
	GetConfigSchema() map[string]interface{}
	CreateMessageSender(timeout time.Duration) (interface{}, error)
}

type TestPlugin struct{}

func (p *TestPlugin) GetInfo() PluginInfo {
	return PluginInfo{
		Name:        "test-api",
		Version:     "1.0.0",
		Description: "Test plugin via API",
		Author:      "System",
		LoadedAt:    time.Now(),
	}
}

func (p *TestPlugin) GetChannelType() ChannelTypeDefinition {
	return &TestChannelType{}
}

func (p *TestPlugin) Initialize(config map[string]interface{}) error {
	fmt.Println("Test plugin initialized via API")
	return nil
}

func (p *TestPlugin) Cleanup() error {
	fmt.Println("Test plugin cleaned up")
	return nil
}

type TestChannelType struct{}

func (t *TestChannelType) GetName() string {
	return "test-api"
}

func (t *TestChannelType) GetDisplayName() string {
	return "Test API Channel"
}

func (t *TestChannelType) GetDescription() string {
	return "Test channel loaded via API"
}

func (t *TestChannelType) ValidateConfig(config map[string]interface{}) error {
	return nil
}

func (t *TestChannelType) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url": map[string]interface{}{
				"type": "string",
				"description": "Test URL",
			},
		},
	}
}

func (t *TestChannelType) CreateMessageSender(timeout time.Duration) (interface{}, error) {
	return &TestSender{}, nil
}

type TestSender struct{}

func NewPlugin() Plugin {
	return &TestPlugin{}
}'

echo "📤 Test 1: Loading plugin via API..."
curl -X POST http://localhost:8080/api/v1/plugins/load \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"test-api\", \"source\": $(echo "$PLUGIN_SOURCE" | jq -Rs .)}" \
  2>/dev/null | jq '.'

echo ""
echo "📋 Test 2: Listing all plugins..."
curl -X GET http://localhost:8080/api/v1/plugins \
  2>/dev/null | jq '.'

echo ""
echo "🔍 Test 3: Getting specific plugin status..."
curl -X GET http://localhost:8080/api/v1/plugins/test-api \
  2>/dev/null | jq '.'

echo ""
echo "🗑️ Test 4: Unloading plugin..."
curl -X DELETE http://localhost:8080/api/v1/plugins/test-api \
  2>/dev/null | jq '.'

echo ""
echo "✅ Plugin API testing completed!"

# Clean up
# kill $SERVER_PID 2>/dev/null