package keeper_test

import (
	"encoding/hex"
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"

	"github.com/initia-labs/OPinit/x/ophost/keeper"
	"github.com/initia-labs/OPinit/x/ophost/types"
)

func Test_RecordBatch(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)
	_, err := ms.RecordBatch(ctx, types.NewMsgRecordBatch(addrs[0], 1, []byte{1, 2, 3}))
	require.NoError(t, err)
}

func Test_CreateBridge(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)
	config := types.BridgeConfig{
		Challenger:          addrs[0].String(),
		Proposer:            addrs[0].String(),
		SubmissionInterval:  time.Second * 10,
		FinalizationPeriod:  time.Second * 60,
		SubmissionStartTime: time.Now().UTC(),
		Metadata:            []byte{1, 2, 3},
	}
	res, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(addrs[0], config))
	require.NoError(t, err)
	require.Equal(t, uint64(1), res.BridgeId)

	_config, err := input.OPHostKeeper.GetBridgeConfig(ctx, res.BridgeId)
	require.NoError(t, err)
	require.Equal(t, config, _config)
}

func Test_ProposeOutput(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)
	config := types.BridgeConfig{
		Challenger:          addrs[0].String(),
		Proposer:            addrs[0].String(),
		SubmissionInterval:  time.Second * 10,
		FinalizationPeriod:  time.Second * 60,
		SubmissionStartTime: time.Now().UTC(),
		Metadata:            []byte{1, 2, 3},
	}
	createRes, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(addrs[0], config))
	require.NoError(t, err)
	require.Equal(t, uint64(1), createRes.BridgeId)

	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// unauthorized
	_, err = ms.ProposeOutput(ctx, types.NewMsgProposeOutput(addrs[1], 1, 100, []byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
	require.Error(t, err)

	// valid
	proposeRes, err := ms.ProposeOutput(ctx, types.NewMsgProposeOutput(addrs[0], 1, 100, []byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
	require.NoError(t, err)
	require.Equal(t, uint64(1), proposeRes.OutputIndex)

	output, err := input.OPHostKeeper.GetOutputProposal(ctx, 1, 1)
	require.NoError(t, err)
	require.Equal(t, types.Output{
		OutputRoot:    []byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		L1BlockTime:   blockTime,
		L2BlockNumber: 100,
	}, output)
}

func Test_DeleteOutput(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)
	config := types.BridgeConfig{
		Proposer:            addrs[0].String(),
		Challenger:          addrs[1].String(),
		SubmissionInterval:  time.Second * 10,
		FinalizationPeriod:  time.Second * 60,
		SubmissionStartTime: time.Now().UTC(),
		Metadata:            []byte{1, 2, 3},
	}
	createReq := types.NewMsgCreateBridge(addrs[0], config)
	createRes, err := ms.CreateBridge(ctx, createReq)
	require.NoError(t, err)
	require.Equal(t, uint64(1), createRes.BridgeId)

	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	proposeRes, err := ms.ProposeOutput(ctx, types.NewMsgProposeOutput(addrs[0], 1, 100, []byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
	require.NoError(t, err)
	require.Equal(t, uint64(1), proposeRes.OutputIndex)

	// unauthorized
	_, err = ms.DeleteOutput(ctx, types.NewMsgDeleteOutput(addrs[0], 1, 1))
	require.Error(t, err)

	// valid
	_, err = ms.DeleteOutput(ctx, types.NewMsgDeleteOutput(addrs[1], 1, 1))
	require.NoError(t, err)

	// should return error; deleted
	_, err = input.OPHostKeeper.GetOutputProposal(ctx, 1, 1)
	require.Error(t, err)
}

func Test_InitiateTokenDeposit(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)
	config := types.BridgeConfig{
		Proposer:            addrs[0].String(),
		Challenger:          addrs[1].String(),
		SubmissionInterval:  time.Second * 10,
		FinalizationPeriod:  time.Second * 60,
		SubmissionStartTime: time.Now().UTC(),
		Metadata:            []byte{1, 2, 3},
	}
	createRes, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(addrs[0], config))
	require.NoError(t, err)
	require.Equal(t, uint64(1), createRes.BridgeId)

	amount := sdk.NewCoin(baseDenom, math.NewInt(100))
	input.Faucet.Fund(ctx, addrs[1], amount)
	_, err = ms.InitiateTokenDeposit(
		ctx,
		types.NewMsgInitiateTokenDeposit(addrs[1], 1, addrs[2], amount, []byte("messages")),
	)
	require.NoError(t, err)
	require.True(t, input.BankKeeper.GetBalance(ctx, addrs[1], baseDenom).IsZero())
	require.Equal(t, amount, input.BankKeeper.GetBalance(ctx, types.BridgeAddress(1), baseDenom))
}

func Test_FinalizeTokenWithdrawal(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)
	config := types.BridgeConfig{
		Proposer:            addrs[0].String(),
		Challenger:          addrs[1].String(),
		SubmissionInterval:  time.Second * 10,
		FinalizationPeriod:  time.Second * 60,
		SubmissionStartTime: time.Now().UTC(),
		Metadata:            []byte{1, 2, 3},
	}
	_, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(addrs[0], config))
	require.NoError(t, err)

	// fund amount
	amount := sdk.NewCoin("l1denom", math.NewInt(3_000_000))
	input.Faucet.Fund(ctx, types.BridgeAddress(1), amount)

	outputRoot := decodeHex(t, "d87b15f515e52e234f5ddca84627128ad842fa6c741d6b85d589a13bbdad3a89")
	version := decodeHex(t, "0000000000000000000000000000000000000000000000000000000000000001")
	stateRoot := decodeHex(t, "0000000000000000000000000000000000000000000000000000000000000002")
	storageRoot := decodeHex(t, "326ca35f4738f837ad9f335349fc71bdecf4c4ed3485fff1763d3bab55efc88a")
	blockHash := decodeHex(t, "0000000000000000000000000000000000000000000000000000000000000003")
	proofs := [][]byte{
		decodeHex(t, "32e1a72a7c215563f9426bfe267b6fa22ba49b1fba7162d80094dc2f2b6c5a3a"),
		decodeHex(t, "627dc2af9ee001b0e119100599dc3923ccdff2c53f06d89f40400edb1e7907e1"),
		decodeHex(t, "bafac86e9ebc05a07701c151846c6de7bca68cd315f7a82fffe05fc4301ac47e"),
	}

	now := time.Now().UTC()
	ctx = ctx.WithBlockTime(now)
	_, err = ms.ProposeOutput(ctx, types.NewMsgProposeOutput(addrs[0], 1, 100, outputRoot))
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(now.Add(time.Second * 60))
	_, err = ms.FinalizeTokenWithdrawal(ctx, types.NewMsgFinalizeTokenWithdrawal(
		1, 1, 4, proofs,
		decodeHex(t, "0000000000000000000000000000000000000004"),
		decodeHex(t, "0000000000000000000000000000000000000001"),
		amount,
		version, stateRoot, storageRoot, blockHash,
	))
	require.NoError(t, err)
	require.Equal(t, amount, input.BankKeeper.GetBalance(ctx, decodeHex(t, "0000000000000000000000000000000000000001"), amount.Denom))
}

