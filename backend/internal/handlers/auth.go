package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"voicegenie/internal/config"
	"voicegenie/internal/middleware"
	"voicegenie/pkg/database"
	"voicegenie/pkg/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthHandler handles authentication related requests
type AuthHandler struct {
	db     *database.DB
	config *config.Config
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *database.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		db:     db,
		config: cfg,
	}
}

// PhoneLoginRequest represents phone login request
type PhoneLoginRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

// WechatLoginRequest represents WeChat login request
type WechatLoginRequest struct {
	Code     string                 `json:"code" binding:"required"`
	UserInfo map[string]interface{} `json:"userInfo" binding:"required"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Token        string             `json:"token"`
	RefreshToken string             `json:"refresh_token"`
	ExpiresIn    int                `json:"expires_in"`
	UserInfo     *database.User     `json:"user_info"`
}

// PhoneLogin handles phone number login
func (h *AuthHandler) PhoneLogin(c *gin.Context) {
	var req PhoneLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      40000,
			"message":   "Invalid request parameters",
			"details":   err.Error(),
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Validate phone number format
	if !isValidPhoneNumber(req.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      40001,
			"message":   "Invalid phone number format",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Verify SMS code
	if !h.verifySMSCode(req.Phone, req.Code) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      40002,
			"message":   "Invalid or expired verification code",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Find or create user
	var user database.User
	result := h.db.Where("phone = ?", req.Phone).First(&user)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Create new user
			user = database.User{
				Username:  generateUsername(req.Phone),
				Nickname:  "手机用户",
				Phone:     req.Phone,
				LoginType: "phone",
				Status:    "active",
			}

			if err := h.db.Create(&user).Error; err != nil {
				logger.WithError(err).Error("Failed to create user")
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":      50000,
					"message":   "Failed to create user account",
					"timestamp": time.Now().Unix(),
				})
				return
			}
		} else {
			logger.WithError(result.Error).Error("Database error during phone login")
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":      50001,
				"message":   "Database error",
				"timestamp": time.Now().Unix(),
			})
			return
		}
	}

	// Update last login info
	user.LastLoginAt = &[]time.Time{time.Now()}[0]
	user.LastLoginIP = c.ClientIP()
	h.db.Save(&user)

	// Generate tokens
	token, err := middleware.GenerateToken(
		strconv.Itoa(int(user.ID)),
		user.Username,
		"user",
		h.config.JWT,
	)
	if err != nil {
		logger.WithError(err).Error("Failed to generate access token")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      50002,
			"message":   "Failed to generate authentication token",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	refreshToken, err := middleware.GenerateRefreshToken(
		strconv.Itoa(int(user.ID)),
		h.config.JWT,
	)
	if err != nil {
		logger.WithError(err).Error("Failed to generate refresh token")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      50003,
			"message":   "Failed to generate refresh token",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Clear sensitive data
	user.Password = ""

	// Return successful response
	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"message":   "Login successful",
		"data": AuthResponse{
			Token:        token,
			RefreshToken: refreshToken,
			ExpiresIn:    h.config.JWT.ExpirationHours * 3600,
			UserInfo:     &user,
		},
		"timestamp": time.Now().Unix(),
	})

	logger.WithFields(map[string]interface{}{
		"user_id": user.ID,
		"phone":   req.Phone,
		"ip":      c.ClientIP(),
	}).Info("User logged in successfully via phone")
}

// SendSMSCode handles SMS code sending
func (h *AuthHandler) SendSMSCode(c *gin.Context) {
	var req struct {
		Phone string `json:"phone" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      40000,
			"message":   "Invalid request parameters",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Validate phone number
	if !isValidPhoneNumber(req.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      40001,
			"message":   "Invalid phone number format",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Check rate limiting for SMS sending
	if !h.checkSMSRateLimit(req.Phone) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"code":      42900,
			"message":   "SMS sending too frequently, please try again later",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Generate and send SMS code
	code := generateSMSCode()
	if err := h.sendSMSCode(req.Phone, code); err != nil {
		logger.WithError(err).Error("Failed to send SMS code")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      50004,
			"message":   "Failed to send verification code",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Store code in cache/database for verification
	h.storeSMSCode(req.Phone, code)

	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"message":   "Verification code sent successfully",
		"data": gin.H{
			"expire_time": time.Now().Add(5 * time.Minute).Unix(),
		},
		"timestamp": time.Now().Unix(),
	})

	logger.WithFields(map[string]interface{}{
		"phone": req.Phone,
		"ip":    c.ClientIP(),
	}).Info("SMS verification code sent")
}

