package presentation

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"notification/internal/presentation/http/handlers"
	"notification/internal/presentation/http/middleware"
	"notification/internal/presentation/http/routes"
	natshandlers "notification/internal/presentation/nats/handlers"
	"notification/pkg/logger"
)

// Server represents the presentation layer server
type Server struct {
	httpServer     *http.Server
	natsManager    *natshandlers.HandlerManager
	router         *gin.Engine
	config         *ServerConfig
}

// ServerConfig holds the server configuration
type ServerConfig struct {
	HTTPPort    string
	HTTPTimeout time.Duration
	
	// HTTP handlers
	ChannelHandler     *handlers.ChannelHandler
	CQRSChannelHandler *handlers.CQRSChannelHandler
	
	// NATS handler manager
	NATSManager     *natshandlers.HandlerManager
	CQRSNATSHandler *natshandlers.CQRSChannelNATSHandler
	
	// Middleware configuration
	MiddlewareConfig *middleware.MiddlewareConfig
}

// NewServer creates a new presentation layer server
func NewServer(config *ServerConfig) *Server {
	// Setup HTTP router
	routerConfig := &routes.RouterConfig{
		ChannelHandler:     config.ChannelHandler,
		CQRSChannelHandler: config.CQRSChannelHandler,
		MiddlewareConfig:   config.MiddlewareConfig,
	}
	router := routes.SetupRouter(routerConfig)

	// Setup HTTP server
	httpServer := &http.Server{
		Addr:         ":" + config.HTTPPort,
		Handler:      router,
		ReadTimeout:  config.HTTPTimeout,
		WriteTimeout: config.HTTPTimeout,
		IdleTimeout:  config.HTTPTimeout * 2,
	}

	return &Server{
		httpServer:  httpServer,
		natsManager: config.NATSManager,
		router:      router,
		config:      config,
	}
}

// Start starts the presentation layer server
func (s *Server) Start(ctx context.Context) error {
	logger.Info("Starting presentation layer server")

	// Register NATS handlers
	if s.natsManager != nil {
		if err := s.natsManager.RegisterAllHandlers(); err != nil {
			return fmt.Errorf("failed to register NATS handlers: %w", err)
		}
		logger.Info("Traditional NATS handlers registered successfully")
	}

	// Register CQRS NATS handlers
	if s.config.CQRSNATSHandler != nil {
		if err := s.config.CQRSNATSHandler.RegisterHandlers(); err != nil {
			return fmt.Errorf("failed to register CQRS NATS handlers: %w", err)
		}
		logger.Info("CQRS NATS handlers registered successfully")
	}

	// Start HTTP server in a goroutine
	go func() {
		logger.Info("Starting HTTP server", zap.String("port", s.config.HTTPPort))
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed to start", zap.Error(err))
		}
	}()

	logger.Info("Presentation layer server started successfully")
	return nil
}

// Stop gracefully stops the presentation layer server
func (s *Server) Stop(ctx context.Context) error {
	logger.Info("Stopping presentation layer server")

	// Stop HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		logger.Error("Failed to shutdown HTTP server gracefully", zap.Error(err))
		return err
	}
	logger.Info("HTTP server stopped")

	// Stop NATS handlers
	if s.natsManager != nil {
		if err := s.natsManager.Close(); err != nil {
			logger.Error("Failed to close NATS handler manager", zap.Error(err))
			return err
		}
		logger.Info("NATS handler manager stopped")
	}

	logger.Info("Presentation layer server stopped successfully")
	return nil
}

// HealthCheck performs a health check on the server
func (s *Server) HealthCheck() error {
	// Check NATS handlers
	if s.natsManager != nil {
		if err := s.natsManager.HealthCheck(); err != nil {
			return fmt.Errorf("NATS health check failed: %w", err)
		}
	}

	return nil
}

// GetRouter returns the HTTP router (useful for testing)
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}