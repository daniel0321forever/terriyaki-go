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

type WalletMethodRequest struct {
	UserID        string
	WalletAddress string
	Network       string
	ProgramID     string
}

// SettlementAdapter abstracts the shared settlement lifecycle across providers.
// Covers the full money-movement lifecycle: collection intents, settlement intents, resolution, and disbursements.
type SettlementAdapter interface {
	// Initialize payment (prep phase)
	CreateCollectionIntent(req CollectionIntentRequest) (*CollectionIntentResult, error)
	// Charge using saved method
	CreateSettlementIntent(req SettlementIntentRequest) (*SettlementIntentResult, error)
	// Confirm outcome
	ResolveSettlement(req SettlementResolutionRequest) (*SettlementResolutionResult, error)
	// Check progress
	QuerySettlementStatus(providerReference string) (*SettlementResolutionResult, error)
	// Send funds
	CreateDisbursement(req DisbursementRequest) (*DisbursementResult, error)
}

// CardMethodAdapter abstracts card onboarding capabilities used by Stripe.
// These methods are called only during the AddPaymentMethod flow, separate from settlement.
type CardMethodAdapter interface {
	CreatePaymentMethodSetupIntent(req PaymentMethodSetupIntentRequest) (*PaymentMethodSetupIntentResult, error)
	EnsurePayerProfile(req PayerProfileRequest) (*PayerProfileResult, error)
	GetPaymentMethodDetails(paymentMethodID string) (*entities.PaymentMethodInfo, error)
	LinkPaymentMethodToPayer(req PaymentMethodLinkRequest) error
}

// WalletMethodAdapter abstracts wallet onboarding capabilities used by Solana.
type WalletMethodAdapter interface {
	ValidateWalletOwnership(req WalletMethodRequest) error
	NormalizeWalletMethod(req WalletMethodRequest) (*entities.PaymentMethodInfo, error)
}

// PaymentGatewayAdapter is the provider contract for settlement operations.
// All providers must implement the shared settlement lifecycle.
type PaymentGatewayAdapter interface {
	SettlementAdapter
}
