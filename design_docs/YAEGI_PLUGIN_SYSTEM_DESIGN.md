# Yaegi 第三方插件系統設計

## 概述

本文件說明如何使用 Yaegi Go 解釋器來實現第三方插件系統，讓 Channel 可以動態載入和使用第三方插件進行通知。

## 背景

參考 Traefik 的插件系統，我們希望：
1. 支援動態載入第三方 Go 插件
2. 插件可以實現自定義的通知邏輯
3. 無需重新編譯主程式即可添加新的 Channel 類型
4. 確保插件的安全性和隔離性

## 現有架構優勢

### 1. 完善的插件介面
```go
type ChannelTypeDefinition interface {
    GetName() string
    GetDisplayName() string
    GetDescription() string
    ValidateConfig(config map[string]interface{}) error
    GetConfigSchema() map[string]interface{}
    CreateMessageSender(timeout time.Duration) (interface{}, error)
}
```

### 2. 動態註冊機制
```go
type ChannelTypeRegistry interface {
    RegisterChannelType(channelType ChannelTypeDefinition) error
    GetChannelType(name string) (ChannelTypeDefinition, error)
    // ...
}
```

### 3. 配置驗證和 Schema
- 每個插件都有自己的配置驗證邏輯
- 支援 JSON Schema 定義

## Yaegi 整合設計

### 1. 插件載入器架構

```go
// PluginLoader 負責載入和管理 Yaegi 插件
type PluginLoader interface {
    LoadPlugin(pluginPath string) error
    LoadPluginFromSource(source string) error
    UnloadPlugin(pluginName string) error
    ListLoadedPlugins() []string
}

// YaegiPluginLoader 實現 PluginLoader
type YaegiPluginLoader struct {
    interpreter *interp.Interpreter
    plugins     map[string]*PluginInfo
    mutex       sync.RWMutex
}

type PluginInfo struct {
    Name        string
    Version     string
    Source      string
    ChannelType ChannelTypeDefinition
    LoadedAt    time.Time
}
```

### 2. 插件標準格式

```go
// 插件必須實現的標準介面
package main

import (
    "notification/internal/domain/shared"
)

// Plugin 插件主要介面
type Plugin interface {
    // GetInfo 返回插件資訊
    GetInfo() PluginInfo
    
    // GetChannelType 返回 Channel 類型定義
    GetChannelType() shared.ChannelTypeDefinition
    
    // Initialize 初始化插件
    Initialize(config map[string]interface{}) error
    
    // Cleanup 清理插件資源
    Cleanup() error
}

// 插件必須導出的函數
func NewPlugin() Plugin {
    return &MyCustomPlugin{}
}
```

### 3. 插件目錄結構

```
plugins/
├── discord/
│   ├── plugin.go          # 主插件檔案
│   ├── config.yaml        # 插件配置
│   └── README.md          # 說明文件
├── telegram/
│   ├── plugin.go
│   ├── config.yaml
│   └── README.md
└── custom-webhook/
    ├── plugin.go
    ├── config.yaml
    └── README.md
```

### 4. 插件配置格式

```yaml
# config.yaml
name: "discord"
version: "1.0.0"
description: "Discord webhook notifications"
author: "Your Name"
entry_point: "plugin.go"
dependencies:
  - "net/http"
  - "encoding/json"
permissions:
  - "network.http"
  - "config.read"
```

## 實現步驟

### 階段 1：基礎 Yaegi 整合

1. **添加 Yaegi 依賴**
```bash
go get github.com/traefik/yaegi
```

2. **實現插件載入器**
```go
// internal/infrastructure/plugins/yaegi_loader.go
type YaegiPluginLoader struct {
    interpreter *interp.Interpreter
    plugins     map[string]*PluginInfo
    mutex       sync.RWMutex
}

func NewYaegiPluginLoader() *YaegiPluginLoader {
    i := interp.New(interp.Options{})
    
    // 註冊必要的標準庫
    i.Use(stdlib.Symbols)
    
    // 註冊我們的介面
    i.Use(map[string]map[string]reflect.Value{
        "notification/internal/domain/shared": {
            "ChannelTypeDefinition": reflect.ValueOf((*shared.ChannelTypeDefinition)(nil)),
        },
    })
    
    return &YaegiPluginLoader{
        interpreter: i,
        plugins:     make(map[string]*PluginInfo),
    }
}
```

3. **實現插件載入邏輯**
```go
func (l *YaegiPluginLoader) LoadPlugin(pluginPath string) error {
    // 讀取插件原始碼
    source, err := os.ReadFile(pluginPath)
    if err != nil {
        return fmt.Errorf("failed to read plugin file: %w", err)
    }
    
    // 執行插件程式碼
    _, err = l.interpreter.Eval(string(source))
    if err != nil {
        return fmt.Errorf("failed to evaluate plugin: %w", err)
    }
    
    // 取得插件實例
    newPluginFunc, err := l.interpreter.Eval("NewPlugin")
    if err != nil {
        return fmt.Errorf("plugin must export NewPlugin function: %w", err)
    }
    
    // 呼叫 NewPlugin() 建立插件實例
    pluginValue := newPluginFunc.Call(nil)[0]
    plugin := pluginValue.Interface().(Plugin)
    
    // 註冊 Channel 類型
    channelType := plugin.GetChannelType()
    registry := shared.GetChannelTypeRegistry()
    return registry.RegisterChannelType(channelType)
}
```

