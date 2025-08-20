package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"notification/internal/application/template/dtos"
	"notification/internal/application/template/usecases"
	"notification/internal/domain/shared"
)

// TemplateHandler handles HTTP requests for templates.
type TemplateHandler struct {
	createTemplateUC *usecases.CreateTemplateUseCase
	getTemplateUC    *usecases.GetTemplateUseCase
	listTemplatesUC  *usecases.ListTemplatesUseCase
	updateTemplateUC *usecases.UpdateTemplateUseCase
	deleteTemplateUC *usecases.DeleteTemplateUseCase
}

// NewTemplateHandler creates a new TemplateHandler.
func NewTemplateHandler(
	createTemplateUC *usecases.CreateTemplateUseCase,
	getTemplateUC *usecases.GetTemplateUseCase,
	listTemplatesUC *usecases.ListTemplatesUseCase,
	updateTemplateUC *usecases.UpdateTemplateUseCase,
	deleteTemplateUC *usecases.DeleteTemplateUseCase,
) *TemplateHandler {
	return &TemplateHandler{
		createTemplateUC: createTemplateUC,
		getTemplateUC:    getTemplateUC,
		listTemplatesUC:  listTemplatesUC,
		updateTemplateUC: updateTemplateUC,
		deleteTemplateUC: deleteTemplateUC,
	}
}

// CreateTemplate handles POST /api/v1/templates
// @Summary Create a new template
// @Description Create a new message template for a specific channel type
// @Tags templates
// @Accept json
// @Produce json
// @Param request body dtos.CreateTemplateRequest true "Create template request"
// @Success 201 {object} map[string]interface{} "Template created successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security ApiKeyAuth
// @Router /templates [post]
func (h *TemplateHandler) CreateTemplate(c *gin.Context) {
	var req dtos.CreateTemplateRequest
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

	response, err := h.createTemplateUC.Execute(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data":  nil,
			"error": map[string]interface{}{
				"code":    "CREATE_TEMPLATE_FAILED",
				"message": "Failed to create template: " + err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":  response,
		"error": nil,
	})
}

// GetTemplate handles GET /api/v1/templates/{id}
// @Summary Get a template by ID
// @Description Retrieve a specific template by its ID
// @Tags templates
// @Accept json
// @Produce json
// @Param id path string true "Template ID"
// @Success 200 {object} map[string]interface{} "Success response with template data"
// @Failure 404 {object} map[string]interface{} "Template not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security ApiKeyAuth
// @Router /templates/{id} [get]
func (h *TemplateHandler) GetTemplate(c *gin.Context) {
	id := c.Param("id")

	response, err := h.getTemplateUC.Execute(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"data":  nil,
			"error": map[string]interface{}{
				"code":    "TEMPLATE_NOT_FOUND",
				"message": "Template not found: " + err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  response,
		"error": nil,
	})
}

// ListTemplates handles GET /api/v1/templates
// @Summary List templates
// @Description Retrieve a list of templates with optional filtering
// @Tags templates
// @Accept json
// @Produce json
// @Param channelType query string false "Filter by channel type"
// @Param tags query []string false "Filter by tags"
// @Param skipCount query int false "Number of records to skip for pagination" default(0)
// @Param maxResultCount query int false "Maximum number of records to return per page (1-100)" default(20)
// @Success 200 {object} map[string]interface{} "Success response with templates list"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security ApiKeyAuth
// @Router /templates [get]
func (h *TemplateHandler) ListTemplates(c *gin.Context) {
	var req dtos.ListTemplatesRequest

	// Parse query parameters
	if channelType := c.Query("channelType"); channelType != "" {
		if ct, err := shared.NewChannelType(channelType); err == nil {
			req.ChannelType = &ct
		}
	}

	// Parse tags
	if tags := c.QueryArray("tags"); len(tags) > 0 {
		req.Tags = tags
	}

	// Parse pagination
	if skipCount := c.Query("skipCount"); skipCount != "" {
		if sc, err := strconv.Atoi(skipCount); err == nil && sc >= 0 {
			req.SkipCount = sc
		}
	}

	if maxResultCount := c.Query("maxResultCount"); maxResultCount != "" {
		if mrc, err := strconv.Atoi(maxResultCount); err == nil && mrc > 0 {
			req.MaxResultCount = mrc
		}
	}

	response, err := h.listTemplatesUC.Execute(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data":  nil,
			"error": map[string]interface{}{
				"code":    "LIST_TEMPLATES_FAILED",
				"message": "Failed to list templates: " + err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  response,
		"error": nil,
	})
}

// UpdateTemplate handles PUT /api/v1/templates/{id}
// @Summary Update a template
// @Description Update an existing template by its ID
// @Tags templates
// @Accept json
// @Produce json
// @Param id path string true "Template ID"
// @Param request body dtos.UpdateTemplateRequest true "Update template request"
// @Success 200 {object} map[string]interface{} "Template updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Template not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security ApiKeyAuth
// @Router /templates/{id} [put]
func (h *TemplateHandler) UpdateTemplate(c *gin.Context) {
	id := c.Param("id")

	var req dtos.UpdateTemplateRequest
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

	response, err := h.updateTemplateUC.Execute(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data":  nil,
			"error": map[string]interface{}{
				"code":    "UPDATE_TEMPLATE_FAILED",
				"message": "Failed to update template: " + err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  response,
		"error": nil,
	})
}

// DeleteTemplate handles DELETE /api/v1/templates/{id}
// @Summary Delete a template
// @Description Delete an existing template by its ID
// @Tags templates
// @Accept json
// @Produce json
// @Param id path string true "Template ID"
// @Success 200 {object} map[string]interface{} "Template deleted successfully"
// @Failure 404 {object} map[string]interface{} "Template not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security ApiKeyAuth
// @Router /templates/{id} [delete]
func (h *TemplateHandler) DeleteTemplate(c *gin.Context) {
	id := c.Param("id")

	err := h.deleteTemplateUC.Execute(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data":  nil,
			"error": map[string]interface{}{
				"code":    "DELETE_TEMPLATE_FAILED",
				"message": "Failed to delete template: " + err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  map[string]interface{}{"deleted": true},
		"error": nil,
	})
}