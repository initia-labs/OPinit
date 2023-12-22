package keeper_test

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"

	testutilsims "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/initia-labs/OPinit/x/opchild/keeper"
	"github.com/initia-labs/OPinit/x/opchild/types"
)

/////////////////////////////////////////
// The messages for Validators

func Test_MsgServer_ExecuteMessages(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	ms := keeper.NewMsgServerImpl(input.OPChildKeeper)
	valPubKeys := testutilsims.CreateTestPubKeys(2)

	// register validator
	val, err := types.NewValidator(valAddrs[0], valPubKeys[0], "val1")
	require.NoError(t, err)

	input.OPChildKeeper.SetValidator(ctx, val)

	// apply validator updates
	input.OPChildKeeper.BlockValidatorUpdates(ctx)

	addMsg, err := types.NewMsgAddValidator("val2", authtypes.NewModuleAddress(types.ModuleName), valAddrs[1], valPubKeys[1])
	require.NoError(t, err)

	removeMsg, err := types.NewMsgRemoveValidator(authtypes.NewModuleAddress(types.ModuleName), valAddrs[0])
	require.NoError(t, err)

	msg, err := types.NewMsgExecuteMessages(addrs[0], []sdk.Msg{addMsg, removeMsg})
	require.NoError(t, err)

	_, err = ms.ExecuteMessages(ctx, msg)
	require.NoError(t, err)

	// apply validator updates
	input.OPChildKeeper.BlockValidatorUpdates(ctx)

	vals, err := input.OPChildKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	require.Len(t, vals, 1)
	require.Equal(t, vals[0].Moniker, "val2")
}

func Test_MsgServer_ExecuteLegacyContents(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	ms := keeper.NewMsgServerImpl(input.OPChildKeeper)
	valPubKeys := testutilsims.CreateTestPubKeys(2)

	// register validator
	val, err := types.NewValidator(valAddrs[0], valPubKeys[0], "val1")
	require.NoError(t, err)

	input.OPChildKeeper.SetValidator(ctx, val)

	// apply validator updates
	input.OPChildKeeper.BlockValidatorUpdates(ctx)

	// valid legacy content
	legacyContent := testLegacyContent{
		Title:       "title",
		Description: "description",
		Message:     "test",
	}
	msg, err := types.NewMsgExecuteLegacyContents(addrs[0], []govv1beta1.Content{&legacyContent})
	require.NoError(t, err)

	_, err = ms.ExecuteLegacyContents(ctx, msg)
	require.NoError(t, err)

	// unauthorized executor
	msg, err = types.NewMsgExecuteLegacyContents(addrs[1], []govv1beta1.Content{&legacyContent})
	require.NoError(t, err)

	_, err = ms.ExecuteLegacyContents(ctx, msg)
	require.Error(t, err)

	// legacy content with wrong message
	legacyContent.Message = "wrong message"
	msg, err = types.NewMsgExecuteLegacyContents(addrs[0], []govv1beta1.Content{&legacyContent})
	require.NoError(t, err)

	_, err = ms.ExecuteLegacyContents(ctx, msg)
	require.Error(t, err)
}

/////////////////////////////////////////
// The messages for Authority

func Test_MsgServer_AddValidator(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	ms := keeper.NewMsgServerImpl(input.OPChildKeeper)
	valPubKeys := testutilsims.CreateTestPubKeys(2)

	msg, err := types.NewMsgAddValidator("val1", authtypes.NewModuleAddress(types.ModuleName), valAddrs[0], valPubKeys[0])
	require.NoError(t, err)

	_, err = ms.AddValidator(ctx, msg)
	require.NoError(t, err)

	// invalid signer
	msg, err = types.NewMsgAddValidator("val1", addrs[0], valAddrs[0], valPubKeys[0])
	require.NoError(t, err)

	_, err = ms.AddValidator(ctx, msg)
	require.Error(t, err)

	// duplicate add validator
	msg, err = types.NewMsgAddValidator("val1", authtypes.NewModuleAddress(types.ModuleName), valAddrs[0], valPubKeys[1])
	require.NoError(t, err)

	_, err = ms.AddValidator(ctx, msg)
	require.Error(t, err)

	// duplicate cons pubkey
	msg, err = types.NewMsgAddValidator("val1", authtypes.NewModuleAddress(types.ModuleName), valAddrs[1], valPubKeys[0])
	require.NoError(t, err)

	_, err = ms.AddValidator(ctx, msg)
	require.Error(t, err)
}

func Test_MsgServer_RemoveValidator(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(input.OPChildKeeper)

	valPubKeys := testutilsims.CreateTestPubKeys(2)

	// register validator
	val, err := types.NewValidator(valAddrs[0], valPubKeys[0], "val1")
	require.NoError(t, err)

	input.OPChildKeeper.SetValidator(ctx, val)

	// invalid signer
	msg, err := types.NewMsgRemoveValidator(addrs[0], valAddrs[0])
	require.NoError(t, err)

	_, err = ms.RemoveValidator(
		ctx,
		msg,
	)
	require.Error(t, err)

	// remove not existing validator
	msg, err = types.NewMsgRemoveValidator(authtypes.NewModuleAddress(types.ModuleName), valAddrs[1])
	require.NoError(t, err)

	_, err = ms.RemoveValidator(
		ctx,
		msg,
	)
	require.Error(t, err)

	// valid remove validator
	msg, err = types.NewMsgRemoveValidator(authtypes.NewModuleAddress(types.ModuleName), valAddrs[0])
	require.NoError(t, err)

	_, err = ms.RemoveValidator(
		ctx,
		msg,
	)
	require.NoError(t, err)
}

