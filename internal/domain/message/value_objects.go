package message

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"channel-api/internal/domain/channel"
	"channel-api/internal/domain/shared"
	"channel-api/internal/domain/template"
)

// MessageID 訊息唯一識別碼
type MessageID struct {
	value string
}

// NewMessageID 建立新的訊息 ID
func NewMessageID() *MessageID {
	timestamp := time.Now().Unix()
	return &MessageID{
		value: fmt.Sprintf("msg_%d_%s", timestamp, uuid.New().String()[:8]),
	}
}

// NewMessageIDFromString 從字串建立訊息 ID
func NewMessageIDFromString(id string) (*MessageID, error) {
	if id == "" {
		return nil, errors.New("message ID cannot be empty")
	}
	return &MessageID{value: id}, nil
}

// String 返回字串表示
func (m *MessageID) String() string {
	return m.value
}

// Equals 比較兩個訊息 ID 是否相等
func (m *MessageID) Equals(other *MessageID) bool {
	if other == nil {
		return false
	}
	return m.value == other.value
}

// MessageStatus 訊息狀態
type MessageStatus string

const (
	MessageStatusSuccess        MessageStatus = "success"
	MessageStatusFailed         MessageStatus = "failed"
	MessageStatusPartialSuccess MessageStatus = "partial_success"
	MessageStatusPending        MessageStatus = "pending"
)

// IsValid 驗證訊息狀態是否有效
func (ms MessageStatus) IsValid() bool {
	switch ms {
	case MessageStatusSuccess, MessageStatusFailed, MessageStatusPartialSuccess, MessageStatusPending:
		return true
	default:
		return false
	}
}

// Variables 範本變數
type Variables struct {
	variables map[string]interface{}
}

// NewVariables 建立範本變數
func NewVariables(variables map[string]interface{}) *Variables {
	if variables == nil {
		variables = make(map[string]interface{})
	}
	return &Variables{variables: variables}
}

// Get 取得變數值
func (v *Variables) Get(key string) (interface{}, bool) {
	value, exists := v.variables[key]
	return value, exists
}

// Set 設定變數值
func (v *Variables) Set(key string, value interface{}) {
	v.variables[key] = value
}

// ToMap 轉換為 map
func (v *Variables) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range v.variables {
		result[k] = v
	}
	return result
}

// Keys 取得所有變數鍵
func (v *Variables) Keys() []string {
	keys := make([]string, 0, len(v.variables))
	for key := range v.variables {
		keys = append(keys, key)
	}
	return keys
}

// HasKey 檢查是否包含指定鍵
func (v *Variables) HasKey(key string) bool {
	_, exists := v.variables[key]
	return exists
}

// ChannelOverride 通道覆寫設定
type ChannelOverride struct {
	Recipients       *channel.Recipients      `json:"recipients,omitempty"`
	TemplateOverride *TemplateOverride        `json:"templateOverride,omitempty"`
	SettingsOverride *shared.CommonSettings   `json:"settingsOverride,omitempty"`
}

// NewChannelOverride 建立通道覆寫設定
func NewChannelOverride() *ChannelOverride {
	return &ChannelOverride{}
}

// WithRecipients 設定收件人覆寫
func (c *ChannelOverride) WithRecipients(recipients *channel.Recipients) *ChannelOverride {
	c.Recipients = recipients
	return c
}

// WithTemplateOverride 設定範本覆寫
func (c *ChannelOverride) WithTemplateOverride(templateOverride *TemplateOverride) *ChannelOverride {
	c.TemplateOverride = templateOverride
	return c
}

// WithSettingsOverride 設定設定覆寫
func (c *ChannelOverride) WithSettingsOverride(settingsOverride *shared.CommonSettings) *ChannelOverride {
	c.SettingsOverride = settingsOverride
	return c
}

// HasRecipientsOverride 檢查是否有收件人覆寫
func (c *ChannelOverride) HasRecipientsOverride() bool {
	return c.Recipients != nil
}

// HasTemplateOverride 檢查是否有範本覆寫
func (c *ChannelOverride) HasTemplateOverride() bool {
	return c.TemplateOverride != nil
}

