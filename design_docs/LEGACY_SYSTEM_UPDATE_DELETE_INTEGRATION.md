# Legacy 系統 Update 與 Delete 整合實作

## 概述

本次實作為 `UpdateChannelUseCase` 和 `DeleteChannelUseCase` 新增了轉發到 legacy 系統的功能，與現有的 `CreateChannelUseCase` 保持一致的整合模式。

## 實作內容

### 1. UpdateChannelUseCase 修改

#### 新增依賴
- 新增 `templateRepo template.TemplateRepository` 依賴
- 新增 `config *config.Config` 依賴

#### 新增功能
- 實作 `forwardUpdateToLegacySystem` 函數
- 在更新 channel 前先轉發到 legacy 系統
- 使用 PUT 方法呼叫 `/api/v2.0/Groups/{groupId}` endpoint

#### API 規格
```
PUT /v2.0/Groups/{groupId}
Authorization: Bearer {token}
Content-Type: application/json

Body:
{
  "name": "emailGroupTest",
  "description": "email group",
  "type": "email",
  "levelName": "Critical",
  "config": {
    "host": "mailapp.advantech.com.tw",
    "port": 465,
    "secure": true,
    "method": "ssl",
    "username": "TEST_USER",
    "password": "TEST_PWD",
    "senderEmail": "test@advantech.com.tw",
    "emailSubject": "Test Subject",
    "template": "Hi, Have a good day!"
  },
  "sendList": [{
    "firstName": "Firstname",
    "lastName": "Lastname",
    "recipientType": "to",
    "target": "test@advantech.com.tw"
  }]
}
```

### 2. DeleteChannelUseCase 修改

#### 新增依賴
- 新增 `config *config.Config` 依賴

#### 新增功能
- 實作 `forwardDeleteToLegacySystem` 函數
- 在刪除 channel 前先轉發到 legacy 系統
- 使用 DELETE 方法呼叫 `/api/v2.0/Groups` endpoint

#### API 規格
```
DELETE /v2.0/Groups
Authorization: Bearer {token}
Content-Type: application/json

Body:
[
  "string"
]
```

### 3. 依賴注入更新

在 `cmd/server/main.go` 中更新了 use case 的建構函數呼叫：

```go
// 更新前
updateChannelUseCase := usecases.NewUpdateChannelUseCase(channelRepo, channelValidator)
deleteChannelUseCase := usecases.NewDeleteChannelUseCase(channelRepo, channelValidator)

// 更新後
updateChannelUseCase := usecases.NewUpdateChannelUseCase(channelRepo, templateRepo, channelValidator, cfg)
deleteChannelUseCase := usecases.NewDeleteChannelUseCase(channelRepo, channelValidator, cfg)
```

## 實作細節

### 共用資料結構

重複使用了 `CreateChannelUseCase` 中定義的資料結構：
- `LegacyChannelRequest`
- `LegacyConfig`
- `SendListItem`

### 錯誤處理

- HTTP 狀態碼 >= 400 時回傳錯誤
- 包含 legacy 系統回應的詳細錯誤訊息
- 如果 legacy 系統呼叫失敗，整個操作會回滾

### Template 整合

Update 操作中：
- 如果有 `TemplateID`，會從資料庫查詢 template
- 優先使用 template 的 subject 和 content
- 如果沒有 template，則使用 config 中的值

## 執行流程

### Update Channel 流程
1. 驗證輸入參數
2. 轉換為 domain objects
3. 業務邏輯驗證
4. 查詢現有 channel
5. 檢查 channel 是否已刪除
6. **轉發到 legacy 系統 (新增)**
7. 更新 channel
8. 持久化
9. 回傳回應

### Delete Channel 流程
1. 驗證輸入參數
2. 轉換為 domain objects
3. 業務邏輯驗證
4. 查詢 channel
5. **轉發到 legacy 系統 (新增)**
6. 執行軟刪除
7. 持久化
8. 回傳回應

## 設定需求

需要在環境變數中設定 legacy 系統的連線資訊：
```bash
LEGACY_SYSTEM_URL=https://legacy-system.example.com
LEGACY_SYSTEM_TOKEN=your_bearer_token_here
```

## 測試驗證

- ✅ 程式碼編譯成功
- ✅ 保持與 CreateChannelUseCase 一致的實作模式
- ✅ 錯誤處理完整
- ✅ 依賴注入正確更新

## 向後相容性

此變更完全向後相容：
- 現有的 API 介面不變
- 只是在內部流程中新增了 legacy 系統的呼叫
- 如果 legacy 系統無法連線，會回傳錯誤但不會影響系統穩定性

## 未來改進建議

1. 可考慮新增重試機制處理 legacy 系統暫時無法連線的情況
2. 可新增 circuit breaker 模式避免 legacy 系統故障影響主系統
3. 可考慮非同步處理 legacy 系統呼叫以提升效能
4. 可新增詳細的 logging 記錄 legacy 系統互動過程