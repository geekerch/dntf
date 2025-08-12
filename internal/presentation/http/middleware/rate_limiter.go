package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"notification/pkg/logger"
)

// RateLimiterConfig holds rate limiting configuration
type RateLimiterConfig struct {
	// Requests per minute per IP
	RequestsPerMinute int
	// Requests per minute per user (if authenticated)
	RequestsPerMinutePerUser int
	// Burst size (maximum requests in a short burst)
	BurstSize int
	// Skip rate limiting for these paths
	SkipPaths []string
	// Skip rate limiting for these IPs (whitelist)
	WhitelistIPs []string
	// Custom key generator function
	KeyGenerator func(*gin.Context) string
}

// TokenBucket represents a token bucket for rate limiting
type TokenBucket struct {
	capacity     int
	tokens       int
	refillRate   int // tokens per minute
	lastRefill   time.Time
	mutex        sync.Mutex
}

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	config  *RateLimiterConfig
	buckets map[string]*TokenBucket
	mutex   sync.RWMutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config *RateLimiterConfig) *RateLimiter {
	if config == nil {
		config = &RateLimiterConfig{
			RequestsPerMinute:        60,
			RequestsPerMinutePerUser: 120,
			BurstSize:               10,
			SkipPaths:               []string{"/health", "/metrics"},
			WhitelistIPs:            []string{"127.0.0.1", "::1"},
		}
	}

	rl := &RateLimiter{
		config:  config,
		buckets: make(map[string]*TokenBucket),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Handler returns the rate limiting middleware handler
func (rl *RateLimiter) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip rate limiting for certain paths
		if rl.shouldSkipPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Skip rate limiting for whitelisted IPs
		if rl.isWhitelistedIP(c.ClientIP()) {
			c.Next()
			return
		}

		// Generate rate limiting key
		key := rl.generateKey(c)
		
		// Check rate limit
		allowed, remaining, resetTime := rl.checkRateLimit(key, c)
		
		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.getLimit(c)))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

		if !allowed {
			logger.Warn("Rate limit exceeded",
				zap.String("key", key),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("client_ip", c.ClientIP()),
				zap.Int("remaining", remaining))

			c.JSON(http.StatusTooManyRequests, ErrorResponse{
				Error:   "Rate limit exceeded",
				Details: fmt.Sprintf("Too many requests. Try again in %v", time.Until(resetTime)),
				Code:    "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		logger.Debug("Rate limit check passed",
			zap.String("key", key),
			zap.Int("remaining", remaining))

		c.Next()
	}
}

// shouldSkipPath checks if rate limiting should be skipped for the given path
func (rl *RateLimiter) shouldSkipPath(path string) bool {
	for _, skipPath := range rl.config.SkipPaths {
		if path == skipPath || (len(path) > len(skipPath) && path[:len(skipPath)] == skipPath) {
			return true
		}
	}
	return false
}

// isWhitelistedIP checks if the IP is whitelisted
func (rl *RateLimiter) isWhitelistedIP(ip string) bool {
	for _, whitelistIP := range rl.config.WhitelistIPs {
		if ip == whitelistIP {
			return true
		}
	}
	return false
}

// generateKey generates a rate limiting key
func (rl *RateLimiter) generateKey(c *gin.Context) string {
	if rl.config.KeyGenerator != nil {
		return rl.config.KeyGenerator(c)
	}

	// Use user ID if authenticated, otherwise use IP
	if userID, exists := c.Get("user_id"); exists {
		return fmt.Sprintf("user:%s", userID)
	}
	
	return fmt.Sprintf("ip:%s", c.ClientIP())
}

// getLimit returns the rate limit for the context
func (rl *RateLimiter) getLimit(c *gin.Context) int {
	// Use per-user limit if authenticated
	if _, exists := c.Get("user_id"); exists {
		return rl.config.RequestsPerMinutePerUser
	}
	
	return rl.config.RequestsPerMinute
}

// checkRateLimit checks if the request is within rate limits
func (rl *RateLimiter) checkRateLimit(key string, c *gin.Context) (allowed bool, remaining int, resetTime time.Time) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	bucket, exists := rl.buckets[key]
	if !exists {
		limit := rl.getLimit(c)
		bucket = &TokenBucket{
			capacity:   rl.config.BurstSize,
			tokens:     rl.config.BurstSize,
			refillRate: limit,
			lastRefill: time.Now(),
		}
		rl.buckets[key] = bucket
	}

	return bucket.consume()
}

// consume attempts to consume a token from the bucket
func (tb *TokenBucket) consume() (allowed bool, remaining int, resetTime time.Time) {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	now := time.Now()
	
	// Refill tokens based on time elapsed
	elapsed := now.Sub(tb.lastRefill)
	tokensToAdd := int(elapsed.Minutes() * float64(tb.refillRate))
	
	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}

	// Calculate reset time (when bucket will be full)
	var nextResetTime time.Time
	if tb.tokens < tb.capacity {
		tokensNeeded := tb.capacity - tb.tokens
		minutesToFull := float64(tokensNeeded) / float64(tb.refillRate)
		nextResetTime = now.Add(time.Duration(minutesToFull * float64(time.Minute)))
	} else {
		nextResetTime = now.Add(time.Minute)
	}

	// Try to consume a token
	if tb.tokens > 0 {
		tb.tokens--
		return true, tb.tokens, nextResetTime
	}

	return false, 0, nextResetTime
}

// cleanup removes old buckets to prevent memory leaks
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		now := time.Now()
		
		for key, bucket := range rl.buckets {
			bucket.mutex.Lock()
			// Remove buckets that haven't been used for 10 minutes
			if now.Sub(bucket.lastRefill) > 10*time.Minute {
				delete(rl.buckets, key)
			}
			bucket.mutex.Unlock()
		}
		
		rl.mutex.Unlock()
	}
}

// DefaultRateLimiter creates a rate limiter with default settings
func DefaultRateLimiter() gin.HandlerFunc {
	config := &RateLimiterConfig{
		RequestsPerMinute:        60,  // 60 requests per minute per IP
		RequestsPerMinutePerUser: 120, // 120 requests per minute per authenticated user
		BurstSize:               10,   // Allow burst of 10 requests
		SkipPaths: []string{
			"/health",
			"/metrics",
		},
		WhitelistIPs: []string{
			"127.0.0.1",
			"::1",
		},
	}
	
	limiter := NewRateLimiter(config)
	return limiter.Handler()
}

// StrictRateLimiter creates a more restrictive rate limiter
func StrictRateLimiter() gin.HandlerFunc {
	config := &RateLimiterConfig{
		RequestsPerMinute:        30,  // 30 requests per minute per IP
		RequestsPerMinutePerUser: 60,  // 60 requests per minute per authenticated user
		BurstSize:               5,    // Allow burst of 5 requests
		SkipPaths: []string{
			"/health",
		},
		WhitelistIPs: []string{
			"127.0.0.1",
		},
	}
	
	limiter := NewRateLimiter(config)
	return limiter.Handler()
}