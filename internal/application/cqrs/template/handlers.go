package template

import (
	"context"
	"fmt"

	"notification/internal/application/cqrs"
	"notification/internal/application/template/dtos"
	"notification/internal/application/template/usecases"
)

// TemplateCommandHandlers handles template commands
type TemplateCommandHandlers struct {
	createTemplateUC *usecases.CreateTemplateUseCase
	updateTemplateUC *usecases.UpdateTemplateUseCase
	deleteTemplateUC *usecases.DeleteTemplateUseCase
	eventBus         cqrs.EventBus
}

// NewTemplateCommandHandlers creates a new template command handlers
func NewTemplateCommandHandlers(
	createTemplateUC *usecases.CreateTemplateUseCase,
	updateTemplateUC *usecases.UpdateTemplateUseCase,
	deleteTemplateUC *usecases.DeleteTemplateUseCase,
	eventBus cqrs.EventBus,
) *TemplateCommandHandlers {
	return &TemplateCommandHandlers{
		createTemplateUC: createTemplateUC,
		updateTemplateUC: updateTemplateUC,
		deleteTemplateUC: deleteTemplateUC,
		eventBus:         eventBus,
	}
}

// HandleCreateTemplate handles create template command
func (h *TemplateCommandHandlers) HandleCreateTemplate(ctx context.Context, cmd *CreateTemplateCommand) (*dtos.TemplateResponse, error) {
	// Execute use case
	response, err := h.createTemplateUC.Execute(ctx, cmd.Request)
	if err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	// Publish event
	event := NewTemplateCreatedEvent(response)
	if err := h.eventBus.Publish(ctx, event); err != nil {
		// Log error but don't fail the command
		// In production, you might want to use a more sophisticated error handling strategy
		fmt.Printf("Failed to publish template created event: %v\n", err)
	}

	return response, nil
}

