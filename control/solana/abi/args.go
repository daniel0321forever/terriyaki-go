package abi

// InitializePledgeArgs models the Rust InitializePledgeArgs struct.
type InitializePledgeArgs struct {
	PledgeID          string   // UTF-8 string, max 32 bytes
	OraclePubkey      [32]byte // Pubkey: oracle authority
	EscrowAmount      uint64   // lamports to lock in PDA
	DeadlineTimestamp int64    // unix seconds
}

// ResolveSuccessArgs models the Rust ResolveSuccessArgs struct.
type ResolveSuccessArgs struct {
	TxHash      string
	FinalizedAt int64
}

// ResolveFailureArgs models the Rust ResolveFailureArgs struct.
type ResolveFailureArgs struct {
	TxHash      string
	FinalizedAt int64
}

// ClaimTimeoutArgs models the Rust ClaimTimeoutArgs struct.
type ClaimTimeoutArgs struct {
	TxHash      string
	FinalizedAt int64
}
