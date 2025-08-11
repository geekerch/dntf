package message

import (
	"errors"
	"time"

	"channel-api/internal/domain/channel"
)

// Message is the aggregate root for messages.
type Message struct {
	id               *MessageID
	channelIDs       *ChannelIDs
	variables        *Variables
	channelOverrides *ChannelOverrides
	status           MessageStatus
	results          []*MessageResult
	createdAt        int64
}

// NewMessage creates a new message.
func NewMessage(
	channelIDs *ChannelIDs,
	variables *Variables,
	channelOverrides *ChannelOverrides,
) (*Message, error) {
	// Validate required fields
	if channelIDs == nil {
		return nil, errors.New("channel IDs is required")
	}
	if variables == nil {
		return nil, errors.New("variables is required")
	}

	// Set default values
	if channelOverrides == nil {
		channelOverrides = NewChannelOverrides(nil)
	}

	return &Message{
		id:               NewMessageID(),
		channelIDs:       channelIDs,
		variables:        variables,
		channelOverrides: channelOverrides,
		status:           MessageStatusPending,
		results:          make([]*MessageResult, 0),
		createdAt:        time.Now().UnixMilli(),
	}, nil
}

// ReconstructMessage reconstructs a message from persistent data.
func ReconstructMessage(
	id *MessageID,
	channelIDs *ChannelIDs,
	variables *Variables,
	channelOverrides *ChannelOverrides,
	status MessageStatus,
	results []*MessageResult,
	createdAt int64,
) *Message {
	return &Message{
		id:               id,
		channelIDs:       channelIDs,
		variables:        variables,
		channelOverrides: channelOverrides,
		status:           status,
		results:          results,
		createdAt:        createdAt,
	}
}

// ID gets the message ID.
func (m *Message) ID() *MessageID {
	return m.id
}

// ChannelIDs gets the list of channel IDs.
func (m *Message) ChannelIDs() *ChannelIDs {
	return m.channelIDs
}

// Variables gets the template variables.
func (m *Message) Variables() *Variables {
	return m.variables
}

// ChannelOverrides gets the channel override settings.
func (m *Message) ChannelOverrides() *ChannelOverrides {
	return m.channelOverrides
}

// Status gets the message status.
func (m *Message) Status() MessageStatus {
	return m.status
}

// Results gets the message results.
func (m *Message) Results() []*MessageResult {
	return m.results
}

// CreatedAt gets the creation time.
func (m *Message) CreatedAt() int64 {
	return m.createdAt
}

// AddResult adds a message result.
func (m *Message) AddResult(result *MessageResult) error {
	if result == nil {
		return errors.New("message result cannot be nil")
	}
	
	// Check if a result for the same channel already exists
	for _, existingResult := range m.results {
		if existingResult.ChannelID().Equals(result.ChannelID()) {
			return errors.New("result for this channel already exists")
		}
	}
	
	m.results = append(m.results, result)
	m.updateStatus()
	return nil
}

// UpdateResult updates the message result for the specified channel.
func (m *Message) UpdateResult(channelID *channel.ChannelID, result *MessageResult) error {
	if channelID == nil {
		return errors.New("channel ID cannot be nil")
	}
	if result == nil {
		return errors.New("message result cannot be nil")
	}
	
	for i, existingResult := range m.results {
		if existingResult.ChannelID().Equals(channelID) {
			m.results[i] = result
			m.updateStatus()
			return nil
		}
	}
	
	return errors.New("result for this channel not found")
}

// GetResult gets the message result for the specified channel.
func (m *Message) GetResult(channelID *channel.ChannelID) (*MessageResult, bool) {
	if channelID == nil {
		return nil, false
	}
	
	for _, result := range m.results {
		if result.ChannelID().Equals(channelID) {
			return result, true
		}
	}
	
	return nil, false
}

// IsCompleted checks if the message has been processed completely.
func (m *Message) IsCompleted() bool {
	return len(m.results) == m.channelIDs.Count()
}

// GetSuccessfulResults gets the successful results.
func (m *Message) GetSuccessfulResults() []*MessageResult {
	successful := make([]*MessageResult, 0)
	for _, result := range m.results {
		if result.IsSuccess() {
			successful = append(successful, result)
		}
	}
	return successful
}

// GetFailedResults gets the failed results.
func (m *Message) GetFailedResults() []*MessageResult {
	failed := make([]*MessageResult, 0)
	for _, result := range m.results {
		if result.IsFailed() {
			failed = append(failed, result)
		}
	}
	return failed
}

// updateStatus updates the message status based on the results.
func (m *Message) updateStatus() {
	if !m.IsCompleted() {
		m.status = MessageStatusPending
		return
	}
	
	successCount := len(m.GetSuccessfulResults())
	totalCount := len(m.results)
	
	if successCount == totalCount {
		m.status = MessageStatusSuccess
	} else if successCount == 0 {
		m.status = MessageStatusFailed
	} else {
		m.status = MessageStatusPartialSuccess
	}
}

// MessageResult is the result of a message.
type MessageResult struct {
	channelID *channel.ChannelID
	status    MessageResultStatus
	message   string
	error     *MessageError
	sentAt    *int64
}

// MessageResultStatus is the status of a message result.
type MessageResultStatus string

const (
	MessageResultStatusSuccess MessageResultStatus = "success"
	MessageResultStatusFailed  MessageResultStatus = "failed"
)

// NewSuccessfulMessageResult creates a successful message result.
func NewSuccessfulMessageResult(channelID *channel.ChannelID, message string) (*MessageResult, error) {
	if channelID == nil {
		return nil, errors.New("channel ID is required")
	}
	if message == "" {
		return nil, errors.New("message is required")
	}
	
	now := time.Now().UnixMilli()
	return &MessageResult{
		channelID: channelID,
		status:    MessageResultStatusSuccess,
		message:   message,
		error:     nil,
		sentAt:    &now,
	}, nil
}

// NewFailedMessageResult creates a failed message result.
func NewFailedMessageResult(channelID *channel.ChannelID, message string, err *MessageError) (*MessageResult, error) {
	if channelID == nil {
		return nil, errors.New("channel ID is required")
	}
	if message == "" {
		return nil, errors.New("message is required")
	}
	if err == nil {
		return nil, errors.New("error is required for failed result")
	}
	
	return &MessageResult{
		channelID: channelID,
		status:    MessageResultStatusFailed,
		message:   message,
		error:     err,
		sentAt:    nil,
	}, nil
}

// ChannelID gets the channel ID.
func (mr *MessageResult) ChannelID() *channel.ChannelID {
	return mr.channelID
}

// Status gets the result status.
func (mr *MessageResult) Status() MessageResultStatus {
	return mr.status
}

// Message gets the message.
func (mr *MessageResult) Message() string {
	return mr.message
}

// Error gets the error.
func (mr *MessageResult) Error() *MessageError {
	return mr.error
}

// SentAt gets the sending time.
func (mr *MessageResult) SentAt() *int64 {
	return mr.sentAt
}

// IsSuccess checks if it is successful.
func (mr *MessageResult) IsSuccess() bool {
	return mr.status == MessageResultStatusSuccess
}

// IsFailed checks if it has failed.
func (mr *MessageResult) IsFailed() bool {
	return mr.status == MessageResultStatusFailed
}

// MessageError is a message error.
type MessageError struct {
	Code    string `json:"code"`
	Details string `json:"details"`
}

// NewMessageError creates a message error.
func NewMessageError(code, details string) *MessageError {
	return &MessageError{
		Code:    code,
		Details: details,
	}
}