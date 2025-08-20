package routes

import (
	"github.com/gin-gonic/gin"

	"notification/internal/infrastructure/plugins"
	"notification/internal/presentation/http/handlers"
)

// SetupPluginRoutes sets up the plugin management routes
func SetupPluginRoutes(router *gin.RouterGroup) {
	pluginLoader := plugins.GetPluginLoader()
	pluginHandler := handlers.NewPluginHandler(pluginLoader)

	// Plugin management routes
	pluginGroup := router.Group("/plugins")
	{
		pluginGroup.POST("/load", pluginHandler.LoadPlugin)
		pluginGroup.POST("/load-file", pluginHandler.LoadPluginFromFile)
		pluginGroup.GET("", pluginHandler.ListPlugins)
		pluginGroup.GET("/:name", pluginHandler.GetPlugin)
		pluginGroup.DELETE("/:name", pluginHandler.UnloadPlugin)
	}
}