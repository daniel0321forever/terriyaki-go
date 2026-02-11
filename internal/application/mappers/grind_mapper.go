package mappers

import (
	"fmt"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/container"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

// Converts Grind entity to GroupGrindDTO
func GrindToGroupGrindDTO(grind *entities.Grind) *dto.GroupGrindDTO {
	// Convert Tasks
	tasks := grind.Tasks
	taskProgressDTOs := make([]dto.TaskProgressDTO, 0, len(tasks))
	for i := range tasks {
		taskProgressDTO := TaskToTaskProgressDTO(&tasks[i])
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
		todayTaskDTO = TaskToTaskDTO(todayTask)
	}

	// Convert participants (users)
	participants, err := container.Repos.UserRepository.FindByGrindID(grind.ID)
	if err != nil {
		fmt.Println("Error finding participants by grind ID:", err)
		panic(err)
	}
	participantsDTOs := make([]dto.UserDTO, 0, len(participants))
	for _, participant := range participants {
		participantsDTOs = append(participantsDTOs, *UserToUserDTO(&participant))
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

func GrindToMessageGrindDTO(grind *entities.Grind) *dto.MessageGrindDTO {
	return &dto.MessageGrindDTO{
		Duration:     grind.Duration,
		StartDate:    grind.StartDate,
		Budget:       grind.Budget,
		Participants: []string{},
	}
}
