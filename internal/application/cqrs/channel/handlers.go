package channel

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"notification/internal/application/cqrs"
	"notification/internal/application/channel/dtos"
	"notification/internal/application/channel/usecases"
	"notification/pkg/logger"
)

// ChannelCommandHandlers contains all command handlers for channel operations
type ChannelCommandHandlers struct {
	createUseCase *usecases.CreateChannelUseCase
	updateUseCase *usecases.UpdateChannelUseCase
	deleteUseCase *usecases.DeleteChannelUseCase
	eventBus      cqrs.EventBus
}

// NewChannelCommandHandlers creates new channel command handlers
func NewChannelCommandHandlers(
	createUseCase *usecases.CreateChannelUseCase,
	updateUseCase *usecases.UpdateChannelUseCase,
	deleteUseCase *usecases.DeleteChannelUseCase,
	eventBus cqrs.EventBus,
) *ChannelCommandHandlers {
	return &ChannelCommandHandlers{
		createUseCase: createUseCase,
		updateUseCase: updateUseCase,
		deleteUseCase: deleteUseCase,
		eventBus:      eventBus,
	}
}

// CreateChannelCommandHandler handles create channel commands
type CreateChannelCommandHandler struct {
	handlers *ChannelCommandHandlers
}

// NewCreateChannelCommandHandler creates a new create channel command handler
func NewCreateChannelCommandHandler(handlers *ChannelCommandHandlers) *CreateChannelCommandHandler {
	return &CreateChannelCommandHandler{handlers: handlers}
}

// Handle handles the create channel command
func (h *CreateChannelCommandHandler) Handle(ctx context.Context, command cqrs.Command) (*cqrs.CommandResult, error) {
	cmd, ok := command.(*CreateChannelCommand)
	if !ok {
		return nil, fmt.Errorf("invalid command type")
	}

	logger.Info("Handling create channel command",
		zap.String("command_id", cmd.GetCommandID()),
		zap.String("channel_name", cmd.Request.ChannelName))

	// Execute the use case
	response, err := h.handlers.createUseCase.Execute(ctx, cmd.Request)
	if err != nil {
		return &cqrs.CommandResult{
			CommandID: cmd.GetCommandID(),
			Success:   false,
			Error:     err,
		}, err
	}

	// Create and publish event
	eventData := &ChannelCreatedEventData{
		ChannelID:   response.ChannelID,
		ChannelName: response.ChannelName,
		Description: response.Description,
		ChannelType: response.ChannelType,
		TemplateID:  response.TemplateID,
		Config:      response.Config,
		Recipients:  response.Recipients,
		Tags:        response.Tags,
		Enabled:     response.Enabled,
		CreatedAt:   response.CreatedAt,
	}

	event := NewChannelCreatedEvent(response.ChannelID, 1, eventData)
	events := []cqrs.Event{event}

	// Publish event
	if err := h.handlers.eventBus.Publish(ctx, event); err != nil {
		logger.Error("Failed to publish channel created event", zap.Error(err))
		// Don't fail the command if event publishing fails
	}

	return &cqrs.CommandResult{
		CommandID: cmd.GetCommandID(),
		Success:   true,
		Data:      response,
		Events:    events,
	}, nil
}

// GetCommandType returns the command type this handler processes
func (h *CreateChannelCommandHandler) GetCommandType() string {
	return CreateChannelCommandType
}

// UpdateChannelCommandHandler handles update channel commands
type UpdateChannelCommandHandler struct {
	handlers *ChannelCommandHandlers
}

// NewUpdateChannelCommandHandler creates a new update channel command handler
func NewUpdateChannelCommandHandler(handlers *ChannelCommandHandlers) *UpdateChannelCommandHandler {
	return &UpdateChannelCommandHandler{handlers: handlers}
}

// Handle handles the update channel command
func (h *UpdateChannelCommandHandler) Handle(ctx context.Context, command cqrs.Command) (*cqrs.CommandResult, error) {
	cmd, ok := command.(*UpdateChannelCommand)
	if !ok {
		return nil, fmt.Errorf("invalid command type")
	}

	logger.Info("Handling update channel command",
		zap.String("command_id", cmd.GetCommandID()),
		zap.String("channel_id", cmd.ChannelID))

	// Execute the use case
	response, err := h.handlers.updateUseCase.Execute(ctx, cmd.ChannelID, cmd.Request)
	if err != nil {
		return &cqrs.CommandResult{
			CommandID: cmd.GetCommandID(),
			Success:   false,
			Error:     err,
		}, err
	}

	// Create and publish event
	eventData := &ChannelUpdatedEventData{
		ChannelID:   response.ChannelID,
		ChannelName: response.ChannelName,
		Description: response.Description,
		ChannelType: response.ChannelType,
		TemplateID:  response.TemplateID,
		Config:      response.Config,
		Recipients:  response.Recipients,
		Tags:        response.Tags,
		Enabled:     response.Enabled,
		UpdatedAt:   response.UpdatedAt,
		Changes:     make(map[string]interface{}), // TODO: Track actual changes
	}

	event := NewChannelUpdatedEvent(response.ChannelID, 2, eventData) // TODO: Get actual version
	events := []cqrs.Event{event}

	// Publish event
	if err := h.handlers.eventBus.Publish(ctx, event); err != nil {
		logger.Error("Failed to publish channel updated event", zap.Error(err))
	}

	return &cqrs.CommandResult{
		CommandID: cmd.GetCommandID(),
		Success:   true,
		Data:      response,
		Events:    events,
	}, nil
}

// GetCommandType returns the command type this handler processes
func (h *UpdateChannelCommandHandler) GetCommandType() string {
	return UpdateChannelCommandType
}

