package postgres

import (
	"context"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CompletionEventSchema struct {
	gorm.Model
	ID          string         `json:"id" gorm:"primaryKey"`
	HabitTaskID string         `json:"habit_task_id" gorm:"not null"`
	UserID      string         `json:"user_id" gorm:"not null"`
	Provider    string         `json:"provider" gorm:"not null"`
	OccurredAt  time.Time      `json:"occurred_at" gorm:"not null"`
	Metadata    datatypes.JSON `json:"metadata"`
}

func (CompletionEventSchema) TableName() string { return "completion_events" }

type GormCompletionEventRepository struct {
	db *gorm.DB
}

func NewGormCompletionEventRepository(db *gorm.DB) *GormCompletionEventRepository {
	return &GormCompletionEventRepository{db: db}
}

func completionEventSchemaToEntity(s *CompletionEventSchema) *entities.CompletionEvent {
	return &entities.CompletionEvent{
		ID:          s.ID,
		HabitTaskID: s.HabitTaskID,
		UserID:      s.UserID,
		Provider:    entities.CompletionProvider(s.Provider),
		OccurredAt:  s.OccurredAt,
		Metadata:    s.Metadata,
	}
}

func (r *GormCompletionEventRepository) Create(event *entities.CompletionEvent) error {
	ctx := context.Background()
	model := CompletionEventSchema{
		ID:          event.ID,
		HabitTaskID: event.HabitTaskID,
		UserID:      event.UserID,
		Provider:    string(event.Provider),
		OccurredAt:  event.OccurredAt,
		Metadata:    event.Metadata,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *GormCompletionEventRepository) FindByHabitTaskID(habitTaskID string) ([]*entities.CompletionEvent, error) {
	ctx := context.Background()
	var models []CompletionEventSchema
	err := r.db.WithContext(ctx).Where("habit_task_id = ?", habitTaskID).Find(&models).Error
	if err != nil {
		return nil, err
	}
	events := make([]*entities.CompletionEvent, len(models))
	for i := range models {
		events[i] = completionEventSchemaToEntity(&models[i])
	}
	return events, nil
}

func (r *GormCompletionEventRepository) FindByUserIDAndProvider(userID string, provider entities.CompletionProvider) ([]*entities.CompletionEvent, error) {
	ctx := context.Background()
	var models []CompletionEventSchema
	err := r.db.WithContext(ctx).Where("user_id = ? AND provider = ?", userID, string(provider)).Find(&models).Error
	if err != nil {
		return nil, err
	}
	events := make([]*entities.CompletionEvent, len(models))
	for i := range models {
		events[i] = completionEventSchemaToEntity(&models[i])
	}
	return events, nil
}
