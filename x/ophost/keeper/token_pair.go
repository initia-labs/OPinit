package keeper

import (
	"context"

	"cosmossdk.io/collections"

	"github.com/initia-labs/OPinit/x/ophost/types"
)

func (k Keeper) SetTokenPair(ctx context.Context, bridgeId uint64, l2Denom, l1Denom string) error {
	return k.TokenPairs.Set(ctx, collections.Join(bridgeId, l2Denom), l1Denom)
}

func (k Keeper) HasTokenPair(ctx context.Context, bridgeId uint64, l2Denom string) (bool, error) {
	return k.TokenPairs.Has(ctx, collections.Join(bridgeId, l2Denom))
}

func (k Keeper) GetTokenPair(
	ctx context.Context,
	bridgeId uint64,
	l2Denom string,
) (l1Denom string, err error) {
	return k.TokenPairs.Get(ctx, collections.Join(bridgeId, l2Denom))
}

func (k Keeper) IterateTokenPair(
	ctx context.Context,
	bridgeId uint64,
	cb func(bridgeId uint64, tokenPair types.TokenPair) (stop bool, err error),
) error {
	return k.TokenPairs.Walk(ctx, collections.NewPrefixedPairRange[uint64, string](bridgeId), func(key collections.Pair[uint64, string], l1Denom string) (stop bool, err error) {
		return cb(key.K1(), types.TokenPair{
			L1Denom: l1Denom,
			L2Denom: key.K2(),
		})
	})
}
