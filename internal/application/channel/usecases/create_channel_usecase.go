package usecases

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"notification/internal/application/channel/dtos"
	"notification/internal/domain/channel"
	"notification/internal/domain/services"
	"notification/internal/domain/shared"
	"notification/internal/domain/template"
	"notification/pkg/config"
)

// CreateChannelUseCase is the use case for creating a channel.
type CreateChannelUseCase struct {
	channelRepo  channel.ChannelRepository
	templateRepo template.TemplateRepository
	validator    *services.ChannelValidator
	config       *config.Config
}

// NewCreateChannelUseCase creates a use case instance.
func NewCreateChannelUseCase(
	channelRepo channel.ChannelRepository,
	templateRepo template.TemplateRepository,
	validator *services.ChannelValidator,
	config *config.Config,
) *CreateChannelUseCase {
	return &CreateChannelUseCase{
		channelRepo:  channelRepo,
		templateRepo: templateRepo,
		validator:    validator,
		config:       config,
	}
}

// Execute executes the create channel operation.
func (uc *CreateChannelUseCase) Execute(ctx context.Context, request *dtos.CreateChannelRequest) (*dtos.ChannelResponse, error) {
	// 1. Validate input parameters
	if err := uc.validateRequest(request); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// 2. Convert to domain objects
	domainObjects, err := uc.convertToDomainObjects(request)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to domain objects: %w", err)
	}

	// 3. Business validation
	if err := uc.validator.ValidateChannelForCreation(
		ctx,
		domainObjects.Name,
		domainObjects.ChannelType,
		domainObjects.TemplateID,
		domainObjects.Config,
	); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 4. Forward to legacy system to get the channel ID
	groupID, err := uc.forwardToLegacySystem(ctx, domainObjects, request)
	if err != nil {
		return nil, fmt.Errorf("failed to forward to legacy system: %w", err)
	}

	channelID, err := channel.NewChannelIDFromString(groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to create channel ID from group ID: %w", err)
	}

	// 5. Create a channel entity with the ID from the legacy system
	ch, err := channel.NewChannelWithID(
		channelID,
		domainObjects.Name,
		domainObjects.Description,
		request.Enabled,
		domainObjects.ChannelType,
		domainObjects.TemplateID,
		domainObjects.CommonSettings,
		domainObjects.Config,
		domainObjects.Recipients,
		domainObjects.Tags,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	// 6. Persist
	if err := uc.channelRepo.Save(ctx, ch); err != nil {
		return nil, fmt.Errorf("failed to save channel: %w", err)
	}

	// 7. Convert to response DTO
	response := uc.convertToResponse(ch)
	return response, nil
}

// validateRequest validates the request parameters.
func (uc *CreateChannelUseCase) validateRequest(request *dtos.CreateChannelRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if request.ChannelName == "" {
		return fmt.Errorf("channel name is required")
	}

	if request.ChannelType == "" {
		return fmt.Errorf("channel type is required")
	}

	return nil
}

// DomainObjects are the converted domain objects.
type DomainObjects struct {
	Name           *channel.ChannelName
	Description    *channel.Description
	ChannelType    shared.ChannelType
	TemplateID     *template.TemplateID
	CommonSettings *shared.CommonSettings
	Config         *channel.ChannelConfig
	Recipients     *channel.Recipients
	Tags           *channel.Tags
}

// LegacyChannelRequest defines the request payload for the legacy system.
type LegacyChannelRequest struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Type        string         `json:"type"`
	LevelName   string         `json:"levelName"`
	Config      LegacyConfig   `json:"config"`
	SendList    []SendListItem `json:"sendList"`
}

