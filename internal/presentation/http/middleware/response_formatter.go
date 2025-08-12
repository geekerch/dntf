package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// StandardResponse represents a standardized API response
type StandardResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	RequestID string      `json:"requestId,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// ErrorInfo represents error information in the response
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ResponseFormatter is a middleware that formats responses in a standard format
func ResponseFormatter() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Only format JSON responses
		if c.GetHeader("Content-Type") == "application/json" {
			return
		}

		// Get the response data
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			// Success response - data should already be set by handlers
			return
		}

		// Error responses are handled by error middleware
	}
}

// CORS middleware for handling Cross-Origin Resource Sharing
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}