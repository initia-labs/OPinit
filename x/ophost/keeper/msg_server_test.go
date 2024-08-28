package keeper_test

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	v1 "github.com/cometbft/cometbft/api/cometbft/crypto/v1"
	"github.com/cometbft/cometbft/crypto/tmhash"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	comettypes "github.com/cometbft/cometbft/types"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store/iavl"
	"cosmossdk.io/store/metrics"
	"cosmossdk.io/store/rootmulti"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"

	opchildtypes "github.com/initia-labs/OPinit/x/opchild/types"
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
		Challengers:           []string{addrsStr[0]},
		Proposer:              addrsStr[0],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: addrsStr[0], ChainType: types.BatchInfo_CHAIN_TYPE_INITIA},
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
		Challengers:           []string{addrsStr[0]},
		Proposer:              addrsStr[0],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: addrsStr[0], ChainType: types.BatchInfo_CHAIN_TYPE_INITIA},
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
		Challengers:           []string{addrsStr[1]},
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: addrsStr[0], ChainType: types.BatchInfo_CHAIN_TYPE_INITIA},
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
	_, err = ms.ProposeOutput(ctx, types.NewMsgProposeOutput(addrsStr[0], 1, 1, 100, []byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
	require.NoError(t, err)

	// invalid output index: nextoutputindex is 2 now
	_, err = ms.DeleteOutput(ctx, types.NewMsgDeleteOutput(addrsStr[1], 1, 2))
	require.Error(t, err)
}

func Test_InitiateTokenDeposit(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)
	config := types.BridgeConfig{
		Proposer:              addrsStr[0],
		Challengers:           []string{addrsStr[1]},
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: addrsStr[0], ChainType: types.BatchInfo_CHAIN_TYPE_INITIA},
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
		Proposer:              addrsStr[0],
		Challengers:           []string{addrsStr[1]},
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: addrsStr[0], ChainType: types.BatchInfo_CHAIN_TYPE_INITIA},
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
		Challengers:           []string{addrsStr[1]},
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: addrsStr[0], ChainType: types.BatchInfo_CHAIN_TYPE_INITIA},
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
		Proposer:              addrsStr[0],
		Challengers:           []string{addrsStr[1], addrsStr[2], addrsStr[3]},
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: addrsStr[0], ChainType: types.BatchInfo_CHAIN_TYPE_INITIA},
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
		Proposer:              addrsStr[0],
		Challengers:           []string{addrsStr[1]},
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo: types.BatchInfo{
			Submitter: addrsStr[1],
			ChainType: types.BatchInfo_CHAIN_TYPE_INITIA,
		},
	}

	_, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(addrsStr[0], config))
	require.NoError(t, err)

	// gov signer
	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress("gov"))
	require.NoError(t, err)
	msg := types.NewMsgUpdateBatchInfo(govAddr, 1, types.BatchInfo{
		Submitter: addrsStr[2],
		ChainType: types.BatchInfo_CHAIN_TYPE_CELESTIA,
	})
	_, err = ms.UpdateBatchInfo(ctx, msg)
	require.NoError(t, err)
	_config, err := ms.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, _config.BatchInfo.Submitter, addrsStr[2])
	require.Equal(t, types.BatchInfo_CHAIN_TYPE_CELESTIA, _config.BatchInfo.ChainType)
	require.Equal(t, input.BridgeHook.batchInfo, _config.BatchInfo)

	// current proposer signer
	msg = types.NewMsgUpdateBatchInfo(addrsStr[0], 1, types.BatchInfo{
		Submitter: addrsStr[3],
		ChainType: types.BatchInfo_CHAIN_TYPE_INITIA,
	})
	_, err = ms.UpdateBatchInfo(ctx, msg)
	require.NoError(t, err)
	_config, err = ms.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, _config.BatchInfo.Submitter, addrsStr[3])
	require.Equal(t, types.BatchInfo_CHAIN_TYPE_INITIA, _config.BatchInfo.ChainType)
	require.Equal(t, input.BridgeHook.batchInfo, _config.BatchInfo)

	// invalid signer
	invalidAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)
	msg = types.NewMsgUpdateBatchInfo(invalidAddr, 1, types.BatchInfo{
		Submitter: addrsStr[2],
		ChainType: types.BatchInfo_CHAIN_TYPE_CELESTIA,
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
		Proposer:              addrsStr[0],
		Challengers:           []string{addrsStr[1]},
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: addrsStr[0], ChainType: types.BatchInfo_CHAIN_TYPE_INITIA},
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

func Test_MsgServer_ForceWithdrawal(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	config := types.BridgeConfig{
		Proposer:              addrsStr[0],
		Challengers:           []string{addrsStr[1]},
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: addrsStr[0], ChainType: types.BatchInfo_CHAIN_TYPE_INITIA},
	}

	_, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(addrsStr[0], config))
	require.NoError(t, err)

	// fund amount
	amount := sdk.NewCoin("uinit", math.NewInt(1_000_000))
	input.Faucet.Fund(ctx, types.BridgeAddress(1), amount.Add(amount).Add(amount))

	sender := "osmo174knscjg688ddtxj8smyjz073r3w5mms8ugvx6"
	receiver := "cosmos174knscjg688ddtxj8smyjz073r3w5mms08musg"

	version := byte(1)

	leaf1 := types.GenerateWithdrawalHash(1, 1, sender, receiver, amount.Denom, amount.Amount.Uint64())
	leaf2 := types.GenerateWithdrawalHash(1, 2, sender, receiver, amount.Denom, amount.Amount.Uint64())
	leaf3 := types.GenerateWithdrawalHash(1, 3, sender, receiver, amount.Denom, amount.Amount.Uint64())

	node33 := types.GenerateNodeHash(leaf3[:], leaf3[:])
	node12 := types.GenerateNodeHash(leaf1[:], leaf2[:])

	storageRoot := types.GenerateNodeHash(node12[:], node33[:])
	appHash, commitmentProofs := makeAppHashWithCommitmentProof(t, []testInput{
		{recipient: receiver, amount: amount, l2Sequence: 1},
		{recipient: receiver, amount: amount, l2Sequence: 2},
		{recipient: receiver, amount: amount, l2Sequence: 3},
	})

	block := makeRandBlock(t, appHash)
	header := block.Header
	blockHash := block.Hash()
	appHashProof := opchildtypes.NewAppHashProof(&header)
	err = opchildtypes.VerifyAppHash(blockHash, header.AppHash, appHashProof)
	require.NoError(t, err)
	outputRoot := types.GenerateOutputRoot(version, storageRoot[:], blockHash)

	now := time.Now().UTC()
	ctx = ctx.WithBlockTime(now)
	_, err = ms.ProposeOutput(ctx, types.NewMsgProposeOutput(addrsStr[0], 1, 1, 100, outputRoot[:]))
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(now.Add(time.Second * 60))

	// force withdraw 1
	_, err = ms.ForceTokenWithdrwal(ctx, types.NewMsgForceTokenWithdrawal(
		1, 1, 1, sender, receiver, amount, *commitmentProofs[0], appHash, *appHashProof, version, storageRoot[:], blockHash,
	))
	require.NoError(t, err)

	receiverAddr, err := sdk.AccAddressFromBech32(receiver)
	require.NoError(t, err)
	require.Equal(t, amount, input.BankKeeper.GetBalance(ctx, receiverAddr, amount.Denom))

	// cannot finalize 1 again
	proofs := [][]byte{leaf2[:], node33[:]}
	_, err = ms.FinalizeTokenWithdrawal(ctx, types.NewMsgFinalizeTokenWithdrawal(1, 1, 1, proofs,
		sender,
		receiver,
		amount,
		[]byte{version}, storageRoot[:], blockHash,
	))
	require.Error(t, err)

	// can finalize withdrawal 2
	proofs = [][]byte{leaf1[:], node33[:]}
	_, err = ms.FinalizeTokenWithdrawal(ctx, types.NewMsgFinalizeTokenWithdrawal(1, 1, 2, proofs,
		sender,
		receiver,
		amount,
		[]byte{version}, storageRoot[:], blockHash,
	))
	require.NoError(t, err)
	require.Equal(t, amount.Add(amount), input.BankKeeper.GetBalance(ctx, receiverAddr, amount.Denom))

	// cannot force withdraw 2 again
	_, err = ms.ForceTokenWithdrwal(ctx, types.NewMsgForceTokenWithdrawal(
		1, 1, 2, sender, receiver, amount, *commitmentProofs[1], appHash, *appHashProof, version, storageRoot[:], blockHash,
	))
	require.Error(t, err)

	// can force withdrawal 3
	_, err = ms.ForceTokenWithdrwal(ctx, types.NewMsgForceTokenWithdrawal(
		1, 1, 3, sender, receiver, amount, *commitmentProofs[2], appHash, *appHashProof, version, storageRoot[:], blockHash,
	))
	require.NoError(t, err)
	require.Equal(t, amount.Add(amount).Add(amount), input.BankKeeper.GetBalance(ctx, receiverAddr, amount.Denom))
}

