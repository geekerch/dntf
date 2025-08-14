---
title: Channel API 設計文件

---

# Channel & Template API 設計文件

## API 格式統一原則

### Request 格式

- **NATS**: 將 RESTful 的請求內容放入 `data` 欄位中
- **RESTful**: 直接使用 Request Body、路徑參數 (Path Parameter) 或查詢參數 (Query Parameter)

### 成功回應格式

**NATS:**
```json
{
  "reqSeqId": "...",
  "rspSeqId": "...",
  "httpStatus": 200,
  "data": { ... },
  "error": null
}
```

**RESTful:**
```json
{
  "data": { ... },
  "error": null
}
```

### 錯誤回應格式

**NATS:**
```json
{
  "reqSeqId": "...",
  "rspSeqId": "...",
  "httpStatus": 4xx,
  "data": null,
  "error": {
    "code": "ERROR_CODE",
    "message": "Error description"
  }
}
```

**RESTful:**
```json
{
  "data": null,
  "error": {
    "code": "ERROR_CODE", 
    "message": "Error description"
  }
}
```

## HTTP 狀態碼說明

| HTTP Code | 說明 | 使用場景 |
|-----------|------|----------|
| 200 | 成功 | 取得、更新或刪除成功 |
| 201 | 建立成功 | 建立新資源成功 |
| 400 | 請求錯誤 | 參數驗證失敗、格式錯誤、業務規則違反 |
| 404 | 資源不存在 | 指定的資源 ID 不存在 |
| 409 | 衝突 | 資源已存在 (如名稱或 ID 重複) |
| 500 | 服務器內部錯誤 | 系統異常、資料庫錯誤等 |

## 分頁規範 (Pagination)

### 分頁請求參數

| 欄位 | 類型 | 必填 | 說明 |
|------|------|------|------|
| skipCount | number | 選填 | 跳過的記錄數，從 0 開始，預設：0 |
| maxResultCount | number | 選填 | 每頁最大結果數，範圍：1-100，預設：10 |

### 分頁回應欄位

| 欄位 | 類型 | 說明 |
|------|------|------|
| items | array | 本次回應的資料列表 |
| skipCount | number | 本次回應的分頁起始點 |
| maxResultCount | number | 回應中每頁最大筆數 |
| totalCount | number | 符合條件的資料總數 |
| hasMore | boolean | 是否還有更多頁的資料 |

> **注意**: 所有時間欄位 (createdAt, updatedAt, lastUsed, deletedAt, sentAt) 均為 Unix timestamp，單位為毫秒 (milliseconds)

---

# Channel API

## 1. 建立 Channel (Create Channel)

**NATS Topic**: `eco1j.infra.eventcenter.channel.create`  
**RESTful API**: `POST /api/v1/channels`

### Request 格式

**Request Body:**
```json
{
  "channelName": "testEmail",
  "description": "channel description",
  "enabled": true,
  "channelType": "email",
  "templateId": "tpl_001",
  "commonSettings": {
    "timeout": 5000,
    "retryAttempts": 3,
    "retryDelay": 1000
  },
  "config": {
    "host": "smtp.example.com",
    "port": 587,
    "username": "user",
    "password": "password"
  },
  "recipients": [
    { "name": "John Doe", "email": "john.doe@example.com", "type": "to" }
  ],
  "tags": ["tag1", "tag2"]
}
```

### Response 格式

**NATS Response:**
```json
{
  "reqSeqId": "...",
  "rspSeqId": "...",
  "httpStatus": 201,
  "data": {
    "channelId": "channel_123",
    "channelName": "testEmail",
    "description": "channel description",
    "enabled": true,
    "channelType": "email",
    "templateId": "tpl_001",
    "commonSettings": { ... },
    "config": { ... },
    "recipients": [ ... ],
    "tags": ["tag1", "tag2"],
    "createdAt": 1754379199000,
    "updatedAt": 1754379199000,
    "lastUsed": null
  },
  "error": null
}
```

**RESTful Response:**
```json
{
  "data": {
    "channelId": "channel_123",
    "channelName": "testEmail",
    "description": "channel description",
    "enabled": true,
    "channelType": "email",
    "templateId": "tpl_001",
    "commonSettings": { ... },
    "config": { ... },
    "recipients": [ ... ],
    "tags": ["tag1", "tag2"],
    "createdAt": 1754379199000,
    "updatedAt": 1754379199000,
    "lastUsed": null
  },
  "error": null
}
```

