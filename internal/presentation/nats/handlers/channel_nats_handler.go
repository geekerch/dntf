package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"notification/internal/application/channel/dtos"
	"notification/internal/application/channel/usecases"
	"notification/pkg/logger"
)

// ChannelNATSHandler handles NATS messages for channel operations
type ChannelNATSHandler struct {
	createUseCase *usecases.CreateChannelUseCase
	getUseCase    *usecases.GetChannelUseCase
	listUseCase   *usecases.ListChannelsUseCase
	updateUseCase *usecases.UpdateChannelUseCase
	deleteUseCase *usecases.DeleteChannelUseCase
	natsConn      *nats.Conn
}

// NATSRequest represents a generic NATS request message
type NATSRequest struct {
	ReqSeqId  string      `json:"reqSeqId"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// NATSResponse represents a generic NATS response message
type NATSResponse struct {
	ReqSeqId  string      `json:"reqSeqId"`
	RspSeqId  string      `json:"rspSeqId"`
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *NATSError  `json:"error,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// NATSError represents error information in NATS response
type NATSError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// NewChannelNATSHandler creates a new NATS handler for channel operations
func NewChannelNATSHandler(
	createUseCase *usecases.CreateChannelUseCase,
	getUseCase *usecases.GetChannelUseCase,
	listUseCase *usecases.ListChannelsUseCase,
	updateUseCase *usecases.UpdateChannelUseCase,
	deleteUseCase *usecases.DeleteChannelUseCase,
	natsConn *nats.Conn,
) *ChannelNATSHandler {
	return &ChannelNATSHandler{
		createUseCase: createUseCase,
		getUseCase:    getUseCase,
		listUseCase:   listUseCase,
		updateUseCase: updateUseCase,
		deleteUseCase: deleteUseCase,
		natsConn:      natsConn,
	}
}

// RegisterHandlers registers all NATS message handlers for channel operations
func (h *ChannelNATSHandler) RegisterHandlers() error {
	// Register create channel handler
	if _, err := h.natsConn.Subscribe("eco1j.infra.eventcenter.channel.create", h.handleCreateChannel); err != nil {
		return fmt.Errorf("failed to subscribe to create channel topic: %w", err)
	}

	// Register get channel handler
	if _, err := h.natsConn.Subscribe("eco1j.infra.eventcenter.channel.get", h.handleGetChannel); err != nil {
		return fmt.Errorf("failed to subscribe to get channel topic: %w", err)
	}

	// Register list channels handler
	if _, err := h.natsConn.Subscribe("eco1j.infra.eventcenter.channel.list", h.handleListChannels); err != nil {
		return fmt.Errorf("failed to subscribe to list channels topic: %w", err)
	}

	// Register update channel handler
	if _, err := h.natsConn.Subscribe("eco1j.infra.eventcenter.channel.update", h.handleUpdateChannel); err != nil {
		return fmt.Errorf("failed to subscribe to update channel topic: %w", err)
	}

	// Register delete channel handler
	if _, err := h.natsConn.Subscribe("eco1j.infra.eventcenter.channel.delete", h.handleDeleteChannel); err != nil {
		return fmt.Errorf("failed to subscribe to delete channel topic: %w", err)
	}

	logger.Info("Channel NATS handlers registered successfully")
	return nil
}

// handleCreateChannel handles create channel NATS messages
func (h *ChannelNATSHandler) handleCreateChannel(msg *nats.Msg) {
	ctx := context.Background()

	logger.Info("Received create channel NATS message",
		zap.String("subject", msg.Subject),
		zap.String("reply", msg.Reply),
	)

	var natsReq NATSRequest
	if err := json.Unmarshal(msg.Data, &natsReq); err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to parse request", err.Error())
		return
	}

	// Convert data to CreateChannelRequest
	dataBytes, err := json.Marshal(natsReq.Data)
	if err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to marshal request data", err.Error())
		return
	}

	var request dtos.CreateChannelRequest
	if err := json.Unmarshal(dataBytes, &request); err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to parse create channel request", err.Error())
		return
	}

	// Execute use case
	response, err := h.createUseCase.Execute(ctx, &request)
	if err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "EXECUTION_ERROR", "Failed to create channel", err.Error())
		return
	}

	h.sendSuccessResponse(msg, natsReq.ReqSeqId, response)
}

