package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for our application
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Log      LogConfig
	AI       AIConfig
	Upload   UploadConfig
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Name    string
	Version string
	Env     string
	Port    string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host         string
	Port         string
	Name         string
	User         string
	Password     string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
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
	Secret           string
	ExpiresIn        time.Duration
	RefreshExpiresIn time.Duration
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string
	Format string
	Output string
}

// AIConfig holds AI service configuration
type AIConfig struct {
	OpenAI    OpenAIConfig
	Deepgram  DeepgramConfig
	ElevenLabs ElevenLabsConfig
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
	Path              string
	MaxSize           int64
	AllowedAudioTypes []string
}

// New creates a new configuration instance
func New() *Config {
	return &Config{
		App: AppConfig{
			Name:    getEnv("APP_NAME", "VoiceGenie"),
			Version: getEnv("APP_VERSION", "0.1.0"),
			Env:     getEnv("APP_ENV", "development"),
			Port:    getEnv("APP_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnv("DB_PORT", "5432"),
			Name:         getEnv("DB_NAME", "voicegenie"),
			User:         getEnv("DB_USER", "postgres"),
			Password:     getEnv("DB_PASSWORD", "postgres123"),
			MaxOpenConns: getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
			MaxLifetime:  getEnvAsDuration("DB_MAX_LIFETIME", 5*time.Minute),
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
			Secret:           getEnv("JWT_SECRET", "your-super-secret-jwt-key"),
			ExpiresIn:        getEnvAsDuration("JWT_EXPIRES_IN", 24*time.Hour),
			RefreshExpiresIn: getEnvAsDuration("JWT_REFRESH_EXPIRES_IN", 168*time.Hour),
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
		},
		Upload: UploadConfig{
			Path:              getEnv("UPLOAD_PATH", "./uploads"),
			MaxSize:           getEnvAsInt64("UPLOAD_MAX_SIZE", 10485760), // 10MB
			AllowedAudioTypes: getEnvAsSlice("ALLOWED_AUDIO_TYPES", []string{"mp3", "wav", "m4a", "aac"}),
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