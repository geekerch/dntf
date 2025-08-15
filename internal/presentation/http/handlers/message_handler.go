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
	listMessagesUC *usecases.ListMessagesUseCase
}

// NewMessageHandler creates a new MessageHandler.
func NewMessageHandler(
	sendMessageUC *usecases.SendMessageUseCase,
	getMessageUC *usecases.GetMessageUseCase,
	listMessagesUC *usecases.ListMessagesUseCase,
) *MessageHandler {
	return &MessageHandler{
		sendMessageUC: sendMessageUC,
		getMessageUC:  getMessageUC,
		listMessagesUC: listMessagesUC,
	}
}

// SendMessage handles POST /api/v1/messages
func (h *MessageHandler) SendMessage(c *gin.Context) {
	var req dtos.SendMessageRequest
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

	response, err := h.sendMessageUC.Execute(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data":  nil,
			"error": map[string]interface{}{
				"code":    "SEND_MESSAGE_FAILED",
				"message": "Failed to send message: " + err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  response,
		"error": nil,
	})
}

// GetMessage handles GET /api/v1/messages/{id}
func (h *MessageHandler) GetMessage(c *gin.Context) {
	id := c.Param("id")

	response, err := h.getMessageUC.Execute(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"data":  nil,
			"error": map[string]interface{}{
				"code":    "MESSAGE_NOT_FOUND",
				"message": "Message not found: " + err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  response,
		"error": nil,
	})
}

// ListMessages handles GET /api/v1/messages
func (h *MessageHandler) ListMessages(c *gin.Context) {
	var req dtos.ListMessagesRequest
	
	// Parse query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data":  nil,
			"error": map[string]interface{}{
				"code":    "INVALID_REQUEST",
				"message": "Invalid query parameters: " + err.Error(),
			},
		})
		return
	}

	response, err := h.listMessagesUC.Execute(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data":  nil,
			"error": map[string]interface{}{
				"code":    "LIST_MESSAGES_FAILED",
				"message": "Failed to list messages: " + err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  response,
		"error": nil,
	})
}