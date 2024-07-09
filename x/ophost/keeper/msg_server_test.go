package keeper_test

import (
	"encoding/base64"
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
	_, err := ms.RecordBatch(ctx, types.NewMsgRecordBatch(addrsStr[0], 1, []byte{1, 2, 3}))
	require.NoError(t, err)
}

func Test_CreateBridge(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	params := input.OPHostKeeper.GetParams(ctx)
	params.RegistrationFee = sdk.NewCoins(sdk.NewCoin("foo", math.NewInt(100)))
	require.NoError(t, input.OPHostKeeper.SetParams(ctx, params))

	input.Faucet.Fund(ctx, addrs[0], sdk.NewCoin("foo", math.NewInt(1000)))

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)
	config := types.BridgeConfig{
		Challengers:         []string{addrsStr[0]},
		Proposer:            addrsStr[0],
		SubmissionInterval:  time.Second * 10,
		FinalizationPeriod:  time.Second * 60,
		SubmissionStartTime: time.Now().UTC(),
		Metadata:            []byte{1, 2, 3},
		BatchInfo:           types.BatchInfo{Submitter: addrsStr[0], Chain: "l1"},
	}
	res, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(addrsStr[0], config))
	require.NoError(t, err)
	require.Equal(t, uint64(1), res.BridgeId)

	_config, err := input.OPHostKeeper.GetBridgeConfig(ctx, res.BridgeId)
	require.NoError(t, err)
	require.Equal(t, config, _config)

	// check community pool
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("foo", math.NewInt(100))), input.CommunityPoolKeeper.CommunityPool)
}

