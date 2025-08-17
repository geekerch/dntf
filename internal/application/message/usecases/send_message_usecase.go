package usecases

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"notification/internal/application/message/dtos"
	"notification/internal/domain/channel"
	"notification/internal/domain/message"
	"notification/internal/domain/services"
	"notification/internal/domain/template"
	"notification/pkg/config"
	"time"

	"github.com/google/uuid"
)

// LegacyMessageRequest defines the request payload for the legacy system.
type LegacyMessageRequest struct {
	GroupID     string                 `json:"groupId"`
	Subject     string                 `json:"subject"`
	UseTemplate bool                   `json:"useTemplate"`
	Message     string                 `json:"message"`
	Header      string                 `json:"header"`
	Footer      string                 `json:"footer"`
	Variables   map[string]interface{} `json:"variables"`
	SendList    []LegacySendListItem   `json:"sendList"`
	Attachments []LegacyAttachment     `json:"attachments"`
}

// LegacySendListItem defines a recipient for the legacy system.
type LegacySendListItem struct {
	Target        string `json:"target"`
	RecipientType string `json:"recipientType"`
}

// LegacyAttachment defines an attachment for the legacy system.
type LegacyAttachment struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
	Type     string `json:"type"`
}

// LegacyMessageResponse represents the response from the legacy system.
type LegacyMessageResponse struct {
	Result  []LegacyResult `json:"result"`
	GroupID string         `json:"groupId"`
}

// LegacyResult represents a result object in the legacy response.
type LegacyResult struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

// SendMessageUseCase handles sending messages.
type SendMessageUseCase struct {
	messageRepo   message.MessageRepository
	channelRepo   channel.ChannelRepository
	templateRepo  template.TemplateRepository
	messageSender *services.EnhancedMessageSender
	config        *config.Config
}

// NewSendMessageUseCase creates a new SendMessageUseCase.
func NewSendMessageUseCase(
	messageRepo message.MessageRepository,
	channelRepo channel.ChannelRepository,
	templateRepo template.TemplateRepository,
	messageSender *services.EnhancedMessageSender,
	config *config.Config,
) *SendMessageUseCase {
	return &SendMessageUseCase{
		messageRepo:   messageRepo,
		channelRepo:   channelRepo,
		templateRepo:  templateRepo,
		messageSender: messageSender,
		config:        config,
	}
}

// Execute sends a message.
func (uc *SendMessageUseCase) Execute(ctx context.Context, req *dtos.SendMessageRequest) (*dtos.MessageResponse, error) {
	// Validate request
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	if len(req.ChannelIDs) == 0 {
		return nil, fmt.Errorf("at least one channel ID is required")
	}

	// Create channel IDs from string slice
	var channelIDEntities []*channel.ChannelID
	for _, channelIDStr := range req.ChannelIDs {
		channelID, err := channel.NewChannelIDFromString(channelIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid channel ID '%s': %w", channelIDStr, err)
		}
		channelIDEntities = append(channelIDEntities, channelID)
	}

	// Create template ID
	templateID, err := template.NewTemplateIDFromString(req.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("invalid template ID: %w", err)
	}

	// Validate all channels exist and get the first one for template validation
	var firstChannelEntity *channel.Channel
	for i, channelID := range channelIDEntities {
		channelEntity, err := uc.channelRepo.FindByID(ctx, channelID)
		if err != nil {
			return nil, fmt.Errorf("failed to find channel '%s': %w", req.ChannelIDs[i], err)
		}
		if i == 0 {
			firstChannelEntity = channelEntity
		}
	}

	// Validate template exists
	templateEntity, err := uc.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to find template: %w", err)
	}

	// Validate channel type matches template channel type (using first channel)
	if firstChannelEntity.ChannelType() != templateEntity.ChannelType() {
		return nil, fmt.Errorf("channel type '%s' does not match template channel type '%s'",
			firstChannelEntity.ChannelType(), templateEntity.ChannelType())
	}

	// Create channel IDs
	channelIDs, err := message.NewChannelIDs(channelIDEntities)
	if err != nil {
		return nil, fmt.Errorf("invalid channel IDs: %w", err)
	}

	// Create variables if provided
	var variables *message.Variables
	if req.Variables != nil {
		variables = message.NewVariables(req.Variables)
	} else {
		variables = message.NewVariables(nil)
	}

	// Create channel overrides if provided
	var channelOverrides *message.ChannelOverrides
	if req.ChannelOverrides != nil {
		channelOverrides = message.NewChannelOverrides(req.ChannelOverrides.ToMap())
	} else {
		channelOverrides = message.NewChannelOverrides(nil)
	}

	// Send message using domain service
	messageEntity, err := uc.messageSender.SendMessage(
		ctx,
		channelIDs,
		variables,
		channelOverrides,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	// Convert to response
	return dtos.ToMessageResponse(messageEntity), nil
}

