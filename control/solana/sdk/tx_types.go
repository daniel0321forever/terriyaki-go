package sdk

// AccountMeta describes one account passed to an instruction.
type AccountMeta struct {
	Pubkey     [32]byte
	IsSigner   bool
	IsWritable bool
}

// TransactionInstruction represents a single instruction destined for a program.
type TransactionInstruction struct {
	ProgramID [32]byte
	Accounts  []AccountMeta
	Data      []byte
}
