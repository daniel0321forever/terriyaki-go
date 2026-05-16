package dto

import "github.com/daniel0321forever/terriyaki-go/internal/domain/entities"

// AddPaymentMethodDTO is the unified request for adding a payment method (card or wallet).
type AddPaymentMethodDTO struct {
	UserID              string `json:"user_id" binding:"required"`
	MethodType          string `json:"method_type" binding:"required"`      // "card" or "solana_wallet"
	CardPaymentMethodID string `json:"card_payment_method_id"`              // Required for card method type
	WalletAddress       string `json:"wallet_address"`                      // Required for solana_wallet method type
	Network             string `json:"network"`                             // Required for solana_wallet (e.g., "mainnet", "devnet")
	ProgramID           string `json:"program_id"`                          // Optional for solana_wallet
}

type GetAvailablePaymentMethodsDTO struct {
	UserID string `json:"user_id" binding:"required"`
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

// SolanaLinkWalletDTO is a placeholder for linking user wallet to payment profile.
type SolanaLinkWalletDTO struct {
	UserID        string `json:"user_id" binding:"required"`
	Network       string `json:"network" binding:"required"`
	WalletAddress string `json:"wallet_address" binding:"required"`
}

// SolanaCreateIntentDTO is a placeholder for creating on-chain settlement intent.
type SolanaCreateIntentDTO struct {
	UserID        string `json:"user_id" binding:"required"`
	Network       string `json:"network" binding:"required"`
	AmountLamports int64 `json:"amount_lamports" binding:"required"`
	MintAddress   string `json:"mint_address"`
}

// SolanaVerifySettlementDTO is a placeholder for tx signature verification callbacks.
type SolanaVerifySettlementDTO struct {
	UserID               string `json:"user_id" binding:"required"`
	Network              string `json:"network" binding:"required"`
	TransactionSignature string `json:"transaction_signature" binding:"required"`
}

// SolanaSettlementStatusDTO is a placeholder response view for on-chain settlement state.
type SolanaSettlementStatusDTO struct {
	Network              string `json:"network"`
	TransactionSignature string `json:"transaction_signature"`
	Status               string `json:"status"`
	FinalizedAtUnix      int64  `json:"finalized_at_unix"`
}
