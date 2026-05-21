package mocks

import (
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/stretchr/testify/mock"
)

type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) Create(message *entities.Message) error {
	args := m.Called(message)
	return args.Error(0)
}

func (m *MockMessageRepository) FindByID(id string) (*entities.Message, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*entities.Message), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockMessageRepository) FindAllForReceiver(receiverID string, offset, limit int) ([]*entities.Message, error) {
	args := m.Called(receiverID, offset, limit)
	if args.Get(0) != nil {
		return args.Get(0).([]*entities.Message), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockMessageRepository) FindAllFromSender(senderID string, offset, limit int) ([]*entities.Message, error) {
	args := m.Called(senderID, offset, limit)
	if args.Get(0) != nil {
		return args.Get(0).([]*entities.Message), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockMessageRepository) Update(message *entities.Message) error {
	args := m.Called(message)
	return args.Error(0)
}
