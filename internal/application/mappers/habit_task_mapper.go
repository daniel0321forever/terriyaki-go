package mappers

import (
	"encoding/json"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

// BuildHabitTaskDTO constructs a HabitTaskDTO from a HabitTask entity.
func BuildHabitTaskDTO(task *entities.HabitTask) *dto.HabitTaskDTO {
	var metadata interface{}
	if task.Metadata != nil {
		if err := json.Unmarshal(task.Metadata, &metadata); err != nil {
			metadata = nil
		}
	}

	return &dto.HabitTaskDTO{
		ID:           task.ID,
		TaskType:     task.TaskType,
		UserID:       task.UserID,
		GrindID:      task.GrindID,
		Date:         task.Date,
		FinishedTime: task.FinishedTime,
		Completed:    task.Completed,
		Metadata:     metadata,
	}
}

// BuildHabitTaskProgressDTO builds a compact progress DTO from a HabitTask entity.
func BuildHabitTaskProgressDTO(task *entities.HabitTask) *dto.HabitTaskProgressDTO {
	status := "pending"
	if task.Completed {
		status = "completed"
	} else {
		startOfTaskDay := task.Date.UTC().Truncate(24 * time.Hour).Add(time.Hour)
		startOfToday := time.Now().UTC().Truncate(24 * time.Hour)
		if startOfTaskDay.Before(startOfToday) {
			status = "missed"
		}
	}
	return &dto.HabitTaskProgressDTO{
		ID:           task.ID,
		Date:         task.Date,
		FinishedTime: task.FinishedTime,
		Status:       status,
	}
}
