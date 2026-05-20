package services

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	controlsolana "github.com/daniel0321forever/terriyaki-go/control/solana"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

var ErrSolanaAdapterNotImplemented = errors.New("solana adapter is enabled but not implemented yet")

// Compile-time checks for the Solana adapter satisfies PaymentGatewayAdapter and WalletMethodAdapter.
var _ PaymentGatewayAdapter = (*SolanaPaymentGatewayAdapter)(nil)
var _ WalletMethodAdapter = (*SolanaPaymentGatewayAdapter)(nil)

// "inherit" from PaymentGatewayAdapter
type SolanaPaymentGatewayAdapter struct {
	rpcEndpoint      string
	programID        [32]byte
	oraclePubkey     [32]byte
	oraclePrivateKey [64]byte
	ctx              context.Context
}

func NewSolanaPaymentGatewayAdapter(rpcEndpoint string, programID [32]byte, oraclePubkey [32]byte, oraclePrivateKey [64]byte) *SolanaPaymentGatewayAdapter {
	return &SolanaPaymentGatewayAdapter{
		rpcEndpoint:      rpcEndpoint,
		programID:        programID,
		oraclePubkey:     oraclePubkey,
		oraclePrivateKey: oraclePrivateKey,
		ctx:              context.Background(),
	}
}

func (a *SolanaPaymentGatewayAdapter) RPCEndpoint() string {
	return a.rpcEndpoint
}

func (a *SolanaPaymentGatewayAdapter) ProgramID() [32]byte {
	return a.programID
}

func (a *SolanaPaymentGatewayAdapter) OraclePubkey() [32]byte {
	return a.oraclePubkey
}

func (a *SolanaPaymentGatewayAdapter) OraclePrivateKey() [64]byte {
	return a.oraclePrivateKey
}

func (a *SolanaPaymentGatewayAdapter) ValidateWalletOwnership(req WalletMethodRequest) error {
	if req.WalletAddress == "" {
		return errors.New("wallet address is required")
	}
	if req.Network == "" {
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
	info.ProviderPaymentMethodID = req.WalletAddress
	info.WalletAddress = req.WalletAddress
	return info, nil
}

func (a *SolanaPaymentGatewayAdapter) CreateCollectionIntent(req_ CollectionIntentRequestPayload) (CollectionIntentResultPayload, error) {
	req, ok := req_.(SolanaCollectionIntentRequest)
	if !ok {
		return nil, fmt.Errorf("solana CreateCollectionIntent requires SolanaCollectionIntentRequest")
	}

	// Non-custodial flow: build an unsigned transaction the client wallet will sign.
	// Ensure we have a recent blockhash from RPC.
	rpcClient := rpc.New(a.rpcEndpoint)
	rpcResult, err := rpcClient.GetLatestBlockhash(a.ctx, rpc.CommitmentFinalized)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recent blockhash: %w", err)
	}
	recentBlockhash := rpcResult.Value.Blockhash

	// map request inputs into control/solana builder args
	// Payer pubkey is required for unsigned tx; if not provided, return guidance error.
	if req.PayerPubkey == "" || req.PledgeID == "" {
		return nil, fmt.Errorf("payer_pubkey and pledge_id are required for Solana collection intents")
	}

	// NOTE: parsing of base58 pubkeys and program ids will be added when
	// the runtime configuration and canonical parsing helpers are finalized.
	// For now use zero-value placeholders for the builder call.
	payerKey, err := solana.PublicKeyFromBase58(req.PayerPubkey)
	if err != nil {
		return nil, fmt.Errorf("invalid payer pubkey: %w", err)
	}

	var payerArr [32]byte
	copy(payerArr[:], payerKey.Bytes())
	var systemProgram [32]byte
	copy(systemProgram[:], solana.SystemProgramID.Bytes())

	// build unsigned transaction
	unsignedTxJSON, pledgePDA, err := controlsolana.BuildInitializePledgeUnsignedTx(
		recentBlockhash,
		payerArr,
		req.PledgeID,
		a.oraclePubkey,
		uint64(req.Amount),
		req.DeadlineUnix,
		systemProgram,
		a.programID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build unsigned pledge transaction: %w", err)
	}

	ref := solanaReference("collection", req.PledgeID)
	return &SolanaCollectionIntentResult{
		ProviderReference: ref,
		ClientSecret:      ref + "_secret",
		Status:            entities.SettlementStatusPending,
		UnsignedTxJSON:    string(unsignedTxJSON),
		PledgePDA:         solana.PublicKeyFromBytes(pledgePDA[:]).String(),
		RecentBlockhash:   recentBlockhash,
		ExpiresAtUnix:     req.DeadlineUnix,
	}, nil
}

