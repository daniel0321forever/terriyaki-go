package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/config"
	"github.com/daniel0321forever/terriyaki-go/internal/database"
	"github.com/daniel0321forever/terriyaki-go/internal/models"
	"github.com/daniel0321forever/terriyaki-go/internal/services"
	"github.com/daniel0321forever/terriyaki-go/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Just stores conversation, lets ElevenLabs handle responses
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
	session, err := models.GetInterviewSession(req.SessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	// 3. Check if session is still active
	if session.Status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session is not active"})
		return
	}

	// 4. Parse conversation history
	var conversationHistory []map[string]interface{}
	if session.ConversationHistory != nil {
		json.Unmarshal(session.ConversationHistory, &conversationHistory)
	}

	// 5. Add user message to history
	conversationHistory = append(conversationHistory, map[string]interface{}{
		"role":    "user",
		"message": req.TranscribedText,
		"time":    time.Now().Format(time.RFC3339),
	})

	fmt.Printf("[WEBHOOK] User message received for session %s: %s\n", req.SessionID, req.TranscribedText)
	fmt.Printf("[WEBHOOK] Total messages in history: %d\n", len(conversationHistory))

	// 6. Count user responses to check if interview should end
	userMessageCount := 0
	for _, msg := range conversationHistory {
		if msg["role"] == "user" {
			userMessageCount++
		}
	}

	fmt.Printf("[WEBHOOK] User response count: %d/4\n", userMessageCount)

	// 7. Check if we've reached the limit (4 user responses)
	if userMessageCount >= 4 {
		// End the interview automatically after candidate's 4th response
		session.Status = "completed"
		now := time.Now()
		session.EndedAt = &now

		// Save conversation history
		historyBytes, _ := json.Marshal(conversationHistory)
		session.ConversationHistory = datatypes.JSON(historyBytes)
		database.Db.Save(session)

		c.JSON(http.StatusOK, gin.H{
			"response": "Thank you for the interview! The session has ended. Please check your results.",
			"ended":    true,
		})
		return
	}

	// 8. Save updated conversation history (without calling Gemini)
	historyBytes, _ := json.Marshal(conversationHistory)
	session.ConversationHistory = datatypes.JSON(historyBytes)
	database.Db.Save(session)

	// 9. Return empty response - let ElevenLabs system prompt handle the response
	// OR return a simple acknowledgment if ElevenLabs requires a response
	c.JSON(http.StatusOK, gin.H{
		"response": "", // Empty - ElevenLabs will use its system prompt
		"ended":    false,
	})
}

// Helper function to format conversation history for prompt
func formatConversationHistory(history []map[string]interface{}) string {
	var formatted strings.Builder
	for i, msg := range history {
		if i > 0 {
			formatted.WriteString("\n")
		}
		role := msg["role"].(string)
		message := msg["message"].(string)
		formatted.WriteString(fmt.Sprintf("%s: %s", role, message))
	}
	return formatted.String()
}

// Try to query the API using curl
// curl -X POST http://localhost:8080/api/v1/interviews/llm \
//   -H "Content-Type: application/json" \
//   -d '{
//     "transcribed_text": "I will use a hash map to solve this problem",
//     "session_id": "test-session-123",
//     "conversation_id": "test-conv-456"
//   }'

