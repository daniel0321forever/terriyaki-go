package integration

import (
	"net/http"
	"testing"

	"github.com/daniel0321forever/terriyaki-go/internal/database"
	"github.com/daniel0321forever/terriyaki-go/internal/models"
	"github.com/daniel0321forever/terriyaki-go/test"
	"github.com/stretchr/testify/assert"
)

func TestUpdateProfileAPI(t *testing.T) {
	// Create test user
	testEmail := "test_profile@example.com"
	user, _ := models.CreateUser("profileuser", testEmail, "password123", "https://example.com/avatar.jpg")
	defer database.Db.Delete(&user)

	t.Run("Successful Profile Update", func(t *testing.T) {
		// Login to get token
		loginReq := map[string]string{
			"email":    testEmail,
			"password": "password123",
		}
		loginRr := test.MakeRequest("POST", "/api/v1/login", loginReq, false)
		loginResponse := test.ParseResponseMap(t, loginRr.Body.Bytes())
		token := loginResponse["token"].(string)

		// Update profile request
		requestBody := map[string]string{
			"username": "newusername",
			"avatar":   "https://example.com/new-avatar.jpg",
		}

		// Execute with auth
		rr := test.MakeAuthenticatedRequest("PATCH", "/api/v1/profile", requestBody, token)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)

		response := test.ParseResponseMap(t, rr.Body.Bytes())
		expected := map[string]interface{}{
			"message": "Username updated successfully",
			"user": map[string]interface{}{
				"id":       "*",
				"username": "newusername",
				"email":    testEmail,
				"avatar":   "https://example.com/new-avatar.jpg",
			},
		}
		test.AssertMapValues(t, response, expected, "")
	})

	t.Run("Update Profile - Unauthorized", func(t *testing.T) {
		// Update profile request without auth
		requestBody := map[string]string{
			"username": "hackedusername",
		}

		// Execute without auth
		rr := test.MakeRequest("PATCH", "/api/v1/profile", requestBody, false)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Update Profile - Invalid Body", func(t *testing.T) {
		// Login to get token
		loginReq := map[string]string{
			"email":    testEmail,
			"password": "password123",
		}
		loginRr := test.MakeRequest("POST", "/api/v1/login", loginReq, false)
		loginResponse := test.ParseResponseMap(t, loginRr.Body.Bytes())
		token := loginResponse["token"].(string)

		// Invalid request body
		requestBody := "invalid"

		// Execute with auth
		rr := test.MakeAuthenticatedRequest("PATCH", "/api/v1/profile", requestBody, token)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
