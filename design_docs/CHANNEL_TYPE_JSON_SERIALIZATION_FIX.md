# Channel Type JSON 序列化修正設計文件

## 問題描述

在建立模板時，發現 `channelType` 欄位無法正確從 JSON 字串轉換為 `shared.ChannelType` 類型。這導致 API 請求失敗，無法正常建立模板。

### 問題根因

1. `shared.ChannelType` 結構體缺少 JSON 序列化和反序列化方法
2. 應用程式啟動時沒有初始化 channel type registry
3. JSON 解析器無法將字串 "email" 轉換為 `ChannelType` 物件

## 解決方案

### 1. 添加 JSON 序列化支援

為 `shared.ChannelType` 添加了 `MarshalJSON` 和 `UnmarshalJSON` 方法：

```go
// MarshalJSON implements json.Marshaler interface for ChannelType
func (ct ChannelType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + ct.name + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler interface for ChannelType
func (ct *ChannelType) UnmarshalJSON(data []byte) error {
	// Remove quotes from JSON string
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("invalid JSON string for ChannelType: %s", string(data))
	}
	
	name := string(data[1 : len(data)-1])
	if name == "" {
		return errors.New("channel type name cannot be empty")
	}
	
	// Create channel type (this will validate against registry)
	channelType, err := NewChannelType(name)
	if err != nil {
		return err
	}
	
	*ct = channelType
	return nil
}
```

### 2. 初始化 Channel Type Registry

在 `cmd/server/main.go` 中添加了 channel types 的初始化：

```go
// Initialize channel types registry
shared.MustInitializeChannelTypes()
log.Info("Channel types initialized successfully")
```

## 修正效果

### 修正前
- JSON 請求中的 `"channelType": "email"` 無法正確解析
- 建立模板時會出現 channel type 轉換錯誤
- API 回應錯誤訊息

### 修正後
- JSON 字串可以正確轉換為 `ChannelType` 物件
- 支援所有已註冊的 channel types：email、slack、sms
- 無效的 channel type 會回傳適當的錯誤訊息
- 模板建立功能正常運作

## 測試驗證

建立了測試腳本驗證以下功能：

1. ✅ JSON 反序列化正常工作
2. ✅ Channel type 驗證正常工作
3. ✅ 支援的 channel types (email, slack, sms) 都能正確處理
4. ✅ 無效的 channel type 會正確回傳錯誤

## 相容性

此修正完全向後相容：

- 現有的 API 介面沒有變更
- 現有的 channel type 定義保持不變
- 擴充性機制依然有效，可以繼續添加新的 channel types

## 相關檔案

- `internal/domain/shared/value_objects.go` - 添加 JSON 序列化方法
- `cmd/server/main.go` - 添加 channel types 初始化
- `internal/domain/shared/init.go` - Channel types 註冊邏輯
- `internal/domain/shared/channel_type_registry.go` - Registry 實作

## 其他檢查結果

### Message 和 Channel 相關檢查

經過全面檢查，發現以下情況：

#### ✅ Channel 相關
- Channel DTO 中的 `channelType` 使用 `string` 類型
- 在 use case 中通過 `shared.NewChannelTypeFromString()` 正確轉換
- JSON 序列化和反序列化正常工作
- 支援所有已註冊的 channel types (email, slack, sms)

#### ✅ Message 相關
- Message DTO 中沒有直接使用 `shared.ChannelType`
- `MessageStatus` 和 `MessageResultStatus` 基於 `string` 類型
- 這些狀態類型的 JSON 序列化自動正常工作
- 所有相關的 JSON 轉換都正常運作

#### ✅ 驗證結果
- 所有 JSON 序列化/反序列化測試通過
- Channel 建立功能正常
- Message 發送功能正常
- 狀態類型轉換正常

### 結論

**只有 Template 存在 channelType JSON 轉換問題**，其他模組 (Channel 和 Message) 都正常工作：

- Template: ❌ → ✅ (已修正)
- Channel: ✅ (正常工作)
- Message: ✅ (正常工作)

## 後續建議

1. 考慮添加更多的 channel type 驗證邏輯
2. 可以考慮添加 channel type 的配置驗證
3. 建議添加更完整的單元測試覆蓋 JSON 序列化功能
4. 考慮為其他自定義類型也添加 JSON 序列化方法以保持一致性