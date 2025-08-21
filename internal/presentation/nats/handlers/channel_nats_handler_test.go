package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"notification/internal/application/channel/dtos"
	"notification/internal/application/channel/usecases"
	"notification/internal/domain/shared"
)

// MockChannelUseCase mocks for channel use cases
type MockCreateChannelUseCase struct {
	mock.Mock
}

func (m *MockCreateChannelUseCase) Execute(ctx context.Context, req *dtos.CreateChannelRequest) (*dtos.ChannelResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*dtos.ChannelResponse), args.Error(1)
}

type MockGetChannelUseCase struct {
	mock.Mock
}

func (m *MockGetChannelUseCase) Execute(ctx context.Context, channelID string) (*dtos.ChannelResponse, error) {
	args := m.Called(ctx, channelID)
	return args.Get(0).(*dtos.ChannelResponse), args.Error(1)
}

type MockListChannelsUseCase struct {
	mock.Mock
}

func (m *MockListChannelsUseCase) Execute(ctx context.Context, req *dtos.ListChannelsRequest) (*dtos.ListChannelsResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*dtos.ListChannelsResponse), args.Error(1)
}

type MockUpdateChannelUseCase struct {
	mock.Mock
}

func (m *MockUpdateChannelUseCase) Execute(ctx context.Context, channelID string, req *dtos.UpdateChannelRequest) (*dtos.ChannelResponse, error) {
	args := m.Called(ctx, channelID, req)
	return args.Get(0).(*dtos.ChannelResponse), args.Error(1)
}

type MockDeleteChannelUseCase struct {
	mock.Mock
}

func (m *MockDeleteChannelUseCase) Execute(ctx context.Context, channelID string) (*dtos.DeleteChannelResponse, error) {
	args := m.Called(ctx, channelID)
	return args.Get(0).(*dtos.DeleteChannelResponse), args.Error(1)
}

// Test setup helpers
func setupNATSServer(t *testing.T) (*server.Server, *nats.Conn) {
	opts := &server.Options{
		Host: "127.0.0.1",
		Port: -1, // Use random port
	}
	
	ns, err := server.NewServer(opts)
	require.NoError(t, err)
	
	go ns.Start()
	
	if !ns.ReadyForConnections(5 * time.Second) {
		t.Fatal("NATS server not ready")
	}
	
	nc, err := nats.Connect(ns.ClientURL())
	require.NoError(t, err)
	
	t.Cleanup(func() {
		nc.Close()
		ns.Shutdown()
	})
	
	return ns, nc
}

func setupOldSystemMockServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()
	
	// Mock GET /v2.0/Groups endpoint for verification
	mux.HandleFunc("/v2.0/Groups", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			response := map[string]interface{}{
				"groups": []map[string]interface{}{},
				"total":  0,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	})
	
	// Mock POST /v2/groups endpoint for channel creation
	mux.HandleFunc("/v2/groups", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			response := map[string]interface{}{
				"id":      uuid.New().String(),
				"message": "Group created successfully",
				"status":  "success",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(response)
		}
	})
	
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	
	return server
}