// LegacyConfig defines the config for the legacy system.
type LegacyConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Secure       bool   `json:"secure"`
	Method       string `json:"method"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	SenderEmail  string `json:"senderEmail"`
	EmailSubject string `json:"emailSubject"`
	Template     string `json:"template"`
}

// SendListItem defines a recipient for the legacy system.
type SendListItem struct {
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	RecipientType string `json:"recipientType"`
	Target        string `json:"target"`
}

// convertToDomainObjects converts to domain objects.
func (uc *CreateChannelUseCase) convertToDomainObjects(request *dtos.CreateChannelRequest) (*DomainObjects, error) {
	// Channel name
	name, err := channel.NewChannelName(request.ChannelName)
	if err != nil {
		return nil, fmt.Errorf("invalid channel name: %w", err)
	}

	// Description
	description, err := channel.NewDescription(request.Description)
	if err != nil {
		return nil, fmt.Errorf("invalid description: %w", err)
	}

	// Channel type
	channelType := shared.ChannelType(request.ChannelType)
	if !channelType.IsValid() {
		return nil, fmt.Errorf("invalid channel type: %s", request.ChannelType)
	}

	// Template ID
	var templateID *template.TemplateID
	if request.TemplateID != "" {
		templateID, err = template.NewTemplateIDFromString(request.TemplateID)
		if err != nil {
			return nil, fmt.Errorf("invalid template ID: %w", err)
		}
	}

	// Common settings
	commonSettings, err := request.CommonSettings.ToCommonSettings()
	if err != nil {
		return nil, fmt.Errorf("invalid common settings: %w", err)
	}

	// Channel configuration
	config := channel.NewChannelConfig(request.Config)

	// Recipients
	recipientSlice, err := dtos.ToRecipientsSlice(request.Recipients)
	if err != nil {
		return nil, fmt.Errorf("invalid recipients: %w", err)
	}
	recipients := channel.NewRecipients(recipientSlice)

	// Tags
	tags := channel.NewTags(request.Tags)

	return &DomainObjects{
		Name:           name,
		Description:    description,
		ChannelType:    channelType,
		TemplateID:     templateID,
		CommonSettings: commonSettings,
		Config:         config,
		Recipients:     recipients,
		Tags:           tags,
	}, nil
}

// convertToResponse converts to a response DTO.
func (uc *CreateChannelUseCase) convertToResponse(ch *channel.Channel) *dtos.ChannelResponse {
	var templateID string
	if ch.TemplateID() != nil {
		templateID = ch.TemplateID().String()
	}

	return &dtos.ChannelResponse{
		ChannelID:      ch.ID().String(),
		ChannelName:    ch.Name().String(),
		Description:    ch.Description().String(),
		Enabled:        ch.IsEnabled(),
		ChannelType:    string(ch.ChannelType()),
		TemplateID:     templateID,
		CommonSettings: dtos.FromCommonSettings(ch.CommonSettings()),
		Config:         ch.Config().ToMap(),
		Recipients:     dtos.FromRecipientsSlice(ch.Recipients().ToSlice()),
		Tags:           ch.Tags().ToSlice(),
		CreatedAt:      ch.Timestamps().CreatedAt,
		UpdatedAt:      ch.Timestamps().UpdatedAt,
		LastUsed:       ch.LastUsed(),
	}
}

func (uc *CreateChannelUseCase) forwardToLegacySystem(ctx context.Context, domainObjects *DomainObjects, request *dtos.CreateChannelRequest) (string, error) {
	legacyURL := uc.config.LegacySystem.URL + "/api/v2.0/Groups"
	bearerToken := uc.config.LegacySystem.Token

	// 1. Construct the request body for the legacy system
	legacyReq := LegacyChannelRequest{
		Name:        domainObjects.Name.String(),
		Description: domainObjects.Description.String(),
		Type:        string(domainObjects.ChannelType),
		LevelName:   "Critical", // Assuming this is a default or derived value
		Config:      LegacyConfig{},
		SendList:    []SendListItem{},
	}

	// 1a. Fetch template if TemplateID exists
	var foundTemplate *template.Template
	if domainObjects.TemplateID != nil {
		var err error
		foundTemplate, err = uc.templateRepo.FindByID(ctx, domainObjects.TemplateID)
		if err != nil {
			// Decide if a missing template is a fatal error. For now, let's assume it is.
			return "", fmt.Errorf("failed to find template with ID %s: %w", domainObjects.TemplateID.String(), err)
		}
	}

	// 2. Populate Config from ch.Config() and template
	configMap := domainObjects.Config.ToMap()
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

	// Prioritize template values for subject and content
	if foundTemplate != nil {
		legacyReq.Config.EmailSubject = foundTemplate.Subject().String()
		legacyReq.Config.Template = foundTemplate.Content().String()
	} else {
		// Fallback to configMap if no template
		if emailSubject, ok := configMap["emailSubject"].(string); ok {
			legacyReq.Config.EmailSubject = emailSubject
		}
		if template, ok := configMap["template"].(string); ok {
			legacyReq.Config.Template = template
		}
	}

	// 3. Populate SendList from ch.Recipients()
	recipientDTOs := request.Recipients
	for _, r := range recipientDTOs {
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

	// 4. Marshal the request body to JSON
	reqBody, err := json.Marshal(legacyReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal legacy request body: %w", err)
	}

	// 5. Create and send the HTTP POST request
	req, err := http.NewRequestWithContext(ctx, "POST", legacyURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create legacy http request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to legacy system: %w", err)
	}
	defer resp.Body.Close()

	// 6. Check response status
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("legacy system returned error status %d: %s", resp.StatusCode, string(body))
	}

	// 7. Parse the response
	var legacyResp struct {
		GroupID string `json:"groupId"`
		Name    string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&legacyResp); err != nil {
		return "", fmt.Errorf("failed to decode legacy response body: %w", err)
	}

	return legacyResp.GroupID, nil
}
