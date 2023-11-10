package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/initia-labs/OPinit/x/ophost/types"
)

func (k Keeper) SetTokenPair(ctx sdk.Context, bridgeId uint64, l2Denom, l1Denom string) {
	kvStore := ctx.KVStore(k.storeKey)
	kvStore.Set(types.GetTokenPairKey(bridgeId, l2Denom), []byte(l1Denom))
}

func (k Keeper) HasTokenPair(ctx sdk.Context, bridgeId uint64, l2Denom string) bool {
	kvStore := ctx.KVStore(k.storeKey)
	return kvStore.Has(types.GetTokenPairKey(bridgeId, l2Denom))
}

func (k Keeper) GetTokenPair(
	ctx sdk.Context,
	bridgeId uint64,
	l2Denom string,
) (l1Denom string, err error) {
	kvStore := ctx.KVStore(k.storeKey)
	bz := kvStore.Get(types.GetTokenPairKey(bridgeId, l2Denom))
	if len(bz) == 0 {
		err = errors.ErrKeyNotFound
		return
	}

	l1Denom = string(bz)
	return
}

func (k Keeper) IterateTokenPair(
	ctx sdk.Context,
	bridgeId uint64,
	cb func(bridgeId uint64, tokenPair types.TokenPair) bool,
) error {
	kvStore := ctx.KVStore(k.storeKey)

	prefixStore := prefix.NewStore(kvStore, types.GetTokenPairBridgePrefixKey(bridgeId))
	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		val := iterator.Value()

		l2Denom := string(key)
		l1Denom := string(val)

		if cb(bridgeId, types.TokenPair{
			L2Denom: l2Denom,
			L1Denom: l1Denom,
		}) {
			break
		}
	}

	return nil
}
