package mocks

import (
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/stretchr/testify/mock"
)

type MockParticipationRepository struct {
	mock.Mock
}

func (m *MockParticipationRepository) Create(participation *entities.Participation) error {
	args := m.Called(participation)
	return args.Error(0)
}

func (m *MockParticipationRepository) FindByParticipationId(participationID string) (*entities.Participation, error) {
	args := m.Called(participationID)
	if args.Get(0) != nil {
		return args.Get(0).(*entities.Participation), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockParticipationRepository) FindByUserAndGrind(userID, grindID string) (*entities.Participation, error) {
	args := m.Called(userID, grindID)
	if args.Get(0) != nil {
		return args.Get(0).(*entities.Participation), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockParticipationRepository) Update(participation *entities.Participation) error {
	args := m.Called(participation)
	return args.Error(0)
}

func (m *MockParticipationRepository) DeleteByGrindID(grindID string) error {
	args := m.Called(grindID)
	return args.Error(0)
}
