package cqrs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"notification/pkg/logger"
)

// DefaultCommandBus is the default implementation of CommandBus
type DefaultCommandBus struct {
	handlers map[string]CommandHandler
	mutex    sync.RWMutex
}

// NewDefaultCommandBus creates a new default command bus
func NewDefaultCommandBus() *DefaultCommandBus {
	return &DefaultCommandBus{
		handlers: make(map[string]CommandHandler),
	}
}

// Execute executes a command
func (bus *DefaultCommandBus) Execute(ctx context.Context, command Command) (*CommandResult, error) {
	startTime := time.Now()
	
	logger.Info("Executing command",
		zap.String("command_id", command.GetCommandID()),
		zap.String("command_type", command.GetCommandType()))

	// Validate command
	if err := command.Validate(); err != nil {
		logger.Error("Command validation failed",
			zap.String("command_id", command.GetCommandID()),
			zap.Error(err))
		return &CommandResult{
			CommandID:  command.GetCommandID(),
			Success:    false,
			Error:      fmt.Errorf("command validation failed: %w", err),
			ExecutedAt: time.Now(),
			Duration:   time.Since(startTime),
		}, err
	}

	// Get handler
	handler, err := bus.GetHandler(command.GetCommandType())
	if err != nil {
		logger.Error("No handler found for command",
			zap.String("command_id", command.GetCommandID()),
			zap.String("command_type", command.GetCommandType()),
			zap.Error(err))
		return &CommandResult{
			CommandID:  command.GetCommandID(),
			Success:    false,
			Error:      err,
			ExecutedAt: time.Now(),
			Duration:   time.Since(startTime),
		}, err
	}

	// Execute command
	result, err := handler.Handle(ctx, command)
	if err != nil {
		logger.Error("Command execution failed",
			zap.String("command_id", command.GetCommandID()),
			zap.String("command_type", command.GetCommandType()),
			zap.Error(err))
		return &CommandResult{
			CommandID:  command.GetCommandID(),
			Success:    false,
			Error:      err,
			ExecutedAt: time.Now(),
			Duration:   time.Since(startTime),
		}, err
	}

	// Update result with timing information
	result.Duration = time.Since(startTime)
	result.ExecutedAt = time.Now()

	logger.Info("Command executed successfully",
		zap.String("command_id", command.GetCommandID()),
		zap.String("command_type", command.GetCommandType()),
		zap.Duration("duration", result.Duration))

	return result, nil
}

// RegisterHandler registers a command handler
func (bus *DefaultCommandBus) RegisterHandler(handler CommandHandler) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	commandType := handler.GetCommandType()
	if _, exists := bus.handlers[commandType]; exists {
		return fmt.Errorf("handler for command type %s already registered", commandType)
	}

	bus.handlers[commandType] = handler
	logger.Info("Command handler registered",
		zap.String("command_type", commandType))

	return nil
}

// GetHandler returns the handler for a command type
func (bus *DefaultCommandBus) GetHandler(commandType string) (CommandHandler, error) {
	bus.mutex.RLock()
	defer bus.mutex.RUnlock()

	handler, exists := bus.handlers[commandType]
	if !exists {
		return nil, fmt.Errorf("no handler registered for command type: %s", commandType)
	}

	return handler, nil
}

// DefaultQueryBus is the default implementation of QueryBus
type DefaultQueryBus struct {
	handlers map[string]QueryHandler
	mutex    sync.RWMutex
}

// NewDefaultQueryBus creates a new default query bus
func NewDefaultQueryBus() *DefaultQueryBus {
	return &DefaultQueryBus{
		handlers: make(map[string]QueryHandler),
	}
}

