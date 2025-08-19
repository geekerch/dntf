package channel_types

import (
	"log"

	"notification/internal/domain/shared"
)

// RegisterDefaultChannelTypes registers all default channel types
func RegisterDefaultChannelTypes() {
	registry := shared.GetChannelTypeRegistry()
	
	// Register email channel type
	if err := registry.RegisterChannelType(NewEmailChannelType()); err != nil {
		log.Printf("Warning: Failed to register email channel type: %v", err)
	}
	
	// Register Slack channel type
	if err := registry.RegisterChannelType(NewSlackChannelType()); err != nil {
		log.Printf("Warning: Failed to register slack channel type: %v", err)
	}
	
	// Register SMS channel type
	if err := registry.RegisterChannelType(NewSMSChannelType()); err != nil {
		log.Printf("Warning: Failed to register sms channel type: %v", err)
	}
}

// MustRegisterDefaultChannelTypes registers all default channel types and panics on error
func MustRegisterDefaultChannelTypes() {
	registry := shared.GetChannelTypeRegistry()
	
	// Register email channel type
	if err := registry.RegisterChannelType(NewEmailChannelType()); err != nil {
		panic("Failed to register email channel type: " + err.Error())
	}
	
	// Register Slack channel type
	if err := registry.RegisterChannelType(NewSlackChannelType()); err != nil {
		panic("Failed to register slack channel type: " + err.Error())
	}
	
	// Register SMS channel type
	if err := registry.RegisterChannelType(NewSMSChannelType()); err != nil {
		panic("Failed to register sms channel type: " + err.Error())
	}
}