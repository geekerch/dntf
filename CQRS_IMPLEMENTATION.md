# CQRS 實作文件

## 概述

本文件描述 Channel API 的 CQRS（Command Query Responsibility Segregation）實作，將讀取和寫入操作分離，提供更好的效能、可擴展性和維護性。

## CQRS 架構設計

CQRS 模式將應用程式的讀取（Query）和寫入（Command）操作分離，每個操作都有專門的模型和處理器。

```
┌─────────────────────────────────────────────────────────────┐
│                    Presentation Layer                       │
├─────────────────────────────────────────────────────────────┤
│                     CQRS Facade                            │
├─────────────────┬─────────────────┬─────────────────────────┤
│   Command Bus   │    Query Bus    │      Event Bus          │
├─────────────────┼─────────────────┼─────────────────────────┤
│ Command Handler │ Query Handler   │   Event Handler         │
├─────────────────┼─────────────────┼─────────────────────────┤
│   Use Cases     │   Use Cases     │   Event Projections     │
├─────────────────┴─────────────────┴─────────────────────────┤
│                    Domain Layer                             │
└─────────────────────────────────────────────────────────────┘
```

## 核心組件

### 1. Command（命令）

Commands 代表改變系統狀態的操作，遵循 CQS（Command Query Separation）原則。

#### 基礎命令介面

```go
type Command interface {
    GetCommandID() string
    GetCommandType() string
    GetTimestamp() time.Time
    Validate() error
}
```

#### Channel Commands

| Command | 功能 | 觸發事件 |
|---------|------|----------|
| `CreateChannelCommand` | 建立通道 | `ChannelCreatedEvent` |
| `UpdateChannelCommand` | 更新通道 | `ChannelUpdatedEvent` |
| `DeleteChannelCommand` | 刪除通道 | `ChannelDeletedEvent` |

### 2. Query（查詢）

Queries 代表讀取系統狀態的操作，不會改變系統狀態。

#### 基礎查詢介面

```go
type Query interface {
    GetQueryID() string
    GetQueryType() string
    GetTimestamp() time.Time
    Validate() error
}
```

#### Channel Queries

| Query | 功能 | 回傳資料 |
|-------|------|----------|
| `GetChannelQuery` | 取得單一通道 | `ChannelResponse` |
| `ListChannelsQuery` | 查詢通道列表 | `ListChannelsResponse` |

#### 查詢選項

```go
type QueryOptions struct {
    Pagination *Pagination  // 分頁參數
    Sorting    []Sorting    // 排序參數
    Filtering  []Filtering  // 篩選參數
    Fields     []string     // 欄位選擇
    Include    []string     // 關聯資料
}
```

### 3. Event（事件）

Events 代表系統中發生的重要事件，用於事件溯源和系統間通訊。

#### 基礎事件介面

```go
type Event interface {
    GetEventID() string
    GetEventType() string
    GetAggregateID() string
    GetAggregateType() string
    GetTimestamp() time.Time
    GetVersion() int64
    GetData() interface{}
}
```

#### Channel Events

| Event | 觸發時機 | 事件資料 |
|-------|----------|----------|
| `ChannelCreatedEvent` | 通道建立時 | 完整通道資訊 |
| `ChannelUpdatedEvent` | 通道更新時 | 更新後資訊 + 變更記錄 |
| `ChannelDeletedEvent` | 通道刪除時 | 通道 ID + 刪除時間 |
| `ChannelEnabledEvent` | 通道啟用時 | 通道 ID + 啟用時間 |
| `ChannelDisabledEvent` | 通道停用時 | 通道 ID + 停用時間 |

## Bus 系統

### 1. Command Bus

負責分發和執行 Commands。

#### 特色功能

- **命令驗證**: 執行前驗證命令格式和業務規則
- **處理器路由**: 自動路由到對應的命令處理器
- **結果追蹤**: 記錄執行結果和效能指標
- **錯誤處理**: 統一的錯誤處理和回報機制

