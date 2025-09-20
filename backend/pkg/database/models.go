package database

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel contains common fields for all models
type BaseModel struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// User represents a user account
type User struct {
	BaseModel

	// Basic info
	Username    string    `json:"username" gorm:"uniqueIndex;size:50"`
	Nickname    string    `json:"nickname" gorm:"size:100"`
	Avatar      string    `json:"avatar" gorm:"size:500"`
	Email       string    `json:"email" gorm:"uniqueIndex;size:100"`
	Phone       string    `json:"phone" gorm:"uniqueIndex;size:20"`

	// Authentication
	Password    string    `json:"-" gorm:"size:255"`
	LoginType   string    `json:"login_type" gorm:"size:20;default:'phone'"` // phone, wechat, guest
	WechatID    string    `json:"wechat_id" gorm:"size:100"`

	// Status
	Status      string    `json:"status" gorm:"size:20;default:'active'"` // active, inactive, banned
	LastLoginAt *time.Time `json:"last_login_at"`
	LastLoginIP string    `json:"last_login_ip" gorm:"size:45"`

	// Preferences
	Language    string    `json:"language" gorm:"size:10;default:'zh-CN'"`
	Timezone    string    `json:"timezone" gorm:"size:50;default:'Asia/Shanghai'"`
	Theme       string    `json:"theme" gorm:"size:20;default:'auto'"`

	// Relationships
	Conversations []Conversation `json:"conversations,omitempty" gorm:"foreignKey:UserID"`
	Messages      []Message      `json:"messages,omitempty" gorm:"foreignKey:UserID"`
	Settings      []Setting      `json:"settings,omitempty" gorm:"foreignKey:UserID"`
}

// Conversation represents a conversation session
type Conversation struct {
	BaseModel

	UserID      uint      `json:"user_id" gorm:"not null;index"`
	Title       string    `json:"title" gorm:"size:200"`
	Description string    `json:"description" gorm:"size:500"`

	// Status
	Status      string    `json:"status" gorm:"size:20;default:'active'"` // active, archived, deleted

	// Statistics
	MessageCount int       `json:"message_count" gorm:"default:0"`
	Duration     int       `json:"duration" gorm:"default:0"` // Total duration in seconds
	LastMessage  string    `json:"last_message" gorm:"size:500"`
	LastMessageAt *time.Time `json:"last_message_at"`

	// AI Context
	Context     string    `json:"context" gorm:"type:text"` // JSON string for AI context
	Model       string    `json:"model" gorm:"size:50;default:'gpt-3.5-turbo'"`
	Temperature float32   `json:"temperature" gorm:"default:0.7"`

	// Relationships
	User        User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Messages    []Message `json:"messages,omitempty" gorm:"foreignKey:ConversationID"`
}

// Message represents a message in a conversation
type Message struct {
	BaseModel

	UserID         uint      `json:"user_id" gorm:"not null;index"`
	ConversationID uint      `json:"conversation_id" gorm:"not null;index"`

	// Content
	Type        string    `json:"type" gorm:"size:20;not null"` // user, ai, system
	Content     string    `json:"content" gorm:"type:text"`
	ContentType string    `json:"content_type" gorm:"size:20;default:'text'"` // text, audio, image

	// Audio related
	AudioURL     string    `json:"audio_url,omitempty" gorm:"size:500"`
	AudioDuration int      `json:"audio_duration,omitempty"` // Duration in seconds
	AudioSize    int64     `json:"audio_size,omitempty"` // File size in bytes

	// Processing status
	Status      string    `json:"status" gorm:"size:20;default:'sent'"` // sending, sent, failed, processed
	ProcessedAt *time.Time `json:"processed_at,omitempty"`

	// AI related (for AI messages)
	Model       string    `json:"model,omitempty" gorm:"size:50"`
	TokensUsed  int       `json:"tokens_used,omitempty"`
	Cost        float64   `json:"cost,omitempty"`

	// Metadata
	Metadata    string    `json:"metadata,omitempty" gorm:"type:text"` // JSON string for additional data

	// Relationships
	User         User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Conversation Conversation `json:"conversation,omitempty" gorm:"foreignKey:ConversationID"`
}

