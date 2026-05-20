package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"

	"github.com/daniel0321forever/terriyaki-go/internal/infrastructure/db/postgres"
	"github.com/daniel0321forever/terriyaki-go/internal/interface/api"
	"github.com/daniel0321forever/terriyaki-go/migrations"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	solanaGo "github.com/gagliardetto/solana-go"
)

var (
	openAPIDocumentOnce sync.Once
	openAPIDocument     *openapi3.T
	openAPIDocumentErr  error

	openAPIRouterOnce sync.Once
	openAPIRouter     routers.Router
	openAPIRouterErr  error
)

const openAPISpecPath = "../../openapi.yaml"

func Router() *gin.Engine {
	db := postgres.Db

	router := gin.Default()

	api.RegisterRoutes(router, db)

	return router
}

func Setup() {
	envLoaded := false
	if wd, err := os.Getwd(); err == nil {
		for dir := wd; ; dir = filepath.Dir(dir) {
			envPath := filepath.Join(dir, ".env")
			if _, statErr := os.Stat(envPath); statErr == nil {
				if loadErr := godotenv.Load(envPath); loadErr != nil {
					log.Printf("Warning: found .env but failed to load %s: %v", envPath, loadErr)
				} else {
					envLoaded = true
				}
				break
			}

			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
		}
	}
	if !envLoaded {
		log.Printf("Warning: could not locate .env file from current working directory")
	}

	if os.Getenv("TEST_POSTGRES_HOST") == "" {
		_ = os.Setenv("TEST_POSTGRES_HOST", os.Getenv("POSTGRES_HOST"))
	}
	if os.Getenv("TEST_POSTGRES_PORT") == "" {
		_ = os.Setenv("TEST_POSTGRES_PORT", os.Getenv("POSTGRES_PORT"))
	}
	if os.Getenv("TEST_POSTGRES_USER") == "" {
		_ = os.Setenv("TEST_POSTGRES_USER", os.Getenv("POSTGRES_USER"))
	}
	if os.Getenv("TEST_POSTGRES_DB") == "" {
		_ = os.Setenv("TEST_POSTGRES_DB", os.Getenv("POSTGRES_DB"))
	}
	if os.Getenv("TEST_POSTGRES_PASSWORD") == "" {
		_ = os.Setenv("TEST_POSTGRES_PASSWORD", os.Getenv("POSTGRES_PASSWORD"))
	}
	if os.Getenv("TEST_POSTGRES_SSLMODE") == "" {
		sslMode := os.Getenv("POSTGRES_SSLMODE")
		if sslMode == "" {
			sslMode = "disable"
		}
		_ = os.Setenv("TEST_POSTGRES_SSLMODE", sslMode)
	}

	if os.Getenv("SOLANA_RPC_ENDPOINT") == "" {
		_ = os.Setenv("SOLANA_RPC_ENDPOINT", "http://127.0.0.1:8899")
	}
	if os.Getenv("SOLANA_PROGRAM_ID") == "" {
		programKey, err := solanaGo.NewRandomPrivateKey()
		if err == nil {
			_ = os.Setenv("SOLANA_PROGRAM_ID", programKey.PublicKey().String())
		}
	}
	if os.Getenv("SOLANA_ORACLE_PUBKEY") == "" || os.Getenv("SOLANA_ORACLE_PRIVATE_KEY") == "" {
		oracleKey, err := solanaGo.NewRandomPrivateKey()
		if err == nil {
			_ = os.Setenv("SOLANA_ORACLE_PUBKEY", oracleKey.PublicKey().String())
			_ = os.Setenv("SOLANA_ORACLE_PRIVATE_KEY", oracleKey.String())
		}
	}

	// connect to test database
	if _, err := postgres.ConnectTestDB(); err != nil {
		panic(err)
	}

	if err := migrations.MigrateDatabase(postgres.Db); err != nil {
		panic(err)
	}
}

func Teardown() {
	if err := migrations.MigrateDown(postgres.Db, 0); err != nil {
		log.Printf("Warning: failed to run migration down during teardown: %v", err)
	}
}

func MakeRequest(method, url string, body interface{}, token string) (*http.Request, *httptest.ResponseRecorder) {
	var reader io.Reader
	if body != nil {
		requestBody, _ := json.Marshal(body)
		reader = bytes.NewBuffer(requestBody)
	}

	requestURL := url
	if len(url) > 0 && url[0] == '/' {
		requestURL = "http://localhost:8080" + url
	}

	request, _ := http.NewRequest(method, requestURL, reader)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	writer := httptest.NewRecorder()
	Router().ServeHTTP(writer, request)
	return request, writer
}

// ValidateOpenAPIContract validates a request/response pair against openapi.yaml.
func ValidateOpenAPIContract(t interface {
	Helper()
	Fatalf(format string, args ...interface{})
}, request *http.Request, response *httptest.ResponseRecorder) {
	t.Helper()

	router, err := getOpenAPIRouter()
	if err != nil {
		t.Fatalf("failed to load OpenAPI router: %v", err)
	}

	route, pathParams, err := router.FindRoute(request)
	if err != nil {
		t.Fatalf("failed to match request %s %s against OpenAPI spec: %v", request.Method, request.URL.Path, err)
	}

	// read the request body into memory so it can be validated
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    request,
		PathParams: pathParams,
		Route:      route,
		Options: &openapi3filter.Options{
			AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
		},
	}
	if request.GetBody != nil {
		if body, bodyErr := request.GetBody(); bodyErr == nil {
			request.Body = body
		}
	}
	if err := openapi3filter.ValidateRequest(context.Background(), requestValidationInput); err != nil {
		t.Fatalf("OpenAPI request validation failed for %s %s: %v", request.Method, request.URL.Path, err)
	}

	// read the response body into memory for validation
	responseValidationInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: requestValidationInput,
		Status:                 response.Code,
		Header:                 response.Header().Clone(),
	}
	responseValidationInput.SetBodyBytes(response.Body.Bytes())

	if err := openapi3filter.ValidateResponse(context.Background(), responseValidationInput); err != nil {
		t.Fatalf("OpenAPI response validation failed for %s %s (status %d): %v", request.Method, request.URL.Path, response.Code, err)
	}
}

func getOpenAPIRouter() (routers.Router, error) {
	// singleton pattern to avoid reloading and reparsing the OpenAPI document for every test
	openAPIRouterOnce.Do(func() {
		document, err := loadOpenAPIDocument()
		if err != nil {
			openAPIRouterErr = err
			return
		}

		router, err := legacy.NewRouter(document)
		if err != nil {
			openAPIRouterErr = err
			return
		}

		openAPIRouter = router
	})

	return openAPIRouter, openAPIRouterErr
}

func loadOpenAPIDocument() (*openapi3.T, error) {
	openAPIDocumentOnce.Do(func() {
		loader := openapi3.NewLoader()
		document, err := loader.LoadFromFile(openAPISpecPath)
		if err != nil {
			openAPIDocumentErr = err
			return
		}

		if err := document.Validate(context.Background()); err != nil {
			openAPIDocumentErr = err
			return
		}

		openAPIDocument = document
	})

	return openAPIDocument, openAPIDocumentErr
}

func BearerToken() string {
	user := gin.H{
		"email":    "test@test.com",
		"password": "test",
	}

	_, writer := MakeRequest("POST", "/api/v1/login", user, "")
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
