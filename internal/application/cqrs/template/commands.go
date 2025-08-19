package template

import (
	"fmt"

	"notification/internal/application/cqrs"
	"notification/internal/application/template/dtos"
)

// Command types
const (
	CreateTemplateCommandType = "template.create"
	UpdateTemplateCommandType = "template.update"
	DeleteTemplateCommandType = "template.delete"
)

// CreateTemplateCommand represents a command to create a template
type CreateTemplateCommand struct {
	*cqrs.BaseCommand
	Request *dtos.CreateTemplateRequest `json:"request"`
}

// NewCreateTemplateCommand creates a new create template command
func NewCreateTemplateCommand(request *dtos.CreateTemplateRequest) *CreateTemplateCommand {
	return &CreateTemplateCommand{
		BaseCommand: cqrs.NewBaseCommand(CreateTemplateCommandType),
		Request:     request,
	}
}

// Validate validates the create template command
func (c *CreateTemplateCommand) Validate() error {
	if c.Request == nil {
		return fmt.Errorf("request cannot be nil")
	}
	
	if c.Request.Name == "" {
		return fmt.Errorf("template name is required")
	}
	
	if c.Request.ChannelType.String() == "" {
		return fmt.Errorf("channel type is required")
	}
	
	if c.Request.Content == "" {
		return fmt.Errorf("template content is required")
	}
	
	return nil
}

// UpdateTemplateCommand represents a command to update a template
type UpdateTemplateCommand struct {
	*cqrs.BaseCommand
	TemplateID string                     `json:"templateId"`
	Request    *dtos.UpdateTemplateRequest `json:"request"`
}

// NewUpdateTemplateCommand creates a new update template command
func NewUpdateTemplateCommand(templateID string, request *dtos.UpdateTemplateRequest) *UpdateTemplateCommand {
	return &UpdateTemplateCommand{
		BaseCommand: cqrs.NewBaseCommand(UpdateTemplateCommandType),
		TemplateID:  templateID,
		Request:     request,
	}
}

// Validate validates the update template command
func (c *UpdateTemplateCommand) Validate() error {
	if c.TemplateID == "" {
		return fmt.Errorf("template ID is required")
	}
	
	if c.Request == nil {
		return fmt.Errorf("request cannot be nil")
	}
	
	// At least one field should be provided for update
	if c.Request.Name == nil && c.Request.Subject == nil && c.Request.Content == nil &&
		c.Request.Variables == nil && c.Request.Tags == nil && c.Request.Settings == nil {
		return fmt.Errorf("at least one field must be provided for update")
	}
	
	return nil
}

// DeleteTemplateCommand represents a command to delete a template
type DeleteTemplateCommand struct {
	*cqrs.BaseCommand
	TemplateID string `json:"templateId"`
}

// NewDeleteTemplateCommand creates a new delete template command
func NewDeleteTemplateCommand(templateID string) *DeleteTemplateCommand {
	return &DeleteTemplateCommand{
		BaseCommand: cqrs.NewBaseCommand(DeleteTemplateCommandType),
		TemplateID:  templateID,
	}
}

// Validate validates the delete template command
func (c *DeleteTemplateCommand) Validate() error {
	if c.TemplateID == "" {
		return fmt.Errorf("template ID is required")
	}
	
	return nil
}