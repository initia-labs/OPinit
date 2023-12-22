package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func Test_Params(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	params, err := input.OPChildKeeper.GetParams(ctx)
	require.NoError(t, err)
	params.MinGasPrices = sdk.NewDecCoins()

	input.OPChildKeeper.SetParams(ctx, params)

	minGasPrices, err := input.OPChildKeeper.MinGasPrices(ctx)
	require.NoError(t, err)
	require.True(t, minGasPrices.Empty())
	bridgeExecutor, err := input.OPChildKeeper.BridgeExecutor(ctx)
	require.NoError(t, err)
	require.Equal(t, addrs[0], bridgeExecutor)
}
