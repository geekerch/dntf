package models

// APIResponse represents a standard API response structure
type APIResponse struct {
	Data  interface{} `json:"data,omitempty"`
	Error *APIError   `json:"error,omitempty"`
}

// APIError represents an error response structure
type APIError struct {
	Code    string `json:"code" example:"INVALID_REQUEST"`
	Message string `json:"message" example:"The request is invalid"`
}

// SuccessResponse represents a successful response
type SuccessResponse struct {
	Success bool        `json:"success" example:"true"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty" example:"Operation completed successfully"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"Error message"`
	Details string `json:"details,omitempty" example:"Detailed error information"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status  string `json:"status" example:"healthy"`
	Service string `json:"service" example:"notification-api"`
	Version string `json:"version" example:"1.0.0"`
}

// InfoResponse represents the API info response
type InfoResponse struct {
	Service   string   `json:"service" example:"notification-api"`
	Version   string   `json:"version" example:"1.0.0"`
	Endpoints []string `json:"endpoints"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Items          interface{} `json:"items"`
	SkipCount      int         `json:"skipCount" example:"0"`
	MaxResultCount int         `json:"maxResultCount" example:"20"`
	TotalCount     int         `json:"totalCount" example:"100"`
	HasMore        bool        `json:"hasMore" example:"true"`
}