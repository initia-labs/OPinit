package keeper_test

import (
	"testing"

	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"
	cosmostypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

func Test_HistoricalInfo(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	historicalInfo := cosmostypes.HistoricalInfo{
		Header: tmtypes.Header{
			Height:  100,
			ChainID: "testnet",
		},
		Valset: []cosmostypes.Validator{{
			OperatorAddress: "hihi",
		}},
	}

	input.OPChildKeeper.SetHistoricalInfo(ctx, 100, &historicalInfo)

	_historicalInfo, found := input.OPChildKeeper.GetHistoricalInfo(ctx, 101)
	require.True(t, found)
	require.Equal(t, historicalInfo.Header.Height, _historicalInfo.Header.Height)
	require.Equal(t, historicalInfo.Header.ChainID, _historicalInfo.Header.ChainID)
	require.Equal(t, historicalInfo.Valset[0].OperatorAddress, _historicalInfo.Valset[0].OperatorAddress)

	_, found = input.OPChildKeeper.GetHistoricalInfo(ctx, 99)
	require.False(t, found)
}

func Test_TrackHistoricalInfo(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	emptyHistoricalInfo := cosmostypes.HistoricalInfo{}
	input.OPChildKeeper.SetHistoricalInfo(ctx, 100, &emptyHistoricalInfo)
	input.OPChildKeeper.SetHistoricalInfo(ctx, 101, &emptyHistoricalInfo)
	input.OPChildKeeper.SetHistoricalInfo(ctx, 102, &emptyHistoricalInfo)
	input.OPChildKeeper.SetHistoricalInfo(ctx, 103, &emptyHistoricalInfo)

	ctx = ctx.WithBlockHeight(104)
	params := input.OPChildKeeper.GetParams(ctx)
	params.HistoricalEntries = 1
	input.OPChildKeeper.SetParams(ctx, params)

	input.OPChildKeeper.TrackHistoricalInfo(ctx)

	_, found := input.OPChildKeeper.GetHistoricalInfo(ctx, 102)
	require.True(t, found)
	_, found = input.OPChildKeeper.GetHistoricalInfo(ctx, 101)
	require.False(t, found)
	_, found = input.OPChildKeeper.GetHistoricalInfo(ctx, 100)
	require.False(t, found)
}

func Test_DeleteHistoricalInfo(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	emptyHistoricalInfo := cosmostypes.HistoricalInfo{}
	input.OPChildKeeper.SetHistoricalInfo(ctx, 100, &emptyHistoricalInfo)

	input.OPChildKeeper.DeleteHistoricalInfo(ctx, 100)
	_, found := input.OPChildKeeper.GetHistoricalInfo(ctx, 100)
	require.False(t, found)
}
