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
		if err := k.SetValidator(ctx, validator); err != nil {
			panic(err)
		}

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

			if err := k.SetLastValidatorPower(ctx, valAddr, lv.Power); err != nil {
				panic(err)
			}

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

	if err := k.SetNextL1Sequence(ctx, data.NextL1Sequence); err != nil {
		panic(err)
	}

	if err := k.SetNextL2Sequence(ctx, data.NextL2Sequence); err != nil {
		panic(err)
	}

	if data.BridgeInfo != nil {
		if err := data.BridgeInfo.Validate(k.addressCodec); err != nil {
			panic(err)
		}

		if err := k.BridgeInfo.Set(ctx, *data.BridgeInfo); err != nil {
			panic(err)
		}
	}

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

	finalizedL1Sequence, err := k.GetNextL1Sequence(ctx)
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

	var bridgeInfo *types.BridgeInfo
	if ok, err := k.BridgeInfo.Has(ctx); err != nil {
		panic(err)
	} else if ok {
		bridgeInfo_, err := k.BridgeInfo.Get(ctx)
		if err != nil {
			panic(err)
		}

		bridgeInfo = &bridgeInfo_
	}

	return &types.GenesisState{
		Params:              params,
		LastValidatorPowers: lastValidatorPowers,
		Validators:          validators,
		Exported:            true,
		NextL1Sequence:      finalizedL1Sequence,
		NextL2Sequence:      nextL2Sequence,
		BridgeInfo:          bridgeInfo,
	}
}
