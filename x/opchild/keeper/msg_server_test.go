package keeper_test

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"slices"
	"testing"
	"time"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	testutilsims "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	protoio "github.com/cosmos/gogoproto/io"
	"github.com/cosmos/gogoproto/proto"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"

	cometabci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/skip-mev/connect/v2/abci/strategies/currencypair"
	vetypes "github.com/skip-mev/connect/v2/abci/ve/types"
	connecttypes "github.com/skip-mev/connect/v2/pkg/types"
	oracletypes "github.com/skip-mev/connect/v2/x/oracle/types"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"

	"github.com/initia-labs/OPinit/x/opchild/keeper"
	"github.com/initia-labs/OPinit/x/opchild/types"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
)

/////////////////////////////////////////
// The messages for Validators

func Test_MsgServer_ExecuteMessages(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	params, err := ms.GetParams(ctx)
	require.NoError(t, err)

	// admin to 0
	params.Admin = addrsStr[0]
	require.NoError(t, ms.SetParams(ctx, params))

	valPubKeys := testutilsims.CreateTestPubKeys(2)

	// register validator
	val, err := types.NewValidator(valAddrs[0], valPubKeys[0], "val1")
	require.NoError(t, err)

	err = input.OPChildKeeper.SetValidator(ctx, val)
	require.NoError(t, err)

	// apply validator updates
	_, err = input.OPChildKeeper.BlockValidatorUpdates(ctx)
	require.NoError(t, err)

	moduleAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)

	addMsg, err := types.NewMsgAddAttestor("val2", moduleAddr, valAddrsStr[1], valPubKeys[1])
	require.NoError(t, err)

	removeMsg, err := types.NewMsgRemoveAttestor(moduleAddr, valAddrsStr[1])
	require.NoError(t, err)

	// should failed with unauthorized
	msg, err := types.NewMsgExecuteMessages(addrsStr[1], []sdk.Msg{addMsg})
	require.NoError(t, err)

	_, err = ms.ExecuteMessages(ctx, msg)
	require.Error(t, err)

	// success
	msg, err = types.NewMsgExecuteMessages(addrsStr[0], []sdk.Msg{addMsg})
	require.NoError(t, err)

	_, err = ms.ExecuteMessages(ctx, msg)
	require.NoError(t, err)

	// apply validator updates
	_, err = input.OPChildKeeper.BlockValidatorUpdates(ctx)
	require.NoError(t, err)

	vals, err := input.OPChildKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	require.Len(t, vals, 2)

	// remove attestor
	msg, err = types.NewMsgExecuteMessages(addrsStr[0], []sdk.Msg{removeMsg})
	require.NoError(t, err)

	_, err = ms.ExecuteMessages(ctx, msg)
	require.NoError(t, err)

	// apply validator updates
	_, err = input.OPChildKeeper.BlockValidatorUpdates(ctx)
	require.NoError(t, err)

	vals, err = input.OPChildKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	require.Len(t, vals, 1)
	require.Equal(t, vals[0].Moniker, "val1")

	// should failed with err (denom not sorted)
	params.MinGasPrices = sdk.DecCoins{{
		Denom:  "22222",
		Amount: math.LegacyNewDec(1),
	}, {
		Denom:  "11111",
		Amount: math.LegacyNewDec(2),
	}}
	updateParamsMsg := types.NewMsgUpdateParams(moduleAddr, &params)
	msg, err = types.NewMsgExecuteMessages(addrsStr[0], []sdk.Msg{updateParamsMsg})
	require.NoError(t, err)

	_, err = ms.ExecuteMessages(ctx, msg)
	require.Error(t, err)
}

/////////////////////////////////////////
// The messages for Authority

func Test_MsgServer_AddAttestor(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)
	valPubKeys := testutilsims.CreateTestPubKeys(2)

	moduleAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)

	msg, err := types.NewMsgAddAttestor("val1", moduleAddr, valAddrsStr[0], valPubKeys[0])
	require.NoError(t, err)

	_, err = ms.AddAttestor(ctx, msg)
	require.NoError(t, err)

	// invalid signer
	msg, err = types.NewMsgAddAttestor("val1", addrsStr[0], valAddrsStr[0], valPubKeys[0])
	require.NoError(t, err)

	_, err = ms.AddAttestor(ctx, msg)
	require.Error(t, err)

	// duplicate add attestor
	msg, err = types.NewMsgAddAttestor("val1", moduleAddr, valAddrsStr[0], valPubKeys[1])
	require.NoError(t, err)

	_, err = ms.AddAttestor(ctx, msg)
	require.Error(t, err)

	// duplicate cons pubkey
	msg, err = types.NewMsgAddAttestor("val1", moduleAddr, valAddrsStr[1], valPubKeys[0])
	require.NoError(t, err)

	_, err = ms.AddAttestor(ctx, msg)
	require.Error(t, err)

	params, err := ms.GetParams(ctx)
	require.NoError(t, err)
	params.MaxValidators = 1
	err = ms.SetParams(ctx, params)
	require.NoError(t, err)

	msg, err = types.NewMsgAddAttestor("val2", moduleAddr, valAddrsStr[1], valPubKeys[1])
	require.NoError(t, err)

	// max validators reached
	_, err = ms.AddAttestor(ctx, msg)
	require.Error(t, err)

	params, err = ms.GetParams(ctx)
	require.NoError(t, err)
	params.MaxValidators = 2
	err = ms.SetParams(ctx, params)
	require.NoError(t, err)

	_, err = ms.AddAttestor(ctx, msg)
	require.NoError(t, err)
}