// DeleteChannelCommandHandler handles delete channel commands
type DeleteChannelCommandHandler struct {
	handlers *ChannelCommandHandlers
}

// NewDeleteChannelCommandHandler creates a new delete channel command handler
func NewDeleteChannelCommandHandler(handlers *ChannelCommandHandlers) *DeleteChannelCommandHandler {
	return &DeleteChannelCommandHandler{handlers: handlers}
}

// Handle handles the delete channel command
func (h *DeleteChannelCommandHandler) Handle(ctx context.Context, command cqrs.Command) (*cqrs.CommandResult, error) {
	cmd, ok := command.(*DeleteChannelCommand)
	if !ok {
		return nil, fmt.Errorf("invalid command type")
	}

	logger.Info("Handling delete channel command",
		zap.String("command_id", cmd.GetCommandID()),
		zap.String("channel_id", cmd.ChannelID))

	// Execute the use case
	response, err := h.handlers.deleteUseCase.Execute(ctx, cmd.ChannelID)
	if err != nil {
		return &cqrs.CommandResult{
			CommandID: cmd.GetCommandID(),
			Success:   false,
			Error:     err,
		}, err
	}

	// Create and publish event
	eventData := &ChannelDeletedEventData{
		ChannelID:   response.ChannelID,
		ChannelName: "", // TODO: Get channel name before deletion
		DeletedAt:   response.DeletedAt,
	}

	event := NewChannelDeletedEvent(response.ChannelID, 3, eventData) // TODO: Get actual version
	events := []cqrs.Event{event}

	// Publish event
	if err := h.handlers.eventBus.Publish(ctx, event); err != nil {
		logger.Error("Failed to publish channel deleted event", zap.Error(err))
	}

	return &cqrs.CommandResult{
		CommandID: cmd.GetCommandID(),
		Success:   true,
		Data:      response,
		Events:    events,
	}, nil
}

// GetCommandType returns the command type this handler processes
func (h *DeleteChannelCommandHandler) GetCommandType() string {
	return DeleteChannelCommandType
}

// ChannelQueryHandlers contains all query handlers for channel operations
type ChannelQueryHandlers struct {
	getUseCase  *usecases.GetChannelUseCase
	listUseCase *usecases.ListChannelsUseCase
}

// NewChannelQueryHandlers creates new channel query handlers
func NewChannelQueryHandlers(
	getUseCase *usecases.GetChannelUseCase,
	listUseCase *usecases.ListChannelsUseCase,
) *ChannelQueryHandlers {
	return &ChannelQueryHandlers{
		getUseCase:  getUseCase,
		listUseCase: listUseCase,
	}
}

// GetChannelQueryHandler handles get channel queries
type GetChannelQueryHandler struct {
	handlers *ChannelQueryHandlers
}

// NewGetChannelQueryHandler creates a new get channel query handler
func NewGetChannelQueryHandler(handlers *ChannelQueryHandlers) *GetChannelQueryHandler {
	return &GetChannelQueryHandler{handlers: handlers}
}

// Handle handles the get channel query
func (h *GetChannelQueryHandler) Handle(ctx context.Context, query cqrs.Query) (*cqrs.QueryResult, error) {
	q, ok := query.(*GetChannelQuery)
	if !ok {
		return nil, fmt.Errorf("invalid query type")
	}

	logger.Debug("Handling get channel query",
		zap.String("query_id", q.GetQueryID()),
		zap.String("channel_id", q.ChannelID))

	// Execute the use case
	response, err := h.handlers.getUseCase.Execute(ctx, q.ChannelID)
	if err != nil {
		return &cqrs.QueryResult{
			QueryID: q.GetQueryID(),
			Success: false,
			Error:   err,
		}, err
	}

	return &cqrs.QueryResult{
		QueryID: q.GetQueryID(),
		Success: true,
		Data:    response,
	}, nil
}

// GetQueryType returns the query type this handler processes
func (h *GetChannelQueryHandler) GetQueryType() string {
	return GetChannelQueryType
}

// ListChannelsQueryHandler handles list channels queries
type ListChannelsQueryHandler struct {
	handlers *ChannelQueryHandlers
}

// NewListChannelsQueryHandler creates a new list channels query handler
func NewListChannelsQueryHandler(handlers *ChannelQueryHandlers) *ListChannelsQueryHandler {
	return &ListChannelsQueryHandler{handlers: handlers}
}

// Handle handles the list channels query
func (h *ListChannelsQueryHandler) Handle(ctx context.Context, query cqrs.Query) (*cqrs.QueryResult, error) {
	q, ok := query.(*ListChannelsQuery)
	if !ok {
		return nil, fmt.Errorf("invalid query type")
	}

	logger.Debug("Handling list channels query",
		zap.String("query_id", q.GetQueryID()),
		zap.String("channel_type", q.ChannelType))

	// Convert CQRS query to DTO
	request := &dtos.ListChannelsRequest{
		ChannelType: q.ChannelType,
		Tags:        q.Tags,
	}

	if q.Options != nil && q.Options.Pagination != nil {
		request.SkipCount = q.Options.Pagination.Offset
		request.MaxResultCount = q.Options.Pagination.Limit
	}

	// Execute the use case
	response, err := h.handlers.listUseCase.Execute(ctx, request)
	if err != nil {
		return &cqrs.QueryResult{
			QueryID: q.GetQueryID(),
			Success: false,
			Error:   err,
		}, err
	}

	return &cqrs.QueryResult{
		QueryID: q.GetQueryID(),
		Success: true,
		Data:    response,
		Metadata: map[string]interface{}{
			"total_count": response.TotalCount,
			"has_more":    response.HasMore,
		},
	}, nil
}

// GetQueryType returns the query type this handler processes
func (h *ListChannelsQueryHandler) GetQueryType() string {
	return ListChannelsQueryType
}