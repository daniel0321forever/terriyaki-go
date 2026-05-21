package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/infrastructure/db/postgres"
	testharness "github.com/daniel0321forever/terriyaki-go/tests"
	"github.com/stretchr/testify/assert"
)

func registerUserAndSeedTaskData(t *testing.T) (string, string) {
	t.Helper()

	email := fmt.Sprintf("tasks-e2e-user-%d@example.com", time.Now().UTC().UnixNano())
	requestBody := map[string]string{
		"username": "tasks-e2e-user",
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

	grindRepo := postgres.NewGormGrindRepository(postgres.Db)
	participationRepo := postgres.NewGormParticipationRepository(postgres.Db)
	taskRepo := postgres.NewGormTaskRepository(postgres.Db)

	startDate := time.Now().UTC().Truncate(24 * time.Hour)
	grind, err := entities.NewGrind(7, 100, startDate)
	assert.NoError(t, err)
	assert.NoError(t, grindRepo.Create(grind))

	participation, err := entities.NewParticipation(userID, grind.ID)
	assert.NoError(t, err)
	assert.NoError(t, participationRepo.Create(participation))

	task, err := entities.NewTask(userID, grind.ID, startDate)
	assert.NoError(t, err)
	assert.NoError(t, taskRepo.Create(task))

	return token, task.ID
}

func TestTaskTodayUnauthorized(t *testing.T) {
	request, writer := testharness.MakeRequest(http.MethodGet, "/api/v1/tasks/today", nil, "invalid-token")
	assert.Equal(t, http.StatusUnauthorized, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	assert.Equal(t, "UNAUTHORIZED", resp["errorCode"])

	testharness.ValidateOpenAPIContract(t, request, writer)
}

func TestTaskTodaySuccess(t *testing.T) {
	token, _ := registerUserAndSeedTaskData(t)

	request, writer := testharness.MakeRequest(http.MethodGet, "/api/v1/tasks/today", nil, token)
	assert.Equal(t, http.StatusOK, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	task, ok := resp["task"].(map[string]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, task["id"])

	testharness.ValidateOpenAPIContract(t, request, writer)
}

func TestTaskByIDSuccess(t *testing.T) {
	token, taskID := registerUserAndSeedTaskData(t)

	request, writer := testharness.MakeRequest(http.MethodGet, "/api/v1/tasks/"+taskID, nil, token)
	assert.Equal(t, http.StatusOK, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	task, ok := resp["task"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, taskID, task["id"])

	testharness.ValidateOpenAPIContract(t, request, writer)
}

func TestTaskByIDNotFound(t *testing.T) {
	token, _ := registerUserAndSeedTaskData(t)

	missingTaskID := "missing-task-id"
	request, writer := testharness.MakeRequest(http.MethodGet, "/api/v1/tasks/"+missingTaskID, nil, token)
	assert.Equal(t, http.StatusNotFound, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	assert.Equal(t, "NOT_FOUND", resp["errorCode"])

	testharness.ValidateOpenAPIContract(t, request, writer)
}

func TestTaskFinishUnauthorized(t *testing.T) {
	requestBody := map[string]string{
		"code":     "print('done')",
		"language": "python",
	}

	request, writer := testharness.MakeRequest(http.MethodPost, "/api/v1/tasks/finish", requestBody, "invalid-token")
	assert.Equal(t, http.StatusUnauthorized, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	assert.Equal(t, "UNAUTHORIZED", resp["errorCode"])

	testharness.ValidateOpenAPIContract(t, request, writer)
}

func TestTaskFinishSuccess(t *testing.T) {
	token, _ := registerUserAndSeedTaskData(t)

	requestBody := map[string]string{
		"code":     "print('done')",
		"language": "python",
	}

	request, writer := testharness.MakeRequest(http.MethodPost, "/api/v1/tasks/finish", requestBody, token)
	assert.Equal(t, http.StatusOK, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	assert.NotEmpty(t, resp["task"])

	testharness.ValidateOpenAPIContract(t, request, writer)
}

func TestTaskFinishInvalidBody(t *testing.T) {
	token, _ := registerUserAndSeedTaskData(t)

	requestBody := map[string]any{
		"code":     123,
		"language": "python",
	}

	_, writer := testharness.MakeRequest(http.MethodPost, "/api/v1/tasks/finish", requestBody, token)
	assert.Equal(t, http.StatusBadRequest, writer.Code)

	resp := testharness.ParseResponseMap(t, writer.Body.Bytes())
	assert.Equal(t, "BAD_REQUEST", resp["errorCode"])
}
