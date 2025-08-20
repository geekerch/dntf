package routes

import (
	"github.com/gin-gonic/gin"

	"notification/internal/presentation/http/handlers"
	"notification/internal/presentation/http/middleware"

	"github.com/swaggo/gin-swagger" // gin-swagger middleware
	"github.com/swaggo/files" // swagger embed files
)

// RouterConfig holds the configuration for setting up routes
type RouterConfig struct {
	ChannelHandler     *handlers.ChannelHandler
	CQRSChannelHandler *handlers.CQRSChannelHandler
	TemplateHandler    *handlers.TemplateHandler
	MessageHandler     *handlers.MessageHandler
	
	// CQRS handlers
	CQRSTemplateHandler *handlers.CQRSTemplateHandler
	CQRSMessageHandler  *handlers.CQRSMessageHandler
	
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
	// @Summary Health check
	// @Description Check if the API is running and healthy
	// @Tags system
	// @Produce json
	// @Success 200 {object} models.HealthResponse "API is healthy"
	// @Router /health [get]
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
		// @Summary API information
		// @Description Get information about the API and available endpoints
		// @Tags system
		// @Produce json
		// @Success 200 {object} models.InfoResponse "API information"
		// @Router /api/v1/public/info [get]
		publicV1.GET("/info", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"service": "notification-api",
				"version": "1.0.0",
				"endpoints": []string{
					"/api/v1/channels",
					"/api/v1/templates", 
					"/api/v1/messages",
					"/api/v2/channels (CQRS)",
					"/api/v2/templates (CQRS)",
					"/api/v2/messages (CQRS)",
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

		// Template routes
		if config.TemplateHandler != nil {
			SetupTemplateRoutes(protectedV1, config.TemplateHandler)
		}

		// Message routes
		if config.MessageHandler != nil {
			SetupMessageRoutes(protectedV1, config.MessageHandler)
		}

		// Plugin management routes
		SetupPluginRoutes(protectedV1)
	}

	// CQRS API v2 routes (using CQRS pattern)
	cqrsV2 := router.Group("/api/v2")
	middlewareManager.SetupProtectedRoutes(cqrsV2)
	{
		// CQRS Channel routes
		if config.CQRSChannelHandler != nil {
			SetupCQRSChannelRoutes(cqrsV2, config.CQRSChannelHandler)
		}
		
		// CQRS Template routes
		if config.CQRSTemplateHandler != nil {
			SetupCQRSTemplateRoutes(cqrsV2, config.CQRSTemplateHandler)
		}
		
		// CQRS Message routes
		if config.CQRSMessageHandler != nil {
			SetupCQRSMessageRoutes(cqrsV2, config.CQRSMessageHandler)
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

	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Handle 404
	router.NoRoute(middleware.NotFoundHandler())
	router.NoMethod(middleware.MethodNotAllowedHandler())

	return router
}