// Forward sends a message via the legacy system.
func (uc *SendMessageUseCase) Forward(ctx context.Context, req *dtos.SendMessageRequest) ([]*dtos.MessageResponse, error) {
	legacyURL := uc.config.LegacySystem.URL + "/api/v2.0/Groups/send" // This might need adjustment
	bearerToken := uc.config.LegacySystem.Token

	// Validate request
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	if len(req.ChannelIDs) == 0 {
		return nil, fmt.Errorf("at least one channel ID is required")
	}

	// 1. Get Template info
	var templateEntity *template.Template
	if req.TemplateID != "" {
		templateID, err := template.NewTemplateIDFromString(req.TemplateID)
		if err != nil {
			return nil, fmt.Errorf("invalid template ID: %w", err)
		}
		templateEntity, err = uc.templateRepo.FindByID(ctx, templateID)
		if err != nil {
			return nil, fmt.Errorf("failed to find template: %w", err)
		}
	}

	// 2. Create legacy requests for each channel
	var legacyRequests []LegacyMessageRequest

	for _, channelIDStr := range req.ChannelIDs {
		// Get Channel info
		channelID, err := channel.NewChannelIDFromString(channelIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid channel ID '%s': %w", channelIDStr, err)
		}
		channelEntity, err := uc.channelRepo.FindByID(ctx, channelID)
		if err != nil {
			return nil, fmt.Errorf("failed to find channel '%s': %w", channelIDStr, err)
		}

		// Construct the request body for the legacy system
		sendList := make([]LegacySendListItem, len(req.Recipients))
		for i, r := range req.Recipients {
			sendList[i] = LegacySendListItem{
				Target:        r,
				RecipientType: string(channelEntity.ChannelType()),
			}
		}

		legacyReq := LegacyMessageRequest{
			GroupID:     channelIDStr,
			Header:      "", // Assuming no header from SendMessageRequest
			Footer:      "", // Assuming no footer from SendMessageRequest
			UseTemplate: true,
			Variables:   req.Variables,
			SendList:    sendList,
			Attachments: []LegacyAttachment{}, // Assuming no attachments from SendMessageRequest
			Subject:     "test",
			Message:     "test",
		}

		if templateEntity != nil {
			legacyReq.UseTemplate = false
			legacyReq.Subject = templateEntity.Subject().String()
			legacyReq.Message = templateEntity.Content().String()

		}
		legacyRequests = append(legacyRequests, legacyReq)
	}

	// 3. Marshal the request body
	reqBody, err := json.Marshal(legacyRequests)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal legacy request body: %w", err)
	}

	// 4. Create and send the HTTP POST request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", legacyURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create legacy http request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+bearerToken)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second, // Set reasonable timeout
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to legacy system: %w", err)
	}
	defer resp.Body.Close()

	// 5. Check response status
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("legacy system returned error status %d: %s", resp.StatusCode, string(body))
	}

	// 6. Parse the response and convert to a MessageResponse DTO
	var legacyResp []LegacyMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&legacyResp); err != nil {
		return nil, fmt.Errorf("failed to decode legacy response body: %w", err)
	}

	if len(legacyResp) == 0 {
		return nil, fmt.Errorf("legacy system returned an empty response array")
	}

	// 7. Process responses - individual channel status will be handled in response creation

	// 8. Create response array with information from all processed channels
	var messageResponses []*dtos.MessageResponse
	currentTime := time.Now().UnixMilli()

	for _, result := range legacyResp {
		// Determine status for this specific channel
		channelStatus := message.MessageStatusSuccess
		var channelResults []*dtos.MessageResultResponse

		// Create results for each recipient based on legacy response
		for i, r := range result.Result {
			var recipient string
			if i < len(req.Recipients) {
				recipient = req.Recipients[i]
			} else {
				recipient = "unknown"
			}

			resultResponse := &dtos.MessageResultResponse{
				Recipient: recipient,
				Status:    message.MessageResultStatusSuccess,
			}

			if r.StatusCode >= 400 {
				channelStatus = message.MessageStatusFailed
				resultResponse.Status = message.MessageResultStatusFailed
				resultResponse.Error = r.Message
			} else {
				resultResponse.SentAt = &currentTime
			}

			channelResults = append(channelResults, resultResponse)
		}

		messageResponse := &dtos.MessageResponse{
			ID:         uuid.New().String(),
			ChannelID:  result.GroupID,
			TemplateID: req.TemplateID,
			Recipients: req.Recipients,
			Variables:  req.Variables,
			Status:     channelStatus,
			Results:    channelResults,
			CreatedAt:  currentTime,
			SentAt:     currentTime,
		}

		messageResponses = append(messageResponses, messageResponse)
	}

	return messageResponses, nil
}
