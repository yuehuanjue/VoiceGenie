package middleware

import (
	"net/http"
	"time"

	"voicegenie/pkg/logger"

	"github.com/gin-gonic/gin"
)

// Logger returns a gin.HandlerFunc for logging HTTP requests
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status and size
		status := c.Writer.Status()
		bodySize := c.Writer.Size()

		// Build query string
		if raw != "" {
			path = path + "?" + raw
		}

		// Get client IP
		clientIP := c.ClientIP()

		// Get request ID
		requestID := c.GetString("X-Request-ID")

		// Log request
		logger.WithFields(map[string]interface{}{
			"status":     status,
			"method":     c.Request.Method,
			"path":       path,
			"ip":         clientIP,
			"latency":    latency,
			"user_agent": c.Request.UserAgent(),
			"body_size":  bodySize,
			"request_id": requestID,
		}).Info("HTTP Request")

		// Log errors if status >= 400
		if status >= 400 {
			if len(c.Errors) > 0 {
				logger.WithFields(map[string]interface{}{
					"request_id": requestID,
					"errors":     c.Errors.String(),
				}).Error("Request errors")
			}
		}
	}
}

// RequestID adds a request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Set("X-Request-ID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// Security adds security headers
func Security() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}

// ErrorHandler handles errors and formats responses
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Process any errors that occurred
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			// Get request ID for error tracking
			requestID := c.GetString("X-Request-ID")

			// Log the error
			logger.WithFields(map[string]interface{}{
				"request_id": requestID,
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
				"error":      err.Error(),
			}).Error("Request error")

			// Send error response if not already sent
			if !c.Writer.Written() {
				switch e := err.Err.(type) {
				case *ValidationError:
					c.JSON(http.StatusBadRequest, gin.H{
						"code":       40000,
						"message":    e.Message,
						"details":    e.Details,
						"request_id": requestID,
						"timestamp":  time.Now().Unix(),
					})
				case *AuthenticationError:
					c.JSON(http.StatusUnauthorized, gin.H{
						"code":       40100,
						"message":    e.Message,
						"request_id": requestID,
						"timestamp":  time.Now().Unix(),
					})
				case *AuthorizationError:
					c.JSON(http.StatusForbidden, gin.H{
						"code":       40300,
						"message":    e.Message,
						"request_id": requestID,
						"timestamp":  time.Now().Unix(),
					})
				case *NotFoundError:
					c.JSON(http.StatusNotFound, gin.H{
						"code":       40400,
						"message":    e.Message,
						"request_id": requestID,
						"timestamp":  time.Now().Unix(),
					})
				case *BusinessError:
					c.JSON(http.StatusBadRequest, gin.H{
						"code":       e.Code,
						"message":    e.Message,
						"request_id": requestID,
						"timestamp":  time.Now().Unix(),
					})
				default:
					c.JSON(http.StatusInternalServerError, gin.H{
						"code":       50000,
						"message":    "Internal server error",
						"request_id": requestID,
						"timestamp":  time.Now().Unix(),
					})
				}
			}
		}
	}
}

// RequestSizeLimit limits the size of request body
func RequestSizeLimit(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"code":      41300,
				"message":   "Request entity too large",
				"max_size":  maxSize,
				"timestamp": time.Now().Unix(),
			})
			return
		}
		c.Next()
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	// Simple implementation using timestamp and random
	return time.Now().Format("20060102150405") + randomString(6)
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// Custom error types
type ValidationError struct {
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

type AuthenticationError struct {
	Message string `json:"message"`
}

func (e *AuthenticationError) Error() string {
	return e.Message
}

type AuthorizationError struct {
	Message string `json:"message"`
}

func (e *AuthorizationError) Error() string {
	return e.Message
}

type NotFoundError struct {
	Message string `json:"message"`
}

func (e *NotFoundError) Error() string {
	return e.Message
}

type BusinessError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *BusinessError) Error() string {
	return e.Message
}