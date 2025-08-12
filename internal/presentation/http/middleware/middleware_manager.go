package middleware

import (
	"github.com/gin-gonic/gin"
)

// MiddlewareConfig holds configuration for all middleware
type MiddlewareConfig struct {
	// Environment: "development", "staging", "production"
	Environment string
	
	// Authentication configuration
	Auth *AuthConfig
	
	// Rate limiting configuration
	RateLimit *RateLimiterConfig
	
	// CORS configuration
	CORS *CORSConfig
	
	// Security configuration
	Security *SecurityConfig
	
	// IP whitelist configuration
	IPWhitelist *IPWhitelistConfig
	
	// Basic auth configuration (for admin endpoints)
	BasicAuth *BasicAuthConfig
	
	// Enable/disable specific middleware
	EnableAuth      bool
	EnableRateLimit bool
	EnableCORS      bool
	EnableSecurity  bool
	EnableIPWhitelist bool
	EnableBasicAuth   bool
}

// MiddlewareManager manages all middleware setup
type MiddlewareManager struct {
	config *MiddlewareConfig
}

// NewMiddlewareManager creates a new middleware manager
func NewMiddlewareManager(config *MiddlewareConfig) *MiddlewareManager {
	if config == nil {
		config = DefaultMiddlewareConfig()
	}
	return &MiddlewareManager{config: config}
}

// SetupMiddleware sets up all middleware for the given router
func (mm *MiddlewareManager) SetupMiddleware(router *gin.Engine) {
	// Core middleware (always enabled)
	router.Use(RequestLogger())
	router.Use(RequestID())
	router.Use(ErrorHandler())

	// Security middleware
	if mm.config.EnableSecurity {
		if mm.config.Security != nil {
			securityMiddleware := NewSecurityMiddleware(mm.config.Security)
			router.Use(securityMiddleware.Handler())
		} else {
			router.Use(mm.getDefaultSecurityMiddleware())
		}
	}

	// CORS middleware
	if mm.config.EnableCORS {
		if mm.config.CORS != nil {
			corsMiddleware := NewCORSMiddleware(mm.config.CORS)
			router.Use(corsMiddleware.Handler())
		} else {
			router.Use(mm.getDefaultCORSMiddleware())
		}
	}

	// IP whitelist middleware (if configured)
	if mm.config.EnableIPWhitelist && mm.config.IPWhitelist != nil {
		router.Use(IPWhitelistMiddleware(mm.config.IPWhitelist))
	}

	// Rate limiting middleware
	if mm.config.EnableRateLimit {
		if mm.config.RateLimit != nil {
			rateLimiter := NewRateLimiter(mm.config.RateLimit)
			router.Use(rateLimiter.Handler())
		} else {
			router.Use(DefaultRateLimiter())
		}
	}

	// Authentication middleware (applied to protected routes)
	if mm.config.EnableAuth && mm.config.Auth != nil {
		authMiddleware := NewAuthMiddleware(mm.config.Auth)
		// Note: Auth middleware is typically applied to specific route groups
		// rather than globally. See SetupProtectedRoutes method.
		router.Use(authMiddleware.Handler())
	}
}

// SetupProtectedRoutes sets up middleware for protected routes
func (mm *MiddlewareManager) SetupProtectedRoutes(routeGroup *gin.RouterGroup) {
	// Authentication middleware for protected routes
	if mm.config.EnableAuth && mm.config.Auth != nil {
		authMiddleware := NewAuthMiddleware(mm.config.Auth)
		routeGroup.Use(authMiddleware.Handler())
	}

	// Additional rate limiting for protected routes (if needed)
	if mm.config.EnableRateLimit {
		routeGroup.Use(StrictRateLimiter())
	}
}

// SetupAdminRoutes sets up middleware for admin routes
func (mm *MiddlewareManager) SetupAdminRoutes(routeGroup *gin.RouterGroup) {
	// Basic auth for admin routes
	if mm.config.EnableBasicAuth && mm.config.BasicAuth != nil {
		routeGroup.Use(BasicAuthMiddleware(mm.config.BasicAuth))
	}

	// Strict rate limiting for admin routes
	if mm.config.EnableRateLimit {
		routeGroup.Use(StrictRateLimiter())
	}

	// IP whitelist for admin routes (if configured)
	if mm.config.EnableIPWhitelist && mm.config.IPWhitelist != nil {
		routeGroup.Use(IPWhitelistMiddleware(mm.config.IPWhitelist))
	}
}

// getDefaultSecurityMiddleware returns default security middleware based on environment
func (mm *MiddlewareManager) getDefaultSecurityMiddleware() gin.HandlerFunc {
	switch mm.config.Environment {
	case "development":
		return DevelopmentSecurity()
	case "production":
		return StrictSecurity([]string{}) // Add your allowed hosts
	default:
		return DefaultSecurity()
	}
}

