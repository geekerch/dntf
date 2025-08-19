# 資料庫遷移路徑配置化實作

## 概述

本次實作將資料庫遷移（migrations）路徑從硬編碼改為可配置的環境變數，解決開發環境與 Docker 環境路徑不同的問題。

## 問題描述

原本的實作中，資料庫遷移路徑在 `pkg/database/gorm.go` 中被硬編碼為 `"file://migrations"`，這導致：

1. 開發環境和 Docker 環境需要不同的路徑設定
2. 無法靈活調整遷移檔案的位置
3. 部署時需要修改程式碼才能適應不同環境

## 解決方案

### 1. 配置結構擴展

在 `pkg/config/config.go` 中的 `DatabaseConfig` 結構體新增 `MigrationsPath` 欄位：

```go
type DatabaseConfig struct {
    // ... 其他欄位
    MigrationsPath string `json:"migrationsPath"`
}
```

### 2. 環境變數配置

新增 `DB_MIGRATIONS_PATH` 環境變數，預設值為 `"migrations"`：

```go
MigrationsPath: getEnv("DB_MIGRATIONS_PATH", "migrations"),
```

### 3. 動態路徑使用

修改 `pkg/database/gorm.go` 中的 `runFileBasedMigrations` 函數，使用配置中的路徑：

```go
// 使用配置中的遷移路徑
migrationsURL := fmt.Sprintf("file://%s", db.config.MigrationsPath)
m, err := migrate.New(migrationsURL, databaseURL)
```

## 環境變數設定

### 開發環境
```bash
DB_MIGRATIONS_PATH=migrations
```

### Docker 環境
```bash
DB_MIGRATIONS_PATH=./migrations
```

### 其他自訂路徑
```bash
DB_MIGRATIONS_PATH=/app/database/migrations
```

## 檔案變更清單

1. `pkg/config/config.go`
   - 新增 `MigrationsPath` 欄位到 `DatabaseConfig`
   - 在配置載入時設定預設值

2. `pkg/database/gorm.go`
   - 修改 `runFileBasedMigrations` 函數使用動態路徑

3. `.env.example`
   - 新增 `DB_MIGRATIONS_PATH` 環境變數說明

## 向後相容性

此變更完全向後相容：
- 預設值 `"migrations"` 與原本硬編碼的路徑相同
- 現有的部署不需要修改即可正常運作
- 只有需要自訂路徑的環境才需要設定新的環境變數

## 測試驗證

- ✅ 程式碼編譯成功
- ✅ 保持原有功能不變
- ✅ 支援環境變數配置

## 未來改進建議

1. 可考慮支援絕對路徑和相對路徑的自動判斷
2. 可新增路徑存在性驗證
3. 可支援多個遷移路徑來源