// focus on high-level model for canonical settlement (payment) lifecycles across all payment providers (e.g., Stripe and Solana)
package entities

import "time"

// specific to Stripe for now
type StripePaymentInfo struct {
	UserID                string `json:"user_id" gorm:"not null"`
	StripeCustomerID      string `json:"stripe_customer_id" gorm:"not null"`
	StripePaymentMethodID string `json:"stripe_payment_method_id" gorm:"primaryKey;not null;unique"`
	Brand                 string `json:"brand" gorm:""`
	Last4                 string `json:"last4" gorm:""`
	ExpMonth              int    `json:"exp_month" gorm:""`
	ExpYear               int    `json:"exp_year" gorm:""`
}

func NewStripePaymentInfo(userID string, stripeCustomerID string, stripePaymentMethodID string, brand string, last4 string, expMonth int, expYear int) *StripePaymentInfo {
	return &StripePaymentInfo{
		UserID:                userID,
		StripeCustomerID:      stripeCustomerID,
		StripePaymentMethodID: stripePaymentMethodID,
		Brand:                 brand,
		Last4:                 last4,
		ExpMonth:              expMonth,
		ExpYear:               expYear,
	}
}

// PaymentMethodInfo is the provider-neutral alias used by application/domain layers.
// It is intentionally provider-neutral to support multi-rail payment methods.
type PaymentMethodInfo struct {
	UserID                  string          `json:"user_id" gorm:"not null"`
	Provider                PaymentProvider `json:"provider" gorm:"not null"`
	ProviderCustomerID      string          `json:"provider_customer_id" gorm:""`
	ProviderPaymentMethodID string          `json:"provider_payment_method_id" gorm:"primaryKey;not null;unique"`
	MethodType              string          `json:"method_type" gorm:""`
	Brand                   string          `json:"brand" gorm:""`
	Last4                   string          `json:"last4" gorm:""`
	ExpMonth                int             `json:"exp_month" gorm:""`
	ExpYear                 int             `json:"exp_year" gorm:""`
	Network                 string          `json:"network" gorm:""`
	WalletAddress           string          `json:"wallet_address" gorm:""`
}

func NewPaymentMethodInfo(provider PaymentProvider, methodType string, userID string, providerCustomerID string, providerPaymentMethodID string, brand string, last4 string, expMonth int, expYear int) *PaymentMethodInfo {
	return &PaymentMethodInfo{
		UserID:                  userID,
		Provider:                provider,
		ProviderCustomerID:      providerCustomerID,
		ProviderPaymentMethodID: providerPaymentMethodID,
		MethodType:              methodType,
		Brand:                   brand,
		Last4:                   last4,
		ExpMonth:                expMonth,
		ExpYear:                 expYear,
	}
}

type PaymentProvider string

const (
	PaymentProviderStripe PaymentProvider = "stripe"
	PaymentProviderSolana PaymentProvider = "solana"
)

type SettlementStatus string

const (
	SettlementStatusPending        SettlementStatus = "pending"
	SettlementStatusAuthorized     SettlementStatus = "authorized"
	SettlementStatusCaptured       SettlementStatus = "captured"
	SettlementStatusFailed         SettlementStatus = "failed"
	SettlementStatusRefunded       SettlementStatus = "refunded"
	SettlementStatusSettledOnChain SettlementStatus = "settled_onchain"
)

// SettlementReference captures provider-neutral references used to reconcile settlements.
type SettlementReference struct {
	ProviderReference string `json:"provider_reference" gorm:""`
	Network           string `json:"network" gorm:""`
	TxHash            string `json:"tx_hash" gorm:""`
	ContractAddress   string `json:"contract_address" gorm:""`
	SettlementProof   string `json:"settlement_proof" gorm:""`
	FinalizedAtUnix   int64  `json:"finalized_at_unix" gorm:""`
}

type PaymentSettlement struct {
	ID              uint                `json:"id"`
	UserID          string              `json:"user_id"`
	Operation       string              `json:"operation"`
	IdempotencyKey  string              `json:"idempotency_key"`
	Provider        PaymentProvider     `json:"provider"`
	PaymentMethodID string              `json:"payment_method_id"`
	Status          SettlementStatus    `json:"status"`
	Amount          int64               `json:"amount"`
	Currency        string              `json:"currency"`
	RetryCount      int                 `json:"retry_count"`
	LastError       string              `json:"last_error"`
	Reference       SettlementReference `json:"reference"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
}

func NewPaymentSettlement(userID string, operation string, idempotencyKey string, provider PaymentProvider, paymentMethodID string, amount int64) *PaymentSettlement {
	return &PaymentSettlement{
		UserID:          userID,
		Operation:       operation,
		IdempotencyKey:  idempotencyKey,
		Provider:        provider,
		PaymentMethodID: paymentMethodID,
		Status:          SettlementStatusPending,
		Amount:          amount,
		Currency:        "usd",
		RetryCount:      0,
		Reference:       SettlementReference{},
	}
}

// SolanaPaymentMethodInfo is a placeholder model for wallet-based payment method linkage.
// It does not alter current Stripe behavior and is reserved for future Solana implementation.
type SolanaPaymentMethodInfo struct {
	UserID        string `json:"user_id" gorm:"not null"`
	Network       string `json:"network" gorm:"not null"`
	WalletAddress string `json:"wallet_address" gorm:"not null"`
	ProgramID     string `json:"program_id" gorm:""`
}

func NewSolanaPaymentMethodInfo(userID string, network string, walletAddress string, programID string) *SolanaPaymentMethodInfo {
	return &SolanaPaymentMethodInfo{
		UserID:        userID,
		Network:       network,
		WalletAddress: walletAddress,
		ProgramID:     programID,
	}
}

// SolanaSettlementInfo is a placeholder model for on-chain settlement tracking.
type SolanaSettlementInfo struct {
	UserID               string           `json:"user_id" gorm:"not null"`
	Network              string           `json:"network" gorm:"not null"`
	TransactionSignature string           `json:"transaction_signature" gorm:"not null"`
	ContractAddress      string           `json:"contract_address" gorm:""`
	Status               SettlementStatus `json:"status" gorm:"not null"`
	FinalizedAtUnix      int64            `json:"finalized_at_unix" gorm:""`
}

func NewSolanaSettlementInfo(userID string, network string, signature string, contractAddress string, status SettlementStatus, finalizedAtUnix int64) *SolanaSettlementInfo {
	return &SolanaSettlementInfo{
		UserID:               userID,
		Network:              network,
		TransactionSignature: signature,
		ContractAddress:      contractAddress,
		Status:               status,
		FinalizedAtUnix:      finalizedAtUnix,
	}
}
