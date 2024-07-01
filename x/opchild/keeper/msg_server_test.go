package keeper_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"testing"
	"time"

	"cosmossdk.io/math"
	testutilsims "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	protoio "github.com/cosmos/gogoproto/io"
	"github.com/cosmos/gogoproto/proto"

	cometabci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/skip-mev/slinky/abci/strategies/currencypair"
	vetypes "github.com/skip-mev/slinky/abci/ve/types"
	slinkytypes "github.com/skip-mev/slinky/pkg/types"
	oracletypes "github.com/skip-mev/slinky/x/oracle/types"

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

	ms := keeper.NewMsgServerImpl(input.OPChildKeeper)

	params, err := ms.GetParams(ctx)
	require.NoError(t, err)

	// admin to 0
	params.Admin = addrsStr[0]
	require.NoError(t, ms.SetParams(ctx, params))

	valPubKeys := testutilsims.CreateTestPubKeys(2)

	// register validator
	val, err := types.NewValidator(valAddrs[0], valPubKeys[0], "val1")
	require.NoError(t, err)

	input.OPChildKeeper.SetValidator(ctx, val)

	// apply validator updates
	input.OPChildKeeper.BlockValidatorUpdates(ctx)

	moduleAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)

	addMsg, err := types.NewMsgAddValidator("val2", moduleAddr, valAddrsStr[1], valPubKeys[1])
	require.NoError(t, err)

	removeMsg, err := types.NewMsgRemoveValidator(moduleAddr, valAddrsStr[0])
	require.NoError(t, err)

	// should failed with unauthorized
	msg, err := types.NewMsgExecuteMessages(addrsStr[1], []sdk.Msg{addMsg, removeMsg})
	require.NoError(t, err)

	_, err = ms.ExecuteMessages(ctx, msg)
	require.Error(t, err)

	// success
	msg, err = types.NewMsgExecuteMessages(addrsStr[0], []sdk.Msg{addMsg, removeMsg})
	require.NoError(t, err)

	_, err = ms.ExecuteMessages(ctx, msg)
	require.NoError(t, err)

	// apply validator updates
	input.OPChildKeeper.BlockValidatorUpdates(ctx)

	vals, err := input.OPChildKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	require.Len(t, vals, 1)
	require.Equal(t, vals[0].Moniker, "val2")

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

