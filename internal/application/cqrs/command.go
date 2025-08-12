package cqrs

import (
	"context"
	"time"
)

// Command represents a command in the CQRS pattern
type Command interface {
	// GetCommandID returns the unique identifier for this command
	GetCommandID() string
	// GetCommandType returns the type of the command
	GetCommandType() string
	// GetTimestamp returns when the command was created
	GetTimestamp() time.Time
	// Validate validates the command
	Validate() error
}

// CommandResult represents the result of executing a command
type CommandResult struct {
	CommandID   string                 `json:"commandId"`
	Success     bool                   `json:"success"`
	Data        interface{}            `json:"data,omitempty"`
	Error       error                  `json:"error,omitempty"`
	Events      []Event                `json:"events,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	ExecutedAt  time.Time              `json:"executedAt"`
	Duration    time.Duration          `json:"duration"`
}

// CommandHandler handles a specific type of command
type CommandHandler interface {
	// Handle processes the command and returns the result
	Handle(ctx context.Context, command Command) (*CommandResult, error)
	// GetCommandType returns the type of command this handler processes
	GetCommandType() string
}

// CommandBus dispatches commands to their appropriate handlers
type CommandBus interface {
	// Execute executes a command
	Execute(ctx context.Context, command Command) (*CommandResult, error)
	// RegisterHandler registers a command handler
	RegisterHandler(handler CommandHandler) error
	// GetHandler returns the handler for a command type
	GetHandler(commandType string) (CommandHandler, error)
}

// BaseCommand provides common command functionality
type BaseCommand struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"userId,omitempty"`
	TraceID   string    `json:"traceId,omitempty"`
}

// GetCommandID returns the command ID
func (c *BaseCommand) GetCommandID() string {
	return c.ID
}

// GetCommandType returns the command type
func (c *BaseCommand) GetCommandType() string {
	return c.Type
}

// GetTimestamp returns the command timestamp
func (c *BaseCommand) GetTimestamp() time.Time {
	return c.Timestamp
}

// NewBaseCommand creates a new base command
func NewBaseCommand(commandType string) *BaseCommand {
	return &BaseCommand{
		ID:        generateID(),
		Type:      commandType,
		Timestamp: time.Now(),
	}
}

// generateID generates a unique ID for commands
func generateID() string {
	// Simple implementation - in production, consider using UUID
	return time.Now().Format("20060102150405.000000")
}