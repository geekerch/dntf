package usecases

import (
	"context"
	"fmt"

	"notification/internal/application/message/dtos"
	"notification/internal/domain/message"
)

// GetMessageUseCase handles getting a single message.
type GetMessageUseCase struct {
	messageRepo message.MessageRepository
}

// NewGetMessageUseCase creates a new GetMessageUseCase.
func NewGetMessageUseCase(messageRepo message.MessageRepository) *GetMessageUseCase {
	return &GetMessageUseCase{
		messageRepo: messageRepo,
	}
}

// Execute gets a message by ID.
func (uc *GetMessageUseCase) Execute(ctx context.Context, id string) (*dtos.MessageResponse, error) {
	// Validate input
	if id == "" {
		return nil, fmt.Errorf("message ID cannot be empty")
	}

	// Create message ID
	messageID, err := message.NewMessageIDFromString(id)
	if err != nil {
		return nil, fmt.Errorf("invalid message ID: %w", err)
	}

	// Find message
	messageEntity, err := uc.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to find message: %w", err)
	}

	// Convert to response
	return dtos.ToMessageResponse(messageEntity), nil
}