// HandleUpdateTemplate handles update template command
func (h *TemplateCommandHandlers) HandleUpdateTemplate(ctx context.Context, cmd *UpdateTemplateCommand) (*dtos.TemplateResponse, error) {
	// Execute use case
	response, err := h.updateTemplateUC.Execute(ctx, cmd.TemplateID, cmd.Request)
	if err != nil {
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	// Publish event
	event := NewTemplateUpdatedEvent(response)
	if err := h.eventBus.Publish(ctx, event); err != nil {
		// Log error but don't fail the command
		fmt.Printf("Failed to publish template updated event: %v\n", err)
	}

	return response, nil
}

// HandleDeleteTemplate handles delete template command
func (h *TemplateCommandHandlers) HandleDeleteTemplate(ctx context.Context, cmd *DeleteTemplateCommand) error {
	// Execute use case
	err := h.deleteTemplateUC.Execute(ctx, cmd.TemplateID)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	// Publish event
	event := NewTemplateDeletedEvent(cmd.TemplateID)
	if err := h.eventBus.Publish(ctx, event); err != nil {
		// Log error but don't fail the command
		fmt.Printf("Failed to publish template deleted event: %v\n", err)
	}

	return nil
}

// TemplateQueryHandlers handles template queries
type TemplateQueryHandlers struct {
	getTemplateUC    *usecases.GetTemplateUseCase
	listTemplatesUC  *usecases.ListTemplatesUseCase
}

// NewTemplateQueryHandlers creates a new template query handlers
func NewTemplateQueryHandlers(
	getTemplateUC *usecases.GetTemplateUseCase,
	listTemplatesUC *usecases.ListTemplatesUseCase,
) *TemplateQueryHandlers {
	return &TemplateQueryHandlers{
		getTemplateUC:   getTemplateUC,
		listTemplatesUC: listTemplatesUC,
	}
}

// HandleGetTemplate handles get template query
func (h *TemplateQueryHandlers) HandleGetTemplate(ctx context.Context, query *GetTemplateQuery) (*dtos.TemplateResponse, error) {
	response, err := h.getTemplateUC.Execute(ctx, query.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	return response, nil
}

// HandleListTemplates handles list templates query
func (h *TemplateQueryHandlers) HandleListTemplates(ctx context.Context, query *ListTemplatesQuery) (*dtos.ListTemplatesResponse, error) {
	// Convert CQRS query to use case request
	request := &dtos.ListTemplatesRequest{}
	
	if query.ChannelType != "" {
		// Note: You might need to convert string to ChannelType enum
		// request.ChannelType = &query.ChannelType
	}
	
	if len(query.Tags) > 0 {
		request.Tags = query.Tags
	}
	
	// Handle pagination
	if query.Options != nil && query.Options.Pagination != nil {
		// Convert offset/limit to page/size
		page := (query.Options.Pagination.Offset / query.Options.Pagination.Limit) + 1
		request.Page = page
		request.Size = query.Options.Pagination.Limit
	}

	response, err := h.listTemplatesUC.Execute(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}

	return response, nil
}

// Command Handlers

// CreateTemplateCommandHandler handles create template commands
type CreateTemplateCommandHandler struct {
	handlers *TemplateCommandHandlers
}

// NewCreateTemplateCommandHandler creates a new create template command handler
func NewCreateTemplateCommandHandler(handlers *TemplateCommandHandlers) *CreateTemplateCommandHandler {
	return &CreateTemplateCommandHandler{
		handlers: handlers,
	}
}

// Handle handles the create template command
func (h *CreateTemplateCommandHandler) Handle(ctx context.Context, cmd cqrs.Command) (*cqrs.CommandResult, error) {
	createCmd, ok := cmd.(*CreateTemplateCommand)
	if !ok {
		return nil, fmt.Errorf("invalid command type")
	}

	response, err := h.handlers.HandleCreateTemplate(ctx, createCmd)
	if err != nil {
		return &cqrs.CommandResult{Success: false, Error: err}, err
	}

	return &cqrs.CommandResult{Success: true, Data: response}, nil
}

// GetCommandType returns the command type this handler supports
func (h *CreateTemplateCommandHandler) GetCommandType() string {
	return CreateTemplateCommandType
}

// UpdateTemplateCommandHandler handles update template commands
type UpdateTemplateCommandHandler struct {
	handlers *TemplateCommandHandlers
}

// NewUpdateTemplateCommandHandler creates a new update template command handler
func NewUpdateTemplateCommandHandler(handlers *TemplateCommandHandlers) *UpdateTemplateCommandHandler {
	return &UpdateTemplateCommandHandler{
		handlers: handlers,
	}
}

// Handle handles the update template command
func (h *UpdateTemplateCommandHandler) Handle(ctx context.Context, cmd cqrs.Command) (*cqrs.CommandResult, error) {
	updateCmd, ok := cmd.(*UpdateTemplateCommand)
	if !ok {
		return nil, fmt.Errorf("invalid command type")
	}

	response, err := h.handlers.HandleUpdateTemplate(ctx, updateCmd)
	if err != nil {
		return &cqrs.CommandResult{Success: false, Error: err}, err
	}

	return &cqrs.CommandResult{Success: true, Data: response}, nil
}

// GetCommandType returns the command type this handler supports
func (h *UpdateTemplateCommandHandler) GetCommandType() string {
	return UpdateTemplateCommandType
}

// DeleteTemplateCommandHandler handles delete template commands
type DeleteTemplateCommandHandler struct {
	handlers *TemplateCommandHandlers
}

// NewDeleteTemplateCommandHandler creates a new delete template command handler
func NewDeleteTemplateCommandHandler(handlers *TemplateCommandHandlers) *DeleteTemplateCommandHandler {
	return &DeleteTemplateCommandHandler{
		handlers: handlers,
	}
}

// Handle handles the delete template command
func (h *DeleteTemplateCommandHandler) Handle(ctx context.Context, cmd cqrs.Command) (*cqrs.CommandResult, error) {
	deleteCmd, ok := cmd.(*DeleteTemplateCommand)
	if !ok {
		return nil, fmt.Errorf("invalid command type")
	}

	err := h.handlers.HandleDeleteTemplate(ctx, deleteCmd)
	if err != nil {
		return &cqrs.CommandResult{Success: false, Error: err}, err
	}

	return &cqrs.CommandResult{Success: true}, nil
}

// GetCommandType returns the command type this handler supports
func (h *DeleteTemplateCommandHandler) GetCommandType() string {
	return DeleteTemplateCommandType
}

// Query Handlers

// GetTemplateQueryHandler handles get template queries
type GetTemplateQueryHandler struct {
	handlers *TemplateQueryHandlers
}

// NewGetTemplateQueryHandler creates a new get template query handler
func NewGetTemplateQueryHandler(handlers *TemplateQueryHandlers) *GetTemplateQueryHandler {
	return &GetTemplateQueryHandler{
		handlers: handlers,
	}
}

// Handle handles the get template query
func (h *GetTemplateQueryHandler) Handle(ctx context.Context, query cqrs.Query) (*cqrs.QueryResult, error) {
	getQuery, ok := query.(*GetTemplateQuery)
	if !ok {
		return nil, fmt.Errorf("invalid query type")
	}

	response, err := h.handlers.HandleGetTemplate(ctx, getQuery)
	if err != nil {
		return &cqrs.QueryResult{Success: false, Error: err}, err
	}

	return &cqrs.QueryResult{Success: true, Data: response}, nil
}

// GetQueryType returns the query type this handler supports
func (h *GetTemplateQueryHandler) GetQueryType() string {
	return GetTemplateQueryType
}

// ListTemplatesQueryHandler handles list templates queries
type ListTemplatesQueryHandler struct {
	handlers *TemplateQueryHandlers
}

// NewListTemplatesQueryHandler creates a new list templates query handler
func NewListTemplatesQueryHandler(handlers *TemplateQueryHandlers) *ListTemplatesQueryHandler {
	return &ListTemplatesQueryHandler{
		handlers: handlers,
	}
}

// Handle handles the list templates query
func (h *ListTemplatesQueryHandler) Handle(ctx context.Context, query cqrs.Query) (*cqrs.QueryResult, error) {
	listQuery, ok := query.(*ListTemplatesQuery)
	if !ok {
		return nil, fmt.Errorf("invalid query type")
	}

	response, err := h.handlers.HandleListTemplates(ctx, listQuery)
	if err != nil {
		return &cqrs.QueryResult{Success: false, Error: err}, err
	}

	return &cqrs.QueryResult{Success: true, Data: response}, nil
}

// GetQueryType returns the query type this handler supports
func (h *ListTemplatesQueryHandler) GetQueryType() string {
	return ListTemplatesQueryType
}