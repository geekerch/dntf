package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"channel-api/internal/application/channel/usecases"
	"channel-api/internal/domain/services"
	"channel-api/internal/infrastructure/external"
	"channel-api/internal/infrastructure/messaging"
	"channel-api/internal/infrastructure/repository"
	"channel-api/pkg/config"
	"channel-api/pkg/database"
	"channel-api/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.InitGlobalLogger(&cfg.Logger); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	log := logger.GetGlobalLogger()

	log.Info("Starting Channel API server",
		zap.String("version", "1.0.0"),
		zap.String("server_address", cfg.GetServerAddress()))

	// Initialize database
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	log.Info("Database connected successfully",
		zap.String("host", cfg.Database.Host),
		zap.Int("port", cfg.Database.Port),
		zap.String("database", cfg.Database.DBName))

	// Run database migrations
	if err := db.RunMigrations("./migrations"); err != nil {
		log.Fatal("Failed to run database migrations", zap.Error(err))
	}
	log.Info("Database migrations completed successfully")

	// Initialize NATS client
	natsClient, err := messaging.NewNATSClient(&cfg.NATS, log)
	if err != nil {
		log.Fatal("Failed to connect to NATS", zap.Error(err))
	}
	defer natsClient.Close()

	log.Info("NATS connected successfully", zap.String("url", cfg.NATS.URL))

	// Build dependency container
	container := buildContainer(db, natsClient, log, cfg)

	// Initialize HTTP server (placeholder for now)
	httpServer := &http.Server{
		Addr:         cfg.GetServerAddress(),
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		Handler:      buildHTTPHandler(container),
	}

	// Initialize NATS handlers
	if err := initializeNATSHandlers(natsClient, container, log); err != nil {
		log.Fatal("Failed to initialize NATS handlers", zap.Error(err))
	}

	// Start HTTP server
	go func() {
		log.Info("Starting HTTP server", zap.String("address", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTP server failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", zap.Error(err))
	} else {
		log.Info("Server shutdown completed")
	}
}

// Container holds all application dependencies
type Container struct {
	// Repositories
	ChannelRepo  repository.ChannelRepositoryImpl
	TemplateRepo repository.TemplateRepositoryImpl
	MessageRepo  repository.MessageRepositoryImpl

	// Services
	MessageSender       *services.EnhancedMessageSender
	ChannelValidator    *services.ChannelValidator
	TemplateRenderer    *services.DefaultTemplateRenderer
	NotificationService *external.DefaultNotificationService

	// Use Cases
	CreateChannelUseCase *usecases.CreateChannelUseCase
	GetChannelUseCase    *usecases.GetChannelUseCase
	ListChannelsUseCase  *usecases.ListChannelsUseCase
	UpdateChannelUseCase *usecases.UpdateChannelUseCase
	DeleteChannelUseCase *usecases.DeleteChannelUseCase

	// Infrastructure
	NATSClient *messaging.NATSClient
	Logger     *logger.Logger
	Config     *config.Config
}

// buildContainer creates and wires all dependencies
func buildContainer(db *database.PostgresDB, natsClient *messaging.NATSClient, log *logger.Logger, cfg *config.Config) *Container {
	// Initialize repositories
	channelRepo := repository.NewChannelRepositoryImpl(db.DB)
	templateRepo := repository.NewTemplateRepositoryImpl(db.DB)
	messageRepo := repository.NewMessageRepositoryImpl(db.DB)

	// Initialize external services
	messageSenderFactory := external.NewDefaultMessageSenderFactory(30 * time.Second)
	notificationService := external.NewDefaultNotificationService(messageSenderFactory)

	// Initialize domain services
	templateRenderer := services.NewDefaultTemplateRenderer()
	channelValidator := services.NewChannelValidator(channelRepo, templateRepo)
	messageSender := services.NewEnhancedMessageSender(
		channelRepo,
		templateRepo,
		messageRepo,
		templateRenderer,
		notificationService,
		log,
	)

	// Initialize use cases
	createChannelUseCase := usecases.NewCreateChannelUseCase(channelRepo, channelValidator)
	getChannelUseCase := usecases.NewGetChannelUseCase(channelRepo)
	listChannelsUseCase := usecases.NewListChannelsUseCase(channelRepo)
	updateChannelUseCase := usecases.NewUpdateChannelUseCase(channelRepo, channelValidator)
	deleteChannelUseCase := usecases.NewDeleteChannelUseCase(channelRepo, channelValidator)

	return &Container{
		// Repositories
		ChannelRepo:  *channelRepo,
		TemplateRepo: *templateRepo,
		MessageRepo:  *messageRepo,

		// Services
		MessageSender:       messageSender,
		ChannelValidator:    channelValidator,
		TemplateRenderer:    templateRenderer,
		NotificationService: notificationService,

		// Use Cases
		CreateChannelUseCase: createChannelUseCase,
		GetChannelUseCase:    getChannelUseCase,
		ListChannelsUseCase:  listChannelsUseCase,
		UpdateChannelUseCase: updateChannelUseCase,
		DeleteChannelUseCase: deleteChannelUseCase,

		// Infrastructure
		NATSClient: natsClient,
		Logger:     log,
		Config:     cfg,
	}
}

// buildHTTPHandler creates HTTP handler (placeholder)
func buildHTTPHandler(container *Container) http.Handler {
	mux := http.NewServeMux()
	
	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "service": "channel-api"}`))
	})

	// Database health check
	mux.HandleFunc("/health/db", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// This would need database health check implementation
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "component": "database"}`))
	})

	// NATS health check
	mux.HandleFunc("/health/nats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if container.NATSClient.IsConnected() {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "healthy", "component": "nats"}`))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status": "unhealthy", "component": "nats"}`))
		}
	})

	return mux
}

// initializeNATSHandlers sets up NATS message handlers
func initializeNATSHandlers(natsClient *messaging.NATSClient, container *Container, log *logger.Logger) error {
	// Subscribe to channel management topics
	subjects := []string{
		"channel.create",
		"channel.get", 
		"channel.list",
		"channel.update",
		"channel.delete",
	}

	for _, subject := range subjects {
		handler := createNATSHandler(subject, container, log)
		if _, err := natsClient.Subscribe(subject, handler); err != nil {
			return fmt.Errorf("failed to subscribe to %s: %w", subject, err)
		}
	}

	log.Info("NATS handlers initialized successfully", zap.Strings("subjects", subjects))
	return nil
}

// createNATSHandler creates a NATS message handler for a specific subject
func createNATSHandler(subject string, container *Container, log *logger.Logger) func(msg *nats.Msg) {
	return func(msg *nats.Msg) {
		log.Debug("Received NATS message",
			zap.String("subject", msg.Subject),
			zap.Int("data_size", len(msg.Data)))

		// Parse NATS message
		natsMessage, err := messaging.FromJSON(msg.Data)
		if err != nil {
			log.Error("Failed to parse NATS message", zap.Error(err))
			return
		}

		// Create response message
		var response *messaging.NATSMessage

		// Route to appropriate handler based on subject
		switch subject {
		case "channel.create":
			response = handleCreateChannel(natsMessage, container, log)
		case "channel.get":
			response = handleGetChannel(natsMessage, container, log)
		case "channel.list":
			response = handleListChannels(natsMessage, container, log)
		case "channel.update":
			response = handleUpdateChannel(natsMessage, container, log)
		case "channel.delete":
			response = handleDeleteChannel(natsMessage, container, log)
		default:
			response = messaging.NewNATSErrorMessage("UNSUPPORTED_OPERATION", "Unsupported operation", 400)
		}

		// Set correlation IDs
		response.SetRequestID(natsMessage.ReqSeqID)
		if response.RspSeqID == "" {
			response.SetResponseID(fmt.Sprintf("rsp_%d", time.Now().UnixNano()))
		}

		// Send response
		if msg.Reply != "" {
			responseData, err := response.ToJSON()
			if err != nil {
				log.Error("Failed to marshal response", zap.Error(err))
				return
			}

			if err := msg.Respond(responseData); err != nil {
				log.Error("Failed to send NATS response", zap.Error(err))
			}
		}
	}
}

// Placeholder NATS handlers - these would implement the actual logic
func handleCreateChannel(msg *messaging.NATSMessage, container *Container, log *logger.Logger) *messaging.NATSMessage {
	// Implementation would go here
	return messaging.NewNATSErrorMessage("NOT_IMPLEMENTED", "Create channel not implemented", 501)
}

func handleGetChannel(msg *messaging.NATSMessage, container *Container, log *logger.Logger) *messaging.NATSMessage {
	// Implementation would go here
	return messaging.NewNATSErrorMessage("NOT_IMPLEMENTED", "Get channel not implemented", 501)
}

func handleListChannels(msg *messaging.NATSMessage, container *Container, log *logger.Logger) *messaging.NATSMessage {
	// Implementation would go here
	return messaging.NewNATSErrorMessage("NOT_IMPLEMENTED", "List channels not implemented", 501)
}

func handleUpdateChannel(msg *messaging.NATSMessage, container *Container, log *logger.Logger) *messaging.NATSMessage {
	// Implementation would go here
	return messaging.NewNATSErrorMessage("NOT_IMPLEMENTED", "Update channel not implemented", 501)
}

func handleDeleteChannel(msg *messaging.NATSMessage, container *Container, log *logger.Logger) *messaging.NATSMessage {
	// Implementation would go here
	return messaging.NewNATSErrorMessage("NOT_IMPLEMENTED", "Delete channel not implemented", 501)
}