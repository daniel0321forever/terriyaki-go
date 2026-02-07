package postgres

import (
	"context"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"gorm.io/gorm"
)

// GrindSchema is a private struct used only for GORM mapping (decoupling for Grind entity)
type GrindSchema struct {
	ID           string          `json:"id" gorm:"primaryKey"`
	Duration     int32           `json:"duration" gorm:"not null"` // stored in days
	Participants []entities.User `json:"participants" gorm:"many2many:participate_records;foreignKey:ID;references:ID"`
	Budget       int32           `json:"budget" gorm:"not null"`
	Tasks        []entities.Task `json:"tasks" gorm:"-"` // Excluded from GORM - tasks are managed separately via TaskRepository
	StartDate    time.Time       `json:"start_date" gorm:"not null"`
	CreatedAt    time.Time       `json:"created_at" gorm:"not null"`
	UpdatedAt    time.Time       `json:"updated_at" gorm:"not null"`
}

// TableName tells GORM which table to use
func (GrindSchema) TableName() string {
	return "grinds"
}

type GormGrindRepository struct {
	db *gorm.DB
}

// NewGormGrindRepository creates the repo.
// We "inject" the database connection here.
func NewGormGrindRepository(db *gorm.DB) *GormGrindRepository {
	return &GormGrindRepository{db: db}
}

func (r *GormGrindRepository) Create(grind *entities.Grind) error {
	ctx := context.Background()
	// 1. Map Entity -> DB Schema
	model := GrindSchema{
		ID:        grind.ID,
		Duration:  grind.Duration,
		Budget:    grind.Budget,
		StartDate: grind.StartDate,
		CreatedAt: grind.CreatedAt,
		UpdatedAt: grind.UpdatedAt,
	}

	// 2. Save to Postgres
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *GormGrindRepository) FindById(id string) (*entities.Grind, error) {
	ctx := context.Background()
	var model GrindSchema
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return nil, err
	}

	// 3. Map DB Model back to Entity
	return &entities.Grind{
		ID:        model.ID,
		Duration:  model.Duration,
		Budget:    model.Budget,
		StartDate: model.StartDate,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}, nil
}

func (r *GormGrindRepository) Update(grind *entities.Grind) error {
	ctx := context.Background()
	// 1. Map Entity -> DB Model
	// We update the UpdatedAt timestamp to the current time
	model := GrindSchema{
		ID:        grind.ID,
		Duration:  grind.Duration,
		Budget:    grind.Budget,
		StartDate: grind.StartDate,
		UpdatedAt: time.Now().UTC(),
	}

	// 2. Save updates to Postgres
	// .Model(&model) tells GORM which record to find based on the ID
	// .Updates(model) only updates the fields that are non-zero
	result := r.db.WithContext(ctx).Model(&model).Updates(model)
	if result.Error != nil {
		return result.Error
	}

	// Update the original entity's timestamp before returning
	grind.UpdatedAt = model.UpdatedAt

	return nil
}

func (r *GormGrindRepository) Delete(id string) error {
	ctx := context.Background()
	// In GORM, if you pass a struct with a Primary Key to Delete,
	// it uses that key to find the record.
	result := r.db.WithContext(ctx).Delete(&GrindSchema{}, "id = ?", id)

	if result.Error != nil {
		return result.Error
	}

	// If no rows were affected, it means the ID didn't exist
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *GormGrindRepository) FindAllByUserID(userID string) ([]*entities.Grind, error) {
	ctx := context.Background()
	var models []GrindSchema

	// Join with participation table to find grinds for a user
	err := r.db.WithContext(ctx).
		Table("grinds").
		Joins("INNER JOIN participation ON grinds.id = participation.grind_id").
		Where("participation.user_id = ?", userID).
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	grinds := make([]*entities.Grind, len(models))
	for i, model := range models {
		grinds[i] = &entities.Grind{
			ID:        model.ID,
			Duration:  model.Duration,
			Budget:    model.Budget,
			StartDate: model.StartDate,
			CreatedAt: model.CreatedAt,
			UpdatedAt: model.UpdatedAt,
		}
	}
	return grinds, nil
}

func (r *GormGrindRepository) FindLatestByUserID(userID string) (*entities.Grind, error) {
	ctx := context.Background()
	var model GrindSchema

	// Join with participation table and order by created_at desc, get first
	err := r.db.WithContext(ctx).
		Table("grinds").
		Joins("INNER JOIN participation ON grinds.id = participation.grind_id").
		Where("participation.user_id = ?", userID).
		Order("grinds.created_at DESC").
		First(&model).Error

	if err != nil {
		return nil, err
	}

	return &entities.Grind{
		ID:        model.ID,
		Duration:  model.Duration,
		Budget:    model.Budget,
		StartDate: model.StartDate,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}, nil
}

func (r *GormGrindRepository) DeleteAll() error {
	ctx := context.Background()
	return r.db.WithContext(ctx).Where("1 = 1").Delete(&GrindSchema{}).Error
}
