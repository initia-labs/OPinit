package keeper_test

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	"github.com/initia-labs/OPinit/x/ophost/keeper"
	"github.com/initia-labs/OPinit/x/ophost/testutil"
	"github.com/initia-labs/OPinit/x/ophost/types"
)

func Test_RecordBatch(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)
	_, err := ms.RecordBatch(ctx, types.NewMsgRecordBatch(testutil.AddrsStr[0], 1, []byte{1, 2, 3}))
	require.NoError(t, err)
}

func Test_CreateBridge(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	params := input.OPHostKeeper.GetParams(ctx)
	params.RegistrationFee = sdk.NewCoins(sdk.NewCoin("foo", math.NewInt(100)))
	require.NoError(t, input.OPHostKeeper.SetParams(ctx, params))

	input.Faucet.Fund(ctx, testutil.Addrs[0], sdk.NewCoin("foo", math.NewInt(1000)))

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)
	config := types.BridgeConfig{
		Challenger:            testutil.AddrsStr[0],
		Proposer:              testutil.AddrsStr[0],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: types.BatchInfo_INITIA},
	}
	res, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(testutil.AddrsStr[0], config))
	require.NoError(t, err)
	require.Equal(t, uint64(1), res.BridgeId)

	_config, err := input.OPHostKeeper.GetBridgeConfig(ctx, res.BridgeId)
	require.NoError(t, err)
	require.Equal(t, config, _config)

	// check community pool
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("foo", math.NewInt(100))), input.CommunityPoolKeeper.CommunityPool)
}

