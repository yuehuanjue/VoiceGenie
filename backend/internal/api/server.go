package api

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"voicegenie/internal/config"
	"voicegenie/internal/handlers"
	"voicegenie/internal/middleware"
	"voicegenie/pkg/database"
	"voicegenie/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
)

// Server represents the HTTP server
type Server struct {
	config *config.Config
	router *gin.Engine
	db     *database.DB
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config) *Server {
	// Set Gin mode based on environment
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create Gin router
	router := gin.New()

	// Initialize database
	db, err := database.New(cfg.Database)
	if err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}

	server := &Server{
		config: cfg,
		router: router,
		db:     db,
	}

	// Initialize handlers
	server.initHandlers()

	// Setup middleware
	server.setupMiddleware()

	// Setup routes
	server.setupRoutes()

	return server
}

// setupMiddleware configures all middleware
func (s *Server) setupMiddleware() {
	// Recovery middleware
	s.router.Use(gin.Recovery())

	// Custom logger middleware
	s.router.Use(middleware.Logger())

	// CORS middleware
	s.router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",  // Frontend dev
			"https://voicegenie.app", // Production frontend
			"https://*.voicegenie.app", // Subdomains
		},
		AllowMethods: []string{
			"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH",
		},
		AllowHeaders: []string{
			"Origin", "Content-Length", "Content-Type", "Authorization",
			"X-Requested-With", "X-Device-Type", "X-App-Version", "X-Timestamp",
		},
		ExposeHeaders: []string{
			"Content-Length", "X-Request-ID", "X-Response-Time",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Request ID middleware
	s.router.Use(middleware.RequestID())

	// Security headers middleware
	s.router.Use(middleware.Security())

	// Rate limiting middleware
	s.router.Use(middleware.RateLimit(s.config.RateLimit))

	// Request size limit middleware
	s.router.Use(middleware.RequestSizeLimit(s.config.App.MaxRequestSize))

	// Error handling middleware
	s.router.Use(middleware.ErrorHandler())

	// Performance monitoring (only in development)
	if s.config.App.Environment != "production" {
		pprof.Register(s.router)
	}
}

