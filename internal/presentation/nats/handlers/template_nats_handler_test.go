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

	"notification/internal/application/template/dtos"
	"notification/internal/application/template/usecases"
	"notification/internal/domain/shared"
)

// MockTemplateUseCase mocks for template use cases
type MockCreateTemplateUseCase struct {
	mock.Mock
}

func (m *MockCreateTemplateUseCase) Execute(ctx context.Context, req *dtos.CreateTemplateRequest) (*dtos.TemplateResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*dtos.TemplateResponse), args.Error(1)
}

type MockGetTemplateUseCase struct {
	mock.Mock
}

func (m *MockGetTemplateUseCase) Execute(ctx context.Context, templateID string) (*dtos.TemplateResponse, error) {
	args := m.Called(ctx, templateID)
	return args.Get(0).(*dtos.TemplateResponse), args.Error(1)
}

type MockListTemplatesUseCase struct {
	mock.Mock
}

func (m *MockListTemplatesUseCase) Execute(ctx context.Context, req *dtos.ListTemplatesRequest) (*dtos.ListTemplatesResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*dtos.ListTemplatesResponse), args.Error(1)
}

type MockUpdateTemplateUseCase struct {
	mock.Mock
}

func (m *MockUpdateTemplateUseCase) Execute(ctx context.Context, templateID string, req *dtos.UpdateTemplateRequest) (*dtos.TemplateResponse, error) {
	args := m.Called(ctx, templateID, req)
	return args.Get(0).(*dtos.TemplateResponse), args.Error(1)
}

type MockDeleteTemplateUseCase struct {
	mock.Mock
}

func (m *MockDeleteTemplateUseCase) Execute(ctx context.Context, templateID string) error {
	args := m.Called(ctx, templateID)
	return args.Error(0)
}

