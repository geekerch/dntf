package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
	Code    string `json:"code,omitempty"`
}

// ErrorHandler is a middleware that handles panics and errors
func ErrorHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Internal server error",
				Details: err,
				Code:    "INTERNAL_ERROR",
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: "Internal server error",
				Code:  "INTERNAL_ERROR",
			})
		}
		c.Abort()
	})
}

// NotFoundHandler handles 404 errors
func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error: "Resource not found",
			Code:  "NOT_FOUND",
		})
	}
}

// MethodNotAllowedHandler handles 405 errors
func MethodNotAllowedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, ErrorResponse{
			Error: "Method not allowed",
			Code:  "METHOD_NOT_ALLOWED",
		})
	}
}