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
	"github.com/initia-labs/OPinit/x/opchild/testutil"
	"github.com/initia-labs/OPinit/x/opchild/types"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
)

func mustDecodeHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

/////////////////////////////////////////
// The messages for Validators

func Test_MsgServer_ExecuteMessages(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	params, err := ms.GetParams(ctx)
	require.NoError(t, err)

	// admin to 0
	params.Admin = testutil.AddrsStr[0]
	require.NoError(t, ms.SetParams(ctx, params))

	valPubKeys := testutilsims.CreateTestPubKeys(2)

	// register validator
	val, err := types.NewValidator(testutil.ValAddrs[0], valPubKeys[0], "val1")
	require.NoError(t, err)

	err = input.OPChildKeeper.SetValidator(ctx, val)
	require.NoError(t, err)

	// apply validator updates
	_, err = input.OPChildKeeper.BlockValidatorUpdates(ctx)
	require.NoError(t, err)

	moduleAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)

	vals, err := input.OPChildKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	require.Len(t, vals, 1)

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
	msg, err := types.NewMsgExecuteMessages(testutil.AddrsStr[0], []sdk.Msg{updateParamsMsg})
	require.NoError(t, err)

	_, err = ms.ExecuteMessages(ctx, msg)
	require.Error(t, err)
}

/////////////////////////////////////////
// The messages for Authority

func Test_MsgServer_UpdateSequencer(t *testing.T) {
	t.Run("successfully replaces sequencer", func(t *testing.T) {
		ctx, input := testutil.CreateTestInput(t, false)
		ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

		valPubKeys := testutilsims.CreateTestPubKeys(2)

		currentSeq, err := types.NewValidator(testutil.ValAddrs[0], valPubKeys[0], "current-sequencer")
		require.NoError(t, err)
		require.NoError(t, input.OPChildKeeper.SetValidator(ctx, currentSeq))

		moduleAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
		require.NoError(t, err)

		msg, err := types.NewMsgUpdateSequencer("next-sequencer", moduleAddr, testutil.ValAddrsStr[1], valPubKeys[1])
		require.NoError(t, err)

		_, err = ms.UpdateSequencer(ctx, msg)
		require.NoError(t, err)

		newSeq, found := input.OPChildKeeper.GetValidator(ctx, testutil.ValAddrs[1])
		require.True(t, found)
		require.Equal(t, "next-sequencer", newSeq.Moniker)
		require.Equal(t, int64(types.SequencerConsPower), newSeq.ConsPower)

		oldSeq, found := input.OPChildKeeper.GetValidator(ctx, testutil.ValAddrs[0])
		require.True(t, found)
		require.Zero(t, oldSeq.ConsPower)
	})

	t.Run("fails with unauthorized signer", func(t *testing.T) {
		ctx, input := testutil.CreateTestInput(t, false)
		ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

		valPubKeys := testutilsims.CreateTestPubKeys(2)

		msg, err := types.NewMsgUpdateSequencer("next-sequencer", testutil.AddrsStr[0], testutil.ValAddrsStr[1], valPubKeys[1])
		require.NoError(t, err)

		_, err = ms.UpdateSequencer(ctx, msg)
		require.Error(t, err)
	})

	t.Run("fails when current sequencer missing", func(t *testing.T) {
		ctx, input := testutil.CreateTestInput(t, false)
		ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

		valPubKeys := testutilsims.CreateTestPubKeys(1)

		moduleAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
		require.NoError(t, err)

		msg, err := types.NewMsgUpdateSequencer("next-sequencer", moduleAddr, testutil.ValAddrsStr[0], valPubKeys[0])
		require.NoError(t, err)

		_, err = ms.UpdateSequencer(ctx, msg)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrNoValidatorFound)
	})
}

func Test_MsgServer_AddFeeWhitelistAddresses(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	params, err := ms.GetParams(ctx)
	require.NoError(t, err)
	initialWhitelistCount := len(params.FeeWhitelist)
	newAddresses := []string{testutil.AddrsStr[1], testutil.AddrsStr[2]}

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
	require.Contains(t, params.FeeWhitelist, testutil.AddrsStr[1])
	require.Contains(t, params.FeeWhitelist, testutil.AddrsStr[2])

	_, err = ms.AddFeeWhitelistAddresses(ctx, msg)
	require.NoError(t, err)
	params, err = ms.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, initialWhitelistCount+2, len(params.FeeWhitelist))
}