func Test_MsgServer_UpdateParams(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(input.OPChildKeeper)

	params, err := ms.GetParams(ctx)
	require.NoError(t, err)
	params.MaxValidators = 1
	params.HistoricalEntries = 1
	params.BridgeExecutor = addrs[1].String()

	msg := types.NewMsgUpdateParams(authtypes.NewModuleAddress(types.ModuleName), &params)
	_, err = ms.UpdateParams(ctx, msg)
	require.NoError(t, err)
	_params, err := ms.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, params, _params)

	// invalid signer
	msg = types.NewMsgUpdateParams(authtypes.NewModuleAddress("gov"), &params)
	require.NoError(t, err)

	_, err = ms.UpdateParams(
		ctx,
		msg,
	)
	require.Error(t, err)
}

func Test_MsgServer_SpendFeePool(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(input.OPChildKeeper)

	// fund fee pool
	collectedFees := sdk.NewCoins(sdk.NewCoin(baseDenom, math.NewInt(100)))
	input.Faucet.Fund(ctx, authtypes.NewModuleAddress(authtypes.FeeCollectorName), collectedFees...)

	beforeAmount := input.BankKeeper.GetBalance(ctx, addrs[1], baseDenom).Amount

	msg := types.NewMsgSpendFeePool(
		authtypes.NewModuleAddress(types.ModuleName),
		addrs[1],
		collectedFees,
	)
	_, err := ms.SpendFeePool(ctx, msg)
	require.NoError(t, err)

	afterAmount := input.BankKeeper.GetBalance(ctx, addrs[1], baseDenom).Amount
	require.Equal(t, beforeAmount.Add(math.NewInt(100)), afterAmount)
}

/////////////////////////////////////////
// The messages for User

func Test_MsgServer_Withdraw(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(input.OPChildKeeper)

	bz := sha3.Sum256([]byte("test_token"))
	denom := "l2/" + hex.EncodeToString(bz[:])

	// fund asset
	account := input.Faucet.NewFundedAccount(ctx, sdk.NewCoin(denom, math.NewInt(1_000_000_000)))

	// valid
	msg := types.NewMsgInitiateTokenWithdrawal(account, addrs[1], sdk.NewCoin(denom, math.NewInt(100)))
	_, err := ms.InitiateTokenWithdrawal(ctx, msg)
	require.NoError(t, err)

}

/////////////////////////////////////////
// The messages for Bridge Executor

func Test_MsgServer_Deposit_NoHook(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(input.OPChildKeeper)

	bz := sha3.Sum256([]byte("test_token"))
	denom := "l2/" + hex.EncodeToString(bz[:])

	// unauthorized deposit
	msg := types.NewMsgFinalizeTokenDeposit(addrs[1], addrs[1], addrs[1], sdk.NewCoin(denom, math.NewInt(100)), 1, nil)
	_, err := ms.FinalizeTokenDeposit(ctx, msg)
	require.Error(t, err)

	beforeBalance := input.BankKeeper.GetBalance(ctx, addrs[1], denom)
	require.Equal(t, math.ZeroInt(), beforeBalance.Amount)

	// valid deposit
	msg = types.NewMsgFinalizeTokenDeposit(addrs[0], addrs[1], addrs[1], sdk.NewCoin(denom, math.NewInt(100)), 1, nil)
	_, err = ms.FinalizeTokenDeposit(ctx, msg)
	require.NoError(t, err)

	afterBalance := input.BankKeeper.GetBalance(ctx, addrs[1], denom)
	require.Equal(t, math.NewInt(100), afterBalance.Amount)
}

func Test_MsgServer_Deposit_HookSuccess(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(input.OPChildKeeper)

	bz := sha3.Sum256([]byte("test_token"))
	denom := "l2/" + hex.EncodeToString(bz[:])

	require.Equal(t, math.ZeroInt(), input.BankKeeper.GetBalance(ctx, addrs[1], denom).Amount)

	hookMsgBytes, err := json.Marshal("message bytes")
	require.NoError(t, err)

	input.BridgeHook.msgBytes = hookMsgBytes
	input.BridgeHook.err = nil

	// valid deposit
	msg := types.NewMsgFinalizeTokenDeposit(addrs[0], addrs[1], addrs[1], sdk.NewCoin(denom, math.NewInt(100)), 1, hookMsgBytes)
	_, err = ms.FinalizeTokenDeposit(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, hookMsgBytes, input.BridgeHook.msgBytes)

	for _, event := range sdk.UnwrapSDKContext(ctx).EventManager().Events() {
		if event.Type == types.EventTypeFinalizeTokenDeposit {
			for _, attr := range event.Attributes {
				if attr.Key == types.AttributeKeyHookSuccess {
					require.Equal(t, "true", attr.Value)
				}
			}
		}
	}
}

func Test_MsgServer_Deposit_HookFail(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(input.OPChildKeeper)

	bz := sha3.Sum256([]byte("test_token"))
	denom := "l2/" + hex.EncodeToString(bz[:])

	input.BridgeHook.err = errors.New("should be failed")

	// valid deposit
	msg := types.NewMsgFinalizeTokenDeposit(addrs[0], addrs[1], addrs[1], sdk.NewCoin(denom, math.NewInt(100)), 1, []byte("invalid_message"))
	_, err := ms.FinalizeTokenDeposit(ctx, msg)
	require.NoError(t, err)
	require.Empty(t, input.BridgeHook.msgBytes)

	for _, event := range sdk.UnwrapSDKContext(ctx).EventManager().Events() {
		if event.Type == types.EventTypeFinalizeTokenDeposit {
			for _, attr := range event.Attributes {
				if attr.Key == types.AttributeKeyHookSuccess {
					require.Equal(t, "false", attr.Value)
				}
			}
		}
	}

}
