package repositories

import (
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

type ParticipationRepository interface {
	Create(*entities.Participation) error
	FindByParticipationId(ParticipationID string) (*entities.Participation, error)
	FindByUserAndGrind(userID string, grindID string) (*entities.Participation, error)
	Update(participation *entities.Participation) error
	DeleteByGrindID(grindID string) error
}
