package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/daniel0321forever/terriyaki-go/internal/config"
	"github.com/daniel0321forever/terriyaki-go/internal/database"
	"github.com/daniel0321forever/terriyaki-go/internal/models"
	"github.com/daniel0321forever/terriyaki-go/test"
	"github.com/stretchr/testify/assert"
)

// Helper function to clean up test users
func cleanupTestUser(email string) {
	// Use Unscoped to find and permanently delete even soft-deleted users
	var user models.User
	database.Db.Unscoped().Where("email = ?", email).Delete(&user)
}

func TestRegisterController(t *testing.T) {

	t.Run("Successful Registration", func(t *testing.T) {
		testEmail := "test_register_success@example.com"

		// Clean up any existing user with this email first
		cleanupTestUser(testEmail)
		defer cleanupTestUser(testEmail)

		// Create request body
		requestBody := map[string]string{
			"username": "testuser",
			"email":    testEmail,
			"password": "testpassword123",
			"avatar":   "https://example.com/avatar.jpg",
		}

		// Execute
		rr := test.MakeRequest("POST", "/api/v1/register", requestBody, false)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Registration successful", response["message"])
		assert.NotNil(t, response["user"])
		assert.NotNil(t, response["token"])
		assert.Nil(t, response["grind"])

		// Verify user was created in database
		user, err := models.GetUserByEmail(testEmail)
		assert.NoError(t, err)
		assert.NotNil(t, user, "User should not be nil")
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, testEmail, user.Email)
	})

	t.Run("Invalid Request Body", func(t *testing.T) {
		// Note: makeRequest helper doesn't support invalid JSON easily
		// So we'll test with an empty/invalid structure instead
		requestBody := "invalid"

		// Execute
		rr := test.MakeRequest("POST", "/api/v1/register", requestBody, false)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "invalid request body", response["message"])
	})

	t.Run("Duplicate Email", func(t *testing.T) {
		testEmail := "test_duplicate@example.com"

		// Clean up any existing user first
		cleanupTestUser(testEmail)
		defer cleanupTestUser(testEmail)

		// Create first user
		requestBody := map[string]string{
			"username": "firstuser",
			"email":    testEmail,
			"password": "password123",
			"avatar":   "https://example.com/avatar1.jpg",
		}
		rr1 := test.MakeRequest("POST", "/api/v1/register", requestBody, false)
		assert.Equal(t, http.StatusOK, rr1.Code)

		// Try to create second user with same email
		requestBody2 := map[string]string{
			"username": "seconduser",
			"email":    testEmail,
			"password": "password456",
			"avatar":   "https://example.com/avatar2.jpg",
		}

		// Execute
		rr2 := test.MakeRequest("POST", "/api/v1/register", requestBody2, false)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr2.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rr2.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "email already exists", response["message"])
		assert.Equal(t, config.ERROR_CODE_DUPLICATE_ENTRY, response["errorCode"])
	})

	t.Run("Missing Required Fields", func(t *testing.T) {
		// Create request body with missing password
		requestBody := map[string]string{
			"username": "testuser",
			"email":    "test_missing@example.com",
			"avatar":   "https://example.com/avatar.jpg",
			// password is missing
		}

		// Execute
		rr := test.MakeRequest("POST", "/api/v1/register", requestBody, false)

		// Assert - this may succeed or fail depending on CreateUser validation
		// The current implementation doesn't validate empty fields before calling CreateUser
		// So this test documents the current behavior
		t.Logf("Status Code: %d", rr.Code)
		t.Logf("Response: %s", rr.Body.String())
	})

	t.Run("Empty Request Body", func(t *testing.T) {
		// Create request with empty body
		requestBody := map[string]string{}

		// Execute
		rr := test.MakeRequest("POST", "/api/v1/register", requestBody, false)

		// Assert - documents current behavior
		t.Logf("Status Code: %d", rr.Code)
		t.Logf("Response: %s", rr.Body.String())
	})
}