#### 使用範例

```go
// 建立命令
command := channelcqrs.NewCreateChannelCommand(&request)

// 執行命令
result, err := commandBus.Execute(ctx, command)
if err != nil {
    // 處理錯誤
}

// 檢查結果
if result.Success {
    // 處理成功結果
    data := result.Data
    events := result.Events
}
```

### 2. Query Bus

負責分發和執行 Queries。

#### 特色功能

- **查詢最佳化**: 支援快取和查詢最佳化
- **分頁支援**: 內建分頁、排序、篩選功能
- **效能監控**: 追蹤查詢執行時間和快取命中率
- **欄位選擇**: 支援部分欄位查詢以提升效能

#### 使用範例

```go
// 建立查詢
query := channelcqrs.NewListChannelsQuery().
    WithChannelType("email").
    WithPagination(0, 20).
    WithSorting("createdAt", "desc")

// 執行查詢
result, err := queryBus.Execute(ctx, query)
if err != nil {
    // 處理錯誤
}

// 檢查結果
if result.Success {
    data := result.Data
    cacheHit := result.CacheHit
}
```

### 3. Event Bus

負責發布和訂閱 Events。

#### 特色功能

- **非同步處理**: 事件處理不阻塞主要業務流程
- **多訂閱者**: 支援多個處理器訂閱同一事件
- **錯誤隔離**: 單一處理器失敗不影響其他處理器
- **事件重播**: 支援事件重播和錯誤恢復

#### 使用範例

```go
// 發布事件
event := NewChannelCreatedEvent(channelID, version, eventData)
err := eventBus.Publish(ctx, event)

// 訂閱事件
eventBus.Subscribe("channel.created", eventHandler)
```

## CQRS 管理器

### CQRSManager

統一管理所有 CQRS 組件的核心管理器。

```go
type CQRSManager struct {
    commandBus CommandBus
    queryBus   QueryBus
    eventBus   EventBus
}
```

#### 功能特色

- **統一介面**: 提供統一的 CQRS 操作介面
- **處理器註冊**: 自動註冊和管理所有處理器
- **生命週期管理**: 管理 Bus 的啟動和關閉
- **健康檢查**: 提供系統健康狀態檢查

### CQRSFacade

提供簡化的 CQRS 操作介面。

```go
type CQRSFacade struct {
    manager *CQRSManager
    config  *CQRSConfig
}
```

#### 主要方法

| 方法 | 功能 | 說明 |
|------|------|------|
| `Send(command)` | 執行命令 | 包含日誌和指標收集 |
| `Query(query)` | 執行查詢 | 支援快取和效能監控 |
| `Publish(event)` | 發布事件 | 非同步事件處理 |

## Presentation Layer 整合

### HTTP 處理器

#### CQRSChannelHandler

使用 CQRS 模式的 HTTP 處理器，提供更好的關注點分離。

```go
type CQRSChannelHandler struct {
    cqrsFacade *cqrs.CQRSFacade
}
```

#### 特色功能

- **命令追蹤**: HTTP 標頭包含命令/查詢 ID
- **快取指示**: 查詢回應包含快取狀態
- **效能指標**: 回應標頭包含執行時間
- **錯誤統一**: 統一的錯誤處理和回應格式

#### HTTP 標頭

| 標頭 | 說明 | 範例 |
|------|------|------|
| `X-Command-ID` | 命令唯一識別碼 | `cmd_20231201120000.123456` |
| `X-Query-ID` | 查詢唯一識別碼 | `qry_20231201120000.123456` |
| `X-Cache` | 快取狀態 | `HIT` / `MISS` |
| `X-Duration` | 執行時間 | `150ms` |

### NATS 處理器

#### CQRSChannelNATSHandler

使用 CQRS 模式的 NATS 訊息處理器。

```go
type CQRSChannelNATSHandler struct {
    cqrsFacade *cqrs.CQRSFacade
    natsConn   *nats.Conn
}
```

#### 訊息格式

