package usecases

import (
	"context"
	"fmt"

	"notification/internal/application/channel/dtos"
	"notification/internal/domain/channel"
	"notification/internal/domain/services"
)

// DeleteChannelUseCase is the use case for deleting a channel.
type DeleteChannelUseCase struct {
	channelRepo channel.ChannelRepository
	validator   *services.ChannelValidator
}

// NewDeleteChannelUseCase creates a use case instance.
func NewDeleteChannelUseCase(
	channelRepo channel.ChannelRepository,
	validator *services.ChannelValidator,
) *DeleteChannelUseCase {
	return &DeleteChannelUseCase{
		channelRepo: channelRepo,
		validator:   validator,
	}
}

// Execute executes the delete channel operation.
func (uc *DeleteChannelUseCase) Execute(ctx context.Context, channelID string) (*dtos.DeleteChannelResponse, error) {
	// 1. Validate input parameters
	if channelID == "" {
		return nil, fmt.Errorf("channel ID is required")
	}

	// 2. Convert to domain object
	id, err := channel.NewChannelIDFromString(channelID)
	if err != nil {
		return nil, fmt.Errorf("invalid channel ID: %w", err)
	}

	// 3. Business validation
	if err := uc.validator.ValidateChannelDeletion(ctx, id); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 4. Query the channel
	ch, err := uc.channelRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("channel not found: %w", err)
	}

	// 5. Perform soft deletion
	if err := ch.Delete(); err != nil {
		return nil, fmt.Errorf("failed to delete channel: %w", err)
	}

	// 6. Persist
	if err := uc.channelRepo.Update(ctx, ch); err != nil {
		return nil, fmt.Errorf("failed to save channel deletion: %w", err)
	}

	// 7. Convert to response DTO
	response := &dtos.DeleteChannelResponse{
		ChannelID: ch.ID().String(),
		Deleted:   true,
		DeletedAt: *ch.Timestamps().DeletedAt,
	}

	return response, nil
}