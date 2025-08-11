package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"channel-api/internal/domain/channel"
	"channel-api/internal/domain/services"
	"channel-api/internal/domain/shared"
)

// SlackService implements MessageSender for Slack channel
type SlackService struct {
	httpClient *http.Client
	timeout    time.Duration
}

// NewSlackService creates a new Slack service
func NewSlackService(timeout time.Duration) *SlackService {
	return &SlackService{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

// Send sends a message to Slack
func (s *SlackService) Send(ctx context.Context, ch *channel.Channel, content *services.RenderedContent) error {
	// Validate channel type
	if ch.ChannelType() != shared.ChannelTypeSlack {
		return fmt.Errorf("invalid channel type for Slack service: %s", ch.ChannelType())
	}

	// Extract Slack configuration
	config, err := s.extractSlackConfig(ch.Config())
	if err != nil {
		return fmt.Errorf("failed to extract Slack config: %w", err)
	}

	// Prepare recipients
	targets := s.prepareTargets(ch.Recipients())
	if len(targets) == 0 {
		return fmt.Errorf("no valid Slack targets found")
	}

	// Send to all targets
	for _, target := range targets {
		if err := s.sendToTarget(ctx, config, target, content); err != nil {
			return fmt.Errorf("failed to send to target %s: %w", target, err)
		}
	}

	return nil
}

// GetChannelType returns the supported channel type
func (s *SlackService) GetChannelType() string {
	return string(shared.ChannelTypeSlack)
}

// ValidateConfig validates Slack channel configuration
func (s *SlackService) ValidateConfig(config *channel.ChannelConfig) error {
	requiredFields := map[string]string{
		"token":     "Slack bot token",
		"workspace": "Slack workspace",
	}

	for field, description := range requiredFields {
		value, exists := config.Get(field)
		if !exists {
			return fmt.Errorf("missing required field: %s (%s)", field, description)
		}
		if value == nil || value == "" {
			return fmt.Errorf("empty required field: %s (%s)", field, description)
		}
	}

	// Validate token format
	if token, exists := config.Get("token"); exists {
		tokenStr := fmt.Sprintf("%v", token)
		if !strings.HasPrefix(tokenStr, "xoxb-") && !strings.HasPrefix(tokenStr, "xoxp-") {
			return fmt.Errorf("invalid Slack token format")
		}
	}

	return nil
}

// SlackConfig holds Slack configuration
type SlackConfig struct {
	Token     string
	Workspace string
	WebhookURL string // Optional webhook URL
}

// SlackMessage represents a Slack message payload
type SlackMessage struct {
	Channel     string            `json:"channel"`
	Text        string            `json:"text"`
	Username    string            `json:"username,omitempty"`
	IconEmoji   string            `json:"icon_emoji,omitempty"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

// SlackAttachment represents a Slack message attachment
type SlackAttachment struct {
	Color     string `json:"color,omitempty"`
	Title     string `json:"title,omitempty"`
	Text      string `json:"text,omitempty"`
	Footer    string `json:"footer,omitempty"`
	Timestamp int64  `json:"ts,omitempty"`
}

// SlackResponse represents the response from Slack API
type SlackResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
	TS    string `json:"ts,omitempty"`
}

// extractSlackConfig extracts Slack configuration from channel config
func (s *SlackService) extractSlackConfig(config *channel.ChannelConfig) (*SlackConfig, error) {
	token, _ := config.Get("token")
	workspace, _ := config.Get("workspace")
	webhookURL, _ := config.Get("webhook_url")

	slackConfig := &SlackConfig{
		Token:     fmt.Sprintf("%v", token),
		Workspace: fmt.Sprintf("%v", workspace),
	}

	if webhookURL != nil {
		slackConfig.WebhookURL = fmt.Sprintf("%v", webhookURL)
	}

	return slackConfig, nil
}

// prepareTargets prepares Slack targets from channel recipients
func (s *SlackService) prepareTargets(recipients *channel.Recipients) []string {
	targets := make([]string, 0)

	for _, recipient := range recipients.ToSlice() {
		if recipient.Target == "" {
			continue
		}

		// Support different target types: channel, user, DM
		target := recipient.Target
		switch strings.ToLower(recipient.Type) {
		case "channel":
			// Ensure channel starts with #
			if !strings.HasPrefix(target, "#") {
				target = "#" + target
			}
		case "user", "dm":
			// Ensure user starts with @
			if !strings.HasPrefix(target, "@") {
				target = "@" + target
			}
		default:
			// Default to channel if type is not specified
			if !strings.HasPrefix(target, "#") && !strings.HasPrefix(target, "@") {
				target = "#" + target
			}
		}

		targets = append(targets, target)
	}

	return targets
}

// sendToTarget sends message to a specific Slack target
func (s *SlackService) sendToTarget(ctx context.Context, config *SlackConfig, target string, content *services.RenderedContent) error {
	// Use webhook if available, otherwise use API
	if config.WebhookURL != "" {
		return s.sendViaWebhook(ctx, config.WebhookURL, target, content)
	}
	return s.sendViaAPI(ctx, config.Token, target, content)
}

// sendViaWebhook sends message via Slack webhook
func (s *SlackService) sendViaWebhook(ctx context.Context, webhookURL, target string, content *services.RenderedContent) error {
	message := SlackMessage{
		Channel: target,
		Text:    content.Content,
	}

	// Add subject as attachment if present
	if content.Subject != "" {
		message.Attachments = []SlackAttachment{
			{
				Color: "good",
				Title: content.Subject,
				Text:  content.Content,
			},
		}
		message.Text = content.Subject
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("webhook request failed with status: %d", resp.StatusCode)
	}

	return nil
}

// sendViaAPI sends message via Slack Web API
func (s *SlackService) sendViaAPI(ctx context.Context, token, target string, content *services.RenderedContent) error {
	message := SlackMessage{
		Channel: target,
		Text:    content.Content,
	}

	// Add subject as attachment if present
	if content.Subject != "" {
		message.Attachments = []SlackAttachment{
			{
				Color:     "good",
				Title:     content.Subject,
				Text:      content.Content,
				Timestamp: time.Now().Unix(),
			},
		}
		message.Text = content.Subject
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://slack.com/api/chat.postMessage", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send API request: %w", err)
	}
	defer resp.Body.Close()

	var slackResp SlackResponse
	if err := json.NewDecoder(resp.Body).Decode(&slackResp); err != nil {
		return fmt.Errorf("failed to decode Slack response: %w", err)
	}

	if !slackResp.OK {
		return fmt.Errorf("Slack API error: %s", slackResp.Error)
	}

	return nil
}