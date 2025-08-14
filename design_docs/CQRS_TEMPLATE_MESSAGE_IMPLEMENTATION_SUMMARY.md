# CQRS 範本和訊息實作總結

## 概述

本文件總結了為 Channel API 系統實作的範本和訊息 CQRS（命令查詢責任分離）元件，提供了企業級的可擴展架構模式。

## 已實作的功能

### 1. CQRS 應用層 - 範本

**位置**: `internal/application/cqrs/template/`

#### 命令 (`commands.go`)
- `CreateTemplateCommand` - 建立範本命令，包含完整驗證
- `UpdateTemplateCommand` - 更新範本命令，支援部分更新
- `DeleteTemplateCommand` - 刪除範本命令，包含存在性檢查

#### 查詢 (`queries.go`)
- `GetTemplateQuery` - 取得單一範本查詢
- `ListTemplatesQuery` - 列出範本查詢，支援進階篩選、排序和分頁

#### 事件 (`events.go`)
- `TemplateCreatedEvent` - 範本建立事件
- `TemplateUpdatedEvent` - 範本更新事件
- `TemplateDeletedEvent` - 範本刪除事件

#### 處理器 (`handlers.go`)
- `TemplateCommandHandlers` - 命令處理器集合
- `TemplateQueryHandlers` - 查詢處理器集合
- 個別命令和查詢處理器，實作 CQRS 介面

### 2. CQRS 應用層 - 訊息

**位置**: `internal/application/cqrs/message/`

#### 命令 (`commands.go`)
- `SendMessageCommand` - 發送訊息命令，包含收件人驗證

#### 查詢 (`queries.go`)
- `GetMessageQuery` - 取得訊息狀態查詢

#### 事件 (`events.go`)
- `MessageSentEvent` - 訊息發送成功事件
- `MessageFailedEvent` - 訊息發送失敗事件
- `MessageDeliveredEvent` - 訊息送達事件

#### 處理器 (`handlers.go`)
- `MessageCommandHandlers` - 訊息命令處理器
- `MessageQueryHandlers` - 訊息查詢處理器
- 個別命令和查詢處理器

### 3. CQRS 展示層 - HTTP 處理器

**位置**: `internal/presentation/http/handlers/`

#### CQRS 範本處理器 (`cqrs_template_handler.go`)
- `CreateTemplate` - POST `/api/v2/templates`
- `GetTemplate` - GET `/api/v2/templates/{id}`
- `ListTemplates` - GET `/api/v2/templates` (支援進階查詢參數)
- `UpdateTemplate` - PUT `/api/v2/templates/{id}`
- `DeleteTemplate` - DELETE `/api/v2/templates/{id}`

#### CQRS 訊息處理器 (`cqrs_message_handler.go`)
- `SendMessage` - POST `/api/v2/messages/send`
- `GetMessage` - GET `/api/v2/messages/{id}`

### 4. CQRS 路由設定

**位置**: `internal/presentation/http/routes/`

#### CQRS 範本路由 (`cqrs_template_routes.go`)
- 設定 `/api/v2/templates` 的所有 RESTful 路由
- 整合 Gin 路由框架

#### CQRS 訊息路由 (`cqrs_message_routes.go`)
- 設定 `/api/v2/messages` 的訊息操作路由
- 整合 Gin 路由框架

### 5. 主應用程式整合

**位置**: `cmd/server/main.go`

#### 新增的整合：
- 匯入範本和訊息 CQRS 套件
- 註冊所有 CQRS 命令處理器到 CQRS 管理器
- 註冊所有 CQRS 查詢處理器到 CQRS 管理器
- 建立和配置 CQRS HTTP 處理器
- 更新伺服器配置以包含 CQRS 處理器

## API 端點對比

### 傳統 API (v1) vs CQRS API (v2)

| 功能 | 傳統 API (v1) | CQRS API (v2) |
|------|---------------|---------------|
| 建立範本 | `POST /api/v1/templates` | `POST /api/v2/templates` |
| 取得範本 | `GET /api/v1/templates/{id}` | `GET /api/v2/templates/{id}` |
| 列出範本 | `GET /api/v1/templates` | `GET /api/v2/templates` |
| 更新範本 | `PUT /api/v1/templates/{id}` | `PUT /api/v2/templates/{id}` |
| 刪除範本 | `DELETE /api/v1/templates/{id}` | `DELETE /api/v2/templates/{id}` |
| 發送訊息 | `POST /api/v1/messages/send` | `POST /api/v2/messages/send` |
| 取得訊息 | `GET /api/v1/messages/{id}` | `GET /api/v2/messages/{id}` |

## CQRS 架構優勢

