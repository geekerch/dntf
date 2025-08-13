package external

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// OldSystemClient defines the interface for interacting with the old system's API.
type OldSystemClient interface {
	CreateGroup(req OldSystemCreateGroupRequest) (*OldSystemCreateGroupResponse, error)
}

// oldSystemClient implements OldSystemClient using HTTP.
type oldSystemClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewOldSystemClient creates a new instance of OldSystemClient.
func NewOldSystemClient(baseURL string) OldSystemClient {
	return &oldSystemClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second, // Set a reasonable timeout
		},
	}
}

// OldSystemCreateGroupRequest represents the request body for POST /v2/groups.
type OldSystemCreateGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	LevelName   string `json:"levelName"`
	Config      struct {
		Host        string `json:"host"`
		Port        int    `json:"port"`	
		Secure      bool   `json:"secure"`
		Method      string `json:"method"`
		Username    string `json:"username"`
		Password    string `json:"password"`
		SenderEmail string `json:"senderEmail"`
		EmailSubject string `json:"emailSubject"`
		Template    string `json:"template"`
	} `json:"config"`
	SendList []struct {
		FirstName   string `json:"firstName"`
		LastName    string `json:"lastName"`
		RecipientType string `json:"recipientType"`
		Target      string `json:"target"`
	} `json:"sendList"`
}

// OldSystemCreateGroupResponse represents a simplified response from POST /v2/groups.
// Assuming a simple success/failure or ID return. Adjust based on actual old system response.
type OldSystemCreateGroupResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// CreateGroup calls the old system's POST /v2/groups API.
func (c *oldSystemClient) CreateGroup(req OldSystemCreateGroupRequest) (*OldSystemCreateGroupResponse, error) {
	url := fmt.Sprintf("%s/v2/groups", c.baseURL)
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	// Add any necessary authentication headers here if the old system requires them
	// httpReq.Header.Set("Authorization", "Bearer YOUR_TOKEN")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request to old system: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("old system API returned non-success status: %d %s", resp.StatusCode, resp.Status)
	}

	var apiResp OldSystemCreateGroupResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode old system API response: %w", err)
	}

	return &apiResp, nil
}
