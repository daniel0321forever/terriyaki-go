package postgres

import (
	"context"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type TaskSchema struct {
	gorm.Model
	ID                 string    `json:"id" gorm:"primaryKey"`
	TaskType           string    `json:"task_type" gorm:"not null"`
	UserID             string    `json:"user_id" gorm:"not null"`
	GrindID            string    `json:"grind_id" gorm:"not null"`
	Date               time.Time `json:"date" gorm:"not null"`
	FinishedTime       time.Time `json:"finished_time"`
	Completed          bool      `json:"completed" gorm:"default:false"`
	ProblemTitle       *string   `json:"problem_title"`
	ProblemDescription *string   `json:"problem_description"`
	ProblemURL         *string   `json:"problem_url"`
	ProblemDifficulty  *string   `json:"problem_difficulty"`
	ProblemTopicTags   datatypes.JSON

	Code         string
	CodeLanguage string
}

func (TaskSchema) TableName() string { return "habit_tasks" }

type GormTaskRepository struct {
	db *gorm.DB
}

func NewGormTaskRepository(db *gorm.DB) *GormTaskRepository {
	return &GormTaskRepository{db: db}
}

func (r *GormTaskRepository) Create(task *entities.Task) error {
	ctx := context.Background()
	model := TaskSchema{
		ID:                 task.ID,
		TaskType:           task.TaskType,
		UserID:             task.UserID,
		GrindID:            task.GrindID,
		Date:               task.Date,
		FinishedTime:       task.FinishedTime,
		Completed:          task.Completed,
		ProblemTitle:       task.ProblemTitle,
		ProblemDescription: task.ProblemDescription,
		ProblemURL:         task.ProblemURL,
		ProblemDifficulty:  task.ProblemDifficulty,
		ProblemTopicTags:   task.ProblemTopicTags,
		Code:               task.Code,
		CodeLanguage:       task.CodeLanguage,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *GormTaskRepository) DeleteByGrindID(grindID string) error {
	ctx := context.Background()
	return r.db.WithContext(ctx).Where("grind_id = ?", grindID).Delete(&TaskSchema{}).Error
}

func (r *GormTaskRepository) FindByID(id string) (*entities.Task, error) {
	ctx := context.Background()
	var model TaskSchema
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &entities.Task{
		ID:                 model.ID,
		TaskType:           model.TaskType,
		UserID:             model.UserID,
		GrindID:            model.GrindID,
		Date:               model.Date,
		FinishedTime:       model.FinishedTime,
		Completed:          model.Completed,
		ProblemTitle:       model.ProblemTitle,
		ProblemDescription: model.ProblemDescription,
		ProblemURL:         model.ProblemURL,
		ProblemDifficulty:  model.ProblemDifficulty,
		ProblemTopicTags:   model.ProblemTopicTags,
		Code:               model.Code,
		CodeLanguage:       model.CodeLanguage,
	}, nil
}

func (r *GormTaskRepository) FindTodayTask(userID, grindID string) (*entities.Task, error) {
	ctx := context.Background()
	var model TaskSchema
	today := time.Now().UTC().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND grind_id = ? AND date >= ? AND date < ?", userID, grindID, today, tomorrow).
		First(&model).Error

	if err != nil {
		return nil, err
	}

	return &entities.Task{
		ID:                 model.ID,
		TaskType:           model.TaskType,
		UserID:             model.UserID,
		GrindID:            model.GrindID,
		Date:               model.Date,
		FinishedTime:       model.FinishedTime,
		Completed:          model.Completed,
		ProblemTitle:       model.ProblemTitle,
		ProblemDescription: model.ProblemDescription,
		ProblemURL:         model.ProblemURL,
		ProblemDifficulty:  model.ProblemDifficulty,
		ProblemTopicTags:   model.ProblemTopicTags,
		Code:               model.Code,
		CodeLanguage:       model.CodeLanguage,
	}, nil
}

func (r *GormTaskRepository) FindByGrindIDAndParticipantID(grindID, userID string) ([]entities.Task, error) {
	ctx := context.Background()
	var models []TaskSchema
	err := r.db.WithContext(ctx).Where("grind_id = ? AND user_id = ?", grindID, userID).Find(&models).Error
	if err != nil {
		return nil, err
	}

	tasks := make([]entities.Task, len(models))
	for i, model := range models {
		tasks[i] = entities.Task{
			ID:                 model.ID,
			TaskType:           model.TaskType,
			UserID:             model.UserID,
			GrindID:            model.GrindID,
			Date:               model.Date,
			FinishedTime:       model.FinishedTime,
			Completed:          model.Completed,
			ProblemTitle:       model.ProblemTitle,
			ProblemDescription: model.ProblemDescription,
			ProblemURL:         model.ProblemURL,
			ProblemDifficulty:  model.ProblemDifficulty,
			ProblemTopicTags:   model.ProblemTopicTags,
			Code:               model.Code,
			CodeLanguage:       model.CodeLanguage,
		}
	}
	return tasks, nil
}

func (r *GormTaskRepository) Update(task *entities.Task) error {
	ctx := context.Background()
	model := TaskSchema{
		ID:                 task.ID,
		TaskType:           task.TaskType,
		UserID:             task.UserID,
		GrindID:            task.GrindID,
		Date:               task.Date,
		FinishedTime:       task.FinishedTime,
		Completed:          task.Completed,
		ProblemTitle:       task.ProblemTitle,
		ProblemDescription: task.ProblemDescription,
		ProblemURL:         task.ProblemURL,
		ProblemDifficulty:  task.ProblemDifficulty,
		ProblemTopicTags:   task.ProblemTopicTags,
		Code:               task.Code,
		CodeLanguage:       task.CodeLanguage,
	}
	return r.db.WithContext(ctx).Model(&TaskSchema{}).Where("id = ?", task.ID).Updates(&model).Error
}

func (r *GormTaskRepository) GetCompletionStats(grindID string) (completedCount, totalCount int, err error) {
	ctx := context.Background()
	var count int64
	err = r.db.WithContext(ctx).Model(&TaskSchema{}).Where("grind_id = ? AND completed = ?", grindID, true).Count(&count).Error
	if err != nil {
		return 0, 0, err
	}
	completedCount = int(count)

	err = r.db.WithContext(ctx).Model(&TaskSchema{}).Where("grind_id = ?", grindID).Count(&count).Error
	if err != nil {
		return 0, 0, err
	}
	totalCount = int(count)
	return completedCount, totalCount, nil
}
