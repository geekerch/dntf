package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"notification/internal/infrastructure/plugins"
)

// PluginHandler handles HTTP requests for plugin management
type PluginHandler struct {
	pluginLoader plugins.PluginLoader
}

// NewPluginHandler creates a new plugin handler
func NewPluginHandler(pluginLoader plugins.PluginLoader) *PluginHandler {
	return &PluginHandler{
		pluginLoader: pluginLoader,
	}
}

// LoadPluginRequest represents the request to load a plugin from source
type LoadPluginRequest struct {
	Name   string `json:"name" binding:"required"`
	Source string `json:"source" binding:"required"`
}

// LoadPlugin handles POST /api/v1/plugins/load
// @Summary Load a plugin from source code
// @Description Load a new plugin from Go source code using Yaegi interpreter
// @Tags plugins
// @Accept json
// @Produce json
// @Param request body LoadPluginRequest true "Load plugin request"
// @Success 200 {object} map[string]interface{} "Plugin loaded successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security ApiKeyAuth
// @Router /plugins/load [post]
func (h *PluginHandler) LoadPlugin(c *gin.Context) {
	var req LoadPluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data":  nil,
			"error": map[string]interface{}{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body: " + err.Error(),
			},
		})
		return
	}

	err := h.pluginLoader.LoadPluginFromSource(req.Name, req.Source)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data":  nil,
			"error": map[string]interface{}{
				"code":    "LOAD_PLUGIN_FAILED",
				"message": "Failed to load plugin: " + err.Error(),
			},
		})
		return
	}

	// Get plugin status after loading
	status, err := h.pluginLoader.GetPluginStatus(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"data":  nil,
			"error": map[string]interface{}{
				"code":    "GET_STATUS_FAILED",
				"message": "Plugin loaded but failed to get status: " + err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  status,
		"error": nil,
	})
}

// ListPlugins handles GET /api/v1/plugins
// @Summary List all loaded plugins
// @Description Get a list of all loaded plugins with their statuses
// @Tags plugins
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Success response with plugins list"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security ApiKeyAuth
// @Router /plugins [get]
func (h *PluginHandler) ListPlugins(c *gin.Context) {
	statuses := h.pluginLoader.GetAllPluginStatuses()

	c.JSON(http.StatusOK, gin.H{
		"data":  statuses,
		"error": nil,
	})
}

// GetPlugin handles GET /api/v1/plugins/{name}
// @Summary Get plugin status by name
// @Description Get the status and information of a specific plugin
// @Tags plugins
// @Accept json
// @Produce json
// @Param name path string true "Plugin name"
// @Success 200 {object} map[string]interface{} "Success response with plugin status"
// @Failure 404 {object} map[string]interface{} "Plugin not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security ApiKeyAuth
// @Router /plugins/{name} [get]
func (h *PluginHandler) GetPlugin(c *gin.Context) {
	name := c.Param("name")

	status, err := h.pluginLoader.GetPluginStatus(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"data":  nil,
			"error": map[string]interface{}{
				"code":    "PLUGIN_NOT_FOUND",
				"message": "Plugin not found: " + err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  status,
		"error": nil,
	})
}

// UnloadPlugin handles DELETE /api/v1/plugins/{name}
// @Summary Unload a plugin
// @Description Unload a specific plugin by name
// @Tags plugins
// @Accept json
// @Produce json
// @Param name path string true "Plugin name"
// @Success 200 {object} map[string]interface{} "Plugin unloaded successfully"
// @Failure 404 {object} map[string]interface{} "Plugin not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security ApiKeyAuth
// @Router /plugins/{name} [delete]
func (h *PluginHandler) UnloadPlugin(c *gin.Context) {
	name := c.Param("name")

	err := h.pluginLoader.UnloadPlugin(name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data":  nil,
			"error": map[string]interface{}{
				"code":    "UNLOAD_PLUGIN_FAILED",
				"message": "Failed to unload plugin: " + err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  map[string]interface{}{"unloaded": true, "name": name},
		"error": nil,
	})
}

// LoadPluginFromFile handles POST /api/v1/plugins/load-file
// @Summary Load a plugin from file path
// @Description Load a plugin from a file path on the server
// @Tags plugins
// @Accept json
// @Produce json
// @Param request body map[string]string true "Load plugin from file request"
// @Success 200 {object} map[string]interface{} "Plugin loaded successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security ApiKeyAuth
// @Router /plugins/load-file [post]
func (h *PluginHandler) LoadPluginFromFile(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data":  nil,
			"error": map[string]interface{}{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body: " + err.Error(),
			},
		})
		return
	}

	filePath, exists := req["file_path"]
	if !exists || filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"data":  nil,
			"error": map[string]interface{}{
				"code":    "INVALID_REQUEST",
				"message": "file_path is required",
			},
		})
		return
	}

	err := h.pluginLoader.LoadPlugin(filePath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data":  nil,
			"error": map[string]interface{}{
				"code":    "LOAD_PLUGIN_FAILED",
				"message": "Failed to load plugin from file: " + err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  map[string]interface{}{"loaded": true, "file_path": filePath},
		"error": nil,
	})
}