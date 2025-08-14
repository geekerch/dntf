package message

import (
	"context"
	"fmt"

	"notification/internal/application/cqrs"
	"notification/internal/application/message/dtos"
	"notification/internal/application/message/usecases"
)

// MessageCommandHandlers handles message commands
type MessageCommandHandlers struct {
	sendMessageUC *usecases.SendMessageUseCase
	eventBus      cqrs.EventBus
}

// NewMessageCommandHandlers creates a new message command handlers
func NewMessageCommandHandlers(
	sendMessageUC *usecases.SendMessageUseCase,
	eventBus cqrs.EventBus,
) *MessageCommandHandlers {
	return &MessageCommandHandlers{
		sendMessageUC: sendMessageUC,
		eventBus:      eventBus,
	}
}

// HandleSendMessage handles send message command
func (h *MessageCommandHandlers) HandleSendMessage(ctx context.Context, cmd *SendMessageCommand) (*dtos.MessageResponse, error) {
	// Execute use case
	response, err := h.sendMessageUC.Execute(ctx, cmd.Request)
	if err != nil {
		// Publish failed event
		if response != nil {
			failedEvent := NewMessageFailedEvent(response.ID, err.Error())
			if publishErr := h.eventBus.Publish(ctx, failedEvent); publishErr != nil {
				fmt.Printf("Failed to publish message failed event: %v\n", publishErr)
			}
		}
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	// Publish success event
	event := NewMessageSentEvent(response)
	if err := h.eventBus.Publish(ctx, event); err != nil {
		// Log error but don't fail the command
		fmt.Printf("Failed to publish message sent event: %v\n", err)
	}

	return response, nil
}

// MessageQueryHandlers handles message queries
type MessageQueryHandlers struct {
	getMessageUC  *usecases.GetMessageUseCase
	listMessagesUC *usecases.ListMessagesUseCase
}

// NewMessageQueryHandlers creates a new message query handlers
func NewMessageQueryHandlers(
	getMessageUC *usecases.GetMessageUseCase,
	listMessagesUC *usecases.ListMessagesUseCase,
) *MessageQueryHandlers {
	return &MessageQueryHandlers{
		getMessageUC:   getMessageUC,
		listMessagesUC: listMessagesUC,
	}
}

// HandleGetMessage handles get message query
func (h *MessageQueryHandlers) HandleGetMessage(ctx context.Context, query *GetMessageQuery) (*dtos.MessageResponse, error) {
	response, err := h.getMessageUC.Execute(ctx, query.MessageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return response, nil
}

// HandleListMessages handles list messages query
func (h *MessageQueryHandlers) HandleListMessages(ctx context.Context, query *ListMessagesQuery) (*dtos.ListMessagesResponse, error) {
	// Convert CQRS query to use case request
	request := &dtos.ListMessagesRequest{
		ChannelID: query.ChannelID,
		Status:    query.Status,
	}
	
	// Set pagination if provided
	if query.Options != nil && query.Options.Pagination != nil {
		request.SkipCount = query.Options.Pagination.Offset
		request.MaxResultCount = query.Options.Pagination.Limit
	}

	response, err := h.listMessagesUC.Execute(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	return response, nil
}

// Command Handlers

// SendMessageCommandHandler handles send message commands
type SendMessageCommandHandler struct {
	handlers *MessageCommandHandlers
}

// NewSendMessageCommandHandler creates a new send message command handler
func NewSendMessageCommandHandler(handlers *MessageCommandHandlers) *SendMessageCommandHandler {
	return &SendMessageCommandHandler{
		handlers: handlers,
	}
}

// Handle handles the send message command
func (h *SendMessageCommandHandler) Handle(ctx context.Context, cmd cqrs.Command) (*cqrs.CommandResult, error) {
	sendCmd, ok := cmd.(*SendMessageCommand)
	if !ok {
		return nil, fmt.Errorf("invalid command type")
	}

	response, err := h.handlers.HandleSendMessage(ctx, sendCmd)
	if err != nil {
		return &cqrs.CommandResult{Success: false, Error: err}, err
	}

	return &cqrs.CommandResult{Success: true, Data: response}, nil
}

// GetCommandType returns the command type this handler supports
func (h *SendMessageCommandHandler) GetCommandType() string {
	return SendMessageCommandType
}

// Query Handlers

// GetMessageQueryHandler handles get message queries
type GetMessageQueryHandler struct {
	handlers *MessageQueryHandlers
}

// NewGetMessageQueryHandler creates a new get message query handler
func NewGetMessageQueryHandler(handlers *MessageQueryHandlers) *GetMessageQueryHandler {
	return &GetMessageQueryHandler{
		handlers: handlers,
	}
}

// Handle handles the get message query
func (h *GetMessageQueryHandler) Handle(ctx context.Context, query cqrs.Query) (*cqrs.QueryResult, error) {
	getQuery, ok := query.(*GetMessageQuery)
	if !ok {
		return nil, fmt.Errorf("invalid query type")
	}

	response, err := h.handlers.HandleGetMessage(ctx, getQuery)
	if err != nil {
		return &cqrs.QueryResult{Success: false, Error: err}, err
	}

	return &cqrs.QueryResult{Success: true, Data: response}, nil
}

// GetQueryType returns the query type this handler supports
func (h *GetMessageQueryHandler) GetQueryType() string {
	return GetMessageQueryType
}

// ListMessagesQueryHandler handles list messages queries
type ListMessagesQueryHandler struct {
	handlers *MessageQueryHandlers
}

// NewListMessagesQueryHandler creates a new list messages query handler
func NewListMessagesQueryHandler(handlers *MessageQueryHandlers) *ListMessagesQueryHandler {
	return &ListMessagesQueryHandler{
		handlers: handlers,
	}
}

// Handle handles the list messages query
func (h *ListMessagesQueryHandler) Handle(ctx context.Context, query cqrs.Query) (*cqrs.QueryResult, error) {
	listQuery, ok := query.(*ListMessagesQuery)
	if !ok {
		return nil, fmt.Errorf("invalid query type")
	}

	response, err := h.handlers.HandleListMessages(ctx, listQuery)
	if err != nil {
		return &cqrs.QueryResult{Success: false, Error: err}, err
	}

	return &cqrs.QueryResult{Success: true, Data: response}, nil
}

// GetQueryType returns the query type this handler supports
func (h *ListMessagesQueryHandler) GetQueryType() string {
	return ListMessagesQueryType
}