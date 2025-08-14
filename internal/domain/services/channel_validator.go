package services

import (
	"context"
	"errors"
	"fmt"

	"notification/internal/domain/channel"
	"notification/internal/domain/shared"
	"notification/internal/domain/template"
)

// ChannelValidator is the domain service for channel validation.
type ChannelValidator struct {
	channelRepo  channel.ChannelRepository
	templateRepo template.TemplateRepository
}

// NewChannelValidator creates a channel validation service.
func NewChannelValidator(
	channelRepo channel.ChannelRepository,
	templateRepo template.TemplateRepository,
) *ChannelValidator {
	return &ChannelValidator{
		channelRepo:  channelRepo,
		templateRepo: templateRepo,
	}
}

// ValidationError is a validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (ve *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", ve.Field, ve.Message)
}

// ValidationErrors is a list of validation errors.
type ValidationErrors []*ValidationError

// Error implements the error interface.
func (ves ValidationErrors) Error() string {
	if len(ves) == 0 {
		return "no validation errors"
	}
	if len(ves) == 1 {
		return ves[0].Error()
	}
	return fmt.Sprintf("multiple validation errors: %d errors", len(ves))
}

// HasErrors checks if there are any errors.
func (ves ValidationErrors) HasErrors() bool {
	return len(ves) > 0
}

// Add adds a validation error.
func (ves *ValidationErrors) Add(field, message string) {
	*ves = append(*ves, &ValidationError{
		Field:   field,
		Message: message,
	})
}

// ValidateChannelForCreation validates channel creation.
func (cv *ChannelValidator) ValidateChannelForCreation(
	ctx context.Context,
	name *channel.ChannelName,
	channelType shared.ChannelType,
	templateID *template.TemplateID,
	config *channel.ChannelConfig,
) error {
	var errors ValidationErrors

	// Validate channel name uniqueness
	if err := cv.validateChannelNameUniqueness(ctx, name); err != nil {
		errors.Add("channelName", err.Error())
	}

	// Validate template existence and type matching
	if err := cv.validateTemplateCompatibility(ctx, templateID, channelType); err != nil {
		errors.Add("templateId", err.Error())
	}

	// Validate channel configuration
	if err := cv.validateChannelConfig(channelType, config); err != nil {
		errors.Add("config", err.Error())
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// ValidateChannelForUpdate validates channel update.
func (cv *ChannelValidator) ValidateChannelForUpdate(
	ctx context.Context,
	channelID *channel.ChannelID,
	name *channel.ChannelName,
	channelType shared.ChannelType,
	templateID *template.TemplateID,
	config *channel.ChannelConfig,
) error {
	var errors ValidationErrors

	// Check if the channel exists
	existingChannel, err := cv.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		errors.Add("channelId", "channel not found")
		return errors
	}

	// Validate channel name uniqueness (excluding self)
	if !existingChannel.Name().Equals(name) {
		if err := cv.validateChannelNameUniqueness(ctx, name); err != nil {
			errors.Add("channelName", err.Error())
		}
	}

	// Validate template existence and type matching
	if err := cv.validateTemplateCompatibility(ctx, templateID, channelType); err != nil {
		errors.Add("templateId", err.Error())
	}

	// Validate channel configuration
	if err := cv.validateChannelConfig(channelType, config); err != nil {
		errors.Add("config", err.Error())
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// validateChannelNameUniqueness validates channel name uniqueness.
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

// validateTemplateCompatibility validates template compatibility.
func (cv *ChannelValidator) validateTemplateCompatibility(
	ctx context.Context,
	templateID *template.TemplateID,
	channelType shared.ChannelType,
) error {
	if templateID == nil {
		return nil // Template ID can be empty
	}

	// Check if the template exists
	tmpl, err := cv.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		return fmt.Errorf("template not found: %w", err)
	}

	// Check if the template type matches the channel type
	if !tmpl.MatchesType(channelType) {
		return fmt.Errorf("template type '%s' does not match channel type '%s'",
			tmpl.ChannelType(), channelType)
	}

	return nil
}

// validateChannelConfig validates channel configuration.
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

// validateEmailConfig validates email configuration.
func (cv *ChannelValidator) validateEmailConfig(config *channel.ChannelConfig) error {
	requiredFields := []string{"host", "port", "username", "password", "secure", "username", "password", "senderEmail"}

	for _, field := range requiredFields {
		if value, exists := config.Get(field); !exists || value == "" {
			return fmt.Errorf("email config missing required field: %s", field)
		}
	}

	// Validate if port is a valid number
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

// validateSlackConfig validates Slack configuration.
func (cv *ChannelValidator) validateSlackConfig(config *channel.ChannelConfig) error {
	requiredFields := []string{"token", "workspace"}

	for _, field := range requiredFields {
		if value, exists := config.Get(field); !exists || value == "" {
			return fmt.Errorf("slack config missing required field: %s", field)
		}
	}

	return nil
}

// validateSMSConfig validates SMS configuration.
func (cv *ChannelValidator) validateSMSConfig(config *channel.ChannelConfig) error {
	requiredFields := []string{"provider", "apiKey", "apiSecret"}

	for _, field := range requiredFields {
		if value, exists := config.Get(field); !exists || value == "" {
			return fmt.Errorf("sms config missing required field: %s", field)
		}
	}

	return nil
}

// ValidateChannelDeletion validates channel deletion.
func (cv *ChannelValidator) ValidateChannelDeletion(ctx context.Context, channelID *channel.ChannelID) error {
	// Check if the channel exists
	ch, err := cv.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		return fmt.Errorf("channel not found: %w", err)
	}

	// Check if the channel is already deleted
	if ch.IsDeleted() {
		return errors.New("channel is already deleted")
	}

	// In a real project, you may need to check here:
	// 1. Whether there are ongoing message sending tasks
	// 2. Whether there are other dependent resources
	// 3. Business rule restrictions

	return nil
}