## 2. 取得全部 Channels (List Channels)

**NATS Topic**: `eco1j.infra.eventcenter.channel.list`  
**RESTful API**: `GET /api/v1/channels`

### Request 格式

| 參數 | NATS Request 格式 (JSON) | RESTful Request 格式 (Query Parameter) | 說明 |
|------|--------------------------|---------------------------------------|------|
| channelType | `"channelType": "email"` | `?channelType=email` | 按通道類型過濾 |
| tags | `"tags": ["tag1", "tag2"]` | `?tags=tag1,tag2` | 按標籤過濾，可篩選包含任一指定標籤的通道 |
| skipCount | `"skipCount": 0` | `?skipCount=0` | 跳過的記錄數 |
| maxResultCount | `"maxResultCount": 10` | `?maxResultCount=10` | 每頁最大結果數 |

### Response 格式

**NATS Response:**
```json
{
  "reqSeqId": "...",
  "rspSeqId": "...",
  "httpStatus": 200,
  "data": {
    "items": [
      {
        "channelId": "channel_123",
        "channelName": "testEmail",
        "channelType": "email",
        "tags": ["tag1", "tag2"],
        "enabled": true,
        "createdAt": 1754379199000,
        "updatedAt": 1754379199000
      }
    ],
    "skipCount": 0,
    "maxResultCount": 10,
    "totalCount": 2,
    "hasMore": false
  },
  "error": null
}
```

**RESTful Response:**
```json
{
  "data": {
    "items": [
      {
        "channelId": "channel_123",
        "channelName": "testEmail",
        "channelType": "email",
        "tags": ["tag1", "tag2"],
        "enabled": true,
        "createdAt": 1754379199000,
        "updatedAt": 1754379199000
      }
    ],
    "skipCount": 0,
    "maxResultCount": 10,
    "totalCount": 2,
    "hasMore": false
  },
  "error": null
}
```

## 3. 取得 Channel (Get Channel)

**NATS Topic**: `eco1j.infra.eventcenter.channel.get`  
**RESTful API**: `GET /api/v1/channels/{channelId}`

### Request 格式

| 參數 | NATS Request 格式 (JSON) | RESTful Request 格式 (Path Parameter) | 說明 |
|------|--------------------------|--------------------------------------|------|
| channelId | `"channelId": "channel_123"` | `/api/v1/channels/channel_123` | 通道唯一識別 ID |

### Response 格式

**NATS Response:**
```json
{
  "reqSeqId": "...",
  "rspSeqId": "...",
  "httpStatus": 200,
  "data": {
    "channelId": "channel_123",
    "channelName": "testEmail",
    "description": "channel description",
    "enabled": true,
    "channelType": "email",
    "templateId": "tpl_001",
    "commonSettings": { ... },
    "config": { ... },
    "recipients": [ ... ],
    "tags": ["tag1", "tag2"],
    "createdAt": 1754379199000,
    "updatedAt": 1754379199000,
    "lastUsed": 1754379199000
  },
  "error": null
}
```

**RESTful Response:**
```json
{
  "data": {
    "channelId": "channel_123",
    "channelName": "testEmail",
    "description": "channel description",
    "enabled": true,
    "channelType": "email",
    "templateId": "tpl_001",
    "commonSettings": { ... },
    "config": { ... },
    "recipients": [ ... ],
    "tags": ["tag1", "tag2"],
    "createdAt": 1754379199000,
    "updatedAt": 1754379199000,
    "lastUsed": 1754379199000
  },
  "error": null
}
```

## 4. 更新 Channel (Update Channel)

**NATS Topic**: `eco1j.infra.eventcenter.channel.update`  
**RESTful API**: `PUT /api/v1/channels/{channelId}`

### Request 格式

**Request Body:**
```json
{
  "channelId": "channel_123",
  "channelName": "updatedEmail",
  "description": "updated description",
  "enabled": false,
  "channelType": "email",
  "templateId": "tpl_002",
  "commonSettings": { ... },
  "config": { ... },
  "recipients": [ ... ],
  "tags": ["tagA", "tagB"]
}
```

