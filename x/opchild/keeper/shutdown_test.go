package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	testutilsims "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/OPinit/x/opchild/keeper"
	"github.com/initia-labs/OPinit/x/opchild/types"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
	"github.com/stretchr/testify/require"
)

func TestShutdown(t *testing.T) {
	ctx, keepers := createTestInput(t, false)
	ms := keeper.NewMsgServerImpl(&keepers.OPChildKeeper)

	info := types.BridgeInfo{
		BridgeId:   1,
		BridgeAddr: addrsStr[1],
		L1ChainId:  "test-chain-id",
		L1ClientId: "test-client-id",
		BridgeConfig: ophosttypes.BridgeConfig{
			Challenger: addrsStr[2],
			Proposer:   addrsStr[3],
			BatchInfo: ophosttypes.BatchInfo{
				Submitter: addrsStr[4],
				ChainType: ophosttypes.BatchInfo_INITIA,
			},
			SubmissionInterval:    time.Minute,
			FinalizationPeriod:    time.Hour,
			SubmissionStartHeight: 1,
			Metadata:              []byte("metadata"),
		},
	}

	_, err := ms.SetBridgeInfo(ctx, types.NewMsgSetBridgeInfo(addrsStr[0], info))
	require.NoError(t, err)

	baseDenom := "test_token"
	denom := ophosttypes.L2Denom(1, baseDenom)

	_, err = ms.FinalizeTokenDeposit(ctx, types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], addrsStr[1], sdk.NewCoin(denom, math.NewInt(100)), 1, 1, baseDenom, nil))
	require.NoError(t, err)

	baseDenom2 := "test_token2"
	denom2 := ophosttypes.L2Denom(1, baseDenom2)

	_, err = ms.FinalizeTokenDeposit(ctx, types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[2], addrsStr[3], sdk.NewCoin(denom2, math.NewInt(200)), 2, 1, baseDenom2, nil))
	require.NoError(t, err)

	account0 := keepers.Faucet.NewFundedAccount(ctx, sdk.NewCoin(testDenoms[0], math.NewInt(100)))
	account1 := keepers.Faucet.NewFundedAccount(ctx, sdk.NewCoin(testDenoms[1], math.NewInt(200)))
	account2 := keepers.Faucet.NewFundedAccount(ctx, sdk.NewCoin(testDenoms[2], math.NewInt(300)))

	valPubKeys := testutilsims.CreateTestPubKeys(1)
	val1, err := types.NewValidator(valAddrs[1], valPubKeys[0], "validator1")
	require.NoError(t, err)

	// set validators
	require.NoError(t, keepers.OPChildKeeper.SetValidator(ctx, val1))
	require.NoError(t, keepers.OPChildKeeper.SetValidatorByConsAddr(ctx, val1))

	err = keepers.OPChildKeeper.Shutdown(ctx)
	require.NoError(t, err)

	vals, err := keepers.OPChildKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	require.Len(t, vals, 0)

	balances := keepers.BankKeeper.GetAllBalances(ctx, addrs[1])
	require.Len(t, balances, 0)

	balances = keepers.BankKeeper.GetAllBalances(ctx, addrs[2])
	require.Len(t, balances, 0)

	nextL2Sequence, err := keepers.OPChildKeeper.NextL2Sequence.Peek(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(3), nextL2Sequence)

	balances = keepers.BankKeeper.GetAllBalances(ctx, account0)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin(testDenoms[0], math.NewInt(100))), balances)

	balances = keepers.BankKeeper.GetAllBalances(ctx, account1)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin(testDenoms[1], math.NewInt(200))), balances)

	balances = keepers.BankKeeper.GetAllBalances(ctx, account2)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin(testDenoms[2], math.NewInt(300))), balances)
}

func TestShutdown_BridgeDisabled(t *testing.T) {
	ctx, keepers := createTestInput(t, false)

	disabled, err := keepers.OPChildKeeper.IsBridgeDisabled(ctx)
	require.NoError(t, err)
	require.False(t, disabled)

	info := types.BridgeInfo{
		BridgeId:   1,
		BridgeAddr: addrsStr[1],
		L1ChainId:  "test-chain-id",
		L1ClientId: "test-client-id",
		BridgeConfig: ophosttypes.BridgeConfig{
			Challenger: addrsStr[2],
			Proposer:   addrsStr[3],
			BatchInfo: ophosttypes.BatchInfo{
				Submitter: addrsStr[4],
				ChainType: ophosttypes.BatchInfo_INITIA,
			},
			SubmissionInterval:    time.Minute,
			FinalizationPeriod:    time.Hour,
			SubmissionStartHeight: 1,
			Metadata:              []byte("metadata"),
			BridgeDisabled:        false,
		},
	}

	ms := keeper.NewMsgServerImpl(&keepers.OPChildKeeper)
	_, err = ms.SetBridgeInfo(ctx, types.NewMsgSetBridgeInfo(addrsStr[0], info))
	require.NoError(t, err)

	disabled, err = keepers.OPChildKeeper.IsBridgeDisabled(ctx)
	require.NoError(t, err)
	require.False(t, disabled)

	info = types.BridgeInfo{
		BridgeId:   1,
		BridgeAddr: addrsStr[1],
		L1ChainId:  "test-chain-id",
		L1ClientId: "test-client-id",
		BridgeConfig: ophosttypes.BridgeConfig{
			Challenger: addrsStr[2],
			Proposer:   addrsStr[3],
			BatchInfo: ophosttypes.BatchInfo{
				Submitter: addrsStr[4],
				ChainType: ophosttypes.BatchInfo_INITIA,
			},
			SubmissionInterval:    time.Minute,
			FinalizationPeriod:    time.Hour,
			SubmissionStartHeight: 1,
			Metadata:              []byte("metadata"),
			BridgeDisabled:        true,
			BridgeDisabledAt:      time.Now(),
		},
	}

	_, err = ms.SetBridgeInfo(ctx, types.NewMsgSetBridgeInfo(addrsStr[0], info))
	require.NoError(t, err)

	disabled, err = keepers.OPChildKeeper.IsBridgeDisabled(ctx)
	require.NoError(t, err)
	require.True(t, disabled)
}