// getDefaultCORSMiddleware returns default CORS middleware based on environment
func (mm *MiddlewareManager) getDefaultCORSMiddleware() gin.HandlerFunc {
	switch mm.config.Environment {
	case "development":
		return DevelopmentCORS()
	case "production":
		return ProductionCORS([]string{}) // Add your allowed origins
	default:
		return DefaultCORS()
	}
}

// DefaultMiddlewareConfig returns a default middleware configuration
func DefaultMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		Environment:       "development",
		EnableAuth:        false, // Disabled by default, enable per route group
		EnableRateLimit:   true,
		EnableCORS:        true,
		EnableSecurity:    true,
		EnableIPWhitelist: false,
		EnableBasicAuth:   false,
	}
}

// DevelopmentMiddlewareConfig returns a middleware configuration for development
func DevelopmentMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		Environment:     "development",
		EnableAuth:      false,
		EnableRateLimit: false, // Disabled for easier development
		EnableCORS:      true,
		EnableSecurity:  true,
		EnableIPWhitelist: false,
		EnableBasicAuth:   false,
		CORS:            DevelopmentCORSConfig(),
		Security:        DevelopmentSecurityConfig(),
	}
}

// ProductionMiddlewareConfig returns a middleware configuration for production
func ProductionMiddlewareConfig(allowedOrigins, allowedHosts []string) *MiddlewareConfig {
	return &MiddlewareConfig{
		Environment:     "production",
		EnableAuth:      true,
		EnableRateLimit: true,
		EnableCORS:      true,
		EnableSecurity:  true,
		EnableIPWhitelist: false,
		EnableBasicAuth:   false,
		Auth: &AuthConfig{
			AuthType:  "api-key",
			SkipPaths: []string{"/health", "/metrics"},
			APIKeys: map[string]string{
				// Add your production API keys here
			},
		},
		RateLimit: &RateLimiterConfig{
			RequestsPerMinute:        100,
			RequestsPerMinutePerUser: 200,
			BurstSize:               20,
			SkipPaths:               []string{"/health", "/metrics"},
		},
		CORS:     ProductionCORSConfig(allowedOrigins),
		Security: StrictSecurityConfig(allowedHosts),
	}
}

// StagingMiddlewareConfig returns a middleware configuration for staging
func StagingMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		Environment:     "staging",
		EnableAuth:      true,
		EnableRateLimit: true,
		EnableCORS:      true,
		EnableSecurity:  true,
		EnableIPWhitelist: false,
		EnableBasicAuth:   false,
		Auth: &AuthConfig{
			AuthType:  "api-key",
			SkipPaths: []string{"/health", "/metrics"},
			APIKeys: map[string]string{
				"staging-key-123": "staging-user",
			},
		},
		RateLimit: &RateLimiterConfig{
			RequestsPerMinute:        200,
			RequestsPerMinutePerUser: 400,
			BurstSize:               30,
			SkipPaths:               []string{"/health", "/metrics"},
		},
		CORS:     DefaultCORSConfig(),
		Security: DefaultSecurityConfig(),
	}
}

// TestMiddlewareConfig returns a minimal middleware configuration for testing
func TestMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		Environment:       "test",
		EnableAuth:        false,
		EnableRateLimit:   false,
		EnableCORS:        false,
		EnableSecurity:    false,
		EnableIPWhitelist: false,
		EnableBasicAuth:   false,
	}
}

// AdminMiddlewareConfig returns middleware configuration for admin endpoints
func AdminMiddlewareConfig(adminUsers map[string]string, allowedIPs []string) *MiddlewareConfig {
	config := DefaultMiddlewareConfig()
	config.EnableBasicAuth = true
	config.EnableIPWhitelist = len(allowedIPs) > 0
	
	config.BasicAuth = &BasicAuthConfig{
		Users:     adminUsers,
		Realm:     "Admin Area",
		SkipPaths: []string{},
	}
	
	if len(allowedIPs) > 0 {
		config.IPWhitelist = &IPWhitelistConfig{
			AllowedIPs: allowedIPs,
			SkipPaths:  []string{},
		}
	}
	
	return config
}

// APIMiddlewareConfig returns middleware configuration for API endpoints
func APIMiddlewareConfig(apiKeys map[string]string) *MiddlewareConfig {
	config := DefaultMiddlewareConfig()
	config.EnableAuth = true
	
	config.Auth = &AuthConfig{
		AuthType:  "api-key",
		SkipPaths: []string{"/health", "/metrics", "/swagger"},
		APIKeys:   apiKeys,
	}
	
	config.RateLimit = &RateLimiterConfig{
		RequestsPerMinute:        120,
		RequestsPerMinutePerUser: 240,
		BurstSize:               15,
		SkipPaths:               []string{"/health", "/metrics"},
	}
	
	return config
}