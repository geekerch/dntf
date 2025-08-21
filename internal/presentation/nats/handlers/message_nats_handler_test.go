package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"notification/internal/application/message/dtos"
	"notification/internal/application/message/usecases"
	"notification/internal/domain/message"
	"notification/internal/domain/shared"
)

// MockMessageUseCase mocks for message use cases
type MockSendMessageUseCase struct {
	mock.Mock
}

func (m *MockSendMessageUseCase) Execute(ctx context.Context, req *dtos.SendMessageRequest) (*dtos.MessageResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*dtos.MessageResponse), args.Error(1)
}

type MockGetMessageUseCase struct {
	mock.Mock
}

func (m *MockGetMessageUseCase) Execute(ctx context.Context, messageID string) (*dtos.MessageResponse, error) {
	args := m.Called(ctx, messageID)
	return args.Get(0).(*dtos.MessageResponse), args.Error(1)
}

type MockListMessagesUseCase struct {
	mock.Mock
}

func (m *MockListMessagesUseCase) Execute(ctx context.Context, req *dtos.ListMessagesRequest) (*dtos.ListMessagesResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*dtos.ListMessagesResponse), args.Error(1)
}

func TestMessageNATSHandler_SendMessage(t *testing.T) {
	_, nc := setupNATSServer(t)
	
	// Setup mocks
	mockSendUseCase := &MockSendMessageUseCase{}
	mockGetUseCase := &MockGetMessageUseCase{}
	mockListUseCase := &MockListMessagesUseCase{}
	
	// Create handler
	handler := NewMessageNATSHandler(
		mockSendUseCase,
		mockGetUseCase,
		mockListUseCase,
		nc,
	)
	
	// Register handlers
	err := handler.RegisterHandlers()
	require.NoError(t, err)
	
	channelID := uuid.New().String()
	templateID := uuid.New().String()
	
	tests := []struct {
		name           string
		request        dtos.SendMessageRequest
		mockResponse   *dtos.MessageResponse
		mockError      error
		expectedError  bool
		expectedStatus bool
		description    string
	}{
		{
			name: "成功發送 Email 訊息 - 包含 Template 和 Variables",
			request: dtos.SendMessageRequest{
				ChannelIDs: []string{channelID},
				TemplateID: templateID,
				Recipients: []map[string]interface{}{
					{
						"name":   "John Doe",
						"target": "john@example.com",
						"type":   "to",
					},
					{
						"name":   "Jane Smith",
						"target": "jane@example.com",
						"type":   "cc",
					},
				},
				Variables: map[string]interface{}{
					"UserName":    "John Doe",
					"CompanyName": "Test Company",
					"LoginTime":   time.Now().Format("2006-01-02 15:04:05"),
				},
				Settings: &shared.CommonSettings{
					Timeout:       30,
					RetryAttempts: 3,
					RetryDelay:    5,
				},
			},
			mockResponse: &dtos.MessageResponse{
				ID:         uuid.New().String(),
				ChannelID:  channelID,
				TemplateID: templateID,
				Recipients: []map[string]interface{}{
					{
						"name":   "John Doe",
						"target": "john@example.com",
						"type":   "to",
					},
					{
						"name":   "Jane Smith",
						"target": "jane@example.com",
						"type":   "cc",
					},
				},
				Variables: map[string]interface{}{
					"UserName":    "John Doe",
					"CompanyName": "Test Company",
					"LoginTime":   time.Now().Format("2006-01-02 15:04:05"),
				},
				Status:    message.MessageStatusSent,
				CreatedAt: time.Now().UnixMilli(),
				Results: []*dtos.MessageResultResponse{
					{
						Recipient: "john@example.com",
						Status:    message.MessageResultStatusSuccess,
						SentAt:    func() *int64 { t := time.Now().UnixMilli(); return &t }(),
					},
					{
						Recipient: "jane@example.com",
						Status:    message.MessageResultStatusSuccess,
						SentAt:    func() *int64 { t := time.Now().UnixMilli(); return &t }(),
					},
				},
			},
			mockError:      nil,
			expectedError:  false,
			expectedStatus: true,
			description:    "測試包含 template、variables 和多個收件人的 email 發送",
		},
		{
			name: "成功發送訊息 - 包含 BCC 和 TO",
			request: dtos.SendMessageRequest{
				ChannelIDs: []string{channelID},
				TemplateID: templateID,
				Recipients: []map[string]interface{}{
					{
						"name":   "Primary Recipient",
						"target": "primary@example.com",
						"type":   "to",
					},
					{
						"name":   "CC Recipient",
						"target": "cc@example.com",
						"type":   "cc",
					},
					{
						"name":   "BCC Recipient",
						"target": "bcc@example.com",
						"type":   "bcc",
					},
				},
				Variables: map[string]interface{}{
					"Subject": "Important Notification",
					"Content": "This is a test message",
				},
			},
			mockResponse: &dtos.MessageResponse{
				ID:         uuid.New().String(),
				ChannelID:  channelID,
				TemplateID: templateID,
				Recipients: []map[string]interface{}{
					{
						"name":   "Primary Recipient",
						"target": "primary@example.com",
						"type":   "to",
					},
					{
						"name":   "CC Recipient",
						"target": "cc@example.com",
						"type":   "cc",
					},
					{
						"name":   "BCC Recipient",
						"target": "bcc@example.com",
						"type":   "bcc",
					},
				},
				Status:    message.MessageStatusSent,
				CreatedAt: time.Now().UnixMilli(),
			},
			mockError:      nil,
			expectedError:  false,
			expectedStatus: true,
			description:    "測試包含 TO、CC、BCC 收件人的訊息發送",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			if tt.mockError != nil {
				mockSendUseCase.On("Execute", mock.Anything, mock.MatchedBy(func(req *dtos.SendMessageRequest) bool {
					return len(req.ChannelIDs) == len(tt.request.ChannelIDs) && req.TemplateID == tt.request.TemplateID
				})).Return((*dtos.MessageResponse)(nil), tt.mockError).Once()
			} else {
				mockSendUseCase.On("Execute", mock.Anything, mock.MatchedBy(func(req *dtos.SendMessageRequest) bool {
					return len(req.ChannelIDs) == len(tt.request.ChannelIDs) && req.TemplateID == tt.request.TemplateID
				})).Return(tt.mockResponse, nil).Once()
			}
			
			// Create NATS request
			reqSeqId := uuid.New().String()
			natsReq := NATSRequest{
				ReqSeqId:  reqSeqId,
				Data:      tt.request,
				Timestamp: time.Now().UnixMilli(),
			}
			
			reqData, err := json.Marshal(natsReq)
			require.NoError(t, err)
			
			// Send request and wait for response
			msg, err := nc.Request("eco1j.infra.eventcenter.message.send", reqData, 10*time.Second)
			require.NoError(t, err)
			
			// Parse response
			var response NATSResponse
			err = json.Unmarshal(msg.Data, &response)
			require.NoError(t, err)
			
			// Verify response
			assert.Equal(t, reqSeqId, response.ReqSeqId)
			assert.Equal(t, tt.expectedStatus, response.Success)
			
			if tt.expectedError {
				assert.NotNil(t, response.Error)
				assert.Nil(t, response.Data)
			} else {
				assert.Nil(t, response.Error)
				assert.NotNil(t, response.Data)
				
				// Verify response data
				responseData, ok := response.Data.(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, tt.mockResponse.ChannelID, responseData["channelId"])
				assert.Equal(t, tt.mockResponse.TemplateID, responseData["templateId"])
				
				// Verify recipients
				recipients, ok := responseData["recipients"].([]interface{})
				assert.True(t, ok)
				assert.Equal(t, len(tt.mockResponse.Recipients), len(recipients))
			}
			
			// Verify mock was called
			mockSendUseCase.AssertExpectations(t)
		})
	}
}