func TestChannelNATSHandler_CreateChannel(t *testing.T) {
	_, nc := setupNATSServer(t)
	oldSystemServer := setupOldSystemMockServer(t)
	
	// Setup mocks
	mockCreateUseCase := &MockCreateChannelUseCase{}
	mockGetUseCase := &MockGetChannelUseCase{}
	mockListUseCase := &MockListChannelsUseCase{}
	mockUpdateUseCase := &MockUpdateChannelUseCase{}
	mockDeleteUseCase := &MockDeleteChannelUseCase{}
	
	// Create handler
	handler := NewChannelNATSHandler(
		mockCreateUseCase,
		mockGetUseCase,
		mockListUseCase,
		mockUpdateUseCase,
		mockDeleteUseCase,
		nc,
	)
	
	// Register handlers
	err := handler.RegisterHandlers()
	require.NoError(t, err)
	
	tests := []struct {
		name           string
		request        dtos.CreateChannelRequest
		mockResponse   *dtos.ChannelResponse
		mockError      error
		expectedError  bool
		expectedStatus bool
	}{
		{
			name: "成功創建 Email Channel",
			request: dtos.CreateChannelRequest{
				ChannelName: "Test Email Channel",
				Description: "Test email channel for notifications",
				Enabled:     true,
				ChannelType: "email",
				TemplateID:  uuid.New().String(),
				CommonSettings: dtos.CommonSettingsDTO{
					Timeout:       30,
					RetryAttempts: 3,
					RetryDelay:    5,
				},
				Config: map[string]interface{}{
					"host":        "smtp.gmail.com",
					"port":        465,
					"secure":      true,
					"method":      "ssl",
					"username":    "test@gmail.com",
					"password":    "testpassword",
					"senderEmail": "test@gmail.com",
				},
				Recipients: []dtos.RecipientDTO{
					{
						Name:   "Test User",
						Target: "test@example.com",
						Type:   "to",
					},
				},
				Tags: []string{"test", "email"},
			},
			mockResponse: &dtos.ChannelResponse{
				ChannelID:   uuid.New().String(),
				ChannelName: "Test Email Channel",
				Description: "Test email channel for notifications",
				Enabled:     true,
				ChannelType: "email",
				CommonSettings: dtos.CommonSettingsDTO{
					Timeout:       30,
					RetryAttempts: 3,
					RetryDelay:    5,
				},
				Config: map[string]interface{}{
					"host":        "smtp.gmail.com",
					"port":        465,
					"secure":      true,
					"method":      "ssl",
					"username":    "test@gmail.com",
					"password":    "testpassword",
					"senderEmail": "test@gmail.com",
				},
				Recipients: []dtos.RecipientDTO{
					{
						Name:   "Test User",
						Target: "test@example.com",
						Type:   "to",
					},
				},
				Tags:      []string{"test", "email"},
				CreatedAt: time.Now().UnixMilli(),
				UpdatedAt: time.Now().UnixMilli(),
			},
			mockError:      nil,
			expectedError:  false,
			expectedStatus: true,
		},
		{
			name: "創建失敗 - 無效配置",
			request: dtos.CreateChannelRequest{
				ChannelName: "Invalid Channel",
				Description: "Channel with invalid config",
				Enabled:     true,
				ChannelType: "email",
				CommonSettings: dtos.CommonSettingsDTO{
					Timeout:       30,
					RetryAttempts: 3,
					RetryDelay:    5,
				},
				Config: map[string]interface{}{
					"invalid": "config",
				},
			},
			mockResponse:   nil,
			mockError:      fmt.Errorf("invalid channel configuration"),
			expectedError:  true,
			expectedStatus: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			if tt.mockError != nil {
				mockCreateUseCase.On("Execute", mock.Anything, mock.MatchedBy(func(req *dtos.CreateChannelRequest) bool {
					return req.ChannelName == tt.request.ChannelName
				})).Return((*dtos.ChannelResponse)(nil), tt.mockError).Once()
			} else {
				mockCreateUseCase.On("Execute", mock.Anything, mock.MatchedBy(func(req *dtos.CreateChannelRequest) bool {
					return req.ChannelName == tt.request.ChannelName
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
			msg, err := nc.Request("eco1j.infra.eventcenter.channel.create", reqData, 5*time.Second)
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
				assert.Equal(t, tt.mockResponse.ChannelName, responseData["channelName"])
				assert.Equal(t, tt.mockResponse.ChannelType, responseData["channelType"])
			}
			
			// Verify mock was called
			mockCreateUseCase.AssertExpectations(t)
		})
	}
}

func TestChannelNATSHandler_GetChannel(t *testing.T) {
	_, nc := setupNATSServer(t)
	
	// Setup mocks
	mockCreateUseCase := &MockCreateChannelUseCase{}
	mockGetUseCase := &MockGetChannelUseCase{}
	mockListUseCase := &MockListChannelsUseCase{}
	mockUpdateUseCase := &MockUpdateChannelUseCase{}
	mockDeleteUseCase := &MockDeleteChannelUseCase{}
	
	// Create handler
	handler := NewChannelNATSHandler(
		mockCreateUseCase,
		mockGetUseCase,
		mockListUseCase,
		mockUpdateUseCase,
		mockDeleteUseCase,
		nc,
	)
	
	// Register handlers
	err := handler.RegisterHandlers()
	require.NoError(t, err)
	
	channelID := uuid.New().String()
	
	tests := []struct {
		name           string
		channelID      string
		mockResponse   *dtos.ChannelResponse
		mockError      error
		expectedError  bool
		expectedStatus bool
	}{
		{
			name:      "成功獲取 Channel",
			channelID: channelID,
			mockResponse: &dtos.ChannelResponse{
				ChannelID:   channelID,
				ChannelName: "Test Channel",
				Description: "Test channel description",
				Enabled:     true,
				ChannelType: "email",
				CreatedAt:   time.Now().UnixMilli(),
				UpdatedAt:   time.Now().UnixMilli(),
			},
			mockError:      nil,
			expectedError:  false,
			expectedStatus: true,
		},
		{
			name:           "Channel 不存在",
			channelID:      "non-existent-id",
			mockResponse:   nil,
			mockError:      fmt.Errorf("channel not found"),
			expectedError:  true,
			expectedStatus: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			if tt.mockError != nil {
				mockGetUseCase.On("Execute", mock.Anything, tt.channelID).Return((*dtos.ChannelResponse)(nil), tt.mockError).Once()
			} else {
				mockGetUseCase.On("Execute", mock.Anything, tt.channelID).Return(tt.mockResponse, nil).Once()
			}
			
			// Create NATS request
			reqSeqId := uuid.New().String()
			natsReq := NATSRequest{
				ReqSeqId:  reqSeqId,
				Data:      map[string]interface{}{"channelId": tt.channelID},
				Timestamp: time.Now().UnixMilli(),
			}
			
			reqData, err := json.Marshal(natsReq)
			require.NoError(t, err)
			
			// Send request and wait for response
			msg, err := nc.Request("eco1j.infra.eventcenter.channel.get", reqData, 5*time.Second)
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

func TestChannelNATSHandler_UpdateChannel(t *testing.T) {
	_, nc := setupNATSServer(t)
	
	// Setup mocks
	mockCreateUseCase := &MockCreateChannelUseCase{}
	mockGetUseCase := &MockGetChannelUseCase{}
	mockListUseCase := &MockListChannelsUseCase{}
	mockUpdateUseCase := &MockUpdateChannelUseCase{}
	mockDeleteUseCase := &MockDeleteChannelUseCase{}
	
	// Create handler
	handler := NewChannelNATSHandler(
		mockCreateUseCase,
		mockGetUseCase,
		mockListUseCase,
		mockUpdateUseCase,
		mockDeleteUseCase,
		nc,
	)
	
	// Register handlers
	err := handler.RegisterHandlers()
	require.NoError(t, err)
	
	channelID := uuid.New().String()
	
	// Test successful update
	t.Run("成功更新 Channel", func(t *testing.T) {
		updateReq := dtos.UpdateChannelRequest{
			ChannelID:   channelID,
			ChannelName: "Updated Channel Name",
			Description: "Updated description",
			Enabled:     false,
			ChannelType: "email",
			CommonSettings: dtos.CommonSettingsDTO{
				Timeout:       60,
				RetryAttempts: 5,
				RetryDelay:    10,
			},
			Config: map[string]interface{}{
				"host":        "smtp.updated.com",
				"port":        587,
				"secure":      false,
				"method":      "tls",
				"username":    "updated@test.com",
				"password":    "newpassword",
				"senderEmail": "updated@test.com",
			},
		}
		
		mockResponse := &dtos.ChannelResponse{
			ChannelID:   channelID,
			ChannelName: "Updated Channel Name",
			Description: "Updated description",
			Enabled:     false,
			ChannelType: "email",
			CommonSettings: dtos.CommonSettingsDTO{
				Timeout:       60,
				RetryAttempts: 5,
				RetryDelay:    10,
			},
			Config: updateReq.Config,
			UpdatedAt: time.Now().UnixMilli(),
		}
		
		// Setup mock expectations
		mockUpdateUseCase.On("Execute", mock.Anything, channelID, mock.MatchedBy(func(req *dtos.UpdateChannelRequest) bool {
			return req.ChannelName == updateReq.ChannelName
		})).Return(mockResponse, nil).Once()
		
		// Create NATS request
		reqSeqId := uuid.New().String()
		natsReq := NATSRequest{
			ReqSeqId:  reqSeqId,
			Data:      updateReq,
			Timestamp: time.Now().UnixMilli(),
		}
		
		reqData, err := json.Marshal(natsReq)
		require.NoError(t, err)
		
		// Send request and wait for response
		msg, err := nc.Request("eco1j.infra.eventcenter.channel.update", reqData, 5*time.Second)
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
		mockUpdateUseCase.AssertExpectations(t)
	})
}

func TestChannelNATSHandler_DeleteChannel(t *testing.T) {
	_, nc := setupNATSServer(t)
	
	// Setup mocks
	mockCreateUseCase := &MockCreateChannelUseCase{}
	mockGetUseCase := &MockGetChannelUseCase{}
	mockListUseCase := &MockListChannelsUseCase{}
	mockUpdateUseCase := &MockUpdateChannelUseCase{}
	mockDeleteUseCase := &MockDeleteChannelUseCase{}
	
	// Create handler
	handler := NewChannelNATSHandler(
		mockCreateUseCase,
		mockGetUseCase,
		mockListUseCase,
		mockUpdateUseCase,
		mockDeleteUseCase,
		nc,
	)
	
	// Register handlers
	err := handler.RegisterHandlers()
	require.NoError(t, err)
	
	channelID := uuid.New().String()
	
	// Test successful deletion
	t.Run("成功刪除 Channel", func(t *testing.T) {
		mockResponse := &dtos.DeleteChannelResponse{
			ChannelID: channelID,
			Deleted:   true,
			DeletedAt: time.Now().UnixMilli(),
		}
		
		// Setup mock expectations
		mockDeleteUseCase.On("Execute", mock.Anything, channelID).Return(mockResponse, nil).Once()
		
		// Create NATS request
		reqSeqId := uuid.New().String()
		natsReq := NATSRequest{
			ReqSeqId:  reqSeqId,
			Data:      map[string]interface{}{"channelId": channelID},
			Timestamp: time.Now().UnixMilli(),
		}
		
		reqData, err := json.Marshal(natsReq)
		require.NoError(t, err)
		
		// Send request and wait for response
		msg, err := nc.Request("eco1j.infra.eventcenter.channel.delete", reqData, 5*time.Second)
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
		mockDeleteUseCase.AssertExpectations(t)
	})
}

func TestChannelNATSHandler_ListChannels(t *testing.T) {
	_, nc := setupNATSServer(t)
	
	// Setup mocks
	mockCreateUseCase := &MockCreateChannelUseCase{}
	mockGetUseCase := &MockGetChannelUseCase{}
	mockListUseCase := &MockListChannelsUseCase{}
	mockUpdateUseCase := &MockUpdateChannelUseCase{}
	mockDeleteUseCase := &MockDeleteChannelUseCase{}
	
	// Create handler
	handler := NewChannelNATSHandler(
		mockCreateUseCase,
		mockGetUseCase,
		mockListUseCase,
		mockUpdateUseCase,
		mockDeleteUseCase,
		nc,
	)
	
	// Register handlers
	err := handler.RegisterHandlers()
	require.NoError(t, err)
	
	// Test successful list
	t.Run("成功列出 Channels", func(t *testing.T) {
		listReq := dtos.ListChannelsRequest{
			ChannelType:    "email",
			Tags:           []string{"test"},
			SkipCount:      0,
			MaxResultCount: 10,
		}
		
		mockResponse := &dtos.ListChannelsResponse{
			Items: []dtos.ChannelSummaryResponse{
				{
					ChannelID:   uuid.New().String(),
					ChannelName: "Test Channel 1",
					ChannelType: "email",
					Tags:        []string{"test"},
					Enabled:     true,
					CreatedAt:   time.Now().UnixMilli(),
					UpdatedAt:   time.Now().UnixMilli(),
				},
				{
					ChannelID:   uuid.New().String(),
					ChannelName: "Test Channel 2",
					ChannelType: "email",
					Tags:        []string{"test"},
					Enabled:     false,
					CreatedAt:   time.Now().UnixMilli(),
					UpdatedAt:   time.Now().UnixMilli(),
				},
			},
			SkipCount:      0,
			MaxResultCount: 10,
			TotalCount:     2,
			HasMore:        false,
		}
		
		// Setup mock expectations
		mockListUseCase.On("Execute", mock.Anything, mock.MatchedBy(func(req *dtos.ListChannelsRequest) bool {
			return req.ChannelType == listReq.ChannelType
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
		msg, err := nc.Request("eco1j.infra.eventcenter.channel.list", reqData, 5*time.Second)
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