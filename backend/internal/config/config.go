package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for our application
type Config struct {
	App       AppConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	Log       LogConfig
	AI        AIConfig
	Upload    UploadConfig
	RateLimit RateLimitConfig
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Name           string
	Version        string
	Environment    string
	Port           string
	BaseURL        string
	ReadTimeout    int
	WriteTimeout   int
	IdleTimeout    int
	MaxHeaderBytes int
	MaxRequestSize int64
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type            string
	Host            string
	Port            int
	Name            string
	User            string
	Password        string
	SSLMode         string
	Timezone        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
	AutoMigrate     bool
	LogLevel        string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host            string
	Port            string
	Password        string
	PoolSize        int
	MinIdleConns    int
	IdleTimeout     time.Duration
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret                string
	Issuer                string
	ExpirationHours       int
	RefreshExpirationDays int
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string
	Format string
	Output string
}

// AIConfig holds AI service configuration
type AIConfig struct {
	OpenAI         OpenAIConfig
	Deepgram       DeepgramConfig
	ElevenLabs     ElevenLabsConfig
	MaxTextLength    int
	MaxMessageLength int
	AutoTTS          bool
}

// OpenAIConfig holds OpenAI configuration
type OpenAIConfig struct {
	APIKey    string
	APIBase   string
	Model     string
	MaxTokens int
}

// DeepgramConfig holds Deepgram configuration
type DeepgramConfig struct {
	APIKey string
	APIURL string
}

// ElevenLabsConfig holds ElevenLabs configuration
type ElevenLabsConfig struct {
	APIKey  string
	APIURL  string
	VoiceID string
}

// UploadConfig holds file upload configuration
type UploadConfig struct {
	AudioPath         string
	MaxFileSize       int64
	AllowedAudioTypes []string
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	MaxRequests    int
	WindowDuration time.Duration
}

// New creates a new configuration instance
func New() *Config {
	return &Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "VoiceGenie"),
			Version:     getEnv("APP_VERSION", "0.1.0"),
			Environment: getEnv("APP_ENV", "development"),
			Port:        getEnv("APP_PORT", "8080"),
			BaseURL:     getEnv("APP_BASE_URL", "http://localhost:8080"),
			ReadTimeout:    getEnvAsInt("APP_READ_TIMEOUT", 30),
			WriteTimeout:   getEnvAsInt("APP_WRITE_TIMEOUT", 30),
			IdleTimeout:    getEnvAsInt("APP_IDLE_TIMEOUT", 60),
			MaxHeaderBytes: getEnvAsInt("APP_MAX_HEADER_BYTES", 1048576),
			MaxRequestSize: getEnvAsInt64("APP_MAX_REQUEST_SIZE", 10485760),
		},
		Database: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnvAsInt("DB_PORT", 5432),
			Name:         getEnv("DB_NAME", "voicegenie"),
			User:         getEnv("DB_USER", "postgres"),
			Password:     getEnv("DB_PASSWORD", "postgres123"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnvAsInt("DB_CONN_MAX_LIFETIME", 300),
			AutoMigrate:     getEnvAsBool("DB_AUTO_MIGRATE", true),
			LogLevel:        getEnv("DB_LOG_LEVEL", "warn"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			Timezone:        getEnv("DB_TIMEZONE", "UTC"),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnv("REDIS_PORT", "6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			PoolSize:     getEnvAsInt("REDIS_POOL_SIZE", 10),
			MinIdleConns: getEnvAsInt("REDIS_MIN_IDLE_CONNS", 5),
			IdleTimeout:  getEnvAsDuration("REDIS_IDLE_TIMEOUT", 5*time.Minute),
		},
		JWT: JWTConfig{
			Secret:                getEnv("JWT_SECRET", "your-super-secret-jwt-key"),
			Issuer:                getEnv("JWT_ISSUER", "voicegenie"),
			ExpirationHours:       getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
			RefreshExpirationDays: getEnvAsInt("JWT_REFRESH_EXPIRATION_DAYS", 7),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "debug"),
			Format: getEnv("LOG_FORMAT", "json"),
			Output: getEnv("LOG_OUTPUT", "stdout"),
		},
		AI: AIConfig{
			OpenAI: OpenAIConfig{
				APIKey:    getEnv("OPENAI_API_KEY", ""),
				APIBase:   getEnv("OPENAI_API_BASE", "https://api.openai.com/v1"),
				Model:     getEnv("OPENAI_MODEL", "gpt-3.5-turbo"),
				MaxTokens: getEnvAsInt("OPENAI_MAX_TOKENS", 1000),
			},
			Deepgram: DeepgramConfig{
				APIKey: getEnv("DEEPGRAM_API_KEY", ""),
				APIURL: getEnv("DEEPGRAM_API_URL", "https://api.deepgram.com/v1"),
			},
			ElevenLabs: ElevenLabsConfig{
				APIKey:  getEnv("ELEVENLABS_API_KEY", ""),
				APIURL:  getEnv("ELEVENLABS_API_URL", "https://api.elevenlabs.io/v1"),
				VoiceID: getEnv("ELEVENLABS_VOICE_ID", "21m00Tcm4TlvDq8ikWAM"),
			},
			MaxTextLength:    getEnvAsInt("AI_MAX_TEXT_LENGTH", 2000),
			MaxMessageLength: getEnvAsInt("AI_MAX_MESSAGE_LENGTH", 1000),
			AutoTTS:          getEnvAsBool("AI_AUTO_TTS", false),
		},
		Upload: UploadConfig{
			AudioPath:         getEnv("UPLOAD_AUDIO_PATH", "./uploads/audio"),
			MaxFileSize:       getEnvAsInt64("UPLOAD_MAX_FILE_SIZE", 10485760),
			AllowedAudioTypes: getEnvAsSlice("ALLOWED_AUDIO_TYPES", []string{"mp3", "wav", "m4a", "aac"}),
		},
		RateLimit: RateLimitConfig{
			MaxRequests:    getEnvAsInt("RATE_LIMIT_MAX_REQUESTS", 100),
			WindowDuration: getEnvAsDuration("RATE_LIMIT_WINDOW", 1*time.Minute),
		},
	}
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Split by comma
		return splitString(value, ",")
	}
	return defaultValue
}

func splitString(s, sep string) []string {
	var result []string
	for _, item := range splitByString(s, sep) {
		if trimmed := trimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func splitByString(s, sep string) []string {
	// Simple split implementation
	if sep == "" {
		return []string{s}
	}

	var result []string
	start := 0
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func trimSpace(s string) string {
	start := 0
	end := len(s)

	// Trim leading spaces
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	// Trim trailing spaces
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}