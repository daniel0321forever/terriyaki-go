package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

type ElevenLabsService struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func NewElevenLabsService(apiKey string) *ElevenLabsService {
	return &ElevenLabsService{
		apiKey:  apiKey,
		baseURL: "https://api.elevenlabs.io/v1",
		client:  &http.Client{},
	}
}

// Text-to-Speech example
func (s *ElevenLabsService) TextToSpeech(voiceID string, text string) ([]byte, error) {
	url := fmt.Sprintf("%s/text-to-speech/%s", s.baseURL, voiceID)

	payload := map[string]interface{}{
		"text":     text,
		"model_id": "eleven_multilingual_v2",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("xi-api-key", s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// Speech-to-Text with multipart form data (improved version)
func (s *ElevenLabsService) SpeechToText(audioData []byte) (string, error) {
	url := fmt.Sprintf("%s/speech-to-text", s.baseURL)

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add model_id as a form field
	const sttModelID = "scribe_v1"
	err := writer.WriteField("model_id", sttModelID)
	if err != nil {
		return "", fmt.Errorf("failed to write model_id field: %w", err)
	}

	// Add audio file (API expects field name "file")
	part, err := writer.CreateFormFile("file", "test_audio.mp3")
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write(audioData); err != nil {
		return "", fmt.Errorf("failed to write audio data: %w", err)
	}

	writer.Close()

	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("xi-api-key", s.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Text string `json:"text"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to decode response: %w, body: %s", err, string(body))
	}

	return result.Text, nil
}

func main() {
	// Load .env file from backend root directory
	// Get the directory where this file is located

	_, filename, _, _ := runtime.Caller(0)
	utilsDir := filepath.Dir(filename)
	backendDir := filepath.Join(utilsDir, "..", "..")
	envPath := filepath.Join(backendDir, ".env")

	err := godotenv.Load(envPath)
	if err != nil {
		log.Printf("⚠️  Warning: Could not load .env file from %s: %v", envPath, err)
		log.Println("   Trying to load from current directory...")
		// Fallback to current directory
		if err := godotenv.Load(); err != nil {
			log.Fatalf("❌ Failed to load .env file: %v\n   Please make sure .env file exists in the backend directory", err)
		}
	}

	// Get API key from environment variable (now loaded from .env)
	apiKey := os.Getenv("ELEVENLABS_API_KEY")
	if apiKey == "" {
		log.Fatal("❌ ELEVENLABS_API_KEY not found in .env file. Please add it to your .env file.")
	}

	service := NewElevenLabsService(apiKey)

	// Step 1: Generate test audio using TTS
	fmt.Println("🎤 Step 1: Generating test audio using TTS...")

	// Use a default voice ID (Adam voice - you can change this)
	voiceID := "pNInz6obpgDQGcFmaJgB"

	// Test text - something relevant to coding interviews
	testText := "Hello, I'm going to solve the two sum problem. I'll use a hash map to store the complement of each number as I iterate through the array."

	fmt.Printf("   Original text: %s\n", testText)

	audioData, err := service.TextToSpeech(voiceID, testText)
	if err != nil {
		log.Fatalf("❌ Failed to generate audio: %v", err)
	}

	fmt.Printf("✅ Audio generated successfully (%d bytes)\n", len(audioData))

	// Optionally save the audio file for inspection
	if err := os.WriteFile("test_audio.mp3", audioData, 0644); err != nil {
		log.Printf("⚠️  Warning: Could not save audio file: %v", err)
	} else {
		fmt.Println("💾 Audio file saved as: test_audio.mp3")
	}

	// Step 2: Test Speech-to-Text
	fmt.Println("\n🎧 Step 2: Testing Speech-to-Text...")

	transcribedText, err := service.SpeechToText(audioData)
	if err != nil {
		log.Fatalf("❌ Failed to transcribe audio: %v", err)
	}

	fmt.Printf("✅ Transcription successful!\n")
	fmt.Printf("   Transcribed text: %s\n", transcribedText)

	// Step 3: Compare results
	fmt.Println("\n📊 Step 3: Comparison:")
	fmt.Printf("   Original:  %s\n", testText)
	fmt.Printf("   Transcribed: %s\n", transcribedText)

	// Note: STT might not match exactly due to natural speech variations
	if transcribedText == testText {
		fmt.Println("✅ Perfect match!")
	} else {
		fmt.Println("⚠️  Transcription differs (this is normal for STT)")
		fmt.Println("   The transcribed text may have slight variations due to speech patterns.")
	}

	fmt.Println("\n✨ Test completed successfully!")
}