// WechatLogin handles WeChat login
func (h *AuthHandler) WechatLogin(c *gin.Context) {
	var req WechatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      40000,
			"message":   "Invalid request parameters",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Exchange code for WeChat user info
	wechatUser, err := h.getWechatUserInfo(req.Code)
	if err != nil {
		logger.WithError(err).Error("Failed to get WeChat user info")
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      40003,
			"message":   "Failed to authenticate with WeChat",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Find or create user
	var user database.User
	result := h.db.Where("wechat_id = ?", wechatUser.OpenID).First(&user)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Create new user
			user = database.User{
				Username:  generateUsername(wechatUser.OpenID),
				Nickname:  wechatUser.Nickname,
				Avatar:    wechatUser.Avatar,
				WechatID:  wechatUser.OpenID,
				LoginType: "wechat",
				Status:    "active",
			}

			if err := h.db.Create(&user).Error; err != nil {
				logger.WithError(err).Error("Failed to create WeChat user")
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":      50000,
					"message":   "Failed to create user account",
					"timestamp": time.Now().Unix(),
				})
				return
			}
		} else {
			logger.WithError(result.Error).Error("Database error during WeChat login")
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":      50001,
				"message":   "Database error",
				"timestamp": time.Now().Unix(),
			})
			return
		}
	}

	// Update last login info
	user.LastLoginAt = &[]time.Time{time.Now()}[0]
	user.LastLoginIP = c.ClientIP()
	h.db.Save(&user)

	// Generate tokens
	token, err := middleware.GenerateToken(
		strconv.Itoa(int(user.ID)),
		user.Username,
		"user",
		h.config.JWT,
	)
	if err != nil {
		logger.WithError(err).Error("Failed to generate access token")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      50002,
			"message":   "Failed to generate authentication token",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	refreshToken, err := middleware.GenerateRefreshToken(
		strconv.Itoa(int(user.ID)),
		h.config.JWT,
	)
	if err != nil {
		logger.WithError(err).Error("Failed to generate refresh token")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      50003,
			"message":   "Failed to generate refresh token",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Clear sensitive data
	user.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"message":   "WeChat login successful",
		"data": AuthResponse{
			Token:        token,
			RefreshToken: refreshToken,
			ExpiresIn:    h.config.JWT.ExpirationHours * 3600,
			UserInfo:     &user,
		},
		"timestamp": time.Now().Unix(),
	})

	logger.WithFields(map[string]interface{}{
		"user_id":   user.ID,
		"wechat_id": wechatUser.OpenID,
		"ip":        c.ClientIP(),
	}).Info("User logged in successfully via WeChat")
}

