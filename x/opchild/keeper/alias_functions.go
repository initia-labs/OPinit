package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/OPinit/x/opchild/types"
)

// Validator Set

// iterate through the validator set and perform the provided function
func (k Keeper) IterateValidators(ctx context.Context, fn func(validator types.ValidatorI) (stop bool, err error)) error {
	return k.Validators.Walk(ctx, nil, func(_ []byte, validator types.Validator) (stop bool, err error) {
		return fn(validator)
	})
}

// iterate through the active validator set and perform the provided function
func (k Keeper) IterateLastValidators(ctx context.Context, fn func(validator types.ValidatorI, power int64) (stop bool, err error)) error {
	return k.IterateLastValidatorPowers(ctx, func(operator []byte, power int64) (stop bool, err error) {
		validator, found := k.GetValidator(ctx, operator)
		if !found {
			return true, fmt.Errorf("validator record not found for address: %v", sdk.ValAddress(operator))
		}

		return fn(validator, power)
	})
}

// Validator gets the Validator interface for a particular address
func (k Keeper) Validator(ctx context.Context, address sdk.ValAddress) types.ValidatorI {
	val, found := k.GetValidator(ctx, address)
	if !found {
		return nil
	}

	return val
}

// ValidatorByConsAddr gets the validator interface for a particular pubkey
func (k Keeper) ValidatorByConsAddr(ctx context.Context, addr sdk.ConsAddress) types.ValidatorI {
	val, found := k.GetValidatorByConsAddr(ctx, addr)
	if !found {
		return nil
	}

	return val
}

// Delegation Set

// Returns self as it is both a validatorset and delegationset
func (k Keeper) GetValidatorSet() types.ValidatorSet {
	return k
}
