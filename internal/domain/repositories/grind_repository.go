package repositories

import (
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

// The interface: should be implemented by infrastructure/db/*
type GrindRepository interface {
	Create(grind *entities.Grind) error
	FindById(id string) (*entities.Grind, error)
	FindAllByUserID(userID string) ([]*entities.Grind, error)
	FindLatestByUserID(userID string) (*entities.Grind, error)
	Update(grind *entities.Grind) error
	Delete(id string) error
	DeleteAll() error
	FindDuedGrinds() ([]*entities.Grind, error)
}
