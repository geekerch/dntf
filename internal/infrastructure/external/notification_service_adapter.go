package external

import (
	"context"

	"notification/internal/domain/channel"
	"notification/internal/domain/services"
)

// NotificationServiceAdapter adapts external.NotificationService to services.ExternalNotificationService
type NotificationServiceAdapter struct {
	notificationService NotificationService
}

// NewNotificationServiceAdapter creates a new adapter
func NewNotificationServiceAdapter(notificationService NotificationService) *NotificationServiceAdapter {
	return &NotificationServiceAdapter{
		notificationService: notificationService,
	}
}

// SendSingleNotification adapts the method signature to match services.ExternalNotificationService
func (a *NotificationServiceAdapter) SendSingleNotification(ctx context.Context, request *services.SendRequest) *services.SendResult {
	// Convert services.SendRequest to external.SendRequest
	externalRequest := &SendRequest{
		Channel:   request.Channel,
		Content:   request.Content,
		Variables: request.Variables,
	}

	// Call the external service
	externalResult := a.notificationService.SendSingleNotification(ctx, externalRequest)

	// Convert external.SendResult to services.SendResult
	return &services.SendResult{
		Success: externalResult.Success,
		Message: externalResult.Message,
		Error:   externalResult.Error,
		Details: externalResult.Details,
		SentAt:  externalResult.SentAt,
	}
}

// ValidateChannel validates if a channel can be used for sending
func (a *NotificationServiceAdapter) ValidateChannel(ch *channel.Channel) error {
	return a.notificationService.ValidateChannel(ch)
}