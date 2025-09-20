package middleware

import (
	"net/http"
	"sync"
	"time"

	"voicegenie/internal/config"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	tokens      int
	maxTokens   int
	refillRate  time.Duration
	lastRefill  time.Time
	mutex       sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxTokens int, refillRate time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow() bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)

	// Refill tokens based on elapsed time
	tokensToAdd := int(elapsed / rl.refillRate)
	if tokensToAdd > 0 {
		rl.tokens = min(rl.maxTokens, rl.tokens+tokensToAdd)
		rl.lastRefill = now
	}

	// Check if we have tokens available
	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

// Global rate limiters for different rate limiting strategies
var (
	globalLimiters = make(map[string]*RateLimiter)
	limiterMutex   = sync.RWMutex{}
)

// RateLimit returns a middleware that implements rate limiting
func RateLimit(config config.RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client identifier (IP or user ID)
		identifier := getClientIdentifier(c)

		// Get or create rate limiter for this client
		limiter := getOrCreateLimiter(identifier, config)

		// Check if request is allowed
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":      42900,
				"message":   "Rate limit exceeded",
				"retry_after": int(config.WindowDuration / time.Second),
				"timestamp": time.Now().Unix(),
			})
			return
		}

		c.Next()
	}
}

// APIRateLimit returns a more sophisticated rate limiter for API endpoints
func APIRateLimit(requestsPerMinute int) gin.HandlerFunc {
	limiters := make(map[string]*RateLimiter)
	mutex := sync.RWMutex{}

	return func(c *gin.Context) {
		identifier := getClientIdentifier(c)

		mutex.RLock()
		limiter, exists := limiters[identifier]
		mutex.RUnlock()

		if !exists {
			mutex.Lock()
			// Double-check after acquiring write lock
			if limiter, exists = limiters[identifier]; !exists {
				limiter = NewRateLimiter(requestsPerMinute, time.Minute/time.Duration(requestsPerMinute))
				limiters[identifier] = limiter
			}
			mutex.Unlock()
		}

		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":         42900,
				"message":      "API rate limit exceeded",
				"limit":        requestsPerMinute,
				"window":       "1 minute",
				"retry_after":  60,
				"timestamp":    time.Now().Unix(),
			})
			return
		}

		c.Next()
	}
}

// UserRateLimit implements per-user rate limiting
func UserRateLimit(requestsPerHour int) gin.HandlerFunc {
	limiters := make(map[string]*RateLimiter)
	mutex := sync.RWMutex{}

	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			// Skip rate limiting for unauthenticated requests
			c.Next()
			return
		}

		identifier := userID.(string)

		mutex.RLock()
		limiter, exists := limiters[identifier]
		mutex.RUnlock()

		if !exists {
			mutex.Lock()
			if limiter, exists = limiters[identifier]; !exists {
				limiter = NewRateLimiter(requestsPerHour, time.Hour/time.Duration(requestsPerHour))
				limiters[identifier] = limiter
			}
			mutex.Unlock()
		}

		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":        42901,
				"message":     "User rate limit exceeded",
				"limit":       requestsPerHour,
				"window":      "1 hour",
				"retry_after": 3600,
				"timestamp":   time.Now().Unix(),
			})
			return
		}

		c.Next()
	}
}

// ExpensiveOperationLimit limits expensive operations (like AI requests)
func ExpensiveOperationLimit(requestsPerMinute int) gin.HandlerFunc {
	limiters := make(map[string]*RateLimiter)
	mutex := sync.RWMutex{}

	return func(c *gin.Context) {
		identifier := getClientIdentifier(c)

		mutex.RLock()
		limiter, exists := limiters[identifier]
		mutex.RUnlock()

		if !exists {
			mutex.Lock()
			if limiter, exists = limiters[identifier]; !exists {
				limiter = NewRateLimiter(requestsPerMinute, time.Minute/time.Duration(requestsPerMinute))
				limiters[identifier] = limiter
			}
			mutex.Unlock()
		}

		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":        42902,
				"message":     "Operation rate limit exceeded",
				"description": "This operation is resource-intensive and has stricter limits",
				"limit":       requestsPerMinute,
				"window":      "1 minute",
				"retry_after": 60,
				"timestamp":   time.Now().Unix(),
			})
			return
		}

		c.Next()
	}
}

// getClientIdentifier returns a unique identifier for the client
func getClientIdentifier(c *gin.Context) string {
	// Try to get user ID first (for authenticated requests)
	if userID, exists := c.Get("user_id"); exists {
		return "user:" + userID.(string)
	}

	// Fall back to IP address
	return "ip:" + c.ClientIP()
}

// getOrCreateLimiter gets or creates a rate limiter for the given identifier
func getOrCreateLimiter(identifier string, config config.RateLimitConfig) *RateLimiter {
	limiterMutex.RLock()
	limiter, exists := globalLimiters[identifier]
	limiterMutex.RUnlock()

	if !exists {
		limiterMutex.Lock()
		// Double-check after acquiring write lock
		if limiter, exists = globalLimiters[identifier]; !exists {
			refillInterval := config.WindowDuration / time.Duration(config.MaxRequests)
			limiter = NewRateLimiter(config.MaxRequests, refillInterval)
			globalLimiters[identifier] = limiter
		}
		limiterMutex.Unlock()
	}

	return limiter
}

// CleanupOldLimiters removes old rate limiters to prevent memory leaks
func CleanupOldLimiters() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			limiterMutex.Lock()
			now := time.Now()
			for identifier, limiter := range globalLimiters {
				// Remove limiters that haven't been used for more than 2 hours
				if now.Sub(limiter.lastRefill) > 2*time.Hour {
					delete(globalLimiters, identifier)
				}
			}
			limiterMutex.Unlock()
		}
	}()
}