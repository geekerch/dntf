package routes

import (
	"github.com/gin-gonic/gin"

	"notification/internal/presentation/http/handlers"
	"notification/internal/presentation/http/middleware"
)

// RouterConfig holds the configuration for setting up routes
type RouterConfig struct {
	ChannelHandler     *handlers.ChannelHandler
	CQRSChannelHandler *handlers.CQRSChannelHandler
	// Add other handlers here as they are implemented
	// TemplateHandler *handlers.TemplateHandler
	// MessageHandler  *handlers.MessageHandler
	
	// Middleware configuration
	MiddlewareConfig *middleware.MiddlewareConfig
}

// SetupRouter sets up the main router with all routes and middleware
func SetupRouter(config *RouterConfig) *gin.Engine {
	// Set Gin mode based on environment
	gin.SetMode(gin.ReleaseMode) // Can be configured via environment variable

	router := gin.New()

	// Setup middleware using middleware manager
	middlewareConfig := config.MiddlewareConfig
	if middlewareConfig == nil {
		middlewareConfig = middleware.DefaultMiddlewareConfig()
	}
	
	middlewareManager := middleware.NewMiddlewareManager(middlewareConfig)
	middlewareManager.SetupMiddleware(router)

	// Health check endpoint (public)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "notification-api",
			"version": "1.0.0",
		})
	})

	// Metrics endpoint (public, but could be protected)
	router.GET("/metrics", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"metrics": gin.H{
				"uptime": "placeholder", // TODO: Implement actual metrics
			},
		})
	})

	// Public API v1 routes (no authentication required)
	publicV1 := router.Group("/api/v1/public")
	{
		// Add public endpoints here if needed
		publicV1.GET("/info", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"service": "notification-api",
				"version": "1.0.0",
				"endpoints": []string{
					"/api/v1/channels",
					"/api/v1/templates",
					"/api/v1/messages",
				},
			})
		})
	}

	// Protected API v1 routes (authentication required)
	protectedV1 := router.Group("/api/v1")
	middlewareManager.SetupProtectedRoutes(protectedV1)
	{
		// Traditional Channel routes
		if config.ChannelHandler != nil {
			SetupChannelRoutes(protectedV1, config.ChannelHandler)
		}

		// TODO: Add template routes when TemplateHandler is implemented
		// if config.TemplateHandler != nil {
		//     SetupTemplateRoutes(protectedV1, config.TemplateHandler)
		// }

		// TODO: Add message routes when MessageHandler is implemented
		// if config.MessageHandler != nil {
		//     SetupMessageRoutes(protectedV1, config.MessageHandler)
		// }
	}

	// CQRS API v2 routes (using CQRS pattern)
	cqrsV2 := router.Group("/api/v2")
	middlewareManager.SetupProtectedRoutes(cqrsV2)
	{
		// CQRS Channel routes
		if config.CQRSChannelHandler != nil {
			SetupCQRSChannelRoutes(cqrsV2, config.CQRSChannelHandler)
		}
	}

	// Admin routes (additional authentication/authorization)
	adminV1 := router.Group("/api/v1/admin")
	middlewareManager.SetupAdminRoutes(adminV1)
	{
		// Admin endpoints
		adminV1.GET("/stats", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Admin stats endpoint",
				"user":    c.GetString("auth_user"),
			})
		})

		adminV1.GET("/config", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Admin config endpoint",
				"user":    c.GetString("auth_user"),
			})
		})
	}

	// Handle 404
	router.NoRoute(middleware.NotFoundHandler())
	router.NoMethod(middleware.MethodNotAllowedHandler())

	return router
}