func Test_MsgServer_RemoveFeeWhitelistAddresses(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	moduleAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)

	addAddresses := []string{testutil.AddrsStr[1], testutil.AddrsStr[2], testutil.AddrsStr[3]}
	addMsg := types.NewMsgAddFeeWhitelistAddresses(moduleAddr, addAddresses)
	_, err = ms.AddFeeWhitelistAddresses(ctx, addMsg)
	require.NoError(t, err)

	params, err := ms.GetParams(ctx)
	require.NoError(t, err)
	countAfterAdd := len(params.FeeWhitelist)

	removeAddresses := []string{testutil.AddrsStr[1], testutil.AddrsStr[3]}
	msg := types.NewMsgRemoveFeeWhitelistAddresses(moduleAddr, removeAddresses)
	_, err = ms.RemoveFeeWhitelistAddresses(ctx, msg)
	require.NoError(t, err)

	params, err = ms.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, countAfterAdd-2, len(params.FeeWhitelist))
	require.NotContains(t, params.FeeWhitelist, testutil.AddrsStr[1])
	require.Contains(t, params.FeeWhitelist, testutil.AddrsStr[2])
	require.NotContains(t, params.FeeWhitelist, testutil.AddrsStr[3])

	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress("gov"))
	require.NoError(t, err)

	invalidMsg := types.NewMsgRemoveFeeWhitelistAddresses(govAddr, removeAddresses)
	_, err = ms.RemoveFeeWhitelistAddresses(ctx, invalidMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid authority")
}

func Test_MsgServer_AddBridgeExecutor(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	params, err := ms.GetParams(ctx)
	require.NoError(t, err)
	initialExecutorCount := len(params.BridgeExecutors)

	moduleAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)

	// adding valid addresses
	newAddresses := []string{testutil.AddrsStr[1], testutil.AddrsStr[2]}
	msg := types.NewMsgAddBridgeExecutor(moduleAddr, newAddresses)
	_, err = ms.AddBridgeExecutor(ctx, msg)
	require.NoError(t, err)

	params, err = ms.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, initialExecutorCount+2, len(params.BridgeExecutors))
	require.Contains(t, params.BridgeExecutors, testutil.AddrsStr[1])
	require.Contains(t, params.BridgeExecutors, testutil.AddrsStr[2])

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
	ctx, input := testutil.CreateTestInput(t, false)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	moduleAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)

	addAddresses := []string{testutil.AddrsStr[1], testutil.AddrsStr[2], testutil.AddrsStr[3]}
	addMsg := types.NewMsgAddBridgeExecutor(moduleAddr, addAddresses)
	_, err = ms.AddBridgeExecutor(ctx, addMsg)
	require.NoError(t, err)

	params, err := ms.GetParams(ctx)
	require.NoError(t, err)
	countAfterAdd := len(params.BridgeExecutors)

	removeAddresses := []string{testutil.AddrsStr[1], testutil.AddrsStr[3]}
	msg := types.NewMsgRemoveBridgeExecutor(moduleAddr, removeAddresses)
	_, err = ms.RemoveBridgeExecutor(ctx, msg)
	require.NoError(t, err)

	params, err = ms.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, countAfterAdd-2, len(params.BridgeExecutors))
	require.NotContains(t, params.BridgeExecutors, testutil.AddrsStr[1])
	require.Contains(t, params.BridgeExecutors, testutil.AddrsStr[2])
	require.NotContains(t, params.BridgeExecutors, testutil.AddrsStr[3])

	// invalid authority
	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress("gov"))
	require.NoError(t, err)

	invalidMsg := types.NewMsgRemoveBridgeExecutor(govAddr, removeAddresses)
	_, err = ms.RemoveBridgeExecutor(ctx, invalidMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid authority")
}

func Test_MsgServer_UpdateMinGasPrices(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
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
	ctx, input := testutil.CreateTestInput(t, false)
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
	ctx, input := testutil.CreateTestInput(t, false)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	moduleAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)

	msg := types.NewMsgUpdateAdmin(moduleAddr, testutil.AddrsStr[1])
	_, err = ms.UpdateAdmin(ctx, msg)
	require.NoError(t, err)

	params, err := ms.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, testutil.AddrsStr[1], params.Admin)

	// invalid authority
	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress("gov"))
	require.NoError(t, err)

	invalidMsg := types.NewMsgUpdateAdmin(govAddr, testutil.AddrsStr[2])
	_, err = ms.UpdateAdmin(ctx, invalidMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid authority")
}

