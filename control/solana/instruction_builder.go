package solana

import (
	"crypto/sha256"
	"fmt"

	"github.com/near/borsh-go"

	"github.com/daniel0321forever/terriyaki-go/control/solana/abi"
	"github.com/daniel0321forever/terriyaki-go/control/solana/sdk"
)

// ==============================================================================
// INSTRUCTION BUILDER (Converts Go intent → on-chain instruction bytes)
// ==============================================================================

// InitializePledgeInstruction builds the on-chain instruction bytecode for
// the initialize_pledge function on the Solana program.
//
// This is where user intent (from the backend's payment flow) becomes wire format.
//
// Arguments:
//
//	args: InitializePledgeArgs containing pledge_id, oracle_pubkey, amount, deadline
//	userPubkey: the signer (user creating the pledge)
//	systemProgram: the system program address
//	pledgePDA: the vault address where lamports will be locked
//	programID: the Solana program's address
//
// Returns:
//
//	instruction bytes (Anchor discriminator + serialized args)
//	err if serialization fails
func InitializePledgeInstruction(
	args abi.InitializePledgeArgs,
	userPubkey [32]byte,
	systemProgram [32]byte,
	pledgePDA [32]byte,
	programID [32]byte,
) (*sdk.TransactionInstruction, error) {
	// Step 1: Construct the instruction discriminator.
	//         Anchor uses SHA256("global:initialize_pledge")[:8] as the discriminator.
	//         This tells the on-chain program which instruction is being called.
	discriminator := AnchorDiscriminator("global:initialize_pledge")

	// Step 2: Serialize instruction arguments in Anchor's AnchorSerialize format.
	//         Order and encoding MUST match the Rust struct exactly.
	argBytes, err := serializeInitializePledgeArgs(&args)
	if err != nil {
		return nil, fmt.Errorf("serialize args: %w", err)
	}

	// Step 3: Combine discriminator + serialized args to form the instruction data.
	instruction := make([]byte, 0, len(discriminator)+len(argBytes))
	instruction = append(instruction, discriminator...)
	instruction = append(instruction, argBytes...)

	// Step 4: Build account metadata (accounts array).
	//         The on-chain program expects accounts in this order:
	//         [0] user (signer, writable)
	//         [1] pledge (writable, PDA)
	//         [2] system_program (read-only)
	// This is defined by the Anchor #[derive(Accounts)] in Rust.
	//
	accounts := []sdk.AccountMeta{
		{Pubkey: userPubkey, IsSigner: true, IsWritable: true},
		{Pubkey: pledgePDA, IsSigner: false, IsWritable: true},
		{Pubkey: systemProgram, IsSigner: false, IsWritable: false},
	}
	returnedInstruction := &sdk.TransactionInstruction{
		ProgramID: programID,
		Accounts:  accounts,
		Data:      instruction,
	}

	return returnedInstruction, nil
}

// ResolveSuccessInstruction builds the on-chain instruction bytecode for
// the resolve_success function.
//
// Called by the oracle when the user successfully completed the habit.
// Transfers escrow from PDA back to user account.
//
// Arguments:
//
//	args: ResolveSuccessArgs with tx_hash and finalized_at
//	oraclePubkey: oracle signer (authorization check on-chain)
//	pledgePDA: the PDA holding escrow
//	userAccount: destination for escrow return
//	systemProgram: system program for CPI transfer
//	programID: the Solana program's address
//
// Returns:
//
//	instruction bytes
//	err if serialization fails
func ResolveSuccessInstruction(
	args abi.ResolveSuccessArgs,
	oraclePubkey [32]byte,
	pledgePDA [32]byte,
	userAccount [32]byte,
	systemProgram [32]byte,
	programID [32]byte,
) (*sdk.TransactionInstruction, error) {
	// Step 1: Construct discriminator for resolve_success.
	discriminator := AnchorDiscriminator("global:resolve_success")

	// Step 2: Serialize resolve_success args.
	argBytes, err := serializeResolveSuccessArgs(&args)
	if err != nil {
		return nil, fmt.Errorf("serialize args: %w", err)
	}

	// Step 3: Combine discriminator + args.
	instruction := make([]byte, 0, len(discriminator)+len(argBytes))
	instruction = append(instruction, discriminator...)
	instruction = append(instruction, argBytes...)

	// Step 4: Account list.
	//         [0] oracle (signer, read-only)
	//         [1] pledge (writable)
	//         [2] user (writable)
	//         [3] system_program (read-only)
	accounts := []sdk.AccountMeta{
		{Pubkey: oraclePubkey, IsSigner: true, IsWritable: false},
		{Pubkey: pledgePDA, IsSigner: false, IsWritable: true},
		{Pubkey: userAccount, IsSigner: false, IsWritable: true},
		{Pubkey: systemProgram, IsSigner: false, IsWritable: false},
	}

	returned := &sdk.TransactionInstruction{
		ProgramID: programID,
		Accounts:  accounts,
		Data:      instruction,
	}

	return returned, nil
}

