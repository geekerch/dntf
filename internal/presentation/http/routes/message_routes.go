package routes

import (
	"github.com/gin-gonic/gin"

	"notification/internal/presentation/http/handlers"
)

// SetupMessageRoutes sets up the message routes.
func SetupMessageRoutes(router *gin.RouterGroup, messageHandler *handlers.MessageHandler) {
	// Message routes
	messageRouter := router.Group("/messages")

	// Message operations
	messageRouter.POST("/send", messageHandler.SendMessage)
	messageRouter.GET("/:id", messageHandler.GetMessage)
}