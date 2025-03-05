package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"github.com/initia-labs/OPinit/v1/x/opchild/types"
	ophosttypes "github.com/initia-labs/OPinit/v1/x/ophost/types"
)

func (k Keeper) GetNextL1Sequence(ctx context.Context) (uint64, error) {
	finalizedL1Sequence, err := k.NextL1Sequence.Peek(ctx)
	if err != nil {
		return 0, err
	}

	if finalizedL1Sequence == collections.DefaultSequenceStart {
		return ophosttypes.DefaultL1SequenceStart, nil
	}

	return finalizedL1Sequence, nil
}

func (k Keeper) SetNextL1Sequence(ctx context.Context, l1Sequence uint64) error {
	return k.NextL1Sequence.Set(ctx, l1Sequence)
}

func (k Keeper) IncreaseNextL1Sequence(ctx context.Context) (uint64, error) {
	finalizedL1Sequence, err := k.NextL1Sequence.Next(ctx)
	if err != nil {
		return 0, err
	}

	if finalizedL1Sequence == collections.DefaultSequenceStart {
		if err := k.NextL1Sequence.Set(ctx, ophosttypes.DefaultL1SequenceStart+1); err != nil {
			return 0, err
		}

		return ophosttypes.DefaultL1SequenceStart, nil
	}

	return finalizedL1Sequence, nil
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
