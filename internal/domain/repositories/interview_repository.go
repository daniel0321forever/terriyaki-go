package repositories

import (
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

type InterviewSessionRepository interface {
	Create(session *entities.InterviewSession) error
	FindByID(id string) (*entities.InterviewSession, error)
	Update(session *entities.InterviewSession) error
}

