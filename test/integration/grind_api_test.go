package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/database"
	"github.com/daniel0321forever/terriyaki-go/internal/models"
	"github.com/daniel0321forever/terriyaki-go/test"
	"github.com/stretchr/testify/assert"
)

// Helper function to permanently clean up test users
func cleanupGrindTestUser(email string) {
	var user models.User
	database.Db.Unscoped().Where("email = ?", email).Delete(&user)
}

func TestCreateGrindAPI(t *testing.T) {
	// Create test user
	testEmail := "test_grind_creator@example.com"
	cleanupGrindTestUser(testEmail)
	defer cleanupGrindTestUser(testEmail)

	_, _ = models.CreateUser("grindcreator", testEmail, "password123", "https://example.com/avatar.jpg")

	t.Run("Successful Grind Creation", func(t *testing.T) {
		// Login to get token
		loginReq := map[string]string{
			"email":    testEmail,
			"password": "password123",
		}
		loginRr := test.MakeRequest("POST", "/api/v1/login", loginReq, false)
		loginResponse := test.ParseResponseMap(t, loginRr.Body.Bytes())
		token := loginResponse["token"].(string)

		// Create grind request
		startDate := time.Now().AddDate(0, 0, 1).Format(time.RFC3339)
		requestBody := map[string]interface{}{
			"duration":     7,
			"budget":       100,
			"participants": []interface{}{testEmail},
			"startDate":    startDate,
		}

		// Execute with auth
		rr := test.MakeAuthenticatedRequest("POST", "/api/v1/grinds", requestBody, token)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)

		response := test.ParseResponseMap(t, rr.Body.Bytes())
		expected := map[string]interface{}{
			"message": "Grind created successfully",
			"grind": map[string]interface{}{
				"id":           "*",
				"duration":     float64(7), // JSON unmarshals numbers as float64
				"budget":       float64(100),
				"startDate":    "*",
				"participants": "*",
				"taskToday":    "*",
				"progress":     "*",
				"quitted":      false,
				"todayStats":   "*",
			},
		}
		test.AssertMapValues(t, response, expected, "")

		// Cleanup - delete created grind
		if grindData, ok := response["grind"].(map[string]interface{}); ok {
			if grindID, ok := grindData["id"].(string); ok {
				grind, _ := models.GetGrind(grindID)
				if grind != nil {
					database.Db.Delete(grind)
				}
			}
		}
	})
}

func TestGetAllUserGrindsAPI(t *testing.T) {
	// Create test user
	testEmail := "test_get_grinds@example.com"
	cleanupGrindTestUser(testEmail)
	defer cleanupGrindTestUser(testEmail)

	_, _ = models.CreateUser("getgrindsuser", testEmail, "password123", "https://example.com/avatar.jpg")

	t.Run("Get All User Grinds", func(t *testing.T) {
		// Login to get token
		loginReq := map[string]string{
			"email":    testEmail,
			"password": "password123",
		}
		loginRr := test.MakeRequest("POST", "/api/v1/login", loginReq, false)
		loginResponse := test.ParseResponseMap(t, loginRr.Body.Bytes())
		token := loginResponse["token"].(string)

		// Execute with auth
		rr := test.MakeAuthenticatedRequest("GET", "/api/v1/grinds", nil, token)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)

		response := test.ParseResponseMap(t, rr.Body.Bytes())
		expected := map[string]interface{}{
			"message": "Grinds fetched successfully",
			"grinds":  "*",
		}
		test.AssertMapValues(t, response, expected, "")
	})
}

func TestGetUserCurrentGrindAPI(t *testing.T) {
	// Create test user
	testEmail := "test_current_grind@example.com"
	cleanupGrindTestUser(testEmail)
	defer cleanupGrindTestUser(testEmail)

	_, _ = models.CreateUser("currentgrinduser", testEmail, "password123", "https://example.com/avatar.jpg")

	t.Run("Get Current Grind - No Grind", func(t *testing.T) {
		// Login to get token
		loginReq := map[string]string{
			"email":    testEmail,
			"password": "password123",
		}
		loginRr := test.MakeRequest("POST", "/api/v1/login", loginReq, false)
		loginResponse := test.ParseResponseMap(t, loginRr.Body.Bytes())
		token := loginResponse["token"].(string)

		// Execute with auth
		rr := test.MakeAuthenticatedRequest("GET", "/api/v1/grinds/current", nil, token)

		// Assert - should return 404 when no current grind
		assert.Equal(t, http.StatusNotFound, rr.Code)

		response := test.ParseResponseMap(t, rr.Body.Bytes())
		expected := map[string]interface{}{
			"message": "No current grind found",
		}
		test.AssertMapValues(t, response, expected, "")
	})
}

