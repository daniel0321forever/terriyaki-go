package solana

import (
	"testing"
)

func sampleKey(seed byte) [32]byte {
	var out [32]byte
	for i := range out {
		out[i] = seed + byte(i)
	}
	return out
}

func TestDerivePledgePDADeterministic(t *testing.T) {
	user := sampleKey(1)
	program := sampleKey(101)

	addr1, bump1, err := DerivePledgePDA("pledge-1", user, program)
	if err != nil {
		t.Fatalf("first derive failed: %v", err)
	}

	addr2, bump2, err := DerivePledgePDA("pledge-1", user, program)
	if err != nil {
		t.Fatalf("second derive failed: %v", err)
	}

	if addr1 != addr2 {
		t.Fatalf("expected same PDA for same inputs")
	}
	if bump1 != bump2 {
		t.Fatalf("expected same bump for same inputs")
	}
}

func TestDerivePledgePDAChangesAcrossSeeds(t *testing.T) {
	user := sampleKey(1)
	program := sampleKey(101)

	baseAddr, _, err := DerivePledgePDA("pledge-1", user, program)
	if err != nil {
		t.Fatalf("derive base failed: %v", err)
	}

	addrByPledge, _, err := DerivePledgePDA("pledge-2", user, program)
	if err != nil {
		t.Fatalf("derive by pledge failed: %v", err)
	}
	if baseAddr == addrByPledge {
		t.Fatalf("expected PDA to change when pledge_id changes")
	}

	addrByUser, _, err := DerivePledgePDA("pledge-1", sampleKey(2), program)
	if err != nil {
		t.Fatalf("derive by user failed: %v", err)
	}
	if baseAddr == addrByUser {
		t.Fatalf("expected PDA to change when user pubkey changes")
	}

	addrByProgram, _, err := DerivePledgePDA("pledge-1", user, sampleKey(102))
	if err != nil {
		t.Fatalf("derive by program failed: %v", err)
	}
	if baseAddr == addrByProgram {
		t.Fatalf("expected PDA to change when program id changes")
	}
}

func TestVerifyPDADerivation(t *testing.T) {
	user := sampleKey(1)
	program := sampleKey(101)

	derived, _, err := DerivePledgePDA("pledge-1", user, program)
	if err != nil {
		t.Fatalf("derive failed: %v", err)
	}

	ok, err := VerifyPDADerivation(derived, "pledge-1", user, program)
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if !ok {
		t.Fatalf("expected verify=true for matching PDA")
	}

	wrong := derived
	wrong[0] ^= 0xFF

	ok, err = VerifyPDADerivation(wrong, "pledge-1", user, program)
	if err != nil {
		t.Fatalf("verify wrong failed: %v", err)
	}
	if ok {
		t.Fatalf("expected verify=false for mismatched PDA")
	}
}

func TestPDATDDParityWithRustFindProgramAddress(t *testing.T) {
	t.Skip("TODO: add golden vectors from Rust derive_pledge_pda/find_program_address parity cases")
}

func TestPDATDDEdgeCases(t *testing.T) {
	t.Skip("TODO: cover empty pledge_id, max-length pledge_id, and non-ASCII pledge_id")
}
