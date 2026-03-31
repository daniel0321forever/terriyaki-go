package mocks

import (
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/stretchr/testify/mock"
)

type MockGrindRepository struct {
	mock.Mock
}

func (m *MockGrindRepository) Create(grind *entities.Grind) error {
	args := m.Called(grind)
	return args.Error(0)
}

func (m *MockGrindRepository) FindById(id string) (*entities.Grind, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*entities.Grind), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGrindRepository) FindAllByUserID(userID string) ([]*entities.Grind, error) {
	args := m.Called(userID)
	if args.Get(0) != nil {
		return args.Get(0).([]*entities.Grind), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGrindRepository) FindLatestByUserID(userID string) (*entities.Grind, error) {
	args := m.Called(userID)
	if args.Get(0) != nil {
		return args.Get(0).(*entities.Grind), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGrindRepository) Update(grind *entities.Grind) error {
	args := m.Called(grind)
	return args.Error(0)
}

func (m *MockGrindRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockGrindRepository) DeleteAll() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockGrindRepository) FindDuedGrinds() ([]*entities.Grind, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]*entities.Grind), args.Error(1)
	}
	return nil, args.Error(1)
}
