package message

import (
	"context"
)

// MessageRepository is the interface for the message repository.
type MessageRepository interface {
	// Save saves a message.
	Save(ctx context.Context, message *Message) error
	
	// FindByID finds a message by ID.
	FindByID(ctx context.Context, id *MessageID) (*Message, error)
	
	// Update updates a message.
	Update(ctx context.Context, message *Message) error
	
	// Exists checks if a message exists.
	Exists(ctx context.Context, id *MessageID) (bool, error)
}