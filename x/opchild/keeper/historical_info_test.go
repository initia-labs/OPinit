package keeper_test

import (
	"testing"

	"cosmossdk.io/collections"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmostypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/initia-labs/OPinit/x/opchild/testutil"
)

func Test_HistoricalInfo(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	params, err := input.OPChildKeeper.GetParams(ctx)
	require.NoError(t, err)
	params.HistoricalEntries = 2
	require.NoError(t, input.OPChildKeeper.SetParams(ctx, params))

	require.NoError(t, input.OPChildKeeper.TrackHistoricalInfo(sdkCtx.WithBlockHeight(1)))
	require.NoError(t, input.OPChildKeeper.TrackHistoricalInfo(sdkCtx.WithBlockHeight(2)))
	require.NoError(t, input.OPChildKeeper.TrackHistoricalInfo(sdkCtx.WithBlockHeight(3)))

	_, err = input.OPChildKeeper.GetHistoricalInfo(ctx, 1)
	require.ErrorIs(t, err, collections.ErrNotFound)

	historicalInfo, err := input.OPChildKeeper.GetHistoricalInfo(ctx, 2)
	require.NoError(t, err)
	require.Equal(t, cosmostypes.HistoricalInfo{
		Header: sdkCtx.WithBlockHeight(2).BlockHeader(),
		Valset: nil,
	}, historicalInfo)

	historicalInfo, err = input.OPChildKeeper.GetHistoricalInfo(ctx, 3)
	require.NoError(t, err)
	require.Equal(t, cosmostypes.HistoricalInfo{
		Header: sdkCtx.WithBlockHeight(3).BlockHeader(),
		Valset: nil,
	}, historicalInfo)
}
