package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func Test_Params(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	params := input.OPChildKeeper.GetParams(ctx)
	params.MinGasPrices = sdk.NewDecCoins()

	input.OPChildKeeper.SetParams(ctx, params)

	require.True(t, input.OPChildKeeper.MinGasPrices(ctx).IsZero())
	require.Equal(t, addrs[0], input.OPChildKeeper.BridgeExecutor(ctx))
}