func Test_ProposeOutput(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)
	config := types.BridgeConfig{
		Challengers:         []string{addrsStr[0]},
		Proposer:            addrsStr[0],
		SubmissionInterval:  time.Second * 10,
		FinalizationPeriod:  time.Second * 60,
		SubmissionStartTime: time.Now().UTC(),
		Metadata:            []byte{1, 2, 3},
		BatchInfo:           types.BatchInfo{Submitter: addrsStr[0], Chain: "l1"},
	}
	createRes, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(addrsStr[0], config))
	require.NoError(t, err)
	require.Equal(t, uint64(1), createRes.BridgeId)

	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// unauthorized
	_, err = ms.ProposeOutput(ctx, types.NewMsgProposeOutput(addrsStr[1], 1, 100, []byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
	require.Error(t, err)

	// valid
	proposeRes, err := ms.ProposeOutput(ctx, types.NewMsgProposeOutput(addrsStr[0], 1, 100, []byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
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
		Proposer:            addrsStr[0],
		Challengers:         []string{addrsStr[1]},
		SubmissionInterval:  time.Second * 10,
		FinalizationPeriod:  time.Second * 60,
		SubmissionStartTime: time.Now().UTC(),
		Metadata:            []byte{1, 2, 3},
		BatchInfo:           types.BatchInfo{Submitter: addrsStr[0], Chain: "l1"},
	}
	createReq := types.NewMsgCreateBridge(addrsStr[0], config)
	createRes, err := ms.CreateBridge(ctx, createReq)
	require.NoError(t, err)
	require.Equal(t, uint64(1), createRes.BridgeId)

	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	proposeRes, err := ms.ProposeOutput(ctx, types.NewMsgProposeOutput(addrsStr[0], 1, 100, []byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
	require.NoError(t, err)
	require.Equal(t, uint64(1), proposeRes.OutputIndex)

	proposeRes, err = ms.ProposeOutput(ctx, types.NewMsgProposeOutput(addrsStr[0], 1, 200, []byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
	require.NoError(t, err)
	require.Equal(t, uint64(2), proposeRes.OutputIndex)

	// unauthorized
	_, err = ms.DeleteOutput(ctx, types.NewMsgDeleteOutput(addrsStr[0], 1, 1))
	require.Error(t, err)

	// valid
	_, err = ms.DeleteOutput(ctx, types.NewMsgDeleteOutput(addrsStr[1], 1, 1))
	require.NoError(t, err)

	// should return error; deleted
	_, err = input.OPHostKeeper.GetOutputProposal(ctx, 1, 2)
	require.Error(t, err)

	_, err = input.OPHostKeeper.GetOutputProposal(ctx, 1, 1)
	require.Error(t, err)

	// should be able to resubmit the same output
	proposeRes, err = ms.ProposeOutput(ctx, types.NewMsgProposeOutput(addrsStr[0], 1, 100, []byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
	require.NoError(t, err)
	require.Equal(t, uint64(1), proposeRes.OutputIndex)

	// invalid output index: nextoutputindex is 2 now
	_, err = ms.DeleteOutput(ctx, types.NewMsgDeleteOutput(addrsStr[1], 1, 2))
	require.Error(t, err)
}

func Test_InitiateTokenDeposit(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)
	config := types.BridgeConfig{
		Proposer:            addrsStr[0],
		Challengers:         []string{addrsStr[1]},
		SubmissionInterval:  time.Second * 10,
		FinalizationPeriod:  time.Second * 60,
		SubmissionStartTime: time.Now().UTC(),
		Metadata:            []byte{1, 2, 3},
		BatchInfo:           types.BatchInfo{Submitter: addrsStr[0], Chain: "l1"},
	}
	createRes, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(addrsStr[0], config))
	require.NoError(t, err)
	require.Equal(t, uint64(1), createRes.BridgeId)

	amount := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100))
	input.Faucet.Fund(ctx, addrs[1], amount)
	_, err = ms.InitiateTokenDeposit(
		ctx,
		types.NewMsgInitiateTokenDeposit(addrsStr[1], 1, "hook", amount, []byte("messages")),
	)
	require.NoError(t, err)
	require.True(t, input.BankKeeper.GetBalance(ctx, addrs[1], sdk.DefaultBondDenom).IsZero())
	require.Equal(t, amount, input.BankKeeper.GetBalance(ctx, types.BridgeAddress(1), sdk.DefaultBondDenom))
}

func Test_FinalizeTokenWithdrawal(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)
	config := types.BridgeConfig{
		Proposer:            addrsStr[0],
		Challengers:         []string{addrsStr[1]},
		SubmissionInterval:  time.Second * 10,
		FinalizationPeriod:  time.Second * 60,
		SubmissionStartTime: time.Now().UTC(),
		Metadata:            []byte{1, 2, 3},
		BatchInfo:           types.BatchInfo{Submitter: addrsStr[0], Chain: "l1"},
	}
	_, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(addrsStr[0], config))
	require.NoError(t, err)

	// fund amount
	amount := sdk.NewCoin("uinit", math.NewInt(1_000_000))
	input.Faucet.Fund(ctx, types.BridgeAddress(1), amount)

	outputRoot := decodeBase64(t, "at+mtcWpUvvV+K/uBm+tRufWD0WH4SZVskw9WKn5N/A=")
	version := decodeBase64(t, "Ch4nNnd/gKYr6y33K2SYeEgcDKEBlLgytRNr77rlQBc=")
	stateRoot := decodeBase64(t, "C2ZdjJ7uX41NaadA/FjlMiG6btiDfYnxE2ABqJocHxI=")
	storageRoot := decodeBase64(t, "8EorDbcn/PYtbfU8+e35gHR5e/Liy/mycsULLPWzJww=")
	blockHash := decodeBase64(t, "tgmfQJT4uipVToW631xz0RXdrfzu7n5XxGNoPpX6isI=")
	proofs := [][]byte{
		decodeBase64(t, "Ux19nu4Nl3N7gBy/ID3rzuXrRpScnxOR9u/PUCxlTC0="),
		decodeBase64(t, "vWrbSRyDJ+FnWxY5Plr7Ltgyyusr/uDW7nQDq8PDDQY="),
		decodeBase64(t, "opvMy3Dv9tUfa4pNr/IBM1GOw8qOlwfxqoXVH5gKPIo="),
	}

	now := time.Now().UTC()
	ctx = ctx.WithBlockTime(now)
	_, err = ms.ProposeOutput(ctx, types.NewMsgProposeOutput(addrsStr[0], 1, 100, outputRoot))
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(now.Add(time.Second * 60))
	sender := "osmo174knscjg688ddtxj8smyjz073r3w5mms8ugvx6"
	receiver := "cosmos174knscjg688ddtxj8smyjz073r3w5mms08musg"
	require.NoError(t, err)
	_, err = ms.FinalizeTokenWithdrawal(ctx, types.NewMsgFinalizeTokenWithdrawal(
		1, 1, 1, proofs,
		sender,
		receiver,
		amount,
		version, stateRoot, storageRoot, blockHash,
	))
	require.NoError(t, err)

	receiverAddr, err := sdk.AccAddressFromBech32("cosmos174knscjg688ddtxj8smyjz073r3w5mms08musg")
	require.NoError(t, err)
	require.Equal(t, amount, input.BankKeeper.GetBalance(ctx, receiverAddr, amount.Denom))
}

func decodeBase64(t *testing.T, str string) []byte {
	bz, err := base64.StdEncoding.DecodeString(str)
	require.NoError(t, err)

	return bz
}

func Test_UpdateProposal(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	config := types.BridgeConfig{
		Proposer:            addrsStr[0],
		Challengers:         []string{addrsStr[1]},
		SubmissionInterval:  time.Second * 10,
		FinalizationPeriod:  time.Second * 60,
		SubmissionStartTime: time.Now().UTC(),
		Metadata:            []byte{1, 2, 3},
		BatchInfo:           types.BatchInfo{Submitter: addrsStr[0], Chain: "l1"},
	}

	_, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(addrsStr[0], config))
	require.NoError(t, err)

	// gov signer
	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress("gov"))
	require.NoError(t, err)
	msg := types.NewMsgUpdateProposer(govAddr, 1, addrsStr[1])
	_, err = ms.UpdateProposer(ctx, msg)
	require.NoError(t, err)
	_config, err := ms.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, addrsStr[1], _config.Proposer)
	require.Equal(t, addrsStr[1], input.BridgeHook.proposer)

	// current proposer signer
	msg = types.NewMsgUpdateProposer(addrsStr[1], 1, addrsStr[2])
	_, err = ms.UpdateProposer(ctx, msg)
	require.NoError(t, err)
	_config, err = ms.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, addrs[2].String(), _config.Proposer)
	require.Equal(t, addrs[2].String(), input.BridgeHook.proposer)

	// invalid signer
	invalidAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)
	msg = types.NewMsgUpdateProposer(invalidAddr, 1, addrsStr[1])
	require.NoError(t, err)

	_, err = ms.UpdateProposer(
		ctx,
		msg,
	)
	require.Error(t, err)
}

