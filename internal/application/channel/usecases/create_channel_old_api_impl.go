package usecases

import (
	"context"
	"fmt"
	"time"

	"notification/internal/application/channel/dtos"
	"notification/internal/infrastructure/external"
)

// CreateChannelOldAPIUseCase implements the CreateChannelUseCase interface
// by calling the old system's /v2/groups API.
type CreateChannelOldAPIUseCase struct {
	oldSystemClient external.OldSystemClient
}

// NewCreateChannelOldAPIUseCase creates a new instance of CreateChannelOldAPIUseCase.
func NewCreateChannelOldAPIUseCase(client external.OldSystemClient) *CreateChannelOldAPIUseCase {
	return &CreateChannelOldAPIUseCase{
		oldSystemClient: client,
	}
}

// Execute handles the creation of a channel by calling the old system's API.
// It maps the new system's CreateChannelRequest to the old system's API request
// and maps the old system's API response back to the new system's ChannelResponse.
func (uc *CreateChannelOldAPIUseCase) Execute(ctx context.Context, request *dtos.CreateChannelRequest) (*dtos.ChannelResponse, error) {
	// --- Mapping from new system's CreateChannelRequest to old system's API request ---
	oldAPIReq := external.OldSystemCreateGroupRequest{
		Name:        request.ChannelName,
		Description: request.Description,
		Type:        request.ChannelType, // Assuming ChannelType maps directly to old system's type
		LevelName:   "",                  // No direct mapping from CreateChannelRequest, set default or derive
	}

	// Map Config (map[string]interface{}) to strongly-typed OldSystemCreateGroupRequest.Config
	/*	if configMap, ok := request.Config.(map[string]interface{}); ok {
			if host, ok := configMap["host"].(string); ok {
				oldAPIReq.Config.Host = host
			}
			if port, ok := configMap["port"].(float64); ok {
				oldAPIReq.Config.Port = int(port)回
			}
			if secure, ok := configMap["secure"].(bool); ok {
				oldAPIReq.Config.Secure = secure
			}
			if method, ok := configMap["method"].(string); ok {
				oldAPIReq.Config.Method = method
			}
			if username, ok := configMap["username"].(string); ok {
				oldAPIReq.Config.Username = username
			}
			if password, ok := configMap["password"].(string); ok {
				oldAPIReq.Config.Password = password
			}
			if senderEmail, ok := configMap["senderEmail"].(string); ok {
				oldAPIReq.Config.SenderEmail = senderEmail
			}
			if emailSubject, ok := configMap["emailSubject"].(string); ok {
				oldAPIReq.Config.EmailSubject = emailSubject
			}
			if template, ok := configMap["template"].(string); ok {
				oldAPIReq.Config.Template = template
			}
		}
	*/
	// Map Recipients to SendList
	oldAPIReq.SendList = make([]struct {
		FirstName     string `json:"firstName"`
		LastName      string `json:"lastName"`
		RecipientType string `json:"recipientType"`
		Target        string `json:"target"`
	}, len(request.Recipients))

	for i, r := range request.Recipients {
		oldAPIReq.SendList[i].FirstName = r.Name // Assuming Name can be used as FirstName
		oldAPIReq.SendList[i].LastName = ""      // No direct mapping for LastName
		oldAPIReq.SendList[i].RecipientType = r.Type
		oldAPIReq.SendList[i].Target = r.Target // Assuming Target is the email/phone number
	}

	// --- Call the old system's API ---
	oldAPIResp, err := uc.oldSystemClient.CreateGroup(oldAPIReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create group in old system: %w", err)
	}

	// --- Mapping from old system's API response to new system's ChannelResponse ---
	// Note: The old system's response is very limited compared to ChannelResponse.
	// Many fields in ChannelResponse will be empty or default values.
	response := &dtos.ChannelResponse{
		ChannelID:   oldAPIResp.ID, // Populate ChannelID from old system's ID
		ChannelName: request.ChannelName,
		Description: request.Description,
		Enabled:     request.Enabled,
		ChannelType: request.ChannelType,
		Config:      request.Config,     // Use the original config from request
		Recipients:  request.Recipients, // Use the original recipients from request
		Tags:        request.Tags,
		CreatedAt:   time.Now().Unix(), // Set current time as creation time
		UpdatedAt:   time.Now().Unix(), // Set current time as update time
		// LastUsed will be nil as old system doesn't provide it
	}

	return response, nil
}
