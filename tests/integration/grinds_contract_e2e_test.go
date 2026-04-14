package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/infrastructure/db/postgres"
	testharness "github.com/daniel0321forever/terriyaki-go/tests"
	"github.com/stretchr/testify/assert"
)

func registerGrindTestUser(t *testing.T) (string, string) {
	t.Helper()

	email := fmt.Sprintf("grind-e2e-user-%d@example.com", time.Now().UTC().UnixNano())
	requestBody := map[string]string{
		"username": "grind-e2e-user",
		"email":    email,
		"password": "secure-password-123",
		"avatar":   "",
	}

	_, writer := testharness.MakeRequest(http.MethodPost, "/api/v1/register", requestBody, "")
	assert.Equal(t, http.StatusOK, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	token, ok := resp["token"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, token)

	user, ok := resp["user"].(map[string]interface{})
	assert.True(t, ok)
	userID, ok := user["id"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, userID)

	return token, userID
}

func seedGrindData(t *testing.T, userID string) string {
	t.Helper()

	grindRepo := postgres.NewGormGrindRepository(postgres.Db)
	participationRepo := postgres.NewGormParticipationRepository(postgres.Db)

	startDate := time.Now().UTC().Truncate(24 * time.Hour)
	grind, err := entities.NewGrind(7, 100, startDate)
	assert.NoError(t, err)
	assert.NoError(t, grindRepo.Create(grind))

	participation, err := entities.NewParticipation(userID, grind.ID)
	assert.NoError(t, err)
	assert.NoError(t, participationRepo.Create(participation))

	return grind.ID
}

func TestGrindCreateUnauthorized(t *testing.T) {
	requestBody := map[string]any{
		"duration":  7,
		"budget":    100,
		"startDate": time.Now().UTC().Format(time.RFC3339),
	}

	request, writer := testharness.MakeRequest(http.MethodPost, "/api/v1/grinds", requestBody, "invalid-token")
	assert.Equal(t, http.StatusUnauthorized, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	assert.Equal(t, "UNAUTHORIZED", resp["errorCode"])

	testharness.ValidateOpenAPIContract(t, request, writer)
}

func TestGrindCreateSuccess(t *testing.T) {
	t.Skip("CreateGrindAPI success path depends on CSV fixture path in service (assets/neetcode250.csv); cover list/current/by-id via seeded data")
}

func TestGrindListUnauthorized(t *testing.T) {
	request, writer := testharness.MakeRequest(http.MethodGet, "/api/v1/grinds", nil, "invalid-token")
	assert.Equal(t, http.StatusUnauthorized, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	assert.Equal(t, "UNAUTHORIZED", resp["errorCode"])

	testharness.ValidateOpenAPIContract(t, request, writer)
}

func TestGrindListSuccess(t *testing.T) {
	token, userID := registerGrindTestUser(t)
	createdGrindID := seedGrindData(t, userID)

	request, writer := testharness.MakeRequest(http.MethodGet, "/api/v1/grinds", nil, token)
	assert.Equal(t, http.StatusOK, writer.Code)

	var grinds []map[string]interface{}
	err := json.Unmarshal(writer.Body.Bytes(), &grinds)
	assert.NoError(t, err)

	assert.NotEmpty(t, grinds)

	found := false
	for _, grind := range grinds {
		if grind["id"] == createdGrindID {
			found = true
			break
		}
	}

	assert.True(t, found)

	testharness.ValidateOpenAPIContract(t, request, writer)
}

func TestGrindCurrentSuccess(t *testing.T) {
	token, userID := registerGrindTestUser(t)
	createdGrindID := seedGrindData(t, userID)

	request, writer := testharness.MakeRequest(http.MethodGet, "/api/v1/grinds/current", nil, token)
	assert.Equal(t, http.StatusOK, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	grind, ok := resp["grind"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, createdGrindID, grind["id"])

	testharness.ValidateOpenAPIContract(t, request, writer)
}

func TestGrindByIDSuccess(t *testing.T) {
	token, userID := registerGrindTestUser(t)
	createdGrindID := seedGrindData(t, userID)

	request, writer := testharness.MakeRequest(http.MethodGet, "/api/v1/grinds/"+createdGrindID, nil, token)
	assert.Equal(t, http.StatusOK, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	grind, ok := resp["grind"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, createdGrindID, grind["id"])

	testharness.ValidateOpenAPIContract(t, request, writer)
}
