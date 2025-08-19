# Recipients 欄位一致性問題分析與解決方案

## 問題描述

目前系統中 recipients 欄位在不同 API 中有不一致的設計：

### 1. Channel API 中的 recipients
```json
"recipients": [
  {
    "name": "Chien-Hsiang Chen",
    "target": "chienhsiang.chen@gmail.com",
    "type": "to"
  }
]
```
- 使用結構化的 `RecipientDTO` 物件
- 包含 `name`, `target`, `type` 欄位

### 2. Message Send API 中的 recipients
```json
"recipients": ["email1@example.com", "email2@example.com"]
```
- 使用簡單的字串陣列
- 只包含目標地址

### 3. ChannelOverrides 中的 recipients
```json
"channelOverrides": {
  "ch_email_001": {
    "recipients": [
      {
        "name": "IT Team",
        "email": "it@company.com",
        "type": "to"
      }
    ]
  }
}
```
- 使用結構化格式（與 Channel API 一致）
- 但在實際程式碼中使用 `channel.Recipients` 型別

## 不一致性影響

1. **API 使用困惑**：開發者需要記住不同 API 使用不同格式
2. **資料轉換複雜**：需要在不同格式間進行轉換
3. **維護困難**：不同的資料結構增加維護成本
4. **功能限制**：簡單字串格式無法支援複雜的收件人類型（如 CC、BCC）

## 解決方案選項

### 選項 1：統一使用結構化格式（推薦）

**優點**：
- 支援完整的收件人資訊（姓名、類型等）
- 與 Channel API 和 ChannelOverrides 保持一致
- 支援不同收件人類型（to、cc、bcc）
- 擴展性好，未來可以加入更多欄位

**缺點**：
- 需要修改現有的 Message API
- 可能影響現有客戶端

**實作方式**：
```go
// 修改 SendMessageRequest
type SendMessageRequest struct {
    ChannelIDs       []string                  `json:"channelIds" validate:"required,min=1"`
    TemplateID       string                    `json:"templateId" validate:"required"`
    Recipients       []RecipientDTO            `json:"recipients" validate:"required,min=1"` // 改為結構化
    Variables        map[string]interface{}    `json:"variables,omitempty"`
    ChannelOverrides *message.ChannelOverrides `json:"channelOverrides,omitempty"`
    Settings         *shared.CommonSettings    `json:"settings,omitempty"`
}
```

### 選項 2：保持向後相容的混合方案

**實作方式**：
```go
type SendMessageRequest struct {
    ChannelIDs         []string                  `json:"channelIds" validate:"required,min=1"`
    TemplateID         string                    `json:"templateId" validate:"required"`
    Recipients         []string                  `json:"recipients,omitempty"`         // 保留舊格式
    RecipientsDetailed []RecipientDTO            `json:"recipientsDetailed,omitempty"` // 新增結構化格式
    Variables          map[string]interface{}    `json:"variables,omitempty"`
    ChannelOverrides   *message.ChannelOverrides `json:"channelOverrides,omitempty"`
    Settings           *shared.CommonSettings    `json:"settings,omitempty"`
}
```

**驗證邏輯**：
- 兩個欄位至少要有一個不為空
- 如果兩個都有值，優先使用 `recipientsDetailed`

### 選項 3：維持現狀但改進文件

**實作方式**：
- 在 API 文件中明確說明不同 API 使用不同格式的原因
- 提供轉換工具或範例

## 建議實作方案

推薦採用**選項 1**，原因如下：

1. **長期一致性**：統一的資料格式更容易維護
2. **功能完整性**：支援完整的收件人資訊
3. **擴展性**：未來可以輕鬆加入新功能
4. **與現有設計對齊**：與 Channel API 保持一致

## 實作步驟

1. **第一階段**：修改 DTO 結構
   - 更新 `SendMessageRequest` 使用 `[]RecipientDTO`
   - 更新相關的驗證邏輯

2. **第二階段**：更新業務邏輯
   - 修改 use case 處理新的資料結構
   - 更新資料轉換邏輯

3. **第三階段**：更新 API 文件
   - 更新 Swagger 文件
   - 更新設計文件範例

4. **第四階段**：測試與部署
   - 單元測試
   - 整合測試
   - API 測試

## 向後相容性考量

如果需要保持向後相容性，可以：

1. **API 版本控制**：建立新版本的 API
2. **漸進式遷移**：先支援兩種格式，再逐步淘汰舊格式
3. **轉換層**：在 handler 層進行格式轉換

## 實作完成

✅ **已完成統一化實作**

### 修改內容

1. **SendMessageRequest**：
   ```go
   Recipients []channeldtos.RecipientDTO `json:"recipients" validate:"required,min=1"`
   ```

2. **MessageResponse**：
   ```go
   Recipients []channeldtos.RecipientDTO `json:"recipients"`
   ```

3. **Use Case 修改**：
   - 更新 `send_message_usecase.go` 處理新的 recipients 格式
   - 修改 legacy system 整合邏輯
   - 新增 `ToMessageResponseWithRecipients` 函數

4. **CQRS 驗證**：
   - 更新 command 驗證邏輯，檢查 `Name` 和 `Type` 欄位

### Recipients 格式範例

現在所有 API 都使用統一的結構化格式：

```json
{
  "recipients": [
    {
      "name": "John Doe",
      "target": "john@example.com",
      "type": "to"
    },
    {
      "name": "Jane Smith", 
      "target": "#general",
      "type": "channel"
    },
    {
      "name": "Support Team",
      "target": "+1234567890", 
      "type": "phone"
    }
  ]
}
```

### 不同 Channel 類型的 Recipients

- **Email**: `type="to/cc/bcc"`, `target="email@example.com"`
- **Slack**: `type="channel/user"`, `target="#general"` 或 `"@username"`
- **SMS**: `type="phone"`, `target="+1234567890"`

## 結論

✅ **統一化完成**：所有 API 現在使用一致的結構化 recipients 格式
✅ **靈活支援**：根據不同 channel 類型使用不同的 `type` 和 `target` 值
✅ **向後相容**：由於尚未上線，直接修改不影響現有用戶
✅ **擴展性**：未來可以輕鬆加入新的 recipient 欄位或 channel 類型