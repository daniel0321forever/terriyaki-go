package mappers

import (
	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

// BuildParticipationDTO constructs Participation DTO from Participation-related entity
func BuildParticipationDTO(participation *entities.Participation) *dto.ParticipationDTO {

	return &dto.ParticipationDTO{
		ID:           participation.UserID,
		UserID:       participation.UserID,
		GrindID:      participation.GrindID,
		MissedDays:   participation.MissedDays,
		TotalPenalty: participation.TotalPenalty,
		Quitted:      participation.Quitted,
		QuittedAt:    &participation.QuittedAt,
	}
}
