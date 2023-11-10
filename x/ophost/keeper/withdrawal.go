package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/OPinit/x/ophost/types"
)

func (k Keeper) RecordProvenWithdrawal(ctx sdk.Context, bridgeId uint64, withdrawalHash [32]byte) {
	kvStore := ctx.KVStore(k.storeKey)
	kvStore.Set(types.GetProvenWithdrawalKey(bridgeId, withdrawalHash), []byte{1})
}

func (k Keeper) HasProvenWithdrawal(ctx sdk.Context, bridgeId uint64, withdrawalHash [32]byte) bool {
	kvStore := ctx.KVStore(k.storeKey)
	return kvStore.Has(types.GetProvenWithdrawalKey(bridgeId, withdrawalHash))
}

func (k Keeper) IterateProvenWithdrawals(
	ctx sdk.Context,
	bridgeId uint64,
	cb func(bridgeId uint64, withdrawalHash [32]byte) bool,
) error {
	kvStore := ctx.KVStore(k.storeKey)

	prefixStore := prefix.NewStore(kvStore, types.GetProvenWithdrawalPrefixKey(bridgeId))
	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()

		withdrawalHash := [32]byte{}
		copy(withdrawalHash[:], key)

		if cb(bridgeId, withdrawalHash) {
			break
		}
	}

	return nil
}
