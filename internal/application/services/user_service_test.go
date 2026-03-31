package services

import (
	"errors"
	"testing"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/config"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/mocks"
	"github.com/stretchr/testify/mock"
)

func TestUserServiceCreateUser_DuplicateEmail(t *testing.T) {
	t.Parallel()

	po := new(mocks.MockUserRepository)
	po.On("FindByEmail", "alice@example.com").Return(&entities.User{ID: "u1", Email: "alice@example.com"}, nil)

	svc := NewUserService(po)

	_, err := svc.CreateUser(dto.CreateUserDTO{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "password123",
	})
	if !errors.Is(err, config.ErrUserAlreadyExists) {
		t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestUserServiceCreateUser_Success(t *testing.T) {
	t.Parallel()

	po := new(mocks.MockUserRepository)
	po.On("FindByEmail", "ALICE@example.com").Return(nil, nil)
	po.On("Create", mock.MatchedBy(func(u *entities.User) bool {
		return u.Username == "alice" && u.Email == "alice@example.com"
	})).Return(nil)

	svc := NewUserService(po)

	res, err := svc.CreateUser(dto.CreateUserDTO{
		Username: "alice",
		Email:    "ALICE@example.com",
		Password: "password123",
		Avatar:   "https://example.com/a.png",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res == nil || res.ID == "" {
		t.Fatalf("expected created user dto with id")
	}
	po.AssertExpectations(t)
}

func TestUserServiceGetUser_NotFound(t *testing.T) {
	t.Parallel()

	po := new(mocks.MockUserRepository)
	po.On("FindById", "missing").Return(nil, errors.New("db error"))

	svc := NewUserService(po)

	_, err := svc.GetUser(dto.GetUserDTO{UserID: "missing"})
	if !errors.Is(err, config.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserServiceUpdateUser_UpdatesSelectedFields(t *testing.T) {
	t.Parallel()

	stored := &entities.User{ID: "u1", Username: "old", Avatar: "old.png", DefaultPaymentMethodID: "pm_old"}
	po := new(mocks.MockUserRepository)
	po.On("FindById", "u1").Return(stored, nil)
	po.On("Update", mock.MatchedBy(func(u *entities.User) bool {
		return u.ID == "u1" && u.Username == "new"
	})).Return(nil)

	svc := NewUserService(po)

	newName := "new"
	newAvatar := "new.png"
	newPM := "pm_new"
	res, err := svc.UpdateUser(dto.UpdateUserDTO{
		UserID:                 "u1",
		Username:               &newName,
		Avatar:                 &newAvatar,
		DefaultPaymentMethodID: &newPM,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.Username != newName || res.Avatar != newAvatar {
		t.Fatalf("expected updated fields in result")
	}
	if stored.DefaultPaymentMethodID != newPM {
		t.Fatalf("expected default payment method to be updated")
	}
}
