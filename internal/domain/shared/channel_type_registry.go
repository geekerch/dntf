package shared

import (
	"fmt"
	"sync"
	"time"
)

// ChannelTypeDefinition defines the interface for a channel type
type ChannelTypeDefinition interface {
	// GetName returns the channel type name
	GetName() string
	
	// GetDisplayName returns the display name
	GetDisplayName() string
	
	// GetDescription returns the description
	GetDescription() string
	
	// ValidateConfig validates the channel configuration
	ValidateConfig(config map[string]interface{}) error
	
	// GetConfigSchema returns the configuration schema
	GetConfigSchema() map[string]interface{}
	
	// CreateMessageSender creates the corresponding message sender
	CreateMessageSender(timeout time.Duration) (interface{}, error)
}

// ChannelTypeRegistry manages all registered channel types
type ChannelTypeRegistry interface {
	// RegisterChannelType registers a new channel type
	RegisterChannelType(channelType ChannelTypeDefinition) error
	
	// GetChannelType gets the specified channel type definition
	GetChannelType(name string) (ChannelTypeDefinition, error)
	
	// GetAllChannelTypes gets all registered channel types
	GetAllChannelTypes() []ChannelTypeDefinition
	
	// IsValidChannelType checks if the channel type is valid
	IsValidChannelType(name string) bool
	
	// GetSupportedTypes returns all supported channel type names
	GetSupportedTypes() []string
}

// DefaultChannelTypeRegistry implements ChannelTypeRegistry
type DefaultChannelTypeRegistry struct {
	channelTypes map[string]ChannelTypeDefinition
	mutex        sync.RWMutex
}

// NewDefaultChannelTypeRegistry creates a new channel type registry
func NewDefaultChannelTypeRegistry() *DefaultChannelTypeRegistry {
	return &DefaultChannelTypeRegistry{
		channelTypes: make(map[string]ChannelTypeDefinition),
	}
}

// RegisterChannelType registers a new channel type
func (r *DefaultChannelTypeRegistry) RegisterChannelType(channelType ChannelTypeDefinition) error {
	if channelType == nil {
		return fmt.Errorf("channel type definition cannot be nil")
	}
	
	name := channelType.GetName()
	if name == "" {
		return fmt.Errorf("channel type name cannot be empty")
	}
	
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.channelTypes[name]; exists {
		return fmt.Errorf("channel type '%s' is already registered", name)
	}
	
	r.channelTypes[name] = channelType
	return nil
}

// GetChannelType gets the specified channel type definition
func (r *DefaultChannelTypeRegistry) GetChannelType(name string) (ChannelTypeDefinition, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	channelType, exists := r.channelTypes[name]
	if !exists {
		return nil, fmt.Errorf("channel type '%s' is not registered", name)
	}
	
	return channelType, nil
}

// GetAllChannelTypes gets all registered channel types
func (r *DefaultChannelTypeRegistry) GetAllChannelTypes() []ChannelTypeDefinition {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	types := make([]ChannelTypeDefinition, 0, len(r.channelTypes))
	for _, channelType := range r.channelTypes {
		types = append(types, channelType)
	}
	
	return types
}

// IsValidChannelType checks if the channel type is valid
func (r *DefaultChannelTypeRegistry) IsValidChannelType(name string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	_, exists := r.channelTypes[name]
	return exists
}

// GetSupportedTypes returns all supported channel type names
func (r *DefaultChannelTypeRegistry) GetSupportedTypes() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	types := make([]string, 0, len(r.channelTypes))
	for name := range r.channelTypes {
		types = append(types, name)
	}
	
	return types
}

// Global registry instance
var globalRegistry ChannelTypeRegistry
var registryOnce sync.Once

// GetChannelTypeRegistry returns the global channel type registry
func GetChannelTypeRegistry() ChannelTypeRegistry {
	registryOnce.Do(func() {
		globalRegistry = NewDefaultChannelTypeRegistry()
	})
	return globalRegistry
}

// SetChannelTypeRegistry sets the global channel type registry (for testing)
func SetChannelTypeRegistry(registry ChannelTypeRegistry) {
	globalRegistry = registry
}