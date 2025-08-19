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

// UpdateChannelUseCase is the use case for updating a channel.
type UpdateChannelUseCase struct {
	channelRepo channel.ChannelRepository
	validator   *services.ChannelValidator
}

// NewUpdateChannelUseCase creates a use case instance.
func NewUpdateChannelUseCase(
	channelRepo channel.ChannelRepository,
	validator *services.ChannelValidator,
) *UpdateChannelUseCase {
	return &UpdateChannelUseCase{
		channelRepo: channelRepo,
		validator:   validator,
	}
}

// Execute executes the channel update.
func (uc *UpdateChannelUseCase) Execute(ctx context.Context, channelID string, request *dtos.UpdateChannelRequest) (*dtos.ChannelResponse, error) {
	// 1. Validate input parameters
	if err := uc.validateRequest(channelID, request); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// 2. Convert to domain objects
	id, err := channel.NewChannelIDFromString(channelID)
	if err != nil {
		return nil, fmt.Errorf("invalid channel ID: %w", err)
	}

	domainObjects, err := uc.convertToDomainObjects(request)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to domain objects: %w", err)
	}

	// 3. Business validation
	if err := uc.validator.ValidateChannelForUpdate(
		ctx,
		id,
		domainObjects.Name,
		domainObjects.ChannelType,
		domainObjects.TemplateID,
		domainObjects.Config,
	); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 4. Query existing channel
	ch, err := uc.channelRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("channel not found: %w", err)
	}

	// 5. Check if the channel is deleted
	if ch.IsDeleted() {
		return nil, fmt.Errorf("cannot update deleted channel")
	}

	// 6. Update the channel
	if err := ch.Update(
		domainObjects.Name,
		domainObjects.Description,
		request.Enabled,
		domainObjects.ChannelType,
		domainObjects.TemplateID,
		domainObjects.CommonSettings,
		domainObjects.Config,
		domainObjects.Recipients,
		domainObjects.Tags,
	); err != nil {
		return nil, fmt.Errorf("failed to update channel: %w", err)
	}

	// 7. Persist
	if err := uc.channelRepo.Update(ctx, ch); err != nil {
		return nil, fmt.Errorf("failed to save channel: %w", err)
	}

	// 8. Convert to response DTO
	response := uc.convertToResponse(ch)
	return response, nil
}

// validateRequest validates the request parameters.
func (uc *UpdateChannelUseCase) validateRequest(channelID string, request *dtos.UpdateChannelRequest) error {
	if channelID == "" {
		return fmt.Errorf("channel ID is required")
	}

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

// convertToDomainObjects converts to domain objects.
func (uc *UpdateChannelUseCase) convertToDomainObjects(request *dtos.UpdateChannelRequest) (*DomainObjects, error) {
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
	channelType, err := shared.NewChannelTypeFromString(request.ChannelType)
	if err != nil {
		return nil, fmt.Errorf("invalid channel type: %s, error: %w", request.ChannelType, err)
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
func (uc *UpdateChannelUseCase) convertToResponse(ch *channel.Channel) *dtos.ChannelResponse {
	var templateID string
	if ch.TemplateID() != nil {
		templateID = ch.TemplateID().String()
	}

	return &dtos.ChannelResponse{
		ChannelID:      ch.ID().String(),
		ChannelName:    ch.Name().String(),
		Description:    ch.Description().String(),
		Enabled:        ch.IsEnabled(),
		ChannelType:    ch.ChannelType().String(),
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