package repositories

import (
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

// CompletionEventRepository defines persistence operations for CompletionEvent domain entities.
type CompletionEventRepository interface {
	Create(event *entities.CompletionEvent) error
	FindByHabitTaskID(habitTaskID string) ([]*entities.CompletionEvent, error)
	FindByUserIDAndProvider(userID string, provider entities.CompletionProvider) ([]*entities.CompletionEvent, error)
}
