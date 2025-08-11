package usecases

import (
	"context"
	"fmt"

	"channel-api/internal/application/channel/dtos"
	"channel-api/internal/domain/channel"
)

// GetChannelUseCase 取得單一通道用例
type GetChannelUseCase struct {
	channelRepo channel.ChannelRepository
}

// NewGetChannelUseCase 建立用例實例
func NewGetChannelUseCase(channelRepo channel.ChannelRepository) *GetChannelUseCase {
	return &GetChannelUseCase{
		channelRepo: channelRepo,
	}
}

// Execute 執行取得通道
func (uc *GetChannelUseCase) Execute(ctx context.Context, channelID string) (*dtos.ChannelResponse, error) {
	// 1. 驗證輸入參數
	if channelID == "" {
		return nil, fmt.Errorf("channel ID is required")
	}

	// 2. 轉換為領域物件
	id, err := channel.NewChannelIDFromString(channelID)
	if err != nil {
		return nil, fmt.Errorf("invalid channel ID: %w", err)
	}

	// 3. 查詢通道
	ch, err := uc.channelRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("channel not found: %w", err)
	}

	// 4. 檢查通道是否已刪除
	if ch.IsDeleted() {
		return nil, fmt.Errorf("channel has been deleted")
	}

	// 5. 轉換為回應 DTO
	response := uc.convertToResponse(ch)
	return response, nil
}

// convertToResponse 轉換為回應 DTO
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