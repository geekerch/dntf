package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"notification/internal/application/template/dtos"
	"notification/internal/application/template/usecases"
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
			"success": false,
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	response, err := h.createTemplateUC.Execute(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Failed to create template",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
		"message": "Template created successfully",
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
			"success": false,
			"error":   "Template not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
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
// @Param page query int false "Page number" default(1)
// @Param size query int false "Page size" default(20)
// @Success 200 {object} map[string]interface{} "Success response with templates list"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security ApiKeyAuth
// @Router /templates [get]
func (h *TemplateHandler) ListTemplates(c *gin.Context) {
	var req dtos.ListTemplatesRequest

	// Parse query parameters
	if channelType := c.Query("channelType"); channelType != "" {
		// Note: You might want to add validation for channel type here
		// For now, we'll assume it's valid
	}

	// Parse tags
	if tags := c.QueryArray("tags"); len(tags) > 0 {
		req.Tags = tags
	}

	// Parse pagination
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			req.Page = p
		}
	}

	if size := c.Query("size"); size != "" {
		if s, err := strconv.Atoi(size); err == nil && s > 0 {
			req.Size = s
		}
	}

	response, err := h.listTemplatesUC.Execute(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to list templates",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
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
			"success": false,
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	response, err := h.updateTemplateUC.Execute(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Failed to update template",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Template updated successfully",
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
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Failed to delete template",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Template deleted successfully",
	})
}