func Test_MsgServer_UpdateParams(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	params, err := ms.GetParams(ctx)
	require.NoError(t, err)
	params.MaxValidators = 1
	params.HistoricalEntries = 1
	params.BridgeExecutors = []string{testutil.Addrs[1].String(), testutil.Addrs[2].String()}

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
	ctx, input := testutil.CreateTestInput(t, false)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	// fund fee pool
	collectedFees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100)))
	input.Faucet.Fund(ctx, authtypes.NewModuleAddress(authtypes.FeeCollectorName), collectedFees...)

	beforeAmount := input.BankKeeper.GetBalance(ctx, testutil.Addrs[1], sdk.DefaultBondDenom).Amount

	msg := types.NewMsgSpendFeePool(
		authtypes.NewModuleAddress(types.ModuleName),
		testutil.Addrs[1],
		collectedFees,
	)
	_, err := ms.SpendFeePool(ctx, msg)
	require.NoError(t, err)

	afterAmount := input.BankKeeper.GetBalance(ctx, testutil.Addrs[1], sdk.DefaultBondDenom).Amount
	require.Equal(t, beforeAmount.Add(math.NewInt(100)), afterAmount)
}

/////////////////////////////////////////
// The messages for User

func Test_MsgServer_Withdraw(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	info := types.BridgeInfo{
		BridgeId:   1,
		BridgeAddr: testutil.AddrsStr[1],
		L1ChainId:  "test-chain-id",
		L1ClientId: "test-client-id",
		BridgeConfig: ophosttypes.BridgeConfig{
			Challenger: testutil.AddrsStr[2],
			Proposer:   testutil.AddrsStr[3],
			BatchInfo: ophosttypes.BatchInfo{
				Submitter: testutil.AddrsStr[4],
				ChainType: ophosttypes.BatchInfo_INITIA,
			},
			SubmissionInterval:    time.Minute,
			FinalizationPeriod:    time.Hour,
			SubmissionStartHeight: 1,
			Metadata:              []byte("metadata"),
		},
	}

	_, err := ms.SetBridgeInfo(ctx, types.NewMsgSetBridgeInfo(testutil.AddrsStr[0], info))
	require.NoError(t, err)

	baseDenom := "test_token"
	denom := ophosttypes.L2Denom(1, baseDenom)

	_, err = ms.FinalizeTokenDeposit(ctx, types.NewMsgFinalizeTokenDeposit(testutil.AddrsStr[0], "anyformataddr", testutil.AddrsStr[1], sdk.NewCoin(denom, math.NewInt(100)), 1, 1, "test/token", nil))
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
	msg = types.NewMsgInitiateTokenWithdrawal(accountAddr, testutil.AddrsStr[1], sdk.NewCoin(denom, math.NewInt(100)))
	_, err = ms.InitiateTokenWithdrawal(ctx, msg)
	require.NoError(t, err)
}

/////////////////////////////////////////
// The messages for Bridge Executor

func Test_MsgServer_SetBridgeInfo(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	info := types.BridgeInfo{
		BridgeId:   1,
		BridgeAddr: testutil.AddrsStr[1],
		L1ChainId:  "test-chain-id",
		L1ClientId: "test-client-id",
		BridgeConfig: ophosttypes.BridgeConfig{
			Challenger: testutil.AddrsStr[2],
			Proposer:   testutil.AddrsStr[3],
			BatchInfo: ophosttypes.BatchInfo{
				Submitter: testutil.AddrsStr[4],
				ChainType: ophosttypes.BatchInfo_INITIA,
			},
			SubmissionInterval:    time.Minute,
			FinalizationPeriod:    time.Hour,
			SubmissionStartHeight: 1,
			Metadata:              []byte("metadata"),
		},
	}

	_, err := ms.SetBridgeInfo(ctx, types.NewMsgSetBridgeInfo(testutil.AddrsStr[0], info))
	require.NoError(t, err)

	// reset possible
	_, err = ms.SetBridgeInfo(ctx, types.NewMsgSetBridgeInfo(testutil.AddrsStr[0], info))
	require.NoError(t, err)

	// cannot change chain id
	info.L1ChainId = "test-chain-id-2"
	_, err = ms.SetBridgeInfo(ctx, types.NewMsgSetBridgeInfo(testutil.AddrsStr[0], info))
	require.Error(t, err)

	// cannot change client id
	info.L1ChainId = "test-chain-id"
	info.L1ClientId = "test-client-id-2"
	_, err = ms.SetBridgeInfo(ctx, types.NewMsgSetBridgeInfo(testutil.AddrsStr[0], info))
	require.Error(t, err)

	info.L1ClientId = "test-client-id"

	// invalid bridge id
	info.BridgeId = 0

	_, err = ms.SetBridgeInfo(ctx, types.NewMsgSetBridgeInfo(testutil.AddrsStr[0], info))
	require.Error(t, err)

	// cannot change bridge id
	info.BridgeId = 2

	_, err = ms.SetBridgeInfo(ctx, types.NewMsgSetBridgeInfo(testutil.AddrsStr[0], info))
	require.ErrorContains(t, err, "expected bridge id")

	// cannot change bridge addr
	info.BridgeId = 1
	info.BridgeAddr = testutil.AddrsStr[0]

	_, err = ms.SetBridgeInfo(ctx, types.NewMsgSetBridgeInfo(testutil.AddrsStr[0], info))
	require.Error(t, err)
	require.ErrorContains(t, err, "expected bridge addr")
}

