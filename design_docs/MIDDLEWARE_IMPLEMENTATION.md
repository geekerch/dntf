# 中介軟體實作文件

## 概述

本文件描述 Channel API 的中介軟體（Middleware）實作，包含認證、限流、CORS、安全性等企業級功能的完整實作。

## 架構設計

中介軟體遵循責任鏈模式（Chain of Responsibility），每個中介軟體負責特定的功能，並可以組合使用。

```
internal/presentation/http/middleware/
├── auth.go                    # 認證中介軟體
├── rate_limiter.go           # 限流中介軟體
├── cors.go                   # CORS 中介軟體
├── security.go               # 安全性中介軟體
├── error_handler.go          # 錯誤處理中介軟體
├── request_logger.go         # 請求日誌中介軟體
├── response_formatter.go     # 回應格式化中介軟體
└── middleware_manager.go     # 中介軟體管理器
```

## 中介軟體功能

### 1. 認證中介軟體 (Authentication)

#### 支援的認證方式

| 認證類型 | 說明 | 使用場景 |
|---------|------|----------|
| API Key | 使用 X-API-Key 標頭或 Bearer token | API 存取控制 |
| JWT | JSON Web Token 認證 | 使用者會話管理 |
| Basic Auth | HTTP 基本認證 | 簡單的使用者認證 |

#### 配置選項

```go
type AuthConfig struct {
    AuthType  string              // "api-key", "jwt", "basic"
    JWTSecret string              // JWT 密鑰
    APIKeys   map[string]string   // API 金鑰對應表
    SkipPaths []string            // 跳過認證的路徑
}
```

#### 特色功能

- **路徑白名單**: 可設定跳過認證的路徑（如 `/health`）
- **多重認證**: 支援多種認證方式並存
- **使用者上下文**: 認證成功後設定使用者資訊到請求上下文
- **詳細日誌**: 記錄認證成功/失敗的詳細資訊

### 2. 限流中介軟體 (Rate Limiting)

#### 限流演算法

使用 **Token Bucket（令牌桶）** 演算法：
- 支援突發流量（Burst）
- 平滑的流量控制
- 自動令牌補充

#### 限流策略

| 策略 | 說明 | 限制對象 |
|------|------|----------|
| IP 限流 | 基於客戶端 IP 地址 | 未認證使用者 |
| 使用者限流 | 基於認證使用者 ID | 已認證使用者 |
| 路徑白名單 | 特定路徑不受限制 | 健康檢查等 |
| IP 白名單 | 特定 IP 不受限制 | 內部服務 |

#### 配置選項

```go
type RateLimiterConfig struct {
    RequestsPerMinute        int      // 每分鐘請求數（IP）
    RequestsPerMinutePerUser int      // 每分鐘請求數（使用者）
    BurstSize               int      // 突發請求數量
    SkipPaths               []string // 跳過限流的路徑
    WhitelistIPs            []string // IP 白名單
}
```

#### 特色功能

- **動態限制**: 認證使用者享有更高的限流額度
- **HTTP 標頭**: 回傳限流狀態資訊
- **記憶體管理**: 自動清理過期的限流記錄
- **可配置**: 支援不同環境的限流策略

### 3. CORS 中介軟體

#### CORS 支援

完整的 CORS（跨來源資源共享）支援：
- **預檢請求**: 正確處理 OPTIONS 請求
- **實際請求**: 設定適當的 CORS 標頭
- **萬用字元支援**: 支援 `*.example.com` 格式
- **私有網路**: 支援私有網路存取控制

#### 配置選項

```go
type CORSConfig struct {
    AllowedOrigins      []string      // 允許的來源
    AllowedMethods      []string      // 允許的 HTTP 方法
    AllowedHeaders      []string      // 允許的標頭
    ExposedHeaders      []string      // 暴露給客戶端的標頭
    AllowCredentials    bool          // 是否允許憑證
    MaxAge              time.Duration // 預檢快取時間
    AllowWildcard       bool          // 是否允許萬用字元
    AllowPrivateNetwork bool          // 是否允許私有網路
}
```

#### 預設配置