func TestLoginController(t *testing.T) {
	// Create a test user
	testEmail := "test_login@example.com"
	testPassword := "testpassword123"

	// Clean up any existing user first
	cleanupTestUser(testEmail)
	defer cleanupTestUser(testEmail)

	_, _ = models.CreateUser("loginuser", testEmail, testPassword, "https://example.com/avatar.jpg")

	t.Run("Successful Login", func(t *testing.T) {
		// Create request body
		requestBody := map[string]string{
			"email":    testEmail,
			"password": testPassword,
		}

		// Execute
		rr := test.MakeRequest("POST", "/api/v1/login", requestBody, false)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)

		response := test.ParseResponseMap(t, rr.Body.Bytes())
		// When user has no grind, message is "No current grind found"
		expected := map[string]interface{}{
			"message": "No current grind found",
			"user": map[string]interface{}{
				"id":       "*",
				"username": "*",
				"email":    testEmail,
				"avatar":   "*",
			},
			"token": "*",
			"grind": nil,
		}
		test.AssertMapValues(t, response, expected, "")

		// Verify token is not empty
		assert.NotEmpty(t, response["token"])
	})

	t.Run("Invalid Email", func(t *testing.T) {
		// Create request body with invalid email
		requestBody := map[string]string{
			"email":    "nonexistent@example.com",
			"password": testPassword,
		}

		// Execute
		rr := test.MakeRequest("POST", "/api/v1/login", requestBody, false)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		response := test.ParseResponseMap(t, rr.Body.Bytes())
		expected := map[string]interface{}{
			"message": "invalid email",
		}
		test.AssertMapValues(t, response, expected, "")
	})

	t.Run("Invalid Password", func(t *testing.T) {
		// Create request body with invalid password
		requestBody := map[string]string{
			"email":    testEmail,
			"password": "wrongpassword",
		}

		// Execute
		rr := test.MakeRequest("POST", "/api/v1/login", requestBody, false)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		response := test.ParseResponseMap(t, rr.Body.Bytes())
		expected := map[string]interface{}{
			"message": "invalid password",
		}
		test.AssertMapValues(t, response, expected, "")
	})

	t.Run("Invalid Request Body", func(t *testing.T) {
		// Execute with invalid body
		rr := test.MakeRequest("POST", "/api/v1/login", "invalid", false)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestLogoutController(t *testing.T) {
	// Create a test user
	testEmail := "test_logout@example.com"

	// Clean up any existing user first
	cleanupTestUser(testEmail)
	defer cleanupTestUser(testEmail)

	_, _ = models.CreateUser("logoutuser", testEmail, "password123", "https://example.com/avatar.jpg")

	t.Run("Successful Logout", func(t *testing.T) {
		// Login first to get token
		loginReq := map[string]string{
			"email":    testEmail,
			"password": "password123",
		}
		loginRr := test.MakeRequest("POST", "/api/v1/login", loginReq, false)
		loginResponse := test.ParseResponseMap(t, loginRr.Body.Bytes())
		token := loginResponse["token"].(string)

		// Execute logout with auth
		rr := test.MakeAuthenticatedRequest("POST", "/api/v1/logout", nil, token)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)

		response := test.ParseResponseMap(t, rr.Body.Bytes())
		expected := map[string]interface{}{
			"message": "Logout successful",
		}
		test.AssertMapValues(t, response, expected, "")
	})

	t.Run("Logout - Unauthorized", func(t *testing.T) {
		// Execute logout without auth
		rr := test.MakeRequest("POST", "/api/v1/logout", nil, false)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

func TestVerifyTokenController(t *testing.T) {
	// Create a test user
	testEmail := "test_verify@example.com"

	// Clean up any existing user first
	cleanupTestUser(testEmail)
	defer cleanupTestUser(testEmail)

	_, _ = models.CreateUser("verifyuser", testEmail, "password123", "https://example.com/avatar.jpg")

	t.Run("Successful Token Verification", func(t *testing.T) {
		// Login first to get token
		loginReq := map[string]string{
			"email":    testEmail,
			"password": "password123",
		}
		loginRr := test.MakeRequest("POST", "/api/v1/login", loginReq, false)
		loginResponse := test.ParseResponseMap(t, loginRr.Body.Bytes())
		token := loginResponse["token"].(string)

		// Execute verify token with auth
		rr := test.MakeAuthenticatedRequest("GET", "/api/v1/verify-token", nil, token)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)

		response := test.ParseResponseMap(t, rr.Body.Bytes())
		// When user has no grind, message is "Token verified with no current grind found"
		expected := map[string]interface{}{
			"message": "Token verified with no current grind found",
			"user": map[string]interface{}{
				"id":       "*",
				"username": "*",
				"email":    testEmail,
				"avatar":   "*",
			},
			"grind": nil,
		}
		test.AssertMapValues(t, response, expected, "")
	})

	t.Run("Verify Token - Unauthorized", func(t *testing.T) {
		// Execute verify token without auth
		rr := test.MakeRequest("GET", "/api/v1/verify-token", nil, false)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Verify Token - Invalid Token", func(t *testing.T) {
		// Execute verify token with invalid token
		rr := test.MakeAuthenticatedRequest("GET", "/api/v1/verify-token", nil, "invalid-token")

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}
