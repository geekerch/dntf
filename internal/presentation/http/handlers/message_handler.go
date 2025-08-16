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
// @Summary Send a message
// @Description Send a message to multiple channels using a template
// @Tags messages
// @Accept json
// @Produce json
// @Param request body dtos.SendMessageRequest true "Send message request"
// @Success 200 {object} map[string]interface{} "Success response with message data"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security ApiKeyAuth
// @Router /messages [post]
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
// @Summary Get a message by ID
// @Description Retrieve a specific message by its ID
// @Tags messages
// @Accept json
// @Produce json
// @Param id path string true "Message ID"
// @Success 200 {object} map[string]interface{} "Success response with message data"
// @Failure 404 {object} map[string]interface{} "Message not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security ApiKeyAuth
// @Router /messages/{id} [get]
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
// @Summary List messages
// @Description Retrieve a list of messages with optional filtering
// @Tags messages
// @Accept json
// @Produce json
// @Param channelId query string false "Filter by channel ID"
// @Param status query string false "Filter by message status"
// @Param skipCount query int false "Number of items to skip" default(0)
// @Param maxResultCount query int false "Maximum number of items to return" default(20)
// @Success 200 {object} map[string]interface{} "Success response with messages list"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security ApiKeyAuth
// @Router /messages [get]
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