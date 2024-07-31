package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_NextL1GetNextL1Sequence(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	res, err := input.OPChildKeeper.GetNextL1Sequence(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(1), res)

	_, err = input.OPChildKeeper.IncreaseNextL1Sequence(ctx)
	require.NoError(t, err)
	res, err = input.OPChildKeeper.GetNextL1Sequence(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(2), res)
}

func Test_SetAndSetNextL2Sequence(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	seq, err := input.OPChildKeeper.GetNextL2Sequence(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(1), seq)

	require.NoError(t, input.OPChildKeeper.SetNextL2Sequence(ctx, 1204))
	seq, err = input.OPChildKeeper.GetNextL2Sequence(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(1204), seq)
}

func Test_IncreaseNextL2Sequence(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	seq, err := input.OPChildKeeper.GetNextL2Sequence(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(1), seq)

	seq, err = input.OPChildKeeper.IncreaseNextL2Sequence(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(1), seq)
	seq, err = input.OPChildKeeper.IncreaseNextL2Sequence(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(2), seq)
}
