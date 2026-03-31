package services

import (
	"encoding/json"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/application/mappers"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/config"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/utils"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/repositories"
	"gorm.io/datatypes"
)

type TaskService struct {
	taskRepo repositories.TaskRepository
}

func NewTaskService(taskRepo repositories.TaskRepository) *TaskService {
	return &TaskService{
		taskRepo: taskRepo,
	}
}

func (s *TaskService) GetTaskByID(request dto.GetTaskDTO) (*dto.TaskDTO, error) {
	task, err := s.taskRepo.FindByID(request.TaskID)
	if err != nil {
		return nil, err
	}

	if request.SetProblemIfNeeded && !task.HasProblemAssigned() {
		// Get random problem from neetcode250
		problem, err := utils.GetRandomProblemFromList("neetcode250")
		if err != nil {
			return nil, err
		}

		// Update task with problem details
		problemTitle := problem.Title
		problemURL := "https://leetcode.com/problems/" + problem.Slug
		problemDifficulty := problem.Difficulty
		problemDescription := problem.Description

		task.ProblemTitle = &problemTitle
		task.ProblemURL = &problemURL
		task.ProblemDifficulty = &problemDifficulty
		task.ProblemDescription = &problemDescription

		// Convert topic tags to JSON
		if len(problem.TopicTags) > 0 {
			tagsJSON, _ := json.Marshal(problem.TopicTags)
			task.ProblemTopicTags = datatypes.JSON(tagsJSON)
		}

		// Save updated task
		err = s.taskRepo.Update(task)
		if err != nil {
			return nil, err
		}
	}

	return mappers.TaskToTaskDTO(task), nil
}

func (s *TaskService) GetTodayTask(request dto.GetTodayTaskDTO) (*dto.TaskDTO, error) {
	task, err := s.taskRepo.FindTodayTask(request.UserID, request.GrindID)
	if err != nil {
		return nil, config.ErrTaskNotFound
	}
	return mappers.TaskToTaskDTO(task), nil
}

func (s *TaskService) GetTaskProgressList(request dto.GetTaskProgressListDTO) ([]*dto.TaskProgressDTO, error) {
	// check if the user is a participant of the grind
	taskRecords, err := s.taskRepo.FindByGrindIDAndParticipantID(request.GrindID, request.ParticipationID)
	if err != nil {
		return nil, err
	}

	var progressRecordsDTO []*dto.TaskProgressDTO
	for _, task := range taskRecords {
		progressRecordsDTO = append(progressRecordsDTO, mappers.TaskToTaskProgressDTO(&task))
	}

	return progressRecordsDTO, nil
}

// more like a "update task"
func (s *TaskService) FinishTask(request dto.FinishTaskDTO) error {
	task, err := s.taskRepo.FindByID(request.TaskID)
	if err != nil {
		return err
	}

	task.Completed = true
	task.FinishedTime = time.Now()
	task.Code = request.Code
	task.CodeLanguage = request.CodeLanguage

	err = s.taskRepo.Update(task)
	if err != nil {
		return err
	}

	return nil
}

// FIXME: it this deprecated?
func (s *TaskService) GetCompletionStats(grindID string) (completedCount, totalCount int, err error) {
	return s.taskRepo.GetCompletionStats(grindID)
}
