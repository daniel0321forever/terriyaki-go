package services

import (
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/gagliardetto/solana-go"
)

// Abstract-factory request/response payload contracts.
// Providers implement the same lifecycle methods but with provider-specific payloads.
// NOTE: The dummy method is used for compile-time type safety check
type CollectionIntentRequestPayload interface{ isCollectionIntentRequestPayload() }
type CollectionIntentResultPayload interface{ isCollectionIntentResultPayload() }
type SettlementIntentRequestPayload interface{ isSettlementIntentRequestPayload() }
type SettlementIntentResultPayload interface{ isSettlementIntentResultPayload() }
type SettlementResolutionRequestPayload interface{ isSettlementResolutionRequestPayload() }
type SettlementResolutionResultPayload interface{ isSettlementResolutionResultPayload() }
type QuerySettlementStatusRequestPayload interface{ isQuerySettlementStatusRequestPayload() }
type DisbursementRequestPayload interface{ isDisbursementRequestPayload() }
type DisbursementResultPayload interface{ isDisbursementResultPayload() }

type StripeCollectionIntentRequest struct {
	Amount   int64
	Currency string
}

func (StripeCollectionIntentRequest) isCollectionIntentRequestPayload() {}

type StripeCollectionIntentResult struct {
	ProviderReference string
	ClientSecret      string
	Status            entities.SettlementStatus
}

func (*StripeCollectionIntentResult) isCollectionIntentResultPayload() {}

type SolanaCollectionIntentRequest struct {
	Amount       int64
	Currency     string
	PayerPubkey  string
	PledgeID     string
	DeadlineUnix int64
	Network      string
	ProgramID    string
	OraclePubkey string
}

func (SolanaCollectionIntentRequest) isCollectionIntentRequestPayload() {}

type SolanaCollectionIntentResult struct {
	ProviderReference string
	ClientSecret      string
	Status            entities.SettlementStatus
	UnsignedTxJSON    string
	PledgePDA         string
	RecentBlockhash   solana.Hash
	ExpiresAtUnix     int64
}

func (*SolanaCollectionIntentResult) isCollectionIntentResultPayload() {}

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

type StripeSettlementIntentRequest struct {
	CustomerID      string
	PaymentMethodID string
	Amount          int64
	Currency        string
}

func (StripeSettlementIntentRequest) isSettlementIntentRequestPayload() {}

type StripeSettlementIntentResult struct {
	ProviderReference string
	ClientSecret      string
	Status            entities.SettlementStatus
}

func (*StripeSettlementIntentResult) isSettlementIntentResultPayload() {}

type SolanaSettlementIntentRequest struct {
	CustomerID      string
	PaymentMethodID string
	Amount          int64
	Currency        string
}

func (SolanaSettlementIntentRequest) isSettlementIntentRequestPayload() {}

type SolanaSettlementIntentResult struct {
	ProviderReference string
	ClientSecret      string
	Status            entities.SettlementStatus
}

func (*SolanaSettlementIntentResult) isSettlementIntentResultPayload() {}

type StripeSettlementResolutionRequest struct {
	ProviderReference string
	Resolution        entities.SettlementStatus
	Amount            int64
	Currency          string
}

func (StripeSettlementResolutionRequest) isSettlementResolutionRequestPayload() {}

type StripeSettlementResolutionResult struct {
	ProviderReference string
	Status            entities.SettlementStatus
}

func (*StripeSettlementResolutionResult) isSettlementResolutionResultPayload() {}

type SolanaSettlementResolutionRequest struct {
	ProviderReference string
	Resolution        string // "success" or "failure"
	Amount            int64
	Currency          string
	// Oracle signing parameters (for resolve_success/resolve_failure paths)
	PledgePDA          string // base58 pledge account address
	UserPubkey         string // base58 user wallet pubkey (for success) or penalty pool (for failure)
	PenaltyPoolKey     string // base58 penalty pool pubkey (only for failure)
	TxHashProof        string // off-chain transaction ID for audit trail
	Network            string // solana network (devnet/mainnet)
	Operation          string // operation name for audit/logging
}

func (SolanaSettlementResolutionRequest) isSettlementResolutionRequestPayload() {}

type SolanaSettlementResolutionResult struct {
	ProviderReference string
	Status            entities.SettlementStatus
	// Oracle resolution result
	Signature         string // transaction signature (if signed)
	SettlementProof   string // JSON proof (if signed)
	SignedTxBase64    string // base64 signed transaction (if applicable)
}

func (*SolanaSettlementResolutionResult) isSettlementResolutionResultPayload() {}

type StripeQuerySettlementStatusRequest struct {
	ProviderReference string
}

func (StripeQuerySettlementStatusRequest) isQuerySettlementStatusRequestPayload() {}

type SolanaQuerySettlementStatusRequest struct {
	ProviderReference string
}

func (SolanaQuerySettlementStatusRequest) isQuerySettlementStatusRequestPayload() {}

type StripeDisbursementRequest struct {
	DestinationReference string
	Amount               int64
	Currency             string
}

func (StripeDisbursementRequest) isDisbursementRequestPayload() {}

type StripeDisbursementResult struct {
	ProviderReference string
	Status            entities.SettlementStatus
}

func (*StripeDisbursementResult) isDisbursementResultPayload() {}

type SolanaDisbursementRequest struct {
	DestinationReference string
	Amount               int64
	Currency             string
}

func (SolanaDisbursementRequest) isDisbursementRequestPayload() {}

type SolanaDisbursementResult struct {
	ProviderReference string
	Status            entities.SettlementStatus
}

func (*SolanaDisbursementResult) isDisbursementResultPayload() {}

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
	CreateCollectionIntent(req CollectionIntentRequestPayload) (CollectionIntentResultPayload, error)
	// Charge using saved method
	CreateSettlementIntent(req SettlementIntentRequestPayload) (SettlementIntentResultPayload, error)
	// Confirm outcome
	ResolveSettlement(req SettlementResolutionRequestPayload) (SettlementResolutionResultPayload, error)
	// Check progress
	QuerySettlementStatus(req QuerySettlementStatusRequestPayload) (SettlementResolutionResultPayload, error)
	// Send funds
	CreateDisbursement(req DisbursementRequestPayload) (DisbursementResultPayload, error)
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
