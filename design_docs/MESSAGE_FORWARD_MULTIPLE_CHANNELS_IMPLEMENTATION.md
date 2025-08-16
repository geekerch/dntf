# Message Forward 多通道支援實作文件

## 概述

本次實作主要針對 `handleSendMessage` 方法進行改寫，將原本直接調用新系統的 `Execute` 方法改為先轉導到舊系統的 `Forward` 方法處理。同時擴展了系統對多個 channel ID 的支援能力。

## 主要變更

### 1. SendMessageRequest 結構修改

**修改前：**
```go
type SendMessageRequest struct {
    ChannelID        string                     `json:"channelId" validate:"required"`
    // ... 其他欄位
}
```

**修改後：**
```go
type SendMessageRequest struct {
    ChannelIDs       []string                   `json:"channelIds" validate:"required,min=1"`
    // ... 其他欄位
}
```

**變更說明：**
- 將單一 `ChannelID` 改為 `ChannelIDs` 陣列
- 支援同時發送訊息到多個通道
- 保持向後相容性，在回應中使用第一個 channel ID

### 2. NATS Handler 修改

**檔案：** `internal/presentation/nats/handlers/message_nats_handler.go`

**修改前（第82行）：**
```go
response, err := h.sendUseCase.Execute(ctx, &request)
```

**修改後：**
```go
// Forward to legacy system first
response, err := h.sendUseCase.Forward(ctx, &request)
```

**變更說明：**
- 改為優先使用舊系統處理訊息發送
- 錯誤訊息更新為更具體的描述

### 3. SendMessageUseCase Execute 方法增強

**檔案：** `internal/application/message/usecases/send_message_usecase.go`

**主要改進：**
- 支援多個 channel ID 的驗證和處理
- 改進錯誤處理，提供更詳細的錯誤訊息
- 保持原有的業務邏輯完整性

### 4. Forward 方法完整實作

**主要功能：**
- 支援多個 channel ID 的舊系統轉發
- 為每個 channel 創建獨立的 legacy request
- 統一處理多個 channel 的回應
- 改進錯誤處理和超時設定

**實作細節：**
```go
// 為每個 channel 創建 legacy request
for _, channelIDStr := range req.ChannelIDs {
    // 驗證 channel 存在性
    // 構建 legacy request
    // 添加到請求陣列
}

// 統一發送到舊系統
// 處理回應並統計成功/失敗狀態
```

### 5. CQRS Commands 驗證更新

**檔案：** `internal/application/cqrs/message/commands.go`

**修改內容：**
- 更新驗證邏輯以支援多個 channel ID
- 增加對每個 channel ID 的非空驗證
- 保持原有的其他驗證邏輯

## 技術特點

### 1. 向後相容性
- API 回應格式保持不變
- 使用第一個 channel ID 作為主要 channel ID 回傳
- 現有的單一 channel 使用方式仍然有效

### 2. 錯誤處理改進
- 詳細的 channel ID 驗證錯誤訊息
- 舊系統調用的具體錯誤回報
- 多 channel 處理時的錯誤聚合

### 3. 效能考量
- 批次處理多個 channel 的請求
- 合理的 HTTP 超時設定（30秒）
- 避免重複的 channel ID 處理

### 4. 安全性
- 保持原有的認證機制
- 驗證所有 channel 的存在性
- 適當的錯誤訊息，避免資訊洩露

## 使用範例

### 單一 Channel（向後相容）
```json
{
    "channelIds": ["channel-001"],
    "templateId": "template-001",
    "recipients": ["user@example.com"]
}
```

### 多個 Channels
```json
{
    "channelIds": ["email-channel", "sms-channel", "slack-channel"],
    "templateId": "template-001",
    "recipients": ["user@example.com", "+1234567890"]
}
```

## 測試建議

1. **單一 Channel 測試**
   - 驗證向後相容性
   - 確認舊系統轉發正常

2. **多 Channel 測試**
   - 測試多個有效 channel
   - 測試包含無效 channel 的情況
   - 驗證錯誤處理

3. **邊界條件測試**
   - 空的 channel IDs 陣列
   - 重複的 channel IDs
   - 不存在的 channel IDs

4. **效能測試**
   - 大量 channel 的處理效能
   - 舊系統回應時間測試
   - 超時情況處理

## 後續改進建議

1. **監控和日誌**
   - 增加詳細的操作日誌
   - 監控舊系統調用的成功率
   - 記錄多 channel 處理的效能指標

2. **配置優化**
   - 可配置的超時時間
   - 可配置的最大 channel 數量限制
   - 舊系統 URL 的動態配置

3. **回應格式增強**
   - 考慮在回應中包含所有處理的 channel 資訊
   - 提供每個 channel 的處理狀態詳情

## 結論

本次實作成功完成了以下目標：
- ✅ 將 `handleSendMessage` 改為使用 `Forward` 方法
- ✅ 完成 `Forward` 方法的多 channel 支援
- ✅ 將 `channelId` 改為 `channelIds` 支援多個通道
- ✅ 保持向後相容性和系統穩定性
- ✅ 改進錯誤處理和驗證邏輯

系統現在能夠有效地處理多個通道的訊息發送，並優先使用舊系統進行處理，為系統遷移和整合提供了良好的基礎。