> **注意**: NATS 格式需在 Request Body 中包含 `channelId`，RESTful 格式的 `channelId` 透過路徑參數傳遞。

### Response 格式

**NATS Response:**
```json
{
  "reqSeqId": "...",
  "rspSeqId": "...",
  "httpStatus": 200,
  "data": {
    "channelId": "channel_123",
    "channelName": "updatedEmail",
    "description": "updated description",
    "enabled": false,
    "channelType": "email",
    "templateId": "tpl_002",
    "commonSettings": { ... },
    "config": { ... },
    "recipients": [ ... ],
    "tags": ["tagA", "tagB"],
    "createdAt": 1754379199000,
    "updatedAt": 1754379299000,
    "lastUsed": 1754379199000
  },
  "error": null
}
```

**RESTful Response:**
```json
{
  "data": {
    "channelId": "channel_123",
    "channelName": "updatedEmail",
    "description": "updated description",
    "enabled": false,
    "channelType": "email",
    "templateId": "tpl_002",
    "commonSettings": { ... },
    "config": { ... },
    "recipients": [ ... ],
    "tags": ["tagA", "tagB"],
    "createdAt": 1754379199000,
    "updatedAt": 1754379299000,
    "lastUsed": 1754379199000
  },
  "error": null
}
```

## 5. 刪除 Channel (Delete Channel)

**NATS Topic**: `eco1j.infra.eventcenter.channel.delete`  
**RESTful API**: `DELETE /api/v1/channels/{channelId}`

### Request 格式

| 參數 | NATS Request 格式 (JSON) | RESTful Request 格式 (Path Parameter) | 說明 |
|------|--------------------------|--------------------------------------|------|
| channelId | `"channelId": "channel_123"` | `/api/v1/channels/channel_123` | 通道唯一識別 ID |

### Response 格式

**NATS Response:**
```json
{
  "reqSeqId": "...",
  "rspSeqId": "...",
  "httpStatus": 200,
  "data": {
    "channelId": "channel_123",
    "deleted": true,
    "deletedAt": 1754379299000
  },
  "error": null
}
```

**RESTful Response:**
```json
{
  "data": {
    "channelId": "channel_123",
    "deleted": true,
    "deletedAt": 1754379299000
  },
  "error": null
}
```

---

# Template API

## 1. 建立 Template (Create Template)

**NATS Topic**: `eco1j.infra.eventcenter.template.create`  
**RESTful API**: `POST /api/v1/templates`

### Request 格式

**Request Body:**
```json
{
  "templateName": "template name",
  "description": "template description",
  "channelType": "email",
  "subject": "subject template with {variables}",
  "template": "message template with {variables} OR JSON structure",
  "tags": ["tag1", "tag2"]
}
```

> **注意**: `templateId` 由系統自動生成，不需在建立請求中提供。

### Response 格式

**NATS Response:**
```json
{
  "reqSeqId": "...",
  "rspSeqId": "...",
  "httpStatus": 201,
  "data": {
    "templateId": "tpl_001",
    "templateName": "template name",
    "description": "template description",
    "channelType": "email",
    "subject": "subject template with {variables}",
    "template": "message template with {variables}",
    "tags": ["tag1", "tag2"],
    "createdAt": 1754379199000,
    "updatedAt": 1754379199000,
    "version": 1
  },
  "error": null
}
```

**RESTful Response:**
```json
{
  "data": {
    "templateId": "tpl_001",
    "templateName": "template name",
    "description": "template description",
    "channelType": "email",
    "subject": "subject template with {variables}",
    "template": "message template with {variables}",
    "tags": ["tag1", "tag2"],
    "createdAt": 1754379199000,
    "updatedAt": 1754379199000,
    "version": 1
  },
  "error": null
}
```

## 2. 取得全部 Templates (List Templates)

**NATS Topic**: `eco1j.infra.eventcenter.template.list`  
**RESTful API**: `GET /api/v1/templates`

### Request 格式

