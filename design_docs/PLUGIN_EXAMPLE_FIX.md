# 插件範例修正

## 問題描述

`plugins/example/plugin.go` 中的範例插件存在以下問題：

1. **導入內部套件問題**：插件嘗試導入 `internal/domain/shared` 和 `internal/infrastructure/plugins` 套件，但 Go 的 `internal` 套件有訪問限制，外部插件無法使用。

2. **介面定義缺失**：插件使用了 `PluginInfo` 和 `Plugin` 介面，但沒有正確的導入路徑。

## 解決方案

### 1. 創建公開的插件 API

在 `pkg/plugins/interfaces.go` 中創建了公開的插件介面定義：

```go
// PluginInfo contains plugin metadata
type PluginInfo struct {
    Name        string    `json:"name"`
    Version     string    `json:"version"`
    Description string    `json:"description"`
    Author      string    `json:"author"`
    LoadedAt    time.Time `json:"loadedAt"`
}

// Plugin represents a loaded plugin instance
type Plugin interface {
    GetInfo() PluginInfo
    GetChannelType() ChannelTypeDefinition
    Initialize(config map[string]interface{}) error
    Cleanup() error
}

// ChannelTypeDefinition defines the interface for channel types
type ChannelTypeDefinition interface {
    GetName() string
    GetDisplayName() string
    GetDescription() string
    ValidateConfig(config map[string]interface{}) error
    GetConfigSchema() map[string]interface{}
    CreateMessageSender(timeout time.Duration) (interface{}, error)
}
```

### 2. 修正範例插件

修正了 `plugins/example/plugin.go`：

- 移除了對 `internal` 套件的導入
- 改為使用 `notification/pkg/plugins` 套件
- 修正了所有介面類型的引用

### 3. 編譯驗證

插件現在可以正確編譯：

```bash
cd plugins/example
go mod init example-plugin
go mod edit -replace notification=../../
go mod tidy
go build -buildmode=plugin -o plugin.so plugin.go
```

## 影響

1. **插件開發者**：現在可以使用公開的 API 開發插件，不再受到 `internal` 套件限制
2. **系統架構**：保持了清晰的 API 邊界，內部實作與公開介面分離
3. **向後相容性**：內部系統仍然可以使用原有的 `internal` 套件介面

## 後續工作

1. 需要更新內部插件載入器，使其能夠處理使用公開 API 的插件
2. 可能需要創建適配器來橋接公開 API 和內部實作
3. 更新插件開發文件，說明如何使用新的公開 API

## 測試

範例插件現在可以正確編譯，沒有編譯錯誤。