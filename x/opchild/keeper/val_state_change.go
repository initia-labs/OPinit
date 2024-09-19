package keeper

import (
	"bytes"
	"context"
	"errors"
	"sort"

	"cosmossdk.io/core/address"
	abci "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/OPinit/x/opchild/types"
)

// BlockValidatorUpdates calculates the ValidatorUpdates for the current block
// Called in each EndBlock
func (k Keeper) BlockValidatorUpdates(ctx context.Context) ([]abci.ValidatorUpdate, error) {
	updates, err := k.ApplyAndReturnValidatorSetUpdates(ctx)
	if err != nil {
		return nil, err
	}

	return updates, nil
}

func (k Keeper) ApplyAndReturnValidatorSetUpdates(ctx context.Context) ([]abci.ValidatorUpdate, error) {
	last, err := k.getLastValidatorsByAddr(ctx)
	if err != nil {
		return nil, err
	}

	updates := []abci.ValidatorUpdate{}
	validators, err := k.GetAllValidators(ctx)
	if err != nil {
		return nil, err
	}

	for _, validator := range validators {
		valAddr, err := k.validatorAddressCodec.StringToBytes(validator.GetOperator())
		if err != nil {
			return nil, err
		}

		oldPower, found := last[validator.GetOperator()]
		newPower := validator.ConsensusPower()

		// zero power validator removed from validator set
		if newPower <= 0 {
			continue
		}

		if !found || oldPower != newPower {
			updates = append(updates, validator.ABCIValidatorUpdate())

			if err := k.SetLastValidatorPower(ctx, valAddr, newPower); err != nil {
				return nil, err
			}
		}

		delete(last, validator.GetOperator())
	}

	noLongerBonded, err := sortNoLongerBonded(last, k.validatorAddressCodec)
	if err != nil {
		return nil, err
	}

	for _, valAddrBytes := range noLongerBonded {
		validator := k.mustGetValidator(ctx, sdk.ValAddress(valAddrBytes))
		if validator.ConsPower > 0 {
			return nil, errors.New("deleting validator cannot have positive power")
		}

		valAddr, err := k.validatorAddressCodec.StringToBytes(validator.GetOperator())
		if err != nil {
			return nil, err
		}

		if err := k.RemoveValidator(ctx, valAddr); err != nil {
			return nil, err
		}

		if err := k.DeleteLastValidatorPower(ctx, valAddr); err != nil {
			return nil, err
		}

		updates = append(updates, validator.ABCIValidatorUpdate())
	}
	return updates, nil
}

// map of operator bech32-addresses to serialized power
// We use bech32 strings here, because we can't have slices as keys: map[[]byte][]byte
type validatorsByAddr map[string]int64

// get the last validator set
func (k Keeper) getLastValidatorsByAddr(ctx context.Context) (validatorsByAddr, error) {
	last := make(validatorsByAddr)

	err := k.IterateLastValidators(ctx, func(validator types.ValidatorI, power int64) (stop bool, err error) {
		last[validator.GetOperator()] = power
		return false, nil
	})

	return last, err
}

// given a map of remaining validators to previous bonded power
// returns the list of validators to be unbonded, sorted by operator address
func sortNoLongerBonded(last validatorsByAddr, vc address.Codec) ([][]byte, error) {
	// sort the map keys for determinism
	noLongerBonded := make([][]byte, len(last))
	index := 0

	for valAddrStr := range last {
		valAddrBytes, err := vc.StringToBytes(valAddrStr)
		if err != nil {
			return nil, err
		}
		noLongerBonded[index] = valAddrBytes
		index++
	}
	// sorted by address - order doesn't matter
	sort.SliceStable(noLongerBonded, func(i, j int) bool {
		// -1 means strictly less than
		return bytes.Compare(noLongerBonded[i], noLongerBonded[j]) == -1
	})

	return noLongerBonded, nil
}
