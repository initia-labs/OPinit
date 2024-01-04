package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmostypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// GetHistoricalInfo gets the historical info at a given height
func (k Keeper) GetHistoricalInfo(ctx context.Context, height int64) (*cosmostypes.HistoricalInfo, bool, error) {
	historicalInfo, err := k.HistoricalInfos.Get(ctx, height)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, false, nil
		}

		return nil, false, err
	}

	return &historicalInfo, true, nil
}

// SetHistoricalInfo sets the historical info at a given height
func (k Keeper) SetHistoricalInfo(ctx context.Context, height int64, hi *cosmostypes.HistoricalInfo) error {
	return k.HistoricalInfos.Set(ctx, height, *hi)
}

// DeleteHistoricalInfo deletes the historical info at a given height
func (k Keeper) DeleteHistoricalInfo(ctx context.Context, height int64) error {
	return k.HistoricalInfos.Remove(ctx, height)
}

// TrackHistoricalInfo saves the latest historical-info and deletes the oldest
// heights that are below pruning height
func (k Keeper) TrackHistoricalInfo(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	entryNum, err := k.HistoricalEntries(ctx)
	if err != nil {
		return err
	}

	// Prune store to ensure we only have parameter-defined historical entries.
	// In most cases, this will involve removing a single historical entry.
	// In the rare scenario when the historical entries gets reduced to a lower value k'
	// from the original value k. k - k' entries must be deleted from the store.
	// Since the entries to be deleted are always in a continuous range, we can iterate
	// over the historical entries starting from the most recent version to be pruned
	// and then return at the first empty entry.
	for i := sdkCtx.BlockHeight() - int64(entryNum); i >= 0; i-- {
		_, found, err := k.GetHistoricalInfo(ctx, i)
		if err != nil {
			return err
		} else if found {
			if err := k.DeleteHistoricalInfo(ctx, i); err != nil {
				return err
			}
		} else {
			break
		}
	}

	// if there is no need to persist historicalInfo, return
	if entryNum == 0 {
		return nil
	}

	// Create HistoricalInfo struct
	lastVals, err := k.GetLastValidators(ctx)
	if err != nil {
		return err
	}
	var lastCosmosVals cosmostypes.Validators
	for _, v := range lastVals {
		lastCosmosVals.Validators = append(lastCosmosVals.Validators, cosmostypes.Validator{
			ConsensusPubkey: v.ConsensusPubkey,
			Tokens:          math.NewInt(v.ConsensusPower() * sdk.DefaultPowerReduction.Int64()),
			Status:          cosmostypes.Bonded,
		})
	}
	lastCosmosVals.ValidatorCodec = k.validatorAddressCodec

	historicalEntry := cosmostypes.NewHistoricalInfo(sdkCtx.BlockHeader(), lastCosmosVals, sdk.DefaultPowerReduction)

	// Set latest HistoricalInfo at current height
	return k.SetHistoricalInfo(ctx, sdkCtx.BlockHeight(), &historicalEntry)
}
