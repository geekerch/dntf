package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"notification/pkg/logger"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	// JWT secret key for token validation
	JWTSecret string
	// API keys for simple authentication
	APIKeys map[string]string
	// Skip authentication for these paths
	SkipPaths []string
	// Authentication type: "jwt", "api-key", "basic"
	AuthType string
}

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	config *AuthConfig
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(config *AuthConfig) *AuthMiddleware {
	if config == nil {
		config = &AuthConfig{
			AuthType:  "api-key",
			SkipPaths: []string{"/health", "/metrics"},
			APIKeys:   make(map[string]string),
		}
	}
	return &AuthMiddleware{config: config}
}

// Handler returns the authentication middleware handler
func (a *AuthMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for certain paths
		if a.shouldSkipAuth(c.Request.URL.Path) {
			c.Next()
			return
		}

		var authenticated bool
		var userID string
		var err error

		switch a.config.AuthType {
		case "jwt":
			authenticated, userID, err = a.validateJWT(c)
		case "api-key":
			authenticated, userID, err = a.validateAPIKey(c)
		case "basic":
			authenticated, userID, err = a.validateBasicAuth(c)
		default:
			authenticated, userID, err = a.validateAPIKey(c)
		}

		if err != nil {
			logger.Warn("Authentication error",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("client_ip", c.ClientIP()),
				zap.Error(err))

			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "Authentication failed",
				Details: err.Error(),
				Code:    "AUTH_FAILED",
			})
			c.Abort()
			return
		}

		if !authenticated {
			logger.Warn("Authentication failed - invalid credentials",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("client_ip", c.ClientIP()))

			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error: "Invalid credentials",
				Code:  "INVALID_CREDENTIALS",
			})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", userID)
		c.Set("authenticated", true)

		logger.Debug("Authentication successful",
			zap.String("user_id", userID),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method))

		c.Next()
	}
}

// shouldSkipAuth checks if authentication should be skipped for the given path
func (a *AuthMiddleware) shouldSkipAuth(path string) bool {
	for _, skipPath := range a.config.SkipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// validateAPIKey validates API key authentication
func (a *AuthMiddleware) validateAPIKey(c *gin.Context) (bool, string, error) {
	// Try to get API key from header
	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		// Try to get from Authorization header
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			apiKey = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if apiKey == "" {
		return false, "", nil
	}

	// Validate API key
	if userID, exists := a.config.APIKeys[apiKey]; exists {
		return true, userID, nil
	}

	return false, "", nil
}

// validateJWT validates JWT token authentication
func (a *AuthMiddleware) validateJWT(c *gin.Context) (bool, string, error) {
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return false, "", nil
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return false, "", nil
	}

	// TODO: Implement JWT validation logic
	// This is a placeholder implementation
	// In a real application, you would:
	// 1. Parse the JWT token
	// 2. Validate the signature using the secret
	// 3. Check expiration
	// 4. Extract user information

	// For now, just check if it's a valid format (placeholder)
	if len(token) > 10 {
		return true, "jwt-user", nil
	}

	return false, "", nil
}

// validateBasicAuth validates basic authentication
func (a *AuthMiddleware) validateBasicAuth(c *gin.Context) (bool, string, error) {
	username, password, hasAuth := c.Request.BasicAuth()
	if !hasAuth {
		return false, "", nil
	}

	// TODO: Implement basic auth validation logic
	// This is a placeholder implementation
	// In a real application, you would:
	// 1. Hash the password
	// 2. Compare with stored credentials
	// 3. Check user permissions

	// For now, just check if credentials are provided (placeholder)
	if username != "" && password != "" {
		return true, username, nil
	}

	return false, "", nil
}

// RequireAuth is a helper function to create an auth middleware with default config
func RequireAuth() gin.HandlerFunc {
	config := &AuthConfig{
		AuthType: "api-key",
		SkipPaths: []string{
			"/health",
			"/metrics",
			"/swagger",
		},
		APIKeys: map[string]string{
			"dev-key-123":  "developer",
			"test-key-456": "tester",
			"admin-key-789": "admin",
		},
	}
	
	middleware := NewAuthMiddleware(config)
	return middleware.Handler()
}

// RequireJWTAuth creates JWT authentication middleware
func RequireJWTAuth(jwtSecret string) gin.HandlerFunc {
	config := &AuthConfig{
		AuthType:  "jwt",
		JWTSecret: jwtSecret,
		SkipPaths: []string{
			"/health",
			"/metrics",
			"/auth/login",
		},
	}
	
	middleware := NewAuthMiddleware(config)
	return middleware.Handler()
}