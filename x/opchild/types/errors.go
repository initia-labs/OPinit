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
	ErrInvalidExecuteMsg               = errorsmod.Register(ModuleName, 13, "invalid execute message")
	ErrUnroutableExecuteMsg            = errorsmod.Register(ModuleName, 14, "unroutable execute message")
	ErrInvalidExecutorChangePlan       = errorsmod.Register(ModuleName, 15, "invalid executor chane plan")
	ErrAlreadyRegisteredHeight         = errorsmod.Register(ModuleName, 16, "executor change plan already exists at the height")
	ErrInvalidBridgeInfo               = errorsmod.Register(ModuleName, 17, "invalid bridge info")
	ErrInvalidHeight                   = errorsmod.Register(ModuleName, 18, "invalid oracle height")
	ErrInvalidPrices                   = errorsmod.Register(ModuleName, 19, "invalid oracle prices")
	ErrMaxValidatorsExceeded           = errorsmod.Register(ModuleName, 20, "max validators exceeded")
	ErrMaxValidatorsLowerThanCurrent   = errorsmod.Register(ModuleName, 21, "max validators cannot be lower than current number of validators")
)