func Test_UpdateChallengers(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	config := types.BridgeConfig{
		Proposer:            addrsStr[0],
		Challengers:         []string{addrsStr[1], addrsStr[2], addrsStr[3]},
		SubmissionInterval:  time.Second * 10,
		FinalizationPeriod:  time.Second * 60,
		SubmissionStartTime: time.Now().UTC(),
		Metadata:            []byte{1, 2, 3},
		BatchInfo:           types.BatchInfo{Submitter: addrsStr[0], Chain: "l1"},
	}

	_, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(addrsStr[0], config))
	require.NoError(t, err)

	// gov signer
	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress("gov"))
	require.NoError(t, err)
	msg := types.NewMsgUpdateChallengers(govAddr, 1, []string{addrsStr[2], addrsStr[3]})
	_, err = ms.UpdateChallengers(ctx, msg)
	require.NoError(t, err)
	_config, err := ms.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, []string{addrsStr[2], addrsStr[3]}, _config.Challengers)
	require.Equal(t, input.BridgeHook.challengers, _config.Challengers)

	// current challenger

	// case 1. replace oneself
	msg = types.NewMsgUpdateChallengers(addrsStr[2], 1, []string{addrsStr[3], addrsStr[4]})
	_, err = ms.UpdateChallengers(ctx, msg)
	require.NoError(t, err)
	_config, err = ms.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, []string{addrsStr[3], addrsStr[4]}, _config.Challengers)
	require.Equal(t, input.BridgeHook.challengers, _config.Challengers)

	// case 2. try to remove other challenger
	msg = types.NewMsgUpdateChallengers(addrsStr[4], 1, []string{addrsStr[4]})
	_, err = ms.UpdateChallengers(ctx, msg)
	require.Error(t, err)

	// case 2. try to replace other challenger
	msg = types.NewMsgUpdateChallengers(addrsStr[4], 1, []string{addrsStr[2], addrsStr[4]})
	_, err = ms.UpdateChallengers(ctx, msg)
	require.Error(t, err)

	// case 3. remove oneself
	msg = types.NewMsgUpdateChallengers(addrsStr[3], 1, []string{addrsStr[4]})
	_, err = ms.UpdateChallengers(ctx, msg)
	require.NoError(t, err)
	_config, err = ms.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, []string{addrsStr[4]}, _config.Challengers)
	require.Equal(t, input.BridgeHook.challengers, _config.Challengers)

	// case 4. try to add more challenger
	msg = types.NewMsgUpdateChallengers(addrsStr[4], 1, []string{addrsStr[3], addrsStr[4]})
	_, err = ms.UpdateChallengers(ctx, msg)
	require.Error(t, err)

	// case 5. try to add more challenger with replace
	msg = types.NewMsgUpdateChallengers(addrsStr[4], 1, []string{addrsStr[2], addrsStr[3]})
	_, err = ms.UpdateChallengers(ctx, msg)
	require.Error(t, err)

	// invalid signer
	invalidAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)
	msg = types.NewMsgUpdateChallengers(invalidAddr, 1, []string{addrsStr[1]})
	require.NoError(t, err)

	_, err = ms.UpdateChallengers(
		ctx,
		msg,
	)
	require.Error(t, err)

	// invalid case
	msg = types.NewMsgUpdateChallengers(govAddr, 1, []string{})
	_, err = ms.UpdateChallengers(ctx, msg)
	require.Error(t, err)
}

