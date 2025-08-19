# 可擴充 Channel Type 實作設計文件

## 概述

本文件描述如何實作一個可擴充的 Channel Type 機制，讓開發者可以輕鬆地新增新的通訊類型（如 email、SMS、Slack、Discord 等），而不需要修改核心程式碼。

## 設計目標

1. **可擴充性**: 新的 channel type 可以透過實作介面來擴充
2. **插件化**: 支援動態註冊新的 channel type
3. **向後相容**: 不影響現有的 email、SMS、Slack 實作
4. **類型安全**: 保持強型別檢查
5. **配置驗證**: 每個 channel type 可以定義自己的配置驗證規則

## 核心設計

### 1. Channel Type Registry

建立一個中央註冊表來管理所有可用的 channel types：

```go
// ChannelTypeRegistry 管理所有註冊的 channel types
type ChannelTypeRegistry interface {
    // RegisterChannelType 註冊新的 channel type
    RegisterChannelType(channelType ChannelTypeDefinition) error
    
    // GetChannelType 取得指定的 channel type 定義
    GetChannelType(name string) (ChannelTypeDefinition, error)
    
    // GetAllChannelTypes 取得所有註冊的 channel types
    GetAllChannelTypes() []ChannelTypeDefinition
    
    // IsValidChannelType 檢查 channel type 是否有效
    IsValidChannelType(name string) bool
}
```

### 2. Channel Type Definition

定義 channel type 的介面：

```go
// ChannelTypeDefinition 定義一個 channel type 的所有屬性
type ChannelTypeDefinition interface {
    // GetName 取得 channel type 名稱
    GetName() string
    
    // GetDisplayName 取得顯示名稱
    GetDisplayName() string
    
    // GetDescription 取得描述
    GetDescription() string
    
    // ValidateConfig 驗證 channel 配置
    ValidateConfig(config map[string]interface{}) error
    
    // GetConfigSchema 取得配置結構描述
    GetConfigSchema() map[string]interface{}
    
    // CreateMessageSender 建立對應的 message sender
    CreateMessageSender(timeout time.Duration) (MessageSender, error)
}
```

### 3. 修改 ChannelType 為動態類型

將原本的硬編碼 `ChannelType` 改為動態類型：

```go
// ChannelType 代表通訊管道類型
type ChannelType struct {
    name string
}

// NewChannelType 建立新的 channel type
func NewChannelType(name string) (ChannelType, error) {
    if !GetChannelTypeRegistry().IsValidChannelType(name) {
        return ChannelType{}, fmt.Errorf("invalid channel type: %s", name)
    }
    return ChannelType{name: name}, nil
}

// String 回傳字串表示
func (ct ChannelType) String() string {
    return ct.name
}

// IsValid 檢查是否為有效的 channel type
func (ct ChannelType) IsValid() bool {
    return GetChannelTypeRegistry().IsValidChannelType(ct.name)
}
```

### 4. 預定義 Channel Types

為現有的 channel types 建立定義：

```go
// EmailChannelType email channel type 定義
type EmailChannelType struct{}

func (e EmailChannelType) GetName() string { return "email" }
func (e EmailChannelType) GetDisplayName() string { return "Email" }
func (e EmailChannelType) GetDescription() string { return "Send notifications via email" }

func (e EmailChannelType) ValidateConfig(config map[string]interface{}) error {
    // 驗證 email 相關配置
    // 如：smtp_host, smtp_port, username, password 等
}

func (e EmailChannelType) GetConfigSchema() map[string]interface{} {
    // 回傳 email 配置的 JSON schema
}

func (e EmailChannelType) CreateMessageSender(timeout time.Duration) (MessageSender, error) {
    return NewEmailService(timeout), nil
}
```

### 5. 自動註冊機制

建立初始化函數來自動註冊預設的 channel types：

```go
// RegisterDefaultChannelTypes 註冊預設的 channel types
func RegisterDefaultChannelTypes() {
    registry := GetChannelTypeRegistry()
    
    // 註冊預設的 channel types
    registry.RegisterChannelType(&EmailChannelType{})
    registry.RegisterChannelType(&SlackChannelType{})
    registry.RegisterChannelType(&SMSChannelType{})
}
```

## 實作步驟

### 階段 1: 建立核心介面和註冊表
1. 建立 `ChannelTypeRegistry` 介面和實作
2. 建立 `ChannelTypeDefinition` 介面
3. 修改 `ChannelType` 為動態類型

### 階段 2: 重構現有實作
1. 為現有的 email、SMS、Slack 建立 `ChannelTypeDefinition` 實作
2. 修改 `MessageSenderFactory` 使用新的註冊機制
3. 更新相關的驗證邏輯

### 階段 3: 測試和文件
1. 建立單元測試
2. 建立範例展示如何新增自定義 channel type
3. 更新 API 文件

## 使用範例

### 新增自定義 Channel Type

開發者可以這樣新增一個新的 Discord channel type：

```go
// DiscordChannelType Discord channel type 定義
type DiscordChannelType struct{}

func (d DiscordChannelType) GetName() string { return "discord" }
func (d DiscordChannelType) GetDisplayName() string { return "Discord" }
func (d DiscordChannelType) GetDescription() string { return "Send notifications via Discord webhook" }

func (d DiscordChannelType) ValidateConfig(config map[string]interface{}) error {
    webhookURL, ok := config["webhook_url"].(string)
    if !ok || webhookURL == "" {
        return errors.New("webhook_url is required for Discord channel")
    }
    // 更多驗證邏輯...
    return nil
}

func (d DiscordChannelType) CreateMessageSender(timeout time.Duration) (MessageSender, error) {
    return NewDiscordService(timeout), nil
}

// 註冊新的 channel type
func init() {
    GetChannelTypeRegistry().RegisterChannelType(&DiscordChannelType{})
}
```

### 使用新的 Channel Type

```go
// 建立 Discord channel
discordType, err := shared.NewChannelType("discord")
if err != nil {
    return err
}

channel, err := channel.NewChannel(
    channelName,
    description,
    true, // enabled
    discordType,
    templateID,
    commonSettings,
    config, // 包含 webhook_url 等 Discord 特定配置
    recipients,
    tags,
)
```

## 優點

1. **擴充性**: 新的 channel type 可以獨立開發和部署
2. **模組化**: 每個 channel type 都是獨立的模組
3. **配置驗證**: 每個 channel type 可以定義自己的配置驗證規則
4. **類型安全**: 保持編譯時期的類型檢查
5. **向後相容**: 現有的程式碼不需要修改

## 注意事項

1. **效能**: 註冊表查詢需要考慮效能，建議使用 map 結構
2. **並發安全**: 註冊表需要支援並發讀寫
3. **錯誤處理**: 需要妥善處理無效的 channel type
4. **配置管理**: 需要考慮配置的序列化和反序列化