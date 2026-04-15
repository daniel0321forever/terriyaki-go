package services

import (
	"errors"

	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

var ErrSolanaAdapterNotImplemented = errors.New("solana adapter is enabled but not implemented yet")

// Compile-time check that Solana adapter satisfies PaymentGatewayAdapter.
var _ PaymentGatewayAdapter = (*SolanaPaymentGatewayAdapter)(nil)

// "inherit" from PaymentGatewayAdapter
type SolanaPaymentGatewayAdapter struct{}

func NewSolanaPaymentGatewayAdapter() *SolanaPaymentGatewayAdapter {
	return &SolanaPaymentGatewayAdapter{}
}

func (*SolanaPaymentGatewayAdapter) CreatePaymentIntent(amount int64) (string, error) {
	return "", ErrSolanaAdapterNotImplemented
}

func (*SolanaPaymentGatewayAdapter) CreateSaveCardIntent() (string, error) {
	return "", ErrSolanaAdapterNotImplemented
}

func (*SolanaPaymentGatewayAdapter) CreateCustomer(name string, email string) (string, error) {
	return "", ErrSolanaAdapterNotImplemented
}

func (*SolanaPaymentGatewayAdapter) DescribePaymentMethod(paymentMethodID string) (*entities.PaymentMethodInfo, error) {
	return nil, ErrSolanaAdapterNotImplemented
}

func (*SolanaPaymentGatewayAdapter) AttachPaymentMethodToCustomer(paymentMethodID string, customerID string) error {
	return ErrSolanaAdapterNotImplemented
}

func (*SolanaPaymentGatewayAdapter) Charge(customerID string, paymentMethodID string, amount int64) (string, error) {
	return "", ErrSolanaAdapterNotImplemented
}

func (*SolanaPaymentGatewayAdapter) PayBack(destinationAccountID string, amount int64) error {
	return ErrSolanaAdapterNotImplemented
}
