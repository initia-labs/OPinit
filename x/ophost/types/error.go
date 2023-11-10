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
	ErrSubmissionInterval         = errorsmod.Register(ModuleName, 7, "submission interval has not passed")
	ErrNotFinalized               = errorsmod.Register(ModuleName, 8, "output has not finalized")
	ErrFailedToVerifyWithdrawal   = errorsmod.Register(ModuleName, 9, "failed to verify withdrawal tx")
	ErrWithdrawalAlreadyFinalized = errorsmod.Register(ModuleName, 10, "withdrawal already finalized")
)