func Test_MsgServer_RemoveAttestor(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	valPubKeys := testutilsims.CreateTestPubKeys(2)

	// register validator
	val, err := types.NewValidator(valAddrs[0], valPubKeys[0], "val1")
	require.NoError(t, err)

	err = input.OPChildKeeper.SetValidator(ctx, val)
	require.NoError(t, err)

	// invalid signer
	msg, err := types.NewMsgRemoveAttestor(addrsStr[0], valAddrsStr[0])
	require.NoError(t, err)

	_, err = ms.RemoveAttestor(
		ctx,
		msg,
	)
	require.Error(t, err)

	moduleAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)

	// remove not existing validator
	msg, err = types.NewMsgRemoveAttestor(moduleAddr, valAddrsStr[1])
	require.NoError(t, err)

	_, err = ms.RemoveAttestor(
		ctx,
		msg,
	)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrNoValidatorFound)

	// remove sequencer should fail
	msg, err = types.NewMsgRemoveAttestor(moduleAddr, valAddrsStr[0])
	require.NoError(t, err)

	_, err = ms.RemoveAttestor(ctx, msg)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrValidatorNotAttestor)

	// register attestor
	addMsg, err := types.NewMsgAddAttestor("val2", moduleAddr, valAddrsStr[1], valPubKeys[1])
	require.NoError(t, err)

	_, err = ms.AddAttestor(ctx, addMsg)
	require.NoError(t, err)

	// valid remove validator
	msg, err = types.NewMsgRemoveAttestor(moduleAddr, valAddrsStr[1])
	require.NoError(t, err)

	_, err = ms.RemoveAttestor(
		ctx,
		msg,
	)
	require.NoError(t, err)
}

func Test_MsgServer_UpdateSequencer(t *testing.T) {
	t.Run("successfully replaces sequencer", func(t *testing.T) {
		ctx, input := createDefaultTestInput(t)
		ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

		valPubKeys := testutilsims.CreateTestPubKeys(2)

		currentSeq, err := types.NewValidator(valAddrs[0], valPubKeys[0], "current-sequencer")
		require.NoError(t, err)
		require.NoError(t, input.OPChildKeeper.SetValidator(ctx, currentSeq))

		moduleAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
		require.NoError(t, err)

		msg, err := types.NewMsgUpdateSequencer("next-sequencer", moduleAddr, valAddrsStr[1], valPubKeys[1])
		require.NoError(t, err)

		_, err = ms.UpdateSequencer(ctx, msg)
		require.NoError(t, err)

		newSeq, found := input.OPChildKeeper.GetValidator(ctx, valAddrs[1])
		require.True(t, found)
		require.Equal(t, "next-sequencer", newSeq.Moniker)
		require.Equal(t, int64(types.SequencerConsPower), newSeq.ConsPower)

		oldSeq, found := input.OPChildKeeper.GetValidator(ctx, valAddrs[0])
		require.True(t, found)
		require.Zero(t, oldSeq.ConsPower)
	})

	t.Run("fails with unauthorized signer", func(t *testing.T) {
		ctx, input := createDefaultTestInput(t)
		ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

		valPubKeys := testutilsims.CreateTestPubKeys(2)

		msg, err := types.NewMsgUpdateSequencer("next-sequencer", addrsStr[0], valAddrsStr[1], valPubKeys[1])
		require.NoError(t, err)

		_, err = ms.UpdateSequencer(ctx, msg)
		require.Error(t, err)
	})

	t.Run("fails when current sequencer missing", func(t *testing.T) {
		ctx, input := createDefaultTestInput(t)
		ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

		valPubKeys := testutilsims.CreateTestPubKeys(1)

		moduleAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
		require.NoError(t, err)

		msg, err := types.NewMsgUpdateSequencer("next-sequencer", moduleAddr, valAddrsStr[0], valPubKeys[0])
		require.NoError(t, err)

		_, err = ms.UpdateSequencer(ctx, msg)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrNoValidatorFound)
	})
}

func Test_MsgServer_AddFeeWhitelistAddresses(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	params, err := ms.GetParams(ctx)
	require.NoError(t, err)
	initialWhitelistCount := len(params.FeeWhitelist)
	newAddresses := []string{addrsStr[1], addrsStr[2]}

	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress("gov"))
	require.NoError(t, err)

	msg := types.NewMsgAddFeeWhitelistAddresses(govAddr, newAddresses)
	require.NoError(t, err)
	_, err = ms.AddFeeWhitelistAddresses(
		ctx,
		msg,
	)
	require.Error(t, err)

	moduleAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)

	msg = types.NewMsgAddFeeWhitelistAddresses(moduleAddr, newAddresses)
	_, err = ms.AddFeeWhitelistAddresses(ctx, msg)
	require.NoError(t, err)

	params, err = ms.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, initialWhitelistCount+2, len(params.FeeWhitelist))
	require.Contains(t, params.FeeWhitelist, addrsStr[1])
	require.Contains(t, params.FeeWhitelist, addrsStr[2])

	_, err = ms.AddFeeWhitelistAddresses(ctx, msg)
	require.NoError(t, err)
	params, err = ms.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, initialWhitelistCount+2, len(params.FeeWhitelist))
}

