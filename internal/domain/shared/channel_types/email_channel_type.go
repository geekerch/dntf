package channel_types

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"notification/internal/domain/shared"
)

// EmailChannelType implements ChannelTypeDefinition for email channels
type EmailChannelType struct{}

// GetName returns the channel type name
func (e *EmailChannelType) GetName() string {
	return "email"
}

// GetDisplayName returns the display name
func (e *EmailChannelType) GetDisplayName() string {
	return "Email"
}

// GetDescription returns the description
func (e *EmailChannelType) GetDescription() string {
	return "Send notifications via email using SMTP"
}

// ValidateConfig validates the email channel configuration
func (e *EmailChannelType) ValidateConfig(config map[string]interface{}) error {
	if config == nil {
		return errors.New("email configuration cannot be nil")
	}

	// Validate SMTP host
	smtpHost, ok := config["smtp_host"].(string)
	if !ok || smtpHost == "" {
		return errors.New("smtp_host is required for email channel")
	}

	// Validate SMTP port
	smtpPortRaw, ok := config["smtp_port"]
	if !ok {
		return errors.New("smtp_port is required for email channel")
	}
	
	var smtpPort int
	switch v := smtpPortRaw.(type) {
	case int:
		smtpPort = v
	case float64:
		smtpPort = int(v)
	case string:
		var err error
		smtpPort, err = strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("invalid smtp_port format: %v", v)
		}
	default:
		return fmt.Errorf("invalid smtp_port type: %T", v)
	}
	
	if smtpPort <= 0 || smtpPort > 65535 {
		return fmt.Errorf("smtp_port must be between 1 and 65535, got: %d", smtpPort)
	}

	// Validate username
	username, ok := config["username"].(string)
	if !ok || username == "" {
		return errors.New("username is required for email channel")
	}

	// Validate password
	password, ok := config["password"].(string)
	if !ok || password == "" {
		return errors.New("password is required for email channel")
	}

	// Validate from email
	fromEmail, ok := config["from_email"].(string)
	if !ok || fromEmail == "" {
		return errors.New("from_email is required for email channel")
	}

	// Optional: Validate from name
	if fromName, exists := config["from_name"]; exists {
		if _, ok := fromName.(string); !ok {
			return errors.New("from_name must be a string")
		}
	}

	// Optional: Validate use_tls
	if useTLS, exists := config["use_tls"]; exists {
		if _, ok := useTLS.(bool); !ok {
			return errors.New("use_tls must be a boolean")
		}
	}

	return nil
}

// GetConfigSchema returns the configuration schema for email channels
func (e *EmailChannelType) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"smtp_host": map[string]interface{}{
				"type":        "string",
				"description": "SMTP server hostname",
				"example":     "smtp.gmail.com",
			},
			"smtp_port": map[string]interface{}{
				"type":        "integer",
				"description": "SMTP server port",
				"minimum":     1,
				"maximum":     65535,
				"example":     587,
			},
			"username": map[string]interface{}{
				"type":        "string",
				"description": "SMTP username",
				"example":     "user@example.com",
			},
			"password": map[string]interface{}{
				"type":        "string",
				"description": "SMTP password",
				"format":      "password",
			},
			"from_email": map[string]interface{}{
				"type":        "string",
				"description": "From email address",
				"format":      "email",
				"example":     "noreply@example.com",
			},
			"from_name": map[string]interface{}{
				"type":        "string",
				"description": "From name (optional)",
				"example":     "Notification System",
			},
			"use_tls": map[string]interface{}{
				"type":        "boolean",
				"description": "Use TLS encryption",
				"default":     true,
			},
		},
		"required": []string{"smtp_host", "smtp_port", "username", "password", "from_email"},
	}
}

// CreateMessageSender creates an email message sender
func (e *EmailChannelType) CreateMessageSender(timeout time.Duration) (interface{}, error) {
	// Return a factory identifier that infrastructure layer can use
	return "email_service", nil
}

// NewEmailChannelType creates a new email channel type definition
func NewEmailChannelType() shared.ChannelTypeDefinition {
	return &EmailChannelType{}
}