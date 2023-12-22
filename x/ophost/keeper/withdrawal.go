package keeper

import (
	"context"

	"cosmossdk.io/collections"
)

func (k Keeper) RecordProvenWithdrawal(ctx context.Context, bridgeId uint64, withdrawalHash [32]byte) error {
	return k.ProvenWithdrawals.Set(ctx, collections.Join(bridgeId, withdrawalHash[:]), true)
}

func (k Keeper) HasProvenWithdrawal(ctx context.Context, bridgeId uint64, withdrawalHash [32]byte) (bool, error) {
	return k.ProvenWithdrawals.Has(ctx, collections.Join(bridgeId, withdrawalHash[:]))
}

func (k Keeper) IterateProvenWithdrawals(
	ctx context.Context,
	bridgeId uint64,
	cb func(bridgeId uint64, withdrawalHash [32]byte) (bool, error),
) error {
	return k.ProvenWithdrawals.Walk(ctx, collections.NewPrefixedPairRange[uint64, []byte](bridgeId), func(key collections.Pair[uint64, []byte], value bool) (stop bool, err error) {
		withdrawalHash := [32]byte{}
		copy(withdrawalHash[:], key.K2())
		return cb(bridgeId, withdrawalHash)
	})
}
