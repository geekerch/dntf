package handlers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"notification/internal/application/channel/dtos"
	"notification/internal/application/cqrs"
	channelcqrs "notification/internal/application/cqrs/channel"
	"notification/pkg/logger"
)

// CQRSChannelNATSHandler handles NATS messages for channel operations using CQRS
type CQRSChannelNATSHandler struct {
	cqrsFacade *cqrs.CQRSFacade
	natsConn   *nats.Conn
}

// NewCQRSChannelNATSHandler creates a new CQRS channel NATS handler
func NewCQRSChannelNATSHandler(cqrsFacade *cqrs.CQRSFacade, natsConn *nats.Conn) *CQRSChannelNATSHandler {
	return &CQRSChannelNATSHandler{
		cqrsFacade: cqrsFacade,
		natsConn:   natsConn,
	}
}

// RegisterHandlers registers all NATS message handlers for channel operations using CQRS
func (h *CQRSChannelNATSHandler) RegisterHandlers() error {
	// Register create channel handler
	if _, err := h.natsConn.Subscribe("eco1j.infra.eventcenter.channel.create", h.handleCreateChannel); err != nil {
		return err
	}

	// Register get channel handler
	if _, err := h.natsConn.Subscribe("eco1j.infra.eventcenter.channel.get", h.handleGetChannel); err != nil {
		return err
	}

	// Register list channels handler
	if _, err := h.natsConn.Subscribe("eco1j.infra.eventcenter.channel.list", h.handleListChannels); err != nil {
		return err
	}

	// Register update channel handler
	if _, err := h.natsConn.Subscribe("eco1j.infra.eventcenter.channel.update", h.handleUpdateChannel); err != nil {
		return err
	}

	// Register delete channel handler
	if _, err := h.natsConn.Subscribe("eco1j.infra.eventcenter.channel.delete", h.handleDeleteChannel); err != nil {
		return err
	}

	logger.Info("CQRS Channel NATS handlers registered successfully")
	return nil
}

// handleCreateChannel handles create channel NATS messages using CQRS
func (h *CQRSChannelNATSHandler) handleCreateChannel(msg *nats.Msg) {
	ctx := context.Background()

	logger.Info("Received create channel NATS message",
		zap.String("subject", msg.Subject),
		zap.String("reply", msg.Reply))

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

	// Create command
	command := channelcqrs.NewCreateChannelCommand(&request)
	command.TraceID = natsReq.ReqSeqId

	// Execute command using CQRS
	result, err := h.cqrsFacade.Send(ctx, command)
	if err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "EXECUTION_ERROR", "Failed to create channel", err.Error())
		return
	}

	if !result.Success {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "EXECUTION_ERROR", "Failed to create channel", result.Error.Error())
		return
	}

	h.sendSuccessResponse(msg, natsReq.ReqSeqId, result.Data)
}

// handleGetChannel handles get channel NATS messages using CQRS
func (h *CQRSChannelNATSHandler) handleGetChannel(msg *nats.Msg) {
	ctx := context.Background()

	logger.Info("Received get channel NATS message",
		zap.String("subject", msg.Subject),
		zap.String("reply", msg.Reply))

	var natsReq NATSRequest
	if err := json.Unmarshal(msg.Data, &natsReq); err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to parse request", err.Error())
		return
	}

	// Extract channel ID from data
	channelID, ok := natsReq.Data.(string)
	if !ok {
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

	// Create query
	query := channelcqrs.NewGetChannelQuery(channelID)
	query.TraceID = natsReq.ReqSeqId

	// Execute query using CQRS
	result, err := h.cqrsFacade.Query(ctx, query)
	if err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "EXECUTION_ERROR", "Failed to get channel", err.Error())
		return
	}

	if !result.Success {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "EXECUTION_ERROR", "Failed to get channel", result.Error.Error())
		return
	}

	h.sendSuccessResponse(msg, natsReq.ReqSeqId, result.Data)
}

