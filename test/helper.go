package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/daniel0321forever/terriyaki-go/internal/application/services"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/migrate"
	"github.com/daniel0321forever/terriyaki-go/internal/infrastructure/db/postgres"
	"github.com/daniel0321forever/terriyaki-go/internal/interface/api"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func Router() *gin.Engine {
	db := postgres.Db

	router := gin.Default()

	// Initialize repositories
	userRepo := postgres.NewGormUserRepository(db)
	grindRepo := postgres.NewGormGrindRepository(db)
	taskRepo := postgres.NewGormTaskRepository(db)
	participationRepo := postgres.NewGormParticipationRepository(db)
	messageRepo := postgres.NewGormMessageRepository(db)
	interviewSessionRepo := postgres.NewGormInterviewSessionRepository(db)
	paymentInfoRepo := postgres.NewGormStripePaymentInfoRepository(db)

	// Initialize services
	userService := services.NewUserService(userRepo)
	grindService := services.NewGrindService(grindRepo, userRepo, taskRepo, participationRepo, messageRepo)
	taskService := services.NewTaskService(taskRepo)
	messageService := services.NewMessageService(messageRepo, userRepo)
	interviewService := services.NewInterviewService(interviewSessionRepo)
	paymentService := services.NewStripePaymentService(userRepo, grindRepo, participationRepo, paymentInfoRepo)
	// Initialize API handlers with services
	grindCtrl := api.NewGrindController(grindService, userService, messageService)
	userCtrl := api.NewUserController(grindService, userService)
	healthCtrl := api.NewHealthController()
	taskCtrl := api.NewTaskController(taskService, grindService)
	messageCtrl := api.NewMessageController(userService, messageService, grindService)
	interviewCtrl := api.NewInterviewController(interviewService, userService, taskService)
	paymentCtrl := api.NewPaymentController(userService, paymentService)

	// define routes
	v1 := router.Group("/api/v1")
	{
		v1.POST("grinds", grindCtrl.CreateGrindAPI)
		v1.GET("grinds", grindCtrl.GetAllUserGrindsAPI)
		v1.DELETE("grinds/delete-all", grindCtrl.DeleteAllGrindsAPI)
		v1.GET("grinds/current", grindCtrl.GetUserCurrentGrindAPI)
		v1.GET("grinds/:id", grindCtrl.GetGrindAPI)
		v1.GET("ping", healthCtrl.PingAPI)
		v1.POST("register", userCtrl.RegisterAPI)
		v1.POST("login", userCtrl.LoginAPI)
		v1.POST("logout", userCtrl.LogoutAPI)
		v1.GET("verify-token", userCtrl.VerifyTokenAPI)
		v1.POST("tasks/finish", taskCtrl.FinishTodayTaskAPI)
		v1.GET("tasks/today", taskCtrl.GetTodayTaskAPI)
		v1.GET("tasks/:id", taskCtrl.GetTaskAPI)
		v1.GET("messages", messageCtrl.GetMessageAPI)
		v1.POST("messages/:id/read", messageCtrl.ReadMessageAPI)
		v1.POST("messages/:id/invitation/create", messageCtrl.CreateInvitationAPI)
		v1.POST("messages/:id/invitation/accept", messageCtrl.AcceptInvitationAPI)
		v1.POST("messages/:id/invitation/reject", messageCtrl.RejectInvitationAPI)
		v1.POST("interviews/llm", interviewCtrl.LLMWebhookAPI)
		v1.POST("interviews/start", interviewCtrl.StartInterviewAPI)
		v1.POST("interviews/:id/response", interviewCtrl.SaveAgentResponseAPI)
		v1.POST("interviews/:id/end", interviewCtrl.EndInterviewAPI)
		v1.POST("payments/payment-intent", paymentCtrl.PaymentIntentAPI)
		v1.POST("payments/save-card-intent", paymentCtrl.SaveCardIntentAPI)
		v1.POST("payments/save-card", paymentCtrl.SaveCardAPI)
		v1.POST("payments/force-charging", paymentCtrl.ForceInvestigateDuedPenaltyAPI)
		v1.GET("payments/methods", paymentCtrl.GetAvailablePaymentMethodsAPI)
		v1.POST("payments/methods/select-default", paymentCtrl.SelectPaymentMethodAPI)
		v1.POST("payments/force-charging", paymentCtrl.TestForceChargingAPI) // TODO: for testing only
	}

	v2 := router.Group("/api/v2")
	{
		v2.POST("login", userCtrl.LoginAPIV2)
		v2.GET("verify-token", userCtrl.VerifyTokenAPIV2)
	}

	return router
}

func Setup() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// connect to test database
	_, err = postgres.ConnectTestDB()
	if err != nil {
		panic(err)
	}

	err = migrate.MigrateDatabase()
	if err != nil {
		panic(err)
	}
}

func Teardown() {
	migrator := postgres.Db.Migrator()
	migrator.DropTable(&postgres.UserSchema{})
	migrator.DropTable(&postgres.GrindSchema{})
	migrator.DropTable(&postgres.TaskSchema{})
	migrator.DropTable(&postgres.ParticipationSchema{})
	migrator.DropTable(&postgres.MessageSchema{})
	migrator.DropTable(&postgres.InterviewSessionSchema{})
	migrator.DropTable(&postgres.StripePaymentInfoSchema{})
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
func AssertMapValues(t interface {
	Errorf(format string, args ...interface{})
}, actual map[string]interface{}, expected map[string]interface{}, path string) bool {
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
func ParseResponseMap(t interface {
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}, body []byte) map[string]interface{} {
	var response map[string]interface{}
	err := json.Unmarshal(body, &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	return response
}
