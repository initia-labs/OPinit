package keeper_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"

	"cosmossdk.io/math"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/initia-labs/OPinit/x/opchild/keeper"
	"github.com/initia-labs/OPinit/x/opchild/testutil"
	"github.com/initia-labs/OPinit/x/opchild/types"
)

func Test_MsgServer_Deposit_HookEvents(t *testing.T) {
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
	from, _ := input.AccountKeeper.AddressCodec().BytesToString(addr)
	to, _ := input.AccountKeeper.AddressCodec().BytesToString(testutil.Addrs[2])

	signedTxBz, err := input.EncodingConfig.TxConfig.TxEncoder()(testutil.GenerateTestTx(
		t, input,
		[]sdk.Msg{
			types.NewMsgInitiateTokenWithdrawal(from, to, sdk.NewCoin(denom, math.NewInt(50))),
		},
		privs, accNums, accSeqs, sdk.UnwrapSDKContext(ctx).ChainID(),
	))
	require.NoError(t, err)

	// valid deposit
	ctx = sdk.UnwrapSDKContext(ctx).WithEventManager(sdk.NewEventManager())
	msg = types.NewMsgFinalizeTokenDeposit(testutil.AddrsStr[0], testutil.AddrsStr[1], addr.String(), sdk.NewCoin(denom, math.NewInt(100)), 2, 1, "test_token", signedTxBz)
	_, err = ms.FinalizeTokenDeposit(ctx, msg)
	require.NoError(t, err)

	withdrawalEventFound := false
	for _, event := range sdk.UnwrapSDKContext(ctx).EventManager().Events() {
		if event.Type == types.EventTypeInitiateTokenWithdrawal {
			withdrawalEventFound = true
		} else if event.Type == types.EventTypeFinalizeTokenDeposit {
			require.Equal(t, event.Attributes[len(event.Attributes)-1].Key, types.AttributeKeySuccess)
			require.Equal(t, event.Attributes[len(event.Attributes)-1].Value, "true")
		}
	}

	// but no event was emitted by hooks
	require.True(t, withdrawalEventFound)
}