func TestTemplateNATSHandler_CreateTemplate(t *testing.T) {
	_, nc := setupNATSServer(t)
	
	// Setup mocks
	mockCreateUseCase := &MockCreateTemplateUseCase{}
	mockGetUseCase := &MockGetTemplateUseCase{}
	mockListUseCase := &MockListTemplatesUseCase{}
	mockUpdateUseCase := &MockUpdateTemplateUseCase{}
	mockDeleteUseCase := &MockDeleteTemplateUseCase{}
	
	// Create handler
	handler := NewTemplateNATSHandler(
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
		request        dtos.CreateTemplateRequest
		mockResponse   *dtos.TemplateResponse
		mockError      error
		expectedError  bool
		expectedStatus bool
	}{
		{
			name: "成功創建 Email Template",
			request: dtos.CreateTemplateRequest{
				Name:        "Welcome Email Template",
				ChannelType: shared.ChannelTypeEmail,
				Subject:     "Welcome to {{.CompanyName}}!",
				Content:     "Hello {{.UserName}}, welcome to our platform! Your account has been created successfully.",
				Variables:   []string{"CompanyName", "UserName"},
				Tags:        []string{"welcome", "email", "onboarding"},
				Settings: &shared.CommonSettings{
					Timeout:       30,
					RetryAttempts: 3,
					RetryDelay:    5,
				},
			},
			mockResponse: &dtos.TemplateResponse{
				ID:          uuid.New().String(),
				Name:        "Welcome Email Template",
				ChannelType: shared.ChannelTypeEmail,
				Subject:     "Welcome to {{.CompanyName}}!",
				Content:     "Hello {{.UserName}}, welcome to our platform! Your account has been created successfully.",
				Variables:   []string{"CompanyName", "UserName"},
				Tags:        []string{"welcome", "email", "onboarding"},
				Version:     1,
				Settings: &shared.CommonSettings{
					Timeout:       30,
					RetryAttempts: 3,
					RetryDelay:    5,
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			mockError:      nil,
			expectedError:  false,
			expectedStatus: true,
		},
		{
			name: "成功創建 SMS Template",
			request: dtos.CreateTemplateRequest{
				Name:        "SMS Verification Template",
				ChannelType: shared.ChannelTypeSMS,
				Content:     "Your verification code is: {{.VerificationCode}}. Valid for 5 minutes.",
				Variables:   []string{"VerificationCode"},
				Tags:        []string{"sms", "verification", "security"},
			},
			mockResponse: &dtos.TemplateResponse{
				ID:          uuid.New().String(),
				Name:        "SMS Verification Template",
				ChannelType: shared.ChannelTypeSMS,
				Content:     "Your verification code is: {{.VerificationCode}}. Valid for 5 minutes.",
				Variables:   []string{"VerificationCode"},
				Tags:        []string{"sms", "verification", "security"},
				Version:     1,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			mockError:      nil,
			expectedError:  false,
			expectedStatus: true,
		},
		{
			name: "創建失敗 - 無效內容",
			request: dtos.CreateTemplateRequest{
				Name:        "Invalid Template",
				ChannelType: shared.ChannelTypeEmail,
				Content:     "", // Empty content should fail
			},
			mockResponse:   nil,
			mockError:      fmt.Errorf("template content cannot be empty"),
			expectedError:  true,
			expectedStatus: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			if tt.mockError != nil {
				mockCreateUseCase.On("Execute", mock.Anything, mock.MatchedBy(func(req *dtos.CreateTemplateRequest) bool {
					return req.Name == tt.request.Name
				})).Return((*dtos.TemplateResponse)(nil), tt.mockError).Once()
			} else {
				mockCreateUseCase.On("Execute", mock.Anything, mock.MatchedBy(func(req *dtos.CreateTemplateRequest) bool {
					return req.Name == tt.request.Name
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
			msg, err := nc.Request("eco1j.infra.eventcenter.template.create", reqData, 5*time.Second)
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
				assert.Equal(t, tt.mockResponse.Name, responseData["name"])
				assert.Equal(t, string(tt.mockResponse.ChannelType), responseData["channelType"])
			}
			
			// Verify mock was called
			mockCreateUseCase.AssertExpectations(t)
		})
	}
}

func TestTemplateNATSHandler_GetTemplate(t *testing.T) {
	_, nc := setupNATSServer(t)
	
	// Setup mocks
	mockCreateUseCase := &MockCreateTemplateUseCase{}
	mockGetUseCase := &MockGetTemplateUseCase{}
	mockListUseCase := &MockListTemplatesUseCase{}
	mockUpdateUseCase := &MockUpdateTemplateUseCase{}
	mockDeleteUseCase := &MockDeleteTemplateUseCase{}
	
	// Create handler
	handler := NewTemplateNATSHandler(
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
	
	templateID := uuid.New().String()
	
	tests := []struct {
		name           string
		templateID     string
		mockResponse   *dtos.TemplateResponse
		mockError      error
		expectedError  bool
		expectedStatus bool
	}{
		{
			name:       "成功獲取 Template",
			templateID: templateID,
			mockResponse: &dtos.TemplateResponse{
				ID:          templateID,
				Name:        "Test Template",
				ChannelType: shared.ChannelTypeEmail,
				Subject:     "Test Subject",
				Content:     "Test Content with {{.Variable}}",
				Variables:   []string{"Variable"},
				Tags:        []string{"test"},
				Version:     1,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			mockError:      nil,
			expectedError:  false,
			expectedStatus: true,
		},
		{
			name:           "Template 不存在",
			templateID:     "non-existent-id",
			mockResponse:   nil,
			mockError:      fmt.Errorf("template not found"),
			expectedError:  true,
			expectedStatus: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			if tt.mockError != nil {
				mockGetUseCase.On("Execute", mock.Anything, tt.templateID).Return((*dtos.TemplateResponse)(nil), tt.mockError).Once()
			} else {
				mockGetUseCase.On("Execute", mock.Anything, tt.templateID).Return(tt.mockResponse, nil).Once()
			}
			
			// Create NATS request
			reqSeqId := uuid.New().String()
			natsReq := NATSRequest{
				ReqSeqId:  reqSeqId,
				Data:      map[string]interface{}{"templateId": tt.templateID},
				Timestamp: time.Now().UnixMilli(),
			}
			
			reqData, err := json.Marshal(natsReq)
			require.NoError(t, err)
			
			// Send request and wait for response
			msg, err := nc.Request("eco1j.infra.eventcenter.template.get", reqData, 5*time.Second)
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

func TestTemplateNATSHandler_UpdateTemplate(t *testing.T) {
	_, nc := setupNATSServer(t)
	
	// Setup mocks
	mockCreateUseCase := &MockCreateTemplateUseCase{}
	mockGetUseCase := &MockGetTemplateUseCase{}
	mockListUseCase := &MockListTemplatesUseCase{}
	mockUpdateUseCase := &MockUpdateTemplateUseCase{}
	mockDeleteUseCase := &MockDeleteTemplateUseCase{}
	
	// Create handler
	handler := NewTemplateNATSHandler(
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
	
	templateID := uuid.New().String()
	
	// Test successful update
	t.Run("成功更新 Template", func(t *testing.T) {
		updatedName := "Updated Template Name"
		updatedContent := "Updated content with {{.NewVariable}}"
		
		updateReq := dtos.UpdateTemplateRequest{
			Name:      &updatedName,
			Content:   &updatedContent,
			Variables: []string{"NewVariable"},
			Tags:      []string{"updated", "test"},
		}
		
		mockResponse := &dtos.TemplateResponse{
			ID:          templateID,
			Name:        updatedName,
			ChannelType: shared.ChannelTypeEmail,
			Subject:     "Original Subject",
			Content:     updatedContent,
			Variables:   []string{"NewVariable"},
			Tags:        []string{"updated", "test"},
			Version:     2, // Version should increment
			CreatedAt:   time.Now().Add(-1 * time.Hour),
			UpdatedAt:   time.Now(),
		}
		
		// Setup mock expectations
		mockUpdateUseCase.On("Execute", mock.Anything, templateID, mock.MatchedBy(func(req *dtos.UpdateTemplateRequest) bool {
			return req.Name != nil && *req.Name == updatedName
		})).Return(mockResponse, nil).Once()
		
		// Create NATS request
		reqSeqId := uuid.New().String()
		natsReq := NATSRequest{
			ReqSeqId:  reqSeqId,
			Data:      map[string]interface{}{
				"templateId": templateID,
				"name":       updatedName,
				"content":    updatedContent,
				"variables":  []string{"NewVariable"},
				"tags":       []string{"updated", "test"},
			},
			Timestamp: time.Now().UnixMilli(),
		}
		
		reqData, err := json.Marshal(natsReq)
		require.NoError(t, err)
		
		// Send request and wait for response
		msg, err := nc.Request("eco1j.infra.eventcenter.template.update", reqData, 5*time.Second)
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
		
		// Verify response data
		responseData, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, updatedName, responseData["name"])
		assert.Equal(t, updatedContent, responseData["content"])
		assert.Equal(t, float64(2), responseData["version"]) // JSON numbers are float64
		
		// Verify mock was called
		mockUpdateUseCase.AssertExpectations(t)
	})
}

func TestTemplateNATSHandler_DeleteTemplate(t *testing.T) {
	_, nc := setupNATSServer(t)
	
	// Setup mocks
	mockCreateUseCase := &MockCreateTemplateUseCase{}
	mockGetUseCase := &MockGetTemplateUseCase{}
	mockListUseCase := &MockListTemplatesUseCase{}
	mockUpdateUseCase := &MockUpdateTemplateUseCase{}
	mockDeleteUseCase := &MockDeleteTemplateUseCase{}
	
	// Create handler
	handler := NewTemplateNATSHandler(
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
	
	templateID := uuid.New().String()
	
	tests := []struct {
		name           string
		templateID     string
		mockError      error
		expectedError  bool
		expectedStatus bool
	}{
		{
			name:           "成功刪除 Template",
			templateID:     templateID,
			mockError:      nil,
			expectedError:  false,
			expectedStatus: true,
		},
		{
			name:           "刪除失敗 - Template 不存在",
			templateID:     "non-existent-id",
			mockError:      fmt.Errorf("template not found"),
			expectedError:  true,
			expectedStatus: false,
		},
		{
			name:           "刪除失敗 - Template 正在使用中",
			templateID:     templateID,
			mockError:      fmt.Errorf("template is in use by channels"),
			expectedError:  true,
			expectedStatus: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			mockDeleteUseCase.On("Execute", mock.Anything, tt.templateID).Return(tt.mockError).Once()
			
			// Create NATS request
			reqSeqId := uuid.New().String()
			natsReq := NATSRequest{
				ReqSeqId:  reqSeqId,
				Data:      map[string]interface{}{"templateId": tt.templateID},
				Timestamp: time.Now().UnixMilli(),
			}
			
			reqData, err := json.Marshal(natsReq)
			require.NoError(t, err)
			
			// Send request and wait for response
			msg, err := nc.Request("eco1j.infra.eventcenter.template.delete", reqData, 5*time.Second)
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
			}
			
			// Verify mock was called
			mockDeleteUseCase.AssertExpectations(t)
		})
	}
}

func TestTemplateNATSHandler_ListTemplates(t *testing.T) {
	_, nc := setupNATSServer(t)
	
	// Setup mocks
	mockCreateUseCase := &MockCreateTemplateUseCase{}
	mockGetUseCase := &MockGetTemplateUseCase{}
	mockListUseCase := &MockListTemplatesUseCase{}
	mockUpdateUseCase := &MockUpdateTemplateUseCase{}
	mockDeleteUseCase := &MockDeleteTemplateUseCase{}
	
	// Create handler
	handler := NewTemplateNATSHandler(
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
		name         string
		request      dtos.ListTemplatesRequest
		mockResponse *dtos.ListTemplatesResponse
		mockError    error
	}{
		{
			name: "成功列出所有 Templates",
			request: dtos.ListTemplatesRequest{
				SkipCount:      0,
				MaxResultCount: 10,
			},
			mockResponse: &dtos.ListTemplatesResponse{
				Items: []*dtos.TemplateResponse{
					{
						ID:          uuid.New().String(),
						Name:        "Email Template 1",
						ChannelType: shared.ChannelTypeEmail,
						Subject:     "Subject 1",
						Content:     "Content 1",
						Tags:        []string{"email", "test"},
						Version:     1,
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					},
					{
						ID:          uuid.New().String(),
						Name:        "SMS Template 1",
						ChannelType: shared.ChannelTypeSMS,
						Content:     "SMS Content 1",
						Tags:        []string{"sms", "test"},
						Version:     1,
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					},
				},
				SkipCount:      0,
				MaxResultCount: 10,
				TotalCount:     2,
				HasMore:        false,
			},
			mockError: nil,
		},
		{
			name: "按 ChannelType 篩選 Templates",
			request: dtos.ListTemplatesRequest{
				ChannelType:    &shared.ChannelTypeEmail,
				SkipCount:      0,
				MaxResultCount: 5,
			},
			mockResponse: &dtos.ListTemplatesResponse{
				Items: []*dtos.TemplateResponse{
					{
						ID:          uuid.New().String(),
						Name:        "Email Template 1",
						ChannelType: shared.ChannelTypeEmail,
						Subject:     "Subject 1",
						Content:     "Content 1",
						Tags:        []string{"email"},
						Version:     1,
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					},
				},
				SkipCount:      0,
				MaxResultCount: 5,
				TotalCount:     1,
				HasMore:        false,
			},
			mockError: nil,
		},
		{
			name: "按 Tags 篩選 Templates",
			request: dtos.ListTemplatesRequest{
				Tags:           []string{"welcome"},
				SkipCount:      0,
				MaxResultCount: 10,
			},
			mockResponse: &dtos.ListTemplatesResponse{
				Items: []*dtos.TemplateResponse{
					{
						ID:          uuid.New().String(),
						Name:        "Welcome Template",
						ChannelType: shared.ChannelTypeEmail,
						Subject:     "Welcome!",
						Content:     "Welcome content",
						Tags:        []string{"welcome", "onboarding"},
						Version:     1,
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					},
				},
				SkipCount:      0,
				MaxResultCount: 10,
				TotalCount:     1,
				HasMore:        false,
			},
			mockError: nil,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			if tt.mockError != nil {
				mockListUseCase.On("Execute", mock.Anything, mock.MatchedBy(func(req *dtos.ListTemplatesRequest) bool {
					return req.MaxResultCount == tt.request.MaxResultCount
				})).Return((*dtos.ListTemplatesResponse)(nil), tt.mockError).Once()
			} else {
				mockListUseCase.On("Execute", mock.Anything, mock.MatchedBy(func(req *dtos.ListTemplatesRequest) bool {
					return req.MaxResultCount == tt.request.MaxResultCount
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
			msg, err := nc.Request("eco1j.infra.eventcenter.template.list", reqData, 5*time.Second)
			require.NoError(t, err)
			
			// Parse response
			var response NATSResponse
			err = json.Unmarshal(msg.Data, &response)
			require.NoError(t, err)
			
			// Verify response
			assert.Equal(t, reqSeqId, response.ReqSeqId)
			
			if tt.mockError != nil {
				assert.False(t, response.Success)
				assert.NotNil(t, response.Error)
			} else {
				assert.True(t, response.Success)
				assert.Nil(t, response.Error)
				assert.NotNil(t, response.Data)
				
				// Verify response data structure
				responseData, ok := response.Data.(map[string]interface{})
				assert.True(t, ok)
				
				items, ok := responseData["items"].([]interface{})
				assert.True(t, ok)
				assert.Equal(t, len(tt.mockResponse.Items), len(items))
			}
			
			// Verify mock was called
			mockListUseCase.AssertExpectations(t)
		})
	}
}

// Integration test for template and channel synchronization
func TestTemplateChannelSynchronization(t *testing.T) {
	_, nc := setupNATSServer(t)
	
	// This test would verify that when a template is updated,
	// channels using that template are notified or updated accordingly
	// This is more of an integration test that would require actual use cases
	
	t.Run("Template 更新後 Channel 同步", func(t *testing.T) {
		// This test would be implemented when we have the actual
		// synchronization mechanism between templates and channels
		t.Skip("Integration test - requires actual synchronization implementation")
	})
}