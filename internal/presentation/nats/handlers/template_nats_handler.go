package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"notification/internal/application/template/dtos"
	"notification/internal/application/template/usecases"
	"notification/pkg/logger"
)

// TemplateNATSHandler handles NATS messages for template operations
type TemplateNATSHandler struct {
	createUseCase *usecases.CreateTemplateUseCase
	getUseCase    *usecases.GetTemplateUseCase
	listUseCase   *usecases.ListTemplatesUseCase
	updateUseCase *usecases.UpdateTemplateUseCase
	deleteUseCase *usecases.DeleteTemplateUseCase
	natsConn      *nats.Conn
}

// NewTemplateNATSHandler creates a new NATS handler for template operations
func NewTemplateNATSHandler(
	createUseCase *usecases.CreateTemplateUseCase,
	getUseCase *usecases.GetTemplateUseCase,
	listUseCase *usecases.ListTemplatesUseCase,
	updateUseCase *usecases.UpdateTemplateUseCase,
	deleteUseCase *usecases.DeleteTemplateUseCase,
	natsConn *nats.Conn,
) *TemplateNATSHandler {
	return &TemplateNATSHandler{
		createUseCase: createUseCase,
		getUseCase:    getUseCase,
		listUseCase:   listUseCase,
		updateUseCase: updateUseCase,
		deleteUseCase: deleteUseCase,
		natsConn:      natsConn,
	}
}

// RegisterHandlers registers all NATS message handlers for template operations
func (h *TemplateNATSHandler) RegisterHandlers() error {
	if _, err := h.natsConn.Subscribe("eco1j.infra.eventcenter.template.create", h.handleCreateTemplate); err != nil {
		return fmt.Errorf("failed to subscribe to create template topic: %w", err)
	}
	if _, err := h.natsConn.Subscribe("eco1j.infra.eventcenter.template.get", h.handleGetTemplate); err != nil {
		return fmt.Errorf("failed to subscribe to get template topic: %w", err)
	}
	if _, err := h.natsConn.Subscribe("eco1j.infra.eventcenter.template.list", h.handleListTemplates); err != nil {
		return fmt.Errorf("failed to subscribe to list templates topic: %w", err)
	}
	if _, err := h.natsConn.Subscribe("eco1j.infra.eventcenter.template.update", h.handleUpdateTemplate); err != nil {
		return fmt.Errorf("failed to subscribe to update template topic: %w", err)
	}
	if _, err := h.natsConn.Subscribe("eco1j.infra.eventcenter.template.delete", h.handleDeleteTemplate); err != nil {
		return fmt.Errorf("failed to subscribe to delete template topic: %w", err)
	}
	logger.Info("Template NATS handlers registered successfully")
	return nil
}

// handleCreateTemplate handles create template NATS messages
func (h *TemplateNATSHandler) handleCreateTemplate(msg *nats.Msg) {
	ctx := context.Background()
	logger.Info("Received create template NATS message", zap.String("subject", msg.Subject), zap.String("reply", msg.Reply))

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

	var request dtos.CreateTemplateRequest
	if err := json.Unmarshal(dataBytes, &request); err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to parse create template request", err.Error())
		return
	}

	response, err := h.createUseCase.Execute(ctx, &request)
	if err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "EXECUTION_ERROR", "Failed to create template", err.Error())
		return
	}

	h.sendSuccessResponse(msg, natsReq.ReqSeqId, response)
}

// handleGetTemplate handles get template NATS messages
func (h *TemplateNATSHandler) handleGetTemplate(msg *nats.Msg) {
	ctx := context.Background()
	logger.Info("Received get template NATS message", zap.String("subject", msg.Subject), zap.String("reply", msg.Reply))

	var natsReq NATSRequest
	if err := json.Unmarshal(msg.Data, &natsReq); err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to parse request", err.Error())
		return
	}

	templateID, ok := natsReq.Data.(string)
	if !ok {
		if dataMap, ok := natsReq.Data.(map[string]interface{}); ok {
			if id, exists := dataMap["templateId"]; exists {
				templateID, _ = id.(string)
			}
		}
	}

	if templateID == "" {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Template ID is required", "")
		return
	}

	response, err := h.getUseCase.Execute(ctx, templateID)
	if err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "EXECUTION_ERROR", "Failed to get template", err.Error())
		return
	}

	h.sendSuccessResponse(msg, natsReq.ReqSeqId, response)
}