// ResolveFailureInstruction builds the on-chain instruction bytecode for
// the resolve_failure function.
//
// Called by the oracle when the user failed to complete the habit.
// Transfers escrow from PDA to penalty pool (not back to user).
//
// Arguments:
//
//	args: ResolveFailureArgs with tx_hash and finalized_at
//	oraclePubkey: oracle signer
//	pledgePDA: the PDA holding escrow
//	penaltyPool: destination for penalty transfer
//	systemProgram: system program for CPI transfer
//	programID: the Solana program's address
//
// Returns:
//
//	instruction bytes
//	err if serialization fails
func ResolveFailureInstruction(
	args abi.ResolveFailureArgs,
	oraclePubkey [32]byte,
	pledgePDA [32]byte,
	penaltyPool [32]byte,
	systemProgram [32]byte,
	programID [32]byte,
) (*sdk.TransactionInstruction, error) {
	discriminator := AnchorDiscriminator("global:resolve_failure")

	argBytes, err := serializeResolveFailureArgs(&args)
	if err != nil {
		return nil, fmt.Errorf("serialize args: %w", err)
	}

	instruction := make([]byte, 0, len(discriminator)+len(argBytes))
	instruction = append(instruction, discriminator...)
	instruction = append(instruction, argBytes...)

	accounts := []sdk.AccountMeta{
		{Pubkey: oraclePubkey, IsSigner: true, IsWritable: false},
		{Pubkey: pledgePDA, IsSigner: false, IsWritable: true},
		{Pubkey: penaltyPool, IsSigner: false, IsWritable: true},
		{Pubkey: systemProgram, IsSigner: false, IsWritable: false},
	}

	returned := &sdk.TransactionInstruction{
		ProgramID: programID,
		Accounts:  accounts,
		Data:      instruction,
	}

	return returned, nil
}

// ClaimTimeoutInstruction builds the on-chain instruction bytecode for
// the claim_timeout function.
//
// Called by the user if the oracle backend fails to resolve within grace period.
// User reclaims their funds as a fallback.
// No oracle signature needed; user signature+deadline proof is sufficient.
//
// Arguments:
//
//	args: ClaimTimeoutArgs
//	userPubkey: user signer (authorization check on-chain)
//	pledgePDA: the PDA holding escrow
//	systemProgram: system program for CPI transfer
//	programID: the Solana program's address
//
// Returns:
//
//	instruction bytes
//	err if serialization fails
func ClaimTimeoutInstruction(
	args abi.ClaimTimeoutArgs,
	userPubkey [32]byte,
	pledgePDA [32]byte,
	systemProgram [32]byte,
	programID [32]byte,
) (*sdk.TransactionInstruction, error) {
	discriminator := AnchorDiscriminator("global:claim_timeout")

	argBytes, err := serializeClaimTimeoutArgs(&args)
	if err != nil {
		return nil, fmt.Errorf("serialize args: %w", err)
	}

	instruction := make([]byte, 0, len(discriminator)+len(argBytes))
	instruction = append(instruction, discriminator...)
	instruction = append(instruction, argBytes...)

	accounts := []sdk.AccountMeta{
		{Pubkey: userPubkey, IsSigner: true, IsWritable: true},
		{Pubkey: pledgePDA, IsSigner: false, IsWritable: true},
		{Pubkey: systemProgram, IsSigner: false, IsWritable: false},
	}

	returned := &sdk.TransactionInstruction{
		ProgramID: programID,
		Accounts:  accounts,
		Data:      instruction,
	}

	return returned, nil
}

// ==============================================================================
// SERIALIZATION HELPERS (Mirrors Anchor AnchorSerialize format)
// ==============================================================================

// AnchorDiscriminator computes the 8-byte instruction discriminator.
// Anchor uses: SHA256(namespace + ":" + instruction_name)[:8]
// For global instruction 'foo': SHA256("global:foo")[:8]
func AnchorDiscriminator(discriminatorName string) []byte {
	hash := sha256.New()
	hash.Write([]byte(discriminatorName))
	return hash.Sum(nil)[:8]
}

// serializeInitializePledgeArgs: Write args in Anchor's byte order.
// Used by InitializePledgeInstruction.
//
// Field order MUST match Rust struct InitializePledgeArgs:
//
//	pledge_id: String (Borsh: 4-byte len + UTF-8 bytes)
//	oracle_pubkey: Pubkey ([32]byte)
//	escrow_amount: u64 (little-endian)
//	deadline_timestamp: i64 (little-endian)
func serializeInitializePledgeArgs(args *abi.InitializePledgeArgs) ([]byte, error) {
	serielized, err := borsh.Serialize(args)
	if err != nil {
		return nil, fmt.Errorf("borsh serialize: %w", err)
	}
	return serielized, nil
}

func serializeResolveSuccessArgs(args *abi.ResolveSuccessArgs) ([]byte, error) {
	// Fields: tx_hash (String), finalized_at (i64)
	serielized, err := borsh.Serialize(args)
	if err != nil {
		return nil, fmt.Errorf("borsh serialize: %w", err)
	}
	return serielized, nil
}

func serializeResolveFailureArgs(args *abi.ResolveFailureArgs) ([]byte, error) {
	// Fields: tx_hash (String), finalized_at (i64)
	serielized, err := borsh.Serialize(args)
	if err != nil {
		return nil, fmt.Errorf("borsh serialize: %w", err)
	}
	return serielized, nil
}

func serializeClaimTimeoutArgs(args *abi.ClaimTimeoutArgs) ([]byte, error) {
	// Fields: tx_hash (String), finalized_at (i64)
	serielized, err := borsh.Serialize(args)
	if err != nil {
		return nil, fmt.Errorf("borsh serialize: %w", err)
	}
	return serielized, nil
}
