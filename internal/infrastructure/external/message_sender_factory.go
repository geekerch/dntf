package external

import (
	"context"
	"fmt"
	"sync"
	"time"

	"channel-api/internal/domain/channel"
	"channel-api/internal/domain/services"
)

// DefaultMessageSenderFactory implements MessageSenderFactory
type DefaultMessageSenderFactory struct {
	senders map[string]MessageSender
	mutex   sync.RWMutex
}

// NewDefaultMessageSenderFactory creates a new message sender factory
func NewDefaultMessageSenderFactory(timeout time.Duration) *DefaultMessageSenderFactory {
	factory := &DefaultMessageSenderFactory{
		senders: make(map[string]MessageSender),
	}

	// Register default senders
	factory.RegisterSender(NewEmailService(timeout))
	factory.RegisterSender(NewSlackService(timeout))
	factory.RegisterSender(NewSMSService(timeout))

	return factory
}

// RegisterSender registers a message sender for a specific channel type
func (f *DefaultMessageSenderFactory) RegisterSender(sender MessageSender) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	
	f.senders[sender.GetChannelType()] = sender
}

// CreateSender creates a message sender for the given channel type
func (f *DefaultMessageSenderFactory) CreateSender(channelType string) (MessageSender, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	sender, exists := f.senders[channelType]
	if !exists {
		return nil, fmt.Errorf("unsupported channel type: %s", channelType)
	}

	return sender, nil
}

// GetSupportedTypes returns all supported channel types
func (f *DefaultMessageSenderFactory) GetSupportedTypes() []string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	types := make([]string, 0, len(f.senders))
	for channelType := range f.senders {
		types = append(types, channelType)
	}

	return types
}

// DefaultNotificationService implements NotificationService
type DefaultNotificationService struct {
	factory MessageSenderFactory
}

// NewDefaultNotificationService creates a new notification service
func NewDefaultNotificationService(factory MessageSenderFactory) *DefaultNotificationService {
	return &DefaultNotificationService{
		factory: factory,
	}
}

// SendNotification sends a notification through multiple channels
func (s *DefaultNotificationService) SendNotification(ctx context.Context, requests []*SendRequest) ([]*SendResult, error) {
	results := make([]*SendResult, 0, len(requests))

	for _, request := range requests {
		result := s.SendSingleNotification(ctx, request)
		results = append(results, result)
	}

	return results, nil
}

// SendSingleNotification sends a notification through a single channel
func (s *DefaultNotificationService) SendSingleNotification(ctx context.Context, request *SendRequest) *SendResult {
	startTime := time.Now()

	// Validate request
	if err := s.validateSendRequest(request); err != nil {
		return &SendResult{
			Success: false,
			Message: "Request validation failed",
			Error:   err,
			Details: map[string]interface{}{
				"channel_id":   request.Channel.ID().String(),
				"channel_type": string(request.Channel.ChannelType()),
			},
		}
	}

	// Get sender for channel type
	sender, err := s.factory.CreateSender(string(request.Channel.ChannelType()))
	if err != nil {
		return &SendResult{
			Success: false,
			Message: "Failed to create message sender",
			Error:   err,
			Details: map[string]interface{}{
				"channel_id":   request.Channel.ID().String(),
				"channel_type": string(request.Channel.ChannelType()),
			},
		}
	}

	// Validate channel configuration
	if err := sender.ValidateConfig(request.Channel.Config()); err != nil {
		return &SendResult{
			Success: false,
			Message: "Channel configuration validation failed",
			Error:   err,
			Details: map[string]interface{}{
				"channel_id":   request.Channel.ID().String(),
				"channel_type": string(request.Channel.ChannelType()),
			},
		}
	}

	// Send message
	if err := sender.Send(ctx, request.Channel, request.Content); err != nil {
		return &SendResult{
			Success: false,
			Message: "Failed to send message",
			Error:   err,
			Details: map[string]interface{}{
				"channel_id":   request.Channel.ID().String(),
				"channel_type": string(request.Channel.ChannelType()),
				"duration_ms":  time.Since(startTime).Milliseconds(),
			},
		}
	}

	return &SendResult{
		Success: true,
		Message: "Message sent successfully",
		Error:   nil,
		Details: map[string]interface{}{
			"channel_id":   request.Channel.ID().String(),
			"channel_type": string(request.Channel.ChannelType()),
			"duration_ms":  time.Since(startTime).Milliseconds(),
		},
		SentAt: time.Now().UnixMilli(),
	}
}

// ValidateChannel validates if a channel can be used for sending
func (s *DefaultNotificationService) ValidateChannel(ch *channel.Channel) error {
	// Check if channel is enabled
	if !ch.IsEnabled() {
		return fmt.Errorf("channel is disabled")
	}

	// Check if channel is deleted
	if ch.IsDeleted() {
		return fmt.Errorf("channel is deleted")
	}

	// Check if channel type is supported
	sender, err := s.factory.CreateSender(string(ch.ChannelType()))
	if err != nil {
		return fmt.Errorf("unsupported channel type: %s", ch.ChannelType())
	}

	// Validate channel configuration
	if err := sender.ValidateConfig(ch.Config()); err != nil {
		return fmt.Errorf("invalid channel configuration: %w", err)
	}

	// Check if channel can send messages
	if err := ch.CanSendMessage(); err != nil {
		return fmt.Errorf("channel cannot send messages: %w", err)
	}

	return nil
}

// validateSendRequest validates a send request
func (s *DefaultNotificationService) validateSendRequest(request *SendRequest) error {
	if request == nil {
		return fmt.Errorf("send request cannot be nil")
	}

	if request.Channel == nil {
		return fmt.Errorf("channel cannot be nil")
	}

	if request.Content == nil {
		return fmt.Errorf("content cannot be nil")
	}

	return s.ValidateChannel(request.Channel)
}

// MockMessageSender is a mock implementation for testing
type MockMessageSender struct {
	channelType    string
	shouldSucceed  bool
	errorMessage   string
	sendDelay      time.Duration
}

// NewMockMessageSender creates a new mock message sender
func NewMockMessageSender(channelType string, shouldSucceed bool, errorMessage string, sendDelay time.Duration) *MockMessageSender {
	return &MockMessageSender{
		channelType:   channelType,
		shouldSucceed: shouldSucceed,
		errorMessage:  errorMessage,
		sendDelay:     sendDelay,
	}
}

// Send mock implementation
func (m *MockMessageSender) Send(ctx context.Context, ch *channel.Channel, content *services.RenderedContent) error {
	// Simulate processing delay
	if m.sendDelay > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(m.sendDelay):
		}
	}

	if !m.shouldSucceed {
		return fmt.Errorf(m.errorMessage)
	}

	return nil
}

// GetChannelType mock implementation
func (m *MockMessageSender) GetChannelType() string {
	return m.channelType
}

// ValidateConfig mock implementation
func (m *MockMessageSender) ValidateConfig(config *channel.ChannelConfig) error {
	if !m.shouldSucceed {
		return fmt.Errorf("mock validation error: %s", m.errorMessage)
	}
	return nil
}