func Test_MsgServer_RemoveFeeWhitelistAddresses(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	moduleAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)

	addAddresses := []string{addrsStr[1], addrsStr[2], addrsStr[3]}
	addMsg := types.NewMsgAddFeeWhitelistAddresses(moduleAddr, addAddresses)
	_, err = ms.AddFeeWhitelistAddresses(ctx, addMsg)
	require.NoError(t, err)

	params, err := ms.GetParams(ctx)
	require.NoError(t, err)
	countAfterAdd := len(params.FeeWhitelist)

	removeAddresses := []string{addrsStr[1], addrsStr[3]}
	msg := types.NewMsgRemoveFeeWhitelistAddresses(moduleAddr, removeAddresses)
	_, err = ms.RemoveFeeWhitelistAddresses(ctx, msg)
	require.NoError(t, err)

	params, err = ms.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, countAfterAdd-2, len(params.FeeWhitelist))
	require.NotContains(t, params.FeeWhitelist, addrsStr[1])
	require.Contains(t, params.FeeWhitelist, addrsStr[2])
	require.NotContains(t, params.FeeWhitelist, addrsStr[3])

	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress("gov"))
	require.NoError(t, err)

	invalidMsg := types.NewMsgRemoveFeeWhitelistAddresses(govAddr, removeAddresses)
	_, err = ms.RemoveFeeWhitelistAddresses(ctx, invalidMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid authority")
}

func Test_MsgServer_AddBridgeExecutor(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	params, err := ms.GetParams(ctx)
	require.NoError(t, err)
	initialExecutorCount := len(params.BridgeExecutors)

	moduleAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)

	// adding valid addresses
	newAddresses := []string{addrsStr[1], addrsStr[2]}
	msg := types.NewMsgAddBridgeExecutor(moduleAddr, newAddresses)
	_, err = ms.AddBridgeExecutor(ctx, msg)
	require.NoError(t, err)

	params, err = ms.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, initialExecutorCount+2, len(params.BridgeExecutors))
	require.Contains(t, params.BridgeExecutors, addrsStr[1])
	require.Contains(t, params.BridgeExecutors, addrsStr[2])

	// adding duplicate addresses (should not increase count)
	_, err = ms.AddBridgeExecutor(ctx, msg)
	require.NoError(t, err)
	params, err = ms.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, initialExecutorCount+2, len(params.BridgeExecutors))

	// invalid authority
	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress("gov"))
	require.NoError(t, err)

	invalidMsg := types.NewMsgAddBridgeExecutor(govAddr, newAddresses)
	_, err = ms.AddBridgeExecutor(ctx, invalidMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid authority")
}

func Test_MsgServer_RemoveBridgeExecutor(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	moduleAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)

	addAddresses := []string{addrsStr[1], addrsStr[2], addrsStr[3]}
	addMsg := types.NewMsgAddBridgeExecutor(moduleAddr, addAddresses)
	_, err = ms.AddBridgeExecutor(ctx, addMsg)
	require.NoError(t, err)

	params, err := ms.GetParams(ctx)
	require.NoError(t, err)
	countAfterAdd := len(params.BridgeExecutors)

	removeAddresses := []string{addrsStr[1], addrsStr[3]}
	msg := types.NewMsgRemoveBridgeExecutor(moduleAddr, removeAddresses)
	_, err = ms.RemoveBridgeExecutor(ctx, msg)
	require.NoError(t, err)

	params, err = ms.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, countAfterAdd-2, len(params.BridgeExecutors))
	require.NotContains(t, params.BridgeExecutors, addrsStr[1])
	require.Contains(t, params.BridgeExecutors, addrsStr[2])
	require.NotContains(t, params.BridgeExecutors, addrsStr[3])

	// invalid authority
	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress("gov"))
	require.NoError(t, err)

	invalidMsg := types.NewMsgRemoveBridgeExecutor(govAddr, removeAddresses)
	_, err = ms.RemoveBridgeExecutor(ctx, invalidMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid authority")
}

func Test_MsgServer_UpdateMinGasPrices(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	moduleAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)

	minGasPrices := sdk.NewDecCoinsFromCoins(
		sdk.NewCoin("test1", math.NewInt(100)),
		sdk.NewCoin("test2", math.NewInt(10)),
	)

	msg := types.NewMsgUpdateMinGasPrices(moduleAddr, minGasPrices)
	_, err = ms.UpdateMinGasPrices(ctx, msg)
	require.NoError(t, err)

	params, err := ms.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, len(params.MinGasPrices))
	require.Equal(t, math.LegacyNewDec(100), params.MinGasPrices.AmountOf("test1"))
	require.Equal(t, math.LegacyNewDec(10), params.MinGasPrices.AmountOf("test2"))

	// invalid authority
	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress("gov"))
	require.NoError(t, err)

	invalidMsg := types.NewMsgUpdateMinGasPrices(govAddr, minGasPrices)
	_, err = ms.UpdateMinGasPrices(ctx, invalidMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid authority")
}

func Test_MsgServer_UpdateMinGasPrices_EmptyCoins(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	moduleAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)

	initialParams, err := ms.GetParams(ctx)
	require.NoError(t, err)

	initialParams.MinGasPrices = sdk.NewDecCoinsFromCoins(
		sdk.NewCoin("test1", math.NewInt(1)),
	)
	err = input.OPChildKeeper.SetParams(ctx, initialParams)
	require.NoError(t, err)

	emptyMinGasPrices := sdk.DecCoins{}
	msg := types.NewMsgUpdateMinGasPrices(moduleAddr, emptyMinGasPrices)
	_, err = ms.UpdateMinGasPrices(ctx, msg)
	require.NoError(t, err)

	paramsAfter, err := ms.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, len(paramsAfter.MinGasPrices))
}

func Test_MsgServer_UpdateAdmin(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	moduleAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)

	msg := types.NewMsgUpdateAdmin(moduleAddr, addrsStr[1])
	_, err = ms.UpdateAdmin(ctx, msg)
	require.NoError(t, err)

	params, err := ms.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, addrsStr[1], params.Admin)

	// invalid authority
	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress("gov"))
	require.NoError(t, err)

	invalidMsg := types.NewMsgUpdateAdmin(govAddr, addrsStr[2])
	_, err = ms.UpdateAdmin(ctx, invalidMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid authority")
}

