package services

import (
	"errors"
	"testing"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/cores/config"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func Test_IngestService_Ingest_LeetCode_Success(t *testing.T) {
	t.Parallel()

	habitTaskRepo := new(mocks.MockHabitTaskRepository)
	completionEventRepo := new(mocks.MockCompletionEventRepository)

	todayTask := &entities.HabitTask{
		ID:      "task-1",
		UserID:  "user-1",
		GrindID: "grind-1",
		Date:    time.Now(),
	}
	habitTaskRepo.On("FindTodayTask", "user-1", "grind-1").Return(todayTask, nil)
	completionEventRepo.On("Create", mock.MatchedBy(func(e *entities.CompletionEvent) bool {
		return e.HabitTaskID == "task-1" &&
			e.UserID == "user-1" &&
			e.Provider == entities.ProviderLeetCode
	})).Return(nil)

	svc := NewIngestService(habitTaskRepo, completionEventRepo)

	rawPayload := map[string]interface{}{
		"grindID":           "grind-1",
		"problemTitle":      "Two Sum",
		"problemURL":        "https://leetcode.com/problems/two-sum",
		"problemDifficulty": "Easy",
		"occurredAt":        time.Now().Format(time.RFC3339),
	}

	event, err := svc.Ingest("leetcode", "user-1", "grind-1", rawPayload)
	assert.NoError(t, err)
	assert.NotNil(t, event)
	assert.Equal(t, entities.ProviderLeetCode, event.Provider)
	assert.Equal(t, "task-1", event.HabitTaskID)
	assert.Equal(t, "user-1", event.UserID)

	habitTaskRepo.AssertExpectations(t)
	completionEventRepo.AssertExpectations(t)
}

func Test_IngestService_Ingest_Duolingo_Success(t *testing.T) {
	t.Parallel()

	habitTaskRepo := new(mocks.MockHabitTaskRepository)
	completionEventRepo := new(mocks.MockCompletionEventRepository)

	todayTask := &entities.HabitTask{
		ID:      "task-2",
		UserID:  "user-1",
		GrindID: "grind-1",
		Date:    time.Now(),
	}
	habitTaskRepo.On("FindTodayTask", "user-1", "grind-1").Return(todayTask, nil)
	completionEventRepo.On("Create", mock.MatchedBy(func(e *entities.CompletionEvent) bool {
		return e.HabitTaskID == "task-2" &&
			e.UserID == "user-1" &&
			e.Provider == entities.ProviderDuolingo
	})).Return(nil)

	svc := NewIngestService(habitTaskRepo, completionEventRepo)

	rawPayload := map[string]interface{}{
		"grindID":          "grind-1",
		"streakCount":      30,
		"lessonsCompleted": 5,
		"xpEarned":         120,
		"occurredAt":       time.Now().Format(time.RFC3339),
	}

	event, err := svc.Ingest("duolingo", "user-1", "grind-1", rawPayload)
	assert.NoError(t, err)
	assert.NotNil(t, event)
	assert.Equal(t, entities.ProviderDuolingo, event.Provider)
	assert.Equal(t, "task-2", event.HabitTaskID)

	habitTaskRepo.AssertExpectations(t)
	completionEventRepo.AssertExpectations(t)
}

func Test_IngestService_Ingest_UnsupportedProvider(t *testing.T) {
	t.Parallel()

	habitTaskRepo := new(mocks.MockHabitTaskRepository)
	completionEventRepo := new(mocks.MockCompletionEventRepository)

	svc := NewIngestService(habitTaskRepo, completionEventRepo)

	_, err := svc.Ingest("unknown", "user-1", "grind-1", map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported provider: unknown")

	// No repo calls expected
	habitTaskRepo.AssertExpectations(t)
	completionEventRepo.AssertExpectations(t)
}

func Test_IngestService_Ingest_HabitTaskNotFound(t *testing.T) {
	t.Parallel()

	habitTaskRepo := new(mocks.MockHabitTaskRepository)
	completionEventRepo := new(mocks.MockCompletionEventRepository)

	habitTaskRepo.On("FindTodayTask", "user-1", "grind-1").Return(nil, gorm.ErrRecordNotFound)

	svc := NewIngestService(habitTaskRepo, completionEventRepo)

	_, err := svc.Ingest("leetcode", "user-1", "grind-1", map[string]interface{}{})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, config.ErrHabitTaskNotFound))

	habitTaskRepo.AssertExpectations(t)
	completionEventRepo.AssertExpectations(t)
}
