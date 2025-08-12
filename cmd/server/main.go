package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"notification/internal/application/channel/usecases"
	"notification/internal/domain/services"
	"notification/internal/infrastructure/external"
	"notification/internal/infrastructure/messaging"
	"notification/internal/infrastructure/repository"
	"notification/internal/presentation"
	"notification/internal/presentation/http/handlers"
	"notification/internal/presentation/http/middleware"
	natshandlers "notification/internal/presentation/nats/handlers"
	"notification/pkg/config"
	"notification/pkg/database"
	"notification/pkg/logger"
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

	log.Info("Starting Notification server",
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

	// Initialize HTTP handlers
	channelHandler := handlers.NewChannelHandler(
		container.CreateChannelUseCase,
		container.GetChannelUseCase,
		container.ListChannelsUseCase,
		container.UpdateChannelUseCase,
		container.DeleteChannelUseCase,
	)

	// Initialize NATS handler manager
	natsHandlerConfig := &natshandlers.HandlerConfig{
		NATSConn:             natsClient.GetConnection(),
		CreateChannelUseCase: container.CreateChannelUseCase,
		GetChannelUseCase:    container.GetChannelUseCase,
		ListChannelsUseCase:  container.ListChannelsUseCase,
		UpdateChannelUseCase: container.UpdateChannelUseCase,
		DeleteChannelUseCase: container.DeleteChannelUseCase,
	}
	natsManager := natshandlers.NewHandlerManager(natsHandlerConfig)

	// Initialize middleware configuration based on environment
	var middlewareConfig *middleware.MiddlewareConfig
	// For now, use development config as default
	// TODO: Add Environment field to config.Config
	middlewareConfig = middleware.DevelopmentMiddlewareConfig()

	// Initialize presentation layer server
	serverConfig := &presentation.ServerConfig{
		HTTPPort:         fmt.Sprintf("%d", cfg.Server.Port),
		HTTPTimeout:      time.Duration(cfg.Server.ReadTimeout) * time.Second,
		ChannelHandler:   channelHandler,
		NATSManager:      natsManager,
		MiddlewareConfig: middlewareConfig,
	}
	server := presentation.NewServer(serverConfig)

	// Start the presentation layer server
	ctx := context.Background()
	if err := server.Start(ctx); err != nil {
		log.Fatal("Failed to start presentation layer server", zap.Error(err))
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Stop(shutdownCtx); err != nil {
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
	notificationServiceAdapter := external.NewNotificationServiceAdapter(notificationService)

	// Initialize domain services
	templateRenderer := services.NewDefaultTemplateRenderer()
	channelValidator := services.NewChannelValidator(channelRepo, templateRepo)
	messageSender := services.NewEnhancedMessageSender(
		channelRepo,
		templateRepo,
		messageRepo,
		templateRenderer,
		notificationServiceAdapter,
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

