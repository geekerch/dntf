package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"notification/internal/application/message/dtos"
	"notification/internal/application/message/usecases"
)

// MessageHandler handles HTTP requests for messages.
type MessageHandler struct {
	sendMessageUC *usecases.SendMessageUseCase
	getMessageUC  *usecases.GetMessageUseCase
}

// NewMessageHandler creates a new MessageHandler.
func NewMessageHandler(
	sendMessageUC *usecases.SendMessageUseCase,
	getMessageUC *usecases.GetMessageUseCase,
) *MessageHandler {
	return &MessageHandler{
		sendMessageUC: sendMessageUC,
		getMessageUC:  getMessageUC,
	}
}

// SendMessage handles POST /api/v1/messages/send
func (h *MessageHandler) SendMessage(c *gin.Context) {
	var req dtos.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	response, err := h.sendMessageUC.Execute(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Failed to send message",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
		"message": "Message sent successfully",
	})
}

// GetMessage handles GET /api/v1/messages/{id}
func (h *MessageHandler) GetMessage(c *gin.Context) {
	id := c.Param("id")

	response, err := h.getMessageUC.Execute(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Message not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}