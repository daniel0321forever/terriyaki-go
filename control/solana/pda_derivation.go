package solana

import (
	"crypto/sha256"
	"fmt"

	"filippo.io/edwards25519"
)

// ==============================================================================
// PDA DERIVATION (Mirrors Rust initialize_pledge.rs::derive_pledge_pda())
// ==============================================================================

// computes the Solana Program Derived Address (PDA) that will
// hold the escrow for this pledge. Must be deterministic: same inputs → same address.
//
//	Seeds: ["pledge", user_pubkey_bytes, pledge_id_bytes]
//
// Why this matters:
// - PDA derivation must be identical in Rust and Go, or addresses diverge → funds locked forever
// - Used to pre-compute vault address before transaction submission
// - Used to verify on-chain pledge account is the correct PDA
//
// Arguments:
//
//	pledgeID: the habit commitment ID (e.g., "pledge-1")
//	userPubkey: the [32]byte Pubkey of the user who pledged
//	programID: the [32]byte Pubkey of the Solana program (deployed address)
//
// Returns:
//
//	vaultAddress: [32]byte Pubkey of the PDA (deterministic for this input set)
//	bump: u8 nonce that makes the address valid on-curve (0-255 range)
//	err: non-nil if serialization fails
func DerivePledgePDA(pledgeID string, userPubkey [32]byte, programID [32]byte) ([32]byte, uint8, error) {
	// Step 1: Build the seeds array exactly as Rust does.
	//         Seeds are serialized as [&[u8]; 3] in Solana.
	//         ORDER MATTERS: ["pledge", userPubkey, pledgeID]

	seedPrefix := []byte("pledge") // First seed: literal bytes
	seedUser := userPubkey[:]      // Second seed: pubkey as 32 bytes
	seedPledge := []byte(pledgeID) // Third seed: pledge_id as UTF-8 bytes

	// Step 2: Attempt to find a valid PDA using bump iteration (standard Solana method).
	//         Start from bump=255 (highest nonce) and decrement until address is off-curve.

	for bump := uint8(255); ; bump-- {
		// Construct seeds with the current bump
		seeds := [][]byte{
			seedPrefix,
			seedUser,
			seedPledge,
			{bump}, // Bump added as a single-byte seed
		}

		// Hash the seeds with program ID to compute candidate address
		address, isValid := publishKeyAndBump(seeds, programID)

		if isValid {
			// Found a valid PDA (off-curve)
			return address, bump, nil
		}

		// Decrement bump and try again
		if bump == 0 {
			// Exhausted bump range; should not happen in practice
			return [32]byte{}, 0, fmt.Errorf("failed to find valid PDA for seeds after 256 attempts")
		}
	}
}

// publishKeyAndBump mimics Solana's find_program_address algorithm.
// Hashes (seeds ++ [bump] || program_id) and checks if result is a valid ed25519 point.
// If not on the curve, returns the address and true (PDA found).
// If on the curve, returns zeros and false (invalid bump, try next).
func publishKeyAndBump(seeds [][]byte, programID [32]byte) ([32]byte, bool) {
	// Concatenate all seeds + [bump] + program ID
	h := sha256.New()
	for _, seed := range seeds {
		h.Write(seed)
	}
	h.Write(programID[:])

	result := h.Sum(nil)
	if len(result) != 32 {
		return [32]byte{}, false
	}

	address := *(*[32]byte)(result)
	// Check if address is a valid Curve25519 point (on-curve).
	// Solana uses ed25519; a point is on-curve if it satisfies the curve equation.
	// For PDA purposes, we want a NON-CURVE point.

	_, err := new(edwards25519.Point).SetBytes(address[:])
	if err == nil {
		// Point is on the curve → invalid PDA, try next bump
		return [32]byte{}, false
	}

	// Point is off the curve → valid PDA
	return address, true
}

// VerifyPDADerivation checks that a claimed PDA address is correct for the given seeds.
// Used to verify on-chain state before building transactions.
//
// Rust equivalent: This is not explicitly in the Rust code, but is the logical check
// that the backend should perform before trusting an on-chain pledge account.
//
// Arguments:
//
//	claimedAddress: the address reported by on-chain RPC getAccount()
//	pledgeID: the pledge ID
//	userPubkey: the user's pubkey
//	programID: the program's pubkey
//
// Returns:
//
//	true if claimed address is the correct PDA for these seeds
//	false if there's a mismatch (potential fraud or bug)
func VerifyPDADerivation(claimedAddress [32]byte, pledgeID string, userPubkey [32]byte, programID [32]byte) (bool, error) {
	derivedAddress, _, err := DerivePledgePDA(pledgeID, userPubkey, programID)
	if err != nil {
		return false, err
	}

	return claimedAddress == derivedAddress, nil
}

// ==============================================================================
// SIGNER SEEDS FOR CPI (Mirrors Rust derive_pledge_signer_seeds logic)
// ==============================================================================

// Uncomment when claim_timeout CPI transfer implementation needs signer seeds:
//
// KEY FUNCTION ⭐ (If used in timeout path)
// DerivePledgeSignerSeeds returns the seeds array needed to sign CPI calls
// on behalf of the PDA when transferring escrow lamports.
//
// In the timeout path, the program signs a CPI system-program transfer with:
//   Signer: PDA (derived from ["pledge", user_pubkey, pledge_id])
//   Seeds: [["pledge", user_pubkey, pledge_id, bump]]
//
// The Go backend does NOT sign this; the ON-CHAIN program signs it.
// But Go computes the seeds upfront to verify they match on-chain.
//
// func DerivePledgeSignerSeeds(pledgeID string, userPubkey [32]byte, bump uint8) [][]byte {
//     return [][]byte{
//         []byte("pledge"),
//         userPubkey[:],
//         []byte(pledgeID),
//         {bump},
//     }
// }
