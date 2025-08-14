package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"notification/internal/application/cqrs"
	messagecqrs "notification/internal/application/cqrs/message"
	"notification/internal/application/message/dtos"
)

// CQRSMessageHandler handles CQRS HTTP requests for messages
type CQRSMessageHandler struct {
	cqrsFacade *cqrs.CQRSFacade
}

// NewCQRSMessageHandler creates a new CQRS message handler
func NewCQRSMessageHandler(cqrsFacade *cqrs.CQRSFacade) *CQRSMessageHandler {
	return &CQRSMessageHandler{
		cqrsFacade: cqrsFacade,
	}
}

// SendMessage handles POST /api/v2/messages/send
func (h *CQRSMessageHandler) SendMessage(c *gin.Context) {
	var req dtos.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Create command
	cmd := messagecqrs.NewSendMessageCommand(&req)

	// Execute command
	result, err := h.cqrsFacade.Send(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Failed to send message",
			"details": err.Error(),
		})
		return
	}

	// Type assert the result
	response, ok := result.Data.(*dtos.MessageResponse)
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
		"message": "Message sent successfully via CQRS",
	})
}

// GetMessage handles GET /api/v2/messages/{id}
func (h *CQRSMessageHandler) GetMessage(c *gin.Context) {
	id := c.Param("id")

	// Create query
	query := messagecqrs.NewGetMessageQuery(id)

	// Execute query
	result, err := h.cqrsFacade.Query(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Message not found",
			"details": err.Error(),
		})
		return
	}

	// Type assert the result
	response, ok := result.Data.(*dtos.MessageResponse)
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

// ListMessages handles GET /api/v2/messages
func (h *CQRSMessageHandler) ListMessages(c *gin.Context) {
	// Parse query parameters
	channelID := c.Query("channelId")
	status := c.Query("status")
	skipCount := 0
	maxResultCount := 10

	// Parse pagination parameters
	if skip := c.Query("skipCount"); skip != "" {
		if parsed, err := strconv.Atoi(skip); err == nil && parsed >= 0 {
			skipCount = parsed
		}
	}
	if limit := c.Query("maxResultCount"); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil && parsed > 0 && parsed <= 1000 {
			maxResultCount = parsed
		}
	}

	// Create query
	query := messagecqrs.NewListMessagesQuery()
	
	if channelID != "" {
		query.WithChannelID(channelID)
	}
	if status != "" {
		query.WithStatus(status)
	}
	query.WithPagination(skipCount, maxResultCount)

	// Execute query
	result, err := h.cqrsFacade.Query(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Failed to list messages",
			"details": err.Error(),
		})
		return
	}

	// Type assert the result
	response, ok := result.Data.(*dtos.ListMessagesResponse)
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