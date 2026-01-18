package mappers

import (
	"encoding/json"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

// TaskToTaskDTO converts Task entity to TaskDTO
func TaskToTaskDTO(task *entities.Task) *dto.TaskDTO {
	var topicTags interface{}
	if task.ProblemTopicTags != nil {
		json.Unmarshal(task.ProblemTopicTags, &topicTags)
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
