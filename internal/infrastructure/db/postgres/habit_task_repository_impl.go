package postgres

import (
	"context"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type HabitTaskSchema struct {
	gorm.Model
	ID           string         `json:"id" gorm:"primaryKey"`
	TaskType     string         `json:"task_type" gorm:"not null"`
	UserID       string         `json:"user_id" gorm:"not null"`
	GrindID      string         `json:"grind_id" gorm:"not null"`
	Date         time.Time      `json:"date" gorm:"not null"`
	FinishedTime *time.Time     `json:"finished_time"`
	Completed    bool           `json:"completed" gorm:"default:false"`
	Metadata     datatypes.JSON `json:"metadata"`
}

func (HabitTaskSchema) TableName() string { return "habit_tasks" }

type GormHabitTaskRepository struct {
	db *gorm.DB
}

func NewGormHabitTaskRepository(db *gorm.DB) *GormHabitTaskRepository {
	return &GormHabitTaskRepository{db: db}
}

func (r *GormHabitTaskRepository) WithTx(tx *gorm.DB) *GormHabitTaskRepository {
	return &GormHabitTaskRepository{db: tx}
}

func habitTaskSchemaToEntity(s *HabitTaskSchema) *entities.HabitTask {
	return &entities.HabitTask{
		ID:           s.ID,
		TaskType:     s.TaskType,
		UserID:       s.UserID,
		GrindID:      s.GrindID,
		Date:         s.Date,
		FinishedTime: s.FinishedTime,
		Completed:    s.Completed,
		Metadata:     s.Metadata,
	}
}

func (r *GormHabitTaskRepository) Create(task *entities.HabitTask) error {
	ctx := context.Background()
	model := HabitTaskSchema{
		ID:           task.ID,
		TaskType:     task.TaskType,
		UserID:       task.UserID,
		GrindID:      task.GrindID,
		Date:         task.Date,
		FinishedTime: task.FinishedTime,
		Completed:    task.Completed,
		Metadata:     task.Metadata,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *GormHabitTaskRepository) FindByID(id string) (*entities.HabitTask, error) {
	ctx := context.Background()
	var model HabitTaskSchema
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return habitTaskSchemaToEntity(&model), nil
}

func (r *GormHabitTaskRepository) FindByGrindIDAndUserID(grindID, userID string) ([]*entities.HabitTask, error) {
	ctx := context.Background()
	var models []HabitTaskSchema
	err := r.db.WithContext(ctx).Where("grind_id = ? AND user_id = ?", grindID, userID).Find(&models).Error
	if err != nil {
		return nil, err
	}
	tasks := make([]*entities.HabitTask, len(models))
	for i := range models {
		tasks[i] = habitTaskSchemaToEntity(&models[i])
	}
	return tasks, nil
}

func (r *GormHabitTaskRepository) FindTodayTask(userID, grindID string) (*entities.HabitTask, error) {
	ctx := context.Background()
	var model HabitTaskSchema
	today := time.Now().UTC().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND grind_id = ? AND date >= ? AND date < ?", userID, grindID, today, tomorrow).
		First(&model).Error
	if err != nil {
		return nil, err
	}
	return habitTaskSchemaToEntity(&model), nil
}

func (r *GormHabitTaskRepository) Update(task *entities.HabitTask) error {
	ctx := context.Background()
	model := HabitTaskSchema{
		ID:           task.ID,
		TaskType:     task.TaskType,
		UserID:       task.UserID,
		GrindID:      task.GrindID,
		Date:         task.Date,
		FinishedTime: task.FinishedTime,
		Completed:    task.Completed,
		Metadata:     task.Metadata,
	}
	return r.db.WithContext(ctx).Model(&HabitTaskSchema{}).Where("id = ?", task.ID).Updates(&model).Error
}

func (r *GormHabitTaskRepository) FindByGrindIDAndParticipantID(grindID, participantID string) ([]entities.HabitTask, error) {
	ptrs, err := r.FindByGrindIDAndUserID(grindID, participantID)
	if err != nil {
		return nil, err
	}
	tasks := make([]entities.HabitTask, len(ptrs))
	for i, p := range ptrs {
		tasks[i] = *p
	}
	return tasks, nil
}

func (r *GormHabitTaskRepository) DeleteByGrindID(grindID string) error {
	ctx := context.Background()
	return r.db.WithContext(ctx).Where("grind_id = ?", grindID).Delete(&HabitTaskSchema{}).Error
}
