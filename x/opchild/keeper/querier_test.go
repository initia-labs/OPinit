package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testutilsims "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/OPinit/x/opchild/keeper"
	"github.com/initia-labs/OPinit/x/opchild/types"
)

func Test_QueryValidator(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	valPubKeys := testutilsims.CreateTestPubKeys(1)
	val, err := types.NewValidator(valAddrs[0], valPubKeys[0], "validator1")
	require.NoError(t, err)

	input.OPChildKeeper.SetValidator(ctx, val)
	q := keeper.NewQuerier(input.OPChildKeeper)

	res, err := q.Validator(ctx, &types.QueryValidatorRequest{ValidatorAddr: val.OperatorAddress})
	require.NoError(t, err)
	require.Equal(t, types.QueryValidatorResponse{Validator: val}, *res)
}

func Test_QueryValidators(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	valPubKeys := testutilsims.CreateTestPubKeys(2)
	val1, err := types.NewValidator(valAddrs[0], valPubKeys[0], "validator1")
	require.NoError(t, err)

	val2, err := types.NewValidator(valAddrs[1], valPubKeys[1], "validator2")
	require.NoError(t, err)
	input.OPChildKeeper.SetValidator(ctx, val1)
	input.OPChildKeeper.SetValidator(ctx, val2)
	q := keeper.NewQuerier(input.OPChildKeeper)

	res, err := q.Validators(ctx, &types.QueryValidatorsRequest{})
	require.NoError(t, err)
	require.Len(t, res.Validators, 2)
}

func Test_QueryParams(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	params, err := input.OPChildKeeper.GetParams(ctx)
	require.NoError(t, err)
	params.MinGasPrices = sdk.NewDecCoins(sdk.NewInt64DecCoin("stake", 1))
	input.OPChildKeeper.SetParams(ctx, params)

	q := keeper.NewQuerier(input.OPChildKeeper)
	res, err := q.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, params, res.Params)
}
