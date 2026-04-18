package services

import "github.com/daniel0321forever/terriyaki-go/internal/domain/entities"

type CollectionIntentRequest struct {
	Amount   int64
	Currency string
}

type CollectionIntentResult struct {
	ProviderReference string
	ClientSecret      string
	Status            entities.SettlementStatus
}

type PaymentMethodSetupIntentRequest struct {
	Usage string
}

type PaymentMethodSetupIntentResult struct {
	ProviderReference string
	ClientSecret      string
}

type PayerProfileRequest struct {
	Name                   string
	Email                  string
	ExistingPayerReference string
}

type PayerProfileResult struct {
	PayerReference string
}

type PaymentMethodLinkRequest struct {
	PaymentMethodID string
	PayerReference  string
}

type SettlementIntentRequest struct {
	CustomerID      string
	PaymentMethodID string
	Amount          int64
	Currency        string
}

type SettlementIntentResult struct {
	ProviderReference string
	ClientSecret      string
	Status            entities.SettlementStatus
}

type SettlementResolutionRequest struct {
	ProviderReference string
	Resolution        entities.SettlementStatus
	Amount            int64
	Currency          string
}

type SettlementResolutionResult struct {
	ProviderReference string
	Status            entities.SettlementStatus
}

type DisbursementRequest struct {
	DestinationReference string
	Amount               int64
	Currency             string
}

type DisbursementResult struct {
	ProviderReference string
	Status            entities.SettlementStatus
}

// PaymentGatewayAdapter abstracts provider-specific payment operations.
type PaymentGatewayAdapter interface {
	// Provider-neutral payment method lifecycle.
	CreateCollectionIntent(req CollectionIntentRequest) (*CollectionIntentResult, error)
	CreatePaymentMethodSetupIntent(req PaymentMethodSetupIntentRequest) (*PaymentMethodSetupIntentResult, error)
	EnsurePayerProfile(req PayerProfileRequest) (*PayerProfileResult, error)
	GetPaymentMethodDetails(paymentMethodID string) (*entities.PaymentMethodInfo, error)
	LinkPaymentMethodToPayer(req PaymentMethodLinkRequest) error

	// Provider-neutral settlement lifecycle.
	CreateSettlementIntent(req SettlementIntentRequest) (*SettlementIntentResult, error)
	ResolveSettlement(req SettlementResolutionRequest) (*SettlementResolutionResult, error)
	QuerySettlementStatus(providerReference string) (*SettlementResolutionResult, error)
	CreateDisbursement(req DisbursementRequest) (*DisbursementResult, error)
}
