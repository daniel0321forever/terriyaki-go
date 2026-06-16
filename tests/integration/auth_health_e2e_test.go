//go:build integration
// +build integration

package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/interface/api"
	testharness "github.com/daniel0321forever/terriyaki-go/tests"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func registerUserForV2Auth(t *testing.T) map[string]string {
	t.Helper()

	email := fmt.Sprintf("e2e-v2-user-%d@example.com", time.Now().UTC().UnixNano())
	password := "secure-password-123"
	requestBody := map[string]string{
		"username": "e2e-v2-user",
		"email":    email,
		"password": password,
		"avatar":   "",
	}

	_, writer := testharness.MakeRequest(http.MethodPost, "/api/v2/register", requestBody, "")
	assert.Equal(t, http.StatusOK, writer.Code)

	return map[string]string{
		"email":    email,
		"password": password,
	}
}

func TestV2RegisterE2E(t *testing.T) {
	email := fmt.Sprintf("e2e-user-%d@example.com", time.Now().UTC().UnixNano())
	requestBody := map[string]string{
		"username": "e2e-user",
		"email":    email,
		"password": "secure-password-123",
		"avatar":   "",
	}

	_, writer := testharness.MakeRequest(http.MethodPost, "/api/v2/register", requestBody, "")
	assert.Equal(t, http.StatusOK, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	assert.Equal(t, "Registration successful", resp["message"])
	assert.NotEmpty(t, resp["token"])

	user, ok := resp["user"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, email, user["email"])
	assert.Equal(t, "e2e-user", user["username"])
}

func TestV2VerifyTokenUnauthorizedE2E(t *testing.T) {
	_, writer := testharness.MakeRequest(http.MethodGet, "/api/v2/verify-token", nil, "invalid-token")
	assert.Equal(t, http.StatusUnauthorized, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	assert.Equal(t, "UNAUTHORIZED", resp["errorCode"])
}

func TestV2LoginInvalidEmailE2E(t *testing.T) {
	requestBody := map[string]string{
		"email":    fmt.Sprintf("missing-v2-%d@example.com", time.Now().UTC().UnixNano()),
		"password": "does-not-matter",
	}

	_, writer := testharness.MakeRequest(http.MethodPost, "/api/v2/login", requestBody, "")
	assert.Equal(t, http.StatusBadRequest, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	assert.Equal(t, "BAD_REQUEST", resp["errorCode"])
}

func TestV2LoginSuccessE2E(t *testing.T) {
	credentials := registerUserForV2Auth(t)

	requestBody := map[string]string{
		"email":    credentials["email"],
		"password": credentials["password"],
	}

	_, writer := testharness.MakeRequest(http.MethodPost, "/api/v2/login", requestBody, "")
	assert.Equal(t, http.StatusOK, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	assert.NotEmpty(t, resp["token"])

	user, ok := resp["user"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, credentials["email"], user["email"])

	grinds, ok := resp["grinds"].([]interface{})
	assert.True(t, ok)
	assert.Empty(t, grinds)
}

func TestV2VerifyTokenSuccessE2E(t *testing.T) {
	credentials := registerUserForV2Auth(t)

	loginBody := map[string]string{
		"email":    credentials["email"],
		"password": credentials["password"],
	}
	_, loginWriter := testharness.MakeRequest(http.MethodPost, "/api/v2/login", loginBody, "")
	assert.Equal(t, http.StatusOK, loginWriter.Code)
	loginResp := testharness.ParseResponseMap(t, loginWriter.Body.Bytes())
	token, ok := loginResp["token"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, token)

	_, writer := testharness.MakeRequest(http.MethodGet, "/api/v2/verify-token", nil, token)
	assert.Equal(t, http.StatusOK, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	_, ok = resp["user"].(map[string]interface{})
	assert.True(t, ok)

	grinds, ok := resp["grinds"].([]interface{})
	assert.True(t, ok)
	assert.Empty(t, grinds)
}

// TestHealthCheckWithRealDB verifies that GET /api/health returns 200 with status=="ok"
// when PostgreSQL is reachable (Redis may be absent — skip redis assertion if not configured).
func TestHealthCheckWithRealDB(t *testing.T) {
	_, writer := testharness.MakeRequest(http.MethodGet, "/api/health", nil, "")
	assert.Equal(t, http.StatusOK, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	assert.Equal(t, "ok", resp["status"])
	assert.Equal(t, "ok", resp["postgres"])
}

// TestHealthCheckRedisFail constructs a HealthController with a bad Redis address
// and verifies it returns 503 with redis=="error" (without panicking).
func TestHealthCheckRedisFail(t *testing.T) {
	db := testharness.TestDB()

	// Point at a non-existent Redis server to force a ping failure
	badRdb := redis.NewClient(&redis.Options{
		Addr: "localhost:19999",
	})
	defer badRdb.Close()

	ctrl := api.NewHealthController(db, badRdb)

	// Call via httptest directly
	w := httptest.NewRecorder()
	c, _ := testharness.NewGinContext(w)
	ctrl.HealthAPI(c)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	resp := testharness.ParseResponseMap(t, w.Body.Bytes())
	assert.Equal(t, "degraded", resp["status"])
	assert.Equal(t, "error", resp["redis"])
}
