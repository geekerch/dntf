package routes

import (
	"github.com/gin-gonic/gin"

	"notification/internal/presentation/http/handlers"
)

// SetupCQRSChannelRoutes sets up the CQRS routes for channel operations
func SetupCQRSChannelRoutes(router *gin.RouterGroup, cqrsChannelHandler *handlers.CQRSChannelHandler) {
	channels := router.Group("/channels")
	{
		channels.POST("", cqrsChannelHandler.CreateChannel)
		channels.GET("", cqrsChannelHandler.ListChannels)
		channels.GET("/:id", cqrsChannelHandler.GetChannel)
		channels.PUT("/:id", cqrsChannelHandler.UpdateChannel)
		channels.DELETE("/:id", cqrsChannelHandler.DeleteChannel)
	}
}