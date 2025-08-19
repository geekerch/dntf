package examples

import (
	"context"
	"errors"
	"fmt"
	"time"

	"notification/internal/domain/shared"
)

// DiscordChannelType implements ChannelTypeDefinition for Discord channels
// This is an example of how to implement a custom channel type
type DiscordChannelType struct{}

// GetName returns the channel type name
func (d *DiscordChannelType) GetName() string {
	return "discord"
}

// GetDisplayName returns the display name
func (d *DiscordChannelType) GetDisplayName() string {
	return "Discord"
}

// GetDescription returns the description
func (d *DiscordChannelType) GetDescription() string {
	return "Send notifications to Discord channels via webhook"
}

// ValidateConfig validates the Discord channel configuration
func (d *DiscordChannelType) ValidateConfig(config map[string]interface{}) error {
	if config == nil {
		return errors.New("discord configuration cannot be nil")
	}

	// Validate webhook URL
	webhookURL, ok := config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return errors.New("webhook_url is required for discord channel")
	}

	// Optional: Validate username
	if username, exists := config["username"]; exists {
		if _, ok := username.(string); !ok {
			return errors.New("username must be a string")
		}
	}

	// Optional: Validate avatar URL
	if avatarURL, exists := config["avatar_url"]; exists {
		if _, ok := avatarURL.(string); !ok {
			return errors.New("avatar_url must be a string")
		}
	}

	// Optional: Validate TTS
	if tts, exists := config["tts"]; exists {
		if _, ok := tts.(bool); !ok {
			return errors.New("tts must be a boolean")
		}
	}

	return nil
}

// GetConfigSchema returns the configuration schema for Discord channels
func (d *DiscordChannelType) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"webhook_url": map[string]interface{}{
				"type":        "string",
				"description": "Discord webhook URL",
				"format":      "uri",
				"example":     "https://discord.com/api/webhooks/123456789/abcdefghijklmnop",
			},
			"username": map[string]interface{}{
				"type":        "string",
				"description": "Bot username (optional)",
				"example":     "NotificationBot",
			},
			"avatar_url": map[string]interface{}{
				"type":        "string",
				"description": "Bot avatar URL (optional)",
				"format":      "uri",
			},
			"tts": map[string]interface{}{
				"type":        "boolean",
				"description": "Text-to-speech (optional)",
				"default":     false,
			},
		},
		"required": []string{"webhook_url"},
	}
}

// CreateMessageSender creates a Discord message sender
func (d *DiscordChannelType) CreateMessageSender(timeout time.Duration) (interface{}, error) {
	return NewDiscordService(timeout), nil
}

// NewDiscordChannelType creates a new Discord channel type definition
func NewDiscordChannelType() shared.ChannelTypeDefinition {
	return &DiscordChannelType{}
}

// DiscordService implements MessageSender for Discord channels
// This is an example implementation
type DiscordService struct {
	timeout time.Duration
}

// NewDiscordService creates a new Discord service
func NewDiscordService(timeout time.Duration) *DiscordService {
	return &DiscordService{
		timeout: timeout,
	}
}

// Send sends a message to Discord via webhook
func (s *DiscordService) Send(ctx context.Context, ch interface{}, content interface{}) error {
	// TODO: Implement actual Discord webhook sending logic
	// This is just a placeholder implementation
	fmt.Printf("Sending Discord message via webhook\n")
	return nil
}

// GetChannelType returns the channel type this sender supports
func (s *DiscordService) GetChannelType() string {
	return "discord"
}

// ValidateConfig validates the channel configuration for this sender
func (s *DiscordService) ValidateConfig(config interface{}) error {
	// TODO: Implement proper config validation
	// This is just a placeholder
	return nil
}

// RegisterDiscordChannelType registers the Discord channel type
// Call this function to add Discord support to your application
func RegisterDiscordChannelType() error {
	registry := shared.GetChannelTypeRegistry()
	return registry.RegisterChannelType(NewDiscordChannelType())
}