func Test_ProposeOutput(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)
	config := types.BridgeConfig{
		Challenger:            testutil.AddrsStr[0],
		Proposer:              testutil.AddrsStr[0],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: types.BatchInfo_INITIA},
	}
	createRes, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(testutil.AddrsStr[0], config))
	require.NoError(t, err)
	require.Equal(t, uint64(1), createRes.BridgeId)

	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// unauthorized
	_, err = ms.ProposeOutput(ctx, types.NewMsgProposeOutput(testutil.AddrsStr[1], 1, 1, 100, []byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
	require.Error(t, err)

	// valid
	_, err = ms.ProposeOutput(ctx, types.NewMsgProposeOutput(testutil.AddrsStr[0], 1, 1, 100, []byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
	require.NoError(t, err)

	output, err := input.OPHostKeeper.GetOutputProposal(ctx, 1, 1)
	require.NoError(t, err)
	require.Equal(t, types.Output{
		OutputRoot:    []byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		L1BlockNumber: uint64(ctx.BlockHeight()), //nolint:gosec
		L1BlockTime:   blockTime,
		L2BlockNumber: 100,
	}, output)
}

func Test_DeleteOutput(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)
	config := types.BridgeConfig{
		Proposer:              testutil.AddrsStr[0],
		Challenger:            testutil.AddrsStr[1],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: types.BatchInfo_INITIA},
	}
	createReq := types.NewMsgCreateBridge(testutil.AddrsStr[0], config)
	createRes, err := ms.CreateBridge(ctx, createReq)
	require.NoError(t, err)
	require.Equal(t, uint64(1), createRes.BridgeId)

	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	_, err = ms.ProposeOutput(ctx, types.NewMsgProposeOutput(testutil.AddrsStr[0], 1, 1, 100, []byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
	require.NoError(t, err)

	_, err = ms.ProposeOutput(ctx, types.NewMsgProposeOutput(testutil.AddrsStr[0], 1, 2, 200, []byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
	require.NoError(t, err)

	// unauthorized
	_, err = ms.DeleteOutput(ctx, types.NewMsgDeleteOutput(testutil.AddrsStr[2], 1, 1))
	require.Error(t, err)

	// valid by challenger
	_, err = ms.DeleteOutput(ctx, types.NewMsgDeleteOutput(testutil.AddrsStr[1], 1, 1))
	require.NoError(t, err)

	// should return error; deleted
	_, err = input.OPHostKeeper.GetOutputProposal(ctx, 1, 2)
	require.Error(t, err)

	_, err = input.OPHostKeeper.GetOutputProposal(ctx, 1, 1)
	require.Error(t, err)

	// should be able to resubmit the same output
	_, err = ms.ProposeOutput(ctx, types.NewMsgProposeOutput(testutil.AddrsStr[0], 1, 1, 100, []byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
	require.NoError(t, err)

	// invalid output index: nextoutputindex is 2 now
	_, err = ms.DeleteOutput(ctx, types.NewMsgDeleteOutput(testutil.AddrsStr[1], 1, 2))
	require.Error(t, err)

	// valid delete by gov
	_, err = ms.DeleteOutput(ctx, types.NewMsgDeleteOutput(input.OPHostKeeper.GetAuthority(), 1, 1))
	require.NoError(t, err)

	// should be able to resubmit the same output
	_, err = ms.ProposeOutput(ctx, types.NewMsgProposeOutput(testutil.AddrsStr[0], 1, 1, 100, []byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
	require.NoError(t, err)

	// valid delete by proposer
	_, err = ms.DeleteOutput(ctx, types.NewMsgDeleteOutput(testutil.AddrsStr[0], 1, 1))
	require.NoError(t, err)
}

func Test_InitiateTokenDeposit(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)
	config := types.BridgeConfig{
		Proposer:              testutil.AddrsStr[0],
		Challenger:            testutil.AddrsStr[1],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: types.BatchInfo_INITIA},
	}
	createRes, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(testutil.AddrsStr[0], config))
	require.NoError(t, err)
	require.Equal(t, uint64(1), createRes.BridgeId)

	amount := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100))
	input.Faucet.Fund(ctx, testutil.Addrs[1], amount)
	_, err = ms.InitiateTokenDeposit(
		ctx,
		types.NewMsgInitiateTokenDeposit(testutil.AddrsStr[1], 1, "l2_addr", amount, []byte("messages")),
	)
	require.NoError(t, err)
	require.True(t, input.BankKeeper.GetBalance(ctx, testutil.Addrs[1], sdk.DefaultBondDenom).IsZero())
	require.Equal(t, amount, input.BankKeeper.GetBalance(ctx, types.BridgeAddress(1), sdk.DefaultBondDenom))

	// not existing bridge
	_, err = ms.InitiateTokenDeposit(
		ctx,
		types.NewMsgInitiateTokenDeposit(testutil.AddrsStr[1], 2, "l2_addr", amount, []byte("messages")),
	)
	require.Error(t, err)

	// bridge disabled
	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress("gov"))
	require.NoError(t, err)
	_, err = ms.DisableBridge(ctx, types.NewMsgDisableBridge(govAddr, 1))
	require.NoError(t, err)
	_, err = ms.InitiateTokenDeposit(ctx, types.NewMsgInitiateTokenDeposit(testutil.AddrsStr[1], 1, "l2_addr", amount, []byte("messages")))
	require.ErrorIs(t, err, types.ErrBridgeDisabled)
}

func Test_FinalizeTokenWithdrawal(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)
	config := types.BridgeConfig{
		Proposer:              testutil.AddrsStr[0],
		Challenger:            testutil.AddrsStr[1],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: types.BatchInfo_INITIA},
	}
	_, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(testutil.AddrsStr[0], config))
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
	_, err = ms.ProposeOutput(ctx, types.NewMsgProposeOutput(testutil.AddrsStr[0], 1, 1, 100, outputRoot[:]))
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(now.Add(time.Second * 60))

	_, err = ms.FinalizeTokenWithdrawal(ctx, types.NewMsgFinalizeTokenWithdrawal(
		testutil.AddrsStr[3], // any address can execute this
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

func Test_FinalizeTokenWithdrawal_MigratedToken(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)
	config := types.BridgeConfig{
		Proposer:              testutil.AddrsStr[0],
		Challenger:            testutil.AddrsStr[1],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: types.BatchInfo_INITIA},
	}
	_, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(testutil.AddrsStr[0], config))
	require.NoError(t, err)

	// fund amount
	amount := sdk.NewCoin("uinit", math.NewInt(1_000_000))
	input.Faucet.Fund(ctx, types.BridgeAddress(1), amount.Add(amount))

	// register migration info
	migrationInfo := types.MigrationInfo{
		BridgeId:     1,
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
		L1Denom:      "uinit",
	}
	msg := types.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		1,
		migrationInfo,
	)
	_, err = ms.RegisterMigrationInfo(ctx, msg)
	require.NoError(t, err)

	// build withdrawal proof
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
	_, err = ms.ProposeOutput(ctx, types.NewMsgProposeOutput(testutil.AddrsStr[0], 1, 1, 100, outputRoot[:]))
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(now.Add(time.Second * 60))

	_, err = ms.FinalizeTokenWithdrawal(ctx, types.NewMsgFinalizeTokenWithdrawal(
		testutil.AddrsStr[3], // any address can execute this
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

	transferEscrowAddress := transfertypes.GetEscrowAddress("transfer", "channel-0")
	require.Equal(t, amount, input.BankKeeper.GetBalance(ctx, transferEscrowAddress, amount.Denom))
}

func decodeBase64(t *testing.T, str string) []byte {
	bz, err := base64.StdEncoding.DecodeString(str)
	require.NoError(t, err)

	return bz
}

func Test_UpdateProposal(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	config := types.BridgeConfig{
		Proposer:              testutil.AddrsStr[0],
		Challenger:            testutil.AddrsStr[1],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: types.BatchInfo_INITIA},
	}

	_, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(testutil.AddrsStr[0], config))
	require.NoError(t, err)

	// gov signer
	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress("gov"))
	require.NoError(t, err)
	msg := types.NewMsgUpdateProposer(govAddr, 1, testutil.AddrsStr[1])
	_, err = ms.UpdateProposer(ctx, msg)
	require.NoError(t, err)
	_config, err := ms.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, testutil.AddrsStr[1], _config.Proposer)
	require.Equal(t, testutil.AddrsStr[1], input.BridgeHook.Proposer)

	// current proposer signer
	msg = types.NewMsgUpdateProposer(testutil.AddrsStr[1], 1, testutil.AddrsStr[2])
	_, err = ms.UpdateProposer(ctx, msg)
	require.NoError(t, err)
	_config, err = ms.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, testutil.Addrs[2].String(), _config.Proposer)
	require.Equal(t, testutil.Addrs[2].String(), input.BridgeHook.Proposer)

	// invalid signer
	invalidAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)
	msg = types.NewMsgUpdateProposer(invalidAddr, 1, testutil.AddrsStr[1])
	require.NoError(t, err)

	_, err = ms.UpdateProposer(
		ctx,
		msg,
	)
	require.Error(t, err)
}

