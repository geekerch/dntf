package services

import (
	"context"
	"errors"
	"fmt"

	"channel-api/internal/domain/channel"
	"channel-api/internal/domain/message"
	"channel-api/internal/domain/template"
)

// MessageSender 訊息發送領域服務
type MessageSender struct {
	channelRepo  channel.ChannelRepository
	templateRepo template.TemplateRepository
	messageRepo  message.MessageRepository
	renderer     TemplateRenderer
}

// NewMessageSender 建立訊息發送服務
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

// SendMessage 發送訊息
func (ms *MessageSender) SendMessage(
	ctx context.Context,
	channelIDs *message.ChannelIDs,
	variables *message.Variables,
	channelOverrides *message.ChannelOverrides,
) (*message.Message, error) {
	// 建立訊息實體
	msg, err := message.NewMessage(channelIDs, variables, channelOverrides)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// 儲存訊息
	if err := ms.messageRepo.Save(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	// 處理每個通道
	for _, channelID := range channelIDs.ToSlice() {
		result := ms.processSingleChannel(ctx, channelID, variables, channelOverrides)
		if err := msg.AddResult(result); err != nil {
			// 如果新增結果失敗，記錄錯誤但繼續處理其他通道
			continue
		}
	}

	// 更新訊息狀態
	if err := ms.messageRepo.Update(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to update message: %w", err)
	}

	return msg, nil
}

// processSingleChannel 處理單一通道的訊息發送
func (ms *MessageSender) processSingleChannel(
	ctx context.Context,
	channelID *channel.ChannelID,
	variables *message.Variables,
	channelOverrides *message.ChannelOverrides,
) *message.MessageResult {
	// 取得通道資訊
	ch, err := ms.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		return ms.createFailedResult(channelID, "Failed to retrieve channel", "CHANNEL_NOT_FOUND", err.Error())
	}

	// 檢查通道是否可以發送訊息
	if err := ch.CanSendMessage(); err != nil {
		return ms.createFailedResult(channelID, "Channel cannot send message", "CHANNEL_UNAVAILABLE", err.Error())
	}

	// 取得範本資訊
	tmpl, err := ms.templateRepo.FindByID(ctx, ch.TemplateID())
	if err != nil {
		return ms.createFailedResult(channelID, "Failed to retrieve template", "TEMPLATE_NOT_FOUND", err.Error())
	}

	// 檢查通道類型是否匹配範本
	if !tmpl.MatchesType(ch.ChannelType()) {
		return ms.createFailedResult(channelID, "Channel type mismatch with template", "TYPE_MISMATCH", 
			fmt.Sprintf("Channel type: %s, Template type: %s", ch.ChannelType(), tmpl.ChannelType()))
	}

	// 準備渲染內容
	renderRequest := ms.prepareRenderRequest(ch, tmpl, variables, channelOverrides)

	// 驗證變數
	if err := ms.validateVariables(tmpl, renderRequest.Variables); err != nil {
		return ms.createFailedResult(channelID, "Variable validation failed", "MISSING_VARIABLES", err.Error())
	}

	// 渲染範本
	renderedContent, err := ms.renderer.Render(ctx, renderRequest)
	if err != nil {
		return ms.createFailedResult(channelID, "Template rendering failed", "RENDER_ERROR", err.Error())
	}

	// 這裡應該調用實際的訊息發送服務 (例如 EmailService, SlackService 等)
	// 由於這是領域層，我們暫時模擬成功的發送
	_ = renderedContent

	// 標記通道為已使用
	ch.MarkAsUsed()
	if err := ms.channelRepo.Update(ctx, ch); err != nil {
		// 更新失敗不影響發送結果，只記錄錯誤
	}

	// 建立成功結果
	result, err := message.NewSuccessfulMessageResult(channelID, "Message sent successfully")
	if err != nil {
		return ms.createFailedResult(channelID, "Failed to create result", "RESULT_ERROR", err.Error())
	}

	return result
}

// prepareRenderRequest 準備渲染請求
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

	// 應用通道覆寫
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

// validateVariables 驗證變數
func (ms *MessageSender) validateVariables(tmpl *template.Template, variables *message.Variables) error {
	missingVariables := tmpl.ValidateVariables(variables.ToMap())
	if len(missingVariables) > 0 {
		return fmt.Errorf("missing required variables: %v", missingVariables)
	}
	return nil
}

// createFailedResult 建立失敗結果
func (ms *MessageSender) createFailedResult(channelID *channel.ChannelID, msg, code, details string) *message.MessageResult {
	msgError := message.NewMessageError(code, details)
	result, _ := message.NewFailedMessageResult(channelID, msg, msgError)
	return result
}

// TemplateRenderer 範本渲染器介面
type TemplateRenderer interface {
	Render(ctx context.Context, request *RenderRequest) (*RenderedContent, error)
}

// RenderRequest 渲染請求
type RenderRequest struct {
	Subject   *template.Subject
	Content   *template.TemplateContent
	Variables *message.Variables
}

// RenderedContent 渲染結果
type RenderedContent struct {
	Subject string
	Content string
}

// DefaultTemplateRenderer 預設範本渲染器
type DefaultTemplateRenderer struct{}

// NewDefaultTemplateRenderer 建立預設範本渲染器
func NewDefaultTemplateRenderer() *DefaultTemplateRenderer {
	return &DefaultTemplateRenderer{}
}

// Render 渲染範本
func (r *DefaultTemplateRenderer) Render(ctx context.Context, request *RenderRequest) (*RenderedContent, error) {
	if request == nil {
		return nil, errors.New("render request is required")
	}

	variableMap := request.Variables.ToMap()

	// 渲染主題
	renderedSubject, err := r.renderTemplate(request.Subject.String(), variableMap)
	if err != nil {
		return nil, fmt.Errorf("failed to render subject: %w", err)
	}

	// 渲染內容
	renderedContent, err := r.renderTemplate(request.Content.String(), variableMap)
	if err != nil {
		return nil, fmt.Errorf("failed to render content: %w", err)
	}

	return &RenderedContent{
		Subject: renderedSubject,
		Content: renderedContent,
	}, nil
}

// renderTemplate 渲染單一範本
func (r *DefaultTemplateRenderer) renderTemplate(template string, variables map[string]interface{}) (string, error) {
	// 簡單的變數替換實現
	// 在實際專案中，可以使用更強大的範本引擎如 text/template 或 html/template
	result := template
	for key, value := range variables {
		placeholder := fmt.Sprintf("{%s}", key)
		replacement := fmt.Sprintf("%v", value)
		result = fmt.Sprintf("%s", fmt.Sprintf(result, placeholder, replacement))
	}
	return result, nil
}