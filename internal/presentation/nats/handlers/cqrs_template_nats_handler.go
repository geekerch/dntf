package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"notification/internal/application/cqrs"
	templatecqrs "notification/internal/application/cqrs/template"
	"notification/internal/application/template/dtos"
	"notification/pkg/logger"
)

// CQRSTemplateNATSHandler handles CQRS NATS messages for templates
type CQRSTemplateNATSHandler struct {
	cqrsFacade *cqrs.CQRSFacade
	logger     logger.Logger
}

// NewCQRSTemplateNATSHandler creates a new CQRS template NATS handler
func NewCQRSTemplateNATSHandler(
	cqrsFacade *cqrs.CQRSFacade,
	logger logger.Logger,
) *CQRSTemplateNATSHandler {
	return &CQRSTemplateNATSHandler{
		cqrsFacade: cqrsFacade,
		logger:     logger,
	}
}

// HandleCreateTemplate handles template creation via CQRS NATS
func (h *CQRSTemplateNATSHandler) HandleCreateTemplate(msg *nats.Msg) {
	var req dtos.CreateTemplateRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		h.logger.Error("Failed to unmarshal create template request", zap.Error(err))
		h.respondWithError(msg, "INVALID_REQUEST", "Invalid request format", err)
		return
	}

	// Create command
	cmd := templatecqrs.NewCreateTemplateCommand(&req)

	// Execute command via CQRS
	result, err := h.cqrsFacade.Send(context.Background(), cmd)
	if err != nil {
		h.logger.Error("Failed to create template via CQRS", zap.Error(err))
		h.respondWithError(msg, "CREATE_FAILED", "Failed to create template", err)
		return
	}

	// Type assert the result
	response, ok := result.Data.(*dtos.TemplateResponse)
	if !ok {
		h.logger.Error("Invalid response type from CQRS create template")
		h.respondWithError(msg, "INTERNAL_ERROR", "Invalid response type", fmt.Errorf("invalid response type"))
		return
	}

	h.respondWithSuccess(msg, response, "Template created successfully via CQRS")
}

// HandleGetTemplate handles getting a template via CQRS NATS
func (h *CQRSTemplateNATSHandler) HandleGetTemplate(msg *nats.Msg) {
	var req struct {
		TemplateID string `json:"templateId"`
	}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		h.logger.Error("Failed to unmarshal get template request", zap.Error(err))
		h.respondWithError(msg, "INVALID_REQUEST", "Invalid request format", err)
		return
	}

	// Create query
	query := templatecqrs.NewGetTemplateQuery(req.TemplateID)

	// Execute query via CQRS
	result, err := h.cqrsFacade.Query(context.Background(), query)
	if err != nil {
		h.logger.Error("Failed to get template via CQRS", zap.Error(err), zap.String("templateId", req.TemplateID))
		h.respondWithError(msg, "NOT_FOUND", "Template not found", err)
		return
	}

	// Type assert the result
	response, ok := result.Data.(*dtos.TemplateResponse)
	if !ok {
		h.logger.Error("Invalid response type from CQRS get template")
		h.respondWithError(msg, "INTERNAL_ERROR", "Invalid response type", fmt.Errorf("invalid response type"))
		return
	}

	h.respondWithSuccess(msg, response, "")
}

// HandleListTemplates handles listing templates via CQRS NATS
func (h *CQRSTemplateNATSHandler) HandleListTemplates(msg *nats.Msg) {
	var req struct {
		ChannelType      string   `json:"channelType,omitempty"`
		Tags             []string `json:"tags,omitempty"`
		SkipCount        int      `json:"skipCount,omitempty"`
		MaxResultCount   int      `json:"maxResultCount,omitempty"`
	}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		h.logger.Error("Failed to unmarshal list templates request", zap.Error(err))
		h.respondWithError(msg, "INVALID_REQUEST", "Invalid request format", err)
		return
	}

	// Create query
	query := templatecqrs.NewListTemplatesQuery()
	
	if req.ChannelType != "" {
		query.WithChannelType(req.ChannelType)
	}
	if len(req.Tags) > 0 {
		query.WithTags(req.Tags)
	}
	if req.MaxResultCount > 0 {
		query.WithPagination(req.SkipCount, req.MaxResultCount)
	}

	// Execute query via CQRS
	result, err := h.cqrsFacade.Query(context.Background(), query)
	if err != nil {
		h.logger.Error("Failed to list templates via CQRS", zap.Error(err))
		h.respondWithError(msg, "LIST_FAILED", "Failed to list templates", err)
		return
	}

	// Type assert the result
	response, ok := result.Data.(*dtos.ListTemplatesResponse)
	if !ok {
		h.logger.Error("Invalid response type from CQRS list templates")
		h.respondWithError(msg, "INTERNAL_ERROR", "Invalid response type", fmt.Errorf("invalid response type"))
		return
	}

	h.respondWithSuccess(msg, response, "")
}

