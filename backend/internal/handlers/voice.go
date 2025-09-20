package handlers

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"voicegenie/internal/config"
	"voicegenie/pkg/database"
	"voicegenie/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// VoiceHandler handles voice processing related requests
type VoiceHandler struct {
	db     *database.DB
	config *config.Config
}

// NewVoiceHandler creates a new voice handler
func NewVoiceHandler(db *database.DB, cfg *config.Config) *VoiceHandler {
	return &VoiceHandler{
		db:     db,
		config: cfg,
	}
}

// ASRRequest represents speech-to-text request
type ASRRequest struct {
	AudioURL              string `json:"audio_url" binding:"required"`
	Language              string `json:"language,omitempty"`
	EnablePunctuation     bool   `json:"enable_punctuation,omitempty"`
	EnableWordTimeStamp   bool   `json:"enable_word_time_stamp,omitempty"`
}

// ASRResponse represents speech-to-text response
type ASRResponse struct {
	Text       string  `json:"text"`
	Confidence float32 `json:"confidence"`
	Language   string  `json:"language"`
	Duration   int     `json:"duration"`
}

// TTSRequest represents text-to-speech request
type TTSRequest struct {
	Text   string  `json:"text" binding:"required"`
	Voice  string  `json:"voice,omitempty"`
	Speed  float32 `json:"speed,omitempty"`
	Pitch  float32 `json:"pitch,omitempty"`
	Volume float32 `json:"volume,omitempty"`
}

// TTSResponse represents text-to-speech response
type TTSResponse struct {
	AudioURL string `json:"audio_url"`
	Duration int    `json:"duration"`
	Text     string `json:"text"`
}

// UploadAudio handles audio file upload
func (h *VoiceHandler) UploadAudio(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":      40100,
			"message":   "Authentication required",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Parse multipart form
	err := c.Request.ParseMultipartForm(32 << 20) // 32MB max
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      40000,
			"message":   "Invalid form data",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Get file from form
	file, header, err := c.Request.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      40001,
			"message":   "Audio file is required",
			"timestamp": time.Now().Unix(),
		})
		return
	}
	defer file.Close()

	// Validate file type
	if !isValidAudioFile(header.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      40002,
			"message":   "Invalid audio file format. Supported formats: mp3, wav, m4a, ogg",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Validate file size
	if header.Size > h.config.Upload.MaxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      40003,
			"message":   "File size exceeds limit",
			"max_size":  h.config.Upload.MaxFileSize,
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Generate unique filename
	fileExt := filepath.Ext(header.Filename)
	fileName := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), fileExt)
	filePath := filepath.Join(h.config.Upload.AudioPath, fileName)

	// Save file to storage
	savedPath, err := h.saveAudioFile(file, filePath)
	if err != nil {
		logger.WithError(err).Error("Failed to save audio file")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      50000,
			"message":   "Failed to save audio file",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Get audio metadata
	metadata, err := h.getAudioMetadata(savedPath)
	if err != nil {
		logger.WithError(err).Warn("Failed to get audio metadata")
		metadata = &AudioMetadata{
			Duration: 0,
			SampleRate: 0,
			Channels: 1,
			Bitrate: 0,
		}
	}

	// Convert userID to uint
	uid, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      40004,
			"message":   "Invalid user ID",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Create audio file record
	audioFile := database.AudioFile{
		UserID:       uint(uid),
		Filename:     fileName,
		OriginalName: header.Filename,
		Path:         savedPath,
		URL:          h.generateFileURL(fileName),
		Size:         header.Size,
		MimeType:     getMimeType(fileExt),
		Duration:     metadata.Duration,
		SampleRate:   metadata.SampleRate,
		Channels:     metadata.Channels,
		Bitrate:      metadata.Bitrate,
		Status:       "uploaded",
	}

	if err := h.db.Create(&audioFile).Error; err != nil {
		logger.WithError(err).Error("Failed to create audio file record")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      50001,
			"message":   "Failed to save file information",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"message":   "Audio file uploaded successfully",
		"data": gin.H{
			"url":       audioFile.URL,
			"filename":  audioFile.Filename,
			"size":      audioFile.Size,
			"duration":  audioFile.Duration,
			"mime_type": audioFile.MimeType,
		},
		"timestamp": time.Now().Unix(),
	})

	logger.WithFields(map[string]interface{}{
		"user_id":  userID,
		"filename": audioFile.Filename,
		"size":     audioFile.Size,
		"duration": audioFile.Duration,
	}).Info("Audio file uploaded successfully")
}