func Test_MsgServer_UpdateParams(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	params, err := ms.GetParams(ctx)
	require.NoError(t, err)
	params.MaxValidators = 1
	params.HistoricalEntries = 1
	params.BridgeExecutors = []string{addrs[1].String(), addrs[2].String()}

	moduleAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)

	msg := types.NewMsgUpdateParams(moduleAddr, &params)
	_, err = ms.UpdateParams(ctx, msg)
	require.NoError(t, err)
	_params, err := ms.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, params, _params)

	// invalid signer
	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress("gov"))
	require.NoError(t, err)

	msg = types.NewMsgUpdateParams(govAddr, &params)
	require.NoError(t, err)

	_, err = ms.UpdateParams(
		ctx,
		msg,
	)
	require.Error(t, err)
}

func Test_MsgServer_SpendFeePool(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	// fund fee pool
	collectedFees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100)))
	input.Faucet.Fund(ctx, authtypes.NewModuleAddress(authtypes.FeeCollectorName), collectedFees...)

	beforeAmount := input.BankKeeper.GetBalance(ctx, addrs[1], sdk.DefaultBondDenom).Amount

	msg := types.NewMsgSpendFeePool(
		authtypes.NewModuleAddress(types.ModuleName),
		addrs[1],
		collectedFees,
	)
	_, err := ms.SpendFeePool(ctx, msg)
	require.NoError(t, err)

	afterAmount := input.BankKeeper.GetBalance(ctx, addrs[1], sdk.DefaultBondDenom).Amount
	require.Equal(t, beforeAmount.Add(math.NewInt(100)), afterAmount)
}

/////////////////////////////////////////
// The messages for User

func Test_MsgServer_Withdraw(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

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

	_, err = ms.FinalizeTokenDeposit(ctx, types.NewMsgFinalizeTokenDeposit(addrsStr[0], "anyformataddr", addrsStr[1], sdk.NewCoin(denom, math.NewInt(100)), 1, 1, "test/token", nil))
	require.NoError(t, err)

	coins := sdk.NewCoins(sdk.NewCoin("foo", math.NewInt(1_000_000_000)), sdk.NewCoin(denom, math.NewInt(1_000_000_000)))
	// fund asset
	account := input.Faucet.NewFundedAccount(ctx, coins...)
	accountAddr, err := input.AccountKeeper.AddressCodec().BytesToString(account)
	require.NoError(t, err)

	// not token from l1
	msg := types.NewMsgInitiateTokenWithdrawal(accountAddr, "anyformataddr", sdk.NewCoin("foo", math.NewInt(100)))
	_, err = ms.InitiateTokenWithdrawal(ctx, msg)
	require.Error(t, err)

	// valid
	msg = types.NewMsgInitiateTokenWithdrawal(accountAddr, addrsStr[1], sdk.NewCoin(denom, math.NewInt(100)))
	_, err = ms.InitiateTokenWithdrawal(ctx, msg)
	require.NoError(t, err)
}

/////////////////////////////////////////
// The messages for Bridge Executor

func Test_MsgServer_SetBridgeInfo(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

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

	// reset possible
	_, err = ms.SetBridgeInfo(ctx, types.NewMsgSetBridgeInfo(addrsStr[0], info))
	require.NoError(t, err)

	// cannot change chain id
	info.L1ChainId = "test-chain-id-2"
	_, err = ms.SetBridgeInfo(ctx, types.NewMsgSetBridgeInfo(addrsStr[0], info))
	require.Error(t, err)

	// cannot change client id
	info.L1ChainId = "test-chain-id"
	info.L1ClientId = "test-client-id-2"
	_, err = ms.SetBridgeInfo(ctx, types.NewMsgSetBridgeInfo(addrsStr[0], info))
	require.Error(t, err)

	info.L1ClientId = "test-client-id"

	// invalid bridge id
	info.BridgeId = 0

	_, err = ms.SetBridgeInfo(ctx, types.NewMsgSetBridgeInfo(addrsStr[0], info))
	require.Error(t, err)

	// cannot change bridge id
	info.BridgeId = 2

	_, err = ms.SetBridgeInfo(ctx, types.NewMsgSetBridgeInfo(addrsStr[0], info))
	require.ErrorContains(t, err, "expected bridge id")

	// cannot change bridge addr
	info.BridgeId = 1
	info.BridgeAddr = addrsStr[0]

	_, err = ms.SetBridgeInfo(ctx, types.NewMsgSetBridgeInfo(addrsStr[0], info))
	require.Error(t, err)
	require.ErrorContains(t, err, "expected bridge addr")
}

