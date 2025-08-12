package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"notification/pkg/logger"
)

// SecurityConfig holds security middleware configuration
type SecurityConfig struct {
	// Content Security Policy
	ContentSecurityPolicy string
	// X-Frame-Options header value
	FrameOptions string
	// X-Content-Type-Options header value
	ContentTypeOptions string
	// Referrer-Policy header value
	ReferrerPolicy string
	// Permissions-Policy header value
	PermissionsPolicy string
	// Strict-Transport-Security header value
	StrictTransportSecurity string
	// X-XSS-Protection header value
	XSSProtection string
	// Remove Server header
	RemoveServerHeader bool
	// Force HTTPS
	ForceHTTPS bool
	// Allowed hosts
	AllowedHosts []string
}

// SecurityMiddleware provides security headers and checks
type SecurityMiddleware struct {
	config *SecurityConfig
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(config *SecurityConfig) *SecurityMiddleware {
	if config == nil {
		config = DefaultSecurityConfig()
	}
	return &SecurityMiddleware{config: config}
}

// Handler returns the security middleware handler
func (sm *SecurityMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check allowed hosts
		if !sm.isHostAllowed(c.Request.Host) {
			logger.Warn("Security check failed - host not allowed",
				zap.String("host", c.Request.Host),
				zap.String("client_ip", c.ClientIP()))
			
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Invalid host",
				Code:  "INVALID_HOST",
			})
			c.Abort()
			return
		}

		// Force HTTPS
		if sm.config.ForceHTTPS && c.Request.Header.Get("X-Forwarded-Proto") != "https" && c.Request.TLS == nil {
			httpsURL := "https://" + c.Request.Host + c.Request.RequestURI
			c.Redirect(http.StatusMovedPermanently, httpsURL)
			c.Abort()
			return
		}

		// Set security headers
		sm.setSecurityHeaders(c)

		c.Next()
	}
}

// setSecurityHeaders sets various security headers
func (sm *SecurityMiddleware) setSecurityHeaders(c *gin.Context) {
	// Content Security Policy
	if sm.config.ContentSecurityPolicy != "" {
		c.Header("Content-Security-Policy", sm.config.ContentSecurityPolicy)
	}

	// X-Frame-Options
	if sm.config.FrameOptions != "" {
		c.Header("X-Frame-Options", sm.config.FrameOptions)
	}

	// X-Content-Type-Options
	if sm.config.ContentTypeOptions != "" {
		c.Header("X-Content-Type-Options", sm.config.ContentTypeOptions)
	}

	// Referrer-Policy
	if sm.config.ReferrerPolicy != "" {
		c.Header("Referrer-Policy", sm.config.ReferrerPolicy)
	}

	// Permissions-Policy
	if sm.config.PermissionsPolicy != "" {
		c.Header("Permissions-Policy", sm.config.PermissionsPolicy)
	}

	// Strict-Transport-Security (only for HTTPS)
	if sm.config.StrictTransportSecurity != "" && (c.Request.TLS != nil || c.Request.Header.Get("X-Forwarded-Proto") == "https") {
		c.Header("Strict-Transport-Security", sm.config.StrictTransportSecurity)
	}

	// X-XSS-Protection
	if sm.config.XSSProtection != "" {
		c.Header("X-XSS-Protection", sm.config.XSSProtection)
	}

	// Remove Server header
	if sm.config.RemoveServerHeader {
		c.Header("Server", "")
	}

	// Additional security headers
	c.Header("X-Permitted-Cross-Domain-Policies", "none")
	c.Header("Cross-Origin-Embedder-Policy", "require-corp")
	c.Header("Cross-Origin-Opener-Policy", "same-origin")
	c.Header("Cross-Origin-Resource-Policy", "same-origin")
}

// isHostAllowed checks if the host is allowed
func (sm *SecurityMiddleware) isHostAllowed(host string) bool {
	if len(sm.config.AllowedHosts) == 0 {
		return true // No restrictions
	}

	// Remove port from host if present
	if colonIndex := strings.LastIndex(host, ":"); colonIndex != -1 {
		host = host[:colonIndex]
	}

	for _, allowedHost := range sm.config.AllowedHosts {
		if allowedHost == host {
			return true
		}
		
		// Support wildcard subdomains
		if strings.HasPrefix(allowedHost, "*.") {
			domain := allowedHost[2:]
			if strings.HasSuffix(host, "."+domain) || host == domain {
				return true
			}
		}
	}

	return false
}