// setupRoutes configures all routes
func (s *Server) setupRoutes() {
	// Health check
	s.router.GET("/health", s.healthCheck)
	s.router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// API routes
	api := s.router.Group("/api")
	{
		// System routes
		system := api.Group("/system")
		{
			system.GET("/status", s.getSystemStatus)
			system.GET("/config", s.getSystemConfig)
			system.GET("/update", s.checkUpdate)
			system.POST("/error", s.reportError)
			system.GET("/help", s.getHelp)
		}

		// Authentication routes
		auth := api.Group("/auth")
		{
			auth.POST("/phone/login", s.phoneLogin)
			auth.POST("/sms/send", s.sendSmsCode)
			auth.POST("/wechat/login", s.wechatLogin)
			auth.POST("/guest/login", s.guestLogin)
			auth.GET("/verify", middleware.AuthRequired(s.config.JWT), s.verifyToken)
			auth.POST("/refresh", s.refreshToken)
			auth.POST("/logout", middleware.AuthRequired(s.config.JWT), s.logout)
		}

		// User routes
		user := api.Group("/user")
		user.Use(middleware.AuthRequired(s.config.JWT))
		{
			user.GET("/info", s.getUserInfo)
			user.PUT("/info", s.updateUserInfo)
			user.POST("/avatar", s.uploadAvatar)
			user.POST("/bind/phone", s.bindPhone)
			user.POST("/password/change", s.changePassword)
		}

		// Conversation routes
		conversations := api.Group("/conversations")
		conversations.Use(middleware.AuthRequired(s.config.JWT))
		{
			conversations.GET("", s.getConversations)
			conversations.POST("", s.createConversation)
			conversations.GET("/:id", s.getConversation)
			conversations.PUT("/:id", s.updateConversation)
			conversations.DELETE("/:id", s.deleteConversation)
			conversations.DELETE("/all", s.clearConversations)

			// Message routes
			conversations.GET("/:id/messages", s.getMessages)
		}

		// Message routes
		messages := api.Group("/messages")
		messages.Use(middleware.AuthRequired(s.config.JWT))
		{
			messages.POST("/text", s.sendTextMessage)
			messages.POST("/voice", s.sendVoiceMessage)
			messages.DELETE("/:id", s.deleteMessage)
		}

		// Voice processing routes
		voice := api.Group("/voice")
		voice.Use(middleware.AuthRequired(s.config.JWT))
		{
			voice.POST("/upload", s.uploadAudio)
			voice.POST("/asr", s.speechToText)
			voice.POST("/tts", s.textToSpeech)
			voice.GET("/voices", s.getVoiceList)
		}

		// Chat routes
		chat := api.Group("/chat")
		chat.Use(middleware.AuthRequired(s.config.JWT))
		{
			chat.POST("/send", s.sendChatMessage)
			chat.GET("/stream", s.streamChatMessage)
			chat.GET("/suggestions/:id", s.getChatSuggestions)
			chat.DELETE("/context/:id", s.clearChatContext)
		}

		// Settings routes
		settings := api.Group("/settings")
		settings.Use(middleware.AuthRequired(s.config.JWT))
		{
			settings.GET("", s.getSettings)
			settings.PUT("", s.updateSettings)
			settings.DELETE("", s.resetSettings)
		}

		// Analytics routes
		analytics := api.Group("/analytics")
		analytics.Use(middleware.AuthRequired(s.config.JWT))
		{
			analytics.GET("/stats", s.getStats)
			analytics.POST("/event", s.reportEvent)
		}

		// WebSocket for real-time features
		api.GET("/ws", s.handleWebSocket)
	}

	// Static file serving
	s.router.Static("/static", "./static")
	s.router.StaticFile("/favicon.ico", "./static/favicon.ico")

	// Catch-all for SPA
	s.router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "API endpoint not found",
			"path":    c.Request.URL.Path,
		})
	})
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Create HTTP server
	srv := &http.Server{
		Addr:           ":" + s.config.App.Port,
		Handler:        s.router,
		ReadTimeout:    time.Duration(s.config.App.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(s.config.App.WriteTimeout) * time.Second,
		IdleTimeout:    time.Duration(s.config.App.IdleTimeout) * time.Second,
		MaxHeaderBytes: s.config.App.MaxHeaderBytes,
	}

	// Channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		logger.Infof("Server starting on port %s", s.config.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-quit
	logger.Info("Server shutting down...")

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server gracefully
	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorf("Server forced to shutdown: %v", err)
		return err
	}

	// Close database connection
	s.db.Close()

	logger.Info("Server exited")
	return nil
}

// healthCheck handles health check requests
func (s *Server) healthCheck(c *gin.Context) {
	status := gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"version":   s.config.App.Version,
		"services": gin.H{
			"database": s.db.Health(),
			"redis":    true, // TODO: implement Redis health check
		},
	}

	c.JSON(http.StatusOK, status)
}

// Handler instances
var (
	authHandler  *handlers.AuthHandler
	voiceHandler *handlers.VoiceHandler
	chatHandler  *handlers.ChatHandler
)

// initHandlers initializes all handlers
func (s *Server) initHandlers() {
	authHandler = handlers.NewAuthHandler(s.db, s.config)
	voiceHandler = handlers.NewVoiceHandler(s.db, s.config)
	chatHandler = handlers.NewChatHandler(s.db, s.config)
}

// Authentication handlers
func (s *Server) phoneLogin(c *gin.Context)  { authHandler.PhoneLogin(c) }
func (s *Server) sendSmsCode(c *gin.Context) { authHandler.SendSMSCode(c) }
func (s *Server) wechatLogin(c *gin.Context) { authHandler.WechatLogin(c) }
func (s *Server) guestLogin(c *gin.Context)  { authHandler.GuestLogin(c) }
func (s *Server) verifyToken(c *gin.Context) { authHandler.VerifyToken(c) }
func (s *Server) refreshToken(c *gin.Context) { authHandler.RefreshToken(c) }
func (s *Server) logout(c *gin.Context)      { authHandler.Logout(c) }

