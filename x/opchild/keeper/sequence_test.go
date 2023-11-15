package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_FinalizedL1Sequence(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	res := input.OPChildKeeper.HasFinalizedL1Sequence(ctx, 1)
	require.False(t, res)

	input.OPChildKeeper.RecordFinalizedL1Sequence(ctx, 1)
	res = input.OPChildKeeper.HasFinalizedL1Sequence(ctx, 1)
	require.True(t, res)
}

func Test_IterateFinalizedL1Sequences(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	sequences := []uint64{1, 2, 4}
	for _, v := range sequences {
		input.OPChildKeeper.RecordFinalizedL1Sequence(ctx, v)
	}
	input.OPChildKeeper.IterateFinalizedL1Sequences(ctx, func(l1Sequence uint64) bool {
		require.Equal(t, sequences[0], l1Sequence)
		sequences = sequences[1:]
		return false
	})
}

func Test_SetAndSetNextL2Sequence(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	seq := input.OPChildKeeper.GetNextL2Sequence(ctx)
	require.Equal(t, uint64(1), seq)

	input.OPChildKeeper.SetNextL2Sequence(ctx, 1204)
	seq = input.OPChildKeeper.GetNextL2Sequence(ctx)
	require.Equal(t, uint64(1204), seq)
}

func Test_IncreaseNextL2Sequence(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	seq := input.OPChildKeeper.GetNextL2Sequence(ctx)
	require.Equal(t, uint64(1), seq)

	seq = input.OPChildKeeper.IncreaseNextL2Sequence(ctx)
	require.Equal(t, uint64(1), seq)
	seq = input.OPChildKeeper.IncreaseNextL2Sequence(ctx)
	require.Equal(t, uint64(2), seq)
}