func Test_UpdateChallenger(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	config := types.BridgeConfig{
		Proposer:              testutil.AddrsStr[0],
		Challenger:            testutil.AddrsStr[1],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: types.BatchInfo_INITIA},
	}

	_, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(testutil.AddrsStr[0], config))
	require.NoError(t, err)

	// gov signer
	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress("gov"))
	require.NoError(t, err)
	msg := types.NewMsgUpdateChallenger(govAddr, 1, testutil.AddrsStr[2])
	_, err = ms.UpdateChallenger(ctx, msg)
	require.NoError(t, err)
	_config, err := ms.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, testutil.AddrsStr[2], _config.Challenger)
	require.Equal(t, input.BridgeHook.Challenger, _config.Challenger)

	// current challenger

	// invalid signer
	invalidAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)
	msg = types.NewMsgUpdateChallenger(invalidAddr, 1, testutil.AddrsStr[1])
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
	ctx, input := testutil.CreateTestInput(t, false)
	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	config := types.BridgeConfig{
		Proposer:              testutil.AddrsStr[0],
		Challenger:            testutil.AddrsStr[1],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo: types.BatchInfo{
			Submitter: testutil.AddrsStr[1],
			ChainType: types.BatchInfo_INITIA,
		},
	}

	_, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(testutil.AddrsStr[0], config))
	require.NoError(t, err)

	// gov signer
	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress("gov"))
	require.NoError(t, err)
	msg := types.NewMsgUpdateBatchInfo(govAddr, 1, types.BatchInfo{
		Submitter: testutil.AddrsStr[2],
		ChainType: types.BatchInfo_CELESTIA,
	})
	_, err = ms.UpdateBatchInfo(ctx, msg)
	require.NoError(t, err)
	_config, err := ms.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, _config.BatchInfo.Submitter, testutil.AddrsStr[2])
	require.Equal(t, types.BatchInfo_CELESTIA, _config.BatchInfo.ChainType)
	require.Equal(t, input.BridgeHook.BatchInfo, _config.BatchInfo)

	// current proposer signer
	msg = types.NewMsgUpdateBatchInfo(testutil.AddrsStr[0], 1, types.BatchInfo{
		Submitter: testutil.AddrsStr[3],
		ChainType: types.BatchInfo_INITIA,
	})
	_, err = ms.UpdateBatchInfo(ctx, msg)
	require.NoError(t, err)
	_config, err = ms.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, _config.BatchInfo.Submitter, testutil.AddrsStr[3])
	require.Equal(t, types.BatchInfo_INITIA, _config.BatchInfo.ChainType)
	require.Equal(t, input.BridgeHook.BatchInfo, _config.BatchInfo)

	// invalid signer
	invalidAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)
	msg = types.NewMsgUpdateBatchInfo(invalidAddr, 1, types.BatchInfo{
		Submitter: testutil.AddrsStr[2],
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
	ctx, input := testutil.CreateTestInput(t, false)
	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	config := types.BridgeConfig{
		Proposer:              testutil.AddrsStr[0],
		Challenger:            testutil.AddrsStr[1],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: types.BatchInfo_INITIA},
		OracleEnabled:         true,
	}

	_, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(testutil.AddrsStr[0], config))
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
	msg = types.NewMsgUpdateOracleConfig(testutil.AddrsStr[0], 1, true)
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
	ctx, input := testutil.CreateTestInput(t, false)
	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	config := types.BridgeConfig{
		Proposer:              testutil.AddrsStr[0],
		Challenger:            testutil.AddrsStr[1],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: types.BatchInfo_INITIA},
	}

	_, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(testutil.AddrsStr[0], config))
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
	require.Equal(t, []byte{4, 5, 6}, input.BridgeHook.Metadata)

	// current proposer
	msg = types.NewMsgUpdateMetadata(testutil.AddrsStr[0], 1, []byte{7, 8, 9})
	_, err = ms.UpdateMetadata(ctx, msg)
	require.NoError(t, err)
	_config, err = ms.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, []byte{7, 8, 9}, _config.Metadata)
	require.Equal(t, []byte{7, 8, 9}, input.BridgeHook.Metadata)

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

