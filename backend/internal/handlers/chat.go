package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"voicegenie/internal/config"
	"voicegenie/pkg/database"
	"voicegenie/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// ChatHandler handles AI chat related requests
type ChatHandler struct {
	db       *database.DB
	config   *config.Config
	upgrader websocket.Upgrader
}

// NewChatHandler creates a new chat handler
func NewChatHandler(db *database.DB, cfg *config.Config) *ChatHandler {
	return &ChatHandler{
		db:     db,
		config: cfg,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// In production, implement proper origin checking
				return true
			},
		},
	}
}

// ChatRequest represents a chat message request
type ChatRequest struct {
	Message        string                 `json:"message" binding:"required"`
	ConversationID string                 `json:"conversation_id,omitempty"`
	Context        map[string]interface{} `json:"context,omitempty"`
	Model          string                 `json:"model,omitempty"`
	Temperature    float32                `json:"temperature,omitempty"`
	MaxTokens      int                    `json:"max_tokens,omitempty"`
}

// ChatResponse represents a chat response
type ChatResponse struct {
	Reply          string   `json:"reply"`
	ConversationID string   `json:"conversation_id"`
	AudioURL       string   `json:"audio_url,omitempty"`
	Suggestions    []string `json:"suggestions,omitempty"`
	TokensUsed     int      `json:"tokens_used,omitempty"`
	Model          string   `json:"model,omitempty"`
}