func Test_MsgServer_AddValidator(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	ms := keeper.NewMsgServerImpl(input.OPChildKeeper)
	valPubKeys := testutilsims.CreateTestPubKeys(2)

	moduleAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)

	msg, err := types.NewMsgAddValidator("val1", moduleAddr, valAddrsStr[0], valPubKeys[0])
	require.NoError(t, err)

	_, err = ms.AddValidator(ctx, msg)
	require.NoError(t, err)

	// invalid signer
	msg, err = types.NewMsgAddValidator("val1", addrsStr[0], valAddrsStr[0], valPubKeys[0])
	require.NoError(t, err)

	_, err = ms.AddValidator(ctx, msg)
	require.Error(t, err)

	// duplicate add validator
	msg, err = types.NewMsgAddValidator("val1", moduleAddr, valAddrsStr[0], valPubKeys[1])
	require.NoError(t, err)

	_, err = ms.AddValidator(ctx, msg)
	require.Error(t, err)

	// duplicate cons pubkey
	msg, err = types.NewMsgAddValidator("val1", moduleAddr, valAddrsStr[1], valPubKeys[0])
	require.NoError(t, err)

	_, err = ms.AddValidator(ctx, msg)
	require.Error(t, err)

	params, err := ms.GetParams(ctx)
	require.NoError(t, err)
	params.MaxValidators = 1
	ms.SetParams(ctx, params)

	msg, err = types.NewMsgAddValidator("val2", moduleAddr, valAddrsStr[1], valPubKeys[1])
	require.NoError(t, err)

	// max validators reached
	_, err = ms.AddValidator(ctx, msg)
	require.Error(t, err)

	params, err = ms.GetParams(ctx)
	require.NoError(t, err)
	params.MaxValidators = 2
	ms.SetParams(ctx, params)

	_, err = ms.AddValidator(ctx, msg)
	require.NoError(t, err)
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
	msg, err := types.NewMsgRemoveValidator(addrsStr[0], valAddrsStr[0])
	require.NoError(t, err)

	_, err = ms.RemoveValidator(
		ctx,
		msg,
	)
	require.Error(t, err)

	moduleAddr, err := input.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)

	// remove not existing validator
	msg, err = types.NewMsgRemoveValidator(moduleAddr, valAddrsStr[1])
	require.NoError(t, err)

	_, err = ms.RemoveValidator(
		ctx,
		msg,
	)
	require.Error(t, err)

	// valid remove validator
	msg, err = types.NewMsgRemoveValidator(moduleAddr, valAddrsStr[0])
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
	ms := keeper.NewMsgServerImpl(input.OPChildKeeper)

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
	ms := keeper.NewMsgServerImpl(input.OPChildKeeper)

	info := types.BridgeInfo{
		BridgeId:   1,
		BridgeAddr: addrsStr[1],
		L1ChainId:  "test-chain-id",
		L1ClientId: "test-client-id",
		BridgeConfig: ophosttypes.BridgeConfig{
			Challengers: []string{addrsStr[2]},
			Proposer:    addrsStr[3],
			BatchInfo: ophosttypes.BatchInfo{
				Submitter: addrsStr[4],
				Chain:     "l1",
			},
			SubmissionInterval:  time.Minute,
			FinalizationPeriod:  time.Hour,
			SubmissionStartTime: time.Now().UTC(),
			Metadata:            []byte("metadata"),
		},
	}

	_, err := ms.SetBridgeInfo(ctx, types.NewMsgSetBridgeInfo(addrsStr[0], info))
	require.NoError(t, err)

	baseDenom := "test_token"
	denom := ophosttypes.L2Denom(1, baseDenom)

	_, err = ms.FinalizeTokenDeposit(ctx, types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], addrsStr[1], sdk.NewCoin(denom, math.NewInt(100)), 1, 1, "test_token", nil))
	require.NoError(t, err)

	coins := sdk.NewCoins(sdk.NewCoin("foo", math.NewInt(1_000_000_000)), sdk.NewCoin(denom, math.NewInt(1_000_000_000)))
	// fund asset
	account := input.Faucet.NewFundedAccount(ctx, coins...)
	accountAddr, err := input.AccountKeeper.AddressCodec().BytesToString(account)
	require.NoError(t, err)

	// not token from l1
	msg := types.NewMsgInitiateTokenWithdrawal(accountAddr, addrsStr[1], sdk.NewCoin("foo", math.NewInt(100)))
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
	ms := keeper.NewMsgServerImpl(input.OPChildKeeper)

	info := types.BridgeInfo{
		BridgeId:   1,
		BridgeAddr: addrsStr[1],
		L1ChainId:  "test-chain-id",
		L1ClientId: "test-client-id",
		BridgeConfig: ophosttypes.BridgeConfig{
			Challengers: []string{addrsStr[2]},
			Proposer:    addrsStr[3],
			BatchInfo: ophosttypes.BatchInfo{
				Submitter: addrsStr[4],
				Chain:     "l1",
			},
			SubmissionInterval:  time.Minute,
			FinalizationPeriod:  time.Hour,
			SubmissionStartTime: time.Now().UTC(),
			Metadata:            []byte("metadata"),
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
	ms := keeper.NewMsgServerImpl(input.OPChildKeeper)

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

	afterToBalance := input.BankKeeper.GetBalance(ctx, addrs[1], denom)
	require.Equal(t, math.NewInt(100), afterToBalance.Amount)

	afterModuleBalance := input.BankKeeper.GetBalance(ctx, opchildModuleAddress, denom)
	require.Equal(t, math.ZeroInt(), afterModuleBalance.Amount)
}

func Test_MsgServer_Deposit_NoHook(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	ms := keeper.NewMsgServerImpl(input.OPChildKeeper)

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
	msg := types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], addrsStr[1], sdk.NewCoin(denom, math.NewInt(100)), 1, 1, "test_token", hookMsgBytes)
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
	msg := types.NewMsgFinalizeTokenDeposit(addrsStr[0], addrsStr[1], addrsStr[1], sdk.NewCoin(denom, math.NewInt(100)), 1, 1, "test_token", []byte("invalid_message"))
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

func Test_MsgServer_UpdateOracle(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	opchildKeeper := input.OPChildKeeper
	oracleKeeper := input.OracleKeeper

	defaultHostChainId := "test-host-1"
	defaultClientId := "test-client-id"
	bridgeInfo := types.BridgeInfo{
		L1ChainId:  defaultHostChainId,
		L1ClientId: defaultClientId,
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
		cp, err := slinkytypes.CurrencyPairFromString(currencyPair)
		require.NoError(t, err)
		err = oracleKeeper.CreateCurrencyPair(sdk.UnwrapSDKContext(ctx), cp)
		require.NoError(t, err)
	}

	cpStrategy, extendedCommitCodec, voteExtensionCodec := getSlinky(oracleKeeper)
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
			cp, err := slinkytypes.CurrencyPairFromString(currencyPairID)
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

	ms := keeper.NewMsgServerImpl(opchildKeeper)
	_, err = ms.UpdateOracle(ctx, types.NewMsgUpdateOracle(addrsStr[0], 11, extCommitBz))
	require.NoError(t, err)

	_, err = ms.UpdateOracle(ctx, types.NewMsgUpdateOracle(addrsStr[1], 11, extCommitBz))
	require.Error(t, err)
}