func Test_MsgServer_Deposit_ToModuleAccount(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	bz := sha3.Sum256([]byte("test_token"))
	denom := "l2/" + hex.EncodeToString(bz[:])

	opchildModuleAddress := authtypes.NewModuleAddress(types.ModuleName)

	beforeToBalance := input.BankKeeper.GetBalance(ctx, addrs[1], denom)
	require.Equal(t, math.ZeroInt(), beforeToBalance.Amount)

	beforeModuleBalance := input.BankKeeper.GetBalance(ctx, opchildModuleAddress, denom)
	require.Equal(t, math.ZeroInt(), beforeModuleBalance.Amount)

	// valid deposit
	msg := types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], opchildModuleAddress.String(), sdk.NewCoin(denom, math.NewInt(100)), 1, 1, "test_token", nil)
	_, err := ms.FinalizeTokenDeposit(ctx, msg)
	require.NoError(t, err)

	for _, event := range sdk.UnwrapSDKContext(ctx).EventManager().Events() {
		if event.Type == types.EventTypeFinalizeTokenDeposit {
			attrIdx := slices.Index(event.Attributes, sdk.NewAttribute(types.AttributeKeySuccess, "false").ToKVPair())
			require.Positive(t, attrIdx)
			require.Equal(t, event.Attributes[attrIdx+1].Key, types.AttributeKeyReason)
			require.Contains(t, event.Attributes[attrIdx+1].Value, "deposit failed;")
		}
	}

	afterToBalance := input.BankKeeper.GetBalance(ctx, addrs[1], denom)
	require.Equal(t, math.ZeroInt(), afterToBalance.Amount)

	afterModuleBalance := input.BankKeeper.GetBalance(ctx, opchildModuleAddress, denom)
	require.True(t, afterModuleBalance.Amount.IsZero())

	// token withdrawal initiated
	events := sdk.UnwrapSDKContext(ctx).EventManager().Events()
	lastEvent := events[len(events)-1]
	require.Equal(t, sdk.NewEvent(
		types.EventTypeInitiateTokenWithdrawal,
		sdk.NewAttribute(types.AttributeKeyFrom, opchildModuleAddress.String()),
		sdk.NewAttribute(types.AttributeKeyTo, addrsStr[1]),
		sdk.NewAttribute(types.AttributeKeyDenom, denom),
		sdk.NewAttribute(types.AttributeKeyBaseDenom, "test_token"),
		sdk.NewAttribute(types.AttributeKeyAmount, "100"),
		sdk.NewAttribute(types.AttributeKeyL2Sequence, "1"),
	), lastEvent)
}

func Test_MsgServer_Deposit_InvalidAddress(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	bz := sha3.Sum256([]byte("test_token"))
	denom := "l2/" + hex.EncodeToString(bz[:])

	beforeToBalance := input.BankKeeper.GetBalance(ctx, addrs[1], denom)
	require.Equal(t, math.ZeroInt(), beforeToBalance.Amount)

	// valid deposit
	msg := types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], "invalid_address", sdk.NewCoin(denom, math.NewInt(100)), 1, 1, "test_token", nil)
	_, err := ms.FinalizeTokenDeposit(ctx, msg)
	require.NoError(t, err)

	for _, event := range sdk.UnwrapSDKContext(ctx).EventManager().Events() {
		if event.Type == types.EventTypeFinalizeTokenDeposit {
			attrIdx := slices.Index(event.Attributes, sdk.NewAttribute(types.AttributeKeySuccess, "false").ToKVPair())
			require.Positive(t, attrIdx)
			require.Equal(t, event.Attributes[attrIdx+1].Key, types.AttributeKeyReason)
			require.Contains(t, event.Attributes[attrIdx+1].Value, "deposit failed;")
		}
	}

	afterToBalance := input.BankKeeper.GetBalance(ctx, addrs[1], denom)
	require.Equal(t, math.ZeroInt(), afterToBalance.Amount)

	// token withdrawal initiated
	events := sdk.UnwrapSDKContext(ctx).EventManager().Events()
	lastEvent := events[len(events)-1]
	require.Equal(t, sdk.NewEvent(
		types.EventTypeInitiateTokenWithdrawal,
		sdk.NewAttribute(types.AttributeKeyFrom, "invalid_address"),
		sdk.NewAttribute(types.AttributeKeyTo, addrsStr[1]),
		sdk.NewAttribute(types.AttributeKeyDenom, denom),
		sdk.NewAttribute(types.AttributeKeyBaseDenom, "test_token"),
		sdk.NewAttribute(types.AttributeKeyAmount, "100"),
		sdk.NewAttribute(types.AttributeKeyL2Sequence, "1"),
	), lastEvent)
}

func Test_MsgServer_Deposit_NoHook(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	bz := sha3.Sum256([]byte("test_token"))
	denom := "l2/" + hex.EncodeToString(bz[:])

	// unauthorized deposit
	msg := types.NewMsgFinalizeTokenDeposit(addrsStr[1], addrsStr[1], addrsStr[1], sdk.NewCoin(denom, math.NewInt(100)), 1, 1, "test_token", nil)
	_, err := ms.FinalizeTokenDeposit(ctx, msg)
	require.Error(t, err)

	beforeBalance := input.BankKeeper.GetBalance(ctx, addrs[1], denom)
	require.Equal(t, math.ZeroInt(), beforeBalance.Amount)

	// valid deposit
	msg = types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], addrsStr[1], sdk.NewCoin(denom, math.NewInt(100)), 1, 1, "test_token", nil)
	_, err = ms.FinalizeTokenDeposit(ctx, msg)
	require.NoError(t, err)
	require.True(t, input.TokenCreationFactory.created[denom])

	afterBalance := input.BankKeeper.GetBalance(ctx, addrs[1], denom)
	require.Equal(t, math.NewInt(100), afterBalance.Amount)
}

