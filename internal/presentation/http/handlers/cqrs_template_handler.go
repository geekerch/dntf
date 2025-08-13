package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"notification/internal/application/cqrs"
	templatecqrs "notification/internal/application/cqrs/template"
	"notification/internal/application/template/dtos"
)

// CQRSTemplateHandler handles CQRS HTTP requests for templates
type CQRSTemplateHandler struct {
	cqrsFacade *cqrs.CQRSFacade
}

// NewCQRSTemplateHandler creates a new CQRS template handler
func NewCQRSTemplateHandler(cqrsFacade *cqrs.CQRSFacade) *CQRSTemplateHandler {
	return &CQRSTemplateHandler{
		cqrsFacade: cqrsFacade,
	}
}

// CreateTemplate handles POST /api/v2/templates
func (h *CQRSTemplateHandler) CreateTemplate(c *gin.Context) {
	var req dtos.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Create command
	cmd := templatecqrs.NewCreateTemplateCommand(&req)

	// Execute command
	result, err := h.cqrsFacade.Send(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Failed to create template",
			"details": err.Error(),
		})
		return
	}

	// Type assert the result
	response, ok := result.Data.(*dtos.TemplateResponse)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Invalid response type",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
		"message": "Template created successfully via CQRS",
	})
}

// GetTemplate handles GET /api/v2/templates/{id}
func (h *CQRSTemplateHandler) GetTemplate(c *gin.Context) {
	id := c.Param("id")

	// Create query
	query := templatecqrs.NewGetTemplateQuery(id)

	// Execute query
	result, err := h.cqrsFacade.Query(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Template not found",
			"details": err.Error(),
		})
		return
	}

	// Type assert the result
	response, ok := result.Data.(*dtos.TemplateResponse)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Invalid response type",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ListTemplates handles GET /api/v2/templates
func (h *CQRSTemplateHandler) ListTemplates(c *gin.Context) {
	// Create query
	query := templatecqrs.NewListTemplatesQuery()

	// Parse query parameters
	if channelType := c.Query("channelType"); channelType != "" {
		query.WithChannelType(channelType)
	}

	if tags := c.QueryArray("tags"); len(tags) > 0 {
		query.WithTags(tags)
	}

	// Parse pagination
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			if size := c.Query("size"); size != "" {
				if s, err := strconv.Atoi(size); err == nil && s > 0 {
					offset := (p - 1) * s
					query.WithPagination(offset, s)
				}
			}
		}
	}

	// Parse sorting
	if sortBy := c.Query("sortBy"); sortBy != "" {
		order := c.DefaultQuery("order", "asc")
		query.WithSorting(sortBy, order)
	}

	// Execute query
	result, err := h.cqrsFacade.Query(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to list templates",
			"details": err.Error(),
		})
		return
	}

	// Type assert the result
	response, ok := result.Data.(*dtos.ListTemplatesResponse)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Invalid response type",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// UpdateTemplate handles PUT /api/v2/templates/{id}
func (h *CQRSTemplateHandler) UpdateTemplate(c *gin.Context) {
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

	// Create command
	cmd := templatecqrs.NewUpdateTemplateCommand(id, &req)

	// Execute command
	result, err := h.cqrsFacade.Send(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Failed to update template",
			"details": err.Error(),
		})
		return
	}

	// Type assert the result
	response, ok := result.Data.(*dtos.TemplateResponse)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Invalid response type",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Template updated successfully via CQRS",
	})
}

// DeleteTemplate handles DELETE /api/v2/templates/{id}
func (h *CQRSTemplateHandler) DeleteTemplate(c *gin.Context) {
	id := c.Param("id")

	// Create command
	cmd := templatecqrs.NewDeleteTemplateCommand(id)

	// Execute command
	_, err := h.cqrsFacade.Send(c.Request.Context(), cmd)
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
		"message": "Template deleted successfully via CQRS",
	})
}