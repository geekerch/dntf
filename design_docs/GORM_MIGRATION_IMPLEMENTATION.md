# GORM Migration 實作設計文件

## 概述

本文件說明將現有的 SQL migration 系統改為使用 GORM AutoMigrate 的實作方案，以支援多種資料庫類型（PostgreSQL、SQLite、SQL Server）。

## 實作內容

### 1. 新增的檔案

#### 1.1 GORM 模型檔案
- `internal/infrastructure/models/channel.go` - Channel 資料表模型
- `internal/infrastructure/models/template.go` - Template 資料表模型  
- `internal/infrastructure/models/message.go` - Message 和 MessageResult 資料表模型
- `internal/infrastructure/models/models.go` - 統一的模型管理檔案

#### 1.2 資料庫連接檔案
- `pkg/database/gorm.go` - 新的 GORM 資料庫連接包裝器

### 2. 修改的檔案

#### 2.1 配置檔案
- `pkg/config/config.go` - 新增 `DB_TYPE` 環境變數支援
- `cmd/server/main.go` - 更新為使用 GORM 資料庫連接

#### 2.2 Repository 檔案（部分完成）
- `internal/infrastructure/repository/channel_repository_impl.go` - 完全重寫為 GORM 版本
- `internal/infrastructure/repository/template_repository_impl.go` - 部分更新
- `internal/infrastructure/repository/message_repository_impl.go` - 部分更新

### 3. 主要功能

#### 3.1 多資料庫支援
```go
// 支援的資料庫類型
- PostgreSQL (postgres/postgresql)
- SQLite (sqlite)  
- SQL Server (sqlserver/mssql)
```

#### 3.2 GORM 模型特性
- 自動 migration 支援
- 跨資料庫相容性
- 自定義 JSON 和 StringArray 類型
- 軟刪除支援
- 索引自動創建

#### 3.3 環境變數配置
```bash
DB_TYPE=postgres          # 資料庫類型
DB_HOST=localhost         # 資料庫主機
DB_PORT=5432             # 資料庫端口
DB_USER=postgres         # 資料庫用戶
DB_PASSWORD=password     # 資料庫密碼
DB_NAME=channel_api      # 資料庫名稱
DB_SSL_MODE=disable      # SSL 模式（PostgreSQL）
```

### 4. 實作狀態

#### 4.1 已完成
✅ GORM 模型定義  
✅ 多資料庫驅動支援  
✅ 配置檔案更新  
✅ Channel Repository GORM 實作  
✅ 主程式更新  

#### 4.2 待完成
🔄 Template Repository GORM 實作完成  
🔄 Message Repository GORM 實作完成  
🔄 編譯錯誤修正  
🔄 測試驗證  

### 5. 下一步工作

1. **完成 Repository 實作**
   - 完成 Template Repository 的 GORM 實作
   - 完成 Message Repository 的 GORM 實作
   - 修正所有編譯錯誤

2. **測試驗證**
   - PostgreSQL 連接測試
   - SQLite 連接測試
   - SQL Server 連接測試
   - Migration 功能測試

3. **文件更新**
   - 更新 .env.example 檔案
   - 更新部署文件
   - 更新開發環境設定指南

### 6. 技術優勢

#### 6.1 相較於 SQL Migration 的優勢
- **跨資料庫相容性**: 自動處理不同資料庫的 SQL 語法差異
- **自動 Schema 管理**: GORM AutoMigrate 自動處理資料表結構變更
- **類型安全**: Go 結構體定義確保類型安全
- **維護簡化**: 不需要維護多套 SQL migration 檔案

#### 6.2 開發效率提升
- **快速切換資料庫**: 只需修改環境變數即可切換資料庫類型
- **開發環境簡化**: SQLite 支援讓本地開發更簡單
- **生產環境彈性**: 可根據需求選擇最適合的資料庫

### 7. 注意事項

#### 7.1 資料庫特定功能
- PostgreSQL 的 GIN 索引需要特殊處理
- SQLite 的某些功能限制需要考慮
- SQL Server 的語法差異需要適配

#### 7.2 Migration 策略
- 現有資料的遷移計畫
- 版本控制策略
- 回滾機制

## 結論

GORM Migration 實作提供了更好的跨資料庫支援和維護性，雖然需要重寫部分 Repository 程式碼，但長期來看將大大提升系統的彈性和可維護性。