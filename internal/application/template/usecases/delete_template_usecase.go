package usecases

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"notification/internal/domain/channel"
	"notification/internal/domain/shared"
	"notification/internal/domain/template"
	"notification/pkg/config"
)

// DeleteTemplateUseCase handles deleting templates.
type DeleteTemplateUseCase struct {
	templateRepo template.TemplateRepository
	channelRepo  channel.ChannelRepository
	config       *config.Config
}

// NewDeleteTemplateUseCase creates a new DeleteTemplateUseCase.
func NewDeleteTemplateUseCase(
	templateRepo template.TemplateRepository,
	channelRepo channel.ChannelRepository,
	config *config.Config,
) *DeleteTemplateUseCase {
	return &DeleteTemplateUseCase{
		templateRepo: templateRepo,
		channelRepo:  channelRepo,
		config:       config,
	}
}

// Execute deletes a template.
func (uc *DeleteTemplateUseCase) Execute(ctx context.Context, id string) error {
	// Validate input
	if id == "" {
		return fmt.Errorf("template ID cannot be empty")
	}

	// Create template ID
	templateID, err := template.NewTemplateIDFromString(id)
	if err != nil {
		return fmt.Errorf("invalid template ID: %w", err)
	}

	// Get template entity before deletion (needed for legacy channel updates)
	templateEntity, err := uc.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		return fmt.Errorf("template with ID '%s' not found: %w", id, err)
	}

	// Update legacy channels that use this template before deletion
	// Set template content to empty since template is being deleted
	if err := uc.updateLegacyChannelsForTemplateDelete(ctx, templateEntity); err != nil {
		// Log error but don't fail the operation
		// The template deletion should proceed, legacy sync is best effort
		fmt.Printf("Warning: failed to update legacy channels for template deletion %s: %v\n", templateEntity.ID().String(), err)
	}

	// Delete template
	if err := uc.templateRepo.Delete(ctx, templateID); err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	return nil
}

// updateLegacyChannelsForTemplateDelete updates all legacy channels that use the template being deleted
func (uc *DeleteTemplateUseCase) updateLegacyChannelsForTemplateDelete(ctx context.Context, templateEntity *template.Template) error {
	// Find all channels that use this template
	// Since we don't have FindByTemplateID, we'll get all channels and filter
	filter := channel.NewChannelFilter()
	pagination := &shared.Pagination{MaxResultCount: 100} // Get maximum allowed channels per query

	result, err := uc.channelRepo.FindAll(ctx, filter, pagination)
	if err != nil {
		return fmt.Errorf("failed to find channels: %w", err)
	}

	// Filter channels that use this template
	var channelsUsingTemplate []*channel.Channel
	for _, ch := range result.Items {
		if ch.TemplateID() != nil && ch.TemplateID().String() == templateEntity.ID().String() {
			channelsUsingTemplate = append(channelsUsingTemplate, ch)
		}
	}

	// Update each channel in the legacy system with empty template content
	for _, ch := range channelsUsingTemplate {
		if err := uc.updateLegacyChannelForTemplateDelete(ctx, ch); err != nil {
			// Log error but continue with other channels
			fmt.Printf("Warning: failed to update legacy channel %s for template deletion: %v\n", ch.ID().String(), err)
		}
	}

	return nil
}

// updateLegacyChannelForTemplateDelete updates a single channel in the legacy system for template deletion
func (uc *DeleteTemplateUseCase) updateLegacyChannelForTemplateDelete(ctx context.Context, ch *channel.Channel) error {
	legacyURL := uc.config.LegacySystem.URL + "/Groups/" + ch.ID().String()
	bearerToken := uc.config.LegacySystem.Token

	// Construct the request body for the legacy system
	legacyReq := LegacyChannelRequest{
		Name:        ch.Name().String(),
		Description: ch.Description().String(),
		Type:        ch.ChannelType().String(),
		LevelName:   "Critical", // Assuming this is a default or derived value
		Config:      LegacyConfig{},
		SendList:    []SendListItem{},
	}

	// Populate Config from channel config
	configMap := ch.Config().ToMap()
	if host, ok := configMap["host"].(string); ok {
		legacyReq.Config.Host = host
	}
	if port, ok := configMap["port"].(float64); ok { // JSON numbers are float64
		legacyReq.Config.Port = int(port)
	}
	if secure, ok := configMap["secure"].(bool); ok {
		legacyReq.Config.Secure = secure
	}
	if method, ok := configMap["method"].(string); ok {
		legacyReq.Config.Method = method
	}
	if username, ok := configMap["username"].(string); ok {
		legacyReq.Config.Username = username
	}
	if password, ok := configMap["password"].(string); ok {
		legacyReq.Config.Password = password
	}
	if senderEmail, ok := configMap["senderEmail"].(string); ok {
		legacyReq.Config.SenderEmail = senderEmail
	}

	// Set empty template content since template is being deleted
	// Use fallback values from config or defaults
	if emailSubject, ok := configMap["emailSubject"].(string); ok {
		legacyReq.Config.EmailSubject = emailSubject
	} else {
		legacyReq.Config.EmailSubject = "Default Subject"
	}
	if templateContent, ok := configMap["template"].(string); ok {
		legacyReq.Config.Template = templateContent
	} else {
		legacyReq.Config.Template = "Default Template Content"
	}

	// Populate SendList from channel recipients
	recipients := ch.Recipients().ToSlice()
	for _, r := range recipients {
		firstName := r.Name
		lastName := ""
		if parts := strings.SplitN(r.Name, " ", 2); len(parts) > 1 {
			firstName = parts[0]
			lastName = parts[1]
		}

		legacyReq.SendList = append(legacyReq.SendList, SendListItem{
			FirstName:     firstName,
			LastName:      lastName,
			RecipientType: r.Type,
			Target:        r.Target,
		})
	}

	// Marshal the request body to JSON
	reqBody, err := json.Marshal(legacyReq)
	if err != nil {
		return fmt.Errorf("failed to marshal legacy request body: %w", err)
	}

	// Create and send the HTTP PUT request
	req, err := http.NewRequestWithContext(ctx, "PUT", legacyURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create legacy http request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to legacy system: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("legacy system returned error status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