// SpeechToText handles speech-to-text conversion
func (h *VoiceHandler) SpeechToText(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":      40100,
			"message":   "Authentication required",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	var req ASRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      40000,
			"message":   "Invalid request parameters",
			"details":   err.Error(),
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Validate audio URL
	if !h.isValidAudioURL(req.AudioURL) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      40001,
			"message":   "Invalid audio URL",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Call ASR service
	result, err := h.performASR(req)
	if err != nil {
		logger.WithError(err).Error("ASR processing failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      50002,
			"message":   "Speech recognition failed",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Update audio file record with ASR results
	h.updateAudioFileASR(req.AudioURL, result)

	// Record usage
	h.recordASRUsage(userID, result)

	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"message":   "Speech recognition completed",
		"data":      result,
		"timestamp": time.Now().Unix(),
	})

	logger.WithFields(map[string]interface{}{
		"user_id":    userID,
		"audio_url":  req.AudioURL,
		"text_length": len(result.Text),
		"confidence": result.Confidence,
	}).Info("Speech-to-text conversion completed")
}

// TextToSpeech handles text-to-speech conversion
func (h *VoiceHandler) TextToSpeech(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":      40100,
			"message":   "Authentication required",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	var req TTSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      40000,
			"message":   "Invalid request parameters",
			"details":   err.Error(),
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Validate text length
	if len(req.Text) > h.config.AI.MaxTextLength {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      40001,
			"message":   "Text too long",
			"max_length": h.config.AI.MaxTextLength,
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Set default values
	if req.Voice == "" {
		req.Voice = "alloy" // Default OpenAI voice
	}
	if req.Speed == 0 {
		req.Speed = 1.0
	}
	if req.Volume == 0 {
		req.Volume = 1.0
	}

	// Call TTS service
	result, err := h.performTTS(req)
	if err != nil {
		logger.WithError(err).Error("TTS processing failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      50003,
			"message":   "Text-to-speech conversion failed",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Record usage
	h.recordTTSUsage(userID, req, result)

	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"message":   "Text-to-speech conversion completed",
		"data":      result,
		"timestamp": time.Now().Unix(),
	})

	logger.WithFields(map[string]interface{}{
		"user_id":     userID,
		"text_length": len(req.Text),
		"voice":       req.Voice,
		"duration":    result.Duration,
	}).Info("Text-to-speech conversion completed")
}

// GetVoiceList returns available TTS voices
func (h *VoiceHandler) GetVoiceList(c *gin.Context) {
	voices := []gin.H{
		{
			"id":          "alloy",
			"name":        "Alloy",
			"language":    "en-US",
			"gender":      "neutral",
			"description": "Natural and balanced voice",
		},
		{
			"id":          "echo",
			"name":        "Echo",
			"language":    "en-US",
			"gender":      "male",
			"description": "Clear and articulate male voice",
		},
		{
			"id":          "fable",
			"name":        "Fable",
			"language":    "en-US",
			"gender":      "male",
			"description": "Warm and storytelling voice",
		},
		{
			"id":          "onyx",
			"name":        "Onyx",
			"language":    "en-US",
			"gender":      "male",
			"description": "Deep and authoritative voice",
		},
		{
			"id":          "nova",
			"name":        "Nova",
			"language":    "en-US",
			"gender":      "female",
			"description": "Bright and energetic female voice",
		},
		{
			"id":          "shimmer",
			"name":        "Shimmer",
			"language":    "en-US",
			"gender":      "female",
			"description": "Gentle and soothing female voice",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"message":   "Voice list retrieved successfully",
		"data":      voices,
		"timestamp": time.Now().Unix(),
	})
}

// Helper functions

// AudioMetadata represents audio file metadata
type AudioMetadata struct {
	Duration   int `json:"duration"`
	SampleRate int `json:"sample_rate"`
	Channels   int `json:"channels"`
	Bitrate    int `json:"bitrate"`
}

func isValidAudioFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExts := []string{".mp3", ".wav", ".m4a", ".ogg", ".flac", ".aac"}

	for _, validExt := range validExts {
		if ext == validExt {
			return true
		}
	}
	return false
}

