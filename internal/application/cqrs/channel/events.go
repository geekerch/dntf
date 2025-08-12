package channel

import (
	"notification/internal/application/cqrs"
	"notification/internal/application/channel/dtos"
)

// Event types
const (
	ChannelCreatedEventType = "channel.created"
	ChannelUpdatedEventType = "channel.updated"
	ChannelDeletedEventType = "channel.deleted"
	ChannelEnabledEventType = "channel.enabled"
	ChannelDisabledEventType = "channel.disabled"
)

// Aggregate type
const ChannelAggregateType = "channel"

// ChannelCreatedEvent represents an event when a channel is created
type ChannelCreatedEvent struct {
	*cqrs.BaseEvent
}

// ChannelCreatedEventData represents the data for channel created event
type ChannelCreatedEventData struct {
	ChannelID      string                 `json:"channelId"`
	ChannelName    string                 `json:"channelName"`
	Description    string                 `json:"description"`
	ChannelType    string                 `json:"channelType"`
	TemplateID     string                 `json:"templateId,omitempty"`
	Config         map[string]interface{} `json:"config"`
	Recipients     []dtos.RecipientDTO    `json:"recipients"`
	Tags           []string               `json:"tags"`
	Enabled        bool                   `json:"enabled"`
	CreatedAt      int64                  `json:"createdAt"`
}

// NewChannelCreatedEvent creates a new channel created event
func NewChannelCreatedEvent(channelID string, version int64, data *ChannelCreatedEventData) *ChannelCreatedEvent {
	return &ChannelCreatedEvent{
		BaseEvent: cqrs.NewBaseEvent(
			ChannelCreatedEventType,
			channelID,
			ChannelAggregateType,
			version,
			data,
		),
	}
}

// ChannelUpdatedEvent represents an event when a channel is updated
type ChannelUpdatedEvent struct {
	*cqrs.BaseEvent
}

// ChannelUpdatedEventData represents the data for channel updated event
type ChannelUpdatedEventData struct {
	ChannelID      string                 `json:"channelId"`
	ChannelName    string                 `json:"channelName"`
	Description    string                 `json:"description"`
	ChannelType    string                 `json:"channelType"`
	TemplateID     string                 `json:"templateId,omitempty"`
	Config         map[string]interface{} `json:"config"`
	Recipients     []dtos.RecipientDTO    `json:"recipients"`
	Tags           []string               `json:"tags"`
	Enabled        bool                   `json:"enabled"`
	UpdatedAt      int64                  `json:"updatedAt"`
	Changes        map[string]interface{} `json:"changes"` // What fields were changed
}

// NewChannelUpdatedEvent creates a new channel updated event
func NewChannelUpdatedEvent(channelID string, version int64, data *ChannelUpdatedEventData) *ChannelUpdatedEvent {
	return &ChannelUpdatedEvent{
		BaseEvent: cqrs.NewBaseEvent(
			ChannelUpdatedEventType,
			channelID,
			ChannelAggregateType,
			version,
			data,
		),
	}
}

// ChannelDeletedEvent represents an event when a channel is deleted
type ChannelDeletedEvent struct {
	*cqrs.BaseEvent
}

// ChannelDeletedEventData represents the data for channel deleted event
type ChannelDeletedEventData struct {
	ChannelID   string `json:"channelId"`
	ChannelName string `json:"channelName"`
	DeletedAt   int64  `json:"deletedAt"`
}

// NewChannelDeletedEvent creates a new channel deleted event
func NewChannelDeletedEvent(channelID string, version int64, data *ChannelDeletedEventData) *ChannelDeletedEvent {
	return &ChannelDeletedEvent{
		BaseEvent: cqrs.NewBaseEvent(
			ChannelDeletedEventType,
			channelID,
			ChannelAggregateType,
			version,
			data,
		),
	}
}

// ChannelEnabledEvent represents an event when a channel is enabled
type ChannelEnabledEvent struct {
	*cqrs.BaseEvent
}

// ChannelEnabledEventData represents the data for channel enabled event
type ChannelEnabledEventData struct {
	ChannelID   string `json:"channelId"`
	ChannelName string `json:"channelName"`
	EnabledAt   int64  `json:"enabledAt"`
}

// NewChannelEnabledEvent creates a new channel enabled event
func NewChannelEnabledEvent(channelID string, version int64, data *ChannelEnabledEventData) *ChannelEnabledEvent {
	return &ChannelEnabledEvent{
		BaseEvent: cqrs.NewBaseEvent(
			ChannelEnabledEventType,
			channelID,
			ChannelAggregateType,
			version,
			data,
		),
	}
}

// ChannelDisabledEvent represents an event when a channel is disabled
type ChannelDisabledEvent struct {
	*cqrs.BaseEvent
}

// ChannelDisabledEventData represents the data for channel disabled event
type ChannelDisabledEventData struct {
	ChannelID    string `json:"channelId"`
	ChannelName  string `json:"channelName"`
	DisabledAt   int64  `json:"disabledAt"`
}

// NewChannelDisabledEvent creates a new channel disabled event
func NewChannelDisabledEvent(channelID string, version int64, data *ChannelDisabledEventData) *ChannelDisabledEvent {
	return &ChannelDisabledEvent{
		BaseEvent: cqrs.NewBaseEvent(
			ChannelDisabledEventType,
			channelID,
			ChannelAggregateType,
			version,
			data,
		),
	}
}