| 環境 | 配置特點 |
|------|----------|
| Development | 寬鬆設定，支援本地開發 |
| Production | 嚴格設定，僅允許指定來源 |
| Default | 平衡的安全設定 |

### 4. 安全性中介軟體

#### 安全標頭

自動設定多種安全標頭：

| 標頭 | 功能 | 預設值 |
|------|------|--------|
| Content-Security-Policy | 內容安全政策 | 限制資源載入來源 |
| X-Frame-Options | 防止點擊劫持 | DENY |
| X-Content-Type-Options | 防止 MIME 類型嗅探 | nosniff |
| Referrer-Policy | 控制 Referrer 資訊 | strict-origin-when-cross-origin |
| Strict-Transport-Security | 強制 HTTPS | 僅 HTTPS 時設定 |
| X-XSS-Protection | XSS 保護 | 1; mode=block |

#### 安全檢查

- **主機驗證**: 檢查 Host 標頭是否在允許清單中
- **HTTPS 強制**: 可強制重導向到 HTTPS
- **伺服器標頭**: 可移除或自訂 Server 標頭

#### 配置選項

```go
type SecurityConfig struct {
    ContentSecurityPolicy   string   // CSP 政策
    FrameOptions           string   // X-Frame-Options
    StrictTransportSecurity string   // HSTS 設定
    ForceHTTPS             bool     // 強制 HTTPS
    AllowedHosts           []string // 允許的主機
    RemoveServerHeader     bool     // 移除 Server 標頭
}
```

### 5. 其他中介軟體

#### IP 白名單中介軟體

```go
type IPWhitelistConfig struct {
    AllowedIPs []string // 允許的 IP 地址
    SkipPaths  []string // 跳過檢查的路徑
}
```

#### 基本認證中介軟體

```go
type BasicAuthConfig struct {
    Users     map[string]string // 使用者名稱 -> 密碼
    Realm     string            // 認證領域
    SkipPaths []string          // 跳過認證的路徑
}
```

## 中介軟體管理器

### MiddlewareManager

統一管理所有中介軟體的設定和應用：

```go
type MiddlewareConfig struct {
    Environment       string           // 環境：development, staging, production
    EnableAuth        bool             // 啟用認證
    EnableRateLimit   bool             // 啟用限流
    EnableCORS        bool             // 啟用 CORS
    EnableSecurity    bool             // 啟用安全性
    EnableIPWhitelist bool             // 啟用 IP 白名單
    EnableBasicAuth   bool             // 啟用基本認證
    
    // 各中介軟體的詳細配置
    Auth        *AuthConfig
    RateLimit   *RateLimiterConfig
    CORS        *CORSConfig
    Security    *SecurityConfig
    IPWhitelist *IPWhitelistConfig
    BasicAuth   *BasicAuthConfig
}
```

### 路由分組策略

| 路由分組 | 中介軟體應用 | 用途 |
|----------|-------------|------|
| 全域 | 日誌、錯誤處理、安全性、CORS | 所有請求 |
| 公開路由 | 基本中介軟體 | 健康檢查、資訊查詢 |
| 保護路由 | + 認證、限流 | API 端點 |
| 管理路由 | + 基本認證、IP 白名單 | 管理功能 |

## 環境配置

### 開發環境 (Development)

```go
DevelopmentMiddlewareConfig() *MiddlewareConfig {
    return &MiddlewareConfig{
        Environment:     "development",
        EnableAuth:      false,  // 方便開發測試
        EnableRateLimit: false,  // 不限制請求
        EnableCORS:      true,   // 支援跨域開發
        EnableSecurity:  true,   // 基本安全設定
        CORS:            DevelopmentCORSConfig(),
        Security:        DevelopmentSecurityConfig(),
    }
}
```

### 正式環境 (Production)

```go
ProductionMiddlewareConfig(allowedOrigins, allowedHosts []string) *MiddlewareConfig {
    return &MiddlewareConfig{
        Environment:     "production",
        EnableAuth:      true,   // 強制認證
        EnableRateLimit: true,   // 嚴格限流
        EnableCORS:      true,   // 限制跨域來源
        EnableSecurity:  true,   // 嚴格安全設定
        Auth:            ProductionAuthConfig(),
        RateLimit:       ProductionRateLimitConfig(),
        CORS:            ProductionCORSConfig(allowedOrigins),
        Security:        StrictSecurityConfig(allowedHosts),
    }
}
```

