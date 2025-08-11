package services

import (
	"context"
	"errors"
	"fmt"

	"channel-api/internal/domain/channel"
	"channel-api/internal/domain/shared"
	"channel-api/internal/domain/template"
)

// ChannelValidator 通道驗證領域服務
type ChannelValidator struct {
	channelRepo  channel.ChannelRepository
	templateRepo template.TemplateRepository
}

// NewChannelValidator 建立通道驗證服務
func NewChannelValidator(
	channelRepo channel.ChannelRepository,
	templateRepo template.TemplateRepository,
) *ChannelValidator {
	return &ChannelValidator{
		channelRepo:  channelRepo,
		templateRepo: templateRepo,
	}
}

// ValidationError 驗證錯誤
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error 實作 error 介面
func (ve *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", ve.Field, ve.Message)
}

// ValidationErrors 驗證錯誤列表
type ValidationErrors []*ValidationError

// Error 實作 error 介面
func (ves ValidationErrors) Error() string {
	if len(ves) == 0 {
		return "no validation errors"
	}
	if len(ves) == 1 {
		return ves[0].Error()
	}
	return fmt.Sprintf("multiple validation errors: %d errors", len(ves))
}

// HasErrors 檢查是否有錯誤
func (ves ValidationErrors) HasErrors() bool {
	return len(ves) > 0
}

// Add 新增驗證錯誤
func (ves *ValidationErrors) Add(field, message string) {
	*ves = append(*ves, &ValidationError{
		Field:   field,
		Message: message,
	})
}

// ValidateChannelForCreation 驗證通道建立
func (cv *ChannelValidator) ValidateChannelForCreation(
	ctx context.Context,
	name *channel.ChannelName,
	channelType shared.ChannelType,
	templateID *template.TemplateID,
	config *channel.ChannelConfig,
) error {
	var errors ValidationErrors

	// 驗證通道名稱唯一性
	if err := cv.validateChannelNameUniqueness(ctx, name); err != nil {
		errors.Add("channelName", err.Error())
	}

	// 驗證範本存在性和類型匹配
	if err := cv.validateTemplateCompatibility(ctx, templateID, channelType); err != nil {
		errors.Add("templateId", err.Error())
	}

	// 驗證通道配置
	if err := cv.validateChannelConfig(channelType, config); err != nil {
		errors.Add("config", err.Error())
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// ValidateChannelForUpdate 驗證通道更新
func (cv *ChannelValidator) ValidateChannelForUpdate(
	ctx context.Context,
	channelID *channel.ChannelID,
	name *channel.ChannelName,
	channelType shared.ChannelType,
	templateID *template.TemplateID,
	config *channel.ChannelConfig,
) error {
	var errors ValidationErrors

	// 檢查通道是否存在
	existingChannel, err := cv.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		errors.Add("channelId", "channel not found")
		return errors
	}

	// 驗證通道名稱唯一性 (排除自己)
	if !existingChannel.Name().Equals(name) {
		if err := cv.validateChannelNameUniqueness(ctx, name); err != nil {
			errors.Add("channelName", err.Error())
		}
	}

	// 驗證範本存在性和類型匹配
	if err := cv.validateTemplateCompatibility(ctx, templateID, channelType); err != nil {
		errors.Add("templateId", err.Error())
	}

	// 驗證通道配置
	if err := cv.validateChannelConfig(channelType, config); err != nil {
		errors.Add("config", err.Error())
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// validateChannelNameUniqueness 驗證通道名稱唯一性
func (cv *ChannelValidator) validateChannelNameUniqueness(ctx context.Context, name *channel.ChannelName) error {
	exists, err := cv.channelRepo.ExistsByName(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to check channel name uniqueness: %w", err)
	}
	if exists {
		return errors.New("channel name already exists")
	}
	return nil
}

// validateTemplateCompatibility 驗證範本相容性
func (cv *ChannelValidator) validateTemplateCompatibility(
	ctx context.Context,
	templateID *template.TemplateID,
	channelType shared.ChannelType,
) error {
	if templateID == nil {
		return nil // 範本 ID 可以為空
	}

	// 檢查範本是否存在
	tmpl, err := cv.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		return fmt.Errorf("template not found: %w", err)
	}

	// 檢查範本類型是否匹配通道類型
	if !tmpl.MatchesType(channelType) {
		return fmt.Errorf("template type '%s' does not match channel type '%s'", 
			tmpl.ChannelType(), channelType)
	}

	return nil
}

// validateChannelConfig 驗證通道配置
func (cv *ChannelValidator) validateChannelConfig(channelType shared.ChannelType, config *channel.ChannelConfig) error {
	if config == nil {
		return errors.New("channel config is required")
	}

	switch channelType {
	case shared.ChannelTypeEmail:
		return cv.validateEmailConfig(config)
	case shared.ChannelTypeSlack:
		return cv.validateSlackConfig(config)
	case shared.ChannelTypeSMS:
		return cv.validateSMSConfig(config)
	default:
		return fmt.Errorf("unsupported channel type: %s", channelType)
	}
}

// validateEmailConfig 驗證電子郵件配置
func (cv *ChannelValidator) validateEmailConfig(config *channel.ChannelConfig) error {
	requiredFields := []string{"host", "port", "username", "password"}
	
	for _, field := range requiredFields {
		if value, exists := config.Get(field); !exists || value == "" {
			return fmt.Errorf("email config missing required field: %s", field)
		}
	}

	// 驗證 port 是否為有效數字
	if port, exists := config.Get("port"); exists {
		switch v := port.(type) {
		case float64:
			if v <= 0 || v > 65535 {
				return errors.New("email config port must be between 1 and 65535")
			}
		case int:
			if v <= 0 || v > 65535 {
				return errors.New("email config port must be between 1 and 65535")
			}
		default:
			return errors.New("email config port must be a number")
		}
	}

	return nil
}

// validateSlackConfig 驗證 Slack 配置
func (cv *ChannelValidator) validateSlackConfig(config *channel.ChannelConfig) error {
	requiredFields := []string{"token", "workspace"}
	
	for _, field := range requiredFields {
		if value, exists := config.Get(field); !exists || value == "" {
			return fmt.Errorf("slack config missing required field: %s", field)
		}
	}

	return nil
}

// validateSMSConfig 驗證簡訊配置
func (cv *ChannelValidator) validateSMSConfig(config *channel.ChannelConfig) error {
	requiredFields := []string{"provider", "apiKey", "apiSecret"}
	
	for _, field := range requiredFields {
		if value, exists := config.Get(field); !exists || value == "" {
			return fmt.Errorf("sms config missing required field: %s", field)
		}
	}

	return nil
}

// ValidateChannelDeletion 驗證通道刪除
func (cv *ChannelValidator) ValidateChannelDeletion(ctx context.Context, channelID *channel.ChannelID) error {
	// 檢查通道是否存在
	ch, err := cv.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		return fmt.Errorf("channel not found: %w", err)
	}

	// 檢查通道是否已刪除
	if ch.IsDeleted() {
		return errors.New("channel is already deleted")
	}

	// 在實際專案中，這裡可能需要檢查：
	// 1. 是否有進行中的訊息發送任務
	// 2. 是否有相依的其他資源
	// 3. 業務規則限制

	return nil
}