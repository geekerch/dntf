package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"notification/pkg/logger"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	// AllowedOrigins is a list of allowed origins
	AllowedOrigins []string
	// AllowedMethods is a list of allowed HTTP methods
	AllowedMethods []string
	// AllowedHeaders is a list of allowed headers
	AllowedHeaders []string
	// ExposedHeaders is a list of headers exposed to the client
	ExposedHeaders []string
	// AllowCredentials indicates whether credentials are allowed
	AllowCredentials bool
	// MaxAge indicates how long the results of a preflight request can be cached
	MaxAge time.Duration
	// AllowWildcard allows wildcard origins (use with caution)
	AllowWildcard bool
	// AllowPrivateNetwork allows private network access
	AllowPrivateNetwork bool
}

// CORSMiddleware provides CORS functionality
type CORSMiddleware struct {
	config *CORSConfig
}

// NewCORSMiddleware creates a new CORS middleware
func NewCORSMiddleware(config *CORSConfig) *CORSMiddleware {
	if config == nil {
		config = DefaultCORSConfig()
	}
	return &CORSMiddleware{config: config}
}

// Handler returns the CORS middleware handler
func (cm *CORSMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			cm.handlePreflight(c, origin)
			return
		}

		// Handle actual requests
		cm.handleActualRequest(c, origin)
		c.Next()
	}
}

// handlePreflight handles CORS preflight requests
func (cm *CORSMiddleware) handlePreflight(c *gin.Context, origin string) {
	// Check if origin is allowed
	if !cm.isOriginAllowed(origin) {
		logger.Warn("CORS preflight rejected - origin not allowed",
			zap.String("origin", origin),
			zap.String("client_ip", c.ClientIP()))
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	// Set CORS headers for preflight
	cm.setCORSHeaders(c, origin)

	// Handle preflight-specific headers
	requestMethod := c.Request.Header.Get("Access-Control-Request-Method")
	if requestMethod != "" && cm.isMethodAllowed(requestMethod) {
		c.Header("Access-Control-Allow-Methods", strings.Join(cm.config.AllowedMethods, ", "))
	}

	requestHeaders := c.Request.Header.Get("Access-Control-Request-Headers")
	if requestHeaders != "" {
		allowedHeaders := cm.getAllowedRequestHeaders(requestHeaders)
		if len(allowedHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
		}
	}

	// Set max age for preflight cache
	if cm.config.MaxAge > 0 {
		c.Header("Access-Control-Max-Age", strconv.Itoa(int(cm.config.MaxAge.Seconds())))
	}

	// Handle private network access
	if cm.config.AllowPrivateNetwork && c.Request.Header.Get("Access-Control-Request-Private-Network") == "true" {
		c.Header("Access-Control-Allow-Private-Network", "true")
	}

	logger.Debug("CORS preflight handled",
		zap.String("origin", origin),
		zap.String("method", requestMethod),
		zap.String("headers", requestHeaders))

	c.AbortWithStatus(http.StatusNoContent)
}

// handleActualRequest handles actual CORS requests
func (cm *CORSMiddleware) handleActualRequest(c *gin.Context, origin string) {
	// Check if origin is allowed
	if !cm.isOriginAllowed(origin) {
		logger.Warn("CORS request rejected - origin not allowed",
			zap.String("origin", origin),
			zap.String("method", c.Request.Method),
			zap.String("client_ip", c.ClientIP()))
		return
	}

	// Set CORS headers for actual request
	cm.setCORSHeaders(c, origin)

	logger.Debug("CORS request handled",
		zap.String("origin", origin),
		zap.String("method", c.Request.Method))
}

// setCORSHeaders sets common CORS headers
func (cm *CORSMiddleware) setCORSHeaders(c *gin.Context, origin string) {
	// Set allowed origin
	if cm.config.AllowWildcard && origin == "" {
		c.Header("Access-Control-Allow-Origin", "*")
	} else if origin != "" {
		c.Header("Access-Control-Allow-Origin", origin)
	}

	// Set credentials
	if cm.config.AllowCredentials {
		c.Header("Access-Control-Allow-Credentials", "true")
	}

	// Set exposed headers
	if len(cm.config.ExposedHeaders) > 0 {
		c.Header("Access-Control-Expose-Headers", strings.Join(cm.config.ExposedHeaders, ", "))
	}

	// Set Vary header to indicate that the response varies based on Origin
	c.Header("Vary", "Origin")
}

// isOriginAllowed checks if the origin is allowed
func (cm *CORSMiddleware) isOriginAllowed(origin string) bool {
	if origin == "" {
		return true // Allow requests without origin (e.g., same-origin requests)
	}

	if cm.config.AllowWildcard {
		return true
	}

	for _, allowedOrigin := range cm.config.AllowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return true
		}
		
		// Support wildcard subdomains (e.g., *.example.com)
		if strings.HasPrefix(allowedOrigin, "*.") {
			domain := allowedOrigin[2:]
			if strings.HasSuffix(origin, "."+domain) || origin == domain {
				return true
			}
		}
	}

	return false
}