// handleGetChannel handles get channel NATS messages
func (h *ChannelNATSHandler) handleGetChannel(msg *nats.Msg) {
	ctx := context.Background()

	logger.Info("Received get channel NATS message",
		zap.String("subject", msg.Subject),
		zap.String("reply", msg.Reply),
	)

	var natsReq NATSRequest
	if err := json.Unmarshal(msg.Data, &natsReq); err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to parse request", err.Error())
		return
	}

	// Extract channel ID from data
	channelID, ok := natsReq.Data.(string)
	if !ok {
		// Try to extract from a map structure
		if dataMap, ok := natsReq.Data.(map[string]interface{}); ok {
			if id, exists := dataMap["channelId"]; exists {
				channelID, _ = id.(string)
			}
		}
	}

	if channelID == "" {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Channel ID is required", "")
		return
	}

	// Execute use case
	response, err := h.getUseCase.Execute(ctx, channelID)
	if err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "EXECUTION_ERROR", "Failed to get channel", err.Error())
		return
	}

	h.sendSuccessResponse(msg, natsReq.ReqSeqId, response)
}

// handleListChannels handles list channels NATS messages
func (h *ChannelNATSHandler) handleListChannels(msg *nats.Msg) {
	ctx := context.Background()

	logger.Info("Received list channels NATS message",
		zap.String("subject", msg.Subject),
		zap.String("reply", msg.Reply),
	)

	var natsReq NATSRequest
	if err := json.Unmarshal(msg.Data, &natsReq); err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to parse request", err.Error())
		return
	}

	// Convert data to ListChannelsRequest
	var request dtos.ListChannelsRequest
	if natsReq.Data != nil {
		dataBytes, err := json.Marshal(natsReq.Data)
		if err != nil {
			h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to marshal request data", err.Error())
			return
		}

		if err := json.Unmarshal(dataBytes, &request); err != nil {
			h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to parse list channels request", err.Error())
			return
		}
	}

	// Set default values
	if request.MaxResultCount <= 0 {
		request.MaxResultCount = 20
	}

	// Execute use case
	response, err := h.listUseCase.Execute(ctx, &request)
	if err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "EXECUTION_ERROR", "Failed to list channels", err.Error())
		return
	}

	h.sendSuccessResponse(msg, natsReq.ReqSeqId, response)
}

// handleUpdateChannel handles update channel NATS messages
func (h *ChannelNATSHandler) handleUpdateChannel(msg *nats.Msg) {
	ctx := context.Background()

	logger.Info("Received update channel NATS message",
		zap.String("subject", msg.Subject),
		zap.String("reply", msg.Reply),
	)

	var natsReq NATSRequest
	if err := json.Unmarshal(msg.Data, &natsReq); err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to parse request", err.Error())
		return
	}

	// Convert data to UpdateChannelRequest
	dataBytes, err := json.Marshal(natsReq.Data)
	if err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to marshal request data", err.Error())
		return
	}

	var request dtos.UpdateChannelRequest
	if err := json.Unmarshal(dataBytes, &request); err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to parse update channel request", err.Error())
		return
	}

	// Execute use case
	response, err := h.updateUseCase.Execute(ctx, request.ChannelID, &request)
	if err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "EXECUTION_ERROR", "Failed to update channel", err.Error())
		return
	}

	h.sendSuccessResponse(msg, natsReq.ReqSeqId, response)
}

// handleDeleteChannel handles delete channel NATS messages
func (h *ChannelNATSHandler) handleDeleteChannel(msg *nats.Msg) {
	ctx := context.Background()

	logger.Info("Received delete channel NATS message",
		zap.String("subject", msg.Subject),
		zap.String("reply", msg.Reply),
	)

	var natsReq NATSRequest
	if err := json.Unmarshal(msg.Data, &natsReq); err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to parse request", err.Error())
		return
	}

	// Extract channel ID from data
	channelID, ok := natsReq.Data.(string)
	if !ok {
		// Try to extract from a map structure
		if dataMap, ok := natsReq.Data.(map[string]interface{}); ok {
			if id, exists := dataMap["channelId"]; exists {
				channelID, _ = id.(string)
			}
		}
	}

	if channelID == "" {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Channel ID is required", "")
		return
	}

	// Execute use case
	response, err := h.deleteUseCase.Execute(ctx, channelID)
	if err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "EXECUTION_ERROR", "Failed to delete channel", err.Error())
		return
	}

	h.sendSuccessResponse(msg, natsReq.ReqSeqId, response)
}

// sendSuccessResponse sends a success response via NATS
func (h *ChannelNATSHandler) sendSuccessResponse(msg *nats.Msg, reqSeqId string, data interface{}) {
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
func (h *ChannelNATSHandler) sendErrorResponse(msg *nats.Msg, requestID, code, message, details string) {
	rspId, _ := uuid.NewRandom()
	response := NATSResponse{
		ReqSeqId: requestID,
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
