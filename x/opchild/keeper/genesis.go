package keeper

import (
	"context"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/OPinit/x/opchild/types"
)

// InitGenesis sets the pool and parameters for the provided keeper.  For each
// validator in data, it sets that validator in the keeper along with manually
// setting the indexes. In addition, it also sets any delegations found in
// data. Finally, it updates the bonded validators.
// Returns final validator set after applying all declaration and delegations
func (k Keeper) InitGenesis(ctx context.Context, data *types.GenesisState) (res []abci.ValidatorUpdate) {
	// We need to pretend to be "n blocks before genesis", where "n" is the
	// validator update delay, so that e.g. slashing periods are correctly
	// initialized for the validator set e.g. with a one-block offset - the
	// first TM block is at height 1, so state updates applied from
	// genesis.json are in block 0.
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx = sdkCtx.WithBlockHeight(1 - sdk.ValidatorUpdateDelay)
	ctx = sdkCtx

	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	for _, validator := range data.Validators {
		k.SetValidator(ctx, validator)

		// Manually set indices for the first time
		if err := k.SetValidatorByConsAddr(ctx, validator); err != nil {
			panic(err)
		}
	}

	// don't need to run Tendermint updates if we exported
	if data.Exported {
		for _, lv := range data.LastValidatorPowers {
			valAddr, err := k.validatorAddressCodec.StringToBytes(lv.Address)
			if err != nil {
				panic(err)
			}

			k.SetLastValidatorPower(ctx, valAddr, lv.Power)
			validator, found := k.GetValidator(ctx, valAddr)

			if !found {
				panic(fmt.Sprintf("validator %s not found", lv.Address))
			}

			update := validator.ABCIValidatorUpdate()
			update.Power = lv.Power // keep the next-val-set offset, use the last power for the first block
			res = append(res, update)
		}
	} else {
		var err error

		res, err = k.ApplyAndReturnValidatorSetUpdates(ctx)
		if err != nil {
			panic(err)
		}
	}

	for _, finalizedL1Sequence := range data.FinalizedL1Sequences {
		k.RecordFinalizedL1Sequence(ctx, finalizedL1Sequence)
	}

	k.SetNextL2Sequence(ctx, data.NextL2Sequence)

	return res
}

// ExportGenesis returns a GenesisState for a given context and keeper. The
// GenesisState will contain the pool, params, validators, and bonds found in
// the keeper.
func (k Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
	var lastValidatorPowers []types.LastValidatorPower
	err := k.IterateLastValidatorPowers(ctx, func(addr []byte, power int64) (stop bool, err error) {
		lastValidatorPowers = append(lastValidatorPowers, types.LastValidatorPower{Address: sdk.ValAddress(addr).String(), Power: power})
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	var finalizedL1Sequences []uint64
	err = k.IterateFinalizedL1Sequences(ctx, func(l1Sequence uint64) (bool, error) {
		finalizedL1Sequences = append(finalizedL1Sequences, l1Sequence)
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	validators, err := k.GetAllValidators(ctx)
	if err != nil {
		panic(err)
	}

	nextL2Sequence, err := k.GetNextL2Sequence(ctx)
	if err != nil {
		panic(err)
	}

	return &types.GenesisState{
		Params:               params,
		LastValidatorPowers:  lastValidatorPowers,
		Validators:           validators,
		Exported:             true,
		FinalizedL1Sequences: finalizedL1Sequences,
		NextL2Sequence:       nextL2Sequence,
	}
}