func TestMessageNATSHandler_SendMessage_ErrorCases(t *testing.T) {
	_, nc := setupNATSServer(t)
	
	// Setup mocks
	mockSendUseCase := &MockSendMessageUseCase{}
	mockGetUseCase := &MockGetMessageUseCase{}
	mockListUseCase := &MockListMessagesUseCase{}
	
	// Create handler
	handler := NewMessageNATSHandler(
		mockSendUseCase,
		mockGetUseCase,
		mockListUseCase,
		nc,
	)
	
	// Register handlers
	err := handler.RegisterHandlers()
	require.NoError(t, err)
	
	channelID := uuid.New().String()
	templateID := uuid.New().String()
	
	tests := []struct {
		name        string
		request     dtos.SendMessageRequest
		mockError   error
		description string
	}{
		{
			name: "發送失敗 - 沒有 Template",
			request: dtos.SendMessageRequest{
				ChannelIDs: []string{channelID},
				TemplateID: "", // Empty template ID
				Recipients: []map[string]interface{}{
					{
						"name":   "Test User",
						"target": "test@example.com",
						"type":   "to",
					},
				},
			},
			mockError:   fmt.Errorf("template ID is required"),
			description: "測試沒有提供 template ID 的錯誤情況",
		},
		{
			name: "發送失敗 - 沒有收件人",
			request: dtos.SendMessageRequest{
				ChannelIDs: []string{channelID},
				TemplateID: templateID,
				Recipients: []map[string]interface{}{}, // Empty recipients
			},
			mockError:   fmt.Errorf("at least one recipient is required"),
			description: "測試沒有收件人的錯誤情況",
		},
		{
			name: "發送失敗 - 無效的 Variables",
			request: dtos.SendMessageRequest{
				ChannelIDs: []string{channelID},
				TemplateID: templateID,
				Recipients: []map[string]interface{}{
					{
						"name":   "Test User",
						"target": "test@example.com",
						"type":   "to",
					},
				},
				Variables: map[string]interface{}{
					"RequiredVar": nil, // Missing required variable
				},
			},
			mockError:   fmt.Errorf("required template variable 'RequiredVar' is missing"),
			description: "測試缺少必要變數的錯誤情況",
		},
		{
			name: "發送失敗 - SMTP 配置錯誤",
			request: dtos.SendMessageRequest{
				ChannelIDs: []string{channelID},
				TemplateID: templateID,
				Recipients: []map[string]interface{}{
					{
						"name":   "Test User",
						"target": "invalid-email",
						"type":   "to",
					},
				},
			},
			mockError:   fmt.Errorf("SMTP configuration error: invalid email address"),
			description: "測試 SMTP 配置錯誤的情況",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			mockSendUseCase.On("Execute", mock.Anything, mock.MatchedBy(func(req *dtos.SendMessageRequest) bool {
				return req.TemplateID == tt.request.TemplateID
			})).Return((*dtos.MessageResponse)(nil), tt.mockError).Once()
			
			// Create NATS request
			reqSeqId := uuid.New().String()
			natsReq := NATSRequest{
				ReqSeqId:  reqSeqId,
				Data:      tt.request,
				Timestamp: time.Now().UnixMilli(),
			}
			
			reqData, err := json.Marshal(natsReq)
			require.NoError(t, err)
			
			// Send request and wait for response
			msg, err := nc.Request("eco1j.infra.eventcenter.message.send", reqData, 5*time.Second)
			require.NoError(t, err)
			
			// Parse response
			var response NATSResponse
			err = json.Unmarshal(msg.Data, &response)
			require.NoError(t, err)
			
			// Verify response
			assert.Equal(t, reqSeqId, response.ReqSeqId)
			assert.False(t, response.Success)
			assert.NotNil(t, response.Error)
			assert.Contains(t, response.Error.Message, "Failed to send message")
			
			// Verify mock was called
			mockSendUseCase.AssertExpectations(t)
		})
	}
}

