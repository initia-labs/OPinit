package keeper_test

import (
	"testing"

	testutilsims "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/initia-labs/OPinit/x/op_child/types"
	"github.com/stretchr/testify/require"
)

func Test_GetValidator(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	valPubKey := testutilsims.CreateTestPubKeys(1)[0]
	val, err := types.NewValidator(valAddrs[1], valPubKey, "validator1")
	require.NoError(t, err)

	// should be empty
	_, found := input.OPChildKeeper.GetValidator(ctx, val.GetOperator())
	require.False(t, found)

	// set validator
	input.OPChildKeeper.SetValidator(ctx, val)

	valAfter, found := input.OPChildKeeper.GetValidator(ctx, val.GetOperator())
	require.True(t, found)
	require.Equal(t, val, valAfter)

	// remove validator
	input.OPChildKeeper.RemoveValidator(ctx, val.GetOperator())

	// should be empty
	_, found = input.OPChildKeeper.GetValidator(ctx, val.GetOperator())
	require.False(t, found)
}

func Test_GetValidatorByConsAddr(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	valPubKey := testutilsims.CreateTestPubKeys(1)[0]
	val, err := types.NewValidator(valAddrs[1], valPubKey, "validator1")
	require.NoError(t, err)

	consAddr, err := val.GetConsAddr()
	require.NoError(t, err)

	// should be empty
	_, found := input.OPChildKeeper.GetValidatorByConsAddr(ctx, consAddr)
	require.False(t, found)

	// set validator
	input.OPChildKeeper.SetValidator(ctx, val)
	input.OPChildKeeper.SetValidatorByConsAddr(ctx, val)

	valAfter, found := input.OPChildKeeper.GetValidatorByConsAddr(ctx, consAddr)
	require.True(t, found)
	require.Equal(t, val, valAfter)

	// remove validator
	input.OPChildKeeper.RemoveValidator(ctx, val.GetOperator())

	// should be empty
	_, found = input.OPChildKeeper.GetValidatorByConsAddr(ctx, consAddr)
	require.False(t, found)
}

func Test_GetAllValidators(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	valPubKeys := testutilsims.CreateTestPubKeys(2)
	val1, err := types.NewValidator(valAddrs[1], valPubKeys[0], "validator1")
	require.NoError(t, err)

	val2, err := types.NewValidator(valAddrs[2], valPubKeys[1], "validator2")
	require.NoError(t, err)

	input.OPChildKeeper.SetValidator(ctx, val1)
	input.OPChildKeeper.SetValidator(ctx, val2)

	vals := input.OPChildKeeper.GetAllValidators(ctx)
	require.Len(t, vals, 2)
	require.Contains(t, vals, val1)
	require.Contains(t, vals, val2)
}

func Test_GetValidators(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	valPubKeys := testutilsims.CreateTestPubKeys(2)
	val1, err := types.NewValidator(valAddrs[1], valPubKeys[0], "validator1")
	require.NoError(t, err)

	val2, err := types.NewValidator(valAddrs[2], valPubKeys[1], "validator2")
	require.NoError(t, err)

	input.OPChildKeeper.SetValidator(ctx, val1)
	input.OPChildKeeper.SetValidator(ctx, val2)

	vals := input.OPChildKeeper.GetValidators(ctx, 1)
	require.Len(t, vals, 1)
}

func Test_LastValidatorPower(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	valPubKeys := testutilsims.CreateTestPubKeys(2)
	val1, err := types.NewValidator(valAddrs[1], valPubKeys[0], "validator1")
	require.NoError(t, err)

	val2, err := types.NewValidator(valAddrs[2], valPubKeys[1], "validator2")
	require.NoError(t, err)

	beforePower := input.OPChildKeeper.GetLastValidatorPower(ctx, val1.GetOperator())
	require.Equal(t, int64(0), beforePower)

	// set validator with power index
	input.OPChildKeeper.SetValidator(ctx, val1)
	input.OPChildKeeper.SetValidator(ctx, val2)
	input.OPChildKeeper.SetLastValidatorPower(ctx, val1.GetOperator(), 100)
	input.OPChildKeeper.SetLastValidatorPower(ctx, val2.GetOperator(), 200)

	afterPower := input.OPChildKeeper.GetLastValidatorPower(ctx, val1.GetOperator())
	require.Equal(t, int64(100), afterPower)

	// iterate all powers
	input.OPChildKeeper.IterateLastValidatorPowers(ctx, func(valAddr sdk.ValAddress, power int64) bool {
		if valAddr.Equals(val1.GetOperator()) {
			require.Equal(t, int64(100), power)
		} else {
			require.Equal(t, val2.GetOperator(), val2.GetOperator())
			require.Equal(t, int64(200), power)
		}

		return false
	})

	// get last validators from the power index
	vals := input.OPChildKeeper.GetLastValidators(ctx)
	require.Len(t, vals, 2)
	require.Contains(t, vals, val1)
	require.Contains(t, vals, val2)

	// decrease max validator to 1
	params := input.OPChildKeeper.GetParams(ctx)
	params.MaxValidators = 1
	input.OPChildKeeper.SetParams(ctx, params)

	// should panic if there is more than 1 validators
	require.Panics(t, func() {
		_ = input.OPChildKeeper.GetLastValidators(ctx)
	})
}
