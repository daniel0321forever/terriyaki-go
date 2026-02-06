package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/daniel0321forever/terriyaki-go/api"
	"github.com/daniel0321forever/terriyaki-go/internal/database"
	"github.com/daniel0321forever/terriyaki-go/internal/migrate"
	"github.com/daniel0321forever/terriyaki-go/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func Router() *gin.Engine {
	router := gin.Default()

	// define routes
	router.GET("/api/v1/ping", api.PingAPI)
	router.POST("/api/v1/register", api.RegisterAPI)
	router.POST("/api/v1/login", api.LoginAPI)
	router.POST("/api/v1/logout", api.LogoutAPI)
	router.GET("/api/v1/verify-token", api.VerifyTokenAPI)
	router.DELETE("/api/v1/users/delete", api.DeleteUserAPI)
	router.GET("/api/v1/users/exists", api.CheckUserExistsAPI)
	router.POST("/api/v1/grinds", api.CreateGrindAPI)
	router.GET("/api/v1/grinds", api.GetAllUserGrindsAPI)
	router.DELETE("/api/v1/grinds/delete-all", api.DeleteAllGrindsAPI)
	router.GET("/api/v1/grinds/current", api.GetUserCurrentGrindAPI)
	router.GET("/api/v1/grinds/:id", api.GetGrindAPI)
	router.POST("/api/v1/grinds/:id/quit", api.QuitGrindAPI)
	router.GET("/api/v1/grinds/:id/progress", api.GetProgressRecordsAPI)
	router.POST("/api/v1/tasks/finish", api.FinishTodayTaskAPI)
	router.GET("/api/v1/tasks/today", api.GetTodayTaskAPI)
	router.GET("/api/v1/tasks/:id", api.GetTaskAPI)
	router.GET("/api/v1/messages", api.GetMessageAPI)
	router.GET("/api/v1/messages/sent", api.GetSentMessageAPI)
	router.POST("/api/v1/messages/:id/read", api.ReadMessageAPI)
	router.POST("/api/v1/messages/invitation", api.CreateInvitationAPI)
	router.POST("/api/v1/messages/:id/invitation/accept", api.AcceptInvitationAPI)
	router.POST("/api/v1/messages/:id/invitation/reject", api.RejectInvitationAPI)
	router.POST("/api/v1/interviews/llm", api.LLMWebhookAPI)
	router.POST("/api/v1/interviews/start", api.StartInterviewAPI)
	router.POST("/api/v1/interviews/:id/response", api.SaveAgentResponseAPI)
	router.POST("/api/v1/interviews/:id/end", api.EndInterviewAPI)
	router.PATCH("/api/v1/profile", api.UpdateProfileAPI)
	router.POST("/api/v1/payments/payment-intent", api.PaymentIntentAPI)
	router.POST("/api/v1/payments/save-card-intent", api.SaveCardIntentAPI)
	router.POST("/api/v1/payments/save-card", api.SaveCardAPI)
	router.POST("/api/v1/charges/dued", api.ForceInvestigateDuedPenaltyAPI)
	router.GET("/api/v1/payments/methods", api.GetAvailablePaymentMethodsAPI)
	router.POST("/api/v1/payments/methods/select-default", api.SelectPaymentMethodAPI)
	router.POST("/api/v1/payments/force-charging", api.TestForceChargingAPI) // TODO: for testing only
	router.POST("/api/v2/login", api.LoginAPIV2)
	router.GET("/api/v2/verify-token", api.VerifyTokenAPIV2)

	return router
}

func Setup() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// connect to test database
	_, err = database.ConnectTestDB()
	if err != nil {
		panic(err)
	}

	err = migrate.MigrateDatabase()
	if err != nil {
		panic(err)
	}
}

func Teardown() {
	migrator := database.Db.Migrator()
	migrator.DropTable(&models.User{})
	migrator.DropTable(&models.Grind{})
	migrator.DropTable(&models.Task{})
	migrator.DropTable(&models.ParticipateRecord{})
	migrator.DropTable(&models.Message{})
	migrator.DropTable(&models.InterviewSession{})
	migrator.DropTable(&models.StripePaymentInfo{})
}