func TestMessageNATSHandler_GetMessage(t *testing.T) {
	_, nc := setupNATSServer(t)
	
	// Setup mocks
	mockSendUseCase := &MockSendMessageUseCase{}
	mockGetUseCase := &MockGetMessageUseCase{}
	mockListUseCase := &MockListMessagesUseCase{}
	
	// Create handler
	handler := NewMessageNATSHandler(
		mockSendUseCase,
		mockGetUseCase,
		mockListUseCase,
		nc,
	)
	
	// Register handlers
	err := handler.RegisterHandlers()
	require.NoError(t, err)
	
	messageID := uuid.New().String()
	
	tests := []struct {
		name           string
		messageID      string
		mockResponse   *dtos.MessageResponse
		mockError      error
		expectedError  bool
		expectedStatus bool
	}{
		{
			name:      "成功獲取 Message",
			messageID: messageID,
			mockResponse: &dtos.MessageResponse{
				ID:         messageID,
				ChannelID:  uuid.New().String(),
				TemplateID: uuid.New().String(),
				Recipients: []map[string]interface{}{
					{
						"name":   "Test User",
						"target": "test@example.com",
						"type":   "to",
					},
				},
				Variables: map[string]interface{}{
					"UserName": "Test User",
				},
				Status:    message.MessageStatusSent,
				CreatedAt: time.Now().UnixMilli(),
			},
			mockError:      nil,
			expectedError:  false,
			expectedStatus: true,
		},
		{
			name:           "Message 不存在",
			messageID:      "non-existent-id",
			mockResponse:   nil,
			mockError:      fmt.Errorf("message not found"),
			expectedError:  true,
			expectedStatus: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			if tt.mockError != nil {
				mockGetUseCase.On("Execute", mock.Anything, tt.messageID).Return((*dtos.MessageResponse)(nil), tt.mockError).Once()
			} else {
				mockGetUseCase.On("Execute", mock.Anything, tt.messageID).Return(tt.mockResponse, nil).Once()
			}
			
			// Create NATS request
			reqSeqId := uuid.New().String()
			natsReq := NATSRequest{
				ReqSeqId:  reqSeqId,
				Data:      map[string]interface{}{"messageId": tt.messageID},
				Timestamp: time.Now().UnixMilli(),
			}
			
			reqData, err := json.Marshal(natsReq)
			require.NoError(t, err)
			
			// Send request and wait for response
			msg, err := nc.Request("eco1j.infra.eventcenter.message.get", reqData, 5*time.Second)
			require.NoError(t, err)
			
			// Parse response
			var response NATSResponse
			err = json.Unmarshal(msg.Data, &response)
			require.NoError(t, err)
			
			// Verify response
			assert.Equal(t, reqSeqId, response.ReqSeqId)
			assert.Equal(t, tt.expectedStatus, response.Success)
			
			if tt.expectedError {
				assert.NotNil(t, response.Error)
			} else {
				assert.Nil(t, response.Error)
				assert.NotNil(t, response.Data)
			}
			
			// Verify mock was called
			mockGetUseCase.AssertExpectations(t)
		})
	}
}

