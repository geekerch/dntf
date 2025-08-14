package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"notification/internal/application/cqrs"
	messagecqrs "notification/internal/application/cqrs/message"
	"notification/internal/application/message/dtos"
	"notification/pkg/logger"
)

// CQRSMessageNATSHandler handles CQRS NATS messages for messages
type CQRSMessageNATSHandler struct {
	cqrsFacade *cqrs.CQRSFacade
	logger     logger.Logger
}

// NewCQRSMessageNATSHandler creates a new CQRS message NATS handler
func NewCQRSMessageNATSHandler(
	cqrsFacade *cqrs.CQRSFacade,
	logger logger.Logger,
) *CQRSMessageNATSHandler {
	return &CQRSMessageNATSHandler{
		cqrsFacade: cqrsFacade,
		logger:     logger,
	}
}

// HandleSendMessage handles message sending via CQRS NATS
func (h *CQRSMessageNATSHandler) HandleSendMessage(msg *nats.Msg) {
	var req dtos.SendMessageRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		h.logger.Error("Failed to unmarshal send message request", zap.Error(err))
		h.respondWithError(msg, "INVALID_REQUEST", "Invalid request format", err)
		return
	}

	// Create command
	cmd := messagecqrs.NewSendMessageCommand(&req)

	// Execute command via CQRS
	result, err := h.cqrsFacade.Send(context.Background(), cmd)
	if err != nil {
		h.logger.Error("Failed to send message via CQRS", zap.Error(err))
		h.respondWithError(msg, "SEND_FAILED", "Failed to send message", err)
		return
	}

	// Type assert the result
	response, ok := result.Data.(*dtos.MessageResponse)
	if !ok {
		h.logger.Error("Invalid response type from CQRS send message")
		h.respondWithError(msg, "INTERNAL_ERROR", "Invalid response type", fmt.Errorf("invalid response type"))
		return
	}

	h.respondWithSuccess(msg, response, "Message sent successfully via CQRS")
}

// HandleGetMessage handles getting a message via CQRS NATS
func (h *CQRSMessageNATSHandler) HandleGetMessage(msg *nats.Msg) {
	var req struct {
		MessageID string `json:"messageId"`
	}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		h.logger.Error("Failed to unmarshal get message request", zap.Error(err))
		h.respondWithError(msg, "INVALID_REQUEST", "Invalid request format", err)
		return
	}

	// Create query
	query := messagecqrs.NewGetMessageQuery(req.MessageID)

	// Execute query via CQRS
	result, err := h.cqrsFacade.Query(context.Background(), query)
	if err != nil {
		h.logger.Error("Failed to get message via CQRS", zap.Error(err), zap.String("messageId", req.MessageID))
		h.respondWithError(msg, "NOT_FOUND", "Message not found", err)
		return
	}

	// Type assert the result
	response, ok := result.Data.(*dtos.MessageResponse)
	if !ok {
		h.logger.Error("Invalid response type from CQRS get message")
		h.respondWithError(msg, "INTERNAL_ERROR", "Invalid response type", fmt.Errorf("invalid response type"))
		return
	}

	h.respondWithSuccess(msg, response, "")
}

// HandleListMessages handles listing messages via CQRS NATS
func (h *CQRSMessageNATSHandler) HandleListMessages(msg *nats.Msg) {
	var req struct {
		ChannelID      string `json:"channelId,omitempty"`
		Status         string `json:"status,omitempty"`
		SkipCount      int    `json:"skipCount,omitempty"`
		MaxResultCount int    `json:"maxResultCount,omitempty"`
	}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		h.logger.Error("Failed to unmarshal list messages request", zap.Error(err))
		h.respondWithError(msg, "INVALID_REQUEST", "Invalid request format", err)
		return
	}

	// Create query
	query := messagecqrs.NewListMessagesQuery()
	
	if req.ChannelID != "" {
		query.WithChannelID(req.ChannelID)
	}
	if req.Status != "" {
		query.WithStatus(req.Status)
	}
	if req.MaxResultCount > 0 {
		query.WithPagination(req.SkipCount, req.MaxResultCount)
	}

	// Execute query via CQRS
	result, err := h.cqrsFacade.Query(context.Background(), query)
	if err != nil {
		h.logger.Error("Failed to list messages via CQRS", zap.Error(err))
		h.respondWithError(msg, "LIST_FAILED", "Failed to list messages", err)
		return
	}

	// Type assert the result
	response, ok := result.Data.(*dtos.ListMessagesResponse)
	if !ok {
		h.logger.Error("Invalid response type from CQRS list messages")
		h.respondWithError(msg, "INTERNAL_ERROR", "Invalid response type", fmt.Errorf("invalid response type"))
		return
	}

	h.respondWithSuccess(msg, response, "")
}

// RegisterHandlers registers all CQRS message NATS handlers
func (h *CQRSMessageNATSHandler) RegisterHandlers(nc *nats.Conn, subjectPrefix string) error {
	// Register command handlers
	if _, err := nc.Subscribe(fmt.Sprintf("%s.message.send", subjectPrefix), h.HandleSendMessage); err != nil {
		return fmt.Errorf("failed to subscribe to message.send: %w", err)
	}

	// Register query handlers
	if _, err := nc.Subscribe(fmt.Sprintf("%s.message.get", subjectPrefix), h.HandleGetMessage); err != nil {
		return fmt.Errorf("failed to subscribe to message.get: %w", err)
	}

	if _, err := nc.Subscribe(fmt.Sprintf("%s.message.list", subjectPrefix), h.HandleListMessages); err != nil {
		return fmt.Errorf("failed to subscribe to message.list: %w", err)
	}

	h.logger.Info("CQRS Message NATS handlers registered successfully")
	return nil
}

// respondWithSuccess sends a success response via NATS with CQRS format
func (h *CQRSMessageNATSHandler) respondWithSuccess(msg *nats.Msg, data interface{}, message string) {
	response := map[string]interface{}{
		"reqSeqId":   extractReqSeqId(msg),
		"rspSeqId":   generateRspSeqId(),
		"httpStatus": 200,
		"data":       data,
		"error":      nil,
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

// respondWithError sends an error response via NATS with CQRS format
func (h *CQRSMessageNATSHandler) respondWithError(msg *nats.Msg, code, message string, err error) {
	httpStatus := 400
	if code == "NOT_FOUND" {
		httpStatus = 404
	} else if code == "INTERNAL_ERROR" {
		httpStatus = 500
	}

	response := map[string]interface{}{
		"reqSeqId":   extractReqSeqId(msg),
		"rspSeqId":   generateRspSeqId(),
		"httpStatus": httpStatus,
		"data":       nil,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
			"details": err.Error(),
		},
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