### 1. 命令查詢分離
- **命令 (Commands)**: 處理狀態變更操作（建立、更新、刪除）
- **查詢 (Queries)**: 處理資料檢索操作（取得、列表）
- **清晰的責任分離**: 每個操作都有明確的目的和職責

### 2. 事件驅動架構
- **領域事件**: 每個命令執行後發布相應事件
- **鬆耦合**: 事件消費者可以獨立處理業務邏輯
- **可擴展性**: 容易新增新的事件處理器

### 3. 進階查詢功能
- **複雜篩選**: 支援多欄位篩選和運算子
- **動態排序**: 支援多欄位排序
- **分頁**: 高效的分頁機制
- **欄位選擇**: 可選擇性返回特定欄位

### 4. 效能最佳化
- **讀寫分離**: 查詢和命令可以使用不同的最佳化策略
- **快取友善**: 查詢操作容易實作快取
- **水平擴展**: 讀取和寫入操作可以獨立擴展

## 實作特色

### 1. 型別安全
- 強型別命令和查詢物件
- 編譯時期型別檢查
- 明確的輸入驗證

### 2. 錯誤處理
- 統一的錯誤回應格式
- 詳細的驗證錯誤訊息
- 事件發布失敗的優雅處理

### 3. 可測試性
- 每個處理器都可以獨立測試
- Mock 友善的介面設計
- 清晰的依賴注入

### 4. 監控和日誌
- 事件發布的日誌記錄
- 命令和查詢的執行追蹤
- 錯誤和效能監控支援

## 使用範例

### 建立範本 (CQRS)
```bash
curl -X POST http://localhost:8080/api/v2/templates \
  -H "Content-Type: application/json" \
  -d '{
    "name": "welcome-email",
    "channelType": "email",
    "subject": "歡迎！",
    "content": "您好 {{name}}，歡迎使用我們的服務！",
    "variables": ["name"],
    "tags": ["welcome"]
  }'
```

### 進階查詢範本 (CQRS)
```bash
curl "http://localhost:8080/api/v2/templates?channelType=email&tags=welcome&page=1&size=10&sortBy=name&order=asc"
```

### 發送訊息 (CQRS)
```bash
curl -X POST http://localhost:8080/api/v2/messages/send \
  -H "Content-Type: application/json" \
  -d '{
    "channelId": "channel-id",
    "templateId": "template-id",
    "recipients": ["user@example.com"],
    "variables": {"name": "張小明"}
  }'
```

## 測試

提供了全面的測試腳本 (`tmp_rovodev_test_cqrs.sh`) 來驗證：
- CQRS 範本的所有 CRUD 操作
- CQRS 訊息發送功能
- 傳統 API vs CQRS API 的效能比較
- 進階查詢功能
- 事件發布機制
- 錯誤處理

## 後續增強建議

### 1. 事件溯源 (Event Sourcing)
- 實作事件存儲
- 事件重播功能
- 快照機制

### 2. 讀取模型最佳化
- 專用的讀取資料庫
- 物化視圖
- 快取策略

### 3. 進階 CQRS 功能
- 命令驗證管道
- 查詢快取
- 批次命令處理

### 4. 監控和分析
- 命令/查詢效能指標
- 事件處理監控
- 業務指標追蹤

### 5. 安全性增強
- 命令授權
- 查詢權限控制
- 審計日誌

## 檔案清單

### 新建立的檔案：
- `internal/application/cqrs/template/commands.go`
- `internal/application/cqrs/template/queries.go`
- `internal/application/cqrs/template/events.go`
- `internal/application/cqrs/template/handlers.go`
- `internal/application/cqrs/message/commands.go`
- `internal/application/cqrs/message/queries.go`
- `internal/application/cqrs/message/events.go`
- `internal/application/cqrs/message/handlers.go`
- `internal/presentation/http/handlers/cqrs_template_handler.go`
- `internal/presentation/http/handlers/cqrs_message_handler.go`
- `internal/presentation/http/routes/cqrs_template_routes.go`
- `internal/presentation/http/routes/cqrs_message_routes.go`

### 修改的檔案：
- `internal/presentation/http/routes/router.go` - 新增 CQRS 路由支援
- `internal/presentation/server.go` - 新增 CQRS 處理器支援
- `cmd/server/main.go` - 整合 CQRS 元件

## 總結

CQRS 範本和訊息元件的實作為系統提供了：

- **企業級架構**: 遵循 CQRS 和事件驅動設計模式
- **高可擴展性**: 支援讀寫分離和水平擴展
- **完整功能**: 涵蓋所有 CRUD 操作和進階查詢
- **事件驅動**: 支援複雜的業務流程和整合
- **向後相容**: 與現有傳統 API 並存
- **測試完備**: 提供全面的測試覆蓋

系統現在同時支援傳統 RESTful API (v1) 和現代 CQRS API (v2)，為不同的使用場景提供了靈活的選擇。