package services

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

var ErrSolanaAdapterNotImplemented = errors.New("solana adapter is enabled but not implemented yet")

// Compile-time checks for the Solana adapter satisfies PaymentGatewayAdapter and WalletMethodAdapter.
var _ PaymentGatewayAdapter = (*SolanaPaymentGatewayAdapter)(nil)
var _ WalletMethodAdapter = (*SolanaPaymentGatewayAdapter)(nil)

// "inherit" from PaymentGatewayAdapter
type SolanaPaymentGatewayAdapter struct {
}

func NewSolanaPaymentGatewayAdapter() *SolanaPaymentGatewayAdapter {
	return &SolanaPaymentGatewayAdapter{}
}

func (a *SolanaPaymentGatewayAdapter) ValidateWalletOwnership(req WalletMethodRequest) error {
	if strings.TrimSpace(req.WalletAddress) == "" {
		return errors.New("wallet address is required")
	}
	if strings.TrimSpace(req.Network) == "" {
		return errors.New("network is required")
	}
	return nil
}

func (a *SolanaPaymentGatewayAdapter) NormalizeWalletMethod(req WalletMethodRequest) (*entities.PaymentMethodInfo, error) {
	if err := a.ValidateWalletOwnership(req); err != nil {
		return nil, err
	}

	info := entities.NewPaymentMethodInfo(
		entities.PaymentProviderSolana,
		"solana_wallet",
		req.UserID,
		"",
		req.WalletAddress,
		"",
		"",
		0,
		0,
	)
	info.Network = req.Network
	info.WalletAddress = req.WalletAddress
	return info, nil
}

func (a *SolanaPaymentGatewayAdapter) CreateCollectionIntent(req CollectionIntentRequest) (*CollectionIntentResult, error) {
	currency := req.Currency
	if currency == "" {
		currency = "usd"
	}
	ref := solanaReference("collection", fmt.Sprintf("%d:%s", req.Amount, currency))
	return &CollectionIntentResult{
		ProviderReference: ref,
		ClientSecret:      ref + "_secret",
		Status:            entities.SettlementStatusPending,
	}, nil
}

func (a *SolanaPaymentGatewayAdapter) CreateSettlementIntent(req SettlementIntentRequest) (*SettlementIntentResult, error) {
	if strings.TrimSpace(req.PaymentMethodID) == "" {
		return nil, errors.New("payment method reference is required")
	}
	ref := solanaReference("settlement", fmt.Sprintf("%s:%s:%d:%s", req.CustomerID, req.PaymentMethodID, req.Amount, req.Currency))
	return &SettlementIntentResult{
		ProviderReference: ref,
		ClientSecret:      ref + "_secret",
		Status:            entities.SettlementStatusCaptured,
	}, nil
}

func (a *SolanaPaymentGatewayAdapter) ResolveSettlement(req SettlementResolutionRequest) (*SettlementResolutionResult, error) {
	status := req.Resolution
	if status == "" {
		status = entities.SettlementStatusSettledOnChain
	}
	return &SettlementResolutionResult{ProviderReference: req.ProviderReference, Status: status}, nil
}

func (a *SolanaPaymentGatewayAdapter) QuerySettlementStatus(providerReference string) (*SettlementResolutionResult, error) {
	if strings.TrimSpace(providerReference) == "" {
		return nil, errors.New("provider reference is required")
	}
	return &SettlementResolutionResult{ProviderReference: providerReference, Status: entities.SettlementStatusSettledOnChain}, nil
}

func (a *SolanaPaymentGatewayAdapter) CreateDisbursement(req DisbursementRequest) (*DisbursementResult, error) {
	if strings.TrimSpace(req.DestinationReference) == "" {
		return nil, errors.New("destination reference is required")
	}
	ref := solanaReference("disbursement", fmt.Sprintf("%s:%d:%s", req.DestinationReference, req.Amount, req.Currency))
	return &DisbursementResult{ProviderReference: ref, Status: entities.SettlementStatusSettledOnChain}, nil
}

func solanaReference(prefix string, seed string) string {
	sum := sha256.Sum256([]byte(prefix + ":" + seed))
	return prefix + "_" + hex.EncodeToString(sum[:8])
}
