package mocks

import (
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/stretchr/testify/mock"
)

// MockCompletionEventRepository is a testify mock implementation of repositories.CompletionEventRepository.
type MockCompletionEventRepository struct {
	mock.Mock
}

func (m *MockCompletionEventRepository) Create(event *entities.CompletionEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockCompletionEventRepository) FindByHabitTaskID(habitTaskID string) ([]*entities.CompletionEvent, error) {
	args := m.Called(habitTaskID)
	if args.Get(0) != nil {
		return args.Get(0).([]*entities.CompletionEvent), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCompletionEventRepository) FindByUserIDAndProvider(userID string, provider entities.CompletionProvider) ([]*entities.CompletionEvent, error) {
	args := m.Called(userID, provider)
	if args.Get(0) != nil {
		return args.Get(0).([]*entities.CompletionEvent), args.Error(1)
	}
	return nil, args.Error(1)
}
