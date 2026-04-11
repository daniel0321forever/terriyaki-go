package mappers

import (
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

// BuildGroupGrindDTO constructs GroupGrindDTO from Grind-related entities
func BuildGroupGrindDTO(grind *entities.Grind, participants []entities.User) *dto.GroupGrindDTO {
	// Convert Tasks
	tasks := grind.Tasks
	taskProgressDTOs := make([]dto.TaskProgressDTO, 0, len(tasks))
	for i := range tasks {
			taskProgressDTO := BuildTaskProgressDTO(&tasks[i])
		if taskProgressDTO != nil {
			taskProgressDTOs = append(taskProgressDTOs, *taskProgressDTO)
		}
	}

	// Get today's task
	var todayTask *entities.Task = nil
	var todayTaskDTO *dto.TaskDTO = nil

	for _, task := range tasks {
		if task.Date.Equal(time.Now().UTC().Truncate(24 * time.Hour)) {
			todayTask = &task
			break
		}
	}

	if todayTask != nil {
		todayTaskDTO = BuildTaskDTO(todayTask)
	}

	participantsDTOs := make([]dto.UserDTO, 0, len(participants))
	for _, participant := range participants {
		participantsDTOs = append(participantsDTOs, *BuildUserDTO(&participant))
	}

	// get today's task

	return &dto.GroupGrindDTO{
		ID:           grind.ID,
		Duration:     grind.Duration,
		Budget:       grind.Budget,
		StartDate:    grind.StartDate,
		CreatedAt:    grind.CreatedAt,
		UpdatedAt:    grind.UpdatedAt,
		Progress:     taskProgressDTOs,
		Participants: participantsDTOs,
		TodayTask:    todayTaskDTO,
	}
}

// BuildMessageGrindDTO constructs MessageGrindDTO from Grind-related entity
func BuildMessageGrindDTO(grind *entities.Grind) *dto.MessageGrindDTO {
	return &dto.MessageGrindDTO{
		Duration:     grind.Duration,
		StartDate:    grind.StartDate,
		Budget:       grind.Budget,
		Participants: []string{},
	}
}