**NATS 請求格式**:
```json
{
  "requestId": "req_20231201120000.123456",
  "data": {
    // 命令或查詢資料
  },
  "timestamp": 1701421200
}
```

**NATS 回應格式**:
```json
{
  "requestId": "req_20231201120000.123456",
  "success": true,
  "data": {
    // 回應資料
  },
  "error": {
    "code": "ERROR_CODE",
    "message": "Error message",
    "details": "Detailed error information"
  },
  "timestamp": 1701421200
}
```

## 效能優化

### 1. 查詢最佳化

- **讀取模型**: 專門為查詢最佳化的資料模型
- **索引策略**: 針對常用查詢建立索引
- **快取層**: 多層快取策略提升查詢效能
- **分頁最佳化**: 高效的分頁實作

### 2. 命令最佳化

- **非同步處理**: 命令執行與事件發布分離
- **批次處理**: 支援批次命令執行
- **重試機制**: 自動重試失敗的命令
- **冪等性**: 確保命令的冪等性

### 3. 事件最佳化

- **事件聚合**: 減少事件數量和網路傳輸
- **非同步發布**: 事件發布不阻塞命令執行
- **事件壓縮**: 壓縮事件資料減少儲存空間
- **分區策略**: 事件分區提升處理效能

## 監控和指標

### 1. 命令指標

- 命令執行次數
- 命令執行時間
- 命令成功/失敗率
- 命令佇列長度

### 2. 查詢指標

- 查詢執行次數
- 查詢執行時間
- 快取命中率
- 查詢結果大小

### 3. 事件指標

- 事件發布次數
- 事件處理延遲
- 事件處理成功率
- 事件佇列積壓

## 錯誤處理

### 1. 命令錯誤

- **驗證錯誤**: 命令格式或業務規則驗證失敗
- **執行錯誤**: 命令執行過程中的錯誤
- **並發錯誤**: 並發修改導致的衝突
- **系統錯誤**: 基礎設施或網路錯誤

### 2. 查詢錯誤

- **參數錯誤**: 查詢參數格式錯誤
- **權限錯誤**: 查詢權限不足
- **資源錯誤**: 查詢的資源不存在
- **超時錯誤**: 查詢執行超時

### 3. 事件錯誤

- **發布錯誤**: 事件發布失敗
- **處理錯誤**: 事件處理器執行失敗
- **序列化錯誤**: 事件序列化/反序列化失敗
- **網路錯誤**: 事件傳輸網路錯誤

## 測試策略

### 1. 單元測試

- 命令/查詢驗證測試
- 處理器邏輯測試
- 事件發布測試
- Bus 路由測試

### 2. 整合測試

- 端到端 CQRS 流程測試
- 事件處理整合測試
- 錯誤場景測試
- 效能基準測試

### 3. 負載測試

- 高並發命令執行測試
- 大量查詢效能測試
- 事件處理能力測試
- 系統穩定性測試

## 部署考量

### 1. 可擴展性

- **水平擴展**: 支援多實例部署
- **負載均衡**: 智慧負載分配
- **資源隔離**: 讀寫操作資源隔離
- **彈性伸縮**: 根據負載自動調整

### 2. 可靠性

- **故障轉移**: 自動故障檢測和轉移
- **資料一致性**: 確保最終一致性
- **備份恢復**: 完整的備份和恢復策略
- **監控告警**: 全面的監控和告警機制

## 總結

CQRS 實作為 Channel API 提供了：

1. **清晰的關注點分離**: 讀寫操作完全分離
2. **優異的效能表現**: 針對性的讀寫最佳化
3. **良好的可擴展性**: 支援獨立擴展讀寫能力
4. **完整的事件驅動**: 支援事件溯源和系統整合
5. **統一的操作介面**: 簡化的 CQRS 操作 API
6. **全面的監控支援**: 詳細的效能和錯誤監控

這個 CQRS 實作為系統提供了企業級的架構基礎，支援高效能、高可用的應用場景。