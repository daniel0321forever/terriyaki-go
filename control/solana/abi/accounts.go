package abi

// PledgeState mirrors the on-chain account structure for a pledge.
type PledgeState struct {
    PledgeID     string
    UserPubkey   [32]byte
    OraclePubkey [32]byte
    EscrowAmount uint64
    DeadlineTS   int64
    Status       uint8
}

// ResolutionReceipt mirrors the program's ResolutionReceipt account.
type ResolutionReceipt struct {
    PledgeID    string
    UserSigner  [32]byte
    TxHash      string
    FinalizedAt int64
    ResType     uint8
    ReceiptHash [32]byte
}
