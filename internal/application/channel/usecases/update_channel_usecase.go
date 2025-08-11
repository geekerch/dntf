package usecases

import (
	"context"
	"fmt"

	"channel-api/internal/application/channel/dtos"
	"channel-api/internal/domain/channel"
	"channel-api/internal/domain/services"
	"channel-api/internal/domain/shared"
	"channel-api/internal/domain/template"
)

// UpdateChannelUseCase 更新通道用例
type UpdateChannelUseCase struct {
	channelRepo channel.ChannelRepository
	validator   *services.ChannelValidator
}

// NewUpdateChannelUseCase 建立用例實例
func NewUpdateChannelUseCase(
	channelRepo channel.ChannelRepository,
	validator *services.ChannelValidator,
) *UpdateChannelUseCase {
	return &UpdateChannelUseCase{
		channelRepo: channelRepo,
		validator:   validator,
	}
}

// Execute 執行更新通道
func (uc *UpdateChannelUseCase) Execute(ctx context.Context, channelID string, request *dtos.UpdateChannelRequest) (*dtos.ChannelResponse, error) {
	// 1. 驗證輸入參數
	if err := uc.validateRequest(channelID, request); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// 2. 轉換為領域物件
	id, err := channel.NewChannelIDFromString(channelID)
	if err != nil {
		return nil, fmt.Errorf("invalid channel ID: %w", err)
	}

	domainObjects, err := uc.convertToDomainObjects(request)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to domain objects: %w", err)
	}

	// 3. 業務驗證
	if err := uc.validator.ValidateChannelForUpdate(
		ctx,
		id,
		domainObjects.Name,
		domainObjects.ChannelType,
		domainObjects.TemplateID,
		domainObjects.Config,
	); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 4. 查詢現有通道
	ch, err := uc.channelRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("channel not found: %w", err)
	}

	// 5. 檢查通道是否已刪除
	if ch.IsDeleted() {
		return nil, fmt.Errorf("cannot update deleted channel")
	}

	// 6. 更新通道
	if err := ch.Update(
		domainObjects.Name,
		domainObjects.Description,
		request.Enabled,
		domainObjects.ChannelType,
		domainObjects.TemplateID,
		domainObjects.CommonSettings,
		domainObjects.Config,
		domainObjects.Recipients,
		domainObjects.Tags,
	); err != nil {
		return nil, fmt.Errorf("failed to update channel: %w", err)
	}

	// 7. 持久化
	if err := uc.channelRepo.Update(ctx, ch); err != nil {
		return nil, fmt.Errorf("failed to save channel: %w", err)
	}

	// 8. 轉換為回應 DTO
	response := uc.convertToResponse(ch)
	return response, nil
}

// validateRequest 驗證請求參數
func (uc *UpdateChannelUseCase) validateRequest(channelID string, request *dtos.UpdateChannelRequest) error {
	if channelID == "" {
		return fmt.Errorf("channel ID is required")
	}

	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if request.ChannelName == "" {
		return fmt.Errorf("channel name is required")
	}

	if request.ChannelType == "" {
		return fmt.Errorf("channel type is required")
	}

	return nil
}

// convertToDomainObjects 轉換為領域物件
func (uc *UpdateChannelUseCase) convertToDomainObjects(request *dtos.UpdateChannelRequest) (*DomainObjects, error) {
	// 通道名稱
	name, err := channel.NewChannelName(request.ChannelName)
	if err != nil {
		return nil, fmt.Errorf("invalid channel name: %w", err)
	}

	// 描述
	description, err := channel.NewDescription(request.Description)
	if err != nil {
		return nil, fmt.Errorf("invalid description: %w", err)
	}

	// 通道類型
	channelType := shared.ChannelType(request.ChannelType)
	if !channelType.IsValid() {
		return nil, fmt.Errorf("invalid channel type: %s", request.ChannelType)
	}

	// 範本 ID
	var templateID *template.TemplateID
	if request.TemplateID != "" {
		templateID, err = template.NewTemplateIDFromString(request.TemplateID)
		if err != nil {
			return nil, fmt.Errorf("invalid template ID: %w", err)
		}
	}

	// 通用設定
	commonSettings, err := request.CommonSettings.ToCommonSettings()
	if err != nil {
		return nil, fmt.Errorf("invalid common settings: %w", err)
	}

	// 通道配置
	config := channel.NewChannelConfig(request.Config)

	// 收件人
	recipientSlice, err := dtos.ToRecipientsSlice(request.Recipients)
	if err != nil {
		return nil, fmt.Errorf("invalid recipients: %w", err)
	}
	recipients := channel.NewRecipients(recipientSlice)

	// 標籤
	tags := channel.NewTags(request.Tags)

	return &DomainObjects{
		Name:           name,
		Description:    description,
		ChannelType:    channelType,
		TemplateID:     templateID,
		CommonSettings: commonSettings,
		Config:         config,
		Recipients:     recipients,
		Tags:           tags,
	}, nil
}

// convertToResponse 轉換為回應 DTO
func (uc *UpdateChannelUseCase) convertToResponse(ch *channel.Channel) *dtos.ChannelResponse {
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