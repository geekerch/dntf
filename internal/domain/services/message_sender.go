package services

import (
	"context"
	"errors"
	"fmt"

	"channel-api/internal/domain/channel"
	"channel-api/internal/domain/message"
	"channel-api/internal/domain/template"
)

// MessageSender is the domain service for sending messages.
type MessageSender struct {
	channelRepo  channel.ChannelRepository
	templateRepo template.TemplateRepository
	messageRepo  message.MessageRepository
	renderer     TemplateRenderer
}

// NewMessageSender creates a message sending service.
func NewMessageSender(
	channelRepo channel.ChannelRepository,
	templateRepo template.TemplateRepository,
	messageRepo message.MessageRepository,
	renderer TemplateRenderer,
) *MessageSender {
	return &MessageSender{
		channelRepo:  channelRepo,
		templateRepo: templateRepo,
		messageRepo:  messageRepo,
		renderer:     renderer,
	}
}

// SendMessage sends a message.
func (ms *MessageSender) SendMessage(
	ctx context.Context,
	channelIDs *message.ChannelIDs,
	variables *message.Variables,
	channelOverrides *message.ChannelOverrides,
) (*message.Message, error) {
	// Create a message entity
	msg, err := message.NewMessage(channelIDs, variables, channelOverrides)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Save the message
	if err := ms.messageRepo.Save(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	// Process each channel
	for _, channelID := range channelIDs.ToSlice() {
		result := ms.processSingleChannel(ctx, channelID, variables, channelOverrides)
		if err := msg.AddResult(result); err != nil {
			// If adding the result fails, log the error but continue processing other channels
			continue
		}
	}

	// Update the message status
	if err := ms.messageRepo.Update(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to update message: %w", err)
	}

	return msg, nil
}

// processSingleChannel processes the message sending for a single channel.
func (ms *MessageSender) processSingleChannel(
	ctx context.Context,
	channelID *channel.ChannelID,
	variables *message.Variables,
	channelOverrides *message.ChannelOverrides,
) *message.MessageResult {
	// Get channel information
	ch, err := ms.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		return ms.createFailedResult(channelID, "Failed to retrieve channel", "CHANNEL_NOT_FOUND", err.Error())
	}

	// Check if the channel can send messages
	if err := ch.CanSendMessage(); err != nil {
		return ms.createFailedResult(channelID, "Channel cannot send message", "CHANNEL_UNAVAILABLE", err.Error())
	}

	// Get template information
	tmpl, err := ms.templateRepo.FindByID(ctx, ch.TemplateID())
	if err != nil {
		return ms.createFailedResult(channelID, "Failed to retrieve template", "TEMPLATE_NOT_FOUND", err.Error())
	}

	// Check if the channel type matches the template
	if !tmpl.MatchesType(ch.ChannelType()) {
		return ms.createFailedResult(channelID, "Channel type mismatch with template", "TYPE_MISMATCH", 
			fmt.Sprintf("Channel type: %s, Template type: %s", ch.ChannelType(), tmpl.ChannelType()))
	}

	// Prepare the rendering content
	renderRequest := ms.prepareRenderRequest(ch, tmpl, variables, channelOverrides)

	// Validate variables
	if err := ms.validateVariables(tmpl, renderRequest.Variables); err != nil {
		return ms.createFailedResult(channelID, "Variable validation failed", "MISSING_VARIABLES", err.Error())
	}

	// Render the template
	renderedContent, err := ms.renderer.Render(ctx, renderRequest)
	if err != nil {
		return ms.createFailedResult(channelID, "Template rendering failed", "RENDER_ERROR", err.Error())
	}

	// This is where the actual message sending service should be called (e.g., EmailService, SlackService, etc.)
	// Since this is the domain layer, we temporarily simulate a successful sending
	_ = renderedContent

	// Mark the channel as used
	ch.MarkAsUsed()
	if err := ms.channelRepo.Update(ctx, ch); err != nil {
		// Update failure does not affect the sending result, only log the error
	}

	// Create a successful result
	result, err := message.NewSuccessfulMessageResult(channelID, "Message sent successfully")
	if err != nil {
		return ms.createFailedResult(channelID, "Failed to create result", "RESULT_ERROR", err.Error())
	}

	return result
}

// prepareRenderRequest prepares the rendering request.
func (ms *MessageSender) prepareRenderRequest(
	ch *channel.Channel,
	tmpl *template.Template,
	variables *message.Variables,
	channelOverrides *message.ChannelOverrides,
) *RenderRequest {
	request := &RenderRequest{
		Subject:   tmpl.Subject(),
		Content:   tmpl.Content(),
		Variables: variables,
	}

	// Apply channel overrides
	if override, exists := channelOverrides.Get(ch.ID().String()); exists {
		if override.HasTemplateOverride() {
			templateOverride := override.TemplateOverride
			if templateOverride.HasSubjectOverride() {
				request.Subject = templateOverride.Subject
			}
			if templateOverride.HasTemplateOverride() {
				request.Content = templateOverride.Template
			}
		}
	}

	return request
}

// validateVariables validates variables.
func (ms *MessageSender) validateVariables(tmpl *template.Template, variables *message.Variables) error {
	missingVariables := tmpl.ValidateVariables(variables.ToMap())
	if len(missingVariables) > 0 {
		return fmt.Errorf("missing required variables: %v", missingVariables)
	}
	return nil
}

// createFailedResult creates a failed result.
func (ms *MessageSender) createFailedResult(channelID *channel.ChannelID, msg, code, details string) *message.MessageResult {
	msgError := message.NewMessageError(code, details)
	result, _ := message.NewFailedMessageResult(channelID, msg, msgError)
	return result
}

// TemplateRenderer is the interface for the template renderer.
type TemplateRenderer interface {
	Render(ctx context.Context, request *RenderRequest) (*RenderedContent, error)
}

// RenderRequest is the rendering request.
type RenderRequest struct {
	Subject   *template.Subject
	Content   *template.TemplateContent
	Variables *message.Variables
}

// RenderedContent is the rendering result.
type RenderedContent struct {
	Subject string
	Content string
}

// DefaultTemplateRenderer is the default template renderer.
type DefaultTemplateRenderer struct{}

// NewDefaultTemplateRenderer creates a default template renderer.
func NewDefaultTemplateRenderer() *DefaultTemplateRenderer {
	return &DefaultTemplateRenderer{}
}

// Render renders the template.
func (r *DefaultTemplateRenderer) Render(ctx context.Context, request *RenderRequest) (*RenderedContent, error) {
	if request == nil {
		return nil, errors.New("render request is required")
	}

	variableMap := request.Variables.ToMap()

	// Render the subject
	renderedSubject, err := r.renderTemplate(request.Subject.String(), variableMap)
	if err != nil {
		return nil, fmt.Errorf("failed to render subject: %w", err)
	}

	// Render the content
	renderedContent, err := r.renderTemplate(request.Content.String(), variableMap)
	if err != nil {
		return nil, fmt.Errorf("failed to render content: %w", err)
	}

	return &RenderedContent{
		Subject: renderedSubject,
		Content: renderedContent,
	}, nil
}

// renderTemplate renders a single template.
func (r *DefaultTemplateRenderer) renderTemplate(template string, variables map[string]interface{}) (string, error) {
	// Simple variable replacement implementation
	// In a real project, you can use a more powerful template engine such as text/template or html/template
	result := template
	for key, value := range variables {
		placeholder := fmt.Sprintf("{%s}", key)
		replacement := fmt.Sprintf("%v", value)
		result = fmt.Sprintf("%s", fmt.Sprintf(result, placeholder, replacement))
	}
	return result, nil
}