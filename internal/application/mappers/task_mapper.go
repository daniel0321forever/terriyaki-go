package mappers

import (
	"encoding/json"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

// BuildTaskDTO constructs Task DTO from Task-related entity
func BuildTaskDTO(task *entities.Task) *dto.TaskDTO {
	var topicTags interface{}
	if task.ProblemTopicTags != nil {
		if err := json.Unmarshal(task.ProblemTopicTags, &topicTags); err != nil {
			// If unmarshalling fails, leave topicTags as nil
			topicTags = nil
		}
	}

	return &dto.TaskDTO{
		ID:                 task.ID,
		TaskType:           task.TaskType,
		UserID:             task.UserID,
		GrindID:            task.GrindID,
		Date:               task.Date,
		FinishedTime:       task.FinishedTime,
		Completed:          task.Completed,
		ProblemTitle:       task.ProblemTitle,
		ProblemDescription: task.ProblemDescription,
		ProblemURL:         task.ProblemURL,
		ProblemDifficulty:  task.ProblemDifficulty,
		ProblemTopicTags:   topicTags,
		Code:               task.Code,
		CodeLanguage:       task.CodeLanguage,
	}
}

// BuildTaskProgressDTO constructs TaskProgressDTO from Task-related entity
func BuildTaskProgressDTO(task *entities.Task) *dto.TaskProgressDTO {
	status := "pending"
	if task.Completed {
		status = "completed"
	} else {
		startOfTaskDay := task.Date.UTC().Truncate(24 * time.Hour).Add(time.Hour * 1)
		startOfToday := time.Now().UTC().Truncate(24 * time.Hour)
		if startOfTaskDay.Before(startOfToday) {
			status = "missed"
		}
	}

	return &dto.TaskProgressDTO{
		ID:           task.ID,
		Date:         task.Date,
		FinishedTime: task.FinishedTime,
		Status:       status,
	}
}