// Execute executes a query
func (bus *DefaultQueryBus) Execute(ctx context.Context, query Query) (*QueryResult, error) {
	startTime := time.Now()
	
	logger.Debug("Executing query",
		zap.String("query_id", query.GetQueryID()),
		zap.String("query_type", query.GetQueryType()))

	// Validate query
	if err := query.Validate(); err != nil {
		logger.Error("Query validation failed",
			zap.String("query_id", query.GetQueryID()),
			zap.Error(err))
		return &QueryResult{
			QueryID:    query.GetQueryID(),
			Success:    false,
			Error:      fmt.Errorf("query validation failed: %w", err),
			ExecutedAt: time.Now(),
			Duration:   time.Since(startTime),
		}, err
	}

	// Get handler
	handler, err := bus.GetHandler(query.GetQueryType())
	if err != nil {
		logger.Error("No handler found for query",
			zap.String("query_id", query.GetQueryID()),
			zap.String("query_type", query.GetQueryType()),
			zap.Error(err))
		return &QueryResult{
			QueryID:    query.GetQueryID(),
			Success:    false,
			Error:      err,
			ExecutedAt: time.Now(),
			Duration:   time.Since(startTime),
		}, err
	}

	// Execute query
	result, err := handler.Handle(ctx, query)
	if err != nil {
		logger.Error("Query execution failed",
			zap.String("query_id", query.GetQueryID()),
			zap.String("query_type", query.GetQueryType()),
			zap.Error(err))
		return &QueryResult{
			QueryID:    query.GetQueryID(),
			Success:    false,
			Error:      err,
			ExecutedAt: time.Now(),
			Duration:   time.Since(startTime),
		}, err
	}

	// Update result with timing information
	result.Duration = time.Since(startTime)
	result.ExecutedAt = time.Now()

	logger.Debug("Query executed successfully",
		zap.String("query_id", query.GetQueryID()),
		zap.String("query_type", query.GetQueryType()),
		zap.Duration("duration", result.Duration),
		zap.Bool("cache_hit", result.CacheHit))

	return result, nil
}

// RegisterHandler registers a query handler
func (bus *DefaultQueryBus) RegisterHandler(handler QueryHandler) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	queryType := handler.GetQueryType()
	if _, exists := bus.handlers[queryType]; exists {
		return fmt.Errorf("handler for query type %s already registered", queryType)
	}

	bus.handlers[queryType] = handler
	logger.Info("Query handler registered",
		zap.String("query_type", queryType))

	return nil
}

// GetHandler returns the handler for a query type
func (bus *DefaultQueryBus) GetHandler(queryType string) (QueryHandler, error) {
	bus.mutex.RLock()
	defer bus.mutex.RUnlock()

	handler, exists := bus.handlers[queryType]
	if !exists {
		return nil, fmt.Errorf("no handler registered for query type: %s", queryType)
	}

	return handler, nil
}

// DefaultEventBus is the default implementation of EventBus
type DefaultEventBus struct {
	handlers map[string][]EventHandler
	mutex    sync.RWMutex
}

// NewDefaultEventBus creates a new default event bus
func NewDefaultEventBus() *DefaultEventBus {
	return &DefaultEventBus{
		handlers: make(map[string][]EventHandler),
	}
}

// Publish publishes an event
func (bus *DefaultEventBus) Publish(ctx context.Context, event Event) error {
	logger.Info("Publishing event",
		zap.String("event_id", event.GetEventID()),
		zap.String("event_type", event.GetEventType()),
		zap.String("aggregate_id", event.GetAggregateID()))

	bus.mutex.RLock()
	handlers, exists := bus.handlers[event.GetEventType()]
	bus.mutex.RUnlock()

	if !exists {
		logger.Debug("No handlers registered for event type",
			zap.String("event_type", event.GetEventType()))
		return nil
	}

	// Handle event with all registered handlers
	for _, handler := range handlers {
		if err := handler.Handle(ctx, event); err != nil {
			logger.Error("Event handler failed",
				zap.String("event_id", event.GetEventID()),
				zap.String("event_type", event.GetEventType()),
				zap.Error(err))
			// Continue with other handlers even if one fails
		}
	}

	return nil
}

// PublishBatch publishes multiple events
func (bus *DefaultEventBus) PublishBatch(ctx context.Context, events []Event) error {
	for _, event := range events {
		if err := bus.Publish(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// Subscribe subscribes to events of a specific type
func (bus *DefaultEventBus) Subscribe(eventType string, handler EventHandler) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	if bus.handlers[eventType] == nil {
		bus.handlers[eventType] = make([]EventHandler, 0)
	}

	bus.handlers[eventType] = append(bus.handlers[eventType], handler)
	logger.Info("Event handler subscribed",
		zap.String("event_type", eventType))

	return nil
}

// Unsubscribe unsubscribes from events of a specific type
func (bus *DefaultEventBus) Unsubscribe(eventType string, handler EventHandler) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	handlers, exists := bus.handlers[eventType]
	if !exists {
		return fmt.Errorf("no handlers registered for event type: %s", eventType)
	}

	// Remove the handler from the slice
	for i, h := range handlers {
		if h == handler {
			bus.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			logger.Info("Event handler unsubscribed",
				zap.String("event_type", eventType))
			return nil
		}
	}

	return fmt.Errorf("handler not found for event type: %s", eventType)
}