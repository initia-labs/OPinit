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

	"github.com/initia-labs/OPinit/v1/x/opchild/types"
	"github.com/stretchr/testify/require"
)

func Test_RegisterExecutorChangePlan(t *testing.T) {
	// Setup
	_, input := createTestInput(t, false)

	// Arguments
	l1ProposalID, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	require.NoError(t, err)
	height, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	require.NoError(t, err)
	nextValAddr := valAddrsStr[0]
	nextExecutorAddr := []string{addrsStr[0], addrsStr[1]}
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

	expectedValidator, err := types.NewValidator(valAddrs[0], &ed25519.PubKey{
		Key: consensusPubKeyBytes,
	}, moniker)
	require.NoError(t, err)
	require.Equal(t, types.ExecutorChangePlan{
		ProposalID:    l1ProposalID.Uint64(),
		Height:        height.Uint64(),
		NextExecutors: []string{addrsStr[0], addrsStr[1]},
		NextValidator: expectedValidator,
		Info:          info,
	}, input.OPChildKeeper.ExecutorChangePlans[height.Uint64()])
}

func Test_ExecuteChangePlan(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	valPubKeys := testutilsims.CreateTestPubKeys(2)
	val1, err := types.NewValidator(valAddrs[1], valPubKeys[0], "validator1")
	require.NoError(t, err)

	val2, err := types.NewValidator(valAddrs[2], valPubKeys[1], "validator2")
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
	nextValAddr := valAddrsStr[0]
	nextExecutorAddr := []string{addrsStr[0], addrsStr[1]}
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
	validator, found := input.OPChildKeeper.GetValidator(ctx, valAddrs[0])
	require.True(t, found)
	require.Equal(t, moniker, validator.GetMoniker())
	require.Equal(t, int64(1), validator.ConsPower)

	consAddr, err := validator.GetConsAddr()
	require.NoError(t, err)
	v := input.OPChildKeeper.ValidatorByConsAddr(ctx, consAddr)
	require.Equal(t, moniker, v.GetMoniker())
	require.Equal(t, int64(1), v.GetConsensusPower())

	// Check if the old validators have been updated
	validator, found = input.OPChildKeeper.GetValidator(ctx, valAddrs[1])
	require.True(t, found)
	require.Equal(t, int64(0), validator.ConsPower)

	validator, found = input.OPChildKeeper.GetValidator(ctx, valAddrs[2])
	require.True(t, found)
	require.Equal(t, int64(0), validator.ConsPower)
}