func Test_MsgServer_DisableBridge(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	config := types.BridgeConfig{
		Proposer:              testutil.AddrsStr[0],
		Challenger:            testutil.AddrsStr[1],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: types.BatchInfo_INITIA},
		OracleEnabled:         true,
	}

	_, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(testutil.AddrsStr[0], config))
	require.NoError(t, err)

	// gov signer
	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress("gov"))
	require.NoError(t, err)
	msg := types.NewMsgDisableBridge(govAddr, 1)
	_, err = ms.DisableBridge(ctx, msg)
	require.NoError(t, err)
	_config, err := ms.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, true, _config.BridgeDisabled)
	require.Equal(t, ctx.BlockTime(), _config.BridgeDisabledAt)

	// already disabled
	msg = types.NewMsgDisableBridge(govAddr, 1)
	_, err = ms.DisableBridge(ctx, msg)
	require.Error(t, err)

	// current proposer signer
	msg = types.NewMsgDisableBridge(testutil.AddrsStr[0], 1)
	_, err = ms.DisableBridge(ctx, msg)
	require.Error(t, err)
}

func Test_MsgServer_UpdateParams(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
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
	ctx, input := testutil.CreateTestInput(t, false)
	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	config := types.BridgeConfig{
		Proposer:              testutil.AddrsStr[0],
		Challenger:            testutil.AddrsStr[1],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: types.BatchInfo_INITIA},
	}

	_, err := ms.CreateBridge(ctx, types.NewMsgCreateBridge(testutil.AddrsStr[0], config))
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

