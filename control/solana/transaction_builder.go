package solana

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/daniel0321forever/terriyaki-go/control/solana/abi"
	solanaGo "github.com/gagliardetto/solana-go"
)

// UnsignedTxEnvelope is a convenience representation of an unsigned intent
// that the client wallet can consume and sign. This is intentionally a
// high-level placeholder while binary tx construction is finalized.
type UnsignedTxEnvelope struct {
	ProgramID       string   `json:"program_id"`
	Accounts        []string `json:"accounts"`
	InstructionB64  string   `json:"instruction_b64"`
	RecentBlockhash string   `json:"recent_blockhash"`
}

// BuildInitializePledgeUnsignedTx builds an unsigned transaction envelope for
// the initialize_pledge instruction. It returns a JSON-encoded transaction and
// the derived pledge PDA. This is a placeholder representation used for the
// non-custodial client-sign flow; later we will convert this into a proper
// unsigned Transaction/message that wallets can sign.
func BuildInitializePledgeUnsignedTx(
	recentBlockhash solanaGo.Hash,
	userPubkey [32]byte,
	pledgeID string,
	oraclePubkey [32]byte,
	escrowAmount uint64,
	deadlineTS int64,
	systemProgramID [32]byte,
	habitProgramID [32]byte,
) ([]byte, [32]byte, error) {
	// derive PDA
	pledgePDA, _, err := DerivePledgePDA(pledgeID, userPubkey, habitProgramID)
	if err != nil {
		return nil, [32]byte{}, fmt.Errorf("derive PDA: %w", err)
	}

	// build instruction bytes using existing builder
	args := abi.InitializePledgeArgs{
		PledgeID:          pledgeID,
		OraclePubkey:      oraclePubkey,
		EscrowAmount:      escrowAmount,
		DeadlineTimestamp: deadlineTS,
	}

	instr, err := InitializePledgeInstruction(
		args,
		userPubkey,
		systemProgramID,
		pledgePDA,
		habitProgramID,
	)
	if err != nil {
		return nil, pledgePDA, fmt.Errorf("build instruction: %w", err)
	}

	// prepare account list as base58 strings
	accounts := make([]string, 0, len(instr.Accounts))
	for _, a := range instr.Accounts {
		accounts = append(accounts, solanaGo.PublicKeyFromBytes(a.Pubkey[:]).String())
	}

	env := UnsignedTxEnvelope{
		ProgramID:       solanaGo.PublicKeyFromBytes(instr.ProgramID[:]).String(),
		Accounts:        accounts,
		InstructionB64:  base64.StdEncoding.EncodeToString(instr.Data),
		RecentBlockhash: recentBlockhash.String(),
	}

	unsignedTxJSON, err := json.Marshal(env)
	if err != nil {
		return nil, pledgePDA, fmt.Errorf("marshal unsigned tx: %w", err)
	}

	return unsignedTxJSON, pledgePDA, nil
}