// DefaultSecurityConfig returns a default security configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none';",
		FrameOptions:          "DENY",
		ContentTypeOptions:    "nosniff",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		PermissionsPolicy:     "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=(), speaker=()",
		XSSProtection:         "1; mode=block",
		RemoveServerHeader:    true,
		ForceHTTPS:           false,
		AllowedHosts:         []string{},
	}
}

// StrictSecurityConfig returns a strict security configuration
func StrictSecurityConfig(allowedHosts []string) *SecurityConfig {
	return &SecurityConfig{
		ContentSecurityPolicy:   "default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self'; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self';",
		FrameOptions:           "DENY",
		ContentTypeOptions:     "nosniff",
		ReferrerPolicy:         "no-referrer",
		PermissionsPolicy:      "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=(), speaker=(), fullscreen=(), sync-xhr=()",
		StrictTransportSecurity: "max-age=31536000; includeSubDomains; preload",
		XSSProtection:          "1; mode=block",
		RemoveServerHeader:     true,
		ForceHTTPS:            true,
		AllowedHosts:          allowedHosts,
	}
}

// DevelopmentSecurityConfig returns a relaxed security configuration for development
func DevelopmentSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		ContentSecurityPolicy: "default-src 'self' 'unsafe-inline' 'unsafe-eval'; img-src 'self' data: https:; connect-src 'self' ws: wss:;",
		FrameOptions:          "SAMEORIGIN",
		ContentTypeOptions:    "nosniff",
		ReferrerPolicy:        "origin-when-cross-origin",
		XSSProtection:         "1; mode=block",
		RemoveServerHeader:    false,
		ForceHTTPS:           false,
		AllowedHosts:         []string{},
	}
}

// IPWhitelistConfig holds IP whitelist configuration
type IPWhitelistConfig struct {
	AllowedIPs []string
	SkipPaths  []string
}

// IPWhitelistMiddleware provides IP whitelisting functionality
func IPWhitelistMiddleware(config *IPWhitelistConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip IP check for certain paths
		for _, skipPath := range config.SkipPaths {
			if strings.HasPrefix(c.Request.URL.Path, skipPath) {
				c.Next()
				return
			}
		}

		clientIP := c.ClientIP()
		
		// Check if IP is allowed
		allowed := false
		for _, allowedIP := range config.AllowedIPs {
			if allowedIP == clientIP {
				allowed = true
				break
			}
		}

		if !allowed {
			logger.Warn("IP whitelist check failed",
				zap.String("client_ip", clientIP),
				zap.String("path", c.Request.URL.Path))
			
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error: "Access denied",
				Code:  "IP_NOT_ALLOWED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// BasicAuthConfig holds basic authentication configuration
type BasicAuthConfig struct {
	Users     map[string]string // username -> password
	Realm     string
	SkipPaths []string
}

// BasicAuthMiddleware provides basic authentication
func BasicAuthMiddleware(config *BasicAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for certain paths
		for _, skipPath := range config.SkipPaths {
			if strings.HasPrefix(c.Request.URL.Path, skipPath) {
				c.Next()
				return
			}
		}

		username, password, hasAuth := c.Request.BasicAuth()
		
		if !hasAuth {
			c.Header("WWW-Authenticate", `Basic realm="`+config.Realm+`"`)
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error: "Authentication required",
				Code:  "AUTH_REQUIRED",
			})
			c.Abort()
			return
		}

		// Check credentials
		if expectedPassword, exists := config.Users[username]; exists {
			// Use constant-time comparison to prevent timing attacks
			if subtle.ConstantTimeCompare([]byte(password), []byte(expectedPassword)) == 1 {
				c.Set("auth_user", username)
				c.Next()
				return
			}
		}

		logger.Warn("Basic auth failed",
			zap.String("username", username),
			zap.String("client_ip", c.ClientIP()))

		c.Header("WWW-Authenticate", `Basic realm="`+config.Realm+`"`)
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Invalid credentials",
			Code:  "INVALID_CREDENTIALS",
		})
		c.Abort()
	}
}

// Helper functions for easy setup

// DefaultSecurity creates security middleware with default settings
func DefaultSecurity() gin.HandlerFunc {
	middleware := NewSecurityMiddleware(DefaultSecurityConfig())
	return middleware.Handler()
}

// StrictSecurity creates security middleware with strict settings
func StrictSecurity(allowedHosts []string) gin.HandlerFunc {
	middleware := NewSecurityMiddleware(StrictSecurityConfig(allowedHosts))
	return middleware.Handler()
}

// DevelopmentSecurity creates security middleware for development
func DevelopmentSecurity() gin.HandlerFunc {
	middleware := NewSecurityMiddleware(DevelopmentSecurityConfig())
	return middleware.Handler()
}