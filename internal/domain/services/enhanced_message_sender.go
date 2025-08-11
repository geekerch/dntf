package services

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"channel-api/internal/domain/channel"
	"channel-api/internal/domain/message"
	"channel-api/internal/domain/template"
	"channel-api/pkg/logger"
)

// ExternalNotificationService defines the interface for external notification service
type ExternalNotificationService interface {
	// SendSingleNotification sends a notification through a single channel
	SendSingleNotification(ctx context.Context, request *SendRequest) *SendResult
	
	// ValidateChannel validates if a channel can be used for sending
	ValidateChannel(ch *channel.Channel) error
}

// SendRequest represents a message sending request
type SendRequest struct {
	Channel   *channel.Channel
	Content   *RenderedContent
	Variables map[string]interface{}
}

// SendResult represents the result of a message sending operation
type SendResult struct {
	Success   bool
	Message   string
	Error     error
	Details   map[string]interface{}
	SentAt    int64
}

// EnhancedMessageSender is an improved version of MessageSender with external service integration
type EnhancedMessageSender struct {
	channelRepo           channel.ChannelRepository
	templateRepo          template.TemplateRepository
	messageRepo           message.MessageRepository
	renderer              TemplateRenderer
	notificationService   ExternalNotificationService
	logger                *logger.Logger
}

// NewEnhancedMessageSender creates an enhanced message sender
func NewEnhancedMessageSender(
	channelRepo channel.ChannelRepository,
	templateRepo template.TemplateRepository,
	messageRepo message.MessageRepository,
	renderer TemplateRenderer,
	notificationService ExternalNotificationService,
	logger *logger.Logger,
) *EnhancedMessageSender {
	return &EnhancedMessageSender{
		channelRepo:         channelRepo,
		templateRepo:        templateRepo,
		messageRepo:         messageRepo,
		renderer:            renderer,
		notificationService: notificationService,
		logger:              logger,
	}
}

