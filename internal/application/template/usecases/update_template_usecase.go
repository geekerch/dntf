package usecases

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"notification/internal/application/template/dtos"
	"notification/internal/domain/channel"
	"notification/internal/domain/shared"
	"notification/internal/domain/template"
	"notification/pkg/config"
)

// UpdateTemplateUseCase handles updating templates.
type UpdateTemplateUseCase struct {
	templateRepo template.TemplateRepository
	channelRepo  channel.ChannelRepository
	config       *config.Config
}

// NewUpdateTemplateUseCase creates a new UpdateTemplateUseCase.
func NewUpdateTemplateUseCase(
	templateRepo template.TemplateRepository,
	channelRepo channel.ChannelRepository,
	config *config.Config,
) *UpdateTemplateUseCase {
	return &UpdateTemplateUseCase{
		templateRepo: templateRepo,
		channelRepo:  channelRepo,
		config:       config,
	}
}

// Execute updates a template.
func (uc *UpdateTemplateUseCase) Execute(ctx context.Context, id string, req *dtos.UpdateTemplateRequest) (*dtos.TemplateResponse, error) {
	// Validate input
	if id == "" {
		return nil, fmt.Errorf("template ID cannot be empty")
	}
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Create template ID
	templateID, err := template.NewTemplateIDFromString(id)
	if err != nil {
		return nil, fmt.Errorf("invalid template ID: %w", err)
	}

	// Find existing template
	templateEntity, err := uc.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to find template: %w", err)
	}

	// Update name if provided
	var updatedName *template.TemplateName
	if req.Name != nil {
		templateName, err := template.NewTemplateName(*req.Name)
		if err != nil {
			return nil, fmt.Errorf("invalid template name: %w", err)
		}

		// Check if another template with same name exists
		if templateName.String() != templateEntity.Name().String() {
			exists, err := uc.templateRepo.ExistsByName(ctx, templateName)
			if err != nil {
				return nil, fmt.Errorf("failed to check template name existence: %w", err)
			}
			if exists {
				return nil, fmt.Errorf("template with name '%s' already exists", *req.Name)
			}
		}

		updatedName = templateName
	} else {
		updatedName = templateEntity.Name()
	}

	// Update subject if provided
	var updatedSubject *template.Subject
	if req.Subject != nil {
		if *req.Subject == "" {
			updatedSubject = nil
		} else {
			subject, err := template.NewSubject(*req.Subject)
			if err != nil {
				return nil, fmt.Errorf("invalid subject: %w", err)
			}
			updatedSubject = subject
		}
	} else {
		updatedSubject = templateEntity.Subject()
	}

	// Update content if provided
	var updatedContent *template.TemplateContent
	if req.Content != nil {
		templateContent, err := template.NewTemplateContent(*req.Content)
		if err != nil {
			return nil, fmt.Errorf("invalid template content: %w", err)
		}
		updatedContent = templateContent
	} else {
		updatedContent = templateEntity.Content()
	}

	// Update tags if provided
	var updatedTags *template.Tags
	if req.Tags != nil {
		updatedTags = template.NewTags(req.Tags)
	} else {
		updatedTags = templateEntity.Tags()
	}

	// Create description (keep existing or empty)
	description := templateEntity.Description()

	// Update the template using the Update method
	if err := templateEntity.Update(
		updatedName,
		description,
		templateEntity.ChannelType(),
		updatedSubject,
		updatedContent,
		updatedTags,
	); err != nil {
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	// Save updated template
	if err := uc.templateRepo.Update(ctx, templateEntity); err != nil {
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	// Update legacy channels that use this template
	if err := uc.updateLegacyChannelsUsingTemplate(ctx, templateEntity); err != nil {
		// Log error but don't fail the operation
		// The template update was successful, legacy sync is best effort
		fmt.Printf("Warning: failed to update legacy channels using template %s: %v\n", templateEntity.ID().String(), err)
	}

	// Convert to response
	return dtos.ToTemplateResponse(templateEntity), nil
}

// updateLegacyChannelsUsingTemplate updates all legacy channels that use the given template
func (uc *UpdateTemplateUseCase) updateLegacyChannelsUsingTemplate(ctx context.Context, templateEntity *template.Template) error {
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

	// Update each channel in the legacy system
	for _, ch := range channelsUsingTemplate {
		if err := uc.updateLegacyChannel(ctx, ch, templateEntity); err != nil {
			// Log error but continue with other channels
			fmt.Printf("Warning: failed to update legacy channel %s: %v\n", ch.ID().String(), err)
		}
	}

	return nil
}

// updateLegacyChannel updates a single channel in the legacy system
func (uc *UpdateTemplateUseCase) updateLegacyChannel(ctx context.Context, ch *channel.Channel, templateEntity *template.Template) error {
	legacyURL := uc.config.LegacySystem.URL + "/api/v2.0/Groups/" + ch.ID().String()
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

	// Use updated template values for subject and content
	legacyReq.Config.EmailSubject = templateEntity.Subject().String()
	legacyReq.Config.Template = templateEntity.Content().String()

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