func TestQuitGrindAPI(t *testing.T) {
	// Create test user
	testEmail := "test_quit_grind@example.com"
	cleanupGrindTestUser(testEmail)
	defer cleanupGrindTestUser(testEmail)

	_, _ = models.CreateUser("quitgrinduser", testEmail, "password123", "https://example.com/avatar.jpg")

	t.Run("Quit Grind - Unauthorized", func(t *testing.T) {
		// Execute without auth
		rr := test.MakeRequest("POST", "/api/v1/grinds/some-grind-id/quit", nil, false)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Quit Grind - Not Found", func(t *testing.T) {
		// Login to get token
		loginReq := map[string]string{
			"email":    testEmail,
			"password": "password123",
		}
		loginRr := test.MakeRequest("POST", "/api/v1/login", loginReq, false)
		loginResponse := test.ParseResponseMap(t, loginRr.Body.Bytes())
		token := loginResponse["token"].(string)

		// Execute with auth but non-existent grind
		rr := test.MakeAuthenticatedRequest("POST", "/api/v1/grinds/nonexistent-grind-id/quit", nil, token)

		// Assert - should return error when grind not found
		// The actual status code depends on implementation
		assert.NotEqual(t, http.StatusOK, rr.Code)
	})
}

func TestGetProgressRecordsAPI(t *testing.T) {
	// Create test user
	testEmail := "test_progress@example.com"
	cleanupGrindTestUser(testEmail)
	defer cleanupGrindTestUser(testEmail)

	_, _ = models.CreateUser("progressuser", testEmail, "password123", "https://example.com/avatar.jpg")

	t.Run("Get Progress Records - Unauthorized", func(t *testing.T) {
		// Execute without auth
		rr := test.MakeRequest("GET", "/api/v1/grinds/some-grind-id/progress", nil, false)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Get Progress Records - Not Found", func(t *testing.T) {
		// Login to get token
		loginReq := map[string]string{
			"email":    testEmail,
			"password": "password123",
		}
		loginRr := test.MakeRequest("POST", "/api/v1/login", loginReq, false)
		loginResponse := test.ParseResponseMap(t, loginRr.Body.Bytes())
		token := loginResponse["token"].(string)

		// Execute with auth but non-existent grind
		rr := test.MakeAuthenticatedRequest("GET", "/api/v1/grinds/nonexistent-grind-id/progress", nil, token)

		// Assert - should return error when grind not found
		assert.NotEqual(t, http.StatusOK, rr.Code)
	})
}

func TestDeleteAllGrindsAPI(t *testing.T) {
	// Create test user
	testEmail := "test_delete_grinds@example.com"
	cleanupGrindTestUser(testEmail)
	defer cleanupGrindTestUser(testEmail)

	_, _ = models.CreateUser("deletegrindsuser", testEmail, "password123", "https://example.com/avatar.jpg")

	t.Run("Delete All Grinds - No Auth Required", func(t *testing.T) {
		// Execute without auth - this endpoint doesn't require authentication (it's for testing)
		rr := test.MakeRequest("DELETE", "/api/v1/grinds/delete-all", nil, false)

		// Assert - should return 200 even without auth (it's a test/admin endpoint)
		assert.Equal(t, http.StatusOK, rr.Code)

		response := test.ParseResponseMap(t, rr.Body.Bytes())
		expected := map[string]interface{}{
			"message": "All grinds deleted successfully",
		}
		test.AssertMapValues(t, response, expected, "")
	})

	t.Run("Delete All Grinds - Successful", func(t *testing.T) {
		// Login to get token
		loginReq := map[string]string{
			"email":    testEmail,
			"password": "password123",
		}
		loginRr := test.MakeRequest("POST", "/api/v1/login", loginReq, false)
		loginResponse := test.ParseResponseMap(t, loginRr.Body.Bytes())
		token := loginResponse["token"].(string)

		// Execute with auth
		rr := test.MakeAuthenticatedRequest("DELETE", "/api/v1/grinds/delete-all", nil, token)

		// Assert - should succeed even if no grinds exist
		assert.Equal(t, http.StatusOK, rr.Code)

		response := test.ParseResponseMap(t, rr.Body.Bytes())
		expected := map[string]interface{}{
			"message": "All grinds deleted successfully",
		}
		test.AssertMapValues(t, response, expected, "")
	})
}

func TestGetGrindAPI(t *testing.T) {
	// Create test user
	testEmail := "test_get_grind@example.com"
	cleanupGrindTestUser(testEmail)
	defer cleanupGrindTestUser(testEmail)

	_, _ = models.CreateUser("getgrinduser", testEmail, "password123", "https://example.com/avatar.jpg")

	t.Run("Get Grind - Unauthorized", func(t *testing.T) {
		// Execute without auth
		rr := test.MakeRequest("GET", "/api/v1/grinds/some-grind-id", nil, false)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Get Grind - Not Found", func(t *testing.T) {
		// Login to get token
		loginReq := map[string]string{
			"email":    testEmail,
			"password": "password123",
		}
		loginRr := test.MakeRequest("POST", "/api/v1/login", loginReq, false)
		loginResponse := test.ParseResponseMap(t, loginRr.Body.Bytes())
		token := loginResponse["token"].(string)

		// Execute with auth but non-existent grind
		rr := test.MakeAuthenticatedRequest("GET", "/api/v1/grinds/nonexistent-grind-id", nil, token)

		// Assert - should return 404 when grind not found
		assert.Equal(t, http.StatusNotFound, rr.Code)

		response := test.ParseResponseMap(t, rr.Body.Bytes())
		expected := map[string]interface{}{
			"message": "grind not found",
		}
		test.AssertMapValues(t, response, expected, "")
	})
}
