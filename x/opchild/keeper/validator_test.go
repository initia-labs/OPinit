package keeper_test

import (
	"testing"

	testutilsims "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/initia-labs/OPinit/v1/x/opchild/types"
	"github.com/stretchr/testify/require"
)

func Test_GetValidator(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	valPubKey := testutilsims.CreateTestPubKeys(1)[0]
	val, err := types.NewValidator(valAddrs[1], valPubKey, "validator1")
	require.NoError(t, err)

	// should be empty
	_, found := input.OPChildKeeper.GetValidator(ctx, valAddrs[1])
	require.False(t, found)

	// set validator
	require.NoError(t, input.OPChildKeeper.SetValidator(ctx, val))

	valAfter, found := input.OPChildKeeper.GetValidator(ctx, valAddrs[1])
	require.True(t, found)
	require.Equal(t, val, valAfter)

	// remove validator
	require.NoError(t, input.OPChildKeeper.RemoveValidator(ctx, valAddrs[1]))

	// should be empty
	_, found = input.OPChildKeeper.GetValidator(ctx, valAddrs[1])
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
	require.NoError(t, input.OPChildKeeper.SetValidator(ctx, val))
	require.NoError(t, input.OPChildKeeper.SetValidatorByConsAddr(ctx, val))

	valAfter, found := input.OPChildKeeper.GetValidatorByConsAddr(ctx, consAddr)
	require.True(t, found)
	require.Equal(t, val, valAfter)

	// remove validator
	require.NoError(t, input.OPChildKeeper.RemoveValidator(ctx, valAddrs[1]))

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

	require.NoError(t, input.OPChildKeeper.SetValidator(ctx, val1))
	require.NoError(t, input.OPChildKeeper.SetValidator(ctx, val2))

	vals, err := input.OPChildKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
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

	require.NoError(t, input.OPChildKeeper.SetValidator(ctx, val1))
	require.NoError(t, input.OPChildKeeper.SetValidator(ctx, val2))

	vals, err := input.OPChildKeeper.GetValidators(ctx, 1)
	require.NoError(t, err)
	require.Len(t, vals, 1)
}

func Test_LastValidatorPower(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	valPubKeys := testutilsims.CreateTestPubKeys(2)
	val1, err := types.NewValidator(valAddrs[1], valPubKeys[0], "validator1")
	require.NoError(t, err)

	val2, err := types.NewValidator(valAddrs[2], valPubKeys[1], "validator2")
	require.NoError(t, err)

	beforePower, err := input.OPChildKeeper.GetLastValidatorPower(ctx, valAddrs[1])
	require.NoError(t, err)
	require.Equal(t, int64(0), beforePower)

	// set validator with power index
	require.NoError(t, input.OPChildKeeper.SetValidator(ctx, val1))
	require.NoError(t, input.OPChildKeeper.SetValidator(ctx, val2))
	require.NoError(t, input.OPChildKeeper.SetLastValidatorPower(ctx, valAddrs[1], 100))
	require.NoError(t, input.OPChildKeeper.SetLastValidatorPower(ctx, valAddrs[2], 200))

	afterPower, err := input.OPChildKeeper.GetLastValidatorPower(ctx, valAddrs[1])
	require.NoError(t, err)
	require.Equal(t, int64(100), afterPower)

	// iterate all powers
	require.NoError(t, input.OPChildKeeper.IterateLastValidatorPowers(ctx, func(key []byte, power int64) (stop bool, err error) {
		valAddr := sdk.ValAddress(key)
		if valAddr.Equals(valAddrs[1]) {
			require.Equal(t, int64(100), power)
		} else {
			require.Equal(t, valAddrs[2], valAddr)
			require.Equal(t, int64(200), power)
		}

		return false, nil
	}))

	// get last validators from the power index
	vals, err := input.OPChildKeeper.GetLastValidators(ctx)
	require.NoError(t, err)
	require.Len(t, vals, 2)
	require.Contains(t, vals, val1)
	require.Contains(t, vals, val2)

	// decrease max validator to 1
	params, err := input.OPChildKeeper.GetParams(ctx)
	require.NoError(t, err)
	params.MaxValidators = 1
	err = input.OPChildKeeper.SetParams(ctx, params)
	require.Error(t, err)
}
