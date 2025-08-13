package routes

import (
	"github.com/gin-gonic/gin"

	"notification/internal/presentation/http/handlers"
)

// SetupTemplateRoutes sets up the template routes.
func SetupTemplateRoutes(router *gin.RouterGroup, templateHandler *handlers.TemplateHandler) {
	// Template routes
	templateRouter := router.Group("/templates")

	// CRUD operations
	templateRouter.POST("", templateHandler.CreateTemplate)
	templateRouter.GET("", templateHandler.ListTemplates)
	templateRouter.GET("/:id", templateHandler.GetTemplate)
	templateRouter.PUT("/:id", templateHandler.UpdateTemplate)
	templateRouter.DELETE("/:id", templateHandler.DeleteTemplate)
}