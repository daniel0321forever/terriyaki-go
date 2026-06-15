package solana

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"time"

	"github.com/daniel0321forever/terriyaki-go/control/solana/abi"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// ==============================================================================
// TYPE DEFINITIONS
// ==============================================================================

// SignedTransaction represents a signed and ready-to-broadcast Solana transaction.
type SignedTransaction struct {
	Bytes     []byte   // Full signed transaction bytes (ready for RPC broadcast)
	Signature [64]byte // Transaction signature (for audit trail and tracking)
}

// ==============================================================================
// TRANSACTION SIGNER (Sign and submit pledges to the blockchain)
// ==============================================================================

// SignResolveSuccessTransaction builds, signs, and prepares for submission
// the resolve_success transaction.
//
// Called by the backend when the oracle confirms the user completed the habit.
// ORACLE SIGNER: This transaction is signed by the oracle keypair (not the user).
// Oracle authority is embedded in the program's initialization and checked on-chain.
//
// Arguments:
//
//	oraclePubkey: [32]byte oracle pubkey
//	oraclePrivateKey: [64]byte oracle keypair
//	pledgePDA: [32]byte pledge account address (must exist on-chain)
//	userAccount: [32]byte user's wallet account (destination for escrow)
//	txHash: off-chain transaction ID for audit trail
//	finalizedAt: unix seconds timestamp
//	systemProgramID: Solana system program
//	habitProgramID: Habitat program
//
// Returns:
//
//	SignedTransaction with bytes and signature ready to broadcast
//	err: non-nil if signing fails
func SignResolveSuccessTransaction(
	recentBlockhash solana.Hash,
	oraclePubkey [32]byte,
	oraclePrivateKey [64]byte,
	pledgePDA [32]byte,
	userAccount [32]byte,
	penaltyPool [32]byte,
	txHash string,
	finalizedAt int64,
	systemProgramID [32]byte,
	habitProgramID [32]byte,
) (SignedTransaction, error) {
	// Step 1: Build the instruction.
	args := abi.ResolveSuccessArgs{
		TxHash:      txHash,
		FinalizedAt: finalizedAt,
	}

	instr, err := ResolveSuccessInstruction(
		args,
		oraclePubkey,
		pledgePDA,
		userAccount,
		penaltyPool,
		systemProgramID,
		habitProgramID,
	)
	if err != nil {
		return SignedTransaction{}, fmt.Errorf("build instruction: %w", err)
	}

	// Step 2: Build transaction message with oracle as signer and feepayer.
	//         Account order: [oracle, pledge, user, system_program, penalty_pool]
	accountMetaSlice := solana.AccountMetaSlice{}
	for _, meta := range instr.Accounts {
		accountMetaSlice = append(accountMetaSlice, &solana.AccountMeta{
			PublicKey:  solana.PublicKeyFromBytes(meta.Pubkey[:]),
			IsSigner:   meta.IsSigner,
			IsWritable: meta.IsWritable,
		})
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			solana.NewInstruction(
				solana.PublicKeyFromBytes(instr.ProgramID[:]),
				accountMetaSlice,
				instr.Data,
			),
		},
		recentBlockhash,
	)
	if err != nil {
		return SignedTransaction{}, fmt.Errorf("create transaction: %w", err)
	}

	// Step 3: Sign with oracle private key.
	oraclePrivKey := ed25519.PrivateKey(oraclePrivateKey[:])
	messageBytes, err := tx.Message.MarshalBinary()
	if err != nil {
		return SignedTransaction{}, fmt.Errorf("marshal message: %w", err)
	}

	oracleSig := ed25519.Sign(oraclePrivKey, messageBytes)
	tx.Signatures = []solana.Signature{solana.SignatureFromBytes(oracleSig)}

	// Step 4: Serialize the signed transaction to bytes for RPC submission.
	signedTxBytes, err := tx.MarshalBinary()
	if err != nil {
		return SignedTransaction{}, fmt.Errorf("marshal transaction: %w", err)
	}

	return SignedTransaction{
		Bytes:     signedTxBytes,
		Signature: solana.SignatureFromBytes(oracleSig),
	}, nil
}

