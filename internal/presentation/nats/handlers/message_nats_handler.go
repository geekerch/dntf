package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"notification/internal/application/message/dtos"
	"notification/internal/application/message/usecases"
	"notification/pkg/logger"
)

// MessageNATSHandler handles NATS messages for message operations
type MessageNATSHandler struct {
	sendUseCase *usecases.SendMessageUseCase
	getUseCase  *usecases.GetMessageUseCase
	listUseCase *usecases.ListMessagesUseCase
	natsConn    *nats.Conn
}

// NewMessageNATSHandler creates a new NATS handler for message operations
func NewMessageNATSHandler(
	sendUseCase *usecases.SendMessageUseCase,
	getUseCase *usecases.GetMessageUseCase,
	listUseCase *usecases.ListMessagesUseCase,
	natsConn *nats.Conn,
) *MessageNATSHandler {
	return &MessageNATSHandler{
		sendUseCase: sendUseCase,
		getUseCase:  getUseCase,
		listUseCase: listUseCase,
		natsConn:    natsConn,
	}
}

// RegisterHandlers registers all NATS message handlers for message operations
func (h *MessageNATSHandler) RegisterHandlers() error {
	if _, err := h.natsConn.Subscribe("eco1j.infra.eventcenter.message.send", h.handleSendMessage); err != nil {
		return fmt.Errorf("failed to subscribe to send message topic: %w", err)
	}
	if _, err := h.natsConn.Subscribe("eco1j.infra.eventcenter.message.get", h.handleGetMessage); err != nil {
		return fmt.Errorf("failed to subscribe to get message topic: %w", err)
	}
	if _, err := h.natsConn.Subscribe("eco1j.infra.eventcenter.message.list", h.handleListMessages); err != nil {
		return fmt.Errorf("failed to subscribe to list messages topic: %w", err)
	}
	logger.Info("Message NATS handlers registered successfully")
	return nil
}

// handleSendMessage handles send message NATS messages
func (h *MessageNATSHandler) handleSendMessage(msg *nats.Msg) {
	ctx := context.Background()
	logger.Info("Received send message NATS message",
		zap.String("subject", msg.Subject),
		zap.String("reply", msg.Reply),
	)

	var natsReq NATSRequest
	if err := json.Unmarshal(msg.Data, &natsReq); err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to parse request", err.Error())
		return
	}

	dataBytes, err := json.Marshal(natsReq.Data)
	if err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to marshal request data", err.Error())
		return
	}

	var request dtos.SendMessageRequest
	if err := json.Unmarshal(dataBytes, &request); err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to parse send message request", err.Error())
		return
	}

	// Forward to legacy system first
	responses, err := h.sendUseCase.Forward(ctx, &request)
	if err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "EXECUTION_ERROR", "Failed to forward message to legacy system", err.Error())
		return
	}

	h.sendSuccessResponse(msg, natsReq.ReqSeqId, responses)
}

// handleGetMessage handles get message NATS messages
func (h *MessageNATSHandler) handleGetMessage(msg *nats.Msg) {
	ctx := context.Background()
	logger.Info("Received get message NATS message",
		zap.String("subject", msg.Subject),
		zap.String("reply", msg.Reply),
	)

	var natsReq NATSRequest
	if err := json.Unmarshal(msg.Data, &natsReq); err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to parse request", err.Error())
		return
	}

	messageID, ok := natsReq.Data.(string)
	if !ok {
		if dataMap, ok := natsReq.Data.(map[string]interface{}); ok {
			if id, exists := dataMap["messageId"]; exists {
				messageID, _ = id.(string)
			}
		}
	}

	if messageID == "" {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Message ID is required", "")
		return
	}

	response, err := h.getUseCase.Execute(ctx, messageID)
	if err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "EXECUTION_ERROR", "Failed to get message", err.Error())
		return
	}

	h.sendSuccessResponse(msg, natsReq.ReqSeqId, response)
}

// handleListMessages handles list messages NATS messages
func (h *MessageNATSHandler) handleListMessages(msg *nats.Msg) {
	ctx := context.Background()
	logger.Info("Received list messages NATS message",
		zap.String("subject", msg.Subject),
		zap.String("reply", msg.Reply),
	)

	var natsReq NATSRequest
	if err := json.Unmarshal(msg.Data, &natsReq); err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to parse request", err.Error())
		return
	}

	var request dtos.ListMessagesRequest
	if natsReq.Data != nil {
		dataBytes, err := json.Marshal(natsReq.Data)
		if err != nil {
			h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to marshal request data", err.Error())
			return
		}

		if err := json.Unmarshal(dataBytes, &request); err != nil {
			h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to parse list messages request", err.Error())
			return
		}
	}

	response, err := h.listUseCase.Execute(ctx, &request)
	if err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "EXECUTION_ERROR", "Failed to list messages", err.Error())
		return
	}

	h.sendSuccessResponse(msg, natsReq.ReqSeqId, response)
}

// sendSuccessResponse sends a success response via NATS
func (h *MessageNATSHandler) sendSuccessResponse(msg *nats.Msg, reqSeqId string, data interface{}) {
	rspId, _ := uuid.NewRandom()
	response := NATSResponse{
		ReqSeqId:  reqSeqId,
		RspSeqId:  rspId.String(),
		Success:   true,
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		logger.Error("Failed to marshal success response", zap.Error(err))
		return
	}

	if err := msg.Respond(responseBytes); err != nil {
		logger.Error("Failed to send success response", zap.Error(err))
	}
}

// sendErrorResponse sends an error response via NATS
func (h *MessageNATSHandler) sendErrorResponse(msg *nats.Msg, reqSeqId, code, message, details string) {
	rspId, _ := uuid.NewRandom()
	response := NATSResponse{
		ReqSeqId: reqSeqId,
		RspSeqId: rspId.String(),
		Success:  false,
		Error: &NATSError{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now().UnixMilli(),
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		logger.Error("Failed to marshal error response", zap.Error(err))
		return
	}

	if err := msg.Respond(responseBytes); err != nil {
		logger.Error("Failed to send error response", zap.Error(err))
	}
}
