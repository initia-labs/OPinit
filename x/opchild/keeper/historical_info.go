package keeper

import (
	"encoding/binary"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmostypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/initia-labs/OPinit/x/opchild/types"
)

// GetHistoricalInfo fetch height historical info that is equal or lower than the given height.
func (k Keeper) GetHistoricalInfo(ctx sdk.Context, height int64) (cosmostypes.HistoricalInfo, bool) {
	store := ctx.KVStore(k.storeKey)

	// increase height by 1 because iterator is exclusive.
	height += 1

	prefixStore := prefix.NewStore(store, types.HistoricalInfoKey)

	end := make([]byte, 8)
	binary.BigEndian.PutUint64(end, uint64(height))

	iterator := prefixStore.ReverseIterator(nil, end)
	defer iterator.Close()

	if !iterator.Valid() {
		return cosmostypes.HistoricalInfo{}, false
	}

	value := iterator.Value()
	return cosmostypes.MustUnmarshalHistoricalInfo(k.cdc, value), true
}

// SetHistoricalInfo sets the historical info at a given height
func (k Keeper) SetHistoricalInfo(ctx sdk.Context, height int64, hi *cosmostypes.HistoricalInfo) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetHistoricalInfoKey(uint64(height))
	value := k.cdc.MustMarshal(hi)
	store.Set(key, value)
}

// DeleteHistoricalInfo deletes the historical info at a given height
func (k Keeper) DeleteHistoricalInfo(ctx sdk.Context, height int64) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetHistoricalInfoKey(uint64(height))

	store.Delete(key)
}

// TrackHistoricalInfo saves the latest historical-info and deletes the oldest
// heights that are below pruning height
func (k Keeper) TrackHistoricalInfo(ctx sdk.Context) {
	entryNum := uint64(k.HistoricalEntries(ctx))

	// Prune store to ensure we only have parameter-defined historical entries.
	// In most cases, this will involve removing a single historical entry.
	// In the rare scenario when the historical entries gets reduced to a lower value k'
	// from the original value k. k - k' entries must be deleted from the store.
	// Since the entries to be deleted are always in a continuous range, we can iterate
	// over the historical entries starting from the most recent version to be pruned
	// and then return at the first empty entry.
	height := uint64(ctx.BlockHeight())
	if height > entryNum {
		store := ctx.KVStore(k.storeKey)
		prefixStore := prefix.NewStore(store, types.HistoricalInfoKey)

		end := make([]byte, 8)
		binary.BigEndian.PutUint64(end, height-entryNum)

		iterator := prefixStore.ReverseIterator(nil, end)
		defer iterator.Close()

		// our historical info does not exist for every block to allow
		// empty block, so it is possible when ibc request deleted block
		// historical info. Then opchild module returns height historical
		// historical info that is lower than the given height.
		//
		// Whenever we delete historical info, we have to leave first info
		// for safety.
		if iterator.Valid() {
			iterator.Next()
		}

		for ; iterator.Valid(); iterator.Next() {
			key := iterator.Key()
			prefixStore.Delete(key)
		}
	}

	// if there is no need to persist historicalInfo, return
	if entryNum == 0 {
		return
	}

	// Create HistoricalInfo struct
	lastVals := k.GetLastValidators(ctx)
	var lastCosmosVals cosmostypes.Validators
	for _, v := range lastVals {
		lastCosmosVals = append(lastCosmosVals, cosmostypes.Validator{
			ConsensusPubkey: v.ConsensusPubkey,
			Tokens:          math.NewInt(v.ConsensusPower() * sdk.DefaultPowerReduction.Int64()),
			Status:          cosmostypes.Bonded,
		})
	}

	historicalEntry := cosmostypes.NewHistoricalInfo(ctx.BlockHeader(), lastCosmosVals, sdk.DefaultPowerReduction)

	// Set latest HistoricalInfo at current height
	k.SetHistoricalInfo(ctx, ctx.BlockHeight(), &historicalEntry)
}
