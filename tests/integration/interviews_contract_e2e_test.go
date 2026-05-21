package integration

import (
	"net/http"
	"testing"

	testharness "github.com/daniel0321forever/terriyaki-go/tests"
	"github.com/stretchr/testify/assert"
)

func TestStartInterviewUnauthorized(t *testing.T) {
	requestBody := map[string]string{"taskId": "task_test"}
	request, writer := testharness.MakeRequest(http.MethodPost, "/api/v1/interviews/start", requestBody, "invalid-token")
	assert.Equal(t, http.StatusUnauthorized, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	assert.Equal(t, "UNAUTHORIZED", resp["errorCode"])

	testharness.ValidateOpenAPIContract(t, request, writer)
}
