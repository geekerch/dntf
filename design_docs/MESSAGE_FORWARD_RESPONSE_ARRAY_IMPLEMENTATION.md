# 訊息轉發回應陣列實作

## 概述
修正 `SendMessageUseCase.Forward` 函數的回應格式，從回傳單一 `MessageResponse` 改為回傳 `MessageResponse` 陣列，每個 channel 對應一個回應物件。

## 問題描述
原本的 `Forward` 函數在處理多個 channel 時，只回傳一個整合的回應物件，無法清楚表示每個 channel 的個別處理狀況。這不符合預期的設計，應該要像第288行的處理方式一樣，為每個 channel 提供對應的回應。

## 解決方案

### 1. 函數簽名修改
```go
// 修改前
func (uc *SendMessageUseCase) Forward(ctx context.Context, req *dtos.SendMessageRequest) (*dtos.MessageResponse, error)

// 修改後  
func (uc *SendMessageUseCase) Forward(ctx context.Context, req *dtos.SendMessageRequest) ([]*dtos.MessageResponse, error)
```

### 2. 回應邏輯重構
- **移除整體狀態判斷**：不再統一判斷所有 channel 的成功/失敗狀態
- **個別 channel 處理**：為每個 `LegacyMessageResponse` 建立對應的 `MessageResponse`
- **詳細結果資訊**：為每個收件人建立 `MessageResultResponse`，包含個別的成功/失敗狀態

### 3. 回應結構
每個 `MessageResponse` 包含：
- `ID`: 唯一識別碼
- `ChannelID`: 對應的 channel ID (來自 `result.GroupID`)
- `TemplateID`: 範本 ID
- `Recipients`: 收件人清單
- `Variables`: 變數
- `Status`: 該 channel 的整體狀態
- `Results`: 每個收件人的詳細處理結果
- `CreatedAt` / `SentAt`: 時間戳記

### 4. 錯誤處理改進
- 不再因為部分 channel 失敗就整體回傳錯誤
- 每個 channel 的成功/失敗狀態獨立處理
- 失敗的 channel 會在其 `Status` 和 `Results` 中反映錯誤資訊

## 影響範圍

### 修改的檔案
1. `internal/application/message/usecases/send_message_usecase.go`
   - 函數簽名修改
   - 回應邏輯重構
   - 錯誤處理簡化

2. `internal/presentation/nats/handlers/message_nats_handler.go`
   - 更新變數名稱以反映陣列型別
   - 無需修改 `sendSuccessResponse` 函數（已支援 `interface{}` 型別）

### 相容性
- NATS 回應格式保持相容，因為 `sendSuccessResponse` 接受 `interface{}` 型別
- 呼叫端現在會收到陣列格式的回應，需要相應調整處理邏輯

## 測試建議
1. 測試單一 channel 的情況
2. 測試多個 channel 的情況
3. 測試部分 channel 成功、部分失敗的混合情況
4. 驗證每個 channel 的回應都包含正確的資訊

## 效益
- **清晰的回應結構**：每個 channel 都有明確的處理結果
- **更好的錯誤追蹤**：可以精確知道哪個 channel 或收件人處理失敗
- **符合設計預期**：與第288行的處理邏輯保持一致
- **提升可維護性**：回應結構更加直觀和易於理解