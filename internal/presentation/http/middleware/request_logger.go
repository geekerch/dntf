package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"notification/pkg/logger"
)

// RequestLogger is a middleware that logs HTTP requests
func RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Use structured logging with zap
		logger.Info("HTTP Request",
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.Int("status", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("client_ip", param.ClientIP),
			zap.String("user_agent", param.Request.UserAgent()),
			zap.Int("body_size", param.BodySize),
		)
		return ""
	})
}

// RequestID is a middleware that adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	// Simple implementation using timestamp
	// In production, consider using UUID or other unique ID generators
	return time.Now().Format("20060102150405.000000")
}