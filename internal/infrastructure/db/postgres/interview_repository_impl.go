package postgres

import (
	"context"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type InterviewSessionSchema struct {
	gorm.Model
	ID                 string         `json:"id" gorm:"primaryKey"`
	UserID             string         `json:"user_id" gorm:"not null"`
	TaskID             string         `json:"task_id" gorm:"not null"`
	Status             string         `json:"status" gorm:"not null"`
	ConversationHistory datatypes.JSON `json:"conversation_history" gorm:"type:jsonb"`
	StartedAt          time.Time      `json:"started_at" gorm:"not null"`
	EndedAt            *time.Time     `json:"ended_at" gorm:""`
}

func (InterviewSessionSchema) TableName() string {
	return "interview_sessions"
}

type GormInterviewSessionRepository struct {
	db *gorm.DB
}

func NewGormInterviewSessionRepository(db *gorm.DB) *GormInterviewSessionRepository {
	return &GormInterviewSessionRepository{db: db}
}

func (r *GormInterviewSessionRepository) Create(session *entities.InterviewSession) error {
	ctx := context.Background()
	model := InterviewSessionSchema{
		ID:                 session.ID,
		UserID:             session.UserID,
		TaskID:             session.TaskID,
		Status:             session.Status,
		ConversationHistory: session.ConversationHistory,
		StartedAt:          session.StartedAt,
		EndedAt:            session.EndedAt,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *GormInterviewSessionRepository) FindByID(id string) (*entities.InterviewSession, error) {
	ctx := context.Background()
	var model InterviewSessionSchema
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &entities.InterviewSession{
		ID:                 model.ID,
		UserID:             model.UserID,
		TaskID:             model.TaskID,
		Status:             model.Status,
		ConversationHistory: model.ConversationHistory,
		StartedAt:          model.StartedAt,
		EndedAt:            model.EndedAt,
	}, nil
}

func (r *GormInterviewSessionRepository) Update(session *entities.InterviewSession) error {
	ctx := context.Background()
	model := InterviewSessionSchema{
		ID:                 session.ID,
		UserID:             session.UserID,
		TaskID:             session.TaskID,
		Status:             session.Status,
		ConversationHistory: session.ConversationHistory,
		StartedAt:          session.StartedAt,
		EndedAt:            session.EndedAt,
	}
	return r.db.WithContext(ctx).Model(&InterviewSessionSchema{}).Where("id = ?", session.ID).Updates(&model).Error
}

