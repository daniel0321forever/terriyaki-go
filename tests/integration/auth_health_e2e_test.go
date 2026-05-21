package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	testharness "github.com/daniel0321forever/terriyaki-go/tests"
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

	_, writer := testharness.MakeRequest(http.MethodPost, "/api/v1/register", requestBody, "")
	assert.Equal(t, http.StatusOK, writer.Code)

	return map[string]string{
		"email":    email,
		"password": password,
	}
}

func TestV1HealthPingE2E(t *testing.T) {
	request, writer := testharness.MakeRequest(http.MethodGet, "/api/v1/ping", nil, "")
	assert.Equal(t, http.StatusOK, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	assert.Equal(t, "pong", resp["message"])

	testharness.ValidateOpenAPIContract(t, request, writer)
}

func TestV1RegisterE2E(t *testing.T) {
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

func TestV1VerifyTokenUnauthorizedE2E(t *testing.T) {
	request, writer := testharness.MakeRequest(http.MethodGet, "/api/v1/verify-token", nil, "invalid-token")
	assert.Equal(t, http.StatusUnauthorized, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	assert.Equal(t, "UNAUTHORIZED", resp["errorCode"])

	testharness.ValidateOpenAPIContract(t, request, writer)
}

func TestV1LoginInvalidEmailE2E(t *testing.T) {
	requestBody := map[string]string{
		"email":    fmt.Sprintf("missing-%d@example.com", time.Now().UTC().UnixNano()),
		"password": "does-not-matter",
	}

	request, writer := testharness.MakeRequest(http.MethodPost, "/api/v1/login", requestBody, "")
	assert.Equal(t, http.StatusBadRequest, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	assert.Equal(t, "BAD_REQUEST", resp["errorCode"])

	testharness.ValidateOpenAPIContract(t, request, writer)
}

func TestV2LoginSuccessE2E(t *testing.T) {
	credentials := registerUserForV2Auth(t)

	requestBody := map[string]string{
		"email":    credentials["email"],
		"password": credentials["password"],
	}

	request, writer := testharness.MakeRequest(http.MethodPost, "/api/v2/login", requestBody, "")
	assert.Equal(t, http.StatusOK, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	assert.NotEmpty(t, resp["token"])

	user, ok := resp["user"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, credentials["email"], user["email"])

	grinds, ok := resp["grinds"].([]interface{})
	assert.True(t, ok)
	assert.Empty(t, grinds)

	testharness.ValidateOpenAPIContract(t, request, writer)
}

func TestV2LoginInvalidEmailE2E(t *testing.T) {
	requestBody := map[string]string{
		"email":    fmt.Sprintf("missing-v2-%d@example.com", time.Now().UTC().UnixNano()),
		"password": "does-not-matter",
	}

	request, writer := testharness.MakeRequest(http.MethodPost, "/api/v2/login", requestBody, "")
	assert.Equal(t, http.StatusBadRequest, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	assert.Equal(t, "BAD_REQUEST", resp["errorCode"])

	testharness.ValidateOpenAPIContract(t, request, writer)
}

func TestV2VerifyTokenUnauthorizedE2E(t *testing.T) {
	request, writer := testharness.MakeRequest(http.MethodGet, "/api/v2/verify-token", nil, "invalid-token")
	assert.Equal(t, http.StatusUnauthorized, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	assert.Equal(t, "UNAUTHORIZED", resp["errorCode"])

	testharness.ValidateOpenAPIContract(t, request, writer)
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

	request, writer := testharness.MakeRequest(http.MethodGet, "/api/v2/verify-token", nil, token)
	assert.Equal(t, http.StatusOK, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	_, ok = resp["user"].(map[string]interface{})
	assert.True(t, ok)

	grinds, ok := resp["grinds"].([]interface{})
	assert.True(t, ok)
	assert.Empty(t, grinds)

	testharness.ValidateOpenAPIContract(t, request, writer)
}