// SendMessage sends a message through multiple channels
func (s *EnhancedMessageSender) SendMessage(
	ctx context.Context,
	channelIDs *message.ChannelIDs,
	variables *message.Variables,
	channelOverrides *message.ChannelOverrides,
) (*message.Message, error) {
	startTime := time.Now()
	
	s.logger.Info("Starting message sending process",
		zap.Int("channel_count", channelIDs.Count()),
		zap.Strings("variable_keys", variables.Keys()))

	// Create message entity
	msg, err := message.NewMessage(channelIDs, variables, channelOverrides)
	if err != nil {
		s.logger.Error("Failed to create message entity", zap.Error(err))
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Save initial message
	if err := s.messageRepo.Save(ctx, msg); err != nil {
		s.logger.Error("Failed to save initial message", zap.Error(err))
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	s.logger.Info("Message entity created and saved",
		zap.String("message_id", msg.ID().String()))

	// Process each channel
	successCount := 0
	for _, channelID := range channelIDs.ToSlice() {
		result := s.processSingleChannelEnhanced(ctx, channelID, variables, channelOverrides)
		
		if err := msg.AddResult(result); err != nil {
			s.logger.Error("Failed to add result to message",
				zap.String("channel_id", channelID.String()),
				zap.Error(err))
			continue
		}

		if result.IsSuccess() {
			successCount++
		}

		s.logger.Info("Channel processing completed",
			zap.String("channel_id", channelID.String()),
			zap.String("status", string(result.Status())),
			zap.String("message", result.Message()))
	}

	// Update message with results
	if err := s.messageRepo.Update(ctx, msg); err != nil {
		s.logger.Error("Failed to update message with results", zap.Error(err))
		return nil, fmt.Errorf("failed to update message: %w", err)
	}

	duration := time.Since(startTime)
	s.logger.Info("Message sending process completed",
		zap.String("message_id", msg.ID().String()),
		zap.String("status", string(msg.Status())),
		zap.Int("success_count", successCount),
		zap.Int("total_count", channelIDs.Count()),
		zap.Duration("duration", duration))

	return msg, nil
}

// processSingleChannelEnhanced processes a single channel with enhanced error handling and logging
func (s *EnhancedMessageSender) processSingleChannelEnhanced(
	ctx context.Context,
	channelID *channel.ChannelID,
	variables *message.Variables,
	channelOverrides *message.ChannelOverrides,
) *message.MessageResult {
	channelLogger := s.logger.WithFields(zap.String("channel_id", channelID.String()))

	// Get channel information
	ch, err := s.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		channelLogger.Error("Failed to retrieve channel", zap.Error(err))
		return s.createFailedResult(channelID, "Failed to retrieve channel", "CHANNEL_NOT_FOUND", err.Error())
	}

	channelLogger = channelLogger.WithFields(
		zap.String("channel_name", ch.Name().String()),
		zap.String("channel_type", string(ch.ChannelType())))

	// Check if channel can send messages
	if err := ch.CanSendMessage(); err != nil {
		channelLogger.Warn("Channel cannot send message", zap.Error(err))
		return s.createFailedResult(channelID, "Channel cannot send message", "CHANNEL_UNAVAILABLE", err.Error())
	}

	// Validate channel with external service
	if err := s.notificationService.ValidateChannel(ch); err != nil {
		channelLogger.Warn("Channel validation failed", zap.Error(err))
		return s.createFailedResult(channelID, "Channel validation failed", "CHANNEL_INVALID", err.Error())
	}

	// Get template information if specified
	var tmpl *template.Template
	if ch.TemplateID() != nil {
		tmpl, err = s.templateRepo.FindByID(ctx, ch.TemplateID())
		if err != nil {
			channelLogger.Error("Failed to retrieve template", zap.Error(err))
			return s.createFailedResult(channelID, "Failed to retrieve template", "TEMPLATE_NOT_FOUND", err.Error())
		}

		// Check template compatibility
		if !tmpl.MatchesType(ch.ChannelType()) {
			channelLogger.Error("Template type mismatch",
				zap.String("template_type", string(tmpl.ChannelType())),
				zap.String("channel_type", string(ch.ChannelType())))
			return s.createFailedResult(channelID, "Template type mismatch", "TYPE_MISMATCH", 
				fmt.Sprintf("Template type: %s, Channel type: %s", tmpl.ChannelType(), ch.ChannelType()))
		}

		channelLogger = channelLogger.WithFields(
			zap.String("template_id", tmpl.ID().String()),
			zap.String("template_name", tmpl.Name().String()))
	}

	// Prepare render request
	renderRequest := s.prepareRenderRequestEnhanced(ch, tmpl, variables, channelOverrides)

	// Validate variables if template is used
	if tmpl != nil {
		if err := s.validateVariables(tmpl, renderRequest.Variables); err != nil {
			channelLogger.Warn("Variable validation failed", zap.Error(err))
			return s.createFailedResult(channelID, "Variable validation failed", "MISSING_VARIABLES", err.Error())
		}
	}

	// Render template
	renderedContent, err := s.renderer.Render(ctx, renderRequest)
	if err != nil {
		channelLogger.Error("Template rendering failed", zap.Error(err))
		return s.createFailedResult(channelID, "Template rendering failed", "RENDER_ERROR", err.Error())
	}

	channelLogger.Debug("Template rendered successfully",
		zap.Int("subject_length", len(renderedContent.Subject)),
		zap.Int("content_length", len(renderedContent.Content)))

	// Send message via external service
	sendRequest := &SendRequest{
		Channel:   ch,
		Content:   renderedContent,
		Variables: variables.ToMap(),
	}

	sendResult := s.notificationService.SendSingleNotification(ctx, sendRequest)
	
	if !sendResult.Success {
		channelLogger.Error("Message sending failed",
			zap.Error(sendResult.Error),
			zap.Any("details", sendResult.Details))
		
		errorCode := "SEND_ERROR"
		errorDetails := "Failed to send message"
		if sendResult.Error != nil {
			errorDetails = sendResult.Error.Error()
		}
		
		return s.createFailedResult(channelID, sendResult.Message, errorCode, errorDetails)
	}

	channelLogger.Info("Message sent successfully",
		zap.String("result_message", sendResult.Message),
		zap.Any("details", sendResult.Details))

	// Mark channel as used
	ch.MarkAsUsed()
	if err := s.channelRepo.Update(ctx, ch); err != nil {
		channelLogger.Warn("Failed to update channel last used time", zap.Error(err))
		// This is not a critical error, so we don't fail the operation
	}

	// Create success result
	result, err := message.NewSuccessfulMessageResult(channelID, sendResult.Message)
	if err != nil {
		channelLogger.Error("Failed to create success result", zap.Error(err))
		return s.createFailedResult(channelID, "Failed to create result", "RESULT_ERROR", err.Error())
	}

	return result
}

// prepareRenderRequestEnhanced prepares render request with enhanced override handling
func (s *EnhancedMessageSender) prepareRenderRequestEnhanced(
	ch *channel.Channel,
	tmpl *template.Template,
	variables *message.Variables,
	channelOverrides *message.ChannelOverrides,
) *RenderRequest {
	request := &RenderRequest{
		Variables: variables,
	}

	// Set default subject and content
	if tmpl != nil {
		request.Subject = tmpl.Subject()
		request.Content = tmpl.Content()
	} else {
		// Use empty subject and content if no template
		defaultSubject, _ := template.NewSubject("")
		defaultContent, _ := template.NewTemplateContent("Default message content")
		request.Subject = defaultSubject
		request.Content = defaultContent
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

// validateVariables validates template variables
func (s *EnhancedMessageSender) validateVariables(tmpl *template.Template, variables *message.Variables) error {
	missingVariables := tmpl.ValidateVariables(variables.ToMap())
	if len(missingVariables) > 0 {
		return fmt.Errorf("missing required variables: %v", missingVariables)
	}
	return nil
}

// createFailedResult creates a failed message result
func (s *EnhancedMessageSender) createFailedResult(channelID *channel.ChannelID, msg, code, details string) *message.MessageResult {
	msgError := message.NewMessageError(code, details)
	result, _ := message.NewFailedMessageResult(channelID, msg, msgError)
	return result
}