func (a *SolanaPaymentGatewayAdapter) CreateSettlementIntent(req_ SettlementIntentRequestPayload) (SettlementIntentResultPayload, error) {
	req, ok := req_.(SolanaSettlementIntentRequest)
	if !ok {
		return nil, fmt.Errorf("solana CreateSettlementIntent requires SolanaSettlementIntentRequest")
	}

	rpcClient := rpc.New(a.rpcEndpoint)
	rpcResult, err := rpcClient.GetLatestBlockhash(a.ctx, rpc.CommitmentFinalized)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recent blockhash for settlement intent: %w", err)
	}

	paymentMethodPubkey, err := solana.PublicKeyFromBase58(req.PaymentMethodID)
	if err != nil {
		return nil, fmt.Errorf("invalid payment method reference (expected base58 pubkey): %w", err)
	}

	acct, err := rpcClient.GetAccountInfo(a.ctx, paymentMethodPubkey)
	if err != nil {
		return nil, fmt.Errorf("failed to query payment method account: %w", err)
	}

	if acct.Value == nil {
		return nil, fmt.Errorf("payment method account not found on chain")
	}

	ref := solanaReference("settlement", fmt.Sprintf("%s:%s:%d:%s:%s", req.CustomerID, req.PaymentMethodID, req.Amount, req.Currency, rpcResult.Value.Blockhash.String()))
	return &SolanaSettlementIntentResult{
		ProviderReference: ref,
		ClientSecret:      ref + "_secret",
		Status:            entities.SettlementStatusCaptured,
	}, nil
}

func (a *SolanaPaymentGatewayAdapter) ResolveSettlement(req_ SettlementResolutionRequestPayload) (SettlementResolutionResultPayload, error) {
	req, ok := req_.(SolanaSettlementResolutionRequest)
	if !ok {
		return nil, fmt.Errorf("solana ResolveSettlement requires SolanaSettlementResolutionRequest")
	}

	// Oracle resolution path: sign and submit transaction
	var systemProgram [32]byte
	copy(systemProgram[:], solana.SystemProgramID.Bytes())

	// Parse pledge PDA
	pledgePubkey, err := solana.PublicKeyFromBase58(req.PledgePDA)
	if err != nil {
		return nil, fmt.Errorf("invalid pledge pda: %w", err)
	}
	var pledgePDA [32]byte
	copy(pledgePDA[:], pledgePubkey.Bytes())

	// Sign based on resolution type
	var signedTx controlsolana.SignedTransaction
	switch req.Resolution {
	case "success":
		userPubkey, err := solana.PublicKeyFromBase58(req.UserPubkey)
		if err != nil {
			return nil, fmt.Errorf("invalid user pubkey: %w", err)
		}
		var userPub [32]byte
		copy(userPub[:], userPubkey.Bytes())

		signed, signErr := controlsolana.SignResolveSuccessTransaction(
			a.oraclePubkey,
			a.oraclePrivateKey,
			pledgePDA,
			userPub,
			req.TxHashProof,
			time.Now().Unix(),
			systemProgram,
			a.programID,
		)
		if signErr != nil {
			return nil, signErr
		}
		signedTx = signed

	case "failure":
		penaltyPoolPubkey, err := solana.PublicKeyFromBase58(req.PenaltyPoolKey)
		if err != nil {
			return nil, fmt.Errorf("invalid penalty pool key: %w", err)
		}
		var penaltyPool [32]byte
		copy(penaltyPool[:], penaltyPoolPubkey.Bytes())

		signed, signErr := controlsolana.SignResolveFailureTransaction(
			a.oraclePubkey,
			a.oraclePrivateKey,
			pledgePDA,
			penaltyPool,
			req.TxHashProof,
			time.Now().Unix(),
			systemProgram,
			a.programID,
		)
		if signErr != nil {
			return nil, signErr
		}
		signedTx = signed

	default:
		return nil, fmt.Errorf("unsupported resolution: %s", req.Resolution)
	}

	// Submit transaction to Solana RPC
	submittedSig, submitErr := controlsolana.SubmitTransactionWithRetry(a.rpcEndpoint, signedTx, 3)
	if submitErr != nil {
		return nil, submitErr
	}

	signature := solana.SignatureFromBytes(submittedSig[:]).String()
	proof := map[string]any{
		"provider_reference": req.ProviderReference,
		"signature":          signature,
		"signed_tx_base64":   base64.StdEncoding.EncodeToString(signedTx.Bytes),
		"submitted_at_unix":  time.Now().Unix(),
		"network":            req.Network,
		"resolution":         req.Resolution,
	}
	proofJSON, _ := json.Marshal(proof)

	return &SolanaSettlementResolutionResult{
		ProviderReference: req.ProviderReference,
		Status:            entities.SettlementStatusSettledOnChain,
		Signature:         signature,
		SettlementProof:   string(proofJSON),
		SignedTxBase64:    base64.StdEncoding.EncodeToString(signedTx.Bytes),
	}, nil
}

