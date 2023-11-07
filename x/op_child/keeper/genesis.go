package keeper

import (
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/OPinit/x/op_child/types"
)

// InitGenesis sets the pool and parameters for the provided keeper.  For each
// validator in data, it sets that validator in the keeper along with manually
// setting the indexes. In addition, it also sets any delegations found in
// data. Finally, it updates the bonded validators.
// Returns final validator set after applying all declaration and delegations
func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) (res []abci.ValidatorUpdate) {
	// We need to pretend to be "n blocks before genesis", where "n" is the
	// validator update delay, so that e.g. slashing periods are correctly
	// initialized for the validator set e.g. with a one-block offset - the
	// first TM block is at height 1, so state updates applied from
	// genesis.json are in block 0.
	ctx = ctx.WithBlockHeight(1 - sdk.ValidatorUpdateDelay)

	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	for _, validator := range data.Validators {
		k.SetValidator(ctx, validator)

		// Manually set indices for the first time
		k.SetValidatorByConsAddr(ctx, validator)
	}

	// don't need to run Tendermint updates if we exported
	if data.Exported {
		for _, lv := range data.LastValidatorPowers {
			valAddr, err := sdk.ValAddressFromBech32(lv.Address)
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

		var previousProposer sdk.ConsAddress
		if data.PreviousProposer != "" {
			var err error
			previousProposer, err = sdk.ConsAddressFromBech32(data.PreviousProposer)
			if err != nil {
				panic(err)
			}
		}

		k.SetPreviousProposerConsAddr(ctx, previousProposer)
	} else {
		var err error

		res, err = k.ApplyAndReturnValidatorSetUpdates(ctx)
		if err != nil {
			panic(err)
		}
	}

	return res
}

// ExportGenesis returns a GenesisState for a given context and keeper. The
// GenesisState will contain the pool, params, validators, and bonds found in
// the keeper.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {

	var lastValidatorPowers []types.LastValidatorPower

	k.IterateLastValidatorPowers(ctx, func(addr sdk.ValAddress, power int64) (stop bool) {
		lastValidatorPowers = append(lastValidatorPowers, types.LastValidatorPower{Address: addr.String(), Power: power})
		return false
	})

	return &types.GenesisState{
		Params:              k.GetParams(ctx),
		LastValidatorPowers: lastValidatorPowers,
		Validators:          k.GetAllValidators(ctx),
		PreviousProposer:    k.GetPreviousProposerConsAddr(ctx).String(),
		Exported:            true,
	}
}
