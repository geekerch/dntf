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
	messageRouter.POST("", messageHandler.SendMessage)  // POST /api/v1/messages for sending messages
	messageRouter.GET("", messageHandler.ListMessages)  // GET /api/v1/messages for listing messages
	messageRouter.GET("/:id", messageHandler.GetMessage) // GET /api/v1/messages/{id} for getting specific message
}