// handleListChannels handles list channels NATS messages using CQRS
func (h *CQRSChannelNATSHandler) handleListChannels(msg *nats.Msg) {
	ctx := context.Background()

	logger.Info("Received list channels NATS message",
		zap.String("subject", msg.Subject),
		zap.String("reply", msg.Reply))

	var natsReq NATSRequest
	if err := json.Unmarshal(msg.Data, &natsReq); err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to parse request", err.Error())
		return
	}

	// Create query
	query := channelcqrs.NewListChannelsQuery()
	query.TraceID = natsReq.ReqSeqId

	// Parse request data if provided
	if natsReq.Data != nil {
		if dataMap, ok := natsReq.Data.(map[string]interface{}); ok {
			if channelType, exists := dataMap["channelType"]; exists {
				if ct, ok := channelType.(string); ok {
					query.WithChannelType(ct)
				}
			}

			if tags, exists := dataMap["tags"]; exists {
				if tagSlice, ok := tags.([]interface{}); ok {
					stringTags := make([]string, len(tagSlice))
					for i, tag := range tagSlice {
						if tagStr, ok := tag.(string); ok {
							stringTags[i] = tagStr
						}
					}
					query.WithTags(stringTags)
				}
			}

			if enabled, exists := dataMap["enabled"]; exists {
				if enabledBool, ok := enabled.(bool); ok {
					query.WithEnabled(enabledBool)
				}
			}

			// Parse pagination
			if pagination, exists := dataMap["pagination"]; exists {
				if paginationMap, ok := pagination.(map[string]interface{}); ok {
					offset := 0
					limit := 20

					if offsetVal, exists := paginationMap["offset"]; exists {
						if offsetFloat, ok := offsetVal.(float64); ok {
							offset = int(offsetFloat)
						}
					}

					if limitVal, exists := paginationMap["limit"]; exists {
						if limitFloat, ok := limitVal.(float64); ok {
							limit = int(limitFloat)
						}
					}

					query.WithPagination(offset, limit)
				}
			}
		}
	}

	// Execute query using CQRS
	result, err := h.cqrsFacade.Query(ctx, query)
	if err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "EXECUTION_ERROR", "Failed to list channels", err.Error())
		return
	}

	if !result.Success {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "EXECUTION_ERROR", "Failed to list channels", result.Error.Error())
		return
	}

	h.sendSuccessResponse(msg, natsReq.ReqSeqId, result.Data)
}

// handleUpdateChannel handles update channel NATS messages using CQRS
func (h *CQRSChannelNATSHandler) handleUpdateChannel(msg *nats.Msg) {
	ctx := context.Background()

	logger.Info("Received update channel NATS message",
		zap.String("subject", msg.Subject),
		zap.String("reply", msg.Reply))

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

	if request.ChannelID == "" {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Channel ID is required", "")
		return
	}

	// Create command
	command := channelcqrs.NewUpdateChannelCommand(request.ChannelID, &request)
	command.TraceID = natsReq.ReqSeqId

	// Execute command using CQRS
	result, err := h.cqrsFacade.Send(ctx, command)
	if err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "EXECUTION_ERROR", "Failed to update channel", err.Error())
		return
	}

	if !result.Success {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "EXECUTION_ERROR", "Failed to update channel", result.Error.Error())
		return
	}

	h.sendSuccessResponse(msg, natsReq.ReqSeqId, result.Data)
}

// handleDeleteChannel handles delete channel NATS messages using CQRS
func (h *CQRSChannelNATSHandler) handleDeleteChannel(msg *nats.Msg) {
	ctx := context.Background()

	logger.Info("Received delete channel NATS message",
		zap.String("subject", msg.Subject),
		zap.String("reply", msg.Reply))

	var natsReq NATSRequest
	if err := json.Unmarshal(msg.Data, &natsReq); err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to parse request", err.Error())
		return
	}

	// Extract channel ID from data
	channelID, ok := natsReq.Data.(string)
	if !ok {
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

	// Create command
	command := channelcqrs.NewDeleteChannelCommand(channelID)
	command.TraceID = natsReq.ReqSeqId

	// Execute command using CQRS
	result, err := h.cqrsFacade.Send(ctx, command)
	if err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "EXECUTION_ERROR", "Failed to delete channel", err.Error())
		return
	}

	if !result.Success {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "EXECUTION_ERROR", "Failed to delete channel", result.Error.Error())
		return
	}

	h.sendSuccessResponse(msg, natsReq.ReqSeqId, result.Data)
}

// sendSuccessResponse sends a success response via NATS
func (h *CQRSChannelNATSHandler) sendSuccessResponse(msg *nats.Msg, reqSeqId string, data interface{}) {
	rspId, _ := uuid.NewRandom()
	response := NATSResponse{
		ReqSeqId:  reqSeqId,
		RspSeqId:  rspId.String(),
		Success:   true,
		Data:      data,
		Timestamp: time.Now().Unix(),
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
func (h *CQRSChannelNATSHandler) sendErrorResponse(msg *nats.Msg, reqSeqId, code, message, details string) {
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
		Timestamp: time.Now().Unix(),
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