// Creates interview session and returns agent config
func StartInterviewAPI(c *gin.Context) {
	// 1. Authenticate user
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	_, err = models.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message":   "user not found",
			"errorCode": config.ERROR_CODE_NOT_FOUND,
		})
		return
	}

	// 2. Get task ID from request
	var body struct {
		TaskID string `json:"task_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 3. Get task details
	task, err := models.GetTaskByID(body.TaskID, false)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	// 4. Create interview session
	session, err := models.CreateInterviewSession(userID, task.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}
	sessionID := session.ID

	// 5. Generate agent token (simple UUID for now)
	agentToken := uuid.New().String()

	// 6. Get ElevenLabs agent ID from environment
	agentID := os.Getenv("ELEVENLABS_AGENT_ID")
	if agentID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no ElevenLabs API key in .env"})
		return
	}

	// 7. Build LLM endpoint URL
	llmEndpoint := os.Getenv("BACKEND_URL") + "/api/v1/interviews/llm"

	// 8. Return configuration
	c.JSON(http.StatusOK, gin.H{
		"agent_id":     agentID,
		"llm_endpoint": llmEndpoint,
		"token":        agentToken,
		"session_id":   sessionID,
	})
}

// Ends an interview session and generates evaluation
func EndInterviewAPI(c *gin.Context) {
	// 1. Authenticate user
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	sessionID := c.Param("id")

	// 2. Get interview session
	session, err := models.GetInterviewSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":     "session not found",
			"errorCode": config.ERROR_CODE_NOT_FOUND,
		})
		return
	}

	// 3. Verify session belongs to user
	if session.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
		return
	}

	// 4. Check if already completed to prevent multiple evaluations
	if session.Status == "completed" {
		fmt.Printf("[END_INTERVIEW] Session %s already completed, returning existing data\n", sessionID)
		c.JSON(http.StatusOK, gin.H{
			"message":    "Interview already ended",
			"session_id": sessionID,
			"transcript": session.ConversationHistory,
			"evaluation": map[string]interface{}{
				"score":        0,
				"feedback":     "Evaluation already generated",
				"strengths":    []string{},
				"improvements": []string{},
			},
		})
		return
	}

	// 5. Update session status
	now := time.Now()
	session.Status = "completed"
	session.EndedAt = &now

	// 6. Generate evaluation from conversation history
	fmt.Printf("[END_INTERVIEW] Generating evaluation for session %s\n", sessionID)
	evaluation := generateEvaluation(session)

	// 7. Save session
	database.Db.Save(session)

	// 8. Return evaluation and transcript
	c.JSON(http.StatusOK, gin.H{
		"message":    "Interview ended successfully",
		"session_id": sessionID,
		"transcript": session.ConversationHistory,
		"evaluation": evaluation,
	})
}

// Saves agent response messages from frontend
func SaveAgentResponseAPI(c *gin.Context) {
	// 1. Authenticate user
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	sessionID := c.Param("id")

	// 2. Get interview session
	session, err := models.GetInterviewSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// 3. Verify session belongs to user
	if session.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
		return
	}

	// 4. Parse request
	var body struct {
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 5. Parse conversation history
	var conversationHistory []map[string]interface{}
	if session.ConversationHistory != nil {
		json.Unmarshal(session.ConversationHistory, &conversationHistory)
	}

	// 6. Add agent message to history
	conversationHistory = append(conversationHistory, map[string]interface{}{
		"role":    "agent",
		"message": body.Message,
		"time":    time.Now().Format(time.RFC3339),
	})

	// 7. Count agent messages to determine if we should end
	agentMessageCount := 0
	for _, msg := range conversationHistory {
		if msg["role"] == "agent" {
			agentMessageCount++
		}
	}

	// 8. Check if we've reached the limit (4 agent messages = intro + 3 questions)
	shouldEnd := agentMessageCount >= 4

	fmt.Printf("[CONVERSATION] agent message %d/%d saved for session %s (should_end: %v)\n",
		agentMessageCount, 4, sessionID, shouldEnd)

	// 9. Save conversation history
	historyBytes, _ := json.Marshal(conversationHistory)
	session.ConversationHistory = datatypes.JSON(historyBytes)
	database.Db.Save(session)

	// 10. Return success - signal frontend to end after user's next response
	c.JSON(http.StatusOK, gin.H{
		"message":       "Message saved",
		"message_count": agentMessageCount,
		"should_end":    shouldEnd,
	})
}

// Generates evaluation using Gemini
func generateEvaluation(session *models.InterviewSession) map[string]interface{} {
	// Parse conversation history
	var conversationHistory []map[string]interface{}
	if session.ConversationHistory != nil {
		json.Unmarshal(session.ConversationHistory, &conversationHistory)
	}

	// Get task context
	task, err := models.GetTaskByID(session.TaskID, false)
	if err != nil {
		return map[string]interface{}{
			"score":        0,
			"feedback":     "Unable to generate evaluation: task not found",
			"strengths":    []string{},
			"improvements": []string{},
		}
	}

	// Build comprehensive evaluation prompt
	conversationText := formatConversationHistory(conversationHistory)

	prompt := fmt.Sprintf(`
You are an expert coding interview evaluator. Analyze the following interview conversation and provide a comprehensive evaluation.

Problem: %s
Difficulty: %s

Candidate's submitted code:
%s

Full Conversation:
%s

Please provide:
1. A score from 0-100 based on:
   - Technical accuracy
   - Problem-solving approach
   - Communication clarity
   - Time/space complexity awareness
   - Code quality (if discussed)

2. Detailed feedback on their performance

3. List 2-3 key strengths

4. List 2-3 areas for improvement

Format your response as JSON:
{
  "score": <number>,
  "feedback": "<detailed feedback>",
  "strengths": ["strength1", "strength2"],
  "improvements": ["improvement1", "improvement2"]
}
`,
		*task.ProblemTitle,
		*task.ProblemDifficulty,
		*task.Code,
		conversationText,
	)

	// Call Gemini for evaluation
	fmt.Printf("[EVALUATION] Generating evaluation for session %s\n", session.ID)
	fmt.Printf("[EVALUATION] Conversation length: %d messages\n", len(conversationHistory))

	geminiService := services.NewGeminiService(os.Getenv("GEMINI_API_KEY"))
	response, err := geminiService.GenerateContent(prompt)

	if err != nil {
		fmt.Printf("[EVALUATION] ERROR: %v\n", err)
		return map[string]interface{}{
			"score":        0,
			"feedback":     fmt.Sprintf("Evaluation generation failed: %v", err),
			"strengths":    []string{},
			"improvements": []string{},
		}
	}

	fmt.Printf("[EVALUATION] SUCCESS - Generated evaluation\n")
	fmt.Printf("[EVALUATION] Response: %s\n", response)

	// Parse JSON response - handle markdown code blocks
	responseText := response
	// Remove markdown code blocks if present
	if strings.Contains(responseText, "```json") {
		start := strings.Index(responseText, "```json")
		end := strings.LastIndex(responseText, "```")
		if start >= 0 && end > start {
			responseText = responseText[start+7 : end] // Remove ```json and closing ```
			responseText = strings.TrimSpace(responseText)
		}
	} else if strings.Contains(responseText, "```") {
		// Handle generic code blocks
		start := strings.Index(responseText, "```")
		end := strings.LastIndex(responseText, "```")
		if start >= 0 && end > start {
			responseText = responseText[start+3 : end] // Remove ``` and closing ```
			responseText = strings.TrimSpace(responseText)
		}
	}

	var evaluation map[string]interface{}
	if err := json.Unmarshal([]byte(responseText), &evaluation); err != nil {
		fmt.Printf("[EVALUATION] JSON parsing error: %v\n", err)
		fmt.Printf("[EVALUATION] Attempting to parse cleaned text: %s\n", responseText)
		// If JSON parsing fails, return the raw response as feedback
		return map[string]interface{}{
			"score":        0,
			"feedback":     response,
			"strengths":    []string{},
			"improvements": []string{},
		}
	}

	return evaluation
}