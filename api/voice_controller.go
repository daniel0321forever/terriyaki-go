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

// ConvertVoiceAPI converts audio to a different voice using ElevenLabs Speech-to-Speech
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

	// 2. Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     "Failed to parse multipart form",
			"errorCode": config.ERROR_CODE_BAD_REQUEST,
		})
		return
	}

	// 3. Get audio file
	audioFiles := form.File["audio"]
	if len(audioFiles) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     "No audio file provided",
			"errorCode": config.ERROR_CODE_BAD_REQUEST,
		})
		return
	}

	audioFile := audioFiles[0]
	file, err := audioFile.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     "Failed to open audio file",
			"errorCode": config.ERROR_CODE_BAD_REQUEST,
		})
		return
	}
	defer file.Close()

	// Read audio data
	audioData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     "Failed to read audio file",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	// 4. Get parameters from form
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
		modelID = "eleven_multilingual_sts_v2" // Default
	}

	// Parse style (default: 0.0)
	style := 0.0
	if styleStr := c.PostForm("style"); styleStr != "" {
		style, err = strconv.ParseFloat(styleStr, 64)
		if err != nil {
			style = 0.0
		}
	}

	// Parse stability (default: 1.0)
	stability := 1.0
	if stabilityStr := c.PostForm("stability"); stabilityStr != "" {
		stability, err = strconv.ParseFloat(stabilityStr, 64)
		if err != nil {
			stability = 1.0
		}
	}

	// Parse remove_background_noise (default: false)
	removeBackgroundNoise := false
	if noiseStr := c.PostForm("remove_background_noise"); noiseStr == "true" {
		removeBackgroundNoise = true
	}

	// 5. Get ElevenLabs API key
	apiKey := os.Getenv("ELEVENLABS_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     "ELEVENLABS_API_KEY not configured",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	// 6. Call ElevenLabs service
	service := services.NewElevenLabsService(apiKey)
	convertedAudio, err := service.SpeechToSpeech(
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

	// 7. Return audio file
	c.Data(http.StatusOK, "audio/mpeg", convertedAudio)
}

