package channel

import (
	"errors"

	"channel-api/internal/domain/shared"
	"channel-api/internal/domain/template"
)

// Channel represents the channel aggregate root
type Channel struct {
	id             *ChannelID
	name           *ChannelName
	description    *Description
	enabled        bool
	channelType    shared.ChannelType
	templateID     *template.TemplateID
	commonSettings *shared.CommonSettings
	config         *ChannelConfig
	recipients     *Recipients
	tags           *Tags
	timestamps     *shared.Timestamps
	lastUsed       *int64
}

// NewChannel creates a new channel
func NewChannel(
	name *ChannelName,
	description *Description,
	enabled bool,
	channelType shared.ChannelType,
	templateID *template.TemplateID,
	commonSettings *shared.CommonSettings,
	config *ChannelConfig,
	recipients *Recipients,
	tags *Tags,
) (*Channel, error) {
	// Validate required fields
	if name == nil {
		return nil, errors.New("channel name is required")
	}
	if !channelType.IsValid() {
		return nil, errors.New("invalid channel type")
	}
	if commonSettings == nil {
		return nil, errors.New("common settings is required")
	}
	
	// Set default values
	if description == nil {
		description, _ = NewDescription("")
	}
	if config == nil {
		config = NewChannelConfig(nil)
	}
	if recipients == nil {
		recipients = NewRecipients(nil)
	}
	if tags == nil {
		tags = NewTags(nil)
	}

	return &Channel{
		id:             NewChannelID(),
		name:           name,
		description:    description,
		enabled:        enabled,
		channelType:    channelType,
		templateID:     templateID,
		commonSettings: commonSettings,
		config:         config,
		recipients:     recipients,
		tags:           tags,
		timestamps:     shared.NewTimestamps(),
		lastUsed:       nil,
	}, nil
}

// ReconstructChannel reconstructs a channel from persisted data
func ReconstructChannel(
	id *ChannelID,
	name *ChannelName,
	description *Description,
	enabled bool,
	channelType shared.ChannelType,
	templateID *template.TemplateID,
	commonSettings *shared.CommonSettings,
	config *ChannelConfig,
	recipients *Recipients,
	tags *Tags,
	timestamps *shared.Timestamps,
	lastUsed *int64,
) *Channel {
	return &Channel{
		id:             id,
		name:           name,
		description:    description,
		enabled:        enabled,
		channelType:    channelType,
		templateID:     templateID,
		commonSettings: commonSettings,
		config:         config,
		recipients:     recipients,
		tags:           tags,
		timestamps:     timestamps,
		lastUsed:       lastUsed,
	}
}

// ID 取得通道 ID
func (c *Channel) ID() *ChannelID {
	return c.id
}

// Name 取得通道名稱
func (c *Channel) Name() *ChannelName {
	return c.name
}

// Description 取得描述
func (c *Channel) Description() *Description {
	return c.description
}

// IsEnabled 檢查通道是否啟用
func (c *Channel) IsEnabled() bool {
	return c.enabled
}

// ChannelType 取得通道類型
func (c *Channel) ChannelType() shared.ChannelType {
	return c.channelType
}

// TemplateID 取得範本 ID
func (c *Channel) TemplateID() *template.TemplateID {
	return c.templateID
}

// CommonSettings 取得通用設定
func (c *Channel) CommonSettings() *shared.CommonSettings {
	return c.commonSettings
}

// Config 取得通道配置
func (c *Channel) Config() *ChannelConfig {
	return c.config
}

// Recipients 取得收件人列表
func (c *Channel) Recipients() *Recipients {
	return c.recipients
}

// Tags 取得標籤
func (c *Channel) Tags() *Tags {
	return c.tags
}

// Timestamps 取得時間戳記
func (c *Channel) Timestamps() *shared.Timestamps {
	return c.timestamps
}

// LastUsed 取得最後使用時間
func (c *Channel) LastUsed() *int64 {
	return c.lastUsed
}

// Update 更新通道
func (c *Channel) Update(
	name *ChannelName,
	description *Description,
	enabled bool,
	channelType shared.ChannelType,
	templateID *template.TemplateID,
	commonSettings *shared.CommonSettings,
	config *ChannelConfig,
	recipients *Recipients,
	tags *Tags,
) error {
	// 驗證必要欄位
	if name == nil {
		return errors.New("channel name is required")
	}
	if !channelType.IsValid() {
		return errors.New("invalid channel type")
	}
	if commonSettings == nil {
		return errors.New("common settings is required")
	}

	// 設定預設值
	if description == nil {
		description, _ = NewDescription("")
	}
	if config == nil {
		config = NewChannelConfig(nil)
	}
	if recipients == nil {
		recipients = NewRecipients(nil)
	}
	if tags == nil {
		tags = NewTags(nil)
	}

	// 更新欄位
	c.name = name
	c.description = description
	c.enabled = enabled
	c.channelType = channelType
	c.templateID = templateID
	c.commonSettings = commonSettings
	c.config = config
	c.recipients = recipients
	c.tags = tags
	c.timestamps.UpdateTimestamp()

	return nil
}

// Enable 啟用通道
func (c *Channel) Enable() {
	c.enabled = true
	c.timestamps.UpdateTimestamp()
}

// Disable 停用通道
func (c *Channel) Disable() {
	c.enabled = false
	c.timestamps.UpdateTimestamp()
}

// MarkAsUsed 標記為已使用
func (c *Channel) MarkAsUsed() {
	now := c.timestamps.UpdatedAt
	c.lastUsed = &now
	c.timestamps.UpdateTimestamp()
}

// CanSendMessage 檢查是否可以發送訊息
func (c *Channel) CanSendMessage() error {
	if c.timestamps.IsDeleted() {
		return errors.New("channel is deleted")
	}
	if !c.enabled {
		return errors.New("channel is disabled")
	}
	if c.recipients.Count() == 0 {
		return errors.New("channel has no recipients")
	}
	return nil
}

// Delete 軟刪除通道
func (c *Channel) Delete() error {
	if c.timestamps.IsDeleted() {
		return errors.New("channel is already deleted")
	}
	c.timestamps.MarkDeleted()
	return nil
}

// IsDeleted 檢查通道是否已刪除
func (c *Channel) IsDeleted() bool {
	return c.timestamps.IsDeleted()
}

// HasTag 檢查是否包含指定標籤
func (c *Channel) HasTag(tag string) bool {
	return c.tags.Contains(tag)
}

// HasAnyTag 檢查是否包含任一指定標籤
func (c *Channel) HasAnyTag(tags []string) bool {
	return c.tags.ContainsAny(tags)
}

// MatchesType 檢查通道類型是否匹配
func (c *Channel) MatchesType(channelType shared.ChannelType) bool {
	return c.channelType == channelType
}