func Test_MsgServer_RegisterAttestorSet(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(govtypes.ModuleName))
	require.NoError(t, err)

	bridgeId := uint64(1)
	bridgeConfig := types.BridgeConfig{
		Proposer:              testutil.AddrsStr[0],
		Challenger:            testutil.AddrsStr[1],
		SubmissionInterval:    100,
		FinalizationPeriod:    1000,
		SubmissionStartHeight: 1,
		ChannelId:             "channel-2",
		Metadata:              []byte("{\"perm_channels\":[{\"port_id\":\"opinit\",\"channel_id\":\"channel-2\"},{\"port_id\":\"nft-transfer\",\"channel_id\":\"channel-1\"},{\"port_id\":\"transfer\",\"channel_id\":\"channel-0\"}]}"),
		BatchInfo: types.BatchInfo{
			Submitter: testutil.AddrsStr[2],
			ChainType: types.BatchInfo_INITIA,
		},
		AttestorSet: []types.Attestor{},
	}
	err = input.OPHostKeeper.SetBridgeConfig(ctx, bridgeId, bridgeConfig)
	require.NoError(t, err)

	// create attestors
	privKey1 := ed25519.GenPrivKey()
	pubKey1 := privKey1.PubKey()
	pkAny1, err := codectypes.NewAnyWithValue(pubKey1)
	require.NoError(t, err)

	privKey2 := ed25519.GenPrivKey()
	pubKey2 := privKey2.PubKey()
	pkAny2, err := codectypes.NewAnyWithValue(pubKey2)
	require.NoError(t, err)

	attestorSet := []types.Attestor{
		{
			OperatorAddress: testutil.ValAddrsStr[0],
			ConsensusPubkey: pkAny1,
			Moniker:         "attestor1",
		},
		{
			OperatorAddress: testutil.ValAddrsStr[1],
			ConsensusPubkey: pkAny2,
			Moniker:         "attestor2",
		},
	}

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	// successful registration
	msg := types.NewMsgRegisterAttestorSet(govAddr, bridgeId, attestorSet)
	_, err = ms.RegisterAttestorSet(ctx, msg)
	require.NoError(t, err)

	config, err := input.OPHostKeeper.GetBridgeConfig(ctx, bridgeId)
	require.NoError(t, err)
	require.Len(t, config.AttestorSet, 2)
	require.Equal(t, testutil.ValAddrsStr[0], config.AttestorSet[0].OperatorAddress)
	require.Equal(t, testutil.ValAddrsStr[1], config.AttestorSet[1].OperatorAddress)

	events := sdk.UnwrapSDKContext(ctx).EventManager().Events()
	found := false
	for _, event := range events {
		if event.Type == types.EventTypeRegisterAttestorSet {
			found = true
			break
		}
	}
	require.True(t, found, "expected register attestor set event")
}

