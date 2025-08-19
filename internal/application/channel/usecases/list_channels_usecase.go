package usecases

import (
	"context"
	"fmt"

	"notification/internal/application/channel/dtos"
	"notification/internal/domain/channel"
	"notification/internal/domain/shared"
)

// ListChannelsUseCase is the use case for listing channels.
type ListChannelsUseCase struct {
	channelRepo channel.ChannelRepository
}

// NewListChannelsUseCase creates a use case instance.
func NewListChannelsUseCase(channelRepo channel.ChannelRepository) *ListChannelsUseCase {
	return &ListChannelsUseCase{
		channelRepo: channelRepo,
	}
}

// Execute executes the channel list query.
func (uc *ListChannelsUseCase) Execute(ctx context.Context, request *dtos.ListChannelsRequest) (*dtos.ListChannelsResponse, error) {
	// 1. Create pagination parameters
	pagination, err := uc.createPagination(request)
	if err != nil {
		return nil, fmt.Errorf("invalid pagination: %w", err)
	}

	// 2. Create filter conditions
	filter := uc.createFilter(request)

	// 3. Query data
	result, err := uc.channelRepo.FindAll(ctx, filter, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to list channels: %w", err)
	}

	// 4. Convert to response DTO
	response := uc.convertToResponse(result)
	return response, nil
}

// createPagination creates pagination parameters.
func (uc *ListChannelsUseCase) createPagination(request *dtos.ListChannelsRequest) (*shared.Pagination, error) {
	skipCount := request.SkipCount
	maxResultCount := request.MaxResultCount

	// Set default values
	if skipCount < 0 {
		skipCount = 0
	}
	if maxResultCount <= 0 {
		maxResultCount = 10
	}
	if maxResultCount > 100 {
		maxResultCount = 100
	}

	return shared.NewPagination(skipCount, maxResultCount)
}

// createFilter creates filter conditions.
func (uc *ListChannelsUseCase) createFilter(request *dtos.ListChannelsRequest) *channel.ChannelFilter {
	filter := channel.NewChannelFilter()

	// Channel type filter
	if request.ChannelType != "" {
		channelType, err := shared.NewChannelTypeFromString(request.ChannelType)
		if err == nil && channelType.IsValid() {
			filter.WithChannelType(channelType)
		}
	}

	// Tag filter
	if len(request.Tags) > 0 {
		filter.WithTags(request.Tags)
	}

	return filter
}

// convertToResponse converts to a response DTO.
func (uc *ListChannelsUseCase) convertToResponse(result *shared.PaginatedResult[*channel.Channel]) *dtos.ListChannelsResponse {
	items := make([]dtos.ChannelSummaryResponse, 0, len(result.Items))

	for _, ch := range result.Items {
		items = append(items, dtos.ChannelSummaryResponse{
			ChannelID:   ch.ID().String(),
			ChannelName: ch.Name().String(),
			ChannelType: ch.ChannelType().String(),
			Tags:        ch.Tags().ToSlice(),
			Enabled:     ch.IsEnabled(),
			CreatedAt:   ch.Timestamps().CreatedAt,
			UpdatedAt:   ch.Timestamps().UpdatedAt,
		})
	}

	return &dtos.ListChannelsResponse{
		Items:          items,
		SkipCount:      result.SkipCount,
		MaxResultCount: result.MaxResultCount,
		TotalCount:     result.TotalCount,
		HasMore:        result.HasMore,
	}
}