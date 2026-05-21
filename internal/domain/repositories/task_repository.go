package repositories

import (
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

type TaskRepository interface {
	Create(*entities.Task) error
	FindByID(id string) (*entities.Task, error)
	FindTodayTask(userID, grindID string) (*entities.Task, error)
	FindByGrindIDAndParticipantID(grindID, userID string) ([]entities.Task, error)
	Update(task *entities.Task) error
	DeleteByGrindID(grindID string) error
	GetCompletionStats(grindID string) (completedCount, totalCount int, err error)
}
