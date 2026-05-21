package repositories

import (
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

type UserRepository interface {
	FindById(id string) (*entities.User, error)
	FindByEmail(email string) (*entities.User, error)
	FindByGrindID(grindID string) ([]entities.User, error)
	Create(user *entities.User) error
	Delete(id string) error
	Update(user *entities.User) error
}
