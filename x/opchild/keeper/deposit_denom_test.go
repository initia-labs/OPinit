package keeper_test

import (
	"context"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	"github.com/initia-labs/OPinit/x/opchild/keeper"
	"github.com/initia-labs/OPinit/x/opchild/types"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
)

// requireNoDepositEffect asserts that a refused deposit left nothing behind
func requireNoDepositEffect(t *testing.T, ctx context.Context, input TestKeepers, denom string, l1Sequence, l2Sequence uint64, supply math.Int, pairExists bool) {
	nextL1Sequence, err := input.OPChildKeeper.GetNextL1Sequence(ctx)
	require.NoError(t, err)
	require.Equal(t, l1Sequence, nextL1Sequence)

	nextL2Sequence, err := input.OPChildKeeper.GetNextL2Sequence(ctx)
	require.NoError(t, err)
	require.Equal(t, l2Sequence, nextL2Sequence)

	require.Equal(t, supply, input.BankKeeper.GetSupply(ctx, denom).Amount)

	hasPair, err := input.OPChildKeeper.DenomPairs.Has(ctx, denom)
	require.NoError(t, err)
	require.Equal(t, pairExists, hasPair)

	for _, event := range sdk.UnwrapSDKContext(ctx).EventManager().Events() {
		require.NotEqual(t, types.EventTypeFinalizeTokenDeposit, event.Type)
		require.NotEqual(t, types.EventTypeInitiateTokenWithdrawal, event.Type)
	}
}

func Test_MsgServer_Deposit_NonCanonicalDenom(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)
	setTestBridgeInfo(t, ctx, input)

	l2Sequence, err := input.OPChildKeeper.GetNextL2Sequence(ctx)
	require.NoError(t, err)

	testCases := []struct {
		name      string
		denom     string
		baseDenom string
	}{
		{"native denom", sdk.DefaultBondDenom, "test_token"},
		{"ibc voucher denom", transfertypes.ParseDenomTrace("transfer/channel-0/test_token").IBCDenom(), "test_token"},
		{"l2 denom of another base denom", ophosttypes.L2Denom(testBridgeId, "other_token"), "test_token"},
		{"l2 denom of another bridge", ophosttypes.L2Denom(testBridgeId+1, "test_token"), "test_token"},
		{"l2 denom of the hashed base denom", ophosttypes.L2Denom(testBridgeId, ophosttypes.L2Denom(testBridgeId, "test_token")), "test_token"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			supply := input.BankKeeper.GetSupply(ctx, tc.denom).Amount
			hasPair, err := input.OPChildKeeper.DenomPairs.Has(ctx, tc.denom)
			require.NoError(t, err)

			// a valid recipient, an invalid recipient and a zero amount are all refused the same way
			for _, msg := range []*types.MsgFinalizeTokenDeposit{
				types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], addrsStr[1], sdk.NewCoin(tc.denom, math.NewInt(100)), 1, 1, tc.baseDenom, nil),
				types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], "invalid_address", sdk.NewCoin(tc.denom, math.NewInt(100)), 1, 1, tc.baseDenom, nil),
				types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], addrsStr[1], sdk.NewCoin(tc.denom, math.ZeroInt()), 1, 1, tc.baseDenom, nil),
			} {
				eventCtx := sdk.UnwrapSDKContext(ctx).WithEventManager(sdk.NewEventManager())
				_, err := ms.FinalizeTokenDeposit(eventCtx, msg)
				require.ErrorIs(t, err, types.ErrNonCanonicalDepositDenom)
				requireNoDepositEffect(t, eventCtx, input, tc.denom, 1, l2Sequence, supply, hasPair)
			}
		})
	}

	// the canonical denom of the same base denom is accepted at the same sequence
	denom := ophosttypes.L2Denom(testBridgeId, "test_token")
	res, err := ms.FinalizeTokenDeposit(ctx, types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], addrsStr[1], sdk.NewCoin(denom, math.NewInt(100)), 1, 1, "test_token", nil))
	require.NoError(t, err)
	require.Equal(t, types.SUCCESS, res.Result)
	require.Equal(t, math.NewInt(100), input.BankKeeper.GetBalance(ctx, addrs[1], denom).Amount)
}

