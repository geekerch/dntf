# Notification API Docker 部署總結

## 🎉 部署成功！

應用已成功部署在Docker容器中，並連接到本機的PostgreSQL和NATS服務。

## 📋 部署配置

### Docker配置
- **容器名稱**: `notification-api`
- **端口映射**: `8080:8080`
- **重啟策略**: `unless-stopped`
- **健康檢查**: 每30秒檢查一次

### 數據庫配置
- **類型**: PostgreSQL
- **主機**: `host.docker.internal` (訪問宿主機)
- **端口**: `5432`
- **用戶**: `admin`
- **數據庫**: `admin`
- **Schema**: `notification`

### NATS配置
- **URL**: `nats://host.docker.internal:4222`
- **Credentials**: 外部掛載 `/app/creds/edgesync_shadowagent.creds`
- **Subject前綴**: `eco1j.infra.eventcenter`

## 🌐 API訪問

- **API基礎URL**: http://localhost:8080
- **API文檔**: http://localhost:8080/swagger/index.html (如果有配置)
- **健康檢查**: http://localhost:8080/health (需要檢查路由配置)

### 測試API
```bash
# 獲取頻道列表
curl http://localhost:8080/api/v1/channels

# 檢查應用狀態
docker-compose -f docker-compose.deploy.yml ps
```

## 🛠️ 管理命令

### 部署相關
```bash
# 快速部署
./quick-deploy.sh

# 完整部署（包含健康檢查）
./deploy.sh

# 停止服務
docker-compose -f docker-compose.deploy.yml down

# 重啟服務
docker-compose -f docker-compose.deploy.yml restart

# 查看日誌
docker-compose -f docker-compose.deploy.yml logs -f
```

### 維護命令
```bash
# 查看容器狀態
docker-compose -f docker-compose.deploy.yml ps

# 進入容器
docker exec -it notification-api sh

# 查看資源使用
docker stats notification-api
```

## 📁 重要文件

- `docker-compose.deploy.yml` - 部署配置文件
- `.env` - 環境變數配置
- `cmd/server/edgesync_shadowagent.creds` - NATS認證文件
- `deploy.sh` - 完整部署腳本
- `quick-deploy.sh` - 快速部署腳本

## 🔧 故障排除

### 常見問題

1. **數據庫連接失敗**
   - 檢查PostgreSQL是否運行: `ps aux | grep postgres`
   - 測試連接: `PGPASSWORD=admin psql -h localhost -U admin -d admin -c "SELECT version();"`

2. **NATS連接失敗**
   - 檢查NATS服務: `netstat -tlnp | grep :4222`
   - 檢查credentials文件: `ls -la cmd/server/edgesync_shadowagent.creds`

3. **容器無法啟動**
   - 查看詳細日誌: `docker-compose -f docker-compose.deploy.yml logs`
   - 檢查端口占用: `netstat -tlnp | grep :8080`

### 日誌查看
```bash
# 實時查看日誌
docker-compose -f docker-compose.deploy.yml logs -f

# 查看最近的日誌
docker-compose -f docker-compose.deploy.yml logs --tail=50
```

## ✅ 驗證部署

應用已成功部署，可以通過以下方式驗證：

1. **容器狀態**: ✅ 運行中
2. **數據庫連接**: ✅ 成功連接到PostgreSQL
3. **NATS連接**: ✅ 使用credentials文件連接
4. **API響應**: ✅ 可以正常返回數據

---

部署時間: $(date)
部署狀態: 🟢 成功