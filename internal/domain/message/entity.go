package message

import (
	"errors"
	"time"

	"channel-api/internal/domain/channel"
)

// Message 訊息聚合根
type Message struct {
	id               *MessageID
	channelIDs       *ChannelIDs
	variables        *Variables
	channelOverrides *ChannelOverrides
	status           MessageStatus
	results          []*MessageResult
	createdAt        int64
}

// NewMessage 建立新訊息
func NewMessage(
	channelIDs *ChannelIDs,
	variables *Variables,
	channelOverrides *ChannelOverrides,
) (*Message, error) {
	// 驗證必要欄位
	if channelIDs == nil {
		return nil, errors.New("channel IDs is required")
	}
	if variables == nil {
		return nil, errors.New("variables is required")
	}

	// 設定預設值
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

// ReconstructMessage 從持久化資料重建訊息
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

// ID 取得訊息 ID
func (m *Message) ID() *MessageID {
	return m.id
}

// ChannelIDs 取得通道 ID 列表
func (m *Message) ChannelIDs() *ChannelIDs {
	return m.channelIDs
}

// Variables 取得範本變數
func (m *Message) Variables() *Variables {
	return m.variables
}

// ChannelOverrides 取得通道覆寫設定
func (m *Message) ChannelOverrides() *ChannelOverrides {
	return m.channelOverrides
}

// Status 取得訊息狀態
func (m *Message) Status() MessageStatus {
	return m.status
}

// Results 取得訊息結果
func (m *Message) Results() []*MessageResult {
	return m.results
}

// CreatedAt 取得建立時間
func (m *Message) CreatedAt() int64 {
	return m.createdAt
}

// AddResult 新增訊息結果
func (m *Message) AddResult(result *MessageResult) error {
	if result == nil {
		return errors.New("message result cannot be nil")
	}
	
	// 檢查是否已存在相同通道的結果
	for _, existingResult := range m.results {
		if existingResult.ChannelID().Equals(result.ChannelID()) {
			return errors.New("result for this channel already exists")
		}
	}
	
	m.results = append(m.results, result)
	m.updateStatus()
	return nil
}

// UpdateResult 更新指定通道的訊息結果
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

// GetResult 取得指定通道的訊息結果
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

// IsCompleted 檢查訊息是否已完成處理
func (m *Message) IsCompleted() bool {
	return len(m.results) == m.channelIDs.Count()
}

// GetSuccessfulResults 取得成功的結果
func (m *Message) GetSuccessfulResults() []*MessageResult {
	successful := make([]*MessageResult, 0)
	for _, result := range m.results {
		if result.IsSuccess() {
			successful = append(successful, result)
		}
	}
	return successful
}

// GetFailedResults 取得失敗的結果
func (m *Message) GetFailedResults() []*MessageResult {
	failed := make([]*MessageResult, 0)
	for _, result := range m.results {
		if result.IsFailed() {
			failed = append(failed, result)
		}
	}
	return failed
}

// updateStatus 根據結果更新訊息狀態
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

// MessageResult 訊息結果
type MessageResult struct {
	channelID *channel.ChannelID
	status    MessageResultStatus
	message   string
	error     *MessageError
	sentAt    *int64
}

// MessageResultStatus 訊息結果狀態
type MessageResultStatus string

const (
	MessageResultStatusSuccess MessageResultStatus = "success"
	MessageResultStatusFailed  MessageResultStatus = "failed"
)

// NewSuccessfulMessageResult 建立成功的訊息結果
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

// NewFailedMessageResult 建立失敗的訊息結果
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

// ChannelID 取得通道 ID
func (mr *MessageResult) ChannelID() *channel.ChannelID {
	return mr.channelID
}

// Status 取得結果狀態
func (mr *MessageResult) Status() MessageResultStatus {
	return mr.status
}

// Message 取得訊息
func (mr *MessageResult) Message() string {
	return mr.message
}

// Error 取得錯誤
func (mr *MessageResult) Error() *MessageError {
	return mr.error
}

// SentAt 取得發送時間
func (mr *MessageResult) SentAt() *int64 {
	return mr.sentAt
}

// IsSuccess 檢查是否成功
func (mr *MessageResult) IsSuccess() bool {
	return mr.status == MessageResultStatusSuccess
}

// IsFailed 檢查是否失敗
func (mr *MessageResult) IsFailed() bool {
	return mr.status == MessageResultStatusFailed
}

// MessageError 訊息錯誤
type MessageError struct {
	Code    string `json:"code"`
	Details string `json:"details"`
}

// NewMessageError 建立訊息錯誤
func NewMessageError(code, details string) *MessageError {
	return &MessageError{
		Code:    code,
		Details: details,
	}
}