package external

import (
	"context"

	"notification/internal/domain/channel"
	"notification/internal/domain/services"
)

// MessageSender defines the interface for sending messages through different channels
type MessageSender interface {
	// Send sends a message through the specified channel
	Send(ctx context.Context, ch *channel.Channel, content *services.RenderedContent) error
	
	// GetChannelType returns the channel type this sender supports
	GetChannelType() string
	
	// ValidateConfig validates the channel configuration for this sender
	ValidateConfig(config *channel.ChannelConfig) error
}

// MessageSenderFactory creates message senders for different channel types
type MessageSenderFactory interface {
	// CreateSender creates a message sender for the given channel type
	CreateSender(channelType string) (MessageSender, error)
	
	// GetSupportedTypes returns all supported channel types
	GetSupportedTypes() []string
}

// SendRequest represents a message sending request
type SendRequest struct {
	Channel   *channel.Channel
	Content   *services.RenderedContent
	Variables map[string]interface{}
}

// SendResult represents the result of a message sending operation
type SendResult struct {
	Success   bool
	Message   string
	Error     error
	Details   map[string]interface{}
	SentAt    int64
}

// NotificationService provides a high-level interface for sending notifications
type NotificationService interface {
	// SendNotification sends a notification through multiple channels
	SendNotification(ctx context.Context, requests []*SendRequest) ([]*SendResult, error)
	
	// SendSingleNotification sends a notification through a single channel
	SendSingleNotification(ctx context.Context, request *SendRequest) (*SendResult, error)
	
	// ValidateChannel validates if a channel can be used for sending
	ValidateChannel(ch *channel.Channel) error
}