package routes

import (
	"github.com/gin-gonic/gin"

	"notification/internal/presentation/http/handlers"
)

// SetupCQRSMessageRoutes sets up the CQRS message routes
func SetupCQRSMessageRoutes(router *gin.RouterGroup, messageHandler *handlers.CQRSMessageHandler) {
	// Message routes using CQRS pattern
	messageRouter := router.Group("/messages")

	// Message operations via CQRS
	messageRouter.POST("/send", messageHandler.SendMessage)
	messageRouter.GET("", messageHandler.ListMessages)
	messageRouter.GET("/:id", messageHandler.GetMessage)
}