// Voice processing handlers
func (s *Server) uploadAudio(c *gin.Context)   { voiceHandler.UploadAudio(c) }
func (s *Server) speechToText(c *gin.Context)  { voiceHandler.SpeechToText(c) }
func (s *Server) textToSpeech(c *gin.Context)  { voiceHandler.TextToSpeech(c) }
func (s *Server) getVoiceList(c *gin.Context)  { voiceHandler.GetVoiceList(c) }

// Chat handlers
func (s *Server) sendChatMessage(c *gin.Context)    { chatHandler.SendChatMessage(c) }
func (s *Server) streamChatMessage(c *gin.Context)  { chatHandler.StreamChatMessage(c) }
func (s *Server) getChatSuggestions(c *gin.Context) { chatHandler.GetChatSuggestions(c) }
func (s *Server) clearChatContext(c *gin.Context)   { chatHandler.ClearChatContext(c) }
func (s *Server) handleWebSocket(c *gin.Context)    { chatHandler.HandleWebSocket(c) }

// TODO: Implement other handler methods
func (s *Server) getSystemStatus(c *gin.Context)    { c.JSON(200, gin.H{"todo": "implement"}) }
func (s *Server) getSystemConfig(c *gin.Context)    { c.JSON(200, gin.H{"todo": "implement"}) }
func (s *Server) checkUpdate(c *gin.Context)        { c.JSON(200, gin.H{"todo": "implement"}) }
func (s *Server) reportError(c *gin.Context)        { c.JSON(200, gin.H{"todo": "implement"}) }
func (s *Server) getHelp(c *gin.Context)            { c.JSON(200, gin.H{"todo": "implement"}) }
func (s *Server) getUserInfo(c *gin.Context)        { c.JSON(200, gin.H{"todo": "implement"}) }
func (s *Server) updateUserInfo(c *gin.Context)     { c.JSON(200, gin.H{"todo": "implement"}) }
func (s *Server) uploadAvatar(c *gin.Context)       { c.JSON(200, gin.H{"todo": "implement"}) }
func (s *Server) bindPhone(c *gin.Context)          { c.JSON(200, gin.H{"todo": "implement"}) }
func (s *Server) changePassword(c *gin.Context)     { c.JSON(200, gin.H{"todo": "implement"}) }
func (s *Server) getConversations(c *gin.Context)   { c.JSON(200, gin.H{"todo": "implement"}) }
func (s *Server) createConversation(c *gin.Context) { c.JSON(200, gin.H{"todo": "implement"}) }
func (s *Server) getConversation(c *gin.Context)    { c.JSON(200, gin.H{"todo": "implement"}) }
func (s *Server) updateConversation(c *gin.Context) { c.JSON(200, gin.H{"todo": "implement"}) }
func (s *Server) deleteConversation(c *gin.Context) { c.JSON(200, gin.H{"todo": "implement"}) }
func (s *Server) clearConversations(c *gin.Context) { c.JSON(200, gin.H{"todo": "implement"}) }
func (s *Server) getMessages(c *gin.Context)        { c.JSON(200, gin.H{"todo": "implement"}) }
func (s *Server) sendTextMessage(c *gin.Context)    { c.JSON(200, gin.H{"todo": "implement"}) }
func (s *Server) sendVoiceMessage(c *gin.Context)   { c.JSON(200, gin.H{"todo": "implement"}) }
func (s *Server) deleteMessage(c *gin.Context)      { c.JSON(200, gin.H{"todo": "implement"}) }
func (s *Server) getSettings(c *gin.Context)        { c.JSON(200, gin.H{"todo": "implement"}) }
func (s *Server) updateSettings(c *gin.Context)     { c.JSON(200, gin.H{"todo": "implement"}) }
func (s *Server) resetSettings(c *gin.Context)      { c.JSON(200, gin.H{"todo": "implement"}) }
func (s *Server) getStats(c *gin.Context)           { c.JSON(200, gin.H{"todo": "implement"}) }
func (s *Server) reportEvent(c *gin.Context)        { c.JSON(200, gin.H{"todo": "implement"}) }