func MakeRequest(method, url string, body interface{}, isAuthenticatedRequest bool) *httptest.ResponseRecorder {
	requestBody, _ := json.Marshal(body)
	request, _ := http.NewRequest(method, url, bytes.NewBuffer(requestBody))
	if isAuthenticatedRequest {
		request.Header.Add("Authorization", "Bearer "+BearerToken())
	}
	writer := httptest.NewRecorder()
	Router().ServeHTTP(writer, request)
	return writer
}

// MakeAuthenticatedRequest makes an HTTP request with a custom bearer token
func MakeAuthenticatedRequest(method, url string, body interface{}, token string) *httptest.ResponseRecorder {
	requestBody, _ := json.Marshal(body)
	request, _ := http.NewRequest(method, url, bytes.NewBuffer(requestBody))
	request.Header.Add("Authorization", "Bearer "+token)
	writer := httptest.NewRecorder()
	Router().ServeHTTP(writer, request)
	return writer
}

func BearerToken() string {
	user := gin.H{
		"email":    "test@test.com",
		"password": "test",
	}

	writer := MakeRequest("POST", "/api/v1/login", user, false)
	var response map[string]string
	json.Unmarshal(writer.Body.Bytes(), &response)
	return response["token"]
}

// AssertMapValues recursively checks if expected values exist in actual map
// expectedMap can contain:
// - exact values to match
// - "*" to assert key exists with any value
// - nested maps for nested validation
func AssertMapValues(t interface{ Errorf(format string, args ...interface{}) }, actual map[string]interface{}, expected map[string]interface{}, path string) bool {
	allPassed := true
	
	for key, expectedValue := range expected {
		currentPath := path
		if currentPath != "" {
			currentPath += "."
		}
		currentPath += key
		
		actualValue, exists := actual[key]
		if !exists {
			t.Errorf("Key '%s' does not exist in actual map", currentPath)
			allPassed = false
			continue
		}
		
		// Check if expected is "*" (wildcard - just check existence)
		if expectedStr, ok := expectedValue.(string); ok && expectedStr == "*" {
			continue
		}
		
		// If expected value is nil, check actual is also nil
		if expectedValue == nil {
			if actualValue != nil {
				t.Errorf("At '%s': expected nil but got %v", currentPath, actualValue)
				allPassed = false
			}
			continue
		}
		
		// If expected is a map, recursively validate
		if expectedMap, ok := expectedValue.(map[string]interface{}); ok {
			actualMap, ok := actualValue.(map[string]interface{})
			if !ok {
				t.Errorf("At '%s': expected map but got %T", currentPath, actualValue)
				allPassed = false
				continue
			}
			if !AssertMapValues(t, actualMap, expectedMap, currentPath) {
				allPassed = false
			}
			continue
		}
		
		// If expected is an array/slice, validate each element
		if expectedSlice, ok := expectedValue.([]interface{}); ok {
			actualSlice, ok := actualValue.([]interface{})
			if !ok {
				t.Errorf("At '%s': expected array but got %T", currentPath, actualValue)
				allPassed = false
				continue
			}
			
			// If expected slice has items, validate them
			if len(expectedSlice) > 0 {
				// Check if first element is a validation map
				if validationMap, ok := expectedSlice[0].(map[string]interface{}); ok {
					// Validate each item in actual slice against the validation map
					for i, actualItem := range actualSlice {
						itemPath := fmt.Sprintf("%s[%d]", currentPath, i)
						if actualItemMap, ok := actualItem.(map[string]interface{}); ok {
							if !AssertMapValues(t, actualItemMap, validationMap, itemPath) {
								allPassed = false
							}
						}
					}
				}
			}
			continue
		}
		
		// Direct value comparison
		if actualValue != expectedValue {
			t.Errorf("At '%s': expected %v but got %v", currentPath, expectedValue, actualValue)
			allPassed = false
		}
	}
	
	return allPassed
}

// ParseResponseMap parses JSON response into map
func ParseResponseMap(t interface{ Errorf(format string, args ...interface{}); Fatalf(format string, args ...interface{}) }, body []byte) map[string]interface{} {
	var response map[string]interface{}
	err := json.Unmarshal(body, &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	return response
}