### 階段 2：安全性和隔離

1. **限制插件權限**
```go
func (l *YaegiPluginLoader) setupSandbox() {
    // 限制檔案系統存取
    l.interpreter.Use(map[string]map[string]reflect.Value{
        "os": {
            "Open":   reflect.ValueOf(restrictedOpen),
            "Create": reflect.ValueOf(restrictedCreate),
        },
    })
}
```

2. **資源限制**
```go
type PluginContext struct {
    MaxMemory     int64
    MaxGoroutines int
    Timeout       time.Duration
}
```

### 階段 3：插件管理 API

1. **HTTP API 端點**
```go
// POST /api/v1/plugins/load
// GET /api/v1/plugins
// DELETE /api/v1/plugins/{name}
// GET /api/v1/plugins/{name}/status
```

2. **插件狀態監控**
```go
type PluginStatus struct {
    Name      string    `json:"name"`
    Status    string    `json:"status"` // loaded, error, unloaded
    LoadedAt  time.Time `json:"loadedAt"`
    Error     string    `json:"error,omitempty"`
    MemoryUse int64     `json:"memoryUse"`
}
```

## 範例插件實現

### Discord 插件範例

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "notification/internal/domain/shared"
)

type DiscordPlugin struct {
    config map[string]interface{}
}

func (p *DiscordPlugin) GetInfo() PluginInfo {
    return PluginInfo{
        Name:        "discord",
        Version:     "1.0.0",
        Description: "Discord webhook notifications",
        Author:      "Plugin Developer",
    }
}

func (p *DiscordPlugin) GetChannelType() shared.ChannelTypeDefinition {
    return &DiscordChannelType{}
}

func (p *DiscordPlugin) Initialize(config map[string]interface{}) error {
    p.config = config
    return nil
}

func (p *DiscordPlugin) Cleanup() error {
    return nil
}

// DiscordChannelType 實現 ChannelTypeDefinition
type DiscordChannelType struct{}

func (d *DiscordChannelType) GetName() string {
    return "discord"
}

func (d *DiscordChannelType) GetDisplayName() string {
    return "Discord"
}

func (d *DiscordChannelType) GetDescription() string {
    return "Send notifications to Discord via webhook"
}

func (d *DiscordChannelType) ValidateConfig(config map[string]interface{}) error {
    webhookURL, ok := config["webhook_url"].(string)
    if !ok || webhookURL == "" {
        return fmt.Errorf("webhook_url is required")
    }
    return nil
}

func (d *DiscordChannelType) GetConfigSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "webhook_url": map[string]interface{}{
                "type":        "string",
                "description": "Discord webhook URL",
                "format":      "uri",
            },
        },
        "required": []string{"webhook_url"},
    }
}

func (d *DiscordChannelType) CreateMessageSender(timeout time.Duration) (interface{}, error) {
    return &DiscordSender{timeout: timeout}, nil
}

type DiscordSender struct {
    timeout time.Duration
}

func (s *DiscordSender) Send(ctx context.Context, ch interface{}, content interface{}) error {
    // 實際的 Discord webhook 發送邏輯
    return nil
}

// 插件入口點
func NewPlugin() Plugin {
    return &DiscordPlugin{}
}
```

## 優勢與考量

### ✅ 優勢
1. **動態載入**：無需重新編譯即可添加新功能
2. **Go 原生**：插件使用 Go 語言，與主程式一致
3. **型別安全**：編譯時檢查介面相容性
4. **效能良好**：Yaegi 效能接近原生 Go
5. **安全隔離**：可以限制插件的系統存取權限

### ⚠️ 考量事項
1. **記憶體使用**：每個插件都會佔用額外記憶體
2. **除錯困難**：插件錯誤可能較難追蹤
3. **相依性管理**：需要小心處理插件的相依性
4. **版本相容性**：需要確保插件與主程式的介面相容

## 建議實施策略

### 第一階段：基礎實現
1. 實現基本的 Yaegi 插件載入器
2. 建立插件標準格式和範例
3. 實現簡單的插件管理 API

### 第二階段：安全強化
1. 添加插件沙盒機制
2. 實現資源限制和監控
3. 添加插件簽名驗證

### 第三階段：生態建設
1. 建立插件市場或倉庫
2. 提供插件開發工具和文件
3. 建立插件測試和認證流程

## 結論

使用 Yaegi 實現第三方插件系統是完全可行的，您現有的架構已經為此提供了良好的基礎。建議從簡單的實現開始，逐步添加安全性和管理功能。

這個方案可以讓您的通知系統具備極強的擴展性，使用者可以輕鬆添加自定義的通知渠道，而無需修改核心程式碼。