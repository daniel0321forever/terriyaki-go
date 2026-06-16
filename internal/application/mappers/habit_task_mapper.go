package mappers

import (
	"encoding/json"

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
