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

func TestGetTodayTaskAPI(t *testing.T) {
	// Create test user
	testEmail := "test_task@example.com"
	user, _ := models.CreateUser("taskuser", testEmail, "password123", "https://example.com/avatar.jpg")
	defer database.Db.Delete(&user)

	t.Run("Get Today Task - No Grind", func(t *testing.T) {
		// Login to get token
		loginReq := map[string]string{
			"email":    testEmail,
			"password": "password123",
		}
		loginRr := test.MakeRequest("POST", "/api/v1/login", loginReq, false)
		loginResponse := test.ParseResponseMap(t, loginRr.Body.Bytes())
		token := loginResponse["token"].(string)

		// Execute with auth
		rr := test.MakeAuthenticatedRequest("GET", "/api/v1/tasks/today", nil, token)

		// Assert - should return 404 when no grind
		assert.Equal(t, http.StatusNotFound, rr.Code)

		response := test.ParseResponseMap(t, rr.Body.Bytes())
		expected := map[string]interface{}{
			"message": "Grind not found",
		}
		test.AssertMapValues(t, response, expected, "")
	})
}

func TestFinishTodayTaskAPI(t *testing.T) {
	// Create test user and grind
	testEmail := "test_finish_task@example.com"
	user, _ := models.CreateUser("finishtaskuser", testEmail, "password123", "https://example.com/avatar.jpg")
	defer database.Db.Delete(&user)

	// Create a grind with tasks
	startDate := time.Now().AddDate(0, 0, -1) // Start yesterday so we have a today task
	grind, _ := models.CreateGrind(7, 100, []interface{}{testEmail}, startDate)
	defer database.Db.Delete(grind)

	t.Run("Finish Today Task - No Auth", func(t *testing.T) {
		// Execute without auth
		requestBody := map[string]interface{}{
			"code":     "def solution():\n    return True",
			"language": "python",
		}
		rr := test.MakeRequest("POST", "/api/v1/tasks/finish", requestBody, false)

		// Assert - should return 401 unauthorized
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

func TestGetTaskAPI(t *testing.T) {
	// Create test user
	testEmail := "test_get_task@example.com"
	user, _ := models.CreateUser("gettaskuser", testEmail, "password123", "https://example.com/avatar.jpg")
	defer database.Db.Delete(&user)

	t.Run("Get Task - Not Found", func(t *testing.T) {
		// Login to get token
		loginReq := map[string]string{
			"email":    testEmail,
			"password": "password123",
		}
		loginRr := test.MakeRequest("POST", "/api/v1/login", loginReq, false)
		loginResponse := test.ParseResponseMap(t, loginRr.Body.Bytes())
		token := loginResponse["token"].(string)

		// Execute with auth
		rr := test.MakeAuthenticatedRequest("GET", "/api/v1/tasks/nonexistent-id", nil, token)

		// Assert - should return 404 when task not found
		assert.Equal(t, http.StatusNotFound, rr.Code)

		response := test.ParseResponseMap(t, rr.Body.Bytes())
		expected := map[string]interface{}{
			"message": "Task not found",
		}
		test.AssertMapValues(t, response, expected, "")
	})
}