| 參數 | NATS Request 格式 (JSON) | RESTful Request 格式 (Query Parameter) | 說明 |
|------|--------------------------|---------------------------------------|------|
| channelType | `"channelType": "email"` | `?channelType=email` | 按通道類型過濾 |
| tags | `"tags": ["tag1", "tag2"]` | `?tags=tag1,tag2` | 按標籤過濾，可篩選包含任一指定標籤的範本 |
| skipCount | `"skipCount": 0` | `?skipCount=0` | 跳過的記錄數 |
| maxResultCount | `"maxResultCount": 10` | `?maxResultCount=10` | 每頁最大結果數 |

### Response 格式

**NATS Response:**
```json
{
  "reqSeqId": "...",
  "rspSeqId": "...",
  "httpStatus": 200,
  "data": {
    "items": [
      {
        "templateId": "tpl_001",
        "templateName": "template name",
        "description": "template description",
        "channelType": "email",
        "subject": "subject template with {variables}",
        "template": "message template with {variables}",
        "tags": ["tag1", "tag2"],
        "createdAt": 1754379199000,
        "updatedAt": 1754379199000,
        "version": 1
      }
    ],
    "skipCount": 0,
    "maxResultCount": 10,
    "totalCount": 2,
    "hasMore": false
  },
  "error": null
}
```

**RESTful Response:**
```json
{
  "data": {
    "items": [
      {
        "templateId": "tpl_001",
        "templateName": "template name",
        "description": "template description",
        "channelType": "email",
        "subject": "subject template with {variables}",
        "template": "message template with {variables}",
        "tags": ["tag1", "tag2"],
        "createdAt": 1754379199000,
        "updatedAt": 1754379199000,
        "version": 1
      }
    ],
    "skipCount": 0,
    "maxResultCount": 10,
    "totalCount": 2,
    "hasMore": false
  },
  "error": null
}
```

## 3. 取得 Template (Get Template)

**NATS Topic**: `eco1j.infra.eventcenter.template.get`  
**RESTful API**: `GET /api/v1/templates/{templateId}`

### Request 格式

| 參數 | NATS Request 格式 (JSON) | RESTful Request 格式 (Path Parameter) | 說明 |
|------|--------------------------|--------------------------------------|------|
| templateId | `"templateId": "tpl_001"` | `/api/v1/templates/tpl_001` | 範本唯一識別 ID |

### Response 格式

**NATS Response:**
```json
{
  "reqSeqId": "...",
  "rspSeqId": "...",
  "httpStatus": 200,
  "data": {
    "templateId": "tpl_001",
    "templateName": "template name",
    "description": "template description",
    "channelType": "email",
    "subject": "subject template with {variables}",
    "template": "message template with {variables}",
    "tags": ["tag1", "tag2"],
    "createdAt": 1754379199000,
    "updatedAt": 1754379199000,
    "version": 1
  },
  "error": null
}
```

**RESTful Response:**
```json
{
  "data": {
    "templateId": "tpl_001",
    "templateName": "template name",
    "description": "template description",
    "channelType": "email",
    "subject": "subject template with {variables}",
    "template": "message template with {variables}",
    "tags": ["tag1", "tag2"],
    "createdAt": 1754379199000,
    "updatedAt": 1754379199000,
    "version": 1
  },
  "error": null
}
```

## 4. 更新 Template (Update Template)

**NATS Topic**: `eco1j.infra.eventcenter.template.update`  
**RESTful API**: `PUT /api/v1/templates/{templateId}`

### Request 格式

**Request Body:**
```json
{
  "templateId": "tpl_001",
  "templateName": "updated template name",
  "description": "updated description",
  "channelType": "email",
  "subject": "updated subject with {variables}",
  "template": "updated message template with {variables}",
  "tags": ["tagA", "tagB"]
}
```

> **注意**: NATS 格式需在 Request Body 中包含 `templateId`，RESTful 格式的 `templateId` 透過路徑參數傳遞。

### Response 格式

**NATS Response:**
```json
{
  "reqSeqId": "...",
  "rspSeqId": "...",
  "httpStatus": 200,
  "data": {
    "templateId": "tpl_001",
    "templateName": "updated template name",
    "description": "updated description",
    "channelType": "email",
    "subject": "updated subject with {variables}",
    "template": "updated message template with {variables}",
    "tags": ["tagA", "tagB"],
    "createdAt": 1754379199000,
    "updatedAt": 1754379299000,
    "version": 2
  },
  "error": null
}
```

