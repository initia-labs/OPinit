package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	cosmostypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func Test_HistoricalInfo(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	params := input.OPChildKeeper.GetParams(ctx)
	params.HistoricalEntries = 2
	input.OPChildKeeper.SetParams(ctx, params)

	input.OPChildKeeper.TrackHistoricalInfo(ctx.WithBlockHeight(1))
	input.OPChildKeeper.TrackHistoricalInfo(ctx.WithBlockHeight(2))
	input.OPChildKeeper.TrackHistoricalInfo(ctx.WithBlockHeight(3))

	_, found := input.OPChildKeeper.GetHistoricalInfo(ctx, 1)
	require.False(t, found)

	historicalInfo, found := input.OPChildKeeper.GetHistoricalInfo(ctx, 2)
	require.True(t, found)
	require.Equal(t, cosmostypes.HistoricalInfo{
		Header: ctx.WithBlockHeight(2).BlockHeader(),
		Valset: nil,
	}, historicalInfo)

	historicalInfo, found = input.OPChildKeeper.GetHistoricalInfo(ctx, 3)
	require.True(t, found)
	require.Equal(t, cosmostypes.HistoricalInfo{
		Header: ctx.WithBlockHeight(3).BlockHeader(),
		Valset: nil,
	}, historicalInfo)
}
