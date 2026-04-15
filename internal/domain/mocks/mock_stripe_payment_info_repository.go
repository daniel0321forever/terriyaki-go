package mocks

import (
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/stretchr/testify/mock"
)

type MockStripePaymentInfoRepository struct {
	mock.Mock
}

func (m *MockStripePaymentInfoRepository) Create(paymentMethodInfo *entities.PaymentMethodInfo) (*entities.PaymentMethodInfo, error) {
	args := m.Called(paymentMethodInfo)
	if args.Get(0) != nil {
		return args.Get(0).(*entities.PaymentMethodInfo), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStripePaymentInfoRepository) FindByID(id string) (*entities.PaymentMethodInfo, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*entities.PaymentMethodInfo), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStripePaymentInfoRepository) FindByUserID(userID string) ([]entities.PaymentMethodInfo, error) {
	args := m.Called(userID)
	if args.Get(0) != nil {
		return args.Get(0).([]entities.PaymentMethodInfo), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStripePaymentInfoRepository) Update(paymentMethodInfo *entities.PaymentMethodInfo) (*entities.PaymentMethodInfo, error) {
	args := m.Called(paymentMethodInfo)
	if args.Get(0) != nil {
		return args.Get(0).(*entities.PaymentMethodInfo), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStripePaymentInfoRepository) Delete(stripePaymentMethodID string) error {
	args := m.Called(stripePaymentMethodID)
	return args.Error(0)
}