// GuestLogin handles guest login
func (h *AuthHandler) GuestLogin(c *gin.Context) {
	// Create guest user
	user := database.User{
		Username:  generateGuestUsername(),
		Nickname:  "游客用户",
		LoginType: "guest",
		Status:    "active",
	}

	if err := h.db.Create(&user).Error; err != nil {
		logger.WithError(err).Error("Failed to create guest user")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      50000,
			"message":   "Failed to create guest account",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Update login info
	user.LastLoginAt = &[]time.Time{time.Now()}[0]
	user.LastLoginIP = c.ClientIP()
	h.db.Save(&user)

	// Generate tokens
	token, err := middleware.GenerateToken(
		strconv.Itoa(int(user.ID)),
		user.Username,
		"guest",
		h.config.JWT,
	)
	if err != nil {
		logger.WithError(err).Error("Failed to generate guest token")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      50002,
			"message":   "Failed to generate authentication token",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"message":   "Guest login successful",
		"data": AuthResponse{
			Token:     token,
			ExpiresIn: h.config.JWT.ExpirationHours * 3600,
			UserInfo:  &user,
		},
		"timestamp": time.Now().Unix(),
	})

	logger.WithFields(map[string]interface{}{
		"user_id": user.ID,
		"ip":      c.ClientIP(),
	}).Info("Guest user created and logged in")
}

// VerifyToken handles token verification
func (h *AuthHandler) VerifyToken(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":      40100,
			"message":   "Invalid token",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Get user info
	var user database.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":      40101,
			"message":   "User not found",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Clear sensitive data
	user.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"message":   "Token is valid",
		"data":      user,
		"timestamp": time.Now().Unix(),
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      40000,
			"message":   "Invalid request parameters",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Validate refresh token
	claims, err := middleware.ValidateRefreshToken(req.RefreshToken, h.config.JWT.Secret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":      40102,
			"message":   "Invalid refresh token",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Get user info
	var user database.User
	if err := h.db.First(&user, claims.UserID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":      40103,
			"message":   "User not found",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Generate new access token
	newToken, err := middleware.GenerateToken(
		claims.UserID,
		user.Username,
		"user",
		h.config.JWT,
	)
	if err != nil {
		logger.WithError(err).Error("Failed to generate new access token")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      50002,
			"message":   "Failed to generate new token",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"message":   "Token refreshed successfully",
		"data": gin.H{
			"token":      newToken,
			"expires_in": h.config.JWT.ExpirationHours * 3600,
		},
		"timestamp": time.Now().Unix(),
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	userID := c.GetString("user_id")

	logger.WithFields(map[string]interface{}{
		"user_id": userID,
		"ip":      c.ClientIP(),
	}).Info("User logged out")

	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"message":   "Logout successful",
		"timestamp": time.Now().Unix(),
	})
}

// Helper functions (implementations would be more complex in production)

func isValidPhoneNumber(phone string) bool {
	// Simple validation - in production, use proper regex
	return len(phone) >= 10 && len(phone) <= 15
}

func generateUsername(source string) string {
	return fmt.Sprintf("user_%d", time.Now().Unix())
}

func generateGuestUsername() string {
	return fmt.Sprintf("guest_%d", time.Now().Unix())
}

func generateSMSCode() string {
	return "123456" // In production, generate random 6-digit code
}

func (h *AuthHandler) verifySMSCode(phone, code string) bool {
	// In production, verify against stored code in Redis/database
	return code == "123456"
}

func (h *AuthHandler) checkSMSRateLimit(phone string) bool {
	// In production, implement proper rate limiting
	return true
}

func (h *AuthHandler) sendSMSCode(phone, code string) error {
	// In production, integrate with SMS provider (Twilio, Aliyun, etc.)
	logger.Infof("Sending SMS code %s to %s", code, phone)
	return nil
}

func (h *AuthHandler) storeSMSCode(phone, code string) {
	// In production, store in Redis with expiration
	logger.Infof("Storing SMS code for phone %s", phone)
}

// WechatUser represents WeChat user info
type WechatUser struct {
	OpenID   string `json:"openid"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"headimgurl"`
}

func (h *AuthHandler) getWechatUserInfo(code string) (*WechatUser, error) {
	// In production, exchange code with WeChat API
	return &WechatUser{
		OpenID:   fmt.Sprintf("wx_%d", time.Now().Unix()),
		Nickname: "微信用户",
		Avatar:   "",
	}, nil
}