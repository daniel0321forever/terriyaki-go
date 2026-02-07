package mappers

import (
	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

// Converts Grind entity to GroupGrindDTO
func GrindToGroupGrindDTO(grind *entities.Grind) *dto.GroupGrindDTO {

	return &dto.GroupGrindDTO{
		ID:        grind.ID,
		Duration:  grind.Duration,
		Budget:    grind.Budget,
		StartDate: grind.StartDate,
		CreatedAt: grind.CreatedAt,
		UpdatedAt: grind.UpdatedAt,
	}
}
