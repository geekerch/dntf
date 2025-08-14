package handlers

import (
	"fmt"

	"github.com/nats-io/nats.go"

	channel_uc "notification/internal/application/channel/usecases"
	message_uc "notification/internal/application/message/usecases"
	template_uc "notification/internal/application/template/usecases"
	"notification/pkg/logger"
)

// HandlerManager manages all NATS message handlers
type HandlerManager struct {
	natsConn        *nats.Conn
	channelHandler  *ChannelNATSHandler
	templateHandler *TemplateNATSHandler
	messageHandler  *MessageNATSHandler
}

// HandlerConfig holds the configuration for creating handlers
type HandlerConfig struct {
	NATSConn *nats.Conn

	// Channel use cases
	CreateChannelUseCase *channel_uc.CreateChannelUseCase
	GetChannelUseCase    *channel_uc.GetChannelUseCase
	ListChannelsUseCase  *channel_uc.ListChannelsUseCase
	UpdateChannelUseCase *channel_uc.UpdateChannelUseCase
	DeleteChannelUseCase *channel_uc.DeleteChannelUseCase

	// Template use cases
	CreateTemplateUseCase *template_uc.CreateTemplateUseCase
	GetTemplateUseCase    *template_uc.GetTemplateUseCase
	ListTemplatesUseCase  *template_uc.ListTemplatesUseCase
	UpdateTemplateUseCase *template_uc.UpdateTemplateUseCase
	DeleteTemplateUseCase *template_uc.DeleteTemplateUseCase

	// Message use cases
	SendMessageUseCase *message_uc.SendMessageUseCase
	GetMessageUseCase  *message_uc.GetMessageUseCase
	ListMessagesUseCase *message_uc.ListMessagesUseCase
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

	// Initialize template handler
	if config.CreateTemplateUseCase != nil &&
		config.GetTemplateUseCase != nil &&
		config.ListTemplatesUseCase != nil &&
		config.UpdateTemplateUseCase != nil &&
		config.DeleteTemplateUseCase != nil {
		manager.templateHandler = NewTemplateNATSHandler(
			config.CreateTemplateUseCase,
			config.GetTemplateUseCase,
			config.ListTemplatesUseCase,
			config.UpdateTemplateUseCase,
			config.DeleteTemplateUseCase,
			config.NATSConn,
		)
	}

	// Initialize message handler
	if config.SendMessageUseCase != nil &&
		config.GetMessageUseCase != nil &&
		config.ListMessagesUseCase != nil {
		manager.messageHandler = NewMessageNATSHandler(
			config.SendMessageUseCase,
			config.GetMessageUseCase,
			config.ListMessagesUseCase,
			config.NATSConn,
		)
	}

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

	// Register template handlers
	if m.templateHandler != nil {
		if err := m.templateHandler.RegisterHandlers(); err != nil {
			return fmt.Errorf("failed to register template handlers: %w", err)
		}
		logger.Info("Template NATS handlers registered")
	}

	// Register message handlers
	if m.messageHandler != nil {
		if err := m.messageHandler.RegisterHandlers(); err != nil {
			return fmt.Errorf("failed to register message handlers: %w", err)
		}
		logger.Info("Message NATS handlers registered")
	}

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
