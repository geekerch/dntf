package cqrs

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"notification/pkg/logger"
)

// CQRSManager manages the CQRS infrastructure
type CQRSManager struct {
	commandBus CommandBus
	queryBus   QueryBus
	eventBus   EventBus
}

// NewCQRSManager creates a new CQRS manager
func NewCQRSManager() *CQRSManager {
	return &CQRSManager{
		commandBus: NewDefaultCommandBus(),
		queryBus:   NewDefaultQueryBus(),
		eventBus:   NewDefaultEventBus(),
	}
}

// NewCQRSManagerWithBuses creates a new CQRS manager with custom buses
func NewCQRSManagerWithBuses(commandBus CommandBus, queryBus QueryBus, eventBus EventBus) *CQRSManager {
	return &CQRSManager{
		commandBus: commandBus,
		queryBus:   queryBus,
		eventBus:   eventBus,
	}
}

// GetCommandBus returns the command bus
func (m *CQRSManager) GetCommandBus() CommandBus {
	return m.commandBus
}

// GetQueryBus returns the query bus
func (m *CQRSManager) GetQueryBus() QueryBus {
	return m.queryBus
}

// GetEventBus returns the event bus
func (m *CQRSManager) GetEventBus() EventBus {
	return m.eventBus
}

// RegisterCommandHandler registers a command handler
func (m *CQRSManager) RegisterCommandHandler(handler CommandHandler) error {
	return m.commandBus.RegisterHandler(handler)
}

// RegisterQueryHandler registers a query handler
func (m *CQRSManager) RegisterQueryHandler(handler QueryHandler) error {
	return m.queryBus.RegisterHandler(handler)
}

// RegisterEventHandler registers an event handler
func (m *CQRSManager) RegisterEventHandler(eventType string, handler EventHandler) error {
	return m.eventBus.Subscribe(eventType, handler)
}

// ExecuteCommand executes a command
func (m *CQRSManager) ExecuteCommand(ctx context.Context, command Command) (*CommandResult, error) {
	return m.commandBus.Execute(ctx, command)
}

// ExecuteQuery executes a query
func (m *CQRSManager) ExecuteQuery(ctx context.Context, query Query) (*QueryResult, error) {
	return m.queryBus.Execute(ctx, query)
}

// PublishEvent publishes an event
func (m *CQRSManager) PublishEvent(ctx context.Context, event Event) error {
	return m.eventBus.Publish(ctx, event)
}

// PublishEvents publishes multiple events
func (m *CQRSManager) PublishEvents(ctx context.Context, events []Event) error {
	return m.eventBus.PublishBatch(ctx, events)
}

// CQRSConfig holds configuration for CQRS setup
type CQRSConfig struct {
	EnableCommandLogging bool
	EnableQueryLogging   bool
	EnableEventLogging   bool
	EnableMetrics        bool
}

// DefaultCQRSConfig returns default CQRS configuration
func DefaultCQRSConfig() *CQRSConfig {
	return &CQRSConfig{
		EnableCommandLogging: true,
		EnableQueryLogging:   false, // Queries can be frequent, so disabled by default
		EnableEventLogging:   true,
		EnableMetrics:        true,
	}
}

// CQRSFacade provides a simplified interface for CQRS operations
type CQRSFacade struct {
	manager *CQRSManager
	config  *CQRSConfig
}

// NewCQRSFacade creates a new CQRS facade
func NewCQRSFacade(manager *CQRSManager, config *CQRSConfig) *CQRSFacade {
	if config == nil {
		config = DefaultCQRSConfig()
	}
	
	return &CQRSFacade{
		manager: manager,
		config:  config,
	}
}

// Send executes a command
func (f *CQRSFacade) Send(ctx context.Context, command Command) (*CommandResult, error) {
	if f.config.EnableCommandLogging {
		logger.Info("Sending command",
			zap.String("command_id", command.GetCommandID()),
			zap.String("command_type", command.GetCommandType()))
	}
	
	result, err := f.manager.ExecuteCommand(ctx, command)
	
	if f.config.EnableCommandLogging {
		if err != nil {
			logger.Error("Command failed",
				zap.String("command_id", command.GetCommandID()),
				zap.String("command_type", command.GetCommandType()),
				zap.Error(err))
		} else {
			logger.Info("Command completed",
				zap.String("command_id", command.GetCommandID()),
				zap.String("command_type", command.GetCommandType()),
				zap.Bool("success", result.Success))
		}
	}
	
	return result, err
}

// Query executes a query
func (f *CQRSFacade) Query(ctx context.Context, query Query) (*QueryResult, error) {
	if f.config.EnableQueryLogging {
		logger.Debug("Executing query",
			zap.String("query_id", query.GetQueryID()),
			zap.String("query_type", query.GetQueryType()))
	}
	
	result, err := f.manager.ExecuteQuery(ctx, query)
	
	if f.config.EnableQueryLogging {
		if err != nil {
			logger.Error("Query failed",
				zap.String("query_id", query.GetQueryID()),
				zap.String("query_type", query.GetQueryType()),
				zap.Error(err))
		} else {
			logger.Debug("Query completed",
				zap.String("query_id", query.GetQueryID()),
				zap.String("query_type", query.GetQueryType()),
				zap.Bool("success", result.Success),
				zap.Bool("cache_hit", result.CacheHit))
		}
	}
	
	return result, err
}

// Publish publishes an event
func (f *CQRSFacade) Publish(ctx context.Context, event Event) error {
	if f.config.EnableEventLogging {
		logger.Info("Publishing event",
			zap.String("event_id", event.GetEventID()),
			zap.String("event_type", event.GetEventType()),
			zap.String("aggregate_id", event.GetAggregateID()))
	}
	
	err := f.manager.PublishEvent(ctx, event)
	
	if f.config.EnableEventLogging && err != nil {
		logger.Error("Event publishing failed",
			zap.String("event_id", event.GetEventID()),
			zap.String("event_type", event.GetEventType()),
			zap.Error(err))
	}
	
	return err
}

// RegisterHandlers registers all handlers for a specific domain
func (f *CQRSFacade) RegisterHandlers(handlers interface{}) error {
	switch h := handlers.(type) {
	case CommandHandler:
		return f.manager.RegisterCommandHandler(h)
	case QueryHandler:
		return f.manager.RegisterQueryHandler(h)
	case map[string]EventHandler:
		for eventType, handler := range h {
			if err := f.manager.RegisterEventHandler(eventType, handler); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported handler type: %T", handlers)
	}
}

// HealthCheck performs a health check on the CQRS system
func (f *CQRSFacade) HealthCheck(ctx context.Context) error {
	// TODO: Implement health checks for command bus, query bus, and event bus
	// This could include checking if handlers are registered, buses are responsive, etc.
	return nil
}

// GetMetrics returns CQRS metrics
func (f *CQRSFacade) GetMetrics() map[string]interface{} {
	// TODO: Implement metrics collection
	// This could include command/query execution times, success rates, event publishing rates, etc.
	return map[string]interface{}{
		"commands_executed": 0,
		"queries_executed":  0,
		"events_published":  0,
	}
}