package api

import (
    "fmt"
    "net/http"
    "os"

    "github.com/daniel0321forever/terriyaki-go/internal/services"
    "github.com/gin-gonic/gin"
)

// LLMWebhookAPI is called by ElevenLabs Agent
func LLMWebhookAPI(c *gin.Context) {
    // 1. Parse request from ElevenLabs Agent
    var req struct {
        TranscribedText string                 `json:"transcribed_text"`
        SessionID       string                 `json:"session_id"`
        ConversationID  string                 `json:"conversation_id"`
        Context         map[string]interface{} `json:"context"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    // 2. Get interview session from database
    // TODO: Implement GetInterviewSession - for now, use a simple approach
    // session, err := models.GetInterviewSession(req.SessionID)
    // if err != nil {
    //     c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
    //     return
    // }

    // TEMPORARY: For testing, use a simple prompt without database
    prompt := fmt.Sprintf(`
You are an AI coding interview assistant helping a candidate practice.

Candidate just said: %s

Respond naturally as an interviewer would. Be helpful but challenging. Keep responses concise (2-3 sentences).
`, req.TranscribedText)

    // 3. Call Gemini API
    geminiService := services.NewGeminiService(os.Getenv("GEMINI_API_KEY"))
    response, err := geminiService.GenerateContent(prompt)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "LLM error", "details": err.Error()})
        return
    }

    // 4. Return response to ElevenLabs Agent
    c.JSON(http.StatusOK, gin.H{
        "response": response,
    })
}

// Try to query the API using curl
// curl -X POST http://localhost:8080/api/v1/interviews/llm \
//   -H "Content-Type: application/json" \
//   -d '{
//     "transcribed_text": "I will use a hash map to solve this problem",
//     "session_id": "test-session-123",
//     "conversation_id": "test-conv-456"
//   }'