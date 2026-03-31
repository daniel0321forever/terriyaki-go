package mocks

import (
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindById(id string) (*entities.User, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*entities.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) FindByEmail(email string) (*entities.User, error) {
	args := m.Called(email)
	if args.Get(0) != nil {
		return args.Get(0).(*entities.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) FindByGrindID(grindID string) ([]entities.User, error) {
	args := m.Called(grindID)
	if args.Get(0) != nil {
		return args.Get(0).([]entities.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) Create(user *entities.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) Update(user *entities.User) error {
	args := m.Called(user)
	return args.Error(0)
}
