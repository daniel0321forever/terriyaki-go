package dto

import "github.com/daniel0321forever/terriyaki-go/internal/domain/entities"

type SaveCardDTO struct {
	UserID          string `json:"user_id" binding:"required"`
	PaymentMethodID string `json:"payment_method_id" binding:"required"`
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
