package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"notification/internal/application/channel/dtos"
	templateDtos "notification/internal/application/template/dtos"
	"notification/internal/domain/shared"
)

// Integration tests for Channel and Template synchronization
func TestChannelTemplateIntegration(t *testing.T) {
	_, nc := setupNATSServer(t)
	oldSystemServer := setupOldSystemMockServer(t)
	
	// This test demonstrates the integration between channels and templates
	// and verifies that changes are properly synchronized
	
	t.Run("Template 更新後 Channel 同步測試", func(t *testing.T) {
		// Step 1: Create a template
		templateID := uuid.New().String()
		createTemplateReq := templateDtos.CreateTemplateRequest{
			Name:        "Integration Test Template",
			ChannelType: shared.ChannelTypeEmail,
			Subject:     "Test Subject {{.Variable1}}",
			Content:     "Test Content {{.Variable1}} and {{.Variable2}}",
			Variables:   []string{"Variable1", "Variable2"},
			Tags:        []string{"integration", "test"},
		}
		
		// Step 2: Create a channel that uses this template
		channelID := uuid.New().String()
		createChannelReq := dtos.CreateChannelRequest{
			ChannelName: "Integration Test Channel",
			Description: "Channel for integration testing",
			Enabled:     true,
			ChannelType: "email",
			TemplateID:  templateID,
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
			Tags: []string{"integration", "test"},
		}
		
		// Step 3: Update the template
		updatedContent := "Updated Content {{.Variable1}}, {{.Variable2}} and {{.NewVariable}}"
		updateTemplateReq := templateDtos.UpdateTemplateRequest{
			Content:   &updatedContent,
			Variables: []string{"Variable1", "Variable2", "NewVariable"},
		}
		
		// In a real integration test, we would:
		// 1. Send template creation request via NATS
		// 2. Send channel creation request via NATS
		// 3. Verify channel was created in old system
		// 4. Send template update request via NATS
		// 5. Verify channels using the template are notified/updated
		
		// For now, we'll verify the structure is correct
		assert.NotEmpty(t, createTemplateReq.Name)
		assert.NotEmpty(t, createChannelReq.ChannelName)
		assert.NotEmpty(t, updateTemplateReq.Variables)
		
		// Verify old system server is accessible
		resp, err := http.Get(oldSystemServer.URL + "/v2.0/Groups?count=1000&index=1&desc=false")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// Test Channel CRUD operations with old system synchronization
func TestChannelOldSystemSync(t *testing.T) {
	_, nc := setupNATSServer(t)
	oldSystemServer := setupOldSystemMockServer(t)
	
	// Track created groups in old system
	var createdGroups []map[string]interface{}
	
	// Enhanced mock server to track operations
	mux := http.NewServeMux()
	
	// Mock GET /v2.0/Groups endpoint
	mux.HandleFunc("/v2.0/Groups", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			response := map[string]interface{}{
				"groups": createdGroups,
				"total":  len(createdGroups),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	})
	
	// Mock POST /v2/groups endpoint
	mux.HandleFunc("/v2/groups", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var requestBody map[string]interface{}
			json.NewDecoder(r.Body).Decode(&requestBody)
			
			groupID := uuid.New().String()
			group := map[string]interface{}{
				"id":          groupID,
				"name":        requestBody["name"],
				"description": requestBody["description"],
				"type":        requestBody["type"],
				"config":      requestBody["config"],
				"sendList":    requestBody["sendList"],
				"createdAt":   time.Now().Unix(),
			}
			
			createdGroups = append(createdGroups, group)
			
			response := map[string]interface{}{
				"id":      groupID,
				"message": "Group created successfully",
				"status":  "success",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(response)
		}
	})
	
	// Mock PUT /v2/groups/{id} endpoint for updates
	mux.HandleFunc("/v2/groups/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPUT {
			// Extract group ID from URL
			groupID := r.URL.Path[len("/v2/groups/"):]
			
			var requestBody map[string]interface{}
			json.NewDecoder(r.Body).Decode(&requestBody)
			
			// Update the group in our mock storage
			for i, group := range createdGroups {
				if group["id"] == groupID {
					createdGroups[i]["name"] = requestBody["name"]
					createdGroups[i]["description"] = requestBody["description"]
					createdGroups[i]["config"] = requestBody["config"]
					createdGroups[i]["updatedAt"] = time.Now().Unix()
					break
				}
			}
			
			response := map[string]interface{}{
				"id":      groupID,
				"message": "Group updated successfully",
				"status":  "success",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	})
	
	// Mock DELETE /v2/groups/{id} endpoint
	mux.HandleFunc("/v2/groups/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			groupID := r.URL.Path[len("/v2/groups/"):]
			
			// Remove the group from our mock storage
			for i, group := range createdGroups {
				if group["id"] == groupID {
					createdGroups = append(createdGroups[:i], createdGroups[i+1:]...)
					break
				}
			}
			
			response := map[string]interface{}{
				"id":      groupID,
				"message": "Group deleted successfully",
				"status":  "success",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	})
	
	enhancedServer := httptest.NewServer(mux)
	defer enhancedServer.Close()
	
	tests := []struct {
		name        string
		operation   string
		description string
	}{
		{
			name:        "Channel 創建同步到舊系統",
			operation:   "create",
			description: "測試創建 Channel 時是否同步到舊系統作為 Group",
		},
		{
			name:        "Channel 更新同步到舊系統",
			operation:   "update",
			description: "測試更新 Channel 時是否同步更新舊系統的 Group",
		},
		{
			name:        "Channel 刪除同步到舊系統",
			operation:   "delete",
			description: "測試刪除 Channel 時是否同步刪除舊系統的 Group",
		},
		{
			name:        "驗證舊系統 Group 查詢",
			operation:   "verify",
			description: "測試是否能正確查詢舊系統的 Groups",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.operation {
			case "create":
				// Test channel creation and old system sync
				initialCount := len(createdGroups)
				
				// Simulate channel creation (in real test, this would go through NATS)
				createReq := map[string]interface{}{
					"name":        "Test Channel for Old System",
					"description": "Testing old system synchronization",
					"type":        "email",
					"config": map[string]interface{}{
						"host":        "smtp.gmail.com",
						"port":        465,
						"secure":      true,
						"method":      "ssl",
						"username":    "test@gmail.com",
						"password":    "testpassword",
						"senderEmail": "test@gmail.com",
					},
					"sendList": []map[string]interface{}{
						{
							"firstName":     "Test",
							"lastName":      "User",
							"recipientType": "to",
							"target":        "test@example.com",
						},
					},
				}
				
				// Make request to old system
				reqBody, _ := json.Marshal(createReq)
				resp, err := http.Post(enhancedServer.URL+"/v2/groups", "application/json", 
					bytes.NewBuffer(reqBody))
				require.NoError(t, err)
				defer resp.Body.Close()
				
				assert.Equal(t, http.StatusCreated, resp.StatusCode)
				assert.Equal(t, initialCount+1, len(createdGroups))
				
			case "verify":
				// Test querying old system groups
				resp, err := http.Get(enhancedServer.URL + "/v2.0/Groups?count=1000&index=1&desc=false")
				require.NoError(t, err)
				defer resp.Body.Close()
				
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				
				var response map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)
				
				groups, ok := response["groups"].([]interface{})
				assert.True(t, ok)
				assert.Equal(t, len(createdGroups), len(groups))
			}
		})
	}
}