func Test_MsgServer_Deposit_HookSuccess(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	bz := sha3.Sum256([]byte("test_token"))
	denom := "l2/" + hex.EncodeToString(bz[:])

	require.Equal(t, math.ZeroInt(), input.BankKeeper.GetBalance(ctx, addrs[1], denom).Amount)

	// empty deposit to create account
	priv, _, addr := keyPubAddr()
	msg := types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], addr.String(), sdk.NewCoin(denom, math.ZeroInt()), 1, 1, "test_token", nil)
	_, err := ms.FinalizeTokenDeposit(ctx, msg)
	require.NoError(t, err)

	// create hook data
	acc := input.AccountKeeper.GetAccount(ctx, addr)
	require.NotNil(t, acc)
	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv}, []uint64{acc.GetAccountNumber()}, []uint64{0}
	signedTxBz, err := input.EncodingConfig.TxConfig.TxEncoder()(generateTestTx(
		t, input,
		[]sdk.Msg{banktypes.NewMsgSend(addr, addrs[2], sdk.NewCoins(sdk.NewCoin(denom, math.NewInt(50))))}, // send half tokens to addrs[2]
		privs, accNums, accSeqs, sdk.UnwrapSDKContext(ctx).ChainID(),
	))
	require.NoError(t, err)

	// valid deposit
	ctx = sdk.UnwrapSDKContext(ctx).WithEventManager(sdk.NewEventManager())
	msg = types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], addr.String(), sdk.NewCoin(denom, math.NewInt(100)), 2, 1, "test_token", signedTxBz)
	_, err = ms.FinalizeTokenDeposit(ctx, msg)
	require.NoError(t, err)

	for _, event := range sdk.UnwrapSDKContext(ctx).EventManager().Events() {
		if event.Type == types.EventTypeFinalizeTokenDeposit {
			require.True(t, slices.Contains(event.Attributes, sdk.NewAttribute(types.AttributeKeySuccess, "true").ToKVPair()))
		}
	}

	// check addrs[2] balance
	afterBalance := input.BankKeeper.GetBalance(ctx, addrs[2], denom)
	require.Equal(t, math.NewInt(50), afterBalance.Amount)
}

func Test_MsgServer_Deposit_HookFail(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	bz := sha3.Sum256([]byte("test_token"))
	denom := "l2/" + hex.EncodeToString(bz[:])

	require.Equal(t, math.ZeroInt(), input.BankKeeper.GetBalance(ctx, addrs[1], denom).Amount)

	// empty deposit to create account
	priv, _, addr := keyPubAddr()
	msg := types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], addr.String(), sdk.NewCoin(denom, math.ZeroInt()), 1, 1, "test_token", nil)
	_, err := ms.FinalizeTokenDeposit(ctx, msg)
	require.NoError(t, err)

	// create hook data
	acc := input.AccountKeeper.GetAccount(ctx, addr)
	require.NotNil(t, acc)
	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv}, []uint64{acc.GetAccountNumber()}, []uint64{0}
	signedTxBz, err := input.EncodingConfig.TxConfig.TxEncoder()(generateTestTx(
		t, input,
		[]sdk.Msg{banktypes.NewMsgSend(addr, addrs[2], sdk.NewCoins(sdk.NewCoin(denom, math.NewInt(101))))}, // send more than deposited tokens
		privs, accNums, accSeqs, sdk.UnwrapSDKContext(ctx).ChainID(),
	))
	require.NoError(t, err)

	// valid deposit
	ctx = sdk.UnwrapSDKContext(ctx).WithEventManager(sdk.NewEventManager())
	msg = types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], addr.String(), sdk.NewCoin(denom, math.NewInt(100)), 2, 1, "test_token", signedTxBz)
	_, err = ms.FinalizeTokenDeposit(ctx, msg)
	require.NoError(t, err)

	for _, event := range sdk.UnwrapSDKContext(ctx).EventManager().Events() {
		if event.Type == types.EventTypeFinalizeTokenDeposit {
			attrIdx := slices.Index(event.Attributes, sdk.NewAttribute(types.AttributeKeySuccess, "false").ToKVPair())
			require.Positive(t, attrIdx)
			require.Equal(t, event.Attributes[attrIdx+1].Key, types.AttributeKeyReason)
			require.Contains(t, event.Attributes[attrIdx+1].Value, "hook failed;")
		}
	}

	// check addrs[2] balance
	afterBalance := input.BankKeeper.GetBalance(ctx, addrs[2], denom)
	require.Equal(t, math.NewInt(0), afterBalance.Amount)

	// check receiver has no balance
	afterBalance = input.BankKeeper.GetBalance(ctx, addr, denom)
	require.Equal(t, math.NewInt(100), afterBalance.Amount)
}

func Test_MsgServer_Deposit_HookFail_WithOutOfGas(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	bz := sha3.Sum256([]byte("test_token"))
	denom := "l2/" + hex.EncodeToString(bz[:])

	require.Equal(t, math.ZeroInt(), input.BankKeeper.GetBalance(ctx, addrs[1], denom).Amount)

	// empty deposit to create account
	priv, _, addr := keyPubAddr()
	msg := types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], addr.String(), sdk.NewCoin(denom, math.ZeroInt()), 1, 1, "test_token", nil)
	_, err := ms.FinalizeTokenDeposit(ctx, msg)
	require.NoError(t, err)

	// create hook data
	acc := input.AccountKeeper.GetAccount(ctx, addr)
	require.NotNil(t, acc)
	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv}, []uint64{acc.GetAccountNumber()}, []uint64{0}
	signedTxBz, err := input.EncodingConfig.TxConfig.TxEncoder()(generateTestTx(
		t, input,
		[]sdk.Msg{banktypes.NewMsgSend(addr, addrs[2], sdk.NewCoins(sdk.NewCoin(denom, math.NewInt(100))))},
		privs, accNums, accSeqs, sdk.UnwrapSDKContext(ctx).ChainID(),
	))
	require.NoError(t, err)

	params, err := input.OPChildKeeper.GetParams(ctx)
	require.NoError(t, err)

	params.HookMaxGas = 1
	input.OPChildKeeper.SetParams(ctx, params)

	// valid deposit
	ctx = sdk.UnwrapSDKContext(ctx).WithEventManager(sdk.NewEventManager()).WithGasMeter(storetypes.NewGasMeter(2_000_000))
	msg = types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], addr.String(), sdk.NewCoin(denom, math.NewInt(100)), 2, 1, "test_token", signedTxBz)
	_, err = ms.FinalizeTokenDeposit(ctx, msg)
	require.NoError(t, err)

	for _, event := range sdk.UnwrapSDKContext(ctx).EventManager().Events() {
		if event.Type == types.EventTypeFinalizeTokenDeposit {
			attrIdx := slices.Index(event.Attributes, sdk.NewAttribute(types.AttributeKeySuccess, "false").ToKVPair())
			require.Positive(t, attrIdx)
			require.Equal(t, event.Attributes[attrIdx+1].Key, types.AttributeKeyReason)
			require.Contains(t, event.Attributes[attrIdx+1].Value, "hook failed;")
			require.Contains(t, event.Attributes[attrIdx+1].Value, "panic:")
		}
	}

	// check addrs[2] balance
	afterBalance := input.BankKeeper.GetBalance(ctx, addrs[2], denom)
	require.Equal(t, math.NewInt(0), afterBalance.Amount)

	// check receiver has no balance
	afterBalance = input.BankKeeper.GetBalance(ctx, addr, denom)
	require.Equal(t, math.NewInt(100), afterBalance.Amount)
}

