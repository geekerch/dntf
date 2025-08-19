package channel_types

import (
	"errors"
	"time"

	"notification/internal/domain/shared"
)

// SlackChannelType implements ChannelTypeDefinition for Slack channels
type SlackChannelType struct{}

// GetName returns the channel type name
func (s *SlackChannelType) GetName() string {
	return "slack"
}

// GetDisplayName returns the display name
func (s *SlackChannelType) GetDisplayName() string {
	return "Slack"
}

// GetDescription returns the description
func (s *SlackChannelType) GetDescription() string {
	return "Send notifications to Slack channels via webhook"
}

// ValidateConfig validates the Slack channel configuration
func (s *SlackChannelType) ValidateConfig(config map[string]interface{}) error {
	if config == nil {
		return errors.New("slack configuration cannot be nil")
	}

	// Validate webhook URL
	webhookURL, ok := config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return errors.New("webhook_url is required for slack channel")
	}

	// Optional: Validate channel name
	if channel, exists := config["channel"]; exists {
		if _, ok := channel.(string); !ok {
			return errors.New("channel must be a string")
		}
	}

	// Optional: Validate username
	if username, exists := config["username"]; exists {
		if _, ok := username.(string); !ok {
			return errors.New("username must be a string")
		}
	}

	// Optional: Validate icon emoji
	if iconEmoji, exists := config["icon_emoji"]; exists {
		if _, ok := iconEmoji.(string); !ok {
			return errors.New("icon_emoji must be a string")
		}
	}

	// Optional: Validate icon URL
	if iconURL, exists := config["icon_url"]; exists {
		if _, ok := iconURL.(string); !ok {
			return errors.New("icon_url must be a string")
		}
	}

	return nil
}

// GetConfigSchema returns the configuration schema for Slack channels
func (s *SlackChannelType) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"webhook_url": map[string]interface{}{
				"type":        "string",
				"description": "Slack webhook URL",
				"format":      "uri",
				"example":     "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
			},
			"channel": map[string]interface{}{
				"type":        "string",
				"description": "Slack channel name (optional, overrides webhook default)",
				"example":     "#general",
			},
			"username": map[string]interface{}{
				"type":        "string",
				"description": "Bot username (optional)",
				"example":     "NotificationBot",
			},
			"icon_emoji": map[string]interface{}{
				"type":        "string",
				"description": "Bot icon emoji (optional)",
				"example":     ":bell:",
			},
			"icon_url": map[string]interface{}{
				"type":        "string",
				"description": "Bot icon URL (optional)",
				"format":      "uri",
			},
		},
		"required": []string{"webhook_url"},
	}
}

// CreateMessageSender creates a Slack message sender
func (s *SlackChannelType) CreateMessageSender(timeout time.Duration) (interface{}, error) {
	// Return a factory identifier that infrastructure layer can use
	return "slack_service", nil
}

// NewSlackChannelType creates a new Slack channel type definition
func NewSlackChannelType() shared.ChannelTypeDefinition {
	return &SlackChannelType{}
}