// Setting represents user settings
type Setting struct {
	BaseModel

	UserID uint   `json:"user_id" gorm:"not null;index"`
	Key    string `json:"key" gorm:"size:100;not null"`
	Value  string `json:"value" gorm:"type:text"`
	Type   string `json:"type" gorm:"size:20;default:'string'"` // string, number, boolean, json

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// AudioFile represents uploaded audio files
type AudioFile struct {
	BaseModel

	UserID uint   `json:"user_id" gorm:"not null;index"`

	// File info
	Filename     string `json:"filename" gorm:"size:255;not null"`
	OriginalName string `json:"original_name" gorm:"size:255"`
	Path         string `json:"path" gorm:"size:500;not null"`
	URL          string `json:"url" gorm:"size:500"`

	// File properties
	Size         int64  `json:"size"`
	MimeType     string `json:"mime_type" gorm:"size:100"`
	Duration     int    `json:"duration"` // Duration in seconds
	SampleRate   int    `json:"sample_rate"`
	Channels     int    `json:"channels"`
	Bitrate      int    `json:"bitrate"`

	// Processing status
	Status       string `json:"status" gorm:"size:20;default:'uploaded'"` // uploaded, processing, processed, failed
	ProcessedAt  *time.Time `json:"processed_at,omitempty"`

	// ASR results
	Transcript   string `json:"transcript,omitempty" gorm:"type:text"`
	Confidence   float32 `json:"confidence,omitempty"`
	Language     string `json:"language,omitempty" gorm:"size:10"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// APIKey represents API keys for external services
type APIKey struct {
	BaseModel

	Name        string    `json:"name" gorm:"size:100;not null"`
	Service     string    `json:"service" gorm:"size:50;not null"` // openai, deepgram, elevenlabs, etc.
	Key         string    `json:"key" gorm:"size:500;not null"`
	Encrypted   bool      `json:"encrypted" gorm:"default:true"`

	// Usage limits
	DailyLimit  int       `json:"daily_limit" gorm:"default:0"` // 0 means no limit
	MonthlyLimit int      `json:"monthly_limit" gorm:"default:0"`

	// Status
	Status      string    `json:"status" gorm:"size:20;default:'active'"` // active, inactive, expired
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
}

// Usage represents API usage tracking
type Usage struct {
	BaseModel

	UserID uint   `json:"user_id" gorm:"not null;index"`

	// Service info
	Service     string `json:"service" gorm:"size:50;not null"` // openai, deepgram, elevenlabs
	Operation   string `json:"operation" gorm:"size:50;not null"` // chat, asr, tts
	Model       string `json:"model,omitempty" gorm:"size:50"`

	// Usage metrics
	TokensUsed  int     `json:"tokens_used,omitempty"`
	Characters  int     `json:"characters,omitempty"`
	Seconds     int     `json:"seconds,omitempty"`
	Requests    int     `json:"requests" gorm:"default:1"`
	Cost        float64 `json:"cost,omitempty"`

	// Metadata
	Date        time.Time `json:"date" gorm:"index"`
	Metadata    string    `json:"metadata,omitempty" gorm:"type:text"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// ErrorLog represents error logs
type ErrorLog struct {
	BaseModel

	UserID      *uint     `json:"user_id,omitempty" gorm:"index"`

	// Error info
	Level       string    `json:"level" gorm:"size:20;not null"` // error, warning, info
	Message     string    `json:"message" gorm:"type:text;not null"`
	Stack       string    `json:"stack,omitempty" gorm:"type:text"`

	// Request info
	Method      string    `json:"method,omitempty" gorm:"size:10"`
	URL         string    `json:"url,omitempty" gorm:"size:500"`
	UserAgent   string    `json:"user_agent,omitempty" gorm:"size:500"`
	IP          string    `json:"ip,omitempty" gorm:"size:45"`

	// Additional context
	Context     string    `json:"context,omitempty" gorm:"type:text"`

	// Status
	Resolved    bool      `json:"resolved" gorm:"default:false"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// Table names
func (User) TableName() string         { return "users" }
func (Conversation) TableName() string { return "conversations" }
func (Message) TableName() string      { return "messages" }
func (Setting) TableName() string      { return "settings" }
func (AudioFile) TableName() string    { return "audio_files" }
func (APIKey) TableName() string       { return "api_keys" }
func (Usage) TableName() string        { return "usage" }
func (ErrorLog) TableName() string     { return "error_logs" }

// Indexes for better performance
func (User) Indexes() []string {
	return []string{
		"idx_users_username",
		"idx_users_email",
		"idx_users_phone",
		"idx_users_status",
		"idx_users_login_type",
	}
}

func (Conversation) Indexes() []string {
	return []string{
		"idx_conversations_user_id",
		"idx_conversations_status",
		"idx_conversations_updated_at",
	}
}

func (Message) Indexes() []string {
	return []string{
		"idx_messages_user_id",
		"idx_messages_conversation_id",
		"idx_messages_type",
		"idx_messages_created_at",
		"idx_messages_user_conversation", // composite index
	}
}

func (AudioFile) Indexes() []string {
	return []string{
		"idx_audio_files_user_id",
		"idx_audio_files_status",
		"idx_audio_files_created_at",
	}
}

func (Usage) Indexes() []string {
	return []string{
		"idx_usage_user_id",
		"idx_usage_service",
		"idx_usage_date",
		"idx_usage_user_date", // composite index
	}
}