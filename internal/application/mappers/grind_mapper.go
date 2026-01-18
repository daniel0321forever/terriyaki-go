package mappers

import (
	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

// Converts Grind entity to GroupGrindDTO
func GrindToGroupGrindDTO(grind *entities.Grind) *dto.GroupGrindDTO {
	// Convert Tasks
	taskDTOs := make([]dto.TaskDTO, 0, len(grind.Tasks))
	for i := range grind.Tasks {
		taskDTO := TaskToTaskDTO(&grind.Tasks[i])
		if taskDTO != nil {
			taskDTOs = append(taskDTOs, *taskDTO)
		}
	}
	// Convert participants (users)
	participatnsDTOs := make([]dto.UserDTO, 0, len(grind.Participants))
	for i := range grind.Participants {
		participatnsDTO := UserToUserDTO(&grind.Participants[i])
		if participatnsDTO != nil {
			participatnsDTOs = append(participatnsDTOs, *participatnsDTO)
		}
	}

	return &dto.GroupGrindDTO{
		ID:           grind.ID,
		Duration:     grind.Duration,
		Budget:       grind.Budget,
		StartDate:    grind.StartDate,
		CreatedAt:    grind.CreatedAt,
		UpdatedAt:    grind.UpdatedAt,
		Tasks:        taskDTOs,
		Participants: participatnsDTOs,
	}
}