func Test_MsgServer_UpdateOracle(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	opchildKeeper := input.OPChildKeeper
	oracleKeeper := input.OracleKeeper

	defaultHostChainId := "test-host-1"
	defaultClientId := "test-client-id"
	bridgeInfo := types.BridgeInfo{
		L1ChainId:  defaultHostChainId,
		L1ClientId: defaultClientId,
		BridgeConfig: ophosttypes.BridgeConfig{
			OracleEnabled: true,
		},
	}
	err := opchildKeeper.BridgeInfo.Set(ctx, bridgeInfo)
	require.NoError(t, err)

	oracleKeeper.InitGenesis(sdk.UnwrapSDKContext(ctx), oracletypes.GenesisState{
		CurrencyPairGenesis: make([]oracletypes.CurrencyPairGenesis, 0),
	})

	prices := []map[string]string{
		{"BTC/USD": "10000000", "ETH/USD": "200000", "ATOM/USD": "5000", "TIMESTAMP/NANOSECOND": "10000"},
		{"BTC/USD": "10000000", "ETH/USD": "200000", "ATOM/USD": "5000", "TIMESTAMP/NANOSECOND": "10000"},
		{"BTC/USD": "10000000", "ETH/USD": "210000", "ATOM/USD": "5000", "TIMESTAMP/NANOSECOND": "10000"},
		{"BTC/USD": "11000000", "ETH/USD": "210000", "ATOM/USD": "5000", "TIMESTAMP/NANOSECOND": "10000"},
		{"BTC/USD": "11000000", "ETH/USD": "210000", "ATOM/USD": "5000", "TIMESTAMP/NANOSECOND": "10000"},
	}
	currencyPairs := []string{"ATOM/USD", "ETH/USD", "BTC/USD", "TIMESTAMP/NANOSECOND"}
	numVals := 5

	for _, currencyPair := range currencyPairs {
		cp, err := connecttypes.CurrencyPairFromString(currencyPair)
		require.NoError(t, err)
		err = oracleKeeper.CreateCurrencyPair(sdk.UnwrapSDKContext(ctx), cp)
		require.NoError(t, err)
	}

	cpStrategy, extendedCommitCodec, voteExtensionCodec := getConnect(oracleKeeper)
	valPrivKeys, _, validatorSet := createCmtValidatorSet(t, numVals)
	err = opchildKeeper.UpdateHostValidatorSet(ctx, defaultClientId, 1, validatorSet)
	require.NoError(t, err)

	eci := cometabci.ExtendedCommitInfo{
		Round: 2,
		Votes: make([]cometabci.ExtendedVoteInfo, numVals),
	}

	marshalDelimitedFn := func(msg proto.Message) ([]byte, error) {
		var buf bytes.Buffer
		if err := protoio.NewDelimitedWriter(&buf).WriteMsg(msg); err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	}

	for i, privKey := range valPrivKeys {
		convertedPrices := make(map[uint64][]byte)
		for currencyPairID, priceString := range prices[i] {
			cp, err := connecttypes.CurrencyPairFromString(currencyPairID)
			require.NoError(t, err)
			rawPrice, converted := new(big.Int).SetString(priceString, 10)
			require.True(t, converted)

			sdkCtx := sdk.UnwrapSDKContext(ctx)
			encodedPrice, err := cpStrategy.GetEncodedPrice(sdkCtx, cp, rawPrice)
			require.NoError(t, err)

			id, err := currencypair.CurrencyPairToHashID(currencyPairID)
			require.NoError(t, err)

			convertedPrices[id] = encodedPrice
		}
		ove := vetypes.OracleVoteExtension{
			Prices: convertedPrices,
		}

		exCommitBz, err := voteExtensionCodec.Encode(ove)
		require.NoError(t, err)

		cve := cmtproto.CanonicalVoteExtension{
			Extension: exCommitBz,
			Height:    10,
			Round:     2,
			ChainId:   defaultHostChainId,
		}

		extSignBytes, err := marshalDelimitedFn(&cve)
		require.NoError(t, err)

		signedBytes, err := privKey.Sign(extSignBytes)
		require.NoError(t, err)

		eci.Votes[i] = cometabci.ExtendedVoteInfo{
			Validator: cometabci.Validator{
				Address: validatorSet.Validators[i].Address,
				Power:   1,
			},
			VoteExtension:      exCommitBz,
			ExtensionSignature: signedBytes,
			BlockIdFlag:        cmtproto.BlockIDFlagCommit,
		}
	}

	extCommitBz, err := extendedCommitCodec.Encode(eci)
	require.NoError(t, err)

	ms := keeper.NewMsgServerImpl(&opchildKeeper)
	_, err = ms.UpdateOracle(ctx, types.NewMsgUpdateOracle(addrsStr[0], 11, extCommitBz))
	require.NoError(t, err)

	_, err = ms.UpdateOracle(ctx, types.NewMsgUpdateOracle(addrsStr[1], 11, extCommitBz))
	require.Error(t, err)
}

