package mocks

import (
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/stretchr/testify/mock"
)

type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) Create(task *entities.Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockTaskRepository) FindByID(id string) (*entities.Task, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*entities.Task), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTaskRepository) FindTodayTask(userID, grindID string) (*entities.Task, error) {
	args := m.Called(userID, grindID)
	if args.Get(0) != nil {
		return args.Get(0).(*entities.Task), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTaskRepository) FindByGrindIDAndParticipantID(grindID, userID string) ([]entities.Task, error) {
	args := m.Called(grindID, userID)
	if args.Get(0) != nil {
		return args.Get(0).([]entities.Task), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTaskRepository) Update(task *entities.Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockTaskRepository) DeleteByGrindID(grindID string) error {
	args := m.Called(grindID)
	return args.Error(0)
}

func (m *MockTaskRepository) GetCompletionStats(grindID string) (int, int, error) {
	args := m.Called(grindID)
	return args.Int(0), args.Int(1), args.Error(2)
}
