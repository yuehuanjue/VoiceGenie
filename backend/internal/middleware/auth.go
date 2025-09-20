package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"voicegenie/internal/config"
	"voicegenie/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// Claims represents JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Type     string `json:"type"` // "user", "guest", "admin"
	jwt.RegisteredClaims
}

// AuthRequired returns a middleware that validates JWT tokens
func AuthRequired(jwtConfig config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from header
		token := extractToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":      40100,
				"message":   "Missing authorization token",
				"timestamp": time.Now().Unix(),
			})
			return
		}

		// Parse and validate token
		claims, err := parseToken(token, jwtConfig.Secret)
		if err != nil {
			logger.WithFields(map[string]interface{}{
				"error": err.Error(),
				"token": token[:min(len(token), 20)] + "...",
			}).Warn("Invalid token")

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":      40101,
				"message":   "Invalid or expired token",
				"timestamp": time.Now().Unix(),
			})
			return
		}

		// Check if token is expired
		if claims.ExpiresAt.Time.Before(time.Now()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":      40102,
				"message":   "Token has expired",
				"timestamp": time.Now().Unix(),
			})
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_type", claims.Type)
		c.Set("claims", claims)

		c.Next()
	}
}

// OptionalAuth returns a middleware that optionally validates JWT tokens
// If token is present, it validates it, otherwise continues without auth
func OptionalAuth(jwtConfig config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.Next()
			return
		}

		claims, err := parseToken(token, jwtConfig.Secret)
		if err == nil && claims.ExpiresAt.Time.After(time.Now()) {
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("user_type", claims.Type)
			c.Set("claims", claims)
		}

		c.Next()
	}
}

// AdminRequired returns a middleware that requires admin privileges
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists || userType != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":      40300,
				"message":   "Admin privileges required",
				"timestamp": time.Now().Unix(),
			})
			return
		}
		c.Next()
	}
}

// UserTypeRequired returns a middleware that requires specific user types
func UserTypeRequired(allowedTypes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":      40100,
				"message":   "Authentication required",
				"timestamp": time.Now().Unix(),
			})
			return
		}

		userTypeStr := userType.(string)
		for _, allowedType := range allowedTypes {
			if userTypeStr == allowedType {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"code":      40301,
			"message":   "Insufficient privileges",
			"timestamp": time.Now().Unix(),
		})
	}
}

// extractToken extracts JWT token from request
func extractToken(c *gin.Context) string {
	// Try Authorization header first
	bearerToken := c.GetHeader("Authorization")
	if len(bearerToken) > 7 && strings.ToUpper(bearerToken[0:6]) == "BEARER" {
		return bearerToken[7:]
	}

	// Try query parameter
	token := c.Query("token")
	if token != "" {
		return token
	}

	// Try cookie
	cookie, err := c.Cookie("token")
	if err == nil && cookie != "" {
		return cookie
	}

	return ""
}

// parseToken parses and validates JWT token
func parseToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GenerateToken generates a JWT token for a user
func GenerateToken(userID, username, userType string, jwtConfig config.JWTConfig) (string, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(jwtConfig.ExpirationHours) * time.Hour)

	claims := Claims{
		UserID:   userID,
		Username: username,
		Type:     userType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jwtConfig.Issuer,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtConfig.Secret))
}

// GenerateRefreshToken generates a refresh token
func GenerateRefreshToken(userID string, jwtConfig config.JWTConfig) (string, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(jwtConfig.RefreshExpirationDays) * 24 * time.Hour)

	claims := Claims{
		UserID: userID,
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jwtConfig.Issuer,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtConfig.Secret))
}

// ValidateRefreshToken validates a refresh token
func ValidateRefreshToken(tokenString, secret string) (*Claims, error) {
	claims, err := parseToken(tokenString, secret)
	if err != nil {
		return nil, err
	}

	if claims.Type != "refresh" {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}