func Test_MsgServer_UpdateOracleFail(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	opchildKeeper := input.OPChildKeeper

	defaultHostChainId := "test-host-1"
	defaultClientId := "test-client-id"
	bridgeInfo := types.BridgeInfo{
		L1ChainId:  defaultHostChainId,
		L1ClientId: defaultClientId,
		BridgeConfig: ophosttypes.BridgeConfig{
			OracleEnabled: false,
		},
	}
	err := opchildKeeper.BridgeInfo.Set(ctx, bridgeInfo)
	require.NoError(t, err)

	ms := keeper.NewMsgServerImpl(&opchildKeeper)
	_, err = ms.UpdateOracle(ctx, types.NewMsgUpdateOracle(addrsStr[0], 11, []byte{}))
	require.EqualError(t, err, types.ErrOracleDisabled.Error())
}

func Test_MsgServer_RelayOracleData(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	opchildKeeper := input.OPChildKeeper

	bridgeInfo := types.BridgeInfo{
		BridgeId:   1,
		BridgeAddr: addrsStr[0],
		L1ChainId:  "test-host-1",
		L1ClientId: "test-client-id",
		BridgeConfig: ophosttypes.BridgeConfig{
			Challenger:            addrsStr[1],
			Proposer:              addrsStr[2],
			SubmissionInterval:    100,
			FinalizationPeriod:    100,
			SubmissionStartHeight: 1,
			OracleEnabled:         true,
			Metadata:              []byte{},
			BatchInfo:             ophosttypes.BatchInfo{Submitter: addrsStr[0], ChainType: ophosttypes.BatchInfo_INITIA},
		},
	}
	err := opchildKeeper.BridgeInfo.Set(ctx, bridgeInfo)
	require.NoError(t, err)

	// valid oracle data (without proof verification for unit test)
	oracleData := types.OracleData{
		BridgeId:       1,
		CurrencyPair:   "BTC/USD",
		Price:          "50000000000",
		Decimals:       8,
		L1BlockHeight:  100,
		L1BlockTime:    1000000000,
		CurrencyPairId: 1,
		Nonce:          1,
		Proof:          []byte("test-proof"),
		ProofHeight: clienttypes.Height{
			RevisionNumber: 0,
			RevisionHeight: 100,
		},
	}

	ms := keeper.NewMsgServerImpl(&opchildKeeper)

	msg := types.NewMsgRelayOracleData(addrsStr[0], oracleData)
	_, err = ms.RelayOracleData(ctx, msg)
	require.Error(t, err) // expected to fail at proof verification
	require.Contains(t, err.Error(), "failed to handle oracle data")
}

func Test_MsgServer_RelayOracleData_InvalidMessage(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	opchildKeeper := input.OPChildKeeper

	ms := keeper.NewMsgServerImpl(&opchildKeeper)

	invalidOracleData := types.OracleData{
		BridgeId:      0, // invalid
		CurrencyPair:  "BTC/USD",
		Price:         "50000",
		L1BlockHeight: 100,
		L1BlockTime:   1000,
		Proof:         []byte("proof"),
		ProofHeight: clienttypes.Height{
			RevisionHeight: 100,
		},
	}

	msg := types.NewMsgRelayOracleData(addrsStr[0], invalidOracleData)
	_, err := ms.RelayOracleData(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "bridge id cannot be zero")
}

func Test_MsgServer_RelayOracleData_OracleDisabled(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	opchildKeeper := input.OPChildKeeper

	bridgeInfo := types.BridgeInfo{
		BridgeId:   1,
		BridgeAddr: addrsStr[0],
		L1ChainId:  "test-host-1",
		L1ClientId: "test-client-id",
		BridgeConfig: ophosttypes.BridgeConfig{
			Challenger:            addrsStr[1],
			Proposer:              addrsStr[2],
			SubmissionInterval:    100,
			FinalizationPeriod:    100,
			SubmissionStartHeight: 1,
			OracleEnabled:         false, // disabled
			Metadata:              []byte{},
			BatchInfo:             ophosttypes.BatchInfo{Submitter: addrsStr[0], ChainType: ophosttypes.BatchInfo_INITIA},
		},
	}
	err := opchildKeeper.BridgeInfo.Set(ctx, bridgeInfo)
	require.NoError(t, err)

	oracleData := types.OracleData{
		BridgeId:       1,
		CurrencyPair:   "BTC/USD",
		Price:          "50000000000",
		Decimals:       8,
		L1BlockHeight:  100,
		L1BlockTime:    1000000000,
		CurrencyPairId: 1,
		Nonce:          1,
		Proof:          []byte("test-proof"),
		ProofHeight: clienttypes.Height{
			RevisionNumber: 0,
			RevisionHeight: 100,
		},
	}

	ms := keeper.NewMsgServerImpl(&opchildKeeper)
	msg := types.NewMsgRelayOracleData(addrsStr[0], oracleData)
	_, err = ms.RelayOracleData(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "oracle is disabled")
}
