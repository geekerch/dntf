package handlers

import (
	"fmt"

	"github.com/nats-io/nats.go"

	"notification/internal/application/channel/usecases"
	"notification/pkg/logger"
)

// HandlerManager manages all NATS message handlers
type HandlerManager struct {
	natsConn       *nats.Conn
	channelHandler *ChannelNATSHandler
	// Add other handlers here as they are implemented
	// templateHandler *TemplateNATSHandler
	// messageHandler  *MessageNATSHandler
}

// HandlerConfig holds the configuration for creating handlers
type HandlerConfig struct {
	NATSConn *nats.Conn
	
	// Channel use cases
	CreateChannelUseCase *usecases.CreateChannelUseCase
	GetChannelUseCase    *usecases.GetChannelUseCase
	ListChannelsUseCase  *usecases.ListChannelsUseCase
	UpdateChannelUseCase *usecases.UpdateChannelUseCase
	DeleteChannelUseCase *usecases.DeleteChannelUseCase
	
	// TODO: Add template and message use cases when implemented
	// CreateTemplateUseCase *usecases.CreateTemplateUseCase
	// GetTemplateUseCase    *usecases.GetTemplateUseCase
	// ListTemplatesUseCase  *usecases.ListTemplatesUseCase
	// UpdateTemplateUseCase *usecases.UpdateTemplateUseCase
	// DeleteTemplateUseCase *usecases.DeleteTemplateUseCase
	// SendMessageUseCase    *usecases.SendMessageUseCase
}

// NewHandlerManager creates a new NATS handler manager
func NewHandlerManager(config *HandlerConfig) *HandlerManager {
	manager := &HandlerManager{
		natsConn: config.NATSConn,
	}

	// Initialize channel handler
	if config.CreateChannelUseCase != nil &&
		config.GetChannelUseCase != nil &&
		config.ListChannelsUseCase != nil &&
		config.UpdateChannelUseCase != nil &&
		config.DeleteChannelUseCase != nil {
		
		manager.channelHandler = NewChannelNATSHandler(
			config.CreateChannelUseCase,
			config.GetChannelUseCase,
			config.ListChannelsUseCase,
			config.UpdateChannelUseCase,
			config.DeleteChannelUseCase,
			config.NATSConn,
		)
	}

	// TODO: Initialize template and message handlers when implemented
	// if config.CreateTemplateUseCase != nil && ... {
	//     manager.templateHandler = NewTemplateNATSHandler(...)
	// }
	// if config.SendMessageUseCase != nil {
	//     manager.messageHandler = NewMessageNATSHandler(...)
	// }

	return manager
}

// RegisterAllHandlers registers all NATS message handlers
func (m *HandlerManager) RegisterAllHandlers() error {
	logger.Info("Registering NATS message handlers")

	// Register channel handlers
	if m.channelHandler != nil {
		if err := m.channelHandler.RegisterHandlers(); err != nil {
			return fmt.Errorf("failed to register channel handlers: %w", err)
		}
		logger.Info("Channel NATS handlers registered")
	}

	// TODO: Register template handlers when implemented
	// if m.templateHandler != nil {
	//     if err := m.templateHandler.RegisterHandlers(); err != nil {
	//         return fmt.Errorf("failed to register template handlers: %w", err)
	//     }
	//     logger.Info("Template NATS handlers registered")
	// }

	// TODO: Register message handlers when implemented
	// if m.messageHandler != nil {
	//     if err := m.messageHandler.RegisterHandlers(); err != nil {
	//         return fmt.Errorf("failed to register message handlers: %w", err)
	//     }
	//     logger.Info("Message NATS handlers registered")
	// }

	logger.Info("All NATS message handlers registered successfully")
	return nil
}

// Close gracefully shuts down the handler manager
func (m *HandlerManager) Close() error {
	logger.Info("Shutting down NATS handler manager")
	
	if m.natsConn != nil && !m.natsConn.IsClosed() {
		m.natsConn.Close()
		logger.Info("NATS connection closed")
	}
	
	return nil
}

// GetChannelHandler returns the channel NATS handler
func (m *HandlerManager) GetChannelHandler() *ChannelNATSHandler {
	return m.channelHandler
}

// Health check for NATS handlers
func (m *HandlerManager) HealthCheck() error {
	if m.natsConn == nil {
		return fmt.Errorf("NATS connection is nil")
	}
	
	if m.natsConn.IsClosed() {
		return fmt.Errorf("NATS connection is closed")
	}
	
	if !m.natsConn.IsConnected() {
		return fmt.Errorf("NATS connection is not connected")
	}
	
	return nil
}