// HasSettingsOverride 檢查是否有設定覆寫
func (c *ChannelOverride) HasSettingsOverride() bool {
	return c.SettingsOverride != nil
}

// TemplateOverride 範本覆寫
type TemplateOverride struct {
	Subject  *template.Subject         `json:"subject,omitempty"`
	Template *template.TemplateContent `json:"template,omitempty"`
}

// NewTemplateOverride 建立範本覆寫
func NewTemplateOverride() *TemplateOverride {
	return &TemplateOverride{}
}

// WithSubject 設定主題覆寫
func (t *TemplateOverride) WithSubject(subject *template.Subject) *TemplateOverride {
	t.Subject = subject
	return t
}

// WithTemplate 設定範本覆寫
func (t *TemplateOverride) WithTemplate(templateContent *template.TemplateContent) *TemplateOverride {
	t.Template = templateContent
	return t
}

// HasSubjectOverride 檢查是否有主題覆寫
func (t *TemplateOverride) HasSubjectOverride() bool {
	return t.Subject != nil
}

// HasTemplateOverride 檢查是否有範本覆寫
func (t *TemplateOverride) HasTemplateOverride() bool {
	return t.Template != nil
}

// ChannelOverrides 通道覆寫設定映射
type ChannelOverrides struct {
	overrides map[string]*ChannelOverride
}

// NewChannelOverrides 建立通道覆寫設定映射
func NewChannelOverrides(overrides map[string]*ChannelOverride) *ChannelOverrides {
	if overrides == nil {
		overrides = make(map[string]*ChannelOverride)
	}
	return &ChannelOverrides{overrides: overrides}
}

// Get 取得指定通道的覆寫設定
func (c *ChannelOverrides) Get(channelID string) (*ChannelOverride, bool) {
	override, exists := c.overrides[channelID]
	return override, exists
}

// Set 設定指定通道的覆寫設定
func (c *ChannelOverrides) Set(channelID string, override *ChannelOverride) {
	c.overrides[channelID] = override
}

// ToMap 轉換為 map
func (c *ChannelOverrides) ToMap() map[string]*ChannelOverride {
	result := make(map[string]*ChannelOverride)
	for k, v := range c.overrides {
		result[k] = v
	}
	return result
}

// HasOverride 檢查是否有指定通道的覆寫設定
func (c *ChannelOverrides) HasOverride(channelID string) bool {
	_, exists := c.overrides[channelID]
	return exists
}

// ChannelIDs 通道 ID 列表
type ChannelIDs struct {
	channelIDs []*channel.ChannelID
}

// NewChannelIDs 建立通道 ID 列表
func NewChannelIDs(channelIDs []*channel.ChannelID) (*ChannelIDs, error) {
	if len(channelIDs) == 0 {
		return nil, errors.New("channel IDs cannot be empty")
	}
	
	// 去重
	seen := make(map[string]bool)
	uniqueIDs := make([]*channel.ChannelID, 0)
	
	for _, id := range channelIDs {
		if id != nil && !seen[id.String()] {
			uniqueIDs = append(uniqueIDs, id)
			seen[id.String()] = true
		}
	}
	
	if len(uniqueIDs) == 0 {
		return nil, errors.New("no valid channel IDs provided")
	}
	
	return &ChannelIDs{channelIDs: uniqueIDs}, nil
}

// ToSlice 轉換為切片
func (c *ChannelIDs) ToSlice() []*channel.ChannelID {
	result := make([]*channel.ChannelID, len(c.channelIDs))
	copy(result, c.channelIDs)
	return result
}

// Count 取得通道 ID 數量
func (c *ChannelIDs) Count() int {
	return len(c.channelIDs)
}

// Contains 檢查是否包含指定通道 ID
func (c *ChannelIDs) Contains(channelID *channel.ChannelID) bool {
	if channelID == nil {
		return false
	}
	
	for _, id := range c.channelIDs {
		if id.Equals(channelID) {
			return true
		}
	}
	return false
}