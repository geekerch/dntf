package plugins

import (
	"time"
	
	"notification/internal/domain/shared"
	publicPlugins "notification/pkg/plugins"
)

// PublicPluginAdapter adapts public plugin API to internal plugin interface
type PublicPluginAdapter struct {
	publicPlugin publicPlugins.Plugin
}

// NewPublicPluginAdapter creates a new adapter for public plugins
func NewPublicPluginAdapter(publicPlugin publicPlugins.Plugin) *PublicPluginAdapter {
	return &PublicPluginAdapter{
		publicPlugin: publicPlugin,
	}
}

// GetInfo implements internal Plugin interface
func (a *PublicPluginAdapter) GetInfo() PluginInfo {
	publicInfo := a.publicPlugin.GetInfo()
	return PluginInfo{
		Name:        publicInfo.Name,
		Version:     publicInfo.Version,
		Description: publicInfo.Description,
		Author:      publicInfo.Author,
		LoadedAt:    publicInfo.LoadedAt,
	}
}

// GetChannelType implements internal Plugin interface
func (a *PublicPluginAdapter) GetChannelType() shared.ChannelTypeDefinition {
	publicChannelType := a.publicPlugin.GetChannelType()
	return &PublicChannelTypeAdapter{
		publicChannelType: publicChannelType,
	}
}

// Initialize implements internal Plugin interface
func (a *PublicPluginAdapter) Initialize(config map[string]interface{}) error {
	return a.publicPlugin.Initialize(config)
}

// Cleanup implements internal Plugin interface
func (a *PublicPluginAdapter) Cleanup() error {
	return a.publicPlugin.Cleanup()
}

// PublicChannelTypeAdapter adapts public channel type API to internal interface
type PublicChannelTypeAdapter struct {
	publicChannelType publicPlugins.ChannelTypeDefinition
}

// GetName implements internal ChannelTypeDefinition interface
func (a *PublicChannelTypeAdapter) GetName() string {
	return a.publicChannelType.GetName()
}

// GetDisplayName implements internal ChannelTypeDefinition interface
func (a *PublicChannelTypeAdapter) GetDisplayName() string {
	return a.publicChannelType.GetDisplayName()
}

// GetDescription implements internal ChannelTypeDefinition interface
func (a *PublicChannelTypeAdapter) GetDescription() string {
	return a.publicChannelType.GetDescription()
}

// ValidateConfig implements internal ChannelTypeDefinition interface
func (a *PublicChannelTypeAdapter) ValidateConfig(config map[string]interface{}) error {
	return a.publicChannelType.ValidateConfig(config)
}

// GetConfigSchema implements internal ChannelTypeDefinition interface
func (a *PublicChannelTypeAdapter) GetConfigSchema() map[string]interface{} {
	return a.publicChannelType.GetConfigSchema()
}

// CreateMessageSender implements internal ChannelTypeDefinition interface
func (a *PublicChannelTypeAdapter) CreateMessageSender(timeout time.Duration) (interface{}, error) {
	return a.publicChannelType.CreateMessageSender(timeout)
}