package handlers

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"notification/internal/application/template/dtos"
	"notification/internal/application/template/usecases"
	"notification/pkg/logger"
)

// TemplateNATSHandler handles NATS messages for templates.
type TemplateNATSHandler struct {
	createTemplateUC *usecases.CreateTemplateUseCase
	getTemplateUC    *usecases.GetTemplateUseCase
	listTemplatesUC  *usecases.ListTemplatesUseCase
	updateTemplateUC *usecases.UpdateTemplateUseCase
	deleteTemplateUC *usecases.DeleteTemplateUseCase
	logger           logger.Logger
}

// NewTemplateNATSHandler creates a new TemplateNATSHandler.
func NewTemplateNATSHandler(
	createTemplateUC *usecases.CreateTemplateUseCase,
	getTemplateUC *usecases.GetTemplateUseCase,
	listTemplatesUC *usecases.ListTemplatesUseCase,
	updateTemplateUC *usecases.UpdateTemplateUseCase,
	deleteTemplateUC *usecases.DeleteTemplateUseCase,
	logger logger.Logger,
) *TemplateNATSHandler {
	return &TemplateNATSHandler{
		createTemplateUC: createTemplateUC,
		getTemplateUC:    getTemplateUC,
		listTemplatesUC:  listTemplatesUC,
		updateTemplateUC: updateTemplateUC,
		deleteTemplateUC: deleteTemplateUC,
		logger:           logger,
	}
}

// HandleCreateTemplate handles template creation via NATS.
func (h *TemplateNATSHandler) HandleCreateTemplate(msg *nats.Msg) {
	var req dtos.CreateTemplateRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		h.logger.Error("Failed to unmarshal create template request", zap.Error(err))
		h.respondWithError(msg, "Invalid request format", err)
		return
	}

	response, err := h.createTemplateUC.Execute(context.Background(), &req)
	if err != nil {
		h.logger.Error("Failed to create template", zap.Error(err))
		h.respondWithError(msg, "Failed to create template", err)
		return
	}

	h.respondWithSuccess(msg, response, "Template created successfully")
}

// HandleGetTemplate handles getting a template via NATS.
func (h *TemplateNATSHandler) HandleGetTemplate(msg *nats.Msg) {
	var req struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		h.logger.Error("Failed to unmarshal get template request", zap.Error(err))
		h.respondWithError(msg, "Invalid request format", err)
		return
	}

	response, err := h.getTemplateUC.Execute(context.Background(), req.ID)
	if err != nil {
		h.logger.Error("Failed to get template", zap.Error(err), zap.String("id", req.ID))
		h.respondWithError(msg, "Template not found", err)
		return
	}

	h.respondWithSuccess(msg, response, "")
}

// HandleListTemplates handles listing templates via NATS.
func (h *TemplateNATSHandler) HandleListTemplates(msg *nats.Msg) {
	var req dtos.ListTemplatesRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		h.logger.Error("Failed to unmarshal list templates request", zap.Error(err))
		h.respondWithError(msg, "Invalid request format", err)
		return
	}

	response, err := h.listTemplatesUC.Execute(context.Background(), &req)
	if err != nil {
		h.logger.Error("Failed to list templates", zap.Error(err))
		h.respondWithError(msg, "Failed to list templates", err)
		return
	}

	h.respondWithSuccess(msg, response, "")
}

// HandleUpdateTemplate handles template update via NATS.
func (h *TemplateNATSHandler) HandleUpdateTemplate(msg *nats.Msg) {
	var req struct {
		ID   string                        `json:"id"`
		Data dtos.UpdateTemplateRequest    `json:"data"`
	}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		h.logger.Error("Failed to unmarshal update template request", zap.Error(err))
		h.respondWithError(msg, "Invalid request format", err)
		return
	}

	response, err := h.updateTemplateUC.Execute(context.Background(), req.ID, &req.Data)
	if err != nil {
		h.logger.Error("Failed to update template", zap.Error(err), zap.String("id", req.ID))
		h.respondWithError(msg, "Failed to update template", err)
		return
	}

	h.respondWithSuccess(msg, response, "Template updated successfully")
}

// HandleDeleteTemplate handles template deletion via NATS.
func (h *TemplateNATSHandler) HandleDeleteTemplate(msg *nats.Msg) {
	var req struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		h.logger.Error("Failed to unmarshal delete template request", zap.Error(err))
		h.respondWithError(msg, "Invalid request format", err)
		return
	}

	err := h.deleteTemplateUC.Execute(context.Background(), req.ID)
	if err != nil {
		h.logger.Error("Failed to delete template", zap.Error(err), zap.String("id", req.ID))
		h.respondWithError(msg, "Failed to delete template", err)
		return
	}

	h.respondWithSuccess(msg, nil, "Template deleted successfully")
}

// respondWithSuccess sends a success response via NATS.
func (h *TemplateNATSHandler) respondWithSuccess(msg *nats.Msg, data interface{}, message string) {
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
func (h *TemplateNATSHandler) respondWithError(msg *nats.Msg, message string, err error) {
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