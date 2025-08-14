# Presentation Layer 實作文件

## 概述

本文件描述 Channel API 的 Presentation Layer（展示層）實作，包含 RESTful API 處理器和 NATS 訊息處理器的完整實作。

## 架構設計

Presentation Layer 遵循 Clean Architecture 原則，負責處理外部請求並將其轉換為應用層可理解的格式。

```
internal/presentation/
├── http/                          # HTTP 相關處理
│   ├── handlers/                  # HTTP 處理器
│   │   └── channel_handler.go     # Channel RESTful API 處理器
│   ├── middleware/                # HTTP 中介軟體
│   │   ├── error_handler.go       # 錯誤處理中介軟體
│   │   ├── request_logger.go      # 請求日誌中介軟體
│   │   └── response_formatter.go  # 回應格式化中介軟體
│   └── routes/                    # 路由設定
│       ├── channel_routes.go      # Channel 路由設定
│       └── router.go              # 主路由設定
├── nats/                          # NATS 相關處理
│   └── handlers/                  # NATS 處理器
│       ├── channel_nats_handler.go # Channel NATS 訊息處理器
│       └── handler_manager.go     # NATS 處理器管理器
└── server.go                      # 展示層伺服器
```

## RESTful API 處理器

### Channel Handler

`ChannelHandler` 負責處理所有 Channel 相關的 HTTP 請求：

#### 支援的 API 端點

| 方法 | 路徑 | 功能 | 處理函數 |
|------|------|------|----------|
| POST | `/api/v1/channels` | 建立通道 | `CreateChannel` |
| GET | `/api/v1/channels` | 查詢通道列表 | `ListChannels` |
| GET | `/api/v1/channels/:id` | 取得單一通道 | `GetChannel` |
| PUT | `/api/v1/channels/:id` | 更新通道 | `UpdateChannel` |
| DELETE | `/api/v1/channels/:id` | 刪除通道 | `DeleteChannel` |

#### 特色功能

1. **輸入驗證**: 使用 Gin 的 binding 功能進行請求參數驗證
2. **錯誤處理**: 統一的錯誤回應格式
3. **查詢參數支援**: 支援分頁、篩選等查詢參數
4. **RESTful 設計**: 遵循 RESTful API 設計原則

### HTTP 中介軟體

#### 錯誤處理中介軟體 (`ErrorHandler`)
- 統一處理 panic 和錯誤
- 提供標準化的錯誤回應格式
- 支援 404 和 405 錯誤處理

#### 請求日誌中介軟體 (`RequestLogger`)
- 使用結構化日誌記錄所有 HTTP 請求
- 包含請求方法、路徑、狀態碼、延遲時間等資訊
- 支援請求 ID 追蹤

#### 回應格式化中介軟體 (`ResponseFormatter`)
- 提供標準化的 API 回應格式
- 支援 CORS 跨域請求
- 統一的成功和錯誤回應結構

### 路由設定

#### 模組化路由設計
- 每個功能模組有獨立的路由設定檔
- 支援版本控制 (`/api/v1/`)
- 健康檢查端點 (`/health`)

## NATS 訊息處理器

### Channel NATS Handler

`ChannelNATSHandler` 負責處理所有 Channel 相關的 NATS 訊息：

#### 支援的 NATS 主題

| 主題 | 功能 | 處理函數 |
|------|------|----------|
| `eco1j.infra.eventcenter.channel.create` | 建立通道 | `handleCreateChannel` |
| `eco1j.infra.eventcenter.channel.get` | 取得單一通道 | `handleGetChannel` |
| `eco1j.infra.eventcenter.channel.list` | 查詢通道列表 | `handleListChannels` |
| `eco1j.infra.eventcenter.channel.update` | 更新通道 | `handleUpdateChannel` |
| `eco1j.infra.eventcenter.channel.delete` | 刪除通道 | `handleDeleteChannel` |

#### NATS 訊息格式

**請求格式**:
```json
{
  "requestId": "unique-request-id",
  "data": {
    // 具體的請求資料
  },
  "timestamp": 1640995200
}
```

**回應格式**:
```json
{
  "requestId": "unique-request-id",
  "success": true,
  "data": {
    // 回應資料
  },
  "error": {
    "code": "ERROR_CODE",
    "message": "Error message",
    "details": "Detailed error information"
  },
  "timestamp": 1640995200
}
```

#### 特色功能

1. **請求追蹤**: 每個請求都有唯一的 Request ID
2. **錯誤處理**: 統一的錯誤回應格式
3. **結構化日誌**: 詳細的操作日誌記錄
4. **資料轉換**: 自動處理 JSON 序列化/反序列化

### NATS 處理器管理器

`HandlerManager` 負責管理所有 NATS 訊息處理器：

#### 功能特色

