package routes

import (
	"github.com/gin-gonic/gin"

	"notification/internal/presentation/http/handlers"
)

// SetupCQRSTemplateRoutes sets up the CQRS template routes
func SetupCQRSTemplateRoutes(router *gin.RouterGroup, templateHandler *handlers.CQRSTemplateHandler) {
	// Template routes using CQRS pattern
	templateRouter := router.Group("/templates")

	// CRUD operations via CQRS
	templateRouter.POST("", templateHandler.CreateTemplate)
	templateRouter.GET("", templateHandler.ListTemplates)
	templateRouter.GET("/:id", templateHandler.GetTemplate)
	templateRouter.PUT("/:id", templateHandler.UpdateTemplate)
	templateRouter.DELETE("/:id", templateHandler.DeleteTemplate)
}