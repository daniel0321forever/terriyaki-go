package integration

import (
	"net/http"
	"testing"

	"github.com/daniel0321forever/terriyaki-go/internal/database"
	"github.com/daniel0321forever/terriyaki-go/internal/models"
	"github.com/daniel0321forever/terriyaki-go/test"
	"github.com/stretchr/testify/assert"
)

func TestGetMessageAPI(t *testing.T) {
	// Create test user
	testEmail := "test_messages@example.com"
	user, _ := models.CreateUser("messagesuser", testEmail, "password123", "https://example.com/avatar.jpg")
	defer database.Db.Delete(&user)

	t.Run("Get Messages Successfully", func(t *testing.T) {
		// Login to get token
		loginReq := map[string]string{
			"email":    testEmail,
			"password": "password123",
		}
		loginRr := test.MakeRequest("POST", "/api/v1/login", loginReq, false)
		loginResponse := test.ParseResponseMap(t, loginRr.Body.Bytes())
		token := loginResponse["token"].(string)

		// Execute with auth
		rr := test.MakeAuthenticatedRequest("GET", "/api/v1/messages?offset=0&limit=10", nil, token)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)

		response := test.ParseResponseMap(t, rr.Body.Bytes())

		// Check basic structure
		assert.Equal(t, "Messages fetched successfully", response["message"])
		assert.NotNil(t, response["data"])

		// Data should be an array (even if empty)
		data, ok := response["data"].([]interface{})
		assert.True(t, ok, "data should be an array")

		// If there are messages, validate their structure
		if len(data) > 0 {
			messageValidation := map[string]interface{}{
				"id":                 "*",
				"content":            "*",
				"type":               "*",
				"read":               "*",
				"invitationAccepted": "*",
				"invitationRejected": "*",
				"createdAt":          "*",
				"sender": map[string]interface{}{
					"id":       "*",
					"username": "*",
					"email":    "*",
					"avatar":   "*",
				},
				"receiver": map[string]interface{}{
					"id":       "*",
					"username": "*",
					"email":    "*",
					"avatar":   "*",
				},
			}

			for i, msg := range data {
				if msgMap, ok := msg.(map[string]interface{}); ok {
					test.AssertMapValues(t, msgMap, messageValidation, "data["+string(rune(i))+"]")
				}
			}
		}
	})

	t.Run("Get Messages - Unauthorized", func(t *testing.T) {
		// Execute without auth
		rr := test.MakeRequest("GET", "/api/v1/messages", nil, false)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

func TestGetSentMessageAPI(t *testing.T) {
	// Create test user
	testEmail := "test_sent_messages@example.com"
	user, _ := models.CreateUser("sentmessagesuser", testEmail, "password123", "https://example.com/avatar.jpg")
	defer database.Db.Delete(&user)

	t.Run("Get Sent Messages Successfully", func(t *testing.T) {
		// Login to get token
		loginReq := map[string]string{
			"email":    testEmail,
			"password": "password123",
		}
		loginRr := test.MakeRequest("POST", "/api/v1/login", loginReq, false)
		loginResponse := test.ParseResponseMap(t, loginRr.Body.Bytes())
		token := loginResponse["token"].(string)

		// Execute with auth
		rr := test.MakeAuthenticatedRequest("GET", "/api/v1/messages/sent?offset=0&limit=10", nil, token)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)

		response := test.ParseResponseMap(t, rr.Body.Bytes())

		// Check basic structure
		assert.Equal(t, "Messages fetched successfully", response["message"])
		assert.NotNil(t, response["data"])

		// Data should be an array (even if empty)
		data, ok := response["data"].([]interface{})
		assert.True(t, ok, "data should be an array")

		// If there are messages, validate their structure
		if len(data) > 0 {
			messageValidation := map[string]interface{}{
				"id":                 "*",
				"content":            "*",
				"type":               "*",
				"read":               "*",
				"invitationAccepted": "*",
				"invitationRejected": "*",
				"createdAt":          "*",
				"sender": map[string]interface{}{
					"id":       "*",
					"username": "*",
					"email":    "*",
					"avatar":   "*",
				},
				"receiver": map[string]interface{}{
					"id":       "*",
					"username": "*",
					"email":    "*",
					"avatar":   "*",
				},
			}

			for i, msg := range data {
				if msgMap, ok := msg.(map[string]interface{}); ok {
					test.AssertMapValues(t, msgMap, messageValidation, "data["+string(rune(i))+"]")
				}
			}
		}
	})
}

func TestCreateInvitationAPI(t *testing.T) {
	// Create test users
	testEmail1 := "test_inviter@example.com"
	testEmail2 := "test_invitee@example.com"
	user1, _ := models.CreateUser("inviter", testEmail1, "password123", "https://example.com/avatar.jpg")
	user2, _ := models.CreateUser("invitee", testEmail2, "password123", "https://example.com/avatar.jpg")
	defer database.Db.Delete(&user1)
	defer database.Db.Delete(&user2)

	t.Run("Create Invitation - Missing Grind", func(t *testing.T) {
		// Login to get token
		loginReq := map[string]string{
			"email":    testEmail1,
			"password": "password123",
		}
		loginRr := test.MakeRequest("POST", "/api/v1/login", loginReq, false)
		loginResponse := test.ParseResponseMap(t, loginRr.Body.Bytes())
		token := loginResponse["token"].(string)

		// Create invitation request with non-existent grind
		requestBody := map[string]interface{}{
			"grindID":          "nonexistent-grind-id",
			"participantEmail": testEmail2,
		}

		// Execute with auth
		rr := test.MakeAuthenticatedRequest("POST", "/api/v1/messages/invitation", requestBody, token)

		// Assert - should return 404 when grind not found
		assert.Equal(t, http.StatusNotFound, rr.Code)

		response := test.ParseResponseMap(t, rr.Body.Bytes())
		expected := map[string]interface{}{
			"message": "grind not found",
		}
		test.AssertMapValues(t, response, expected, "")
	})
}

func TestReadMessageAPI(t *testing.T) {
	// Create test user
	testEmail := "test_read_message@example.com"
	user, _ := models.CreateUser("readmessageuser", testEmail, "password123", "https://example.com/avatar.jpg")
	defer database.Db.Delete(&user)

	t.Run("Read Message - Not Found", func(t *testing.T) {
		// Login to get token
		loginReq := map[string]string{
			"email":    testEmail,
			"password": "password123",
		}
		loginRr := test.MakeRequest("POST", "/api/v1/login", loginReq, false)
		loginResponse := test.ParseResponseMap(t, loginRr.Body.Bytes())
		token := loginResponse["token"].(string)

		// Execute with auth
		rr := test.MakeAuthenticatedRequest("POST", "/api/v1/messages/nonexistent-id/read", nil, token)

		// Assert - should return error when message not found
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
