package usecases

import (
	"context"
	"fmt"

	"channel-api/internal/application/channel/dtos"
	"channel-api/internal/domain/channel"
	"channel-api/internal/domain/services"
)

// DeleteChannelUseCase 刪除通道用例
type DeleteChannelUseCase struct {
	channelRepo channel.ChannelRepository
	validator   *services.ChannelValidator
}

// NewDeleteChannelUseCase 建立用例實例
func NewDeleteChannelUseCase(
	channelRepo channel.ChannelRepository,
	validator *services.ChannelValidator,
) *DeleteChannelUseCase {
	return &DeleteChannelUseCase{
		channelRepo: channelRepo,
		validator:   validator,
	}
}

// Execute 執行刪除通道
func (uc *DeleteChannelUseCase) Execute(ctx context.Context, channelID string) (*dtos.DeleteChannelResponse, error) {
	// 1. 驗證輸入參數
	if channelID == "" {
		return nil, fmt.Errorf("channel ID is required")
	}

	// 2. 轉換為領域物件
	id, err := channel.NewChannelIDFromString(channelID)
	if err != nil {
		return nil, fmt.Errorf("invalid channel ID: %w", err)
	}

	// 3. 業務驗證
	if err := uc.validator.ValidateChannelDeletion(ctx, id); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 4. 查詢通道
	ch, err := uc.channelRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("channel not found: %w", err)
	}

	// 5. 執行軟刪除
	if err := ch.Delete(); err != nil {
		return nil, fmt.Errorf("failed to delete channel: %w", err)
	}

	// 6. 持久化
	if err := uc.channelRepo.Update(ctx, ch); err != nil {
		return nil, fmt.Errorf("failed to save channel deletion: %w", err)
	}

	// 7. 轉換為回應 DTO
	response := &dtos.DeleteChannelResponse{
		ChannelID: ch.ID().String(),
		Deleted:   true,
		DeletedAt: *ch.Timestamps().DeletedAt,
	}

	return response, nil
}