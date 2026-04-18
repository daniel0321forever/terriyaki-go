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

func (*SolanaPaymentGatewayAdapter) CreateCollectionIntent(req CollectionIntentRequest) (*CollectionIntentResult, error) {
	return nil, ErrSolanaAdapterNotImplemented
}

func (*SolanaPaymentGatewayAdapter) CreatePaymentMethodSetupIntent(req PaymentMethodSetupIntentRequest) (*PaymentMethodSetupIntentResult, error) {
	return nil, ErrSolanaAdapterNotImplemented
}

func (*SolanaPaymentGatewayAdapter) EnsurePayerProfile(req PayerProfileRequest) (*PayerProfileResult, error) {
	return nil, ErrSolanaAdapterNotImplemented
}

func (*SolanaPaymentGatewayAdapter) GetPaymentMethodDetails(paymentMethodID string) (*entities.PaymentMethodInfo, error) {
	return nil, ErrSolanaAdapterNotImplemented
}

func (*SolanaPaymentGatewayAdapter) LinkPaymentMethodToPayer(req PaymentMethodLinkRequest) error {
	return ErrSolanaAdapterNotImplemented
}

func (*SolanaPaymentGatewayAdapter) CreateSettlementIntent(req SettlementIntentRequest) (*SettlementIntentResult, error) {
	return nil, ErrSolanaAdapterNotImplemented
}

func (*SolanaPaymentGatewayAdapter) ResolveSettlement(req SettlementResolutionRequest) (*SettlementResolutionResult, error) {
	return nil, ErrSolanaAdapterNotImplemented
}

func (*SolanaPaymentGatewayAdapter) QuerySettlementStatus(providerReference string) (*SettlementResolutionResult, error) {
	return nil, ErrSolanaAdapterNotImplemented
}

func (*SolanaPaymentGatewayAdapter) CreateDisbursement(req DisbursementRequest) (*DisbursementResult, error) {
	return nil, ErrSolanaAdapterNotImplemented
}
