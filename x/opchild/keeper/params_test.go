package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/OPinit/x/opchild/testutil"
	"github.com/initia-labs/OPinit/x/opchild/types"
	"github.com/stretchr/testify/require"
)

func Test_Params(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	params, err := input.OPChildKeeper.GetParams(ctx)
	require.NoError(t, err)
	params.MinGasPrices = sdk.NewDecCoins()
	params.FeeWhitelist = []string{testutil.AddrsStr[0], testutil.AddrsStr[1]}

	err = input.OPChildKeeper.SetParams(ctx, params)
	require.NoError(t, err)

	minGasPrices, err := input.OPChildKeeper.MinGasPrices(ctx)
	require.NoError(t, err)
	require.True(t, minGasPrices.Empty())
	bridgeExecutors, err := input.OPChildKeeper.BridgeExecutors(ctx)
	require.NoError(t, err)
	require.Equal(t, []sdk.AccAddress{testutil.Addrs[0]}, bridgeExecutors)

	feeWhitelist, err := input.OPChildKeeper.FeeWhitelist(ctx)
	require.NoError(t, err)
	require.Equal(t, params.FeeWhitelist, feeWhitelist)
}

func Test_Change_MaxValidators(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	params, err := input.OPChildKeeper.GetParams(ctx)
	require.NoError(t, err)

	// cannot be zero
	params.MaxValidators = 0
	err = input.OPChildKeeper.SetParams(ctx, params)
	require.Error(t, err)

	params.MaxValidators = 2
	err = input.OPChildKeeper.SetParams(ctx, params)
	require.NoError(t, err)

	err = input.OPChildKeeper.Validators.Set(ctx, []byte{0}, types.Validator{})
	require.NoError(t, err)
	err = input.OPChildKeeper.Validators.Set(ctx, []byte{1}, types.Validator{})
	require.NoError(t, err)

	// cannot be lower than current number of validators
	params.MaxValidators = 1
	err = input.OPChildKeeper.SetParams(ctx, params)
	require.Error(t, err)
}
