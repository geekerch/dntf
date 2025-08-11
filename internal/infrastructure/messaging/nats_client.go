package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"channel-api/pkg/config"
	"channel-api/pkg/logger"
)

// NATSClient wraps NATS connection with additional functionality
type NATSClient struct {
	conn   *nats.Conn
	config *config.NATSConfig
	logger *logger.Logger
}

// NewNATSClient creates a new NATS client
func NewNATSClient(cfg *config.NATSConfig, log *logger.Logger) (*NATSClient, error) {
	// Configure NATS options
	opts := []nats.Option{
		nats.Name("channel-api"),
		nats.MaxReconnects(cfg.MaxReconnects),
		nats.ReconnectWait(time.Duration(cfg.ReconnectWait) * time.Second),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Warn("NATS disconnected", zap.Error(err))
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Info("NATS reconnected", zap.String("url", nc.ConnectedUrl()))
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Info("NATS connection closed")
		}),
	}

	// Connect to NATS
	conn, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return &NATSClient{
		conn:   conn,
		config: cfg,
		logger: log,
	}, nil
}

// Close closes the NATS connection
func (c *NATSClient) Close() error {
	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}

// Publish publishes a message to a subject
func (c *NATSClient) Publish(subject string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	fullSubject := c.getFullSubject(subject)
	if err := c.conn.Publish(fullSubject, payload); err != nil {
		c.logger.Error("Failed to publish message",
			zap.String("subject", fullSubject),
			zap.Error(err))
		return fmt.Errorf("failed to publish message: %w", err)
	}

	c.logger.Debug("Message published",
		zap.String("subject", fullSubject),
		zap.Int("payload_size", len(payload)))

	return nil
}

// Subscribe subscribes to a subject with a handler
func (c *NATSClient) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	fullSubject := c.getFullSubject(subject)
	sub, err := c.conn.Subscribe(fullSubject, handler)
	if err != nil {
		c.logger.Error("Failed to subscribe to subject",
			zap.String("subject", fullSubject),
			zap.Error(err))
		return nil, fmt.Errorf("failed to subscribe to subject: %w", err)
	}

	c.logger.Info("Subscribed to subject",
		zap.String("subject", fullSubject))

	return sub, nil
}

// Request sends a request and waits for a response
func (c *NATSClient) Request(subject string, data interface{}, timeout time.Duration) (*nats.Msg, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	fullSubject := c.getFullSubject(subject)
	msg, err := c.conn.Request(fullSubject, payload, timeout)
	if err != nil {
		c.logger.Error("Failed to send request",
			zap.String("subject", fullSubject),
			zap.Error(err))
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	c.logger.Debug("Request sent and response received",
		zap.String("subject", fullSubject),
		zap.Int("response_size", len(msg.Data)))

	return msg, nil
}

// RequestWithContext sends a request with context
func (c *NATSClient) RequestWithContext(ctx context.Context, subject string, data interface{}) (*nats.Msg, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	fullSubject := c.getFullSubject(subject)
	msg, err := c.conn.RequestWithContext(ctx, fullSubject, payload)
	if err != nil {
		c.logger.Error("Failed to send request with context",
			zap.String("subject", fullSubject),
			zap.Error(err))
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	c.logger.Debug("Request with context sent and response received",
		zap.String("subject", fullSubject),
		zap.Int("response_size", len(msg.Data)))

	return msg, nil
}

// QueueSubscribe subscribes to a subject with queue group
func (c *NATSClient) QueueSubscribe(subject, queue string, handler nats.MsgHandler) (*nats.Subscription, error) {
	fullSubject := c.getFullSubject(subject)
	sub, err := c.conn.QueueSubscribe(fullSubject, queue, handler)
	if err != nil {
		c.logger.Error("Failed to queue subscribe to subject",
			zap.String("subject", fullSubject),
			zap.String("queue", queue),
			zap.Error(err))
		return nil, fmt.Errorf("failed to queue subscribe to subject: %w", err)
	}

	c.logger.Info("Queue subscribed to subject",
		zap.String("subject", fullSubject),
		zap.String("queue", queue))

	return sub, nil
}

// IsConnected checks if NATS is connected
func (c *NATSClient) IsConnected() bool {
	return c.conn != nil && c.conn.IsConnected()
}

// GetStats returns NATS connection statistics
func (c *NATSClient) GetStats() nats.Statistics {
	if c.conn == nil {
		return nats.Statistics{}
	}
	return c.conn.Stats()
}

// getFullSubject prepends the subject prefix to the subject
func (c *NATSClient) getFullSubject(subject string) string {
	if c.config.SubjectPrefix == "" {
		return subject
	}
	return c.config.SubjectPrefix + "." + subject
}

// NATSMessage represents a NATS message with metadata
type NATSMessage struct {
	ReqSeqID   string      `json:"reqSeqId,omitempty"`
	RspSeqID   string      `json:"rspSeqId,omitempty"`
	HTTPStatus int         `json:"httpStatus,omitempty"`
	Data       interface{} `json:"data"`
	Error      *NATSError  `json:"error"`
	Timestamp  int64       `json:"timestamp"`
}

// NATSError represents an error in NATS response
type NATSError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewNATSMessage creates a new NATS message
func NewNATSMessage(data interface{}) *NATSMessage {
	return &NATSMessage{
		Data:      data,
		Error:     nil,
		Timestamp: time.Now().UnixMilli(),
	}
}

// NewNATSErrorMessage creates a new NATS error message
func NewNATSErrorMessage(code, message string, httpStatus int) *NATSMessage {
	return &NATSMessage{
		Data: nil,
		Error: &NATSError{
			Code:    code,
			Message: message,
		},
		HTTPStatus: httpStatus,
		Timestamp:  time.Now().UnixMilli(),
	}
}

// SetRequestID sets the request sequence ID
func (m *NATSMessage) SetRequestID(reqSeqID string) *NATSMessage {
	m.ReqSeqID = reqSeqID
	return m
}

// SetResponseID sets the response sequence ID
func (m *NATSMessage) SetResponseID(rspSeqID string) *NATSMessage {
	m.RspSeqID = rspSeqID
	return m
}

// SetHTTPStatus sets the HTTP status code
func (m *NATSMessage) SetHTTPStatus(status int) *NATSMessage {
	m.HTTPStatus = status
	return m
}

// IsError checks if the message contains an error
func (m *NATSMessage) IsError() bool {
	return m.Error != nil
}

// ToJSON converts the message to JSON bytes
func (m *NATSMessage) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON creates a NATS message from JSON bytes
func FromJSON(data []byte) (*NATSMessage, error) {
	var msg NATSMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal NATS message: %w", err)
	}
	return &msg, nil
}