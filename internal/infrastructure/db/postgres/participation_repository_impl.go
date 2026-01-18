package postgres

import (
	"context"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"gorm.io/gorm"
)

type ParticipationSchema struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	UserID       string    `json:"user_id" gorm:"not null;constraint:OnDelete:CASCADE;"`
	GrindID      string    `json:"grind_id" gorm:"not null;constraint:OnDelete:CASCADE;"`
	MissedDays   int       `json:"missed_days" gorm:"not null;default:0"`
	TotalPenalty int       `json:"total_penalty" gorm:"not null;default:0"`
	Quitted      bool      `json:"quitted" gorm:"not null;default:false"`
	QuittedAt    time.Time `json:"quitted_at" gorm:""`
}

func (ParticipationSchema) TableName() string { return "participation" }

type GormParticipationRepository struct {
	db *gorm.DB
}

func NewGormParticipationRepository(db *gorm.DB) *GormParticipationRepository {
	return &GormParticipationRepository{db: db}
}

func (r *GormParticipationRepository) Create(participation *entities.Participation) error {
	ctx := context.Background()
	model := ParticipationSchema{
		ID:           participation.ID,
		UserID:       participation.UserID,
		GrindID:      participation.GrindID,
		MissedDays:   participation.MissedDays,
		TotalPenalty: participation.TotalPenalty,
		Quitted:      participation.Quitted,
		QuittedAt:    participation.QuittedAt,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *GormParticipationRepository) FindByParticipationId(ParticipationID string) (*entities.Participation, error) {
	ctx := context.Background()
	var model ParticipationSchema
	err := r.db.WithContext(ctx).Where("id = ?", ParticipationID).First(&model).Error
	if err != nil {
		return nil, err
	}

	return &entities.Participation{
		ID:           model.ID,
		UserID:       model.UserID,
		GrindID:      model.GrindID,
		MissedDays:   model.MissedDays,
		TotalPenalty: model.TotalPenalty,
		Quitted:      model.Quitted,
		QuittedAt:    model.QuittedAt,
	}, nil
}

func (r *GormParticipationRepository) FindByUserAndGrind(userID string, grindID string) (*entities.Participation, error) {
	ctx := context.Background()
	var model ParticipationSchema
	err := r.db.WithContext(ctx).Where("user_id = ? AND grind_id = ?", userID, grindID).First(&model).Error
	if err != nil {
		return nil, err
	}
	return &entities.Participation{
		ID:           model.ID,
		UserID:       model.UserID,
		GrindID:      model.GrindID,
		MissedDays:   model.MissedDays,
		TotalPenalty: model.TotalPenalty,
		Quitted:      model.Quitted,
		QuittedAt:    model.QuittedAt,
	}, nil
}

func (r *GormParticipationRepository) Update(participation *entities.Participation) error {
	ctx := context.Background()
	model := ParticipationSchema{
		ID:           participation.ID,
		UserID:       participation.UserID,
		GrindID:      participation.GrindID,
		MissedDays:   participation.MissedDays,
		TotalPenalty: participation.TotalPenalty,
		Quitted:      participation.Quitted,
		QuittedAt:    participation.QuittedAt,
	}
	return r.db.WithContext(ctx).
		Model(&ParticipationSchema{}).
		Where("user_id = ? AND grind_id = ?", participation.UserID, participation.GrindID).
		Updates(&model).Error
}

func (r *GormParticipationRepository) DeleteByGrindID(grindID string) error {
	ctx := context.Background()
	return r.db.WithContext(ctx).Where("grind_id = ?", grindID).Delete(&ParticipationSchema{}).Error
}