// KEY FUNCTION ⭐ (Critical to project)
// SignResolveFailureTransaction builds, signs, and prepares for submission
// the resolve_failure transaction.
//
// Called by the oracle when the user failed to complete the habit.
// ORACLE SIGNER: Signed by oracle keypair (same as resolve_success path).
// Destination is penalty pool (not the user's wallet).
//
// Arguments:
//
//	recentBlockhash: recent blockhash from RPC
//	oraclePubkey: [32]byte oracle pubkey
//	oraclePrivateKey: [64]byte oracle keypair
//	pledgePDA: [32]byte pledge account address
//	userAccount: [32]byte user wallet address
//	penaltyPool: [32]byte destination for failed pledge escrow
//	txHash: off-chain transaction ID
//	finalizedAt: unix seconds
//	systemProgramID: Solana system program
//	habitProgramID: Habitat program
//
// Returns:
//
//	SignedTransaction with bytes and signature ready to broadcast
//	err: non-nil if signing fails
func SignResolveFailureTransaction(
	recentBlockhash solana.Hash,
	oraclePubkey [32]byte,
	oraclePrivateKey [64]byte,
	pledgePDA [32]byte,
	userAccount [32]byte,
	penaltyPool [32]byte,
	txHash string,
	finalizedAt int64,
	systemProgramID [32]byte,
	habitProgramID [32]byte,
) (SignedTransaction, error) {
	// Step 1: Build the instruction.
	args := abi.ResolveFailureArgs{
		TxHash:      txHash,
		FinalizedAt: finalizedAt,
	}

	instr, err := ResolveFailureInstruction(
		args,
		oraclePubkey,
		pledgePDA,
		userAccount,
		penaltyPool,
		systemProgramID,
		habitProgramID,
	)
	if err != nil {
		return SignedTransaction{}, fmt.Errorf("build instruction: %w", err)
	}

	// Step 2: Build transaction message with oracle as signer and feepayer.
	//         Account order: [oracle, pledge, user, system_program, penalty_pool]
	accountMetaSlice := solana.AccountMetaSlice{}
	for _, meta := range instr.Accounts {
		accountMetaSlice = append(accountMetaSlice, &solana.AccountMeta{
			PublicKey:  solana.PublicKeyFromBytes(meta.Pubkey[:]),
			IsSigner:   meta.IsSigner,
			IsWritable: meta.IsWritable,
		})
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			solana.NewInstruction(
				solana.PublicKeyFromBytes(instr.ProgramID[:]),
				accountMetaSlice,
				instr.Data,
			),
		},
		recentBlockhash,
	)
	if err != nil {
		return SignedTransaction{}, fmt.Errorf("create transaction: %w", err)
	}

	// Step 3: Sign with oracle private key.
	oraclePrivKey := ed25519.PrivateKey(oraclePrivateKey[:])
	messageBytes, err := tx.Message.MarshalBinary()
	if err != nil {
		return SignedTransaction{}, fmt.Errorf("marshal message: %w", err)
	}

	oracleSig := ed25519.Sign(oraclePrivKey, messageBytes)
	tx.Signatures = []solana.Signature{solana.SignatureFromBytes(oracleSig)}

	// Step 4: Serialize the signed transaction to bytes for RPC submission.
	signedTxBytes, err := tx.MarshalBinary()
	if err != nil {
		return SignedTransaction{}, fmt.Errorf("marshal transaction: %w", err)
	}

	return SignedTransaction{
		Bytes:     signedTxBytes,
		Signature: solana.SignatureFromBytes(oracleSig),
	}, nil
}

// KEY FUNCTION ⭐ (Critical to project)
// SignClaimTimeoutTransaction builds, signs, and prepares for submission
// the claim_timeout transaction.
//
// Called by the USER when the oracle backend failed to resolve the pledge
// within the grace period (deadline + 7 days). User can self-refund.
// USER SIGNER: Signed by the user's private key (same as initialize_pledge).
//
// Arguments:
//
//	userPubkey: [32]byte user pubkey
//	userPrivateKey: [64]byte user keypair
//	pledgePDA: [32]byte pledge account address
//	txHash: off-chain transaction ID
//	finalizedAt: unix seconds (current time when user submits)
//	systemProgramID: Solana system program
//	habitProgramID: Habitat program
//
// Returns:
//
//	SignedTransaction with bytes and signature ready to broadcast
//	err: non-nil if signing fails
func SignClaimTimeoutTransaction(
	userPubkey [32]byte,
	userPrivateKey [64]byte,
	pledgePDA [32]byte,
	txHash string,
	finalizedAt int64,
	systemProgramID [32]byte,
	habitProgramID [32]byte,
) (SignedTransaction, error) {
	// Step 1: Build the instruction.
	args := abi.ClaimTimeoutArgs{
		TxHash:      txHash,
		FinalizedAt: finalizedAt,
	}

	instr, err := ClaimTimeoutInstruction(
		args,
		userPubkey,
		pledgePDA,
		systemProgramID,
		habitProgramID,
	)
	if err != nil {
		return SignedTransaction{}, fmt.Errorf("build instruction: %w", err)
	}

	// Step 2: Build transaction message with user as signer and feepayer.
	//         Account order: [user, pledge, system_program]
	var recentBlockhash solana.Hash // TODO: fetch from RPC

	accountMetaSlice := solana.AccountMetaSlice{}
	for _, meta := range instr.Accounts {
		accountMetaSlice = append(accountMetaSlice, &solana.AccountMeta{
			PublicKey:  solana.PublicKeyFromBytes(meta.Pubkey[:]),
			IsSigner:   meta.IsSigner,
			IsWritable: meta.IsWritable,
		})
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			solana.NewInstruction(
				solana.PublicKeyFromBytes(instr.ProgramID[:]),
				accountMetaSlice,
				instr.Data,
			),
		},
		recentBlockhash,
	)
	if err != nil {
		return SignedTransaction{}, fmt.Errorf("create transaction: %w", err)
	}

	// Step 3: Sign with user's private key.
	userPrivKey := ed25519.PrivateKey(userPrivateKey[:])
	messageBytes, err := tx.Message.MarshalBinary()
	if err != nil {
		return SignedTransaction{}, fmt.Errorf("marshal message: %w", err)
	}

	userSig := ed25519.Sign(userPrivKey, messageBytes)
	tx.Signatures = []solana.Signature{solana.SignatureFromBytes(userSig)}

	// Step 4: Serialize the signed transaction to bytes for RPC submission.
	signedTxBytes, err := tx.MarshalBinary()
	if err != nil {
		return SignedTransaction{}, fmt.Errorf("marshal transaction: %w", err)
	}

	return SignedTransaction{
		Bytes:     signedTxBytes,
		Signature: solana.SignatureFromBytes(userSig),
	}, nil
}

