package keeper_test

import (
	"errors"
	"fmt"
	"testing"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	testutilsims "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

func Test_UpdateHostValidators(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	hostValidatorStore := input.OPChildKeeper.HostValidatorStore
	_, valPubKeys, cmtValSet := createCmtValidatorSet(t, 20)

	err := hostValidatorStore.UpdateValidators(ctx, 10, cmtValSet)
	require.NoError(t, err)

	for _, valPubKey := range valPubKeys {
		consAddr := sdk.ConsAddress(valPubKey.Address())
		_, err := hostValidatorStore.GetPubKeyByConsAddr(ctx, consAddr)
		require.NoError(t, err)

		_, err = hostValidatorStore.ValidatorByConsAddr(ctx, consAddr)
		require.NoError(t, err)
	}

	lastHeight, err := hostValidatorStore.GetLastHeight(ctx)
	require.NoError(t, err)

	require.Equal(t, lastHeight, int64(10))
}

func Test_GetHostPubKeyByConsAddr(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	hostValidatorStore := input.OPChildKeeper.HostValidatorStore

	valPubKey := testutilsims.CreateTestPubKeys(1)[0]

	val, err := types.NewValidator("validator1", valPubKey, types.Description{})
	require.NoError(t, err)
	cmtPubKey, err := val.CmtConsPublicKey()
	require.NoError(t, err)

	consAddr, err := val.GetConsAddr()
	require.NoError(t, err)

	// should be empty
	_, err = hostValidatorStore.GetPubKeyByConsAddr(ctx, consAddr)
	require.Error(t, err)

	// set validator
	err = hostValidatorStore.SetValidator(ctx, val)
	require.NoError(t, err)

	valPubKeyAfter, err := hostValidatorStore.GetPubKeyByConsAddr(ctx, consAddr)
	require.NoError(t, err)
	require.Equal(t, cmtPubKey, valPubKeyAfter)
}

func Test_HostValidatorByConsAddr(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	hostValidatorStore := input.OPChildKeeper.HostValidatorStore

	valPubKey := testutilsims.CreateTestPubKeys(1)[0]

	val, err := types.NewValidator("validator1", valPubKey, types.Description{})
	require.NoError(t, err)

	consAddr, err := val.GetConsAddr()
	require.NoError(t, err)

	// should be empty
	_, err = hostValidatorStore.ValidatorByConsAddr(ctx, consAddr)
	require.Error(t, err)

	// set validator
	err = hostValidatorStore.SetValidator(ctx, val)
	require.NoError(t, err)

	valAfter, err := hostValidatorStore.ValidatorByConsAddr(ctx, consAddr)
	require.NoError(t, err)
	require.Equal(t, val, valAfter)
}

func Test_DeleteHostValidators(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	hostValidatorStore := input.OPChildKeeper.HostValidatorStore

	valPubKeys := testutilsims.CreateTestPubKeys(20)
	vals := make([]types.Validator, 20)
	for i, valPubKey := range valPubKeys {
		val, err := types.NewValidator(fmt.Sprintf("validator%d", i), valPubKey, types.Description{})
		require.NoError(t, err)
		err = hostValidatorStore.SetValidator(ctx, val)
		require.NoError(t, err)
		vals[i] = val
	}
	// remove validator
	err := hostValidatorStore.DeleteAllValidators(ctx)
	require.NoError(t, err)

	valAfters, err := hostValidatorStore.GetAllValidators(ctx)
	require.NoError(t, err)
	require.Equal(t, len(valAfters), 0)
}

func Test_GetHostAllValidators(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	hostValidatorStore := input.OPChildKeeper.HostValidatorStore

	valPubKeys := testutilsims.CreateTestPubKeys(20)
	vals := make([]types.Validator, 20)
	for i, valPubKey := range valPubKeys {
		val, err := types.NewValidator(fmt.Sprintf("validator%d", i), valPubKey, types.Description{})
		require.NoError(t, err)
		err = hostValidatorStore.SetValidator(ctx, val)
		require.NoError(t, err)
		vals[i] = val
	}

	valAfters, err := hostValidatorStore.GetAllValidators(ctx)
	require.NoError(t, err)

	require.ElementsMatch(t, vals, valAfters)
}

func Test_HostTotalBondedTokens(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	hostValidatorStore := input.OPChildKeeper.HostValidatorStore

	sum := math.ZeroInt()

	valPubKeys := testutilsims.CreateTestPubKeys(20)
	vals := make([]types.Validator, 20)
	for i, valPubKey := range valPubKeys {
		val, err := types.NewValidator(fmt.Sprintf("validator%d", i), valPubKey, types.Description{})
		require.NoError(t, err)
		val.Status = types.Bonded
		val.Tokens = math.NewInt(int64(i + 1))
		err = hostValidatorStore.SetValidator(ctx, val)
		require.NoError(t, err)
		vals[i] = val
		sum = sum.AddRaw(int64(i + 1))
	}

	totalBondedTokens, err := hostValidatorStore.TotalBondedTokens(ctx)
	require.NoError(t, err)

	require.True(t, totalBondedTokens.Equal(sum))
}

func Test_LastHeight(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	hostValidatorStore := input.OPChildKeeper.HostValidatorStore

	_, err := hostValidatorStore.GetLastHeight(ctx)
	require.Error(t, err)
	require.True(t, errors.Is(err, collections.ErrNotFound))

	err = hostValidatorStore.SetLastHeight(ctx, 123)
	require.NoError(t, err)

	lastHeight, err := hostValidatorStore.GetLastHeight(ctx)
	require.NoError(t, err)

	require.Equal(t, lastHeight, int64(123))
}
