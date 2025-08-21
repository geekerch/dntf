package plugins

import (
	"time"
)

// PluginInfo contains plugin metadata
type PluginInfo struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	LoadedAt    time.Time `json:"loadedAt"`
}

// Plugin represents a loaded plugin instance
type Plugin interface {
	// GetInfo returns plugin information
	GetInfo() PluginInfo
	
	// GetChannelType returns the channel type definition
	GetChannelType() ChannelTypeDefinition
	
	// Initialize initializes the plugin with configuration
	Initialize(config map[string]interface{}) error
	
	// Cleanup cleans up plugin resources
	Cleanup() error
}

// ChannelTypeDefinition defines the interface for channel types
type ChannelTypeDefinition interface {
	GetName() string
	GetDisplayName() string
	GetDescription() string
	ValidateConfig(config map[string]interface{}) error
	GetConfigSchema() map[string]interface{}
	CreateMessageSender(timeout time.Duration) (interface{}, error)
}