func Test_UpdateBatchInfo(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	config := types.BridgeConfig{
		Proposer:            addrsStr[0],
		Challengers:         []string{addrsStr[1]},
		SubmissionInterval:  time.Second * 10,
		FinalizationPeriod:  time.Second * 60,
		SubmissionStartTime: time.Now().UTC(),
		Metadata:            []byte{1, 2, 3},
		BatchInfo: types.BatchInfo{
			Submitter: addrsStr[1],
			Chain:     "l1",
		},
	}

	_, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(addrsStr[0], config))
	require.NoError(t, err)

	// gov signer
	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress("gov"))
	require.NoError(t, err)
	msg := types.NewMsgUpdateBatchInfo(govAddr, 1, types.BatchInfo{
		Submitter: addrsStr[2],
		Chain:     "celestia",
	})
	_, err = ms.UpdateBatchInfo(ctx, msg)
	require.NoError(t, err)
	_config, err := ms.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, _config.BatchInfo.Submitter, addrsStr[2])
	require.Equal(t, "celestia", _config.BatchInfo.Chain)
	require.Equal(t, input.BridgeHook.batchInfo, _config.BatchInfo)

	// current proposer signer
	msg = types.NewMsgUpdateBatchInfo(addrsStr[0], 1, types.BatchInfo{
		Submitter: addrsStr[3],
		Chain:     "l1",
	})
	_, err = ms.UpdateBatchInfo(ctx, msg)
	require.NoError(t, err)
	_config, err = ms.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, _config.BatchInfo.Submitter, addrsStr[3])
	require.Equal(t, "l1", _config.BatchInfo.Chain)
	require.Equal(t, input.BridgeHook.batchInfo, _config.BatchInfo)

	// invalid signer
	invalidAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)
	msg = types.NewMsgUpdateBatchInfo(invalidAddr, 1, types.BatchInfo{
		Submitter: addrsStr[2],
		Chain:     "celestia",
	})
	require.NoError(t, err)

	_, err = ms.UpdateBatchInfo(
		ctx,
		msg,
	)
	require.Error(t, err)
}

func Test_UpdateMetadata(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	config := types.BridgeConfig{
		Proposer:            addrsStr[0],
		Challengers:         []string{addrsStr[1]},
		SubmissionInterval:  time.Second * 10,
		FinalizationPeriod:  time.Second * 60,
		SubmissionStartTime: time.Now().UTC(),
		Metadata:            []byte{1, 2, 3},
		BatchInfo:           types.BatchInfo{Submitter: addrsStr[0], Chain: "l1"},
	}

	_, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(addrsStr[0], config))
	require.NoError(t, err)

	// gov signer
	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress("gov"))
	require.NoError(t, err)
	msg := types.NewMsgUpdateMetadata(govAddr, 1, []byte{4, 5, 6})
	_, err = ms.UpdateMetadata(ctx, msg)
	require.NoError(t, err)
	_config, err := ms.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, []byte{4, 5, 6}, _config.Metadata)
	require.Equal(t, []byte{4, 5, 6}, input.BridgeHook.metadata)

	// current challenger
	msg = types.NewMsgUpdateMetadata(addrsStr[0], 1, []byte{7, 8, 9})
	_, err = ms.UpdateMetadata(ctx, msg)
	require.NoError(t, err)
	_config, err = ms.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, []byte{7, 8, 9}, _config.Metadata)
	require.Equal(t, []byte{7, 8, 9}, input.BridgeHook.metadata)

	// invalid signer
	invalidAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)
	msg = types.NewMsgUpdateMetadata(invalidAddr, 1, []byte{1, 2, 3})
	require.NoError(t, err)

	_, err = ms.UpdateMetadata(
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

	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress("gov"))
	require.NoError(t, err)
	msg := types.NewMsgUpdateParams(govAddr, &params)
	_, err = ms.UpdateParams(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, params, ms.GetParams(ctx))

	// invalid signer
	invalidAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)
	msg = types.NewMsgUpdateParams(invalidAddr, &params)
	require.NoError(t, err)

	_, err = ms.UpdateParams(
		ctx,
		msg,
	)
	require.Error(t, err)
}