func getMimeType(ext string) string {
	mimeTypes := map[string]string{
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".m4a":  "audio/mp4",
		".ogg":  "audio/ogg",
		".flac": "audio/flac",
		".aac":  "audio/aac",
	}

	if mimeType, exists := mimeTypes[strings.ToLower(ext)]; exists {
		return mimeType
	}
	return "audio/mpeg"
}

func (h *VoiceHandler) saveAudioFile(file multipart.File, filePath string) (string, error) {
	// In production, this would save to cloud storage (AWS S3, etc.)
	// For now, we'll simulate saving and return the path
	logger.Infof("Saving audio file to: %s", filePath)

	// Create the file content (in real implementation, save to disk/cloud)
	_, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func (h *VoiceHandler) generateFileURL(filename string) string {
	return fmt.Sprintf("%s/static/audio/%s", h.config.App.BaseURL, filename)
}

func (h *VoiceHandler) getAudioMetadata(filePath string) (*AudioMetadata, error) {
	// In production, use ffprobe or similar tool to get actual metadata
	// For now, return mock data
	return &AudioMetadata{
		Duration:   30, // 30 seconds
		SampleRate: 44100,
		Channels:   2,
		Bitrate:    128000,
	}, nil
}

func (h *VoiceHandler) isValidAudioURL(url string) bool {
	// Basic URL validation
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

func (h *VoiceHandler) performASR(req ASRRequest) (*ASRResponse, error) {
	// In production, integrate with ASR service (OpenAI Whisper, Deepgram, etc.)
	// For now, return mock response

	logger.Infof("Performing ASR on audio URL: %s", req.AudioURL)

	// Simulate ASR processing time
	time.Sleep(2 * time.Second)

	return &ASRResponse{
		Text:       "这是一段语音转文字的测试结果。",
		Confidence: 0.95,
		Language:   "zh-CN",
		Duration:   30,
	}, nil
}

func (h *VoiceHandler) performTTS(req TTSRequest) (*TTSResponse, error) {
	// In production, integrate with TTS service (OpenAI TTS, ElevenLabs, etc.)
	// For now, return mock response

	logger.Infof("Performing TTS for text: %s", req.Text[:min(50, len(req.Text))])

	// Simulate TTS processing time
	time.Sleep(1 * time.Second)

	// Generate mock audio URL
	audioURL := fmt.Sprintf("%s/static/tts/%s.mp3", h.config.App.BaseURL, uuid.New().String())

	return &TTSResponse{
		AudioURL: audioURL,
		Duration: len(req.Text) / 10, // Rough estimate: 10 chars per second
		Text:     req.Text,
	}, nil
}

func (h *VoiceHandler) updateAudioFileASR(audioURL string, result *ASRResponse) {
	// Update audio file record with ASR results
	h.db.Model(&database.AudioFile{}).
		Where("url = ?", audioURL).
		Updates(map[string]interface{}{
			"transcript":   result.Text,
			"confidence":   result.Confidence,
			"language":     result.Language,
			"status":       "processed",
			"processed_at": time.Now(),
		})
}

func (h *VoiceHandler) recordASRUsage(userID string, result *ASRResponse) {
	uid, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		return
	}

	usage := database.Usage{
		UserID:     uint(uid),
		Service:    "deepgram",
		Operation:  "asr",
		Seconds:    result.Duration,
		Characters: len(result.Text),
		Requests:   1,
		Date:       time.Now().Truncate(24 * time.Hour),
	}

	h.db.Create(&usage)
}

func (h *VoiceHandler) recordTTSUsage(userID string, req TTSRequest, result *TTSResponse) {
	uid, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		return
	}

	usage := database.Usage{
		UserID:     uint(uid),
		Service:    "openai",
		Operation:  "tts",
		Model:      req.Voice,
		Characters: len(req.Text),
		Seconds:    result.Duration,
		Requests:   1,
		Date:       time.Now().Truncate(24 * time.Hour),
	}

	h.db.Create(&usage)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}