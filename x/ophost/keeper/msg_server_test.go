package keeper_test

import (
	"encoding/base64"
	"testing"
	"time"

	"cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
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
		Challenger:            addrsStr[0],
		Proposer:              addrsStr[0],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: addrsStr[0], ChainType: types.BatchInfo_INITIA},
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
		Challenger:            addrsStr[0],
		Proposer:              addrsStr[0],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: addrsStr[0], ChainType: types.BatchInfo_INITIA},
	}
	createRes, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(addrsStr[0], config))
	require.NoError(t, err)
	require.Equal(t, uint64(1), createRes.BridgeId)

	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// unauthorized
	_, err = ms.ProposeOutput(ctx, types.NewMsgProposeOutput(addrsStr[1], 1, 1, 100, []byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
	require.Error(t, err)

	// valid
	_, err = ms.ProposeOutput(ctx, types.NewMsgProposeOutput(addrsStr[0], 1, 1, 100, []byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
	require.NoError(t, err)

	output, err := input.OPHostKeeper.GetOutputProposal(ctx, 1, 1)
	require.NoError(t, err)
	require.Equal(t, types.Output{
		OutputRoot:    []byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		L1BlockNumber: uint64(ctx.BlockHeight()),
		L1BlockTime:   blockTime,
		L2BlockNumber: 100,
	}, output)
}

func Test_DeleteOutput(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)
	config := types.BridgeConfig{
		Proposer:              addrsStr[0],
		Challenger:            addrsStr[1],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: addrsStr[0], ChainType: types.BatchInfo_INITIA},
	}
	createReq := types.NewMsgCreateBridge(addrsStr[0], config)
	createRes, err := ms.CreateBridge(ctx, createReq)
	require.NoError(t, err)
	require.Equal(t, uint64(1), createRes.BridgeId)

	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	_, err = ms.ProposeOutput(ctx, types.NewMsgProposeOutput(addrsStr[0], 1, 1, 100, []byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
	require.NoError(t, err)

	_, err = ms.ProposeOutput(ctx, types.NewMsgProposeOutput(addrsStr[0], 1, 2, 200, []byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
	require.NoError(t, err)

	// unauthorized
	_, err = ms.DeleteOutput(ctx, types.NewMsgDeleteOutput(addrsStr[2], 1, 1))
	require.Error(t, err)

	// valid by challenger
	_, err = ms.DeleteOutput(ctx, types.NewMsgDeleteOutput(addrsStr[1], 1, 1))
	require.NoError(t, err)

	// should return error; deleted
	_, err = input.OPHostKeeper.GetOutputProposal(ctx, 1, 2)
	require.Error(t, err)

	_, err = input.OPHostKeeper.GetOutputProposal(ctx, 1, 1)
	require.Error(t, err)

	// should be able to resubmit the same output
	_, err = ms.ProposeOutput(ctx, types.NewMsgProposeOutput(addrsStr[0], 1, 1, 100, []byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
	require.NoError(t, err)

	// invalid output index: nextoutputindex is 2 now
	_, err = ms.DeleteOutput(ctx, types.NewMsgDeleteOutput(addrsStr[1], 1, 2))
	require.Error(t, err)

	// valid delete by gov
	_, err = ms.DeleteOutput(ctx, types.NewMsgDeleteOutput(input.OPHostKeeper.GetAuthority(), 1, 1))
	require.NoError(t, err)

	// should be able to resubmit the same output
	_, err = ms.ProposeOutput(ctx, types.NewMsgProposeOutput(addrsStr[0], 1, 1, 100, []byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
	require.NoError(t, err)

	// valid delete by proposer
	_, err = ms.DeleteOutput(ctx, types.NewMsgDeleteOutput(addrsStr[0], 1, 1))
	require.NoError(t, err)
}

func Test_InitiateTokenDeposit(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)
	config := types.BridgeConfig{
		Proposer:              addrsStr[0],
		Challenger:            addrsStr[1],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: addrsStr[0], ChainType: types.BatchInfo_INITIA},
	}
	createRes, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(addrsStr[0], config))
	require.NoError(t, err)
	require.Equal(t, uint64(1), createRes.BridgeId)

	amount := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100))
	input.Faucet.Fund(ctx, addrs[1], amount)
	_, err = ms.InitiateTokenDeposit(
		ctx,
		types.NewMsgInitiateTokenDeposit(addrsStr[1], 1, "l2_addr", amount, []byte("messages")),
	)
	require.NoError(t, err)
	require.True(t, input.BankKeeper.GetBalance(ctx, addrs[1], sdk.DefaultBondDenom).IsZero())
	require.Equal(t, amount, input.BankKeeper.GetBalance(ctx, types.BridgeAddress(1), sdk.DefaultBondDenom))

	// not existing bridge
	_, err = ms.InitiateTokenDeposit(
		ctx,
		types.NewMsgInitiateTokenDeposit(addrsStr[1], 2, "l2_addr", amount, []byte("messages")),
	)
	require.Error(t, err)
}

func Test_FinalizeTokenWithdrawal(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)
	config := types.BridgeConfig{
		Proposer:              addrsStr[0],
		Challenger:            addrsStr[1],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: addrsStr[0], ChainType: types.BatchInfo_INITIA},
	}
	_, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(addrsStr[0], config))
	require.NoError(t, err)

	// fund amount
	amount := sdk.NewCoin("uinit", math.NewInt(1_000_000))
	input.Faucet.Fund(ctx, types.BridgeAddress(1), amount)

	sender := "osmo174knscjg688ddtxj8smyjz073r3w5mms8ugvx6"
	receiver := "cosmos174knscjg688ddtxj8smyjz073r3w5mms08musg"

	version := byte(1)

	withdrawal1 := types.GenerateWithdrawalHash(1, 1, sender, receiver, amount.Denom, amount.Amount.Uint64())
	withdrawal2 := types.GenerateWithdrawalHash(1, 2, sender, receiver, amount.Denom, amount.Amount.Uint64())
	withdrawal3 := types.GenerateWithdrawalHash(1, 3, sender, receiver, amount.Denom, amount.Amount.Uint64())

	proof1 := withdrawal2
	proof2 := types.GenerateNodeHash(withdrawal3[:], withdrawal3[:])

	node12 := types.GenerateNodeHash(withdrawal1[:], withdrawal2[:])

	storageRoot := types.GenerateNodeHash(node12[:], proof2[:])
	blockHash := decodeBase64(t, "tgmfQJT4uipVToW631xz0RXdrfzu7n5XxGNoPpX6isI=")
	outputRoot := types.GenerateOutputRoot(version, storageRoot[:], blockHash)
	proofs := [][]byte{
		proof1[:],
		proof2[:],
	}

	now := time.Now().UTC()
	ctx = ctx.WithBlockTime(now)
	_, err = ms.ProposeOutput(ctx, types.NewMsgProposeOutput(addrsStr[0], 1, 1, 100, outputRoot[:]))
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(now.Add(time.Second * 60))

	require.NoError(t, err)
	_, err = ms.FinalizeTokenWithdrawal(ctx, types.NewMsgFinalizeTokenWithdrawal(
		addrsStr[3], // any address can execute this
		1, 1, 1, proofs,
		sender,
		receiver,
		amount,
		[]byte{version}, storageRoot[:], blockHash,
	))
	require.NoError(t, err)

	receiverAddr, err := sdk.AccAddressFromBech32(receiver)
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
		Proposer:              addrsStr[0],
		Challenger:            addrsStr[1],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: addrsStr[0], ChainType: types.BatchInfo_INITIA},
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

func Test_UpdateChallenger(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	config := types.BridgeConfig{
		Proposer:              addrsStr[0],
		Challenger:            addrsStr[1],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: addrsStr[0], ChainType: types.BatchInfo_INITIA},
	}

	_, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(addrsStr[0], config))
	require.NoError(t, err)

	// gov signer
	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress("gov"))
	require.NoError(t, err)
	msg := types.NewMsgUpdateChallenger(govAddr, 1, addrsStr[2])
	_, err = ms.UpdateChallenger(ctx, msg)
	require.NoError(t, err)
	_config, err := ms.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, addrsStr[2], _config.Challenger)
	require.Equal(t, input.BridgeHook.challenger, _config.Challenger)

	// current challenger

	// invalid signer
	invalidAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)
	msg = types.NewMsgUpdateChallenger(invalidAddr, 1, addrsStr[1])
	require.NoError(t, err)

	_, err = ms.UpdateChallenger(
		ctx,
		msg,
	)
	require.Error(t, err)

	// invalid case
	msg = types.NewMsgUpdateChallenger(govAddr, 1, "")
	_, err = ms.UpdateChallenger(ctx, msg)
	require.Error(t, err)
}

