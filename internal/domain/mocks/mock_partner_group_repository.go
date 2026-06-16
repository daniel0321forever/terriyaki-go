package mocks

import (
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/stretchr/testify/mock"
)

// MockPartnerGroupRepository is a testify mock implementation of repositories.PartnerGroupRepository.
type MockPartnerGroupRepository struct {
	mock.Mock
}

func (m *MockPartnerGroupRepository) Create(group *entities.PartnerGroup) error {
	args := m.Called(group)
	return args.Error(0)
}

func (m *MockPartnerGroupRepository) FindByID(id string) (*entities.PartnerGroup, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*entities.PartnerGroup), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPartnerGroupRepository) FindByGrindID(grindID string) (*entities.PartnerGroup, error) {
	args := m.Called(grindID)
	if args.Get(0) != nil {
		return args.Get(0).(*entities.PartnerGroup), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPartnerGroupRepository) FindByInviteToken(token string) (*entities.PartnerGroup, error) {
	args := m.Called(token)
	if args.Get(0) != nil {
		return args.Get(0).(*entities.PartnerGroup), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPartnerGroupRepository) AddMember(groupID, userID string) error {
	args := m.Called(groupID, userID)
	return args.Error(0)
}

func (m *MockPartnerGroupRepository) Update(group *entities.PartnerGroup) error {
	args := m.Called(group)
	return args.Error(0)
}
