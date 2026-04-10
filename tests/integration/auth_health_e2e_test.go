package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	testharness "github.com/daniel0321forever/terriyaki-go/tests"
	"github.com/stretchr/testify/assert"
)

func TestHealthPingE2E(t *testing.T) {
	request, writer := testharness.MakeRequest(http.MethodGet, "/api/v1/ping", nil, "")
	assert.Equal(t, http.StatusOK, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	assert.Equal(t, "pong", resp["message"])

	testharness.ValidateOpenAPIContract(t, request, writer)
}

func TestRegisterE2E(t *testing.T) {
	email := fmt.Sprintf("e2e-user-%d@example.com", time.Now().UTC().UnixNano())
	requestBody := map[string]string{
		"username": "e2e-user",
		"email":    email,
		"password": "secure-password-123",
		"avatar":   "",
	}

	request, writer := testharness.MakeRequest(http.MethodPost, "/api/v1/register", requestBody, "")
	assert.Equal(t, http.StatusOK, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	assert.Equal(t, "Registration successful", resp["message"])
	assert.NotEmpty(t, resp["token"])

	user, ok := resp["user"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, email, user["email"])
	assert.Equal(t, "e2e-user", user["username"])

	testharness.ValidateOpenAPIContract(t, request, writer)
}

func TestVerifyTokenUnauthorizedE2E(t *testing.T) {
	request, writer := testharness.MakeRequest(http.MethodGet, "/api/v1/verify-token", nil, "invalid-token")
	assert.Equal(t, http.StatusUnauthorized, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	assert.Equal(t, "unauthorized", resp["message"])

	testharness.ValidateOpenAPIContract(t, request, writer)
}

func TestLoginInvalidEmailE2E(t *testing.T) {
	requestBody := map[string]string{
		"email":    fmt.Sprintf("missing-%d@example.com", time.Now().UTC().UnixNano()),
		"password": "does-not-matter",
	}

	request, writer := testharness.MakeRequest(http.MethodPost, "/api/v1/login", requestBody, "")
	assert.Equal(t, http.StatusBadRequest, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	assert.Equal(t, "invalid email", resp["message"])

	testharness.ValidateOpenAPIContract(t, request, writer)
}
