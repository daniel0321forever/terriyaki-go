package postgres

import (
	"context"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"gorm.io/gorm"
)

type UserSchema struct {
	gorm.Model
	ID                     string    `json:"id" gorm:"primaryKey"`
	Username               string    `json:"username" gorm:"not null"`
	Email                  string    `json:"email" gorm:"uniqueIndex;not null"`
	Password               string    `json:"password" gorm:"not null"`
	Avatar                 string    `json:"avatar" gorm:""`
	StripeCustomerID       string    `json:"stripe_customer_id" gorm:""`
	DefaultPaymentMethodID string    `json:"default_payment_method_id" gorm:""`
	CreatedAt              time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt              time.Time `json:"updated_at" gorm:"autoUpdateTime"`
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

func (r *GormUserRepository) FindByGrindID(grindID string) ([]entities.User, error) {
	ctx := context.Background()
	var models []UserSchema
	err := r.db.WithContext(ctx).Table("users").
		Joins("INNER JOIN participation ON users.id = participation.user_id").
		Where("participation.grind_id = ?", grindID).
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	users := make([]entities.User, len(models))
	for i, model := range models {
		users[i] = entities.User{
			ID:                     model.ID,
			Email:                  model.Email,
			Username:               model.Username,
			Avatar:                 model.Avatar,
			HashedPassword:         model.Password,
			StripeCustomerID:       model.StripeCustomerID,
			DefaultPaymentMethodID: model.DefaultPaymentMethodID,
		}
	}

	return users, nil
}

func (r *GormUserRepository) Delete(id string) error {
	ctx := context.Background()
	return r.db.WithContext(ctx).Delete(&UserSchema{}, "id = ?", id).Error
}

func (r *GormUserRepository) Update(user *entities.User) error {
	ctx := context.Background()
	model := UserSchema{
		ID:       user.ID,
		Username: user.Username,
		Avatar:   user.Avatar,
	}
	return r.db.WithContext(ctx).Model(&UserSchema{}).Where("id = ?", user.ID).Updates(&model).Error
}
