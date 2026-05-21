package integration

import (
	"net/http"
	"testing"

	testharness "github.com/daniel0321forever/terriyaki-go/tests"
	"github.com/stretchr/testify/assert"
)

func TestSelectDefaultPaymentMethodUnauthorized(t *testing.T) {
	requestBody := map[string]string{"payment_method_id": "pm_test"}
	request, writer := testharness.MakeRequest(http.MethodPost, "/api/v1/payments/stripe/methods/select-default", requestBody, "invalid-token")
	assert.Equal(t, http.StatusUnauthorized, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	assert.Equal(t, "UNAUTHORIZED", resp["errorCode"])

	testharness.ValidateOpenAPIContract(t, request, writer)
}

func TestGetPaymentMethodsUnauthorized(t *testing.T) {
	request, writer := testharness.MakeRequest(http.MethodGet, "/api/v1/payments/stripe/methods", nil, "invalid-token")
	assert.Equal(t, http.StatusUnauthorized, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	assert.Equal(t, "UNAUTHORIZED", resp["errorCode"])

	testharness.ValidateOpenAPIContract(t, request, writer)
}
