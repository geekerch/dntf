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

// CreateChannelUseCase 建立通道用例
type CreateChannelUseCase struct {
	channelRepo channel.ChannelRepository
	validator   *services.ChannelValidator
}

// NewCreateChannelUseCase 建立用例實例
func NewCreateChannelUseCase(
	channelRepo channel.ChannelRepository,
	validator *services.ChannelValidator,
) *CreateChannelUseCase {
	return &CreateChannelUseCase{
		channelRepo: channelRepo,
		validator:   validator,
	}
}

// Execute 執行建立通道
func (uc *CreateChannelUseCase) Execute(ctx context.Context, request *dtos.CreateChannelRequest) (*dtos.ChannelResponse, error) {
	// 1. 驗證輸入參數
	if err := uc.validateRequest(request); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// 2. 轉換為領域物件
	domainObjects, err := uc.convertToDomainObjects(request)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to domain objects: %w", err)
	}

	// 3. 業務驗證
	if err := uc.validator.ValidateChannelForCreation(
		ctx,
		domainObjects.Name,
		domainObjects.ChannelType,
		domainObjects.TemplateID,
		domainObjects.Config,
	); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 4. 建立通道實體
	ch, err := channel.NewChannel(
		domainObjects.Name,
		domainObjects.Description,
		request.Enabled,
		domainObjects.ChannelType,
		domainObjects.TemplateID,
		domainObjects.CommonSettings,
		domainObjects.Config,
		domainObjects.Recipients,
		domainObjects.Tags,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	// 5. 持久化
	if err := uc.channelRepo.Save(ctx, ch); err != nil {
		return nil, fmt.Errorf("failed to save channel: %w", err)
	}

	// 6. 轉換為回應 DTO
	response := uc.convertToResponse(ch)
	return response, nil
}

// validateRequest 驗證請求參數
func (uc *CreateChannelUseCase) validateRequest(request *dtos.CreateChannelRequest) error {
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

// DomainObjects 轉換後的領域物件
type DomainObjects struct {
	Name           *channel.ChannelName
	Description    *channel.Description
	ChannelType    shared.ChannelType
	TemplateID     *template.TemplateID
	CommonSettings *shared.CommonSettings
	Config         *channel.ChannelConfig
	Recipients     *channel.Recipients
	Tags           *channel.Tags
}

// convertToDomainObjects 轉換為領域物件
func (uc *CreateChannelUseCase) convertToDomainObjects(request *dtos.CreateChannelRequest) (*DomainObjects, error) {
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
func (uc *CreateChannelUseCase) convertToResponse(ch *channel.Channel) *dtos.ChannelResponse {
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