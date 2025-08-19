package shared

import (
	"fmt"
	"log"
	"sync"
	"time"
)

var initOnce sync.Once

// InitializeChannelTypes initializes all default channel types
// This should be called once during application startup
func InitializeChannelTypes() {
	initOnce.Do(func() {
		registerDefaultChannelTypes()
	})
}

// MustInitializeChannelTypes initializes all default channel types and panics on error
// This should be called once during application startup
func MustInitializeChannelTypes() {
	initOnce.Do(func() {
		mustRegisterDefaultChannelTypes()
	})
}

// registerDefaultChannelTypes registers all default channel types
func registerDefaultChannelTypes() {
	registry := GetChannelTypeRegistry()
	
	// Register email channel type
	if err := registry.RegisterChannelType(newEmailChannelType()); err != nil {
		log.Printf("Warning: Failed to register email channel type: %v", err)
	}
	
	// Register Slack channel type
	if err := registry.RegisterChannelType(newSlackChannelType()); err != nil {
		log.Printf("Warning: Failed to register slack channel type: %v", err)
	}
	
	// Register SMS channel type
	if err := registry.RegisterChannelType(newSMSChannelType()); err != nil {
		log.Printf("Warning: Failed to register sms channel type: %v", err)
	}
}

// mustRegisterDefaultChannelTypes registers all default channel types and panics on error
func mustRegisterDefaultChannelTypes() {
	registry := GetChannelTypeRegistry()
	
	// Register email channel type
	if err := registry.RegisterChannelType(newEmailChannelType()); err != nil {
		panic("Failed to register email channel type: " + err.Error())
	}
	
	// Register Slack channel type
	if err := registry.RegisterChannelType(newSlackChannelType()); err != nil {
		panic("Failed to register slack channel type: " + err.Error())
	}
	
	// Register SMS channel type
	if err := registry.RegisterChannelType(newSMSChannelType()); err != nil {
		panic("Failed to register sms channel type: " + err.Error())
	}
}

// Built-in channel type implementations to avoid circular imports

// emailChannelType implements ChannelTypeDefinition for email channels
type emailChannelType struct{}

func (e *emailChannelType) GetName() string { return "email" }
func (e *emailChannelType) GetDisplayName() string { return "Email" }
func (e *emailChannelType) GetDescription() string { return "Send notifications via email using SMTP" }

func (e *emailChannelType) ValidateConfig(config map[string]interface{}) error {
	// Basic validation - can be enhanced
	if config == nil {
		return fmt.Errorf("email configuration cannot be nil")
	}
	return nil
}

func (e *emailChannelType) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"smtp_host": map[string]interface{}{"type": "string"},
			"smtp_port": map[string]interface{}{"type": "integer"},
			"username":  map[string]interface{}{"type": "string"},
			"password":  map[string]interface{}{"type": "string"},
			"from_email": map[string]interface{}{"type": "string"},
		},
		"required": []string{"smtp_host", "smtp_port", "username", "password", "from_email"},
	}
}

func (e *emailChannelType) CreateMessageSender(timeout time.Duration) (interface{}, error) {
	// Return a factory function that can be used by infrastructure layer
	return func() interface{} {
		// This will be handled by the infrastructure layer
		return "email_service_factory"
	}, nil
}

func newEmailChannelType() ChannelTypeDefinition {
	return &emailChannelType{}
}

// slackChannelType implements ChannelTypeDefinition for Slack channels
type slackChannelType struct{}

func (s *slackChannelType) GetName() string { return "slack" }
func (s *slackChannelType) GetDisplayName() string { return "Slack" }
func (s *slackChannelType) GetDescription() string { return "Send notifications to Slack channels via webhook" }

func (s *slackChannelType) ValidateConfig(config map[string]interface{}) error {
	if config == nil {
		return fmt.Errorf("slack configuration cannot be nil")
	}
	return nil
}

func (s *slackChannelType) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"webhook_url": map[string]interface{}{"type": "string"},
		},
		"required": []string{"webhook_url"},
	}
}

func (s *slackChannelType) CreateMessageSender(timeout time.Duration) (interface{}, error) {
	// Return a factory function that can be used by infrastructure layer
	return func() interface{} {
		// This will be handled by the infrastructure layer
		return "slack_service_factory"
	}, nil
}

func newSlackChannelType() ChannelTypeDefinition {
	return &slackChannelType{}
}

// smsChannelType implements ChannelTypeDefinition for SMS channels
type smsChannelType struct{}

func (s *smsChannelType) GetName() string { return "sms" }
func (s *smsChannelType) GetDisplayName() string { return "SMS" }
func (s *smsChannelType) GetDescription() string { return "Send notifications via SMS" }

func (s *smsChannelType) ValidateConfig(config map[string]interface{}) error {
	if config == nil {
		return fmt.Errorf("sms configuration cannot be nil")
	}
	return nil
}

func (s *smsChannelType) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"provider": map[string]interface{}{"type": "string"},
		},
		"required": []string{"provider"},
	}
}

func (s *smsChannelType) CreateMessageSender(timeout time.Duration) (interface{}, error) {
	// Return a factory function that can be used by infrastructure layer
	return func() interface{} {
		// This will be handled by the infrastructure layer
		return "sms_service_factory"
	}, nil
}

func newSMSChannelType() ChannelTypeDefinition {
	return &smsChannelType{}
}