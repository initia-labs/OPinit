package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_MaxValidators(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	params, err := input.OPChildKeeper.GetParams(ctx)
	require.NoError(t, err)
	params.MaxValidators = 10
	require.NoError(t, input.OPChildKeeper.SetParams(ctx, params))

	maxValidators, err := input.OPChildKeeper.MaxValidators(ctx)
	require.NoError(t, err)
	require.Equal(t, uint32(10), maxValidators)
}

func Test_HistoricalEntries(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	params, err := input.OPChildKeeper.GetParams(ctx)
	require.NoError(t, err)
	params.HistoricalEntries = 10
	require.NoError(t, input.OPChildKeeper.SetParams(ctx, params))

	entries, err := input.OPChildKeeper.HistoricalEntries(ctx)
	require.NoError(t, err)
	require.Equal(t, uint32(10), entries)
}

func Test_UnbondingTime(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	unbondingTime, err := input.OPChildKeeper.UnbondingTime(ctx)
	require.NoError(t, err)
	require.Equal(t, (60 * 60 * 24 * 7 * time.Second), unbondingTime)
}