// handleListTemplates handles list templates NATS messages
func (h *TemplateNATSHandler) handleListTemplates(msg *nats.Msg) {
	ctx := context.Background()
	logger.Info("Received list templates NATS message", zap.String("subject", msg.Subject), zap.String("reply", msg.Reply))

	var natsReq NATSRequest
	if err := json.Unmarshal(msg.Data, &natsReq); err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to parse request", err.Error())
		return
	}

	var request dtos.ListTemplatesRequest
	if natsReq.Data != nil {
		dataBytes, err := json.Marshal(natsReq.Data)
		if err != nil {
			h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to marshal request data", err.Error())
			return
		}

		if err := json.Unmarshal(dataBytes, &request); err != nil {
			h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to parse list templates request", err.Error())
			return
		}
	}

	response, err := h.listUseCase.Execute(ctx, &request)
	if err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "EXECUTION_ERROR", "Failed to list templates", err.Error())
		return
	}

	h.sendSuccessResponse(msg, natsReq.ReqSeqId, response)
}

// handleUpdateTemplate handles update template NATS messages
func (h *TemplateNATSHandler) handleUpdateTemplate(msg *nats.Msg) {
	ctx := context.Background()
	logger.Info("Received update template NATS message", zap.String("subject", msg.Subject), zap.String("reply", msg.Reply))

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

	var payload map[string]interface{}
	if err := json.Unmarshal(dataBytes, &payload); err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to parse update template payload", err.Error())
		return
	}

	templateID, ok := payload["templateId"].(string)
	if !ok || templateID == "" {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "templateId is required in payload", "")
		return
	}
	delete(payload, "templateId")

	updateDtoBytes, err := json.Marshal(payload)
	if err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to marshal update DTO from payload", err.Error())
		return
	}
	var updateDto dtos.UpdateTemplateRequest
	if err := json.Unmarshal(updateDtoBytes, &updateDto); err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to unmarshal update DTO", err.Error())
		return
	}

	response, err := h.updateUseCase.Execute(ctx, templateID, &updateDto)
	if err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "EXECUTION_ERROR", "Failed to update template", err.Error())
		return
	}

	h.sendSuccessResponse(msg, natsReq.ReqSeqId, response)
}

// handleDeleteTemplate handles delete template NATS messages
func (h *TemplateNATSHandler) handleDeleteTemplate(msg *nats.Msg) {
	ctx := context.Background()
	logger.Info("Received delete template NATS message", zap.String("subject", msg.Subject), zap.String("reply", msg.Reply))

	var natsReq NATSRequest
	if err := json.Unmarshal(msg.Data, &natsReq); err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Failed to parse request", err.Error())
		return
	}

	templateID, ok := natsReq.Data.(string)
	if !ok {
		if dataMap, ok := natsReq.Data.(map[string]interface{}); ok {
			if id, exists := dataMap["templateId"]; exists {
				templateID, _ = id.(string)
			}
		}
	}

	if templateID == "" {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "INVALID_REQUEST", "Template ID is required", "")
		return
	}

	if err := h.deleteUseCase.Execute(ctx, templateID); err != nil {
		h.sendErrorResponse(msg, natsReq.ReqSeqId, "EXECUTION_ERROR", "Failed to delete template", err.Error())
		return
	}

	h.sendSuccessResponse(msg, natsReq.ReqSeqId, map[string]interface{}{"deleted": true})
}

// sendSuccessResponse sends a success response via NATS
func (h *TemplateNATSHandler) sendSuccessResponse(msg *nats.Msg, reqSeqId string, data interface{}) {
	rspSeqId, _ := uuid.NewRandom()
	response := NATSResponse{
		ReqSeqId:  reqSeqId,
		RspSeqId:  rspSeqId.String(),
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
func (h *TemplateNATSHandler) sendErrorResponse(msg *nats.Msg, reqSeqId, code, message, details string) {
	rspSeqId, _ := uuid.NewRandom()
	response := NATSResponse{
		ReqSeqId: reqSeqId,
		RspSeqId: rspSeqId.String(),
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