func TestMessageNATSHandler_ListMessages(t *testing.T) {
	_, nc := setupNATSServer(t)
	
	// Setup mocks
	mockSendUseCase := &MockSendMessageUseCase{}
	mockGetUseCase := &MockGetMessageUseCase{}
	mockListUseCase := &MockListMessagesUseCase{}
	
	// Create handler
	handler := NewMessageNATSHandler(
		mockSendUseCase,
		mockGetUseCase,
		mockListUseCase,
		nc,
	)
	
	// Register handlers
	err := handler.RegisterHandlers()
	require.NoError(t, err)
	
	channelID := uuid.New().String()
	
	// Test successful list
	t.Run("成功列出 Messages", func(t *testing.T) {
		listReq := dtos.ListMessagesRequest{
			ChannelID:      channelID,
			Status:         "sent",
			SkipCount:      0,
			MaxResultCount: 10,
		}
		
		mockResponse := &dtos.ListMessagesResponse{
			Items: []*dtos.MessageResponse{
				{
					ID:         uuid.New().String(),
					ChannelID:  channelID,
					TemplateID: uuid.New().String(),
					Status:     message.MessageStatusSent,
					CreatedAt:  time.Now().UnixMilli(),
				},
				{
					ID:         uuid.New().String(),
					ChannelID:  channelID,
					TemplateID: uuid.New().String(),
					Status:     message.MessageStatusSent,
					CreatedAt:  time.Now().UnixMilli(),
				},
			},
			SkipCount:      0,
			MaxResultCount: 10,
			TotalCount:     2,
			HasMore:        false,
		}
		
		// Setup mock expectations
		mockListUseCase.On("Execute", mock.Anything, mock.MatchedBy(func(req *dtos.ListMessagesRequest) bool {
			return req.ChannelID == listReq.ChannelID
		})).Return(mockResponse, nil).Once()
		
		// Create NATS request
		reqSeqId := uuid.New().String()
		natsReq := NATSRequest{
			ReqSeqId:  reqSeqId,
			Data:      listReq,
			Timestamp: time.Now().UnixMilli(),
		}
		
		reqData, err := json.Marshal(natsReq)
		require.NoError(t, err)
		
		// Send request and wait for response
		msg, err := nc.Request("eco1j.infra.eventcenter.message.list", reqData, 5*time.Second)
		require.NoError(t, err)
		
		// Parse response
		var response NATSResponse
		err = json.Unmarshal(msg.Data, &response)
		require.NoError(t, err)
		
		// Verify response
		assert.Equal(t, reqSeqId, response.ReqSeqId)
		assert.True(t, response.Success)
		assert.Nil(t, response.Error)
		assert.NotNil(t, response.Data)
		
		// Verify mock was called
		mockListUseCase.AssertExpectations(t)
	})
}