// OpenAIMessage represents a message in OpenAI format
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIRequest represents an OpenAI chat completion request
type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	Temperature float32         `json:"temperature,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

// OpenAIResponse represents an OpenAI chat completion response
type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// StreamResponse represents a streaming response chunk
type StreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

// SendChatMessage handles regular chat message sending
func (h *ChatHandler) SendChatMessage(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":      40100,
			"message":   "Authentication required",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      40000,
			"message":   "Invalid request parameters",
			"details":   err.Error(),
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Validate message length
	if len(req.Message) > h.config.AI.MaxMessageLength {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      40001,
			"message":   "Message too long",
			"max_length": h.config.AI.MaxMessageLength,
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Get or create conversation
	conversation, err := h.getOrCreateConversation(userID, req.ConversationID)
	if err != nil {
		logger.WithError(err).Error("Failed to get or create conversation")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      50000,
			"message":   "Failed to process conversation",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Save user message
	_, err = h.saveUserMessage(conversation.ID, userID, req.Message)
	if err != nil {
		logger.WithError(err).Error("Failed to save user message")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      50001,
			"message":   "Failed to save message",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Get conversation context
	messages, err := h.getConversationMessages(conversation.ID)
	if err != nil {
		logger.WithError(err).Error("Failed to get conversation messages")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      50002,
			"message":   "Failed to get conversation context",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Set default model and parameters
	if req.Model == "" {
		req.Model = "gpt-3.5-turbo"
	}
	if req.Temperature == 0 {
		req.Temperature = 0.7
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 1000
	}

	// Call OpenAI API
	aiResponse, err := h.callOpenAI(messages, req)
	if err != nil {
		logger.WithError(err).Error("OpenAI API call failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      50003,
			"message":   "AI service temporarily unavailable",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Save AI response
	aiMessage, err := h.saveAIMessage(conversation.ID, userID, aiResponse.Choices[0].Message.Content, req.Model, aiResponse.Usage.TotalTokens)
	if err != nil {
		logger.WithError(err).Error("Failed to save AI message")
	}

	// Update conversation
	h.updateConversation(conversation, aiResponse.Choices[0].Message.Content)

	// Generate TTS audio (optional)
	var audioURL string
	if h.config.AI.AutoTTS {
		ttsResult, err := h.performTTS(TTSRequest{
			Text:  aiResponse.Choices[0].Message.Content,
			Voice: "alloy",
			Speed: 1.0,
		})
		if err == nil {
			audioURL = ttsResult.AudioURL
			// Update AI message with audio URL
			h.db.Model(&aiMessage).Update("audio_url", audioURL)
		}
	}

	// Generate suggestions
	suggestions := h.generateSuggestions(aiResponse.Choices[0].Message.Content)

	// Record usage
	h.recordChatUsage(userID, req, aiResponse)

	response := ChatResponse{
		Reply:          aiResponse.Choices[0].Message.Content,
		ConversationID: strconv.Itoa(int(conversation.ID)),
		AudioURL:       audioURL,
		Suggestions:    suggestions,
		TokensUsed:     aiResponse.Usage.TotalTokens,
		Model:          req.Model,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"message":   "Chat message processed successfully",
		"data":      response,
		"timestamp": time.Now().Unix(),
	})

	logger.WithFields(map[string]interface{}{
		"user_id":         userID,
		"conversation_id": conversation.ID,
		"message_length":  len(req.Message),
		"response_length": len(aiResponse.Choices[0].Message.Content),
		"tokens_used":     aiResponse.Usage.TotalTokens,
	}).Info("Chat message processed successfully")
}

// StreamChatMessage handles streaming chat responses
func (h *ChatHandler) StreamChatMessage(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":      40100,
			"message":   "Authentication required",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      40000,
			"message":   "Invalid request parameters",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Get or create conversation
	conversation, err := h.getOrCreateConversation(userID, req.ConversationID)
	if err != nil {
		h.sendSSEError(c, "Failed to process conversation")
		return
	}

	// Save user message
	_, err = h.saveUserMessage(conversation.ID, userID, req.Message)
	if err != nil {
		h.sendSSEError(c, "Failed to save message")
		return
	}

	// Get conversation context
	messages, err := h.getConversationMessages(conversation.ID)
	if err != nil {
		h.sendSSEError(c, "Failed to get conversation context")
		return
	}

	// Set default parameters
	if req.Model == "" {
		req.Model = "gpt-3.5-turbo"
	}
	if req.Temperature == 0 {
		req.Temperature = 0.7
	}

	// Call OpenAI streaming API
	err = h.callOpenAIStream(c, messages, req, conversation, userID)
	if err != nil {
		h.sendSSEError(c, "AI service temporarily unavailable")
		return
	}
}

// HandleWebSocket handles WebSocket connections for real-time chat
func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
	// Upgrade HTTP connection to WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.WithError(err).Error("Failed to upgrade to WebSocket")
		return
	}
	defer conn.Close()

	// Get user ID from query parameter (in production, validate JWT)
	userID := c.Query("user_id")
	if userID == "" {
		conn.WriteJSON(gin.H{
			"type":    "error",
			"message": "Authentication required",
		})
		return
	}

	logger.WithField("user_id", userID).Info("WebSocket connection established")

	for {
		// Read message from client
		var req ChatRequest
		err := conn.ReadJSON(&req)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.WithError(err).Error("WebSocket error")
			}
			break
		}

		// Process chat message
		response, err := h.processChatMessage(userID, req)
		if err != nil {
			conn.WriteJSON(gin.H{
				"type":    "error",
				"message": "Failed to process message",
			})
			continue
		}

		// Send response
		conn.WriteJSON(gin.H{
			"type": "message",
			"data": response,
		})
	}

	logger.WithField("user_id", userID).Info("WebSocket connection closed")
}

// GetChatSuggestions returns conversation suggestions
func (h *ChatHandler) GetChatSuggestions(c *gin.Context) {
	conversationID := c.Param("id")
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      40000,
			"message":   "Conversation ID is required",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Get last few messages
	var messages []database.Message
	err := h.db.Where("conversation_id = ?", conversationID).
		Order("created_at DESC").
		Limit(5).
		Find(&messages).Error

	if err != nil {
		logger.WithError(err).Error("Failed to get conversation messages")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      50000,
			"message":   "Failed to get suggestions",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Generate suggestions based on conversation
	var suggestions []string
	if len(messages) > 0 {
		suggestions = h.generateSuggestions(messages[0].Content)
	} else {
		suggestions = []string{
			"你好，我想了解一下...",
			"请帮我解释一下...",
			"我有一个问题想咨询...",
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"message":   "Suggestions retrieved successfully",
		"data":      suggestions,
		"timestamp": time.Now().Unix(),
	})
}

// ClearChatContext clears conversation context
func (h *ChatHandler) ClearChatContext(c *gin.Context) {
	conversationID := c.Param("id")
	userID := c.GetString("user_id")

	// Verify conversation ownership
	var conversation database.Conversation
	err := h.db.Where("id = ? AND user_id = ?", conversationID, userID).First(&conversation).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":      40400,
			"message":   "Conversation not found",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Clear context (in this case, we'll just add a system message)
	uid, _ := strconv.ParseUint(userID, 10, 32)
	systemMessage := database.Message{
		UserID:         uint(uid),
		ConversationID: conversation.ID,
		Type:           "system",
		Content:        "对话上下文已清除",
		ContentType:    "text",
		Status:         "sent",
	}

	if err := h.db.Create(&systemMessage).Error; err != nil {
		logger.WithError(err).Error("Failed to create system message")
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"message":   "Chat context cleared successfully",
		"timestamp": time.Now().Unix(),
	})

	logger.WithFields(map[string]interface{}{
		"user_id":         userID,
		"conversation_id": conversationID,
	}).Info("Chat context cleared")
}

// Helper functions

func (h *ChatHandler) getOrCreateConversation(userID, conversationID string) (*database.Conversation, error) {
	uid, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		return nil, err
	}

	if conversationID != "" {
		// Get existing conversation
		var conversation database.Conversation
		err := h.db.Where("id = ? AND user_id = ?", conversationID, uid).First(&conversation).Error
		if err == nil {
			return &conversation, nil
		}
	}

	// Create new conversation
	conversation := database.Conversation{
		UserID:      uint(uid),
		Title:       "新对话",
		Status:      "active",
		Model:       "gpt-3.5-turbo",
		Temperature: 0.7,
	}

	if err := h.db.Create(&conversation).Error; err != nil {
		return nil, err
	}

	return &conversation, nil
}

func (h *ChatHandler) saveUserMessage(conversationID uint, userID, content string) (*database.Message, error) {
	uid, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		return nil, err
	}

	message := database.Message{
		UserID:         uint(uid),
		ConversationID: conversationID,
		Type:           "user",
		Content:        content,
		ContentType:    "text",
		Status:         "sent",
	}

	if err := h.db.Create(&message).Error; err != nil {
		return nil, err
	}

	return &message, nil
}

func (h *ChatHandler) saveAIMessage(conversationID uint, userID, content, model string, tokensUsed int) (*database.Message, error) {
	uid, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		return nil, err
	}

	message := database.Message{
		UserID:         uint(uid),
		ConversationID: conversationID,
		Type:           "ai",
		Content:        content,
		ContentType:    "text",
		Status:         "sent",
		Model:          model,
		TokensUsed:     tokensUsed,
		ProcessedAt:    &[]time.Time{time.Now()}[0],
	}

	if err := h.db.Create(&message).Error; err != nil {
		return nil, err
	}

	return &message, nil
}

func (h *ChatHandler) getConversationMessages(conversationID uint) ([]OpenAIMessage, error) {
	var messages []database.Message
	err := h.db.Where("conversation_id = ?", conversationID).
		Order("created_at ASC").
		Limit(20). // Limit context to last 20 messages
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	var openAIMessages []OpenAIMessage
	for _, msg := range messages {
		if msg.Type == "user" || msg.Type == "ai" {
			role := "user"
			if msg.Type == "ai" {
				role = "assistant"
			}
			openAIMessages = append(openAIMessages, OpenAIMessage{
				Role:    role,
				Content: msg.Content,
			})
		}
	}

	return openAIMessages, nil
}

func (h *ChatHandler) callOpenAI(messages []OpenAIMessage, req ChatRequest) (*OpenAIResponse, error) {
	// In production, implement actual OpenAI API call
	// For now, return mock response

	logger.Infof("Calling OpenAI API with %d messages", len(messages))

	// Simulate API call delay
	time.Sleep(2 * time.Second)

	return &OpenAIResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", uuid.New().String()[:8]),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []struct {
			Index   int `json:"index"`
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		}{
			{
				Index: 0,
				Message: struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}{
					Role:    "assistant",
					Content: "这是AI的回复。我理解了您的问题，让我为您详细解答...",
				},
				FinishReason: "stop",
			},
		},
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     len(req.Message) / 4, // Rough estimate
			CompletionTokens: 50,
			TotalTokens:      len(req.Message)/4 + 50,
		},
	}, nil
}

func (h *ChatHandler) callOpenAIStream(c *gin.Context, messages []OpenAIMessage, req ChatRequest, conversation *database.Conversation, userID string) error {
	// In production, implement actual OpenAI streaming API call
	// For now, simulate streaming response

	responseChunks := []string{
		"这是", "AI的", "流式", "回复。", "我理解了", "您的问题，", "让我为您", "详细解答...",
	}

	var fullResponse strings.Builder

	for i, chunk := range responseChunks {
		// Send chunk
		h.sendSSEMessage(c, "data", gin.H{
			"type":    "chunk",
			"content": chunk,
			"index":   i,
		})

		fullResponse.WriteString(chunk)

		// Simulate streaming delay
		time.Sleep(200 * time.Millisecond)

		// Flush the response
		c.Writer.Flush()
	}

	// Send completion message
	h.sendSSEMessage(c, "data", gin.H{
		"type": "done",
		"conversation_id": strconv.Itoa(int(conversation.ID)),
	})

	// Save AI message
	h.saveAIMessage(conversation.ID, userID, fullResponse.String(), req.Model, 50)

	// Update conversation
	h.updateConversation(conversation, fullResponse.String())

	return nil
}

func (h *ChatHandler) processChatMessage(userID string, req ChatRequest) (*ChatResponse, error) {
	// Get or create conversation
	conversation, err := h.getOrCreateConversation(userID, req.ConversationID)
	if err != nil {
		return nil, err
	}

	// Save user message
	_, err = h.saveUserMessage(conversation.ID, userID, req.Message)
	if err != nil {
		return nil, err
	}

	// Get conversation context
	messages, err := h.getConversationMessages(conversation.ID)
	if err != nil {
		return nil, err
	}

	// Call OpenAI API
	if req.Model == "" {
		req.Model = "gpt-3.5-turbo"
	}
	if req.Temperature == 0 {
		req.Temperature = 0.7
	}

	aiResponse, err := h.callOpenAI(messages, req)
	if err != nil {
		return nil, err
	}

	// Save AI response
	_, err = h.saveAIMessage(conversation.ID, userID, aiResponse.Choices[0].Message.Content, req.Model, aiResponse.Usage.TotalTokens)
	if err != nil {
		return nil, err
	}

	// Update conversation
	h.updateConversation(conversation, aiResponse.Choices[0].Message.Content)

	return &ChatResponse{
		Reply:          aiResponse.Choices[0].Message.Content,
		ConversationID: strconv.Itoa(int(conversation.ID)),
		TokensUsed:     aiResponse.Usage.TotalTokens,
		Model:          req.Model,
	}, nil
}

func (h *ChatHandler) updateConversation(conversation *database.Conversation, lastMessage string) {
	// Update conversation with last message
	updates := map[string]interface{}{
		"last_message":    lastMessage,
		"last_message_at": time.Now(),
		"message_count":   conversation.MessageCount + 2, // User + AI message
	}

	h.db.Model(conversation).Updates(updates)
}

func (h *ChatHandler) generateSuggestions(lastMessage string) []string {
	// Simple suggestion generation based on keywords
	// In production, use AI to generate contextual suggestions

	defaultSuggestions := []string{
		"请继续解释",
		"能举个例子吗？",
		"还有其他观点吗？",
	}

	return defaultSuggestions
}

func (h *ChatHandler) recordChatUsage(userID string, req ChatRequest, response *OpenAIResponse) {
	uid, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		return
	}

	usage := database.Usage{
		UserID:    uint(uid),
		Service:   "openai",
		Operation: "chat",
		Model:     req.Model,
		TokensUsed: response.Usage.TotalTokens,
		Requests:  1,
		Date:      time.Now().Truncate(24 * time.Hour),
	}

	h.db.Create(&usage)
}

func (h *ChatHandler) sendSSEMessage(c *gin.Context, event string, data interface{}) {
	dataBytes, _ := json.Marshal(data)
	fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, string(dataBytes))
}

func (h *ChatHandler) sendSSEError(c *gin.Context, message string) {
	h.sendSSEMessage(c, "error", gin.H{"message": message})
}

// performTTS is a wrapper for TTS functionality
func (h *ChatHandler) performTTS(req TTSRequest) (*TTSResponse, error) {
	// Mock TTS response for now
	// In production, this should call the actual TTS service
	audioURL := fmt.Sprintf("%s/static/tts/%s.mp3", h.config.App.BaseURL, uuid.New().String())

	return &TTSResponse{
		AudioURL: audioURL,
		Duration: len(req.Text) / 10, // Rough estimate: 10 chars per second
		Text:     req.Text,
	}, nil
}