func (a *SolanaPaymentGatewayAdapter) QuerySettlementStatus(req_ QuerySettlementStatusRequestPayload) (SettlementResolutionResultPayload, error) {
	req, ok := req_.(SolanaQuerySettlementStatusRequest)
	if !ok {
		return nil, fmt.Errorf("solana QuerySettlementStatus requires SolanaQuerySettlementStatusRequest")
	}

	rpcClient := rpc.New(a.rpcEndpoint)

	// Best-effort path 1: treat provider reference as a transaction signature.
	if sig, sigErr := solana.SignatureFromBase58(req.ProviderReference); sigErr == nil {
		statuses, err := rpcClient.GetSignatureStatuses(a.ctx, true, sig)
		if err != nil {
			return nil, fmt.Errorf("failed to query signature status: %w", err)
		}

		status := entities.SettlementStatusPending
		proof := map[string]any{
			"provider_reference": req.ProviderReference,
			"queried_at_unix":    time.Now().Unix(),
			"query_type":         "signature",
		}

		if len(statuses.Value) > 0 && statuses.Value[0] != nil {
			sigStatus := statuses.Value[0]
			proof["confirmation_status"] = sigStatus.ConfirmationStatus
			proof["slot"] = sigStatus.Slot

			if sigStatus.Err != nil {
				status = entities.SettlementStatusFailed
				proof["rpc_error"] = sigStatus.Err
			} else if sigStatus.ConfirmationStatus == rpc.ConfirmationStatusFinalized {
				status = entities.SettlementStatusSettledOnChain
			}
		} else {
			proof["confirmation_status"] = "not_found"
		}

		proofJSON, _ := json.Marshal(proof)
		return &SolanaSettlementResolutionResult{
			ProviderReference: req.ProviderReference,
			Status:            status,
			Signature:         req.ProviderReference,
			SettlementProof:   string(proofJSON),
		}, nil
	}

	// Best-effort path 2: treat provider reference as a pledge PDA/account address.
	if pledgeAccount, pkErr := solana.PublicKeyFromBase58(req.ProviderReference); pkErr == nil {
		acct, err := rpcClient.GetAccountInfo(a.ctx, pledgeAccount)
		if err != nil {
			return nil, fmt.Errorf("failed to query pledge account: %w", err)
		}

		status := entities.SettlementStatusPending
		proof := map[string]any{
			"provider_reference": req.ProviderReference,
			"queried_at_unix":    time.Now().Unix(),
			"query_type":         "account",
		}

		if acct.Value == nil {
			proof["account_state"] = "not_found"
		} else {
			proof["owner"] = acct.Value.Owner.String()
			proof["lamports"] = acct.Value.Lamports
			proof["executable"] = acct.Value.Executable

			// In this program, escrow lamports are drained on successful/failure resolution.
			if acct.Value.Lamports == 0 {
				status = entities.SettlementStatusSettledOnChain
			}
		}

		proofJSON, _ := json.Marshal(proof)
		return &SolanaSettlementResolutionResult{
			ProviderReference: req.ProviderReference,
			Status:            status,
			SettlementProof:   string(proofJSON),
		}, nil
	}

	// Unknown provider reference format. Keep as pending until a canonical reference is provided.
	proofJSON, _ := json.Marshal(map[string]any{
		"provider_reference": req.ProviderReference,
		"queried_at_unix":    time.Now().Unix(),
		"query_type":         "unknown",
		"note":               "provider_reference is neither a signature nor a public key",
	})
	return &SolanaSettlementResolutionResult{
		ProviderReference: req.ProviderReference,
		Status:            entities.SettlementStatusPending,
		SettlementProof:   string(proofJSON),
	}, nil
}

func (a *SolanaPaymentGatewayAdapter) CreateDisbursement(req_ DisbursementRequestPayload) (DisbursementResultPayload, error) {
	req, ok := req_.(SolanaDisbursementRequest)
	if !ok {
		return nil, fmt.Errorf("solana CreateDisbursement requires SolanaDisbursementRequest")
	}

	if _, err := solana.PublicKeyFromBase58(req.DestinationReference); err != nil {
		return nil, fmt.Errorf("invalid destination reference (expected base58 pubkey): %w", err)
	}

	// RPC preflight: validate endpoint reachability and chain liveness.
	rpcClient := rpc.New(a.rpcEndpoint)
	rpcResult, err := rpcClient.GetLatestBlockhash(a.ctx, rpc.CommitmentFinalized)
	if err != nil {
		return nil, fmt.Errorf("failed to validate rpc endpoint for disbursement: %w", err)
	}

	ref := solanaReference(
		"disbursement",
		fmt.Sprintf("%s:%d:%s:%s", req.DestinationReference, req.Amount, req.Currency, rpcResult.Value.Blockhash.String()),
	)

	// TODO: Wire a dedicated on-chain disbursement instruction + signer path once
	// the Solana program exposes a canonical disbursement entrypoint.
	return &SolanaDisbursementResult{
		ProviderReference: ref,
		Status:            entities.SettlementStatusCaptured,
	}, nil
}

func solanaReference(prefix string, seed string) string {
	sum := sha256.Sum256([]byte(prefix + ":" + seed))
	return prefix + "_" + hex.EncodeToString(sum[:8])
}
