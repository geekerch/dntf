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

// ID gets the channel ID.
func (c *Channel) ID() *ChannelID {
	return c.id
}

// Name gets the channel name.
func (c *Channel) Name() *ChannelName {
	return c.name
}

// Description gets the description.
func (c *Channel) Description() *Description {
	return c.description
}

// IsEnabled checks if the channel is enabled.
func (c *Channel) IsEnabled() bool {
	return c.enabled
}

// ChannelType gets the channel type.
func (c *Channel) ChannelType() shared.ChannelType {
	return c.channelType
}

// TemplateID gets the template ID.
func (c *Channel) TemplateID() *template.TemplateID {
	return c.templateID
}

// CommonSettings gets the common settings.
func (c *Channel) CommonSettings() *shared.CommonSettings {
	return c.commonSettings
}

// Config gets the channel configuration.
func (c *Channel) Config() *ChannelConfig {
	return c.config
}

// Recipients gets the list of recipients.
func (c *Channel) Recipients() *Recipients {
	return c.recipients
}

// Tags gets the tags.
func (c *Channel) Tags() *Tags {
	return c.tags
}

// Timestamps gets the timestamps.
func (c *Channel) Timestamps() *shared.Timestamps {
	return c.timestamps
}

// LastUsed gets the last used time.
func (c *Channel) LastUsed() *int64 {
	return c.lastUsed
}

// Update updates the channel.
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
	// Validate required fields
	if name == nil {
		return errors.New("channel name is required")
	}
	if !channelType.IsValid() {
		return errors.New("invalid channel type")
	}
	if commonSettings == nil {
		return errors.New("common settings is required")
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

	// Update fields
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

// Enable enables the channel.
func (c *Channel) Enable() {
	c.enabled = true
	c.timestamps.UpdateTimestamp()
}

// Disable disables the channel.
func (c *Channel) Disable() {
	c.enabled = false
	c.timestamps.UpdateTimestamp()
}

// MarkAsUsed marks the channel as used.
func (c *Channel) MarkAsUsed() {
	now := c.timestamps.UpdatedAt
	c.lastUsed = &now
	c.timestamps.UpdateTimestamp()
}

// CanSendMessage checks if a message can be sent.
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

// Delete soft deletes the channel.
func (c *Channel) Delete() error {
	if c.timestamps.IsDeleted() {
		return errors.New("channel is already deleted")
	}
	c.timestamps.MarkDeleted()
	return nil
}

// IsDeleted checks if the channel is deleted.
func (c *Channel) IsDeleted() bool {
	return c.timestamps.IsDeleted()
}

// HasTag checks if it contains the specified tag.
func (c *Channel) HasTag(tag string) bool {
	return c.tags.Contains(tag)
}

// HasAnyTag checks if it contains any of the specified tags.
func (c *Channel) HasAnyTag(tags []string) bool {
	return c.tags.ContainsAny(tags)
}

// MatchesType checks if the channel type matches.
func (c *Channel) MatchesType(channelType shared.ChannelType) bool {
	return c.channelType == channelType
}