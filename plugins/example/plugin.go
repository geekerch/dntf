package main

import (
	"context"
	"fmt"
	"time"

	"notification/internal/domain/shared"
)

// ExamplePlugin demonstrates how to create a plugin
type ExamplePlugin struct {
	config map[string]interface{}
}

func (p *ExamplePlugin) GetInfo() PluginInfo {
	return PluginInfo{
		Name:        "example",
		Version:     "1.0.0",
		Description: "Example plugin for demonstration",
		Author:      "System",
		LoadedAt:    time.Now(),
	}
}

func (p *ExamplePlugin) GetChannelType() shared.ChannelTypeDefinition {
	return &ExampleChannelType{}
}

func (p *ExamplePlugin) Initialize(config map[string]interface{}) error {
	p.config = config
	fmt.Println("Example plugin initialized")
	return nil
}

func (p *ExamplePlugin) Cleanup() error {
	fmt.Println("Example plugin cleaned up")
	return nil
}

// ExampleChannelType implements ChannelTypeDefinition
type ExampleChannelType struct{}

func (e *ExampleChannelType) GetName() string {
	return "example"
}

func (e *ExampleChannelType) GetDisplayName() string {
	return "Example Channel"
}

func (e *ExampleChannelType) GetDescription() string {
	return "Example channel type for demonstration purposes"
}

func (e *ExampleChannelType) ValidateConfig(config map[string]interface{}) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	
	// Example validation
	if url, ok := config["url"].(string); !ok || url == "" {
		return fmt.Errorf("url is required")
	}
	
	return nil
}

func (e *ExampleChannelType) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url": map[string]interface{}{
				"type":        "string",
				"description": "Example URL",
				"example":     "https://example.com/webhook",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Timeout in seconds",
				"default":     30,
			},
		},
		"required": []string{"url"},
	}
}

func (e *ExampleChannelType) CreateMessageSender(timeout time.Duration) (interface{}, error) {
	return &ExampleSender{timeout: timeout}, nil
}

// ExampleSender implements message sending
type ExampleSender struct {
	timeout time.Duration
}

func (s *ExampleSender) Send(ctx context.Context, ch interface{}, content interface{}) error {
	fmt.Printf("Example sender: sending message with timeout %v\n", s.timeout)
	fmt.Printf("Channel: %+v\n", ch)
	fmt.Printf("Content: %+v\n", content)
	return nil
}

func (s *ExampleSender) GetChannelType() string {
	return "example"
}

func (s *ExampleSender) ValidateConfig(config interface{}) error {
	return nil
}

// Plugin entry point - this function must be exported
func NewPlugin() Plugin {
	return &ExamplePlugin{}
}