func Test_MsgServer_Deposit_MigratedDenom(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	denom := testL2Denom(t, ctx, input, "test_token")
	otherDenom := testL2Denom(t, ctx, input, "other_token")

	// a deposit before the migration is registered
	_, err := ms.FinalizeTokenDeposit(ctx, types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], addrsStr[1], sdk.NewCoin(denom, math.NewInt(100)), 1, 1, "test_token", nil))
	require.NoError(t, err)

	_, err = ms.RegisterMigrationInfo(ctx, types.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(types.ModuleName).String(),
		types.MigrationInfo{Denom: denom, IbcChannelId: "channel-0", IbcPortId: "transfer"},
	))
	require.NoError(t, err)

	l2Sequence, err := input.OPChildKeeper.GetNextL2Sequence(ctx)
	require.NoError(t, err)
	supply := input.BankKeeper.GetSupply(ctx, denom).Amount
	require.Equal(t, math.NewInt(100), supply)

	// a valid recipient, an invalid recipient and a zero amount are refused
	for _, msg := range []*types.MsgFinalizeTokenDeposit{
		types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], addrsStr[1], sdk.NewCoin(denom, math.NewInt(100)), 2, 1, "test_token", nil),
		types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], "invalid_address", sdk.NewCoin(denom, math.NewInt(100)), 2, 1, "test_token", nil),
		types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], authtypes.NewModuleAddress(types.ModuleName).String(), sdk.NewCoin(denom, math.NewInt(100)), 2, 1, "test_token", nil),
		types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], addrsStr[1], sdk.NewCoin(denom, math.ZeroInt()), 2, 1, "test_token", nil),
	} {
		eventCtx := sdk.UnwrapSDKContext(ctx).WithEventManager(sdk.NewEventManager())
		_, err := ms.FinalizeTokenDeposit(eventCtx, msg)
		require.ErrorIs(t, err, types.ErrMigratedDenomDeposit)
		requireNoDepositEffect(t, eventCtx, input, denom, 2, l2Sequence, supply, true)
	}
	require.Equal(t, math.NewInt(100), input.BankKeeper.GetBalance(ctx, addrs[1], denom).Amount)

	// a denom without migration info is still deposited at that sequence
	res, err := ms.FinalizeTokenDeposit(ctx, types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], addrsStr[1], sdk.NewCoin(otherDenom, math.NewInt(70)), 2, 1, "other_token", nil))
	require.NoError(t, err)
	require.Equal(t, types.SUCCESS, res.Result)
	require.Equal(t, math.NewInt(70), input.BankKeeper.GetBalance(ctx, addrs[1], otherDenom).Amount)

	// the migrated denom still arrives over IBC: the voucher is converted to the OP token
	ibcCoin := sdk.NewCoin(transfertypes.ParseDenomTrace("transfer/channel-0/test_token").IBCDenom(), math.NewInt(30))
	require.NoError(t, input.BankKeeper.MintCoins(ctx, transfertypes.ModuleName, sdk.NewCoins(ibcCoin)))
	require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(ctx, transfertypes.ModuleName, addrs[1], sdk.NewCoins(ibcCoin)))
	l2Coin, err := input.OPChildKeeper.HandleMigratedTokenDeposit(ctx, addrs[1], ibcCoin, "")
	require.NoError(t, err)
	require.Equal(t, sdk.NewCoin(denom, math.NewInt(30)), l2Coin)
	require.Equal(t, math.NewInt(130), input.BankKeeper.GetBalance(ctx, addrs[1], denom).Amount)
}

func Test_MsgServer_Deposit_CanonicalDenom(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	denom := testL2Denom(t, ctx, input, "test_token")

	// zero amount deposit to create an account
	priv, _, addr := keyPubAddr()
	res, err := ms.FinalizeTokenDeposit(ctx, types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], addr.String(), sdk.NewCoin(denom, math.ZeroInt()), 1, 1, "test_token", nil))
	require.NoError(t, err)
	require.Equal(t, types.SUCCESS, res.Result)
	acc := input.AccountKeeper.GetAccount(ctx, addr)
	require.NotNil(t, acc)
	require.True(t, input.BankKeeper.GetSupply(ctx, denom).IsZero())

	// normal deposit
	res, err = ms.FinalizeTokenDeposit(ctx, types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], addr.String(), sdk.NewCoin(denom, math.NewInt(100)), 2, 1, "test_token", nil))
	require.NoError(t, err)
	require.Equal(t, types.SUCCESS, res.Result)
	require.Equal(t, math.NewInt(100), input.BankKeeper.GetBalance(ctx, addr, denom).Amount)

	baseDenom, err := input.OPChildKeeper.GetBaseDenom(ctx, denom)
	require.NoError(t, err)
	require.Equal(t, "test_token", baseDenom)

	// deposit with hook
	signedTxBz, err := input.EncodingConfig.TxConfig.TxEncoder()(generateTestTx(
		t, input,
		[]sdk.Msg{banktypes.NewMsgSend(addr, addrs[2], sdk.NewCoins(sdk.NewCoin(denom, math.NewInt(50))))},
		[]cryptotypes.PrivKey{priv}, []uint64{acc.GetAccountNumber()}, []uint64{0}, sdk.UnwrapSDKContext(ctx).ChainID(),
	))
	require.NoError(t, err)

	eventCtx := sdk.UnwrapSDKContext(ctx).WithEventManager(sdk.NewEventManager())
	res, err = ms.FinalizeTokenDeposit(eventCtx, types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], addr.String(), sdk.NewCoin(denom, math.NewInt(100)), 3, 1, "test_token", signedTxBz))
	require.NoError(t, err)
	require.Equal(t, types.SUCCESS, res.Result)
	for _, event := range eventCtx.EventManager().Events() {
		if event.Type == types.EventTypeFinalizeTokenDeposit {
			require.True(t, slices.Contains(event.Attributes, sdk.NewAttribute(types.AttributeKeySuccess, "true").ToKVPair()))
		}
	}
	require.Equal(t, math.NewInt(150), input.BankKeeper.GetBalance(ctx, addr, denom).Amount)
	require.Equal(t, math.NewInt(50), input.BankKeeper.GetBalance(ctx, addrs[2], denom).Amount)

	// zero amount deposit with hook
	signedTxBz, err = input.EncodingConfig.TxConfig.TxEncoder()(generateTestTx(
		t, input,
		[]sdk.Msg{banktypes.NewMsgSend(addr, addrs[2], sdk.NewCoins(sdk.NewCoin(denom, math.NewInt(10))))},
		[]cryptotypes.PrivKey{priv}, []uint64{acc.GetAccountNumber()}, []uint64{1}, sdk.UnwrapSDKContext(ctx).ChainID(),
	))
	require.NoError(t, err)
	res, err = ms.FinalizeTokenDeposit(ctx, types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], addr.String(), sdk.NewCoin(denom, math.ZeroInt()), 4, 1, "test_token", signedTxBz))
	require.NoError(t, err)
	require.Equal(t, types.SUCCESS, res.Result)
	require.Equal(t, math.NewInt(60), input.BankKeeper.GetBalance(ctx, addrs[2], denom).Amount)

	// a base denom that is an IBC voucher on L1 or a slashed denom has a canonical l2 denom too
	for i, base := range []string{transfertypes.ParseDenomTrace("transfer/channel-9/ufoo").IBCDenom(), "move/0123456789abcdef"} {
		baseL2Denom := ophosttypes.L2Denom(testBridgeId, base)
		res, err = ms.FinalizeTokenDeposit(ctx, types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], addrsStr[3], sdk.NewCoin(baseL2Denom, math.NewInt(7)), uint64(5+i), 1, base, nil)) //nolint:gosec
		require.NoError(t, err)
		require.Equal(t, types.SUCCESS, res.Result)
		require.Equal(t, math.NewInt(7), input.BankKeeper.GetBalance(ctx, addrs[3], baseL2Denom).Amount)
	}
}