func Test_UpdateBatchInfo(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	config := types.BridgeConfig{
		Proposer:              addrsStr[0],
		Challenger:            addrsStr[1],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo: types.BatchInfo{
			Submitter: addrsStr[1],
			ChainType: types.BatchInfo_INITIA,
		},
	}

	_, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(addrsStr[0], config))
	require.NoError(t, err)

	// gov signer
	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress("gov"))
	require.NoError(t, err)
	msg := types.NewMsgUpdateBatchInfo(govAddr, 1, types.BatchInfo{
		Submitter: addrsStr[2],
		ChainType: types.BatchInfo_CELESTIA,
	})
	_, err = ms.UpdateBatchInfo(ctx, msg)
	require.NoError(t, err)
	_config, err := ms.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, _config.BatchInfo.Submitter, addrsStr[2])
	require.Equal(t, types.BatchInfo_CELESTIA, _config.BatchInfo.ChainType)
	require.Equal(t, input.BridgeHook.batchInfo, _config.BatchInfo)

	// current proposer signer
	msg = types.NewMsgUpdateBatchInfo(addrsStr[0], 1, types.BatchInfo{
		Submitter: addrsStr[3],
		ChainType: types.BatchInfo_INITIA,
	})
	_, err = ms.UpdateBatchInfo(ctx, msg)
	require.NoError(t, err)
	_config, err = ms.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, _config.BatchInfo.Submitter, addrsStr[3])
	require.Equal(t, types.BatchInfo_INITIA, _config.BatchInfo.ChainType)
	require.Equal(t, input.BridgeHook.batchInfo, _config.BatchInfo)

	// invalid signer
	invalidAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)
	msg = types.NewMsgUpdateBatchInfo(invalidAddr, 1, types.BatchInfo{
		Submitter: addrsStr[2],
		ChainType: types.BatchInfo_CELESTIA,
	})
	require.NoError(t, err)

	_, err = ms.UpdateBatchInfo(
		ctx,
		msg,
	)
	require.Error(t, err)
}

