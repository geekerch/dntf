# Template 更新與刪除時 Legacy Channel 同步實現

## 概述

本文件說明在 Template 更新和刪除操作時，如何同步更新使用該 Template 的 Legacy Channel 的實現方案。

## 背景

Legacy 系統在建立 Channel 時會將 Template 的內容直接嵌入到 Channel 配置中。因此當 Template 更新或刪除時，需要同步更新所有使用該 Template 的 Legacy Channel，以確保資料一致性。

## 實現方案

### 1. Template 更新時的 Legacy Channel 同步

**位置**: `internal/application/template/usecases/update_template_usecase.go`

**實現邏輯**:
1. 更新 Template 實體
2. 儲存更新後的 Template
3. 查找所有使用該 Template 的 Channel
4. 對每個 Legacy Channel 發送 HTTP PUT 請求更新其配置

**關鍵方法**:
- `updateLegacyChannelsUsingTemplate()`: 查找並更新所有使用該 Template 的 Legacy Channel
- `updateLegacyChannel()`: 更新單個 Legacy Channel

### 2. Template 刪除時的 Legacy Channel 同步

**位置**: `internal/application/template/usecases/delete_template_usecase.go`

**實現邏輯**:
1. 在刪除 Template 前先取得 Template 實體
2. 查找所有使用該 Template 的 Channel
3. 對每個 Legacy Channel 發送 HTTP PUT 請求，使用預設或備用的 Template 內容
4. 刪除 Template

**關鍵方法**:
- `updateLegacyChannelsForTemplateDelete()`: 查找並更新所有使用該 Template 的 Legacy Channel
- `updateLegacyChannelForTemplateDelete()`: 更新單個 Legacy Channel，使用預設內容

## 技術細節

### Legacy System API 整合

**API 端點**: `PUT /api/v2.0/Groups/{channelId}`

**請求格式**:
```json
{
  "name": "channel name",
  "description": "channel description", 
  "type": "channel type",
  "levelName": "Critical",
  "config": {
    "host": "smtp.example.com",
    "port": 587,
    "secure": true,
    "method": "SMTP",
    "username": "user@example.com",
    "password": "password",
    "senderEmail": "sender@example.com",
    "emailSubject": "updated subject",
    "template": "updated template content"
  },
  "sendList": [
    {
      "firstName": "John",
      "lastName": "Doe", 
      "recipientType": "email",
      "target": "john.doe@example.com"
    }
  ]
}
```

### 錯誤處理策略

- **Best Effort 原則**: Legacy Channel 同步失敗不會影響 Template 的更新或刪除操作
- **錯誤記錄**: 同步失敗時會記錄警告訊息，但不會中斷主要操作
- **個別處理**: 每個 Channel 的同步失敗不會影響其他 Channel 的處理

### 分頁處理

由於 Channel 數量可能很大，使用分頁查詢來避免記憶體問題：
- 每次查詢最多 100 個 Channel (`MaxResultCount: 100`)
- 如需處理更多 Channel，可以實現多次查詢的邏輯

## 資料結構

### LegacyChannelRequest
```go
type LegacyChannelRequest struct {
    Name        string         `json:"name"`
    Description string         `json:"description"`
    Type        string         `json:"type"`
    LevelName   string         `json:"levelName"`
    Config      LegacyConfig   `json:"config"`
    SendList    []SendListItem `json:"sendList"`
}
```

### LegacyConfig
```go
type LegacyConfig struct {
    Host         string `json:"host"`
    Port         int    `json:"port"`
    Secure       bool   `json:"secure"`
    Method       string `json:"method"`
    Username     string `json:"username"`
    Password     string `json:"password"`
    SenderEmail  string `json:"senderEmail"`
    EmailSubject string `json:"emailSubject"`
    Template     string `json:"template"`
}
```

## 配置需求

需要在 `config.Config` 中配置 Legacy System 的連接資訊：
- `LegacySystem.URL`: Legacy System 的基礎 URL
- `LegacySystem.Token`: 用於認證的 Bearer Token

## 測試考量

1. **單元測試**: 測試 Template 更新/刪除的核心邏輯
2. **整合測試**: 測試與 Legacy System 的 HTTP 通信
3. **錯誤情境測試**: 測試 Legacy System 不可用時的錯誤處理
4. **效能測試**: 測試大量 Channel 時的處理效能

## 未來改進

1. **批次處理**: 實現批次更新 Legacy Channel 以提高效能
2. **非同步處理**: 使用訊息佇列進行非同步的 Legacy Channel 更新
3. **重試機制**: 實現自動重試機制處理暫時性的網路錯誤
4. **監控與告警**: 添加監控指標和告警機制

## 結論

此實現確保了 Template 更新和刪除操作與 Legacy System 的資料一致性，同時採用 Best Effort 策略避免影響主要業務流程。