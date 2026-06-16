package mocks

import (
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/stretchr/testify/mock"
)

// MockHabitTaskRepository is a testify mock implementation of repositories.HabitTaskRepository.
type MockHabitTaskRepository struct {
	mock.Mock
}

func (m *MockHabitTaskRepository) Create(task *entities.HabitTask) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockHabitTaskRepository) FindByID(id string) (*entities.HabitTask, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*entities.HabitTask), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockHabitTaskRepository) FindByGrindIDAndUserID(grindID, userID string) ([]*entities.HabitTask, error) {
	args := m.Called(grindID, userID)
	if args.Get(0) != nil {
		return args.Get(0).([]*entities.HabitTask), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockHabitTaskRepository) FindTodayTask(userID, grindID string) (*entities.HabitTask, error) {
	args := m.Called(userID, grindID)
	if args.Get(0) != nil {
		return args.Get(0).(*entities.HabitTask), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockHabitTaskRepository) Update(task *entities.HabitTask) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockHabitTaskRepository) FindByGrindIDAndParticipantID(grindID, participantID string) ([]entities.HabitTask, error) {
	args := m.Called(grindID, participantID)
	if args.Get(0) != nil {
		return args.Get(0).([]entities.HabitTask), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockHabitTaskRepository) DeleteByGrindID(grindID string) error {
	args := m.Called(grindID)
	return args.Error(0)
}