**RESTful Response:**
```json
{
  "data": {
    "templateId": "tpl_001",
    "templateName": "updated template name",
    "description": "updated description",
    "channelType": "email",
    "subject": "updated subject with {variables}",
    "template": "updated message template with {variables}",
    "tags": ["tagA", "tagB"],
    "createdAt": 1754379199000,
    "updatedAt": 1754379299000,
    "version": 2
  },
  "error": null
}
```

## 5. 刪除 Template (Delete Template)

**NATS Topic**: `eco1j.infra.eventcenter.template.delete`  
**RESTful API**: `DELETE /api/v1/templates/{templateId}`

### Request 格式

| 參數 | NATS Request 格式 (JSON) | RESTful Request 格式 (Path Parameter) | 說明 |
|------|--------------------------|--------------------------------------|------|
| templateId | `"templateId": "tpl_001"` | `/api/v1/templates/tpl_001` | 範本唯一識別 ID |

### Response 格式

**NATS Response:**
```json
{
  "reqSeqId": "...",
  "rspSeqId": "...",
  "httpStatus": 200,
  "data": {
    "templateId": "tpl_001",
    "deleted": true,
    "deletedAt": 1754379299000
  },
  "error": null
}
```

**RESTful Response:**
```json
{
  "data": {
    "templateId": "tpl_001",
    "deleted": true,
    "deletedAt": 1754379299000
  },
  "error": null
}
```

---

# Message API

## 1. 發送訊息 (Send Message)

**NATS Topic**: `eco1j.infra.eventcenter.message.send`  
**RESTful API**: `POST /api/v1/messages`

### Request 格式

**Request Body:**
```json
{
  "channelIds": ["ch_email_001", "ch_slack_001"],
  "variables": {
    "name": "John Doe",
    "title": "System Maintenance",
    "message": "Scheduled maintenance at 2AM"
  },
  "channelOverrides": {
    "ch_email_001": {
      "recipients": [
        {
          "name": "IT Team",
          "email": "it@company.com",
          "type": "to"
        }
      ],
      "templateOverride": {
        "subject": "重要通知: {title}",
        "template": "嗨 {name}, 系統將於 {timestamp} 進行維護。"
      }
    },
    "ch_slack_001": {
      "recipients": [
        {
          "name": "General Channel",
          "target": "#general",
          "type": "channel"
        }
      ],
      "settingsOverride": {
        "retryAttempts": 5
      }
    }
  }
}
```

### Request 欄位說明

| 欄位 | 類型 | 必填 | 說明 |
|------|------|------|------|
| channelIds | array<string> | 必填 | 要發送的通道 ID 列表 |
| variables | object | 必填 | 動態變數，用來填充模板中的 {} 佔位符 |
| channelOverrides | object | 選填 | 針對特定通道的覆寫設定。鍵為 channelId，值為覆寫物件 |

### Response 格式

**NATS Response:**
```json
{
  "reqSeqId": "...",
  "rspSeqId": "...",
  "httpStatus": 200,
  "data": {
    "messageId": "msg_001_1234567890",
    "status": "partial_success",
    "results": [
      {
        "channelId": "ch_email_001",
        "status": "success",
        "message": "Email sent successfully.",
        "sentAt": 1754379299000
      },
      {
        "channelId": "ch_slack_001",
        "status": "failed",
        "message": "Slack API returned an error.",
        "error": {
          "code": "SLACK_API_ERROR",
          "details": "Invalid channel ID or authentication token."
        }
      }
    ]
  },
  "error": null
}
```

**RESTful Response:**
```json
{
  "data": {
    "messageId": "msg_001_1234567890",
    "status": "partial_success",
    "results": [
      {
        "channelId": "ch_email_001",
        "status": "success",
        "message": "Email sent successfully.",
        "sentAt": 1754379299000
      },
      {
        "channelId": "ch_slack_001",
        "status": "failed",
        "message": "Slack API returned an error.",
        "error": {
          "code": "SLACK_API_ERROR",
          "details": "Invalid channel ID or authentication token."
        }
      }
    ]
  },
  "error": null
}
```