package external

import (
	"context"
	"fmt"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"notification/internal/domain/channel"
	"notification/internal/domain/services"
	"notification/internal/domain/shared"
)

// EmailService implements MessageSender for email channel
type EmailService struct {
	timeout time.Duration
}

// NewEmailService creates a new email service
func NewEmailService(timeout time.Duration) *EmailService {
	return &EmailService{
		timeout: timeout,
	}
}

// Send sends an email through SMTP
func (s *EmailService) Send(ctx context.Context, ch *channel.Channel, content *services.RenderedContent) error {
	// Validate channel type
	if !ch.ChannelType().Equals(shared.ChannelTypeEmail) {
		return fmt.Errorf("invalid channel type for email service: %s", ch.ChannelType().String())
	}

	// Extract SMTP configuration
	config, err := s.extractSMTPConfig(ch.Config())
	if err != nil {
		return fmt.Errorf("failed to extract SMTP config: %w", err)
	}

	// Prepare recipients
	recipients := s.prepareRecipients(ch.Recipients())
	if len(recipients.To) == 0 {
		return fmt.Errorf("no valid email recipients found")
	}

	// Create email message
	message := s.buildEmailMessage(config, recipients, content)

	// Send email with timeout context
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	return s.sendSMTP(ctx, config, recipients.To, message)
}

// GetChannelType returns the supported channel type
func (s *EmailService) GetChannelType() string {
	return shared.ChannelTypeEmail.String()
}

// ValidateConfig validates email channel configuration
func (s *EmailService) ValidateConfig(config *channel.ChannelConfig) error {
	requiredFields := map[string]string{
		"host":     "SMTP host",
		"port":     "SMTP port",
		"username": "SMTP username",
		"password": "SMTP password",
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

	// Validate port
	if port, exists := config.Get("port"); exists {
		switch v := port.(type) {
		case float64:
			if v <= 0 || v > 65535 {
				return fmt.Errorf("invalid port number: %v", v)
			}
		case int:
			if v <= 0 || v > 65535 {
				return fmt.Errorf("invalid port number: %v", v)
			}
		case string:
			if portInt, err := strconv.Atoi(v); err != nil || portInt <= 0 || portInt > 65535 {
				return fmt.Errorf("invalid port number: %s", v)
			}
		default:
			return fmt.Errorf("invalid port type: %T", v)
		}
	}

	return nil
}

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	UseTLS   bool
}

// EmailRecipients holds email recipients
type EmailRecipients struct {
	To  []string
	CC  []string
	BCC []string
}

// extractSMTPConfig extracts SMTP configuration from channel config
func (s *EmailService) extractSMTPConfig(config *channel.ChannelConfig) (*SMTPConfig, error) {
	host, _ := config.Get("host")
	port, _ := config.Get("port")
	username, _ := config.Get("username")
	password, _ := config.Get("password")
	from, _ := config.Get("from")
	useTLS, _ := config.Get("use_tls")

	// Convert port to int
	var portInt int
	switch v := port.(type) {
	case float64:
		portInt = int(v)
	case int:
		portInt = v
	case string:
		var err error
		portInt, err = strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid port format: %s", v)
		}
	default:
		return nil, fmt.Errorf("unsupported port type: %T", v)
	}

	// Set default from address if not provided
	fromStr := ""
	if from != nil {
		fromStr = fmt.Sprintf("%v", from)
	}
	if fromStr == "" {
		fromStr = fmt.Sprintf("%v", username)
	}

	// Convert useTLS to bool
	useTLSBool := false
	if useTLS != nil {
		switch v := useTLS.(type) {
		case bool:
			useTLSBool = v
		case string:
			useTLSBool = strings.ToLower(v) == "true"
		}
	}

	return &SMTPConfig{
		Host:     fmt.Sprintf("%v", host),
		Port:     portInt,
		Username: fmt.Sprintf("%v", username),
		Password: fmt.Sprintf("%v", password),
		From:     fromStr,
		UseTLS:   useTLSBool,
	}, nil
}

// prepareRecipients prepares email recipients from channel recipients
func (s *EmailService) prepareRecipients(recipients *channel.Recipients) *EmailRecipients {
	emailRecipients := &EmailRecipients{
		To:  make([]string, 0),
		CC:  make([]string, 0),
		BCC: make([]string, 0),
	}

	for _, recipient := range recipients.ToSlice() {
		if recipient.Target == "" {
			continue
		}

		switch strings.ToLower(recipient.Type) {
		case "to", "":
			emailRecipients.To = append(emailRecipients.To, recipient.Target)
		case "cc":
			emailRecipients.CC = append(emailRecipients.CC, recipient.Target)
		case "bcc":
			emailRecipients.BCC = append(emailRecipients.BCC, recipient.Target)
		}
	}

	return emailRecipients
}

// buildEmailMessage builds the email message
func (s *EmailService) buildEmailMessage(config *SMTPConfig, recipients *EmailRecipients, content *services.RenderedContent) string {
	var message strings.Builder

	// Headers
	message.WriteString(fmt.Sprintf("From: %s\r\n", config.From))

	if len(recipients.To) > 0 {
		message.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(recipients.To, ", ")))
	}

	if len(recipients.CC) > 0 {
		message.WriteString(fmt.Sprintf("CC: %s\r\n", strings.Join(recipients.CC, ", ")))
	}

	message.WriteString(fmt.Sprintf("Subject: %s\r\n", content.Subject))
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	message.WriteString("\r\n")

	// Body
	message.WriteString(content.Content)

	return message.String()
}

// sendSMTP sends email via SMTP
func (s *EmailService) sendSMTP(ctx context.Context, config *SMTPConfig, recipients []string, message string) error {
	// Create SMTP address
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	// Create auth
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)

	// Combine all recipients (To + CC + BCC)
	allRecipients := make([]string, 0, len(recipients))
	allRecipients = append(allRecipients, recipients...)

	// Send email with context cancellation support
	done := make(chan error, 1)
	go func() {
		err := smtp.SendMail(addr, auth, config.From, allRecipients, []byte(message))
		done <- err
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("email sending cancelled: %w", ctx.Err())
	case err := <-done:
		if err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}
		return nil
	}
}