func Test_MsgServer_RegisterAttestorSet_InvalidSigner(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	bridgeId := uint64(1)
	bridgeConfig := types.BridgeConfig{
		Proposer:              testutil.AddrsStr[0],
		Challenger:            testutil.AddrsStr[1],
		SubmissionInterval:    100,
		FinalizationPeriod:    1000,
		SubmissionStartHeight: 1,
		Metadata:              []byte("test-metadata"),
		BatchInfo: types.BatchInfo{
			Submitter: testutil.AddrsStr[2],
			ChainType: types.BatchInfo_INITIA,
		},
		AttestorSet: []types.Attestor{},
	}
	err := input.OPHostKeeper.SetBridgeConfig(ctx, bridgeId, bridgeConfig)
	require.NoError(t, err)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	// invalid authority (not gov)
	msg := types.NewMsgRegisterAttestorSet(testutil.AddrsStr[0], bridgeId, []types.Attestor{})
	_, err = ms.RegisterAttestorSet(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid authority")
}

func Test_MsgServer_RegisterAttestorSet_BridgeNotFound(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(govtypes.ModuleName))
	require.NoError(t, err)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	// non-existent bridge
	msg := types.NewMsgRegisterAttestorSet(govAddr, 999, []types.Attestor{})
	_, err = ms.RegisterAttestorSet(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func Test_MsgServer_UpdateChannelId_Success(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(govtypes.ModuleName))
	require.NoError(t, err)

	bridgeId := uint64(1)
	bridgeConfig := types.BridgeConfig{
		Proposer:              testutil.AddrsStr[0],
		Challenger:            testutil.AddrsStr[1],
		SubmissionInterval:    100,
		FinalizationPeriod:    1000,
		SubmissionStartHeight: 1,
		ChannelId:             "channel-0",
		BatchInfo: types.BatchInfo{
			Submitter: testutil.AddrsStr[2],
			ChainType: types.BatchInfo_INITIA,
		},
		AttestorSet: []types.Attestor{},
	}
	err = input.OPHostKeeper.SetBridgeConfig(ctx, bridgeId, bridgeConfig)
	require.NoError(t, err)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	// update channel id as gov
	msg := types.NewMsgUpdateChannelId(govAddr, bridgeId, "channel-5")
	_, err = ms.UpdateChannelId(ctx, msg)
	require.NoError(t, err)

	// verify channel id was updated
	config, err := input.OPHostKeeper.GetBridgeConfig(ctx, bridgeId)
	require.NoError(t, err)
	require.Equal(t, "channel-5", config.ChannelId)

	events := sdk.UnwrapSDKContext(ctx).EventManager().Events()
	found := false
	for _, event := range events {
		if event.Type == types.EventTypeUpdateChannelId {
			found = true
			break
		}
	}
	require.True(t, found, "expected update channel id event")
}

func Test_MsgServer_UpdateChannelId_AsProposer(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	// create bridge
	bridgeId := uint64(1)
	bridgeConfig := types.BridgeConfig{
		Proposer:              testutil.AddrsStr[0],
		Challenger:            testutil.AddrsStr[1],
		SubmissionInterval:    100,
		FinalizationPeriod:    1000,
		SubmissionStartHeight: 1,
		ChannelId:             "channel-0",
		BatchInfo: types.BatchInfo{
			Submitter: testutil.AddrsStr[2],
			ChainType: types.BatchInfo_INITIA,
		},
		AttestorSet: []types.Attestor{},
	}
	err := input.OPHostKeeper.SetBridgeConfig(ctx, bridgeId, bridgeConfig)
	require.NoError(t, err)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	// proposer cannot update channel id
	msg := types.NewMsgUpdateChannelId(testutil.AddrsStr[0], bridgeId, "channel-10")
	_, err = ms.UpdateChannelId(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid authority")
}

func Test_MsgServer_UpdateChannelId_InvalidAuthority(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	// create bridge
	bridgeId := uint64(1)
	bridgeConfig := types.BridgeConfig{
		Proposer:              testutil.AddrsStr[0],
		Challenger:            testutil.AddrsStr[1],
		SubmissionInterval:    100,
		FinalizationPeriod:    1000,
		SubmissionStartHeight: 1,
		ChannelId:             "channel-0",
		BatchInfo: types.BatchInfo{
			Submitter: testutil.AddrsStr[2],
			ChainType: types.BatchInfo_INITIA,
		},
		AttestorSet: []types.Attestor{},
	}
	err := input.OPHostKeeper.SetBridgeConfig(ctx, bridgeId, bridgeConfig)
	require.NoError(t, err)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	// unauthorized user cannot update channel id
	msg := types.NewMsgUpdateChannelId(testutil.AddrsStr[2], bridgeId, "channel-5")
	_, err = ms.UpdateChannelId(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid authority")
}

func Test_MsgServer_UpdateChannelId_EmptyChannelId(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(govtypes.ModuleName))
	require.NoError(t, err)

	// create bridge
	bridgeId := uint64(1)
	bridgeConfig := types.BridgeConfig{
		Proposer:              testutil.AddrsStr[0],
		Challenger:            testutil.AddrsStr[1],
		SubmissionInterval:    100,
		FinalizationPeriod:    1000,
		SubmissionStartHeight: 1,
		ChannelId:             "channel-0",
		BatchInfo: types.BatchInfo{
			Submitter: testutil.AddrsStr[2],
			ChainType: types.BatchInfo_INITIA,
		},
		AttestorSet: []types.Attestor{},
	}
	err = input.OPHostKeeper.SetBridgeConfig(ctx, bridgeId, bridgeConfig)
	require.NoError(t, err)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	// empty channel id should fail validation
	msg := types.NewMsgUpdateChannelId(govAddr, bridgeId, "")
	_, err = ms.UpdateChannelId(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid channel id")
}

func Test_MsgServer_RegisterAttestorSet_EmptyChannelId(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(govtypes.ModuleName))
	require.NoError(t, err)

	// create bridge without channel_id
	bridgeId := uint64(1)
	bridgeConfig := types.BridgeConfig{
		Proposer:              testutil.AddrsStr[0],
		Challenger:            testutil.AddrsStr[1],
		SubmissionInterval:    100,
		FinalizationPeriod:    1000,
		SubmissionStartHeight: 1,
		ChannelId:             "",
		BatchInfo: types.BatchInfo{
			Submitter: testutil.AddrsStr[2],
			ChainType: types.BatchInfo_INITIA,
		},
		AttestorSet: []types.Attestor{},
	}
	err = input.OPHostKeeper.SetBridgeConfig(ctx, bridgeId, bridgeConfig)
	require.NoError(t, err)

	// create attestor
	privKey1 := ed25519.GenPrivKey()
	pubKey1 := privKey1.PubKey()
	pkAny1, err := codectypes.NewAnyWithValue(pubKey1)
	require.NoError(t, err)

	attestorSet := []types.Attestor{
		{
			OperatorAddress: testutil.ValAddrsStr[0],
			ConsensusPubkey: pkAny1,
			Moniker:         "attestor1",
		},
	}

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	// registration should fail due to empty channel id
	msg := types.NewMsgRegisterAttestorSet(govAddr, bridgeId, attestorSet)
	_, err = ms.RegisterAttestorSet(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "channel_id must be set")
}

func Test_MsgServer_RegisterAttestorSet_WithChannelId(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(govtypes.ModuleName))
	require.NoError(t, err)

	// create bridge
	bridgeId := uint64(1)
	bridgeConfig := types.BridgeConfig{
		Proposer:              testutil.AddrsStr[0],
		Challenger:            testutil.AddrsStr[1],
		SubmissionInterval:    100,
		FinalizationPeriod:    1000,
		SubmissionStartHeight: 1,
		ChannelId:             "channel-0",
		BatchInfo: types.BatchInfo{
			Submitter: testutil.AddrsStr[2],
			ChainType: types.BatchInfo_INITIA,
		},
		AttestorSet: []types.Attestor{},
	}
	err = input.OPHostKeeper.SetBridgeConfig(ctx, bridgeId, bridgeConfig)
	require.NoError(t, err)

	// create attestor
	privKey1 := ed25519.GenPrivKey()
	pubKey1 := privKey1.PubKey()
	pkAny1, err := codectypes.NewAnyWithValue(pubKey1)
	require.NoError(t, err)

	attestorSet := []types.Attestor{
		{
			OperatorAddress: testutil.ValAddrsStr[0],
			ConsensusPubkey: pkAny1,
			Moniker:         "attestor1",
		},
	}

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	// registration should succeed
	msg := types.NewMsgRegisterAttestorSet(govAddr, bridgeId, attestorSet)
	_, err = ms.RegisterAttestorSet(ctx, msg)
	require.NoError(t, err)

	// verify attestor set was updated
	config, err := input.OPHostKeeper.GetBridgeConfig(ctx, bridgeId)
	require.NoError(t, err)
	require.Len(t, config.AttestorSet, 1)
	require.Equal(t, testutil.ValAddrsStr[0], config.AttestorSet[0].OperatorAddress)
}

func Test_MsgServer_AddAttestor_EmptyChannelId(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(govtypes.ModuleName))
	require.NoError(t, err)

	// create bridge without channel_id
	bridgeId := uint64(1)
	bridgeConfig := types.BridgeConfig{
		Proposer:              testutil.AddrsStr[0],
		Challenger:            testutil.AddrsStr[1],
		SubmissionInterval:    100,
		FinalizationPeriod:    1000,
		SubmissionStartHeight: 1,
		ChannelId:             "",
		BatchInfo: types.BatchInfo{
			Submitter: testutil.AddrsStr[2],
			ChainType: types.BatchInfo_INITIA,
		},
		AttestorSet: []types.Attestor{},
	}
	err = input.OPHostKeeper.SetBridgeConfig(ctx, bridgeId, bridgeConfig)
	require.NoError(t, err)

	// create attestor
	privKey1 := ed25519.GenPrivKey()
	pubKey1 := privKey1.PubKey()
	pkAny1, err := codectypes.NewAnyWithValue(pubKey1)
	require.NoError(t, err)

	attestor := types.Attestor{
		OperatorAddress: testutil.ValAddrsStr[0],
		ConsensusPubkey: pkAny1,
		Moniker:         "attestor1",
	}

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	// add attestor should fail due to empty channel id
	msg := types.NewMsgAddAttestor(govAddr, bridgeId, attestor)
	_, err = ms.AddAttestor(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "channel_id must be set")
}