func Test_MsgServer_Deposit_ToModuleAccount(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	bz := sha3.Sum256([]byte("test_token"))
	denom := "l2/" + hex.EncodeToString(bz[:])

	opchildModuleAddress := authtypes.NewModuleAddress(types.ModuleName)

	beforeToBalance := input.BankKeeper.GetBalance(ctx, testutil.Addrs[1], denom)
	require.Equal(t, math.ZeroInt(), beforeToBalance.Amount)

	beforeModuleBalance := input.BankKeeper.GetBalance(ctx, opchildModuleAddress, denom)
	require.Equal(t, math.ZeroInt(), beforeModuleBalance.Amount)

	// valid deposit
	msg := types.NewMsgFinalizeTokenDeposit(testutil.AddrsStr[0], testutil.AddrsStr[1], opchildModuleAddress.String(), sdk.NewCoin(denom, math.NewInt(100)), 1, 1, "test_token", nil)
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

	afterToBalance := input.BankKeeper.GetBalance(ctx, testutil.Addrs[1], denom)
	require.Equal(t, math.ZeroInt(), afterToBalance.Amount)

	afterModuleBalance := input.BankKeeper.GetBalance(ctx, opchildModuleAddress, denom)
	require.True(t, afterModuleBalance.Amount.IsZero())

	// token withdrawal initiated
	events := sdk.UnwrapSDKContext(ctx).EventManager().Events()
	lastEvent := events[len(events)-1]
	require.Equal(t, sdk.NewEvent(
		types.EventTypeInitiateTokenWithdrawal,
		sdk.NewAttribute(types.AttributeKeyFrom, opchildModuleAddress.String()),
		sdk.NewAttribute(types.AttributeKeyTo, testutil.AddrsStr[1]),
		sdk.NewAttribute(types.AttributeKeyDenom, denom),
		sdk.NewAttribute(types.AttributeKeyBaseDenom, "test_token"),
		sdk.NewAttribute(types.AttributeKeyAmount, "100"),
		sdk.NewAttribute(types.AttributeKeyL2Sequence, "1"),
	), lastEvent)
}

func Test_MsgServer_Deposit_InvalidAddress(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	bz := sha3.Sum256([]byte("test_token"))
	denom := "l2/" + hex.EncodeToString(bz[:])

	beforeToBalance := input.BankKeeper.GetBalance(ctx, testutil.Addrs[1], denom)
	require.Equal(t, math.ZeroInt(), beforeToBalance.Amount)

	// valid deposit
	msg := types.NewMsgFinalizeTokenDeposit(testutil.AddrsStr[0], testutil.AddrsStr[1], "invalid_address", sdk.NewCoin(denom, math.NewInt(100)), 1, 1, "test_token", nil)
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

	afterToBalance := input.BankKeeper.GetBalance(ctx, testutil.Addrs[1], denom)
	require.Equal(t, math.ZeroInt(), afterToBalance.Amount)

	// token withdrawal initiated
	events := sdk.UnwrapSDKContext(ctx).EventManager().Events()
	lastEvent := events[len(events)-1]
	require.Equal(t, sdk.NewEvent(
		types.EventTypeInitiateTokenWithdrawal,
		sdk.NewAttribute(types.AttributeKeyFrom, "invalid_address"),
		sdk.NewAttribute(types.AttributeKeyTo, testutil.AddrsStr[1]),
		sdk.NewAttribute(types.AttributeKeyDenom, denom),
		sdk.NewAttribute(types.AttributeKeyBaseDenom, "test_token"),
		sdk.NewAttribute(types.AttributeKeyAmount, "100"),
		sdk.NewAttribute(types.AttributeKeyL2Sequence, "1"),
	), lastEvent)
}

func Test_MsgServer_Deposit_NoHook(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	bz := sha3.Sum256([]byte("test_token"))
	denom := "l2/" + hex.EncodeToString(bz[:])

	// unauthorized deposit
	msg := types.NewMsgFinalizeTokenDeposit(testutil.AddrsStr[1], testutil.AddrsStr[1], testutil.AddrsStr[1], sdk.NewCoin(denom, math.NewInt(100)), 1, 1, "test_token", nil)
	_, err := ms.FinalizeTokenDeposit(ctx, msg)
	require.Error(t, err)

	beforeBalance := input.BankKeeper.GetBalance(ctx, testutil.Addrs[1], denom)
	require.Equal(t, math.ZeroInt(), beforeBalance.Amount)

	// valid deposit
	msg = types.NewMsgFinalizeTokenDeposit(testutil.AddrsStr[0], testutil.AddrsStr[1], testutil.AddrsStr[1], sdk.NewCoin(denom, math.NewInt(100)), 1, 1, "test_token", nil)
	_, err = ms.FinalizeTokenDeposit(ctx, msg)
	require.NoError(t, err)
	require.True(t, input.TokenCreationFactory.Created[denom])

	afterBalance := input.BankKeeper.GetBalance(ctx, testutil.Addrs[1], denom)
	require.Equal(t, math.NewInt(100), afterBalance.Amount)
}