func Test_UpdateOracleConfig(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	config := types.BridgeConfig{
		Proposer:              addrsStr[0],
		Challenger:            addrsStr[1],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: addrsStr[0], ChainType: types.BatchInfo_INITIA},
		OracleEnabled:         true,
	}

	_, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(addrsStr[0], config))
	require.NoError(t, err)

	// gov signer
	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress("gov"))
	require.NoError(t, err)
	msg := types.NewMsgUpdateOracleConfig(govAddr, 1, false)
	_, err = ms.UpdateOracleConfig(ctx, msg)
	require.NoError(t, err)
	_config, err := ms.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, false, _config.OracleEnabled)

	// current proposer signer
	msg = types.NewMsgUpdateOracleConfig(addrsStr[0], 1, true)
	_, err = ms.UpdateOracleConfig(ctx, msg)
	require.NoError(t, err)
	_config, err = ms.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, true, _config.OracleEnabled)

	// invalid signer
	invalidAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)
	msg = types.NewMsgUpdateOracleConfig(invalidAddr, 1, false)
	require.NoError(t, err)

	_, err = ms.UpdateOracleConfig(ctx, msg)
	require.Error(t, err)
}

func Test_UpdateMetadata(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	config := types.BridgeConfig{
		Proposer:              addrsStr[0],
		Challenger:            addrsStr[1],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: addrsStr[0], ChainType: types.BatchInfo_INITIA},
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

	// current proposer
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

func Test_MsgServer_UpdateFinalizationPeriod(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	config := types.BridgeConfig{
		Proposer:              addrsStr[0],
		Challenger:            addrsStr[1],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: addrsStr[0], ChainType: types.BatchInfo_INITIA},
	}

	_, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(addrsStr[0], config))
	require.NoError(t, err)

	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress("gov"))
	require.NoError(t, err)
	msg := types.NewMsgUpdateFinalizationPeriod(govAddr, 1, time.Second*10)
	_, err = ms.UpdateFinalizationPeriod(ctx, msg)
	require.NoError(t, err)

	msg = types.NewMsgUpdateFinalizationPeriod(govAddr, 1, time.Second*20)
	_, err = ms.UpdateFinalizationPeriod(ctx, msg)
	require.NoError(t, err)

	msg = types.NewMsgUpdateFinalizationPeriod(govAddr, 1, time.Second*30)
	_, err = ms.UpdateFinalizationPeriod(ctx, msg)
	require.NoError(t, err)

	// check finalization period
	config, err = ms.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, time.Second*30, config.FinalizationPeriod)

	// invalid signer
	invalidAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)
	msg = types.NewMsgUpdateFinalizationPeriod(invalidAddr, 1, time.Second*10)
	_, err = ms.UpdateFinalizationPeriod(ctx, msg)
	require.Error(t, err)

	// invalid bridge id
	msg = types.NewMsgUpdateFinalizationPeriod(govAddr, 0, time.Second*10)
	_, err = ms.UpdateFinalizationPeriod(ctx, msg)
	require.Error(t, err)

	// invalid finalization period
	msg = types.NewMsgUpdateFinalizationPeriod(govAddr, 1, time.Second*0)
	_, err = ms.UpdateFinalizationPeriod(ctx, msg)
	require.Error(t, err)

	// not exist bridge
	msg = types.NewMsgUpdateFinalizationPeriod(govAddr, 2, time.Second*10)
	_, err = ms.UpdateFinalizationPeriod(ctx, msg)
	require.Error(t, err)
}

