package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/ophost module sentinel errors
var (
	ErrInvalidBridgeId            = errorsmod.Register(ModuleName, 2, "invalid bridge id")
	ErrInvalidHashLength          = errorsmod.Register(ModuleName, 3, "invalid hash length")
	ErrInvalidOutputIndex         = errorsmod.Register(ModuleName, 4, "invalid output index")
	ErrInvalidAmount              = errorsmod.Register(ModuleName, 5, "invalid bridge amount")
	ErrInvalidSequence            = errorsmod.Register(ModuleName, 6, "invalid sequence")
	ErrInvalidL2BlockNumber       = errorsmod.Register(ModuleName, 7, "invalid l2 block number")
	ErrNotFinalized               = errorsmod.Register(ModuleName, 8, "output has not finalized")
	ErrAlreadyFinalized           = errorsmod.Register(ModuleName, 9, "output has already finalized")
	ErrFailedToVerifyWithdrawal   = errorsmod.Register(ModuleName, 10, "failed to verify withdrawal tx")
	ErrWithdrawalAlreadyFinalized = errorsmod.Register(ModuleName, 11, "withdrawal already finalized")
	ErrEmptyBatchInfo             = errorsmod.Register(ModuleName, 12, "empty batch info")
	ErrInvalidBridgeMetadata      = errorsmod.Register(ModuleName, 13, "invalid bridge metadata")
	ErrInvalidBatchInfo           = errorsmod.Register(ModuleName, 14, "invalid batch info")
	ErrInvalidChallengerUpdate    = errorsmod.Register(ModuleName, 15, "invalid challenger update")
	ErrEmptyBatchBytes            = errorsmod.Register(ModuleName, 16, "empty batch bytes")
)