1. **統一管理**: 集中管理所有 NATS 處理器的註冊和生命週期
2. **健康檢查**: 提供 NATS 連線狀態檢查
3. **優雅關閉**: 支援優雅關閉所有 NATS 連線
4. **擴展性**: 易於新增新的訊息處理器

## 展示層伺服器

### Server 組件

`Server` 是展示層的主要組件，負責：

1. **HTTP 伺服器管理**: 啟動和關閉 HTTP 伺服器
2. **NATS 處理器管理**: 註冊和管理 NATS 訊息處理器
3. **優雅關閉**: 支援優雅關閉所有服務
4. **健康檢查**: 提供整體健康狀態檢查

#### 配置選項

```go
type ServerConfig struct {
    HTTPPort    string                        // HTTP 伺服器埠號
    HTTPTimeout time.Duration                 // HTTP 請求超時時間
    ChannelHandler *handlers.ChannelHandler   // Channel HTTP 處理器
    NATSManager *natshandlers.HandlerManager  // NATS 處理器管理器
}
```

## 依賴注入設計

### 依賴關係

```
Presentation Layer
├── HTTP Handlers → Application Use Cases
├── NATS Handlers → Application Use Cases
└── Middleware → Infrastructure Services (Logger, etc.)
```

### 注入流程

1. **Use Cases 注入**: HTTP 和 NATS 處理器接收 Application Layer 的 Use Cases
2. **基礎設施注入**: 中介軟體接收 Infrastructure Layer 的服務（如 Logger）
3. **配置注入**: 伺服器接收所有必要的處理器和配置

## 錯誤處理策略

### HTTP 錯誤處理

1. **輸入驗證錯誤**: 回傳 400 Bad Request
2. **資源不存在**: 回傳 404 Not Found
3. **業務邏輯錯誤**: 回傳 500 Internal Server Error
4. **方法不支援**: 回傳 405 Method Not Allowed

### NATS 錯誤處理

1. **請求解析錯誤**: `INVALID_REQUEST` 錯誤碼
2. **執行錯誤**: `EXECUTION_ERROR` 錯誤碼
3. **系統錯誤**: `SYSTEM_ERROR` 錯誤碼

## 日誌記錄

### 結構化日誌

使用 Zap 進行結構化日誌記錄：

```go
logger.Info("HTTP Request",
    zap.String("method", "POST"),
    zap.String("path", "/api/v1/channels"),
    zap.Int("status", 201),
    zap.Duration("latency", time.Millisecond*150),
)
```

### 日誌級別

- **Info**: 正常操作記錄
- **Error**: 錯誤情況記錄
- **Debug**: 除錯資訊（開發環境）

## 效能考量

### HTTP 效能

1. **連線池**: 使用 HTTP/1.1 Keep-Alive
2. **超時設定**: 合理的讀寫超時時間
3. **中介軟體順序**: 優化中介軟體執行順序

### NATS 效能

1. **連線重用**: 重用 NATS 連線
2. **非同步處理**: 非阻塞的訊息處理
3. **錯誤恢復**: 自動重連機制

## 安全考量

### HTTP 安全

1. **CORS 設定**: 適當的跨域資源共享設定
2. **請求大小限制**: 防止大型請求攻擊
3. **請求 ID**: 追蹤和審計請求

### NATS 安全

1. **主題驗證**: 驗證 NATS 主題格式
2. **訊息驗證**: 驗證訊息格式和內容
3. **錯誤資訊**: 避免洩露敏感資訊

## 測試策略

### 單元測試

1. **處理器測試**: 測試各個處理器的邏輯
2. **中介軟體測試**: 測試中介軟體功能
3. **Mock 依賴**: 使用 Mock 隔離外部依賴

### 整合測試

1. **HTTP API 測試**: 端到端 API 測試
2. **NATS 訊息測試**: 訊息處理整合測試
3. **錯誤場景測試**: 各種錯誤情況測試

## 擴展指南

### 新增 HTTP 處理器

1. 在 `handlers/` 目錄建立新的處理器檔案
2. 實作處理器結構和方法
3. 在 `routes/` 目錄新增路由設定
4. 更新 `router.go` 註冊新路由

### 新增 NATS 處理器

1. 在 `nats/handlers/` 目錄建立新的處理器檔案
2. 實作處理器結構和訊息處理方法
3. 更新 `handler_manager.go` 註冊新處理器
4. 新增對應的 Use Cases 依賴

## 總結

Presentation Layer 的實作提供了：

1. **完整的 RESTful API**: 支援所有 Channel 操作
2. **完整的 NATS 訊息處理**: 支援事件驅動架構
3. **統一的錯誤處理**: 標準化的錯誤回應
4. **結構化日誌**: 便於監控和除錯
5. **優雅關閉**: 支援零停機部署
6. **高度可擴展**: 易於新增新功能

這個實作為整個 Channel API 系統提供了穩定、可靠的對外介面，支援多種通訊協定和使用場景。