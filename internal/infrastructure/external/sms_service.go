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

// SMSService implements MessageSender for SMS channel
type SMSService struct {
	httpClient *http.Client
	timeout    time.Duration
}

// NewSMSService creates a new SMS service
func NewSMSService(timeout time.Duration) *SMSService {
	return &SMSService{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

// Send sends an SMS message
func (s *SMSService) Send(ctx context.Context, ch *channel.Channel, content *services.RenderedContent) error {
	// Validate channel type
	if ch.ChannelType() != shared.ChannelTypeSMS {
		return fmt.Errorf("invalid channel type for SMS service: %s", ch.ChannelType())
	}

	// Extract SMS configuration
	config, err := s.extractSMSConfig(ch.Config())
	if err != nil {
		return fmt.Errorf("failed to extract SMS config: %w", err)
	}

	// Prepare phone numbers
	phoneNumbers := s.preparePhoneNumbers(ch.Recipients())
	if len(phoneNumbers) == 0 {
		return fmt.Errorf("no valid phone numbers found")
	}

	// Send to all phone numbers
	for _, phoneNumber := range phoneNumbers {
		if err := s.sendToPhoneNumber(ctx, config, phoneNumber, content); err != nil {
			return fmt.Errorf("failed to send to phone number %s: %w", phoneNumber, err)
		}
	}

	return nil
}

// GetChannelType returns the supported channel type
func (s *SMSService) GetChannelType() string {
	return string(shared.ChannelTypeSMS)
}

// ValidateConfig validates SMS channel configuration
func (s *SMSService) ValidateConfig(config *channel.ChannelConfig) error {
	requiredFields := map[string]string{
		"provider":   "SMS provider (twilio, aws_sns, etc.)",
		"api_key":    "API key",
		"api_secret": "API secret",
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

	// Validate provider
	if provider, exists := config.Get("provider"); exists {
		providerStr := strings.ToLower(fmt.Sprintf("%v", provider))
		supportedProviders := []string{"twilio", "aws_sns", "nexmo", "messagebird"}
		
		isSupported := false
		for _, supported := range supportedProviders {
			if providerStr == supported {
				isSupported = true
				break
			}
		}
		
		if !isSupported {
			return fmt.Errorf("unsupported SMS provider: %s. Supported providers: %v", providerStr, supportedProviders)
		}
	}

	return nil
}

// SMSConfig holds SMS configuration
type SMSConfig struct {
	Provider  string
	APIKey    string
	APISecret string
	From      string
	BaseURL   string
}

// SMSMessage represents an SMS message payload
type SMSMessage struct {
	From string `json:"from"`
	To   string `json:"to"`
	Body string `json:"body"`
}

// SMSResponse represents the response from SMS provider
type SMSResponse struct {
	Success   bool   `json:"success"`
	MessageID string `json:"message_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// extractSMSConfig extracts SMS configuration from channel config
func (s *SMSService) extractSMSConfig(config *channel.ChannelConfig) (*SMSConfig, error) {
	provider, _ := config.Get("provider")
	apiKey, _ := config.Get("api_key")
	apiSecret, _ := config.Get("api_secret")
	from, _ := config.Get("from")
	baseURL, _ := config.Get("base_url")

	smsConfig := &SMSConfig{
		Provider:  strings.ToLower(fmt.Sprintf("%v", provider)),
		APIKey:    fmt.Sprintf("%v", apiKey),
		APISecret: fmt.Sprintf("%v", apiSecret),
	}

	if from != nil {
		smsConfig.From = fmt.Sprintf("%v", from)
	}

	if baseURL != nil {
		smsConfig.BaseURL = fmt.Sprintf("%v", baseURL)
	} else {
		// Set default base URL based on provider
		smsConfig.BaseURL = s.getDefaultBaseURL(smsConfig.Provider)
	}

	return smsConfig, nil
}

// getDefaultBaseURL returns default base URL for SMS providers
func (s *SMSService) getDefaultBaseURL(provider string) string {
	switch provider {
	case "twilio":
		return "https://api.twilio.com/2010-04-01"
	case "aws_sns":
		return "https://sns.us-east-1.amazonaws.com"
	case "nexmo":
		return "https://rest.nexmo.com"
	case "messagebird":
		return "https://rest.messagebird.com"
	default:
		return ""
	}
}

// preparePhoneNumbers prepares phone numbers from channel recipients
func (s *SMSService) preparePhoneNumbers(recipients *channel.Recipients) []string {
	phoneNumbers := make([]string, 0)

	for _, recipient := range recipients.ToSlice() {
		// Check both Target and Email fields for phone numbers
		phoneNumber := ""
		if recipient.Target != "" {
			phoneNumber = recipient.Target
		} else if recipient.Email != "" && s.isPhoneNumber(recipient.Email) {
			phoneNumber = recipient.Email
		}

		if phoneNumber != "" {
			// Clean and validate phone number
			cleanNumber := s.cleanPhoneNumber(phoneNumber)
			if s.isValidPhoneNumber(cleanNumber) {
				phoneNumbers = append(phoneNumbers, cleanNumber)
			}
		}
	}

	return phoneNumbers
}

// isPhoneNumber checks if a string looks like a phone number
func (s *SMSService) isPhoneNumber(str string) bool {
	// Simple check for phone number patterns
	cleaned := s.cleanPhoneNumber(str)
	return len(cleaned) >= 10 && len(cleaned) <= 15
}

// cleanPhoneNumber removes non-digit characters from phone number
func (s *SMSService) cleanPhoneNumber(phoneNumber string) string {
	var cleaned strings.Builder
	for _, char := range phoneNumber {
		if char >= '0' && char <= '9' || char == '+' {
			cleaned.WriteRune(char)
		}
	}
	return cleaned.String()
}

// isValidPhoneNumber validates phone number format
func (s *SMSService) isValidPhoneNumber(phoneNumber string) bool {
	if len(phoneNumber) < 10 || len(phoneNumber) > 15 {
		return false
	}

	// Check for international format
	if strings.HasPrefix(phoneNumber, "+") {
		return len(phoneNumber) >= 11 && len(phoneNumber) <= 16
	}

	return true
}

// sendToPhoneNumber sends SMS to a specific phone number
func (s *SMSService) sendToPhoneNumber(ctx context.Context, config *SMSConfig, phoneNumber string, content *services.RenderedContent) error {
	// Combine subject and content for SMS
	messageBody := content.Content
	if content.Subject != "" {
		messageBody = content.Subject + "\n\n" + content.Content
	}

	// Truncate message if too long (SMS has character limits)
	if len(messageBody) > 1600 {
		messageBody = messageBody[:1597] + "..."
	}

	switch config.Provider {
	case "twilio":
		return s.sendViaTwilio(ctx, config, phoneNumber, messageBody)
	case "aws_sns":
		return s.sendViaAWSSNS(ctx, config, phoneNumber, messageBody)
	case "nexmo":
		return s.sendViaNexmo(ctx, config, phoneNumber, messageBody)
	case "messagebird":
		return s.sendViaMessageBird(ctx, config, phoneNumber, messageBody)
	default:
		return fmt.Errorf("unsupported SMS provider: %s", config.Provider)
	}
}

// sendViaTwilio sends SMS via Twilio API
func (s *SMSService) sendViaTwilio(ctx context.Context, config *SMSConfig, phoneNumber, message string) error {
	// This is a simplified implementation
	// In production, you would use the official Twilio SDK
	
	payload := map[string]string{
		"From": config.From,
		"To":   phoneNumber,
		"Body": message,
	}

	return s.sendHTTPRequest(ctx, config, payload, "/Accounts/"+config.APIKey+"/Messages.json")
}

// sendViaAWSSNS sends SMS via AWS SNS
func (s *SMSService) sendViaAWSSNS(ctx context.Context, config *SMSConfig, phoneNumber, message string) error {
	// This is a simplified implementation
	// In production, you would use the AWS SDK
	
	payload := map[string]interface{}{
		"PhoneNumber": phoneNumber,
		"Message":     message,
	}

	return s.sendHTTPRequest(ctx, config, payload, "/")
}

// sendViaNexmo sends SMS via Nexmo API
func (s *SMSService) sendViaNexmo(ctx context.Context, config *SMSConfig, phoneNumber, message string) error {
	payload := map[string]interface{}{
		"from": config.From,
		"to":   phoneNumber,
		"text": message,
	}

	return s.sendHTTPRequest(ctx, config, payload, "/sms/json")
}

// sendViaMessageBird sends SMS via MessageBird API
func (s *SMSService) sendViaMessageBird(ctx context.Context, config *SMSConfig, phoneNumber, message string) error {
	payload := map[string]interface{}{
		"originator": config.From,
		"recipients": []string{phoneNumber},
		"body":       message,
	}

	return s.sendHTTPRequest(ctx, config, payload, "/messages")
}

// sendHTTPRequest sends HTTP request to SMS provider
func (s *SMSService) sendHTTPRequest(ctx context.Context, config *SMSConfig, payload interface{}, endpoint string) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal SMS payload: %w", err)
	}

	url := config.BaseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	
	// Set authentication based on provider
	switch config.Provider {
	case "twilio":
		req.SetBasicAuth(config.APIKey, config.APISecret)
	case "messagebird":
		req.Header.Set("Authorization", "AccessKey "+config.APIKey)
	default:
		req.Header.Set("Authorization", "Bearer "+config.APIKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send SMS request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("SMS request failed with status: %d", resp.StatusCode)
	}

	return nil
}