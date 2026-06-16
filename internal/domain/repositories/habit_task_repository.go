package repositories

import (
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

// HabitTaskRepository defines persistence operations for HabitTask domain entities.
type HabitTaskRepository interface {
	Create(task *entities.HabitTask) error
	FindByID(id string) (*entities.HabitTask, error)
	FindByGrindIDAndUserID(grindID, userID string) ([]*entities.HabitTask, error)
	FindTodayTask(userID, grindID string) (*entities.HabitTask, error)
	Update(task *entities.HabitTask) error
	DeleteByGrindID(grindID string) error
}
