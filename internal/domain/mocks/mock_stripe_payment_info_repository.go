package mocks

import (
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/stretchr/testify/mock"
)

type MockStripePaymentInfoRepository struct {
	mock.Mock
}

func (m *MockStripePaymentInfoRepository) Create(userID, stripeCustomerID, stripePaymentMethodID string) (*entities.StripePaymentInfo, error) {
	args := m.Called(userID, stripeCustomerID, stripePaymentMethodID)
	if args.Get(0) != nil {
		return args.Get(0).(*entities.StripePaymentInfo), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStripePaymentInfoRepository) FindByID(id string) (*entities.StripePaymentInfo, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*entities.StripePaymentInfo), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStripePaymentInfoRepository) FindByUserID(userID string) ([]entities.StripePaymentInfo, error) {
	args := m.Called(userID)
	if args.Get(0) != nil {
		return args.Get(0).([]entities.StripePaymentInfo), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStripePaymentInfoRepository) Update(stripePaymentInfo *entities.StripePaymentInfo) (*entities.StripePaymentInfo, error) {
	args := m.Called(stripePaymentInfo)
	if args.Get(0) != nil {
		return args.Get(0).(*entities.StripePaymentInfo), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStripePaymentInfoRepository) Delete(stripePaymentMethodID string) error {
	args := m.Called(stripePaymentMethodID)
	return args.Error(0)
}
