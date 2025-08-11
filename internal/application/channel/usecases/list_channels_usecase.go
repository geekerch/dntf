package usecases

import (
	"context"
	"fmt"

	"channel-api/internal/application/channel/dtos"
	"channel-api/internal/domain/channel"
	"channel-api/internal/domain/shared"
)

// ListChannelsUseCase 查詢通道列表用例
type ListChannelsUseCase struct {
	channelRepo channel.ChannelRepository
}

// NewListChannelsUseCase 建立用例實例
func NewListChannelsUseCase(channelRepo channel.ChannelRepository) *ListChannelsUseCase {
	return &ListChannelsUseCase{
		channelRepo: channelRepo,
	}
}

// Execute 執行查詢通道列表
func (uc *ListChannelsUseCase) Execute(ctx context.Context, request *dtos.ListChannelsRequest) (*dtos.ListChannelsResponse, error) {
	// 1. 建立分頁參數
	pagination, err := uc.createPagination(request)
	if err != nil {
		return nil, fmt.Errorf("invalid pagination: %w", err)
	}

	// 2. 建立過濾條件
	filter := uc.createFilter(request)

	// 3. 查詢資料
	result, err := uc.channelRepo.FindAll(ctx, filter, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to list channels: %w", err)
	}

	// 4. 轉換為回應 DTO
	response := uc.convertToResponse(result)
	return response, nil
}

// createPagination 建立分頁參數
func (uc *ListChannelsUseCase) createPagination(request *dtos.ListChannelsRequest) (*shared.Pagination, error) {
	skipCount := request.SkipCount
	maxResultCount := request.MaxResultCount

	// 設定預設值
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

// createFilter 建立過濾條件
func (uc *ListChannelsUseCase) createFilter(request *dtos.ListChannelsRequest) *channel.ChannelFilter {
	filter := channel.NewChannelFilter()

	// 通道類型過濾
	if request.ChannelType != "" {
		channelType := shared.ChannelType(request.ChannelType)
		if channelType.IsValid() {
			filter.WithChannelType(channelType)
		}
	}

	// 標籤過濾
	if len(request.Tags) > 0 {
		filter.WithTags(request.Tags)
	}

	return filter
}

// convertToResponse 轉換為回應 DTO
func (uc *ListChannelsUseCase) convertToResponse(result *shared.PaginatedResult[*channel.Channel]) *dtos.ListChannelsResponse {
	items := make([]dtos.ChannelSummaryResponse, 0, len(result.Items))

	for _, ch := range result.Items {
		items = append(items, dtos.ChannelSummaryResponse{
			ChannelID:   ch.ID().String(),
			ChannelName: ch.Name().String(),
			ChannelType: string(ch.ChannelType()),
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