// isMethodAllowed checks if the HTTP method is allowed
func (cm *CORSMiddleware) isMethodAllowed(method string) bool {
	for _, allowedMethod := range cm.config.AllowedMethods {
		if allowedMethod == method {
			return true
		}
	}
	return false
}

// getAllowedRequestHeaders filters request headers to only include allowed ones
func (cm *CORSMiddleware) getAllowedRequestHeaders(requestHeaders string) []string {
	headers := strings.Split(requestHeaders, ",")
	var allowedHeaders []string

	for _, header := range headers {
		header = strings.TrimSpace(header)
		if cm.isHeaderAllowed(header) {
			allowedHeaders = append(allowedHeaders, header)
		}
	}

	return allowedHeaders
}

// isHeaderAllowed checks if the header is allowed
func (cm *CORSMiddleware) isHeaderAllowed(header string) bool {
	header = strings.ToLower(header)
	
	// Always allow simple headers
	simpleHeaders := []string{
		"accept",
		"accept-language",
		"content-language",
		"content-type",
	}
	
	for _, simpleHeader := range simpleHeaders {
		if header == simpleHeader {
			return true
		}
	}

	// Check configured allowed headers
	for _, allowedHeader := range cm.config.AllowedHeaders {
		if strings.ToLower(allowedHeader) == header {
			return true
		}
	}

	return false
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
			"HEAD",
			"PATCH",
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-Request-ID",
			"X-API-Key",
		},
		ExposedHeaders: []string{
			"X-Request-ID",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
		},
		AllowCredentials:    false,
		MaxAge:             12 * time.Hour,
		AllowWildcard:      true,
		AllowPrivateNetwork: false,
	}
}

// RestrictiveCORSConfig returns a more restrictive CORS configuration
func RestrictiveCORSConfig(allowedOrigins []string) *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-Request-ID",
			"X-API-Key",
		},
		ExposedHeaders: []string{
			"X-Request-ID",
		},
		AllowCredentials:    true,
		MaxAge:             1 * time.Hour,
		AllowWildcard:      false,
		AllowPrivateNetwork: false,
	}
}

// DevelopmentCORSConfig returns a CORS configuration suitable for development
func DevelopmentCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:8080",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:3001",
			"http://127.0.0.1:8080",
		},
		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
			"HEAD",
			"PATCH",
		},
		AllowedHeaders: []string{
			"*",
		},
		ExposedHeaders: []string{
			"*",
		},
		AllowCredentials:    true,
		MaxAge:             1 * time.Hour,
		AllowWildcard:      false,
		AllowPrivateNetwork: true,
	}
}

// ProductionCORSConfig returns a CORS configuration suitable for production
func ProductionCORSConfig(allowedOrigins []string) *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-Request-ID",
			"X-API-Key",
		},
		ExposedHeaders: []string{
			"X-Request-ID",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
		},
		AllowCredentials:    true,
		MaxAge:             24 * time.Hour,
		AllowWildcard:      false,
		AllowPrivateNetwork: false,
	}
}

// CORS helper functions for easy setup

// DefaultCORS creates CORS middleware with default settings
func DefaultCORS() gin.HandlerFunc {
	middleware := NewCORSMiddleware(DefaultCORSConfig())
	return middleware.Handler()
}

// DevelopmentCORS creates CORS middleware for development
func DevelopmentCORS() gin.HandlerFunc {
	middleware := NewCORSMiddleware(DevelopmentCORSConfig())
	return middleware.Handler()
}

// ProductionCORS creates CORS middleware for production
func ProductionCORS(allowedOrigins []string) gin.HandlerFunc {
	middleware := NewCORSMiddleware(ProductionCORSConfig(allowedOrigins))
	return middleware.Handler()
}