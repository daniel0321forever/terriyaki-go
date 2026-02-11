package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/application/services"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/config"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InterviewController struct {
	interviewService *services.InterviewService
	userService      *services.UserService
	taskService      *services.TaskService
}

func NewInterviewController(
	is *services.InterviewService,
	us *services.UserService,
	ts *services.TaskService,
) *InterviewController {
	return &InterviewController{
		interviewService: is,
		userService:      us,
		taskService:      ts,
	}
}

// Just stores conversation, lets ElevenLabs handle responses
func (ctrl *InterviewController) LLMWebhookAPI(c *gin.Context) {
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
	getInterviewDTO := dto.GetInterviewSessionDTO{
		SessionID: req.SessionID,
	}
	sessionDTO, err := ctrl.interviewService.GetSession(getInterviewDTO)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	// 3. Check if session is still active
	if sessionDTO.Status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session is not active"})
		return
	}

	// 4. Parse conversation history
	var conversationHistory []map[string]interface{}
	if sessionDTO.ConversationHistory != nil {
		// ConversationHistory is already interface{} from DTO
		if history, ok := sessionDTO.ConversationHistory.([]interface{}); ok {
			for _, item := range history {
				if msg, ok := item.(map[string]interface{}); ok {
					conversationHistory = append(conversationHistory, msg)
				}
			}
		}
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
	now := time.Now()
	if userMessageCount >= 4 {
		// End the interview automatically after candidate's 4th response
		updateInterviewDTO := dto.UpdateInterviewSessionDTO{
			SessionID:           req.SessionID,
			Status:              stringPtr("completed"),
			ConversationHistory: conversationHistory,
			EndedAt:             &now,
		}
		_, err := ctrl.interviewService.UpdateSession(updateInterviewDTO)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update session"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"response": "Thank you for the interview! The session has ended. Please check your results.",
			"ended":    true,
		})
		return
	}

	// 8. Save updated conversation history (without calling Gemini)
	updateInterviewDTO := dto.UpdateInterviewSessionDTO{
		SessionID:           req.SessionID,
		ConversationHistory: conversationHistory,
	}
	_, err = ctrl.interviewService.UpdateSession(updateInterviewDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update session"})
		return
	}

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
func (ctrl *InterviewController) StartInterviewAPI(c *gin.Context) {
	// 1. Authenticate user
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	getUserDTO := dto.GetUserDTO{
		UserID: userID,
	}
	_, err = ctrl.userService.GetUser(getUserDTO)
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
	getTaskDTO := dto.GetTaskDTO{
		TaskID:             body.TaskID,
		SetProblemIfNeeded: false,
	}
	taskDTO, err := ctrl.taskService.GetTaskByID(getTaskDTO)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	// 4. Create interview session
	createInterviewDTO := dto.CreateInterviewSessionDTO{
		UserID: userID,
		TaskID: taskDTO.ID,
	}
	sessionDTO, err := ctrl.interviewService.CreateSession(createInterviewDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}
	sessionID := sessionDTO.ID

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
func (ctrl *InterviewController) EndInterviewAPI(c *gin.Context) {
	// 1. Authenticate user
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	sessionID := c.Param("id")

	// 2. Get interview session
	getInterviewDTO := dto.GetInterviewSessionDTO{
		SessionID: sessionID,
	}
	sessionDTO, err := ctrl.interviewService.GetSession(getInterviewDTO)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":     "session not found",
			"errorCode": config.ERROR_CODE_NOT_FOUND,
		})
		return
	}

	// 3. Verify session belongs to user
	if sessionDTO.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
		return
	}

	// 4. Check if already completed to prevent multiple evaluations
	if sessionDTO.Status == "completed" {
		fmt.Printf("[END_INTERVIEW] Session %s already completed, returning existing data\n", sessionID)
		c.JSON(http.StatusOK, gin.H{
			"message":    "Interview already ended",
			"session_id": sessionID,
			"transcript": sessionDTO.ConversationHistory,
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
	updateInterviewDTO := dto.UpdateInterviewSessionDTO{
		SessionID: sessionID,
		Status:    stringPtr("completed"),
		EndedAt:   &now,
	}
	updatedSessionDTO, err := ctrl.interviewService.UpdateSession(updateInterviewDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update session"})
		return
	}

	// 6. Generate evaluation from conversation history
	fmt.Printf("[END_INTERVIEW] Generating evaluation for session %s\n", sessionID)
	evaluation := ctrl.GenerateEvaluation(updatedSessionDTO)

	// 7. Return evaluation and transcript
	c.JSON(http.StatusOK, gin.H{
		"message":    "Interview ended successfully",
		"session_id": sessionID,
		"transcript": updatedSessionDTO.ConversationHistory,
		"evaluation": evaluation,
	})
}

// Saves agent response messages from frontend
func (ctrl *InterviewController) SaveAgentResponseAPI(c *gin.Context) {
	// 1. Authenticate user
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	sessionID := c.Param("id")

	// 2. Get interview session
	getInterviewDTO := dto.GetInterviewSessionDTO{
		SessionID: sessionID,
	}
	sessionDTO, err := ctrl.interviewService.GetSession(getInterviewDTO)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// 3. Verify session belongs to user
	if sessionDTO.UserID != userID {
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
	if sessionDTO.ConversationHistory != nil {
		// ConversationHistory is already interface{} from DTO
		if history, ok := sessionDTO.ConversationHistory.([]interface{}); ok {
			for _, item := range history {
				if msg, ok := item.(map[string]interface{}); ok {
					conversationHistory = append(conversationHistory, msg)
				}
			}
		}
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
	updateInterviewDTO := dto.UpdateInterviewSessionDTO{
		SessionID:           sessionID,
		ConversationHistory: conversationHistory,
	}
	_, err = ctrl.interviewService.UpdateSession(updateInterviewDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update session"})
		return
	}

	// 10. Return success - signal frontend to end after user's next response
	c.JSON(http.StatusOK, gin.H{
		"message":       "Message saved",
		"message_count": agentMessageCount,
		"should_end":    shouldEnd,
	})
}

// Generates evaluation using Gemini
func (ctrl *InterviewController) GenerateEvaluation(sessionDTO *dto.InterviewSessionDTO) map[string]interface{} {
	// Parse conversation history
	var conversationHistory []map[string]interface{}
	if sessionDTO.ConversationHistory != nil {
		// ConversationHistory is already interface{} from DTO
		if history, ok := sessionDTO.ConversationHistory.([]interface{}); ok {
			for _, item := range history {
				if msg, ok := item.(map[string]interface{}); ok {
					conversationHistory = append(conversationHistory, msg)
				}
			}
		}
	}

	// Get task context
	getTaskDTO := dto.GetTaskDTO{
		TaskID:             sessionDTO.TaskID,
		SetProblemIfNeeded: false,
	}
	taskDTO, err := ctrl.taskService.GetTaskByID(getTaskDTO)
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
		*taskDTO.ProblemTitle,
		*taskDTO.ProblemDifficulty,
		taskDTO.Code,
		conversationText,
	)

	// Call Gemini for evaluation
	fmt.Printf("[EVALUATION] Generating evaluation for session %s\n", sessionDTO.ID)
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

// Helper function to create a string pointer
func stringPtr(s string) *string {
	return &s
}
