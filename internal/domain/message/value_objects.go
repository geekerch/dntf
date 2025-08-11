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

// MessageID is the unique identifier for a message.
type MessageID struct {
	value string
}

// NewMessageID creates a new message ID.
func NewMessageID() *MessageID {
	timestamp := time.Now().Unix()
	return &MessageID{
		value: fmt.Sprintf("msg_%d_%s", timestamp, uuid.New().String()[:8]),
	}
}

// NewMessageIDFromString creates a message ID from a string.
func NewMessageIDFromString(id string) (*MessageID, error) {
	if id == "" {
		return nil, errors.New("message ID cannot be empty")
	}
	return &MessageID{value: id}, nil
}

// String returns the string representation.
func (m *MessageID) String() string {
	return m.value
}

// Equals compares whether two message IDs are equal.
func (m *MessageID) Equals(other *MessageID) bool {
	if other == nil {
		return false
	}
	return m.value == other.value
}

// MessageStatus is the status of the message.
type MessageStatus string

const (
	MessageStatusSuccess        MessageStatus = "success"
	MessageStatusFailed         MessageStatus = "failed"
	MessageStatusPartialSuccess MessageStatus = "partial_success"
	MessageStatusPending        MessageStatus = "pending"
)

// IsValid validates if the message status is valid.
func (ms MessageStatus) IsValid() bool {
	switch ms {
	case MessageStatusSuccess, MessageStatusFailed, MessageStatusPartialSuccess, MessageStatusPending:
		return true
	default:
		return false
	}
}

// Variables are the template variables.
type Variables struct {
	variables map[string]interface{}
}

// NewVariables creates template variables.
func NewVariables(variables map[string]interface{}) *Variables {
	if variables == nil {
		variables = make(map[string]interface{})
	}
	return &Variables{variables: variables}
}

// Get gets the variable value.
func (v *Variables) Get(key string) (interface{}, bool) {
	value, exists := v.variables[key]
	return value, exists
}

// Set sets the variable value.
func (v *Variables) Set(key string, value interface{}) {
	v.variables[key] = value
}

// ToMap converts to a map.
func (v *Variables) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range v.variables {
		result[k] = v
	}
	return result
}

// Keys gets all variable keys.
func (v *Variables) Keys() []string {
	keys := make([]string, 0, len(v.variables))
	for key := range v.variables {
		keys = append(keys, key)
	}
	return keys
}

// HasKey checks if it contains the specified key.
func (v *Variables) HasKey(key string) bool {
	_, exists := v.variables[key]
	return exists
}

// ChannelOverride is the channel override setting.
type ChannelOverride struct {
	Recipients       *channel.Recipients      `json:"recipients,omitempty"`
	TemplateOverride *TemplateOverride        `json:"templateOverride,omitempty"`
	SettingsOverride *shared.CommonSettings   `json:"settingsOverride,omitempty"`
}

// NewChannelOverride creates a channel override setting.
func NewChannelOverride() *ChannelOverride {
	return &ChannelOverride{}
}

// WithRecipients sets the recipient override.
func (c *ChannelOverride) WithRecipients(recipients *channel.Recipients) *ChannelOverride {
	c.Recipients = recipients
	return c
}

// WithTemplateOverride sets the template override.
func (c *ChannelOverride) WithTemplateOverride(templateOverride *TemplateOverride) *ChannelOverride {
	c.TemplateOverride = templateOverride
	return c
}

// WithSettingsOverride sets the setting override.
func (c *ChannelOverride) WithSettingsOverride(settingsOverride *shared.CommonSettings) *ChannelOverride {
	c.SettingsOverride = settingsOverride
	return c
}

// HasRecipientsOverride checks if there is a recipient override.
func (c *ChannelOverride) HasRecipientsOverride() bool {
	return c.Recipients != nil
}

// HasTemplateOverride checks if there is a template override.
func (c *ChannelOverride) HasTemplateOverride() bool {
	return c.TemplateOverride != nil
}

// HasSettingsOverride checks if there is a setting override.
func (c *ChannelOverride) HasSettingsOverride() bool {
	return c.SettingsOverride != nil
}

// TemplateOverride is the template override.
type TemplateOverride struct {
	Subject  *template.Subject         `json:"subject,omitempty"`
	Template *template.TemplateContent `json:"template,omitempty"`
}

// NewTemplateOverride creates a template override.
func NewTemplateOverride() *TemplateOverride {
	return &TemplateOverride{}
}

// WithSubject sets the subject override.
func (t *TemplateOverride) WithSubject(subject *template.Subject) *TemplateOverride {
	t.Subject = subject
	return t
}

// WithTemplate sets the template override.
func (t *TemplateOverride) WithTemplate(templateContent *template.TemplateContent) *TemplateOverride {
	t.Template = templateContent
	return t
}

// HasSubjectOverride checks if there is a subject override.
func (t *TemplateOverride) HasSubjectOverride() bool {
	return t.Subject != nil
}

// HasTemplateOverride checks if there is a template override.
func (t *TemplateOverride) HasTemplateOverride() bool {
	return t.Template != nil
}

// ChannelOverrides is the channel override setting map.
type ChannelOverrides struct {
	overrides map[string]*ChannelOverride
}

// NewChannelOverrides creates a channel override setting map.
func NewChannelOverrides(overrides map[string]*ChannelOverride) *ChannelOverrides {
	if overrides == nil {
		overrides = make(map[string]*ChannelOverride)
	}
	return &ChannelOverrides{overrides: overrides}
}

// Get gets the override setting for the specified channel.
func (c *ChannelOverrides) Get(channelID string) (*ChannelOverride, bool) {
	override, exists := c.overrides[channelID]
	return override, exists
}

// Set sets the override setting for the specified channel.
func (c *ChannelOverrides) Set(channelID string, override *ChannelOverride) {
	c.overrides[channelID] = override
}

// ToMap converts to a map.
func (c *ChannelOverrides) ToMap() map[string]*ChannelOverride {
	result := make(map[string]*ChannelOverride)
	for k, v := range c.overrides {
		result[k] = v
	}
	return result
}

// HasOverride checks if there is an override setting for the specified channel.
func (c *ChannelOverrides) HasOverride(channelID string) bool {
	_, exists := c.overrides[channelID]
	return exists
}

// ChannelIDs is the list of channel IDs.
type ChannelIDs struct {
	channelIDs []*channel.ChannelID
}

// NewChannelIDs creates a list of channel IDs.
func NewChannelIDs(channelIDs []*channel.ChannelID) (*ChannelIDs, error) {
	if len(channelIDs) == 0 {
		return nil, errors.New("channel IDs cannot be empty")
	}
	
	// Deduplicate
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

// ToSlice converts to a slice.
func (c *ChannelIDs) ToSlice() []*channel.ChannelID {
	result := make([]*channel.ChannelID, len(c.channelIDs))
	copy(result, c.channelIDs)
	return result
}

// Count gets the number of channel IDs.
func (c *ChannelIDs) Count() int {
	return len(c.channelIDs)
}

// Contains checks if it contains the specified channel ID.
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