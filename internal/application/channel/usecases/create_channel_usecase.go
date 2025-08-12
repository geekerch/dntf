package usecases

import (
	"context"
	"fmt"

	"notification/internal/application/channel/dtos"
	"notification/internal/domain/channel"
	"notification/internal/domain/services"
	"notification/internal/domain/shared"
	"notification/internal/domain/template"
)

// CreateChannelUseCase is the use case for creating a channel.
type CreateChannelUseCase struct {
	channelRepo channel.ChannelRepository
	validator   *services.ChannelValidator
}

// NewCreateChannelUseCase creates a use case instance.
func NewCreateChannelUseCase(
	channelRepo channel.ChannelRepository,
	validator *services.ChannelValidator,
) *CreateChannelUseCase {
	return &CreateChannelUseCase{
		channelRepo: channelRepo,
		validator:   validator,
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

	// 4. Create a channel entity
	ch, err := channel.NewChannel(
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

	// 5. Persist
	if err := uc.channelRepo.Save(ctx, ch); err != nil {
		return nil, fmt.Errorf("failed to save channel: %w", err)
	}

	// 6. Convert to response DTO
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