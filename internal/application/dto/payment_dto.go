package dto

import (
	"fmt"
	"strings"

	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

func NewAddPaymentMethodDTO(userID, methodType, cardPaymentMethodID, walletAddress, network, programID string) (AddPaymentMethodDTO, error) {
	methodType = strings.TrimSpace(methodType)
	if strings.TrimSpace(userID) == "" {
		return AddPaymentMethodDTO{}, fmt.Errorf("user_id is required")
	}
	if methodType == "" {
		return AddPaymentMethodDTO{}, fmt.Errorf("method_type is required")
	}

	dto := AddPaymentMethodDTO{
		UserID:              strings.TrimSpace(userID),
		MethodType:          methodType,
		CardPaymentMethodID: strings.TrimSpace(cardPaymentMethodID),
		WalletAddress:       strings.TrimSpace(walletAddress),
		Network:             strings.TrimSpace(network),
		ProgramID:           strings.TrimSpace(programID),
	}

	switch dto.MethodType {
	case "card":
		if dto.CardPaymentMethodID == "" {
			return AddPaymentMethodDTO{}, fmt.Errorf("card payment method ID is required")
		}
	case "solana_wallet":
		if dto.WalletAddress == "" {
			return AddPaymentMethodDTO{}, fmt.Errorf("solana wallet address is required")
		}
		if dto.Network == "" {
			return AddPaymentMethodDTO{}, fmt.Errorf("solana network is required")
		}
	default:
		return AddPaymentMethodDTO{}, fmt.Errorf("unsupported payment method type: %s", dto.MethodType)
	}

	return dto, nil
}

func NewGetAvailablePaymentMethodsDTO(userID string) (GetAvailablePaymentMethodsDTO, error) {
	if strings.TrimSpace(userID) == "" {
		return GetAvailablePaymentMethodsDTO{}, fmt.Errorf("user_id is required")
	}
	return GetAvailablePaymentMethodsDTO{UserID: strings.TrimSpace(userID)}, nil
}

func NewSolanaLinkWalletDTO(userID, network, walletAddress string) (SolanaLinkWalletDTO, error) {
	if strings.TrimSpace(userID) == "" {
		return SolanaLinkWalletDTO{}, fmt.Errorf("user_id is required")
	}
	if strings.TrimSpace(network) == "" {
		return SolanaLinkWalletDTO{}, fmt.Errorf("network is required")
	}
	if strings.TrimSpace(walletAddress) == "" {
		return SolanaLinkWalletDTO{}, fmt.Errorf("wallet_address is required")
	}
	return SolanaLinkWalletDTO{
		UserID:        strings.TrimSpace(userID),
		Network:       strings.TrimSpace(network),
		WalletAddress: strings.TrimSpace(walletAddress),
	}, nil
}

func NewSolanaCreateIntentDTO(userID, walletAddress, network, programID, pledgeID, oraclePubkey string, amountLamports, deadlineUnix int64) (SolanaCreateIntentDTO, error) {
	if strings.TrimSpace(userID) == "" {
		return SolanaCreateIntentDTO{}, fmt.Errorf("user_id is required")
	}
	if strings.TrimSpace(walletAddress) == "" {
		return SolanaCreateIntentDTO{}, fmt.Errorf("wallet_address is required")
	}
	if strings.TrimSpace(network) == "" {
		return SolanaCreateIntentDTO{}, fmt.Errorf("network is required")
	}
	if strings.TrimSpace(pledgeID) == "" {
		return SolanaCreateIntentDTO{}, fmt.Errorf("pledge_id is required")
	}
	if amountLamports <= 0 {
		return SolanaCreateIntentDTO{}, fmt.Errorf("amount_lamports must be > 0")
	}
	if deadlineUnix <= 0 {
		return SolanaCreateIntentDTO{}, fmt.Errorf("deadline_unix must be > 0")
	}
	return SolanaCreateIntentDTO{
		UserID:         strings.TrimSpace(userID),
		WalletAddress:  strings.TrimSpace(walletAddress),
		Network:        strings.TrimSpace(network),
		ProgramID:      strings.TrimSpace(programID),
		PledgeID:       strings.TrimSpace(pledgeID),
		AmountLamports: amountLamports,
		DeadlineUnix:   deadlineUnix,
		OraclePubkey:   strings.TrimSpace(oraclePubkey),
	}, nil
}

func NewStripeCreateIntentDTO(userID string, amountCents int64, currency string) (StripeCreateIntentDTO, error) {
	if strings.TrimSpace(userID) == "" {
		return StripeCreateIntentDTO{}, fmt.Errorf("user_id is required")
	}
	if amountCents <= 0 {
		return StripeCreateIntentDTO{}, fmt.Errorf("amount_cents must be > 0")
	}
	return StripeCreateIntentDTO{
		UserID:      strings.TrimSpace(userID),
		AmountCents: amountCents,
		Currency:    strings.TrimSpace(currency),
	}, nil
}

func NewSolanaSubmitSignedTransactionDTO(providerReference, signedBase64, network string) (SolanaSubmitSignedTransactionDTO, error) {
	if strings.TrimSpace(providerReference) == "" {
		return SolanaSubmitSignedTransactionDTO{}, fmt.Errorf("provider_reference is required")
	}
	if strings.TrimSpace(signedBase64) == "" {
		return SolanaSubmitSignedTransactionDTO{}, fmt.Errorf("signed_transaction_base64 is required")
	}
	if strings.TrimSpace(network) == "" {
		return SolanaSubmitSignedTransactionDTO{}, fmt.Errorf("network is required")
	}
	return SolanaSubmitSignedTransactionDTO{
		ProviderReference:       strings.TrimSpace(providerReference),
		SignedTransactionBase64: strings.TrimSpace(signedBase64),
		Network:                 strings.TrimSpace(network),
	}, nil
}

func NewSolanaVerifySettlementDTO(userID, network, transactionSignature string) (SolanaVerifySettlementDTO, error) {
	if strings.TrimSpace(userID) == "" {
		return SolanaVerifySettlementDTO{}, fmt.Errorf("user_id is required")
	}
	if strings.TrimSpace(network) == "" {
		return SolanaVerifySettlementDTO{}, fmt.Errorf("network is required")
	}
	if strings.TrimSpace(transactionSignature) == "" {
		return SolanaVerifySettlementDTO{}, fmt.Errorf("transaction_signature is required")
	}
	return SolanaVerifySettlementDTO{
		UserID:               strings.TrimSpace(userID),
		Network:              strings.TrimSpace(network),
		TransactionSignature: strings.TrimSpace(transactionSignature),
	}, nil
}

func NewChargeWithIdempotencyDTO(paymentMethod entities.PaymentMethodInfo, amountCents int64, operation, userID string) (ChargeWithIdempotencyDTO, error) {
	if strings.TrimSpace(userID) == "" {
		return ChargeWithIdempotencyDTO{}, fmt.Errorf("user_id is required")
	}
	if amountCents <= 0 {
		return ChargeWithIdempotencyDTO{}, fmt.Errorf("amount_cents must be > 0")
	}
	if strings.TrimSpace(operation) == "" {
		return ChargeWithIdempotencyDTO{}, fmt.Errorf("operation is required")
	}
	return ChargeWithIdempotencyDTO{
		PaymentMethodInfo: paymentMethod,
		AmountCents:       amountCents,
		Operation:         strings.TrimSpace(operation),
		UserID:            strings.TrimSpace(userID),
	}, nil
}

func NewPayBackDTO(destinationAccountID string, amountCents int64) (PayBackDTO, error) {
	if strings.TrimSpace(destinationAccountID) == "" {
		return PayBackDTO{}, fmt.Errorf("destination_account_id is required")
	}
	if amountCents <= 0 {
		return PayBackDTO{}, fmt.Errorf("amount_cents must be > 0")
	}
	return PayBackDTO{
		DestinationAccountID: strings.TrimSpace(destinationAccountID),
		AmountCents:          amountCents,
	}, nil
}

func NewSettlementIntentRequestDTO(paymentMethod entities.PaymentMethodInfo, amountCents int64) (SettlementIntentRequestDTO, error) {
	if amountCents <= 0 {
		return SettlementIntentRequestDTO{}, fmt.Errorf("amount_cents must be > 0")
	}
	return SettlementIntentRequestDTO{
		PaymentMethodInfo: paymentMethod,
		AmountCents:       amountCents,
	}, nil
}

func NewReconcileSettlementsDTO(limit int) (ReconcileSettlementsDTO, error) {
	if limit <= 0 {
		return ReconcileSettlementsDTO{}, fmt.Errorf("limit must be > 0")
	}
	return ReconcileSettlementsDTO{Limit: limit}, nil
}

func NewSolanaResolvePledgeDTO(operation, resolution, penaltyPoolKey, pledgePDA, userPubkey, network, txHashProof string) (SolanaResolvePledgeDTO, error) {
	if strings.TrimSpace(operation) == "" {
		return SolanaResolvePledgeDTO{}, fmt.Errorf("operation is required")
	}
	resolution = strings.TrimSpace(resolution)
	if resolution != "success" && resolution != "failure" {
		return SolanaResolvePledgeDTO{}, fmt.Errorf("resolution must be 'success' or 'failure'")
	}
	if strings.TrimSpace(pledgePDA) == "" {
		return SolanaResolvePledgeDTO{}, fmt.Errorf("pledge_pda is required")
	}
	if strings.TrimSpace(userPubkey) == "" {
		return SolanaResolvePledgeDTO{}, fmt.Errorf("user_pubkey is required")
	}
	if strings.TrimSpace(network) == "" {
		return SolanaResolvePledgeDTO{}, fmt.Errorf("network is required")
	}
	if strings.TrimSpace(txHashProof) == "" {
		return SolanaResolvePledgeDTO{}, fmt.Errorf("tx_hash_proof is required")
	}
	return SolanaResolvePledgeDTO{
		Operation:      strings.TrimSpace(operation),
		Resolution:     resolution,
		PenaltyPoolKey: strings.TrimSpace(penaltyPoolKey),
		PledgePDA:      strings.TrimSpace(pledgePDA),
		UserPubkey:     strings.TrimSpace(userPubkey),
		Network:        strings.TrimSpace(network),
		TxHashProof:    strings.TrimSpace(txHashProof),
	}, nil
}

/*
Input DTOs
*/
// AddPaymentMethodDTO is the unified request for adding a payment method (card or wallet).
type AddPaymentMethodDTO struct {
	UserID              string `json:"user_id" binding:"required"`
	MethodType          string `json:"method_type" binding:"required"` // "card" or "solana_wallet"
	CardPaymentMethodID string `json:"card_payment_method_id"`         // Required for card method type
	WalletAddress       string `json:"wallet_address"`                 // Required for solana_wallet method type
	Network             string `json:"network"`                        // Required for solana_wallet (e.g., "mainnet", "devnet")
	ProgramID           string `json:"program_id"`                     // Optional for solana_wallet
}

type GetAvailablePaymentMethodsDTO struct {
	UserID string `json:"user_id" binding:"required"`
}

// SolanaLinkWalletDTO is a placeholder for linking user wallet to payment profile.
type SolanaLinkWalletDTO struct {
	UserID        string `json:"user_id" binding:"required"`
	Network       string `json:"network" binding:"required"`
	WalletAddress string `json:"wallet_address" binding:"required"`
}

// SolanaCreateIntentDTO is a placeholder for creating on-chain settlement intent.
type SolanaCreateIntentDTO struct {
	UserID         string `json:"user_id" binding:"required"`
	WalletAddress  string `json:"wallet_address" binding:"required"`
	Network        string `json:"network" binding:"required"`
	ProgramID      string `json:"program_id"`
	PledgeID       string `json:"pledge_id" binding:"required"`
	AmountLamports int64  `json:"amount_lamports" binding:"required"`
	DeadlineUnix   int64  `json:"deadline_unix" binding:"required"`
	OraclePubkey   string `json:"oracle_pubkey" binding:"required"`
}

// StripeCreateIntentDTO is a placeholder for creating a Stripe collection intent.
type StripeCreateIntentDTO struct {
	UserID      string `json:"user_id"`
	AmountCents int64  `json:"amount_cents" binding:"required"`
	Currency    string `json:"currency"`
}

type SolanaSubmitSignedTransactionDTO struct {
	ProviderReference       string `json:"provider_reference" binding:"required"`
	SignedTransactionBase64 string `json:"signed_transaction_base64" binding:"required"`
	Network                 string `json:"network" binding:"required"`
}

// SolanaVerifySettlementDTO is a placeholder for tx signature verification callbacks.
type SolanaVerifySettlementDTO struct {
	UserID               string `json:"user_id" binding:"required"`
	Network              string `json:"network" binding:"required"`
	TransactionSignature string `json:"transaction_signature" binding:"required"`
}

// ChargeWithIdempotencyDTO is the unified request for charging a pending payment
// or performing a provider-specific charge operation. All amounts are in cents.
type ChargeWithIdempotencyDTO struct {
	PaymentMethodInfo entities.PaymentMethodInfo `json:"payment_method_info"`
	AmountCents       int64                      `json:"amount_cents"`
	Operation         string                     `json:"operation"`
	UserID            string                     `json:"user_id"`
}

// PayBackDTO represents a disbursement request (amount in cents).
type PayBackDTO struct {
	DestinationAccountID string `json:"destination_account_id" binding:"required"`
	AmountCents          int64  `json:"amount_cents" binding:"required"`
}

type SettlementIntentRequestDTO struct {
	PaymentMethodInfo entities.PaymentMethodInfo `json:"payment_method_info"`
	AmountCents       int64                      `json:"amount_cents"`
}

// ReconcileSettlementsDTO allows configuring reconciliation batch limits.
type ReconcileSettlementsDTO struct {
	Limit int `json:"limit"`
}

// SolanaResolvePledgeDTO represents a request to resolve a pledge using the oracle.
// This enables the hybrid flow where the oracle backend signs the resolution (success or failure).
type SolanaResolvePledgeDTO struct {
	Operation      string `json:"operation" binding:"required"`     // "solana_collection_intent" or similar
	Resolution     string `json:"resolution" binding:"required"`    // "success" or "failure"
	PenaltyPoolKey string `json:"penalty_pool_key"`                 // Required for failure resolution
	PledgePDA      string `json:"pledge_pda" binding:"required"`    // Pledge account address
	UserPubkey     string `json:"user_pubkey" binding:"required"`   // User wallet address
	Network        string `json:"network" binding:"required"`       // "mainnet", "devnet", etc.
	TxHashProof    string `json:"tx_hash_proof" binding:"required"` // The off-chain TX hash for audit trail
}

// SolanaResolvePledgeResultDTO is the response after oracle-signed resolution.
type SolanaResolvePledgeResultDTO struct {
	ProviderReference string `json:"provider_reference"`
	Signature         string `json:"signature"`
	Status            string `json:"status"`
	Resolution        string `json:"resolution"` // "success" or "failure"
	SettlementProof   string `json:"settlement_proof"`
}

/*
Output DTOs
*/
type AvailablePaymentMethodsDTO struct {
	PaymentInfos       []entities.PaymentMethodInfo `json:"payment_infos"`
	DefaultPaymentInfo entities.PaymentMethodInfo   `json:"default_payment_info"`
}

type PendingPaymentDTO struct {
	PaymentMethodInfo entities.PaymentMethodInfo `json:"payment_method_info"`
	PaymentAmount     int64                      `json:"payment_amount"`
}

type PendingPaymentsResultDTO struct {
	PendingPayments []PendingPaymentDTO `json:"pending_payments"`
}

type StripeCreateCollectionIntentResultDTO struct {
	ClientSecret     string `json:"client_secret"`
	IdempotentReplay bool   `json:"idempotent_replay"`
}

type SolanaCreateCollectionIntentResultDTO struct {
	ProviderReference string `json:"provider_reference"`
	UnsignedTxJSON    string `json:"unsigned_tx_json"`
	PledgePDA         string `json:"pledge_pda"`
	RecentBlockhash   string `json:"recent_blockhash"`
	ExpiresAtUnix     int64  `json:"expires_at_unix"`
}

type SolanaSubmitSignedTransactionResultDTO struct {
	ProviderReference string `json:"provider_reference"`
	Signature         string `json:"signature"`
	Status            string `json:"status"`
	SettlementProof   string `json:"settlement_proof"`
}

type SolanaSettlementStatusDTO struct {
	Network              string `json:"network"`
	TransactionSignature string `json:"transaction_signature"`
	Status               string `json:"status"`
	FinalizedAtUnix      int64  `json:"finalized_at_unix"`
}

type ChargeWithIdempotencyResultDTO struct {
	ProviderReference string `json:"provider_reference"`
	IdempotentReplay  bool   `json:"idempotent_replay"`
}

type ClaimIdempotencyResultDTO struct {
	Claimed bool `json:"claimed"`
}

type AddPaymentMethodResultDTO struct {
	PaymentMethod entities.PaymentMethodInfo `json:"payment_method"`
}

type SettlementReferenceDTO struct {
	ProviderReference string `json:"provider_reference"`
	Network           string `json:"network"`
	TxHash            string `json:"tx_hash"`
	ContractAddress   string `json:"contract_address"`
	SettlementProof   string `json:"settlement_proof"`
	FinalizedAtUnix   int64  `json:"finalized_at_unix"`
}

type PaymentSettlementDTO struct {
	ID              uint                      `json:"id"`
	UserID          string                    `json:"user_id"`
	Operation       string                    `json:"operation"`
	IdempotencyKey  string                    `json:"idempotency_key"`
	Provider        entities.PaymentProvider  `json:"provider"`
	PaymentMethodID string                    `json:"payment_method_id"`
	Status          entities.SettlementStatus `json:"status"`
	Amount          int64                     `json:"amount"`
	Currency        string                    `json:"currency"`
	RetryCount      int                       `json:"retry_count"`
	LastError       string                    `json:"last_error"`
	Reference       SettlementReferenceDTO    `json:"reference"`
	CreatedAtUnix   int64                     `json:"created_at_unix"`
	UpdatedAtUnix   int64                     `json:"updated_at_unix"`
}

type ReconcileSettlementsResultDTO struct {
	UpdatedSettlements []PaymentSettlementDTO `json:"updated_settlements"`
}

type PayBackResultDTO struct {
	ProviderReference string `json:"provider_reference"`
	Status            string `json:"status"`
}
