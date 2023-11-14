package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_MaxValidators(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	params := input.OPChildKeeper.GetParams(ctx)
	params.MaxValidators = 10
	input.OPChildKeeper.SetParams(ctx, params)

	maxValidators := input.OPChildKeeper.MaxValidators(ctx)
	require.Equal(t, uint32(10), maxValidators)
}

func Test_HistoricalEntries(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	params := input.OPChildKeeper.GetParams(ctx)
	params.HistoricalEntries = 10
	input.OPChildKeeper.SetParams(ctx, params)

	entries := input.OPChildKeeper.HistoricalEntries(ctx)
	require.Equal(t, uint32(10), entries)
}

func Test_UnbondingTime(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	unbondingTime := input.OPChildKeeper.UnbondingTime(ctx)
	require.Equal(t, (60 * 60 * 24 * 7 * time.Second), unbondingTime)
}
