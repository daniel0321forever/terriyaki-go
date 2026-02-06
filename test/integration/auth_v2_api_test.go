package integration

import (
	"net/http"
	"testing"

	"github.com/daniel0321forever/terriyaki-go/internal/database"
	"github.com/daniel0321forever/terriyaki-go/internal/models"
	"github.com/daniel0321forever/terriyaki-go/test"
	"github.com/stretchr/testify/assert"
)

func TestLoginV2Controller(t *testing.T) {
	// Create a test user
	testEmail := "test_login_v2@example.com"
	testPassword := "testpassword123"
	
	// Clean up any existing user first
	cleanupTestUser(testEmail)
	
	user, _ := models.CreateUser("loginv2user", testEmail, testPassword, "https://example.com/avatar.jpg")
	defer database.Db.Delete(&user)

	t.Run("Successful Login V2", func(t *testing.T) {
		// Create request body
		requestBody := map[string]string{
			"email":    testEmail,
			"password": testPassword,
		}

		// Execute
		rr := test.MakeRequest("POST", "/api/v2/login", requestBody, false)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)

		response := test.ParseResponseMap(t, rr.Body.Bytes())
		expected := map[string]interface{}{
			"message": "Login successful",
			"user": map[string]interface{}{
				"id":       "*",
				"username": "*",
				"email":    testEmail,
				"avatar":   "*",
			},
			"token":  "*",
			"grinds": "*",
		}
		test.AssertMapValues(t, response, expected, "")

		// Verify token is not empty
		assert.NotEmpty(t, response["token"])
		// Verify grinds is a map
		assert.IsType(t, map[string]interface{}{}, response["grinds"])
	})

	t.Run("Login V2 - Invalid Email", func(t *testing.T) {
		// Create request body with invalid email
		requestBody := map[string]string{
			"email":    "nonexistent_v2@example.com",
			"password": testPassword,
		}

		// Execute
		rr := test.MakeRequest("POST", "/api/v2/login", requestBody, false)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		response := test.ParseResponseMap(t, rr.Body.Bytes())
		expected := map[string]interface{}{
			"message": "invalid email",
		}
		test.AssertMapValues(t, response, expected, "")
	})

	t.Run("Login V2 - Invalid Password", func(t *testing.T) {
		// Create request body with invalid password
		requestBody := map[string]string{
			"email":    testEmail,
			"password": "wrongpassword",
		}

		// Execute
		rr := test.MakeRequest("POST", "/api/v2/login", requestBody, false)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		response := test.ParseResponseMap(t, rr.Body.Bytes())
		expected := map[string]interface{}{
			"message": "invalid password",
		}
		test.AssertMapValues(t, response, expected, "")
	})

	t.Run("Login V2 - Invalid Request Body", func(t *testing.T) {
		// Execute with invalid body
		rr := test.MakeRequest("POST", "/api/v2/login", "invalid", false)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestVerifyTokenV2Controller(t *testing.T) {
	// Create a test user
	testEmail := "test_verify_v2@example.com"
	
	// Clean up any existing user first
	cleanupTestUser(testEmail)
	
	user, _ := models.CreateUser("verifyv2user", testEmail, "password123", "https://example.com/avatar.jpg")
	defer database.Db.Delete(&user)

	t.Run("Successful Token Verification V2", func(t *testing.T) {
		// Login first to get token (using v2 login)
		loginReq := map[string]string{
			"email":    testEmail,
			"password": "password123",
		}
		loginRr := test.MakeRequest("POST", "/api/v2/login", loginReq, false)
		loginResponse := test.ParseResponseMap(t, loginRr.Body.Bytes())
		token := loginResponse["token"].(string)

		// Execute verify token with auth
		rr := test.MakeAuthenticatedRequest("GET", "/api/v2/verify-token", nil, token)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)

		response := test.ParseResponseMap(t, rr.Body.Bytes())
		expected := map[string]interface{}{
			"message": "Token verified",
			"user": map[string]interface{}{
				"id":       "*",
				"username": "*",
				"email":    testEmail,
				"avatar":   "*",
			},
			"grinds": "*",
		}
		test.AssertMapValues(t, response, expected, "")

		// Verify grinds is a map
		assert.IsType(t, map[string]interface{}{}, response["grinds"])
	})

	t.Run("Verify Token V2 - Unauthorized", func(t *testing.T) {
		// Execute verify token without auth
		rr := test.MakeRequest("GET", "/api/v2/verify-token", nil, false)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Verify Token V2 - Invalid Token", func(t *testing.T) {
		// Execute verify token with invalid token
		rr := test.MakeAuthenticatedRequest("GET", "/api/v2/verify-token", nil, "invalid-token-v2")

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}