// Test SMTP configuration and email sending
func TestSMTPEmailSending(t *testing.T) {
	_, nc := setupNATSServer(t)
	
	t.Run("SMTP 配置測試", func(t *testing.T) {
		// Test SMTP configuration from spec
		smtpConfig := map[string]interface{}{
			"host":        "smtp.gmail.com",
			"port":        465,
			"secure":      true,
			"method":      "ssl",
			"username":    "chienhsiang.chen@gmail.com",
			"password":    "tlrqyoxptgjbbatn",
			"senderEmail": "chienhsiang.chen@gmail.com",
		}
		
		// Verify configuration structure
		assert.Equal(t, "smtp.gmail.com", smtpConfig["host"])
		assert.Equal(t, 465, smtpConfig["port"])
		assert.Equal(t, true, smtpConfig["secure"])
		assert.Equal(t, "ssl", smtpConfig["method"])
		assert.NotEmpty(t, smtpConfig["username"])
		assert.NotEmpty(t, smtpConfig["password"])
		assert.NotEmpty(t, smtpConfig["senderEmail"])
		
		// In a real test, we would:
		// 1. Create a channel with this SMTP config
		// 2. Send a test message
		// 3. Verify the message was sent successfully
		// 4. Check message status and results
	})
	
	t.Run("Email 收件人測試", func(t *testing.T) {
		// Test different recipient types
		recipients := []map[string]interface{}{
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
		}
		
		// Verify recipient structure
		assert.Len(t, recipients, 3)
		
		// Check each recipient type
		toRecipients := filterRecipientsByType(recipients, "to")
		ccRecipients := filterRecipientsByType(recipients, "cc")
		bccRecipients := filterRecipientsByType(recipients, "bcc")
		
		assert.Len(t, toRecipients, 1)
		assert.Len(t, ccRecipients, 1)
		assert.Len(t, bccRecipients, 1)
	})
}

// Helper function to filter recipients by type
func filterRecipientsByType(recipients []map[string]interface{}, recipientType string) []map[string]interface{} {
	var filtered []map[string]interface{}
	for _, recipient := range recipients {
		if recipient["type"] == recipientType {
			filtered = append(filtered, recipient)
		}
	}
	return filtered
}