func Test_MsgServer_AddAttestor_WithChannelId(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(govtypes.ModuleName))
	require.NoError(t, err)

	// create bridge
	bridgeId := uint64(1)
	bridgeConfig := types.BridgeConfig{
		Proposer:              testutil.AddrsStr[0],
		Challenger:            testutil.AddrsStr[1],
		SubmissionInterval:    100,
		FinalizationPeriod:    1000,
		SubmissionStartHeight: 1,
		ChannelId:             "channel-0",
		BatchInfo: types.BatchInfo{
			Submitter: testutil.AddrsStr[2],
			ChainType: types.BatchInfo_INITIA,
		},
		AttestorSet: []types.Attestor{},
	}
	err = input.OPHostKeeper.SetBridgeConfig(ctx, bridgeId, bridgeConfig)
	require.NoError(t, err)

	// create attestor
	privKey1 := ed25519.GenPrivKey()
	pubKey1 := privKey1.PubKey()
	pkAny1, err := codectypes.NewAnyWithValue(pubKey1)
	require.NoError(t, err)

	attestor := types.Attestor{
		OperatorAddress: testutil.ValAddrsStr[0],
		ConsensusPubkey: pkAny1,
		Moniker:         "attestor1",
	}

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	// add attestor should succeed
	msg := types.NewMsgAddAttestor(govAddr, bridgeId, attestor)
	_, err = ms.AddAttestor(ctx, msg)
	require.NoError(t, err)

	// verify attestor was added
	config, err := input.OPHostKeeper.GetBridgeConfig(ctx, bridgeId)
	require.NoError(t, err)
	require.Len(t, config.AttestorSet, 1)
	require.Equal(t, testutil.ValAddrsStr[0], config.AttestorSet[0].OperatorAddress)
}

