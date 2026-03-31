package services

import (
	"errors"
	"testing"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/config"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/mocks"
	"github.com/stretchr/testify/mock"
)

func TestTaskServiceGetTodayTask_NotFoundMapped(t *testing.T) {
	t.Parallel()

	po := new(mocks.MockTaskRepository)
	po.On("FindTodayTask", "u1", "g1").Return(nil, errors.New("not found"))

	svc := NewTaskService(po)

	_, err := svc.GetTodayTask(dto.GetTodayTaskDTO{UserID: "u1", GrindID: "g1"})
	if !errors.Is(err, config.ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestTaskServiceGetTaskByID_NoSetProblem(t *testing.T) {
	t.Parallel()

	task := &entities.Task{ID: "t1", UserID: "u1", GrindID: "g1", TaskType: "leetcode", Date: time.Now()}
	po := new(mocks.MockTaskRepository)
	po.On("FindByID", "t1").Return(task, nil)

	svc := NewTaskService(po)

	res, err := svc.GetTaskByID(dto.GetTaskDTO{TaskID: "t1", SetProblemIfNeeded: false})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res == nil || res.ID != "t1" {
		t.Fatalf("expected task dto id t1")
	}
}

func TestTaskServiceFinishTask_UpdatesTask(t *testing.T) {
	t.Parallel()

	task := &entities.Task{ID: "t1", Completed: false}
	po := new(mocks.MockTaskRepository)
	po.On("FindByID", "t1").Return(task, nil)
	po.On("Update", mock.MatchedBy(func(t *entities.Task) bool {
		return t.ID == "t1" && t.Completed && t.Code == "print(1)" && t.CodeLanguage == "python"
	})).Return(nil)

	svc := NewTaskService(po)

	err := svc.FinishTask(dto.FinishTaskDTO{TaskID: "t1", Code: "print(1)", CodeLanguage: "python"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !task.Completed {
		t.Fatalf("expected task to be marked completed")
	}
	if task.Code != "print(1)" || task.CodeLanguage != "python" {
		t.Fatalf("expected code fields to be updated")
	}
	if task.FinishedTime.IsZero() {
		t.Fatalf("expected finished time to be set")
	}
	po.AssertExpectations(t)
}
