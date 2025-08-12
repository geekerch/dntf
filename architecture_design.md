# Channel API Clean Architecture 設計

## 架構層級 (Architecture Layers)

```
┌─────────────────────────────────────────┐
│              Presentation Layer          │
├─────────────────────────────────────────┤
│               Application Layer          │
├─────────────────────────────────────────┤
│                Domain Layer              │
├─────────────────────────────────────────┤
│              Infrastructure Layer        │
└─────────────────────────────────────────┘
```

## 1. Domain Layer (核心業務邏輯層)

### Entities (實體)
- `Channel`: 通道實體，包含通道的基本屬性和業務規則
- `Template`: 範本實體，包含範本的基本屬性和業務規則
- `Message`: 訊息實體，表示一次發送任務
- `MessageResult`: 訊息發送結果

### Value Objects (值物件)
- `ChannelID`: 通道唯一識別碼
- `TemplateID`: 範本唯一識別碼
- `MessageID`: 訊息唯一識別碼
- `ChannelType`: 通道類型 (email, slack, sms, etc.)
- `ChannelConfig`: 通道配置
- `CommonSettings`: 通用設定
- `Recipients`: 收件人列表
- `Variables`: 範本變數
- `Pagination`: 分頁參數

### Domain Services (領域服務)
- `MessageSender`: 協調跨聚合根的訊息發送邏輯
- `TemplateRenderer`: 範本渲染服務
- `ChannelValidator`: 通道驗證服務

### Repository Interfaces (倉儲介面)
- `ChannelRepository`: 通道資料存取介面
- `TemplateRepository`: 範本資料存取介面
- `MessageRepository`: 訊息資料存取介面

## 2. Application Layer (應用層)

### Use Cases (用例)
**Channel Use Cases:**
- `CreateChannelUseCase`: 建立通道
- `GetChannelUseCase`: 取得單一通道
- `ListChannelsUseCase`: 取得通道列表
- `UpdateChannelUseCase`: 更新通道
- `DeleteChannelUseCase`: 刪除通道

**Template Use Cases:**
- `CreateTemplateUseCase`: 建立範本
- `GetTemplateUseCase`: 取得單一範本
- `ListTemplatesUseCase`: 取得範本列表
- `UpdateTemplateUseCase`: 更新範本
- `DeleteTemplateUseCase`: 刪除範本

**Message Use Cases:**
- `SendMessageUseCase`: 發送訊息

### DTOs (Data Transfer Objects)
- Request/Response DTOs for each use case
- Pagination DTOs

## 3. Infrastructure Layer (基礎設施層)

### Repository Implementations (倉儲實作)
- `ChannelRepositoryImpl`: 通道資料存取實作
- `TemplateRepositoryImpl`: 範本資料存取實作
- `MessageRepositoryImpl`: 訊息資料存取實作

### External Services (外部服務)
- `EmailService`: 電子郵件發送服務
- `SlackService`: Slack 訊息發送服務
- `SMSService`: 簡訊發送服務

### Message Brokers (訊息代理)
- `NATSPublisher`: NATS 訊息發布者
- `NATSSubscriber`: NATS 訊息訂閱者

## 4. Presentation Layer (展示層)

### HTTP Handlers (HTTP 處理器)
- `ChannelHandler`: 處理 Channel RESTful API
- `TemplateHandler`: 處理 Template RESTful API
- `MessageHandler`: 處理 Message RESTful API

### NATS Handlers (NATS 處理器)
- `ChannelNATSHandler`: 處理 Channel NATS 訊息
- `TemplateNATSHandler`: 處理 Template NATS 訊息
- `MessageNATSHandler`: 處理 Message NATS 訊息

### Middleware (中介軟體)
- `ErrorHandler`: 錯誤處理中介軟體
- `RequestLogger`: 請求日誌中介軟體
- `ResponseFormatter`: 回應格式化中介軟體

## 專案目錄結構

```
notification/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── domain/
│   │   ├── channel/
│   │   │   ├── entity.go
│   │   │   ├── repository.go
│   │   │   └── value_objects.go
│   │   ├── template/
│   │   │   ├── entity.go
│   │   │   ├── repository.go
│   │   │   └── value_objects.go
│   │   ├── message/
│   │   │   ├── entity.go
│   │   │   ├── repository.go
│   │   │   └── value_objects.go
│   │   └── shared/
│   │       └── value_objects.go
│   ├── application/
│   │   ├── channel/
│   │   │   ├── usecases/
│   │   │   └── dtos/
│   │   ├── template/
│   │   │   ├── usecases/
│   │   │   └── dtos/
│   │   └── message/
│   │       ├── usecases/
│   │       └── dtos/
│   ├── infrastructure/
│   │   ├── repository/
│   │   ├── external/
│   │   └── messaging/
│   └── presentation/
│       ├── http/
│       │   ├── handlers/
│       │   ├── middleware/
│       │   └── routes/
│       └── nats/
│           └── handlers/
├── pkg/
│   ├── errors/
│   ├── logger/
│   └── config/
└── go.mod
```

## 依賴注入原則

遵循依賴反轉原則：
- Domain Layer 不依賴任何其他層
- Application Layer 只依賴 Domain Layer
- Infrastructure Layer 實作 Domain Layer 的介面
- Presentation Layer 依賴 Application Layer

## 通訊模式

### RESTful API 流程
```
HTTP Request → Handler → Use Case → Domain Service → Repository
```

### NATS 訊息流程
```
NATS Message → NATS Handler → Use Case → Domain Service → Repository
```