type testInput struct {
	recipient  string
	amount     sdk.Coin
	l2Sequence uint64
}

func makeAppHashWithCommitmentProof(t *testing.T, inputs []testInput) ([]byte, []*v1.ProofOps) {
	db := dbm.NewMemDB()
	store := rootmulti.NewStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	iavlStoreKey := storetypes.NewKVStoreKey(opchildtypes.StoreKey)

	store.MountStoreWithDB(iavlStoreKey, storetypes.StoreTypeIAVL, nil)
	require.NoError(t, store.LoadVersion(0))

	iavlStore := store.GetCommitStore(iavlStoreKey).(*iavl.Store)

	commitmentKeys := make([][]byte, len(inputs))
	for i, input := range inputs {
		commitment := opchildtypes.CommitWithdrawal(input.l2Sequence, input.recipient, input.amount)
		commitmentKeys[i] = opchildtypes.WithdrawalCommitmentKey(input.l2Sequence)

		iavlStore.Set(commitmentKeys[i], commitment)
	}

	cid := store.Commit()

	// Get Proof
	proofs := make([]*v1.ProofOps, len(commitmentKeys))
	for i, commitmentKey := range commitmentKeys {
		// same with curl https://rpc.initia.xyz/abci_query\?path\="\"store/opchild/key\""\&data=0xcommitmentkey\&prove=true
		res, err := store.Query(&storetypes.RequestQuery{
			Path:  fmt.Sprintf("/%s/key", opchildtypes.StoreKey), // required path to get key/value+proof
			Data:  commitmentKey,
			Prove: true,
		})
		require.NoError(t, err)
		require.NotNil(t, res.ProofOps)

		proofs[i] = opchildtypes.NewProtoFromProofOps(res.ProofOps)
	}

	return cid.GetHash(), proofs
}

