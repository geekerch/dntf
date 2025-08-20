# HTTP 與 NATS DTO 格式統一實現

## 概述

本文件說明將 HTTP API 的 DTO 格式與 NATS 訊息格式進行統一的實現方案，確保兩種通訊協定使用一致的資料結構。

## 背景

原本 HTTP 和 NATS 的資料格式存在不一致問題：
- 分頁參數不統一（Template 使用 page/size，Channel 使用 skipCount/maxResultCount）
- HTTP 回應格式不統一（Template 使用 success/error/message，其他使用 data/error）
- 部分欄位解析邏輯有問題

## 實現的統一化

### 1. 分頁參數統一

**修正前**：
```go
// Template DTO
type ListTemplatesRequest struct {
    Page int `json:"page,omitempty"`
    Size int `json:"size,omitempty"`
}

// Channel DTO  
type ListChannelsRequest struct {
    SkipCount      int `json:"skipCount"`
    MaxResultCount int `json:"maxResultCount"`
}
```

**修正後**：
```go
// 統一使用 skipCount/maxResultCount
type ListTemplatesRequest struct {
    SkipCount      int `json:"skipCount,omitempty"`
    MaxResultCount int `json:"maxResultCount,omitempty"`
}
```

### 2. HTTP 回應格式統一

**修正前**：
```go
// Template Handler
c.JSON(http.StatusOK, gin.H{
    "success": true,
    "data":    response,
    "message": "Template created successfully",
})

// Channel Handler
c.JSON(http.StatusOK, gin.H{
    "data":  response,
    "error": nil,
})
```

**修正後**：
```go
// 統一使用 data/error 格式
c.JSON(http.StatusOK, gin.H{
    "data":  response,
    "error": nil,
})

// 錯誤回應格式
c.JSON(http.StatusBadRequest, gin.H{
    "data":  nil,
    "error": map[string]interface{}{
        "code":    "ERROR_CODE",
        "message": "Error message",
    },
})
```

### 3. 查詢參數解析修正

**修正前**：
```go
// Template Handler - channelType 解析但未設定
if channelType := c.Query("channelType"); channelType != "" {
    // Note: You might want to add validation for channel type here
    // For now, we'll assume it's valid
}
```

**修正後**：
```go
// 正確解析並設定 channelType
if channelType := c.Query("channelType"); channelType != "" {
    if ct, err := shared.NewChannelType(channelType); err == nil {
        req.ChannelType = &ct
    }
}
```

## 修改的檔案

### DTO 結構修改
- `internal/application/template/dtos/template_dto.go`
  - 將 `ListTemplatesRequest` 的分頁欄位從 `Page/Size` 改為 `SkipCount/MaxResultCount`
  - 更新 `ToPagination()` 方法邏輯

### HTTP Handler 修改
- `internal/presentation/http/handlers/template_handler.go`
  - 統一回應格式為 `{"data": ..., "error": ...}`
  - 修正 channelType 查詢參數解析
  - 更新分頁參數解析邏輯
  - 更新 Swagger 註解

### CQRS Handler 修改
- `internal/application/cqrs/template/handlers.go`
  - 修正分頁欄位引用錯誤

## Swagger 文件更新

重新產生的 Swagger 文件包含：
- 統一的分頁參數說明
- 正確的 DTO 結構定義
- 一致的回應格式範例

## NATS 與 HTTP 的設計原則

### RESTful API 設計原則保持
- ID 參數仍通過 URL 路徑傳遞（如 `/templates/{id}`）
- 查詢參數用於過濾和分頁
- HTTP 方法語義保持標準（GET/POST/PUT/DELETE）

### NATS 訊息格式
- 使用統一的 `NATSRequest` 包裝格式
- `data` 欄位包含實際的業務資料
- ID 資訊包含在 `data` 中（因為訊息傳遞無 URL 概念）

## 驗證方式

1. **編譯檢查**：`go build ./...` 無錯誤
2. **Swagger 產生**：成功重新產生 API 文件
3. **格式一致性**：HTTP 和 NATS 使用相同的 DTO 結構

## 後續維護

1. **新增 API 時**：確保使用統一的回應格式和分頁參數
2. **DTO 修改時**：同時考慮 HTTP 和 NATS 的相容性
3. **測試覆蓋**：確保兩種協定的行為一致

## 結論

此次統一化確保了 HTTP 和 NATS 介面的資料格式一致性，提升了 API 的可維護性和使用者體驗。所有修改都保持了 RESTful 設計原則，同時滿足了 NATS 訊息傳遞的需求。