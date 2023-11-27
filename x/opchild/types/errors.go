package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/opchild module sentinel errors
var (
	ErrNoValidatorFound                = errorsmod.Register(ModuleName, 2, "validator does not exist")
	ErrValidatorOwnerExists            = errorsmod.Register(ModuleName, 3, "validator already exist for this operator address; must use new validator operator address")
	ErrValidatorPubKeyExists           = errorsmod.Register(ModuleName, 4, "validator already exist for this pubkey; must use new validator pubkey")
	ErrValidatorPubKeyTypeNotSupported = errorsmod.Register(ModuleName, 5, "validator pubkey type is not supported")
	ErrInvalidHistoricalInfo           = errorsmod.Register(ModuleName, 6, "invalid historical info")
	ErrEmptyValidatorPubKey            = errorsmod.Register(ModuleName, 7, "empty validator public key")
	ErrInvalidSigner                   = errorsmod.Register(ModuleName, 8, "expected `opchild` module account as only signer for system message")
	ErrDepositAlreadyFinalized         = errorsmod.Register(ModuleName, 9, "deposit already finalized")
	ErrInvalidAmount                   = errorsmod.Register(ModuleName, 10, "invalid amount")
	ErrInvalidSequence                 = errorsmod.Register(ModuleName, 11, "invalid sequence")
	ErrZeroMaxValidators               = errorsmod.Register(ModuleName, 12, "max validators must be non-zero")
)