func Test_MsgServer_Deposit_HookSuccess(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	bz := sha3.Sum256([]byte("test_token"))
	denom := "l2/" + hex.EncodeToString(bz[:])

	require.Equal(t, math.ZeroInt(), input.BankKeeper.GetBalance(ctx, testutil.Addrs[1], denom).Amount)

	// empty deposit to create account
	priv, _, addr := testutil.KeyPubAddr()
	msg := types.NewMsgFinalizeTokenDeposit(testutil.AddrsStr[0], testutil.AddrsStr[1], addr.String(), sdk.NewCoin(denom, math.ZeroInt()), 1, 1, "test_token", nil)
	_, err := ms.FinalizeTokenDeposit(ctx, msg)
	require.NoError(t, err)

	// create hook data
	acc := input.AccountKeeper.GetAccount(ctx, addr)
	require.NotNil(t, acc)
	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv}, []uint64{acc.GetAccountNumber()}, []uint64{0}
	signedTxBz, err := input.EncodingConfig.TxConfig.TxEncoder()(testutil.GenerateTestTx(
		t, input,
		[]sdk.Msg{banktypes.NewMsgSend(addr, testutil.Addrs[2], sdk.NewCoins(sdk.NewCoin(denom, math.NewInt(50))))}, // send half tokens to addrs[2]
		privs, accNums, accSeqs, sdk.UnwrapSDKContext(ctx).ChainID(),
	))
	require.NoError(t, err)

	// valid deposit
	ctx = sdk.UnwrapSDKContext(ctx).WithEventManager(sdk.NewEventManager())
	msg = types.NewMsgFinalizeTokenDeposit(testutil.AddrsStr[0], testutil.AddrsStr[1], addr.String(), sdk.NewCoin(denom, math.NewInt(100)), 2, 1, "test_token", signedTxBz)
	_, err = ms.FinalizeTokenDeposit(ctx, msg)
	require.NoError(t, err)

	for _, event := range sdk.UnwrapSDKContext(ctx).EventManager().Events() {
		if event.Type == types.EventTypeFinalizeTokenDeposit {
			require.True(t, slices.Contains(event.Attributes, sdk.NewAttribute(types.AttributeKeySuccess, "true").ToKVPair()))
		}
	}

	// check addrs[2] balance
	afterBalance := input.BankKeeper.GetBalance(ctx, testutil.Addrs[2], denom)
	require.Equal(t, math.NewInt(50), afterBalance.Amount)
}

func Test_MsgServer_Deposit_HookFail(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	bz := sha3.Sum256([]byte("test_token"))
	denom := "l2/" + hex.EncodeToString(bz[:])

	require.Equal(t, math.ZeroInt(), input.BankKeeper.GetBalance(ctx, testutil.Addrs[1], denom).Amount)

	// empty deposit to create account
	priv, _, addr := testutil.KeyPubAddr()
	msg := types.NewMsgFinalizeTokenDeposit(testutil.AddrsStr[0], testutil.AddrsStr[1], addr.String(), sdk.NewCoin(denom, math.ZeroInt()), 1, 1, "test_token", nil)
	_, err := ms.FinalizeTokenDeposit(ctx, msg)
	require.NoError(t, err)

	// create hook data
	acc := input.AccountKeeper.GetAccount(ctx, addr)
	require.NotNil(t, acc)
	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv}, []uint64{acc.GetAccountNumber()}, []uint64{0}
	signedTxBz, err := input.EncodingConfig.TxConfig.TxEncoder()(testutil.GenerateTestTx(
		t, input,
		[]sdk.Msg{banktypes.NewMsgSend(addr, testutil.Addrs[2], sdk.NewCoins(sdk.NewCoin(denom, math.NewInt(101))))}, // send more than deposited tokens
		privs, accNums, accSeqs, sdk.UnwrapSDKContext(ctx).ChainID(),
	))
	require.NoError(t, err)

	// valid deposit
	ctx = sdk.UnwrapSDKContext(ctx).WithEventManager(sdk.NewEventManager())
	msg = types.NewMsgFinalizeTokenDeposit(testutil.AddrsStr[0], testutil.AddrsStr[1], addr.String(), sdk.NewCoin(denom, math.NewInt(100)), 2, 1, "test_token", signedTxBz)
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
	afterBalance := input.BankKeeper.GetBalance(ctx, testutil.Addrs[2], denom)
	require.Equal(t, math.NewInt(0), afterBalance.Amount)

	// check receiver has no balance
	afterBalance = input.BankKeeper.GetBalance(ctx, addr, denom)
	require.Equal(t, math.NewInt(100), afterBalance.Amount)
}

