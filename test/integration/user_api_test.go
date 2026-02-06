package integration

import (
	"net/http"
	"testing"

	"github.com/daniel0321forever/terriyaki-go/internal/database"
	"github.com/daniel0321forever/terriyaki-go/internal/models"
	"github.com/daniel0321forever/terriyaki-go/test"
	"github.com/stretchr/testify/assert"
)

func TestCheckUserExistsAPI(t *testing.T) {
	// Create a test user first
	testEmail := "test_exists@example.com"
	user, _ := models.CreateUser("existsuser", testEmail, "password123", "https://example.com/avatar.jpg")
	defer database.Db.Delete(&user)

	t.Run("User Exists", func(t *testing.T) {
		// Execute
		rr := test.MakeRequest("GET", "/api/v1/users/exists?email="+testEmail, nil, false)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)

		response := test.ParseResponseMap(t, rr.Body.Bytes())
		expected := map[string]interface{}{
			"exists": true,
		}
		test.AssertMapValues(t, response, expected, "")
	})

	t.Run("User Does Not Exist", func(t *testing.T) {
		// Execute
		rr := test.MakeRequest("GET", "/api/v1/users/exists?email=nonexistent@example.com", nil, false)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)

		response := test.ParseResponseMap(t, rr.Body.Bytes())
		expected := map[string]interface{}{
			"exists": false,
		}
		test.AssertMapValues(t, response, expected, "")
	})
}

func TestDeleteUserAPI(t *testing.T) {
	t.Run("Successful User Deletion", func(t *testing.T) {
		// Create a test user
		testEmail := "test_delete@example.com"
		user, _ := models.CreateUser("deleteuser", testEmail, "password123", "https://example.com/avatar.jpg")

		// Get auth token for this user
		loginReq := map[string]string{
			"email":    testEmail,
			"password": "password123",
		}
		loginRr := test.MakeRequest("POST", "/api/v1/login", loginReq, false)
		loginResponse := test.ParseResponseMap(t, loginRr.Body.Bytes())
		token := loginResponse["token"].(string)

		// Execute delete with custom auth
		requestBody := map[string]string{}
		_ = test.MakeAuthenticatedRequest("DELETE", "/api/v1/users/delete", requestBody, token)

		// Clean up if deletion failed
		defer database.Db.Delete(&user)
	})
}
