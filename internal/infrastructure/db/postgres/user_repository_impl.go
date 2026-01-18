package postgres

import (
	"context"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"gorm.io/gorm"
)

type UserSchema struct {
	ID       string `gorm:"primaryKey"`
	Username string
	Email    string `gorm:"uniqueIndex;not null"`
	// Name deprecated?
	Password  string
	Avatar    string
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (UserSchema) TableName() string { return "users" }

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) Create(u *entities.User) error {
	ctx := context.Background()
	now := time.Now().UTC()
	model := UserSchema{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Password:  u.HashedPassword, // Already hashed by the service
		Avatar:    u.Avatar,
		CreatedAt: now,
		UpdatedAt: now,
	}
	err := r.db.WithContext(ctx).Create(&model).Error
	return err
}

func (r *GormUserRepository) FindByEmail(email string) (*entities.User, error) {
	ctx := context.Background()
	var model UserSchema
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error; err != nil {
		return nil, err
	}
	return &entities.User{
		ID:             model.ID,
		Email:          model.Email,
		Username:       model.Username,
		Avatar:         model.Avatar,
		HashedPassword: model.Password,
	}, nil
}

func (r *GormUserRepository) FindById(id string) (*entities.User, error) {
	ctx := context.Background()
	var model UserSchema
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &entities.User{
		ID:             model.ID,
		Email:          model.Email,
		Username:       model.Username,
		Avatar:         model.Avatar,
		HashedPassword: model.Password,
	}, nil
}

func (r *GormUserRepository) Delete(id string) error {
	ctx := context.Background()
	return r.db.WithContext(ctx).Delete(&UserSchema{}, "id = ?", id).Error
}
