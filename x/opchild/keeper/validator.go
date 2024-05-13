package keeper

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/initia-labs/OPinit/x/opchild/types"
)

// get a single validator
func (k Keeper) GetValidator(ctx context.Context, addr sdk.ValAddress) (validator types.Validator, found bool) {
	validator, err := k.Validators.Get(ctx, addr)
	if errors.Is(err, collections.ErrNotFound) {
		return validator, false
	}

	return validator, true
}

func (k Keeper) mustGetValidator(ctx context.Context, addr sdk.ValAddress) types.Validator {
	validator, found := k.GetValidator(ctx, addr)
	if !found {
		panic(fmt.Sprintf("validator record not found for address: %X\n", addr))
	}

	return validator
}

// get a single validator by consensus address
func (k Keeper) GetValidatorByConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (validator types.Validator, found bool) {
	opAddr, err := k.ValidatorsByConsAddr.Get(ctx, consAddr)
	if errors.Is(err, collections.ErrNotFound) {
		return validator, false
	}

	return k.GetValidator(ctx, opAddr)
}

// set the main record holding validator details
func (k Keeper) SetValidator(ctx context.Context, validator types.Validator) error {
	valAddr, err := k.validatorAddressCodec.StringToBytes(validator.GetOperator())
	if err != nil {
		return err
	}

	return k.Validators.Set(ctx, valAddr, validator)
}

// validator index
func (k Keeper) SetValidatorByConsAddr(ctx context.Context, validator types.Validator) error {
	consPk, err := validator.GetConsAddr()
	if err != nil {
		return err
	}

	valAddr, err := k.validatorAddressCodec.StringToBytes(validator.GetOperator())
	if err != nil {
		return err
	}

	return k.ValidatorsByConsAddr.Set(ctx, consPk, valAddr)
}

// remove the validator record and associated indexes
// except for the bonded validator index which is only handled in ApplyAndReturnTendermintUpdates
func (k Keeper) RemoveValidator(ctx context.Context, address sdk.ValAddress) error {
	// first retrieve the old validator record
	validator, found := k.GetValidator(ctx, address)
	if !found {
		return nil
	}

	valConsAddr, err := validator.GetConsAddr()
	if err != nil {
		return err
	}

	if err := k.Validators.Remove(ctx, address); err != nil {
		return err
	}
	if err := k.ValidatorsByConsAddr.Remove(ctx, valConsAddr); err != nil {
		return err
	}

	return nil
}

// get groups of validators

// get the set of all validators with no limits, used during genesis dump
func (k Keeper) GetAllValidators(ctx context.Context) (validators []types.Validator, err error) {
	err = k.Validators.Walk(ctx, nil, func(key []byte, validator types.Validator) (stop bool, err error) {
		validators = append(validators, validator)
		return false, nil
	})

	return validators, err
}

// return a given amount of all the validators
func (k Keeper) GetValidators(ctx context.Context, maxRetrieve uint32) (validators []types.Validator, err error) {
	validators = make([]types.Validator, 0, maxRetrieve)
	err = k.Validators.Walk(ctx, nil, func(key []byte, validator types.Validator) (stop bool, err error) {
		validators = append(validators, validator)
		return len(validators) == int(maxRetrieve), nil
	})

	return validators, err
}

// Last Validator Index

// Load the last validator power.
// Returns zero if the operator was not a validator last block.
func (k Keeper) GetLastValidatorPower(ctx context.Context, operator sdk.ValAddress) (int64, error) {
	power, err := k.LastValidatorPowers.Get(ctx, operator)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return 0, nil
		}

		return 0, err
	}

	return power, nil
}

// Set the last validator power.
func (k Keeper) SetLastValidatorPower(ctx context.Context, operator sdk.ValAddress, power int64) error {
	return k.LastValidatorPowers.Set(ctx, operator, power)
}

// Delete the last validator power.
func (k Keeper) DeleteLastValidatorPower(ctx context.Context, operator sdk.ValAddress) error {
	return k.LastValidatorPowers.Remove(ctx, operator)
}

// Iterate over last validator powers.
func (k Keeper) IterateLastValidatorPowers(ctx context.Context, handler func(operator []byte, power int64) (stop bool, err error)) error {
	return k.LastValidatorPowers.Walk(ctx, nil, handler)
}

// get the group of the bonded validators
func (k Keeper) GetLastValidators(ctx context.Context) (validators []types.Validator, err error) {
	maxValidators, err := k.MaxValidators(ctx)
	if err != nil {
		return nil, err
	}

	validators = make([]types.Validator, 0, maxValidators)
	err = k.IterateLastValidatorPowers(ctx, func(operator []byte, power int64) (stop bool, err error) {
		validators = append(validators, k.mustGetValidator(ctx, operator))
		// sanity check
		if len(validators) > int(maxValidators) {
			panic("more validators than maxValidators found")
		}

		return false, nil
	})

	return validators, err
}
