package message

import (
	"notification/internal/application/cqrs"
	"notification/internal/application/message/dtos"
)

// Event types
const (
	MessageSentEventType     = "message.sent"
	MessageFailedEventType   = "message.failed"
	MessageDeliveredEventType = "message.delivered"
)

// MessageSentEvent represents an event when a message is sent
type MessageSentEvent struct {
	*cqrs.BaseEvent
	Message *dtos.MessageResponse `json:"message"`
}

// NewMessageSentEvent creates a new message sent event
func NewMessageSentEvent(message *dtos.MessageResponse) *MessageSentEvent {
	return &MessageSentEvent{
		BaseEvent: cqrs.NewBaseEvent(MessageSentEventType, "message", "sent", 0, message),
		Message:   message,
	}
}

// MessageFailedEvent represents an event when a message fails to send
type MessageFailedEvent struct {
	*cqrs.BaseEvent
	MessageID string `json:"messageId"`
	Error     string `json:"error"`
}

// NewMessageFailedEvent creates a new message failed event
func NewMessageFailedEvent(messageID, error string) *MessageFailedEvent {
	return &MessageFailedEvent{
		BaseEvent: cqrs.NewBaseEvent(MessageFailedEventType, "message", "failed", 0, map[string]string{"messageId": messageID, "error": error}),
		MessageID: messageID,
		Error:     error,
	}
}

// MessageDeliveredEvent represents an event when a message is delivered
type MessageDeliveredEvent struct {
	*cqrs.BaseEvent
	MessageID string `json:"messageId"`
	Recipient string `json:"recipient"`
}

// NewMessageDeliveredEvent creates a new message delivered event
func NewMessageDeliveredEvent(messageID, recipient string) *MessageDeliveredEvent {
	return &MessageDeliveredEvent{
		BaseEvent: cqrs.NewBaseEvent(MessageDeliveredEventType, "message", "delivered", 0, map[string]string{"messageId": messageID, "recipient": recipient}),
		MessageID: messageID,
		Recipient: recipient,
	}
}