package message

import (
	"fmt"

	"notification/internal/application/cqrs"
	"notification/internal/application/message/dtos"
)

// Command types
const (
	SendMessageCommandType = "message.send"
)

// SendMessageCommand represents a command to send a message
type SendMessageCommand struct {
	*cqrs.BaseCommand
	Request *dtos.SendMessageRequest `json:"request"`
}

// NewSendMessageCommand creates a new send message command
func NewSendMessageCommand(request *dtos.SendMessageRequest) *SendMessageCommand {
	return &SendMessageCommand{
		BaseCommand: cqrs.NewBaseCommand(SendMessageCommandType),
		Request:     request,
	}
}

// Validate validates the send message command
func (c *SendMessageCommand) Validate() error {
	if c.Request == nil {
		return fmt.Errorf("request cannot be nil")
	}
	
	if c.Request.ChannelID == "" {
		return fmt.Errorf("channel ID is required")
	}
	
	if c.Request.TemplateID == "" {
		return fmt.Errorf("template ID is required")
	}
	
	if len(c.Request.Recipients) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}
	
	// Validate recipients are not empty
	for i, recipient := range c.Request.Recipients {
		if recipient == "" {
			return fmt.Errorf("recipient at index %d cannot be empty", i)
		}
	}
	
	return nil
}