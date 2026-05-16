package abi

// ProgramErrorCode maps Rust error code integers to human-readable messages.
type ProgramErrorCode uint32

const (
    UnauthorizedUser ProgramErrorCode = 6000
    InvalidInstruction ProgramErrorCode = 6001
    AlreadyResolved ProgramErrorCode = 6002
    TimeoutNotReached ProgramErrorCode = 6003
    InsufficientFunds ProgramErrorCode = 6004
)

func (code ProgramErrorCode) ErrorMessage() string {
    switch code {
    case UnauthorizedUser:
        return "only the pledge owner or oracle can resolve this pledge"
    case InvalidInstruction:
        return "invalid instruction arguments (empty pledge_id, zero escrow, past deadline)"
    case AlreadyResolved:
        return "pledge has already been resolved; idempotent rejection"
    case TimeoutNotReached:
        return "grace period has not elapsed; timeout claim not eligible yet"
    case InsufficientFunds:
        return "source account does not have enough lamports to fund escrow"
    default:
        return "unknown program error code"
    }
}
