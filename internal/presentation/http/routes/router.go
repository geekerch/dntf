package routes

import (
	"github.com/gin-gonic/gin"

	"notification/internal/presentation/http/handlers"
	"notification/internal/presentation/http/middleware"
)

// RouterConfig holds the configuration for setting up routes
type RouterConfig struct {
	ChannelHandler *handlers.ChannelHandler
	// Add other handlers here as they are implemented
	// TemplateHandler *handlers.TemplateHandler
	// MessageHandler  *handlers.MessageHandler
}

// SetupRouter sets up the main router with all routes and middleware
func SetupRouter(config *RouterConfig) *gin.Engine {
	// Set Gin mode based on environment
	gin.SetMode(gin.ReleaseMode) // Can be configured via environment variable

	router := gin.New()

	// Global middleware
	router.Use(middleware.RequestLogger())
	router.Use(middleware.RequestID())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.CORS())
	router.Use(middleware.ResponseFormatter())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "notification-api",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Channel routes
		if config.ChannelHandler != nil {
			SetupChannelRoutes(v1, config.ChannelHandler)
		}

		// TODO: Add template routes when TemplateHandler is implemented
		// if config.TemplateHandler != nil {
		//     SetupTemplateRoutes(v1, config.TemplateHandler)
		// }

		// TODO: Add message routes when MessageHandler is implemented
		// if config.MessageHandler != nil {
		//     SetupMessageRoutes(v1, config.MessageHandler)
		// }
	}

	// Handle 404
	router.NoRoute(middleware.NotFoundHandler())
	router.NoMethod(middleware.MethodNotAllowedHandler())

	return router
}