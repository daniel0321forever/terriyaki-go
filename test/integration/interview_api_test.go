package integration

import (
	"net/http"
	"testing"

	"github.com/daniel0321forever/terriyaki-go/internal/database"
	"github.com/daniel0321forever/terriyaki-go/internal/models"
	"github.com/daniel0321forever/terriyaki-go/test"
	"github.com/stretchr/testify/assert"
)

func TestStartInterviewAPI(t *testing.T) {
	// Create test user
	testEmail := "test_interview@example.com"
	user, _ := models.CreateUser("interviewuser", testEmail, "password123", "https://example.com/avatar.jpg")
	defer database.Db.Delete(&user)

	t.Run("Start Interview - Unauthorized", func(t *testing.T) {
		// Request body
		requestBody := map[string]string{
			"task_id": "some-task-id",
		}

		// Execute without auth
		rr := test.MakeRequest("POST", "/api/v1/interviews/start", requestBody, false)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Start Interview - Task Not Found", func(t *testing.T) {
		// Login to get token
		loginReq := map[string]string{
			"email":    testEmail,
			"password": "password123",
		}
		loginRr := test.MakeRequest("POST", "/api/v1/login", loginReq, false)
		loginResponse := test.ParseResponseMap(t, loginRr.Body.Bytes())
		token := loginResponse["token"].(string)

		// Request body with non-existent task
		requestBody := map[string]string{
			"task_id": "nonexistent-task-id",
		}

		// Execute with auth
		rr := test.MakeAuthenticatedRequest("POST", "/api/v1/interviews/start", requestBody, token)

		// Assert - should return 404 when task not found
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestEndInterviewAPI(t *testing.T) {
	// Create test user
	testEmail := "test_end_interview@example.com"
	user, _ := models.CreateUser("endinterviewuser", testEmail, "password123", "https://example.com/avatar.jpg")
	defer database.Db.Delete(&user)

	t.Run("End Interview - Unauthorized", func(t *testing.T) {
		// Execute without auth
		rr := test.MakeRequest("POST", "/api/v1/interviews/session-id/end", nil, false)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("End Interview - Session Not Found", func(t *testing.T) {
		// Login to get token
		loginReq := map[string]string{
			"email":    testEmail,
			"password": "password123",
		}
		loginRr := test.MakeRequest("POST", "/api/v1/login", loginReq, false)
		loginResponse := test.ParseResponseMap(t, loginRr.Body.Bytes())
		token := loginResponse["token"].(string)

		// Execute with auth
		rr := test.MakeAuthenticatedRequest("POST", "/api/v1/interviews/nonexistent-session/end", nil, token)

		// Assert - should return 404 when session not found
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestSaveAgentResponseAPI(t *testing.T) {
	// Create test user
	testEmail := "test_agent_response@example.com"
	user, _ := models.CreateUser("agentresponseuser", testEmail, "password123", "https://example.com/avatar.jpg")
	defer database.Db.Delete(&user)

	t.Run("Save Agent Response - Unauthorized", func(t *testing.T) {
		// Request body
		requestBody := map[string]string{
			"message": "Hello, this is the agent",
		}

		// Execute without auth
		rr := test.MakeRequest("POST", "/api/v1/interviews/session-id/response", requestBody, false)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Save Agent Response - Session Not Found", func(t *testing.T) {
		// Login to get token
		loginReq := map[string]string{
			"email":    testEmail,
			"password": "password123",
		}
		loginRr := test.MakeRequest("POST", "/api/v1/login", loginReq, false)
		loginResponse := test.ParseResponseMap(t, loginRr.Body.Bytes())
		token := loginResponse["token"].(string)

		// Request body
		requestBody := map[string]string{
			"message": "Hello, this is the agent",
		}

		// Execute with auth
		rr := test.MakeAuthenticatedRequest("POST", "/api/v1/interviews/nonexistent-session/response", requestBody, token)

		// Assert - should return 404 when session not found
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestLLMWebhookAPI(t *testing.T) {
	t.Run("LLM Webhook - Invalid Request", func(t *testing.T) {
		// Invalid request body
		requestBody := "invalid"

		// Execute
		rr := test.MakeRequest("POST", "/api/v1/interviews/llm", requestBody, false)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("LLM Webhook - Session Not Found", func(t *testing.T) {
		// Valid request structure but non-existent session
		requestBody := map[string]interface{}{
			"transcribed_text": "I think we should use a hash map",
			"session_id":       "nonexistent-session-id",
			"conversation_id":  "conv-123",
			"context":          map[string]interface{}{},
		}

		// Execute
		rr := test.MakeRequest("POST", "/api/v1/interviews/llm", requestBody, false)

		// Assert
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}
