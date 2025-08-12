package usecases

import (
	"context"
	"fmt"

	"notification/internal/application/channel/dtos"
	"notification/internal/domain/channel"
)

// GetChannelUseCase is the use case for getting a single channel.
type GetChannelUseCase struct {
	channelRepo channel.ChannelRepository
}

// NewGetChannelUseCase creates a use case instance.
func NewGetChannelUseCase(channelRepo channel.ChannelRepository) *GetChannelUseCase {
	return &GetChannelUseCase{
		channelRepo: channelRepo,
	}
}

// Execute executes the get channel operation.
func (uc *GetChannelUseCase) Execute(ctx context.Context, channelID string) (*dtos.ChannelResponse, error) {
	// 1. Validate input parameters
	if channelID == "" {
		return nil, fmt.Errorf("channel ID is required")
	}

	// 2. Convert to domain object
	id, err := channel.NewChannelIDFromString(channelID)
	if err != nil {
		return nil, fmt.Errorf("invalid channel ID: %w", err)
	}

	// 3. Query the channel
	ch, err := uc.channelRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("channel not found: %w", err)
	}

	// 4. Check if the channel is deleted
	if ch.IsDeleted() {
		return nil, fmt.Errorf("channel has been deleted")
	}

	// 5. Convert to response DTO
	response := uc.convertToResponse(ch)
	return response, nil
}

// convertToResponse converts to a response DTO.
func (uc *GetChannelUseCase) convertToResponse(ch *channel.Channel) *dtos.ChannelResponse {
	var templateID string
	if ch.TemplateID() != nil {
		templateID = ch.TemplateID().String()
	}

	return &dtos.ChannelResponse{
		ChannelID:      ch.ID().String(),
		ChannelName:    ch.Name().String(),
		Description:    ch.Description().String(),
		Enabled:        ch.IsEnabled(),
		ChannelType:    string(ch.ChannelType()),
		TemplateID:     templateID,
		CommonSettings: dtos.FromCommonSettings(ch.CommonSettings()),
		Config:         ch.Config().ToMap(),
		Recipients:     dtos.FromRecipientsSlice(ch.Recipients().ToSlice()),
		Tags:           ch.Tags().ToSlice(),
		CreatedAt:      ch.Timestamps().CreatedAt,
		UpdatedAt:      ch.Timestamps().UpdatedAt,
		LastUsed:       ch.LastUsed(),
	}
}