func makeRandBlock(t *testing.T, appHash []byte) *comettypes.Block {
	txs := []comettypes.Tx{comettypes.Tx("foo"), comettypes.Tx("bar")}
	lastID := makeBlockIDRandom()
	h := int64(3)
	voteSet, valSet, vals := randVoteSet(h-1, 1, cmtproto.PrecommitType, 10, 1, false)
	extCommit, err := comettypes.MakeExtCommit(lastID, h-1, 1, voteSet, vals, time.Now().UTC(), false)
	require.NoError(t, err)

	ev, err := comettypes.NewMockDuplicateVoteEvidenceWithValidator(h, time.Now().UTC(), vals[0], "block-test-chain")
	require.NoError(t, err)
	evList := []comettypes.Evidence{ev}

	block := comettypes.MakeBlock(h, txs, extCommit.ToCommit(), evList)
	block.ValidatorsHash = valSet.Hash()
	block.AppHash = appHash

	return block
}

func makeBlockIDRandom() comettypes.BlockID {
	var (
		blockHash   = make([]byte, tmhash.Size)
		partSetHash = make([]byte, tmhash.Size)
	)
	rand.Read(blockHash)   //nolint: errcheck // ignore errcheck for read
	rand.Read(partSetHash) //nolint: errcheck // ignore errcheck for read
	return comettypes.BlockID{Hash: blockHash, PartSetHeader: comettypes.PartSetHeader{Total: 123, Hash: partSetHash}}
}

// NOTE: privValidators are in order
func randVoteSet(
	height int64,
	round int32,
	signedMsgType cmtproto.SignedMsgType,
	numValidators int,
	votingPower int64,
	extEnabled bool,
) (*comettypes.VoteSet, *comettypes.ValidatorSet, []comettypes.PrivValidator) {
	valSet, privValidators := comettypes.RandValidatorSet(numValidators, votingPower)
	if extEnabled {
		if signedMsgType != cmtproto.PrecommitType {
			return nil, nil, nil
		}
		return comettypes.NewExtendedVoteSet("test_chain_id", height, round, signedMsgType, valSet), valSet, privValidators
	}
	return comettypes.NewVoteSet("test_chain_id", height, round, signedMsgType, valSet), valSet, privValidators
}
