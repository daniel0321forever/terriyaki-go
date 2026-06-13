package solana

import (
	"fmt"

	solanaGo "github.com/gagliardetto/solana-go"
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
//
// func DerivePledgePDA(pledgeID string, userPubkey [32]byte, programID [32]byte) ([32]byte, uint8, error)
func DerivePledgePDA(pledgeID string, userPubkey [32]byte, programID [32]byte) ([32]byte, uint8, error) {
	// Build the seeds array exactly as Rust/Solana expects.
	// Seeds: ["pledge", userPubkey, pledgeID]
	seedPrefix := []byte("pledge")
	seedUser := userPubkey[:]
	seedPledge := []byte(pledgeID)

	seeds := [][]byte{
		seedPrefix,
		seedUser,
		seedPledge,
	}

	pda, bump, err := solanaGo.FindProgramAddress(seeds, solanaGo.PublicKey(programID))
	if err != nil {
		return [32]byte{}, 0, fmt.Errorf("failed to find valid PDA: %w", err)
	}

	return [32]byte(pda), bump, nil
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