func Test_MsgServer_Deposit_HookFail_WithOutOfGas(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	bz := sha3.Sum256([]byte("test_token"))
	denom := "l2/" + hex.EncodeToString(bz[:])

	require.Equal(t, math.ZeroInt(), input.BankKeeper.GetBalance(ctx, testutil.Addrs[1], denom).Amount)

	// empty deposit to create account
	priv, _, addr := testutil.KeyPubAddr()
	msg := types.NewMsgFinalizeTokenDeposit(testutil.AddrsStr[0], testutil.AddrsStr[1], addr.String(), sdk.NewCoin(denom, math.ZeroInt()), 1, 1, "test_token", nil)
	_, err := ms.FinalizeTokenDeposit(ctx, msg)
	require.NoError(t, err)

	// create hook data
	acc := input.AccountKeeper.GetAccount(ctx, addr)
	require.NotNil(t, acc)
	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv}, []uint64{acc.GetAccountNumber()}, []uint64{0}
	signedTxBz, err := input.EncodingConfig.TxConfig.TxEncoder()(testutil.GenerateTestTx(
		t, input,
		[]sdk.Msg{banktypes.NewMsgSend(addr, testutil.Addrs[2], sdk.NewCoins(sdk.NewCoin(denom, math.NewInt(100))))},
		privs, accNums, accSeqs, sdk.UnwrapSDKContext(ctx).ChainID(),
	))
	require.NoError(t, err)

	params, err := input.OPChildKeeper.GetParams(ctx)
	require.NoError(t, err)

	params.HookMaxGas = 1
	input.OPChildKeeper.SetParams(ctx, params)

	// valid deposit
	ctx = sdk.UnwrapSDKContext(ctx).WithEventManager(sdk.NewEventManager()).WithGasMeter(storetypes.NewGasMeter(2_000_000))
	msg = types.NewMsgFinalizeTokenDeposit(testutil.AddrsStr[0], testutil.AddrsStr[1], addr.String(), sdk.NewCoin(denom, math.NewInt(100)), 2, 1, "test_token", signedTxBz)
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
	afterBalance := input.BankKeeper.GetBalance(ctx, testutil.Addrs[2], denom)
	require.Equal(t, math.NewInt(0), afterBalance.Amount)

	// check receiver has no balance
	afterBalance = input.BankKeeper.GetBalance(ctx, addr, denom)
	require.Equal(t, math.NewInt(100), afterBalance.Amount)
}

func Test_MsgServer_UpdateOracle(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
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
	_, err = ms.UpdateOracle(ctx, types.NewMsgUpdateOracle(testutil.AddrsStr[0], 11, extCommitBz))
	require.NoError(t, err)

	_, err = ms.UpdateOracle(ctx, types.NewMsgUpdateOracle(testutil.AddrsStr[1], 11, extCommitBz))
	require.Error(t, err)
}

func Test_MsgServer_UpdateOracleFail(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
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
	_, err = ms.UpdateOracle(ctx, types.NewMsgUpdateOracle(testutil.AddrsStr[0], 11, []byte{}))
	require.EqualError(t, err, types.ErrOracleDisabled.Error())
}