// ==============================================================================
// TRANSACTION SUBMISSION HELPERS
// ==============================================================================

// SubmitTransactionWithRetry sends a signed transaction to Solana RPC and polls
// for confirmation.
//
// Workflow:
// 1. Decode the signed transaction to extract the signature
// 2. Send transaction to RPC endpoint via SendTransaction RPC method
// 3. Poll GetSignatureStatuses() until confirmed (finalized)
// 4. Return transaction signature as proof of submission
//
// Arguments:
//
//	rpcEndpoint: Solana RPC URL (e.g., 'https://api.devnet.solana.com')
//	signedTx: SignedTransaction with bytes and signature
//	maxRetries: how many times to retry on transient failures
//
// Returns:
//
//	signature: [64]byte transaction signature (proof of submission)
//	err: non-nil if all retries exhausted or RPC error
func SubmitTransactionWithRetry(
	rpcEndpoint string,
	signedTx SignedTransaction,
	maxRetries int,
) ([64]byte, error) {
	// Create RPC client.
	client := rpc.New(rpcEndpoint)

	// Parse the signed transaction to verify structure.
	tx, err := solana.TransactionFromBytes(signedTx.Bytes)
	if err != nil {
		return [64]byte{}, fmt.Errorf("parse transaction: %w", err)
	}

	// Use the signature from the first signer (should be the only one).
	if len(tx.Signatures) == 0 {
		return [64]byte{}, fmt.Errorf("transaction has no signatures")
	}
	sig := tx.Signatures[0]

	// Retry loop for sending transaction.
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		// Send transaction to RPC.
		_, err := client.SendTransactionWithOpts(
			ctx,
			tx,
			rpc.TransactionOpts{
				SkipPreflight:       false,
				PreflightCommitment: rpc.CommitmentFinalized,
			},
		)
		cancel()

		if err == nil {
			// Transaction sent successfully, now poll for confirmation.
			break
		}

		lastErr = err

		// Check if error is transient (can retry).
		// For now, we retry on all errors up to maxRetries.
		if attempt < maxRetries-1 {
			// Exponential backoff: 500ms, 1s, 2s, 4s, etc.
			backoff := time.Duration(1<<uint(attempt)) * 500 * time.Millisecond
			time.Sleep(backoff)
			continue
		}
	}

	if lastErr != nil {
		return [64]byte{}, fmt.Errorf("send transaction (after %d attempts): %w", maxRetries, lastErr)
	}

	// Poll for confirmation (finalized status).
	pollTimeout := time.After(2 * time.Minute)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-pollTimeout:
			return [64]byte{}, fmt.Errorf("confirmation timeout: signature %s not finalized within 2 minutes", sig.String())

		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

			// Check signature status.
			statuses, err := client.GetSignatureStatuses(ctx, true, sig)
			cancel()

			if err != nil {
				// Log error but continue polling.
				continue
			}

			if len(statuses.Value) == 0 {
				// Signature not yet processed.
				continue
			}

			status := statuses.Value[0]
			if status == nil {
				// Still processing.
				continue
			}

			// Check confirmation status.
			if status.ConfirmationStatus == rpc.ConfirmationStatusFinalized {
				return sig, nil // Success!
			}

			if status.Err != nil {
				// Transaction failed on-chain.
				return [64]byte{}, fmt.Errorf("transaction failed on-chain: %v", status.Err)
			}
		}
	}
}