func Test_MsgServer_RemoveAttestor_EmptyChannelId(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(govtypes.ModuleName))
	require.NoError(t, err)

	// create attestor
	privKey1 := ed25519.GenPrivKey()
	pubKey1 := privKey1.PubKey()
	pkAny1, err := codectypes.NewAnyWithValue(pubKey1)
	require.NoError(t, err)

	attestor := types.Attestor{
		OperatorAddress: testutil.ValAddrsStr[0],
		ConsensusPubkey: pkAny1,
		Moniker:         "attestor1",
	}

	// create bridge without channel_id but with existing attestor
	bridgeId := uint64(1)
	bridgeConfig := types.BridgeConfig{
		Proposer:              testutil.AddrsStr[0],
		Challenger:            testutil.AddrsStr[1],
		SubmissionInterval:    100,
		FinalizationPeriod:    1000,
		SubmissionStartHeight: 1,
		ChannelId:             "",
		BatchInfo: types.BatchInfo{
			Submitter: testutil.AddrsStr[2],
			ChainType: types.BatchInfo_INITIA,
		},
		AttestorSet: []types.Attestor{attestor},
	}
	err = input.OPHostKeeper.SetBridgeConfig(ctx, bridgeId, bridgeConfig)
	require.NoError(t, err)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	// remove attestor should fail due to empty channel id
	msg := types.NewMsgRemoveAttestor(govAddr, bridgeId, testutil.ValAddrsStr[0])
	_, err = ms.RemoveAttestor(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "channel_id must be set")
}

func Test_MsgServer_RemoveAttestor_WithChannelId(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	govAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(govtypes.ModuleName))
	require.NoError(t, err)

	// create attestor
	privKey1 := ed25519.GenPrivKey()
	pubKey1 := privKey1.PubKey()
	pkAny1, err := codectypes.NewAnyWithValue(pubKey1)
	require.NoError(t, err)

	attestor := types.Attestor{
		OperatorAddress: testutil.ValAddrsStr[0],
		ConsensusPubkey: pkAny1,
		Moniker:         "attestor1",
	}

	// create bridge with both channel_id and attestor
	bridgeId := uint64(1)
	bridgeConfig := types.BridgeConfig{
		Proposer:              testutil.AddrsStr[0],
		Challenger:            testutil.AddrsStr[1],
		SubmissionInterval:    100,
		FinalizationPeriod:    1000,
		SubmissionStartHeight: 1,
		ChannelId:             "channel-0",
		BatchInfo: types.BatchInfo{
			Submitter: testutil.AddrsStr[2],
			ChainType: types.BatchInfo_INITIA,
		},
		AttestorSet: []types.Attestor{attestor},
	}
	err = input.OPHostKeeper.SetBridgeConfig(ctx, bridgeId, bridgeConfig)
	require.NoError(t, err)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	// remove attestor should succeed
	msg := types.NewMsgRemoveAttestor(govAddr, bridgeId, testutil.ValAddrsStr[0])
	_, err = ms.RemoveAttestor(ctx, msg)
	require.NoError(t, err)

	// verify attestor was removed
	config, err := input.OPHostKeeper.GetBridgeConfig(ctx, bridgeId)
	require.NoError(t, err)
	require.Len(t, config.AttestorSet, 0)
}