// HandleUpdateTemplate handles template update via CQRS NATS
func (h *CQRSTemplateNATSHandler) HandleUpdateTemplate(msg *nats.Msg) {
	var req struct {
		TemplateID string                       `json:"templateId"`
		Data       dtos.UpdateTemplateRequest   `json:"data"`
	}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		h.logger.Error("Failed to unmarshal update template request", zap.Error(err))
		h.respondWithError(msg, "INVALID_REQUEST", "Invalid request format", err)
		return
	}

	// Create command
	cmd := templatecqrs.NewUpdateTemplateCommand(req.TemplateID, &req.Data)

	// Execute command via CQRS
	result, err := h.cqrsFacade.Send(context.Background(), cmd)
	if err != nil {
		h.logger.Error("Failed to update template via CQRS", zap.Error(err), zap.String("templateId", req.TemplateID))
		h.respondWithError(msg, "UPDATE_FAILED", "Failed to update template", err)
		return
	}

	// Type assert the result
	response, ok := result.Data.(*dtos.TemplateResponse)
	if !ok {
		h.logger.Error("Invalid response type from CQRS update template")
		h.respondWithError(msg, "INTERNAL_ERROR", "Invalid response type", fmt.Errorf("invalid response type"))
		return
	}

	h.respondWithSuccess(msg, response, "Template updated successfully via CQRS")
}

// HandleDeleteTemplate handles template deletion via CQRS NATS
func (h *CQRSTemplateNATSHandler) HandleDeleteTemplate(msg *nats.Msg) {
	var req struct {
		TemplateID string `json:"templateId"`
	}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		h.logger.Error("Failed to unmarshal delete template request", zap.Error(err))
		h.respondWithError(msg, "INVALID_REQUEST", "Invalid request format", err)
		return
	}

	// Create command
	cmd := templatecqrs.NewDeleteTemplateCommand(req.TemplateID)

	// Execute command via CQRS
	_, err := h.cqrsFacade.Send(context.Background(), cmd)
	if err != nil {
		h.logger.Error("Failed to delete template via CQRS", zap.Error(err), zap.String("templateId", req.TemplateID))
		h.respondWithError(msg, "DELETE_FAILED", "Failed to delete template", err)
		return
	}

	// For delete operations, we expect a success result without specific data
	deleteResponse := map[string]interface{}{
		"templateId": req.TemplateID,
		"deleted":    true,
		"deletedAt":  getCurrentTimestamp(),
	}

	h.respondWithSuccess(msg, deleteResponse, "Template deleted successfully via CQRS")
}

// RegisterHandlers registers all CQRS template NATS handlers
func (h *CQRSTemplateNATSHandler) RegisterHandlers(nc *nats.Conn, subjectPrefix string) error {
	// Register command handlers
	if _, err := nc.Subscribe(fmt.Sprintf("%s.template.create", subjectPrefix), h.HandleCreateTemplate); err != nil {
		return fmt.Errorf("failed to subscribe to template.create: %w", err)
	}

	if _, err := nc.Subscribe(fmt.Sprintf("%s.template.update", subjectPrefix), h.HandleUpdateTemplate); err != nil {
		return fmt.Errorf("failed to subscribe to template.update: %w", err)
	}

	if _, err := nc.Subscribe(fmt.Sprintf("%s.template.delete", subjectPrefix), h.HandleDeleteTemplate); err != nil {
		return fmt.Errorf("failed to subscribe to template.delete: %w", err)
	}

	// Register query handlers
	if _, err := nc.Subscribe(fmt.Sprintf("%s.template.get", subjectPrefix), h.HandleGetTemplate); err != nil {
		return fmt.Errorf("failed to subscribe to template.get: %w", err)
	}

	if _, err := nc.Subscribe(fmt.Sprintf("%s.template.list", subjectPrefix), h.HandleListTemplates); err != nil {
		return fmt.Errorf("failed to subscribe to template.list: %w", err)
	}

	h.logger.Info("CQRS Template NATS handlers registered successfully")
	return nil
}

// respondWithSuccess sends a success response via NATS with CQRS format
func (h *CQRSTemplateNATSHandler) respondWithSuccess(msg *nats.Msg, data interface{}, message string) {
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
func (h *CQRSTemplateNATSHandler) respondWithError(msg *nats.Msg, code, message string, err error) {
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

// Helper functions for sequence ID handling
func extractReqSeqId(msg *nats.Msg) string {
	// Try to extract reqSeqId from message headers or generate one
	if msg.Header != nil {
		if reqSeqId := msg.Header.Get("reqSeqId"); reqSeqId != "" {
			return reqSeqId
		}
	}
	return generateRspSeqId() // Fallback to generated ID
}

func generateRspSeqId() string {
	// Generate a unique response sequence ID
	// This is a simple implementation - in production you might want to use UUID
	return fmt.Sprintf("rsp_%d", getCurrentTimestamp())
}

func getCurrentTimestamp() int64 {
	return 1701421200000 // Placeholder - should use actual timestamp
}