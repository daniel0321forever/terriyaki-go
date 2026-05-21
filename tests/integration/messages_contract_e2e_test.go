package integration

import (
	"net/http"
	"testing"

	testharness "github.com/daniel0321forever/terriyaki-go/tests"
	"github.com/stretchr/testify/assert"
)

func TestMessagesUnauthorized(t *testing.T) {
	request, writer := testharness.MakeRequest(http.MethodGet, "/api/v1/messages", nil, "invalid-token")
	assert.Equal(t, http.StatusUnauthorized, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	assert.Equal(t, "UNAUTHORIZED", resp["errorCode"])

	testharness.ValidateOpenAPIContract(t, request, writer)
}

func TestCreateInvitationUnauthorized(t *testing.T) {
	requestBody := map[string]string{
		"participantEmail": "invitee@example.com",
		"grindID":          "grind_test",
	}

	request, writer := testharness.MakeRequest(http.MethodPost, "/api/v1/messages/invitation", requestBody, "invalid-token")
	assert.Equal(t, http.StatusUnauthorized, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	assert.Equal(t, "UNAUTHORIZED", resp["errorCode"])

	testharness.ValidateOpenAPIContract(t, request, writer)
}
