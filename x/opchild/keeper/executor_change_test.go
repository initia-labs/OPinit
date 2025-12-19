package keeper_test

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	testutilsims "github.com/cosmos/cosmos-sdk/testutil/sims"

	"github.com/initia-labs/OPinit/x/opchild/testutil"
	"github.com/initia-labs/OPinit/x/opchild/types"
	"github.com/stretchr/testify/require"
)

func Test_RegisterExecutorChangePlan(t *testing.T) {
	// Setup
	_, input := testutil.CreateTestInput(t, false)

	// Arguments
	l1ProposalID, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	require.NoError(t, err)
	height, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	require.NoError(t, err)
	nextValAddr := testutil.ValAddrsStr[0]
	nextExecutorAddr := []string{testutil.AddrsStr[0], testutil.AddrsStr[1]}
	consensusPubKey := "l7aqGv+Zjbm0rallfqfqz+3iN31iOmgJCafWV5pGs6o="
	moniker := "moniker"
	info := "info"

	err = input.OPChildKeeper.RegisterExecutorChangePlan(
		l1ProposalID.Uint64(), height.Uint64(), nextValAddr,
		moniker,
		fmt.Sprintf(`{"@type":"/cosmos.crypto.ed25519.PubKey","key":"%s"}`, consensusPubKey),
		info,
		nextExecutorAddr,
	)
	require.NoError(t, err)
	require.Len(t, input.OPChildKeeper.ExecutorChangePlans, 1)

	consensusPubKeyBytes, err := base64.StdEncoding.DecodeString(consensusPubKey)
	require.NoError(t, err)

	expectedValidator, err := types.NewValidator(testutil.ValAddrs[0], &ed25519.PubKey{
		Key: consensusPubKeyBytes,
	}, moniker)
	require.NoError(t, err)
	require.Equal(t, types.ExecutorChangePlan{
		ProposalID:    l1ProposalID.Uint64(),
		Height:        height.Uint64(),
		NextExecutors: []string{testutil.AddrsStr[0], testutil.AddrsStr[1]},
		NextValidator: expectedValidator,
		Info:          info,
	}, input.OPChildKeeper.ExecutorChangePlans[height.Uint64()])
}

func Test_ExecuteChangePlan(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	valPubKeys := testutilsims.CreateTestPubKeys(2)
	val1, err := types.NewValidator(testutil.ValAddrs[1], valPubKeys[0], "validator1")
	require.NoError(t, err)

	val2, err := types.NewValidator(testutil.ValAddrs[2], valPubKeys[1], "validator2")
	require.NoError(t, err)

	// set validators
	require.NoError(t, input.OPChildKeeper.SetValidator(ctx, val1))
	require.NoError(t, input.OPChildKeeper.SetValidatorByConsAddr(ctx, val1))
	require.NoError(t, input.OPChildKeeper.SetValidator(ctx, val2))
	require.NoError(t, input.OPChildKeeper.SetValidatorByConsAddr(ctx, val2))

	// Arguments
	l1ProposalID, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	require.NoError(t, err)
	height, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	require.NoError(t, err)
	nextValAddr := testutil.ValAddrsStr[0]
	nextExecutorAddr := []string{testutil.AddrsStr[0], testutil.AddrsStr[1]}
	consensusPubKey := "l7aqGv+Zjbm0rallfqfqz+3iN31iOmgJCafWV5pGs6o="
	moniker := "moniker"
	info := "info"

	err = input.OPChildKeeper.RegisterExecutorChangePlan(
		l1ProposalID.Uint64(), height.Uint64(), nextValAddr,
		moniker,
		fmt.Sprintf(`{"@type":"/cosmos.crypto.ed25519.PubKey","key":"%s"}`, consensusPubKey),
		info,
		nextExecutorAddr,
	)
	require.NoError(t, err)
	require.Len(t, input.OPChildKeeper.ExecutorChangePlans, 1)

	err = input.OPChildKeeper.ChangeExecutor(ctx, input.OPChildKeeper.ExecutorChangePlans[height.Uint64()])
	require.NoError(t, err)

	// Check if the validator has been updated
	validator, found := input.OPChildKeeper.GetValidator(ctx, testutil.ValAddrs[0])
	require.True(t, found)
	require.Equal(t, moniker, validator.GetMoniker())
	require.Equal(t, int64(1), validator.ConsPower)

	consAddr, err := validator.GetConsAddr()
	require.NoError(t, err)
	v := input.OPChildKeeper.ValidatorByConsAddr(ctx, consAddr)
	require.Equal(t, moniker, v.GetMoniker())
	require.Equal(t, int64(1), v.GetConsensusPower())

	// Check if the old validators have been updated
	validator, found = input.OPChildKeeper.GetValidator(ctx, testutil.ValAddrs[1])
	require.True(t, found)
	require.Equal(t, int64(0), validator.ConsPower)

	validator, found = input.OPChildKeeper.GetValidator(ctx, testutil.ValAddrs[2])
	require.True(t, found)
	require.Equal(t, int64(0), validator.ConsPower)
}