func decodeHex(t *testing.T, str string) []byte {
	bz, err := hex.DecodeString(str)
	require.NoError(t, err)

	return bz
}

func Test_UpdateProposal(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	config := types.BridgeConfig{
		Proposer:            addrs[0].String(),
		Challenger:          addrs[1].String(),
		SubmissionInterval:  time.Second * 10,
		FinalizationPeriod:  time.Second * 60,
		SubmissionStartTime: time.Now().UTC(),
		Metadata:            []byte{1, 2, 3},
	}

	_, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(addrs[0], config))
	require.NoError(t, err)

	// gov signer
	msg := types.NewMsgUpdateProposer(authtypes.NewModuleAddress("gov"), 1, addrs[1])
	_, err = ms.UpdateProposer(ctx, msg)
	require.NoError(t, err)
	_config, err := ms.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, addrs[1].String(), _config.Proposer)
	require.Equal(t, addrs[1].String(), input.BridgeHook.proposer)

	// current proposer signer
	msg = types.NewMsgUpdateProposer(addrs[1], 1, addrs[2])
	_, err = ms.UpdateProposer(sdk.WrapSDKContext(ctx), msg)
	require.NoError(t, err)
	_config, err = ms.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, addrs[2].String(), _config.Proposer)
	require.Equal(t, addrs[2].String(), input.BridgeHook.proposer)

	// invalid signer
	msg = types.NewMsgUpdateProposer(authtypes.NewModuleAddress(types.ModuleName), 1, addrs[1])
	require.NoError(t, err)

	_, err = ms.UpdateProposer(
		ctx,
		msg,
	)
	require.Error(t, err)
}

func Test_UpdateChallenger(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	config := types.BridgeConfig{
		Proposer:            addrs[0].String(),
		Challenger:          addrs[1].String(),
		SubmissionInterval:  time.Second * 10,
		FinalizationPeriod:  time.Second * 60,
		SubmissionStartTime: time.Now().UTC(),
		Metadata:            []byte{1, 2, 3},
	}

	_, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(addrs[0], config))
	require.NoError(t, err)

	// gov signer
	msg := types.NewMsgUpdateChallenger(authtypes.NewModuleAddress("gov"), 1, addrs[2])
	_, err = ms.UpdateChallenger(ctx, msg)
	require.NoError(t, err)
	_config, err := ms.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, addrs[2].String(), _config.Challenger)
	require.Equal(t, addrs[2].String(), input.BridgeHook.challenger)

	// current challenger
	msg = types.NewMsgUpdateChallenger(addrs[2], 1, addrs[3])
	_, err = ms.UpdateChallenger(sdk.WrapSDKContext(ctx), msg)
	require.NoError(t, err)
	_config, err = ms.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, addrs[3].String(), _config.Challenger)
	require.Equal(t, addrs[3].String(), input.BridgeHook.challenger)

	// invalid signer
	msg = types.NewMsgUpdateChallenger(authtypes.NewModuleAddress(types.ModuleName), 1, addrs[1])
	require.NoError(t, err)

	_, err = ms.UpdateChallenger(
		ctx,
		msg,
	)
	require.Error(t, err)
}

func Test_MsgServer_UpdateParams(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	params := ms.GetParams(ctx)
	params.RegistrationFee = sdk.NewCoins(sdk.NewCoin("foo", math.NewInt(100)))

	msg := types.NewMsgUpdateParams(authtypes.NewModuleAddress("gov"), &params)
	_, err := ms.UpdateParams(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, params, ms.GetParams(ctx))

	// invalid signer
	msg = types.NewMsgUpdateParams(authtypes.NewModuleAddress(types.ModuleName), &params)
	require.NoError(t, err)

	_, err = ms.UpdateParams(
		ctx,
		msg,
	)
	require.Error(t, err)
}