func Test_MsgServer_SetFastBridgeConfig(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	config := types.BridgeConfig{
		Proposer:              addrsStr[0],
		Challenger:            addrsStr[1],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: addrsStr[0], ChainType: types.BatchInfo_INITIA},
	}

	_, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(addrsStr[0], config))
	require.NoError(t, err)

	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress("gov"))
	require.NoError(t, err)

	verifiers := make([]types.FastBridgeVerifier, len(pubKeys))
	for i := range pubKeys {
		verifier, _ := types.NewFastBridgeVerifier(addrs[i], pubKeys[i])
		verifiers[i] = verifier
	}
	fastBridgeConfig := types.FastBridgeConfig{
		Verifiers:      verifiers,
		Threshold:      2,
		MaxRate:        "0.2",
		RecoveryWindow: 86400,
		BaseFee:        sdk.NewCoin("uinit", math.NewInt(100)),
	}

	// success case - new config
	msg := types.NewMsgUpdateFastBridgeConfig(govAddr, 1, &fastBridgeConfig)
	_, err = ms.UpdateFastBridgeConfig(ctx, msg)
	require.NoError(t, err)

	_fastBridgeConfig, err := ms.GetFastBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.True(t, fastBridgeConfig.Equal(_fastBridgeConfig))

	// success case - clear config
	msg = types.NewMsgUpdateFastBridgeConfig(govAddr, 1, nil)
	_, err = ms.UpdateFastBridgeConfig(ctx, msg)
	require.NoError(t, err)

	_fastBridgeConfig, err = ms.GetFastBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Nil(t, _fastBridgeConfig)

	// fail case - non-gov signer
	msg = types.NewMsgUpdateFastBridgeConfig(addrsStr[0], 1, &fastBridgeConfig)
	_, err = ms.UpdateFastBridgeConfig(ctx, msg)
	require.ErrorIs(t, err, govtypes.ErrInvalidSigner)

	// fail case - threshold is 0
	invalidFastBridgeConfig := fastBridgeConfig
	invalidFastBridgeConfig.Threshold = 0
	msg = types.NewMsgUpdateFastBridgeConfig(govAddr, 1, &invalidFastBridgeConfig)
	_, err = ms.UpdateFastBridgeConfig(ctx, msg)
	require.ErrorContains(t, err, "threshold must be greater than 0")

	// fail case - threshold is larger than the number of verifers
	invalidFastBridgeConfig.Threshold = uint32(len(verifiers) + 1)
	msg = types.NewMsgUpdateFastBridgeConfig(govAddr, 1, &invalidFastBridgeConfig)
	_, err = ms.UpdateFastBridgeConfig(ctx, msg)
	require.ErrorContains(t, err, "threshold exceeds 5, got 6")

	// fail case - non-float max rate
	invalidFastBridgeConfig = fastBridgeConfig
	invalidFastBridgeConfig.MaxRate = "invalid"
	msg = types.NewMsgUpdateFastBridgeConfig(govAddr, 1, &invalidFastBridgeConfig)
	_, err = ms.UpdateFastBridgeConfig(ctx, msg)
	require.ErrorContains(t, err, "failed to set decimal string with base 10")

	// fail case - invalid max rate
	invalidFastBridgeConfig.MaxRate = "1.1"
	msg = types.NewMsgUpdateFastBridgeConfig(govAddr, 1, &invalidFastBridgeConfig)
	_, err = ms.UpdateFastBridgeConfig(ctx, msg)
	require.ErrorContains(t, err, "max rate must be within 0 and 1, got 1.100000")

	// fail case - recovery window is 0
	invalidFastBridgeConfig = fastBridgeConfig
	invalidFastBridgeConfig.RecoveryWindow = 0
	msg = types.NewMsgUpdateFastBridgeConfig(govAddr, 1, &invalidFastBridgeConfig)
	_, err = ms.UpdateFastBridgeConfig(ctx, msg)
	require.ErrorContains(t, err, "recovery window must be greater than 0")

	// fail case - invalid verifier address format
	invalidVerifiers := []types.FastBridgeVerifier{
		{
			Address: "invalid",
			Pubkey:  verifiers[0].Pubkey,
		},
		verifiers[1],
		verifiers[2],
	}
	invalidFastBridgeConfig = fastBridgeConfig
	invalidFastBridgeConfig.Verifiers = invalidVerifiers
	msg = types.NewMsgUpdateFastBridgeConfig(govAddr, 1, &invalidFastBridgeConfig)
	_, err = ms.UpdateFastBridgeConfig(ctx, msg)
	require.ErrorContains(t, err, "decoding bech32 failed")

	// fail case - invalid pubkey format
	invalidFastBridgeConfig.Verifiers[0].Address = verifiers[0].Address
	invalidPubkey, _ := codectypes.NewAnyWithValue(&types.FastBridgeVerifier{})
	invalidFastBridgeConfig.Verifiers[0].Pubkey = invalidPubkey
	msg = types.NewMsgUpdateFastBridgeConfig(govAddr, 1, &invalidFastBridgeConfig)
	_, err = ms.UpdateFastBridgeConfig(ctx, msg)
	require.ErrorContains(t, err, "expecting cryptotypes.PubKey")

	// fail case - verifier address and pubkey mismatch
	invalidVerifiers = []types.FastBridgeVerifier{
		{
			Address: verifiers[1].Address,
			Pubkey:  verifiers[0].Pubkey,
		},
		{
			Address: verifiers[0].Address,
			Pubkey:  verifiers[1].Pubkey,
		},
	}
	invalidFastBridgeConfig = fastBridgeConfig
	invalidFastBridgeConfig.Verifiers = invalidVerifiers
	msg = types.NewMsgUpdateFastBridgeConfig(govAddr, 1, &invalidFastBridgeConfig)
	_, err = ms.UpdateFastBridgeConfig(ctx, msg)
	require.ErrorContains(t, err, "mismatch pubkey address")
}