func Test_MsgServer_Deposit_OldSequence(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	denom := testL2Denom(t, ctx, input, "test_token")

	_, err := ms.FinalizeTokenDeposit(ctx, types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], addrsStr[1], sdk.NewCoin(denom, math.NewInt(100)), 1, 1, "test_token", nil))
	require.NoError(t, err)

	_, err = ms.RegisterMigrationInfo(ctx, types.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(types.ModuleName).String(),
		types.MigrationInfo{Denom: denom, IbcChannelId: "channel-0", IbcPortId: "transfer"},
	))
	require.NoError(t, err)

	// an already finalized sequence is a NOOP before the denom checks: a relayer that resends an old deposit
	// gets the same answer as before, whatever the deposit carries
	for _, coin := range []sdk.Coin{
		sdk.NewCoin(denom, math.NewInt(100)),                // migrated in the meantime
		sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100)), // not canonical
	} {
		res, err := ms.FinalizeTokenDeposit(ctx, types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], addrsStr[1], coin, 1, 1, "test_token", nil))
		require.NoError(t, err)
		require.Equal(t, types.NOOP, res.Result)
	}
	require.Equal(t, math.NewInt(100), input.BankKeeper.GetBalance(ctx, addrs[1], denom).Amount)

	// a future sequence is still refused with the sequence error
	_, err = ms.FinalizeTokenDeposit(ctx, types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], addrsStr[1], sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100)), 3, 1, "test_token", nil))
	require.ErrorIs(t, err, types.ErrInvalidSequence)
}

func Test_MsgServer_Deposit_NoBridgeInfo(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	// no bridge info is set: the l2 denom of a deposit cannot be derived
	denom := ophosttypes.L2Denom(testBridgeId, "test_token")
	l2Sequence, err := input.OPChildKeeper.GetNextL2Sequence(ctx)
	require.NoError(t, err)

	eventCtx := sdk.UnwrapSDKContext(ctx).WithEventManager(sdk.NewEventManager())
	_, err = ms.FinalizeTokenDeposit(eventCtx, types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], addrsStr[1], sdk.NewCoin(denom, math.NewInt(100)), 1, 1, "test_token", nil))
	require.ErrorIs(t, err, types.ErrBridgeInfoNotExists)
	requireNoDepositEffect(t, eventCtx, input, denom, 1, l2Sequence, math.ZeroInt(), false)

	// once the bridge info is set the same deposit is accepted
	setTestBridgeInfo(t, ctx, input)
	res, err := ms.FinalizeTokenDeposit(ctx, types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], addrsStr[1], sdk.NewCoin(denom, math.NewInt(100)), 1, 1, "test_token", nil))
	require.NoError(t, err)
	require.Equal(t, types.SUCCESS, res.Result)
}
