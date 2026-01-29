package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/daniel0321forever/terriyaki-go/internal/config"
	"github.com/daniel0321forever/terriyaki-go/internal/services"
	"github.com/daniel0321forever/terriyaki-go/internal/utils"
	"github.com/gin-gonic/gin"
)

// ConvertVoiceAPI converts audio to a different voice using ElevenLabs Speech-to-Speech.
// This is a standalone voice changer endpoint (not tied to interview sessions).
//
// Request (multipart/form-data):
// - audio: file (required)
// - voice_id: string (required)
// - model_id: string (optional, default: eleven_multilingual_sts_v2)
// - style: float (optional, 0.0–1.0, default: 0.0)
// - stability: float (optional, 0.0–1.0, default: 1.0)
// - remove_background_noise: bool (optional, default: false)
func ConvertVoiceAPI(c *gin.Context) {
	// 1. Authenticate user
	token := c.GetHeader("Authorization")
	_, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "unauthorized",
			"errorCode": config.ERROR_CODE_UNAUTHORIZED,
		})
		return
	}

	// 2. Read audio file
	fileHeader, err := c.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     "audio file is required",
			"errorCode": config.ERROR_CODE_BAD_REQUEST,
		})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     "failed to open audio file",
			"errorCode": config.ERROR_CODE_BAD_REQUEST,
		})
		return
	}
	defer file.Close()

	audioData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     "failed to read audio file",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	// 3. Parse parameters
	voiceID := c.PostForm("voice_id")
	if voiceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     "voice_id is required",
			"errorCode": config.ERROR_CODE_BAD_REQUEST,
		})
		return
	}

	modelID := c.PostForm("model_id")
	if modelID == "" {
		modelID = "eleven_multilingual_sts_v2"
	}

	style := 0.0
	if styleStr := c.PostForm("style"); styleStr != "" {
		if v, err := strconv.ParseFloat(styleStr, 64); err == nil {
			style = v
		}
	}

	stability := 1.0
	if stabilityStr := c.PostForm("stability"); stabilityStr != "" {
		if v, err := strconv.ParseFloat(stabilityStr, 64); err == nil {
			stability = v
		}
	}

	removeBackgroundNoise := false
	if noiseStr := c.PostForm("remove_background_noise"); noiseStr == "true" {
		removeBackgroundNoise = true
	}

	// 4. Get ElevenLabs API key
	apiKey := os.Getenv("ELEVENLABS_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     "ELEVENLABS_API_KEY not configured",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	// 5. Call ElevenLabs Speech-to-Speech
	service := services.NewElevenLabsService(apiKey)
	converted, err := service.SpeechToSpeech(
		voiceID,
		audioData,
		modelID,
		style,
		stability,
		removeBackgroundNoise,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     fmt.Sprintf("Failed to convert voice: %v", err),
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	// 6. Return converted audio as binary
	c.Data(http.StatusOK, "audio/mpeg", converted)
}
