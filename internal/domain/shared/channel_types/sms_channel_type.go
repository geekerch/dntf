package channel_types

import (
	"errors"
	"time"

	"notification/internal/domain/shared"
)

// SMSChannelType implements ChannelTypeDefinition for SMS channels
type SMSChannelType struct{}

// GetName returns the channel type name
func (s *SMSChannelType) GetName() string {
	return "sms"
}

// GetDisplayName returns the display name
func (s *SMSChannelType) GetDisplayName() string {
	return "SMS"
}

// GetDescription returns the description
func (s *SMSChannelType) GetDescription() string {
	return "Send notifications via SMS using Twilio or other SMS providers"
}

// ValidateConfig validates the SMS channel configuration
func (s *SMSChannelType) ValidateConfig(config map[string]interface{}) error {
	if config == nil {
		return errors.New("sms configuration cannot be nil")
	}

	// Validate provider
	provider, ok := config["provider"].(string)
	if !ok || provider == "" {
		return errors.New("provider is required for sms channel")
	}

	switch provider {
	case "twilio":
		return s.validateTwilioConfig(config)
	default:
		return errors.New("unsupported SMS provider: " + provider)
	}
}

// validateTwilioConfig validates Twilio-specific configuration
func (s *SMSChannelType) validateTwilioConfig(config map[string]interface{}) error {
	// Validate account SID
	accountSID, ok := config["account_sid"].(string)
	if !ok || accountSID == "" {
		return errors.New("account_sid is required for Twilio SMS")
	}

	// Validate auth token
	authToken, ok := config["auth_token"].(string)
	if !ok || authToken == "" {
		return errors.New("auth_token is required for Twilio SMS")
	}

	// Validate from number
	fromNumber, ok := config["from_number"].(string)
	if !ok || fromNumber == "" {
		return errors.New("from_number is required for Twilio SMS")
	}

	return nil
}

// GetConfigSchema returns the configuration schema for SMS channels
func (s *SMSChannelType) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"provider": map[string]interface{}{
				"type":        "string",
				"description": "SMS provider",
				"enum":        []string{"twilio"},
				"example":     "twilio",
			},
			"account_sid": map[string]interface{}{
				"type":        "string",
				"description": "Twilio Account SID",
				"example":     "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			},
			"auth_token": map[string]interface{}{
				"type":        "string",
				"description": "Twilio Auth Token",
				"format":      "password",
			},
			"from_number": map[string]interface{}{
				"type":        "string",
				"description": "From phone number",
				"example":     "+1234567890",
			},
		},
		"required": []string{"provider", "account_sid", "auth_token", "from_number"},
		"if": map[string]interface{}{
			"properties": map[string]interface{}{
				"provider": map[string]interface{}{
					"const": "twilio",
				},
			},
		},
		"then": map[string]interface{}{
			"required": []string{"account_sid", "auth_token", "from_number"},
		},
	}
}

// CreateMessageSender creates an SMS message sender
func (s *SMSChannelType) CreateMessageSender(timeout time.Duration) (interface{}, error) {
	// Return a factory identifier that infrastructure layer can use
	return "sms_service", nil
}

// NewSMSChannelType creates a new SMS channel type definition
func NewSMSChannelType() shared.ChannelTypeDefinition {
	return &SMSChannelType{}
}