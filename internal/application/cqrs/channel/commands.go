package channel

import (
	"fmt"

	"notification/internal/application/cqrs"
	"notification/internal/application/channel/dtos"
)

// Command types
const (
	CreateChannelCommandType = "channel.create"
	UpdateChannelCommandType = "channel.update"
	DeleteChannelCommandType = "channel.delete"
)

// CreateChannelCommand represents a command to create a channel
type CreateChannelCommand struct {
	*cqrs.BaseCommand
	Request *dtos.CreateChannelRequest `json:"request"`
}

// NewCreateChannelCommand creates a new create channel command
func NewCreateChannelCommand(request *dtos.CreateChannelRequest) *CreateChannelCommand {
	return &CreateChannelCommand{
		BaseCommand: cqrs.NewBaseCommand(CreateChannelCommandType),
		Request:     request,
	}
}

// Validate validates the create channel command
func (c *CreateChannelCommand) Validate() error {
	if c.Request == nil {
		return fmt.Errorf("request cannot be nil")
	}
	
	if c.Request.ChannelName == "" {
		return fmt.Errorf("channel name is required")
	}
	
	if c.Request.ChannelType == "" {
		return fmt.Errorf("channel type is required")
	}
	
	return nil
}

// UpdateChannelCommand represents a command to update a channel
type UpdateChannelCommand struct {
	*cqrs.BaseCommand
	ChannelID string                     `json:"channelId"`
	Request   *dtos.UpdateChannelRequest `json:"request"`
}

// NewUpdateChannelCommand creates a new update channel command
func NewUpdateChannelCommand(channelID string, request *dtos.UpdateChannelRequest) *UpdateChannelCommand {
	return &UpdateChannelCommand{
		BaseCommand: cqrs.NewBaseCommand(UpdateChannelCommandType),
		ChannelID:   channelID,
		Request:     request,
	}
}

// Validate validates the update channel command
func (c *UpdateChannelCommand) Validate() error {
	if c.ChannelID == "" {
		return fmt.Errorf("channel ID is required")
	}
	
	if c.Request == nil {
		return fmt.Errorf("request cannot be nil")
	}
	
	if c.Request.ChannelName == "" {
		return fmt.Errorf("channel name is required")
	}
	
	if c.Request.ChannelType == "" {
		return fmt.Errorf("channel type is required")
	}
	
	return nil
}

// DeleteChannelCommand represents a command to delete a channel
type DeleteChannelCommand struct {
	*cqrs.BaseCommand
	ChannelID string `json:"channelId"`
}

// NewDeleteChannelCommand creates a new delete channel command
func NewDeleteChannelCommand(channelID string) *DeleteChannelCommand {
	return &DeleteChannelCommand{
		BaseCommand: cqrs.NewBaseCommand(DeleteChannelCommandType),
		ChannelID:   channelID,
	}
}

// Validate validates the delete channel command
func (c *DeleteChannelCommand) Validate() error {
	if c.ChannelID == "" {
		return fmt.Errorf("channel ID is required")
	}
	
	return nil
}