### 測試環境 (Test)

```go
TestMiddlewareConfig() *MiddlewareConfig {
    return &MiddlewareConfig{
        Environment:       "test",
        EnableAuth:        false, // 簡化測試
        EnableRateLimit:   false, // 不影響測試效能
        EnableCORS:        false, // 不需要跨域
        EnableSecurity:    false, // 簡化安全設定
    }
}
```

## 使用方式

### 基本使用

```go
// 使用預設配置
middlewareManager := middleware.NewMiddlewareManager(nil)
middlewareManager.SetupMiddleware(router)

// 設定保護路由
protectedRoutes := router.Group("/api/v1")
middlewareManager.SetupProtectedRoutes(protectedRoutes)

// 設定管理路由
adminRoutes := router.Group("/api/v1/admin")
middlewareManager.SetupAdminRoutes(adminRoutes)
```

### 自訂配置

```go
config := &middleware.MiddlewareConfig{
    Environment:     "production",
    EnableAuth:      true,
    EnableRateLimit: true,
    EnableCORS:      true,
    EnableSecurity:  true,
    Auth: &middleware.AuthConfig{
        AuthType: "api-key",
        APIKeys: map[string]string{
            "your-api-key": "user-id",
        },
    },
    RateLimit: &middleware.RateLimiterConfig{
        RequestsPerMinute: 100,
        BurstSize:        10,
    },
}

middlewareManager := middleware.NewMiddlewareManager(config)
```

## 效能考量

### 記憶體管理

- **限流記錄**: 自動清理過期的 Token Bucket
- **認證快取**: 可實作認證結果快取
- **日誌緩衝**: 使用非同步日誌寫入

### 效能優化

- **路徑匹配**: 使用高效的字串比對
- **標頭設定**: 避免重複設定相同標頭
- **中介軟體順序**: 將輕量級中介軟體放在前面

## 安全考量

### 認證安全

- **API 金鑰**: 使用足夠長度的隨機金鑰
- **JWT**: 使用強密鑰和適當的過期時間
- **基本認證**: 使用常數時間比較防止時序攻擊

### 限流安全

- **分散式限流**: 考慮使用 Redis 實作分散式限流
- **限流繞過**: 防止透過不同 IP 繞過限流
- **DDoS 防護**: 結合其他 DDoS 防護機制

### CORS 安全

- **來源驗證**: 嚴格驗證 Origin 標頭
- **萬用字元**: 謹慎使用萬用字元設定
- **憑證處理**: 正確處理跨域憑證

## 監控和日誌

### 日誌記錄

所有中介軟體都提供詳細的結構化日誌：

```go
logger.Info("Authentication successful",
    zap.String("user_id", userID),
    zap.String("path", c.Request.URL.Path),
    zap.String("method", c.Request.Method))

logger.Warn("Rate limit exceeded",
    zap.String("client_ip", c.ClientIP()),
    zap.Int("remaining", remaining))
```

### 監控指標

建議監控的指標：
- 認證成功/失敗率
- 限流觸發頻率
- CORS 請求統計
- 安全事件計數

## 擴展指南

### 新增自訂中介軟體

1. 建立中介軟體檔案
2. 實作 `gin.HandlerFunc` 介面
3. 在 `MiddlewareManager` 中註冊
4. 更新配置結構

### 整合第三方服務

- **認證服務**: 整合 OAuth2、LDAP 等
- **限流服務**: 整合 Redis、Memcached 等
- **監控服務**: 整合 Prometheus、Grafana 等

## 總結

中介軟體實作提供了：

1. **完整的安全防護**: 認證、授權、安全標頭、CORS
2. **流量控制**: 智慧限流和突發處理
3. **靈活配置**: 支援多種環境和使用場景
4. **高效能**: 優化的演算法和記憶體管理
5. **易於擴展**: 模組化設計便於新增功能
6. **詳細監控**: 完整的日誌和指標支援

這個中介軟體系統為 Channel API 提供了企業級的安全性和可靠性保障。