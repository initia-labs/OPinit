package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"github.com/initia-labs/OPinit/x/opchild/types"
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
	nextL2Sequence, err := k.NextL2Sequence.Peek(ctx)
	if err != nil {
		return 0, err
	}

	if nextL2Sequence == collections.DefaultSequenceStart {
		return types.DefaultL2SequenceStart, nil
	}

	return nextL2Sequence, nil
}

func (k Keeper) IncreaseNextL2Sequence(ctx context.Context) (uint64, error) {
	nextL2Sequence, err := k.NextL2Sequence.Next(ctx)
	if err != nil {
		return 0, err
	}

	if nextL2Sequence == collections.DefaultSequenceStart {
		if err := k.NextL2Sequence.Set(ctx, types.DefaultL2SequenceStart+1); err != nil {
			return 0, err
		}

		return types.DefaultL2SequenceStart, nil
	}

	return nextL2Sequence, nil
}
