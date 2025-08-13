package handlers

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"notification/internal/application/message/dtos"
	"notification/internal/application/message/usecases"
	"notification/pkg/logger"
)

// MessageNATSHandler handles NATS messages for messages.
type MessageNATSHandler struct {
	sendMessageUC *usecases.SendMessageUseCase
	getMessageUC  *usecases.GetMessageUseCase
	logger        logger.Logger
}

// NewMessageNATSHandler creates a new MessageNATSHandler.
func NewMessageNATSHandler(
	sendMessageUC *usecases.SendMessageUseCase,
	getMessageUC *usecases.GetMessageUseCase,
	logger logger.Logger,
) *MessageNATSHandler {
	return &MessageNATSHandler{
		sendMessageUC: sendMessageUC,
		getMessageUC:  getMessageUC,
		logger:        logger,
	}
}

// HandleSendMessage handles message sending via NATS.
func (h *MessageNATSHandler) HandleSendMessage(msg *nats.Msg) {
	var req dtos.SendMessageRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		h.logger.Error("Failed to unmarshal send message request", zap.Error(err))
		h.respondWithError(msg, "Invalid request format", err)
		return
	}

	response, err := h.sendMessageUC.Execute(context.Background(), &req)
	if err != nil {
		h.logger.Error("Failed to send message", zap.Error(err))
		h.respondWithError(msg, "Failed to send message", err)
		return
	}

	h.respondWithSuccess(msg, response, "Message sent successfully")
}

// HandleGetMessage handles getting a message via NATS.
func (h *MessageNATSHandler) HandleGetMessage(msg *nats.Msg) {
	var req struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		h.logger.Error("Failed to unmarshal get message request", zap.Error(err))
		h.respondWithError(msg, "Invalid request format", err)
		return
	}

	response, err := h.getMessageUC.Execute(context.Background(), req.ID)
	if err != nil {
		h.logger.Error("Failed to get message", zap.Error(err), zap.String("id", req.ID))
		h.respondWithError(msg, "Message not found", err)
		return
	}

	h.respondWithSuccess(msg, response, "")
}

// respondWithSuccess sends a success response via NATS.
func (h *MessageNATSHandler) respondWithSuccess(msg *nats.Msg, data interface{}, message string) {
	response := map[string]interface{}{
		"success": true,
		"data":    data,
	}
	if message != "" {
		response["message"] = message
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		h.logger.Error("Failed to marshal success response", zap.Error(err))
		return
	}

	if err := msg.Respond(responseData); err != nil {
		h.logger.Error("Failed to send success response", zap.Error(err))
	}
}

// respondWithError sends an error response via NATS.
func (h *MessageNATSHandler) respondWithError(msg *nats.Msg, message string, err error) {
	response := map[string]interface{}{
		"success": false,
		"error":   message,
		"details": err.Error(),
	}

	responseData, marshalErr := json.Marshal(response)
	if marshalErr != nil {
		h.logger.Error("Failed to marshal error response", zap.Error(marshalErr))
		return
	}

	if respondErr := msg.Respond(responseData); respondErr != nil {
		h.logger.Error("Failed to send error response", zap.Error(respondErr))
	}
}