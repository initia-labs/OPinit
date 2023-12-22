package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
)

func (k Keeper) RecordFinalizedL1Sequence(ctx context.Context, l1Sequence uint64) error {
	return k.FinalizedL1Sequence.Set(ctx, l1Sequence, true)
}

func (k Keeper) HasFinalizedL1Sequence(ctx context.Context, l1Sequence uint64) (bool, error) {
	return k.FinalizedL1Sequence.Has(ctx, l1Sequence)
}

func (k Keeper) IterateFinalizedL1Sequences(ctx context.Context, cb func(l1Sequence uint64) (stop bool, err error)) error {
	return k.FinalizedL1Sequence.Walk(ctx, nil, func(l1sequence uint64, _ bool) (stop bool, err error) {
		return cb(l1sequence)
	})
}

func (k Keeper) SetNextL2Sequence(ctx context.Context, l2Sequence uint64) error {
	return k.NextL2Sequence.Set(ctx, l2Sequence)
}

func (k Keeper) GetNextL2Sequence(ctx context.Context) (uint64, error) {
	nextL2Sequence, err := k.NextL2Sequence.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return 1, nil
		}

		return 0, err
	}

	return nextL2Sequence, nil
}

func (k Keeper) IncreaseNextL2Sequence(ctx context.Context) (uint64, error) {
	nextL2Sequence, err := k.GetNextL2Sequence(ctx)
	if err != nil {
		return 0, err
	}

	// increase NextL2Sequence
	if err = k.NextL2Sequence.Set(ctx, nextL2Sequence+1); err != nil {
		return 0, err
	}

	return nextL2Sequence, nil
}
