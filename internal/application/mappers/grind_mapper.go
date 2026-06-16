package mappers

import (
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

// BuildGroupGrindDTO constructs GroupGrindDTO from Grind-related entities.
func BuildGroupGrindDTO(grind *entities.Grind, participants []entities.User) *dto.GroupGrindDTO {
	tasks := grind.Tasks
	progressDTOs := make([]dto.HabitTaskProgressDTO, 0, len(tasks))
	for i := range tasks {
		p := BuildHabitTaskProgressDTO(&tasks[i])
		if p != nil {
			progressDTOs = append(progressDTOs, *p)
		}
	}

	var todayTaskDTO *dto.HabitTaskDTO
	today := time.Now().UTC().Truncate(24 * time.Hour)
	for i := range tasks {
		if tasks[i].Date.UTC().Truncate(24 * time.Hour).Equal(today) {
			todayTaskDTO = BuildHabitTaskDTO(&tasks[i])
			break
		}
	}

	participantDTOs := make([]dto.UserDTO, 0, len(participants))
	for _, p := range participants {
		participantDTOs = append(participantDTOs, *BuildUserDTO(&p))
	}

	return &dto.GroupGrindDTO{
		ID:           grind.ID,
		Duration:     grind.Duration,
		Budget:       grind.Budget,
		StartDate:    grind.StartDate,
		CreatedAt:    grind.CreatedAt,
		UpdatedAt:    grind.UpdatedAt,
		Progress:     progressDTOs,
		Participants: participantDTOs,
		TodayTask:    todayTaskDTO,
	}
}

// BuildMessageGrindDTO constructs MessageGrindDTO from Grind-related entity.
func BuildMessageGrindDTO(grind *entities.Grind) *dto.MessageGrindDTO {
	return &dto.MessageGrindDTO{
		Duration:     grind.Duration,
		StartDate:    grind.StartDate,
		Budget:       grind.Budget,
		Participants: []string{},
	}
}
