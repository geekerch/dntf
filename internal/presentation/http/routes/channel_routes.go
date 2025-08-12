package routes

import (
	"github.com/gin-gonic/gin"

	"notification/internal/presentation/http/handlers"
)

// SetupChannelRoutes sets up the routes for channel operations
func SetupChannelRoutes(router *gin.RouterGroup, channelHandler *handlers.ChannelHandler) {
	channels := router.Group("/channels")
	{
		channels.POST("", channelHandler.CreateChannel)
		channels.GET("", channelHandler.ListChannels)
		channels.GET("/:id", channelHandler.GetChannel)
		channels.PUT("/:id", channelHandler.UpdateChannel)
		channels.DELETE("/:id", channelHandler.DeleteChannel)
	}
}