package integration

import (
	"net/http"
	"testing"

	testharness "github.com/daniel0321forever/terriyaki-go/tests"
	"github.com/stretchr/testify/assert"
)

func TestGrindListUnauthorizedContractE2E(t *testing.T) {
	request, writer := testharness.MakeRequest(http.MethodGet, "/api/v1/grinds", nil, "invalid-token")
	assert.Equal(t, http.StatusUnauthorized, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	errorObj := resp["error"].(map[string]interface{})
	assert.Equal(t, "unauthorized", errorObj["message"])

	testharness.ValidateOpenAPIContract(t, request, writer)
}

func TestTaskTodayUnauthorizedContractE2E(t *testing.T) {
	request, writer := testharness.MakeRequest(http.MethodGet, "/api/v1/tasks/today", nil, "invalid-token")
	assert.Equal(t, http.StatusUnauthorized, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	errorObj := resp["error"].(map[string]interface{})
	assert.Equal(t, "unauthorized", errorObj["message"])

	testharness.ValidateOpenAPIContract(t, request, writer)
}

func TestSelectDefaultPaymentMethodUnauthorizedContractE2E(t *testing.T) {
	body := map[string]string{"payment_method_id": "pm_test"}
	request, writer := testharness.MakeRequest(http.MethodPost, "/api/v1/payments/methods/select-default", body, "invalid-token")
	assert.Equal(t, http.StatusUnauthorized, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	errorObj := resp["error"].(map[string]interface{})
	assert.Equal(t, "Unauthorized", errorObj["message"])

	testharness.ValidateOpenAPIContract(t, request, writer)
}

func TestMessagesUnauthorizedContractE2E(t *testing.T) {
	request, writer := testharness.MakeRequest(http.MethodGet, "/api/v1/messages", nil, "invalid-token")
	assert.Equal(t, http.StatusUnauthorized, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	errorObj := resp["error"].(map[string]interface{})
	assert.Equal(t, "unauthorized", errorObj["message"])

	testharness.ValidateOpenAPIContract(t, request, writer)
}
