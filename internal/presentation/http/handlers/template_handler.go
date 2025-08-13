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