func Test_MsgServer_RelayOracleData(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
	opchildKeeper := input.OPChildKeeper

	bridgeInfo := types.BridgeInfo{
		BridgeId:   1,
		BridgeAddr: testutil.AddrsStr[0],
		L1ChainId:  "test-host-1",
		L1ClientId: "test-client-id",
		BridgeConfig: ophosttypes.BridgeConfig{
			Challenger:            testutil.AddrsStr[1],
			Proposer:              testutil.AddrsStr[2],
			SubmissionInterval:    100,
			FinalizationPeriod:    100,
			SubmissionStartHeight: 1,
			OracleEnabled:         true,
			Metadata:              []byte{},
			BatchInfo:             ophosttypes.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: ophosttypes.BatchInfo_INITIA},
		},
	}
	err := opchildKeeper.BridgeInfo.Set(ctx, bridgeInfo)
	require.NoError(t, err)

	oracleHash, _ := hex.DecodeString("d3f45ee700ee1a37b1b114d7e12f44d2514cafaa2308fa0446c324f3491453f7")
	oracleData := types.OracleData{
		BridgeId:        1,
		OraclePriceHash: oracleHash,
		Prices: []types.OraclePriceData{
			{
				CurrencyPair:   "BTC/USD",
				Price:          "9222376713",
				Decimals:       5,
				CurrencyPairId: 4,
				Nonce:          567,
			},
			{
				CurrencyPair:   "ETH/USD",
				Price:          "3243513053",
				Decimals:       6,
				CurrencyPairId: 15,
				Nonce:          567,
			},
			{
				CurrencyPair:   "ATOM/USD",
				Price:          "2180500000",
				Decimals:       9,
				CurrencyPairId: 1,
				Nonce:          567,
			},
		},
		L1BlockHeight: 663,
		L1BlockTime:   1765532733321115000,
		Proof:         mustDecodeHex("0aeb010a0a69637332333a6961766c1209a100000000000000011ad1010ace010a09a10000000000000001122f0a20d3f45ee700ee1a37b1b114d7e12f44d2514cafaa2308fa0446c324f3491453f710970518f8eab787abd59bc0181a0c0801180120012a040002ae0a222a080112260204ae0a201fdb1e2e60ff88045cd3f68bd0da4593b20d710c13725f98e239dd014fbac36020222a080112260406ae0a20d0a1e370a3c38d3c05c109c408335caa79e61e6005cfba0fb1da3fc640c55cd820222a08011226060aae0a20bbe83488fa11800d0da08a55ae8d9201868378c190c793f5fcc294f7fb61faae200a96020a0c69637332333a73696d706c6512066f70686f73741afd010afa010a066f70686f737412200cecd8af22290cfdd98431e98cccb6418b7b27a3a9914616b0237d3d2f027ccb1a090801180120012a01002225080112210141c04a42e47994d4dbcb3060eedd792d7b623f1bf61f4f7498bd5e3e1119666a22250801122101a9a139f033e62375c436e237e8fb7deddb3a5ade8a485d45e0229d8b8ea9d937222508011221010f31c2c269df8cc8cccdc393e8c168cd397589427bf4fc70646b263f9389529a222708011201011a20cfff89b3f02d88273b060b24913e22bfe25d236a2cd25c21d142c6d93a1eb92e22250801122101bada0d57b7de3594067a4227846a20361bb1429f632d0cd0cde8f31395b96772"),
		ProofHeight: clienttypes.Height{
			RevisionNumber: 1,
			RevisionHeight: 664,
		},
	}

	ms := keeper.NewMsgServerImpl(&opchildKeeper)

	msg := types.NewMsgRelayOracleData(testutil.AddrsStr[0], oracleData)
	_, err = ms.RelayOracleData(ctx, msg)

	// expected to fail at proof verification without IBC client state setup, successful test is in oracle_test.go
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to handle oracle data")
}

func Test_MsgServer_RelayOracleData_InvalidMessage(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
	opchildKeeper := input.OPChildKeeper

	ms := keeper.NewMsgServerImpl(&opchildKeeper)

	testCases := []struct {
		name        string
		oracleData  types.OracleData
		expectedErr string
	}{
		{
			name: "zero bridge id",
			oracleData: types.OracleData{
				BridgeId:        0, // invalid - zero bridge ID
				OraclePriceHash: make([]byte, 32),
				Prices: []types.OraclePriceData{
					{
						CurrencyPair:   "BTC/USD",
						Price:          "50000",
						Decimals:       5,
						CurrencyPairId: 1,
						Nonce:          1,
					},
				},
				L1BlockHeight: 100,
				L1BlockTime:   1000,
				Proof:         []byte("proof"),
				ProofHeight: clienttypes.Height{
					RevisionHeight: 100,
				},
			},
			expectedErr: "bridge id cannot be zero",
		},
		{
			name: "empty oracle price hash",
			oracleData: types.OracleData{
				BridgeId:        1,
				OraclePriceHash: []byte{}, // invalid - empty hash
				Prices: []types.OraclePriceData{
					{
						CurrencyPair:   "BTC/USD",
						Price:          "50000",
						Decimals:       5,
						CurrencyPairId: 1,
						Nonce:          1,
					},
				},
				L1BlockHeight: 100,
				L1BlockTime:   1000,
				Proof:         []byte("proof"),
				ProofHeight: clienttypes.Height{
					RevisionHeight: 100,
				},
			},
			expectedErr: "oracle price hash cannot be empty",
		},
		{
			name: "wrong oracle price hash length",
			oracleData: types.OracleData{
				BridgeId:        1,
				OraclePriceHash: make([]byte, 16), // invalid - should be 32 bytes
				Prices: []types.OraclePriceData{
					{
						CurrencyPair:   "BTC/USD",
						Price:          "50000",
						Decimals:       5,
						CurrencyPairId: 1,
						Nonce:          1,
					},
				},
				L1BlockHeight: 100,
				L1BlockTime:   1000,
				Proof:         []byte("proof"),
				ProofHeight: clienttypes.Height{
					RevisionHeight: 100,
				},
			},
			expectedErr: "oracle price hash must be 32 bytes",
		},
		{
			name: "empty prices array",
			oracleData: types.OracleData{
				BridgeId:        1,
				OraclePriceHash: make([]byte, 32),
				Prices:          []types.OraclePriceData{}, // invalid - empty array
				L1BlockHeight:   100,
				L1BlockTime:     1000,
				Proof:           []byte("proof"),
				ProofHeight: clienttypes.Height{
					RevisionHeight: 100,
				},
			},
			expectedErr: "prices cannot be empty",
		},
		{
			name: "invalid price data - empty currency pair",
			oracleData: types.OracleData{
				BridgeId:        1,
				OraclePriceHash: make([]byte, 32),
				Prices: []types.OraclePriceData{
					{
						CurrencyPair:   "", // invalid - empty
						Price:          "50000",
						Decimals:       5,
						CurrencyPairId: 1,
						Nonce:          1,
					},
				},
				L1BlockHeight: 100,
				L1BlockTime:   1000,
				Proof:         []byte("proof"),
				ProofHeight: clienttypes.Height{
					RevisionHeight: 100,
				},
			},
			expectedErr: "currency pair cannot be empty",
		},
		{
			name: "invalid price data - empty price",
			oracleData: types.OracleData{
				BridgeId:        1,
				OraclePriceHash: make([]byte, 32),
				Prices: []types.OraclePriceData{
					{
						CurrencyPair:   "BTC/USD",
						Price:          "", // invalid - empty
						Decimals:       5,
						CurrencyPairId: 1,
						Nonce:          1,
					},
				},
				L1BlockHeight: 100,
				L1BlockTime:   1000,
				Proof:         []byte("proof"),
				ProofHeight: clienttypes.Height{
					RevisionHeight: 100,
				},
			},
			expectedErr: "price cannot be empty",
		},
		{
			name: "invalid price data - invalid price format",
			oracleData: types.OracleData{
				BridgeId:        1,
				OraclePriceHash: make([]byte, 32),
				Prices: []types.OraclePriceData{
					{
						CurrencyPair:   "BTC/USD",
						Price:          "not-a-number", // invalid format
						Decimals:       5,
						CurrencyPairId: 1,
						Nonce:          1,
					},
				},
				L1BlockHeight: 100,
				L1BlockTime:   1000,
				Proof:         []byte("proof"),
				ProofHeight: clienttypes.Height{
					RevisionHeight: 100,
				},
			},
			expectedErr: "invalid price format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgRelayOracleData(testutil.AddrsStr[0], tc.oracleData)
			_, err := ms.RelayOracleData(ctx, msg)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

func Test_MsgServer_RelayOracleData_OracleDisabled(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
	opchildKeeper := input.OPChildKeeper

	bridgeInfo := types.BridgeInfo{
		BridgeId:   1,
		BridgeAddr: testutil.AddrsStr[0],
		L1ChainId:  "test-host-1",
		L1ClientId: "test-client-id",
		BridgeConfig: ophosttypes.BridgeConfig{
			Challenger:            testutil.AddrsStr[1],
			Proposer:              testutil.AddrsStr[2],
			SubmissionInterval:    100,
			FinalizationPeriod:    100,
			SubmissionStartHeight: 1,
			OracleEnabled:         false, // disabled
			Metadata:              []byte{},
			BatchInfo:             ophosttypes.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: ophosttypes.BatchInfo_INITIA},
		},
	}
	err := opchildKeeper.BridgeInfo.Set(ctx, bridgeInfo)
	require.NoError(t, err)

	oracleData := types.OracleData{
		BridgeId:        1,
		OraclePriceHash: make([]byte, 32),
		Prices: []types.OraclePriceData{
			{
				CurrencyPair:   "BTC/USD",
				Price:          "50000000000",
				Decimals:       8,
				CurrencyPairId: 1,
				Nonce:          1,
			},
		},
		L1BlockHeight: 100,
		L1BlockTime:   1000000000,
		Proof:         []byte("test-proof"),
		ProofHeight: clienttypes.Height{
			RevisionNumber: 0,
			RevisionHeight: 100,
		},
	}

	ms := keeper.NewMsgServerImpl(&opchildKeeper)
	msg := types.NewMsgRelayOracleData(testutil.AddrsStr[0], oracleData)
	_, err = ms.RelayOracleData(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "oracle is disabled")
}
