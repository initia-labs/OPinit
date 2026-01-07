package keeper_test

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"

	testutilsims "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	"github.com/initia-labs/OPinit/x/opchild/keeper"
	"github.com/initia-labs/OPinit/x/opchild/testutil"
	"github.com/initia-labs/OPinit/x/opchild/types"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
)

func Test_QueryValidator(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	valPubKeys := testutilsims.CreateTestPubKeys(1)
	val, err := types.NewValidator(testutil.ValAddrs[0], valPubKeys[0], "validator1")
	require.NoError(t, err)

	require.NoError(t, input.OPChildKeeper.SetValidator(ctx, val))
	q := keeper.NewQuerier(&input.OPChildKeeper)

	res, err := q.Validator(ctx, &types.QueryValidatorRequest{ValidatorAddr: val.OperatorAddress})
	require.NoError(t, err)
	require.Equal(t, types.QueryValidatorResponse{Validator: val}, *res)
}

func Test_QueryValidators(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	valPubKeys := testutilsims.CreateTestPubKeys(2)
	val1, err := types.NewValidator(testutil.ValAddrs[0], valPubKeys[0], "validator1")
	require.NoError(t, err)

	val2, err := types.NewValidator(testutil.ValAddrs[1], valPubKeys[1], "validator2")
	require.NoError(t, err)
	require.NoError(t, input.OPChildKeeper.SetValidator(ctx, val1))
	require.NoError(t, input.OPChildKeeper.SetValidator(ctx, val2))
	q := keeper.NewQuerier(&input.OPChildKeeper)

	res, err := q.Validators(ctx, &types.QueryValidatorsRequest{})
	require.NoError(t, err)
	require.Len(t, res.Validators, 2)
}

func Test_QuerySetBridgeInfo(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	info := types.BridgeInfo{
		BridgeId:   1,
		BridgeAddr: testutil.AddrsStr[1],
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
	err := input.OPChildKeeper.BridgeInfo.Set(ctx, info)
	require.NoError(t, err)

	q := keeper.NewQuerier(&input.OPChildKeeper)
	res, err := q.BridgeInfo(ctx, &types.QueryBridgeInfoRequest{})
	require.NoError(t, err)
	require.Equal(t, info, res.BridgeInfo)
}

func Test_QueryParams(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	params, err := input.OPChildKeeper.GetParams(ctx)
	require.NoError(t, err)
	params.MinGasPrices = sdk.NewDecCoins(sdk.NewInt64DecCoin("stake", 1))
	require.NoError(t, input.OPChildKeeper.SetParams(ctx, params))

	q := keeper.NewQuerier(&input.OPChildKeeper)
	res, err := q.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, params, res.Params)
}

func Test_QueryNextL1Sequence(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	// update the next L1 sequence
	require.NoError(t, input.OPChildKeeper.NextL1Sequence.Set(ctx, 100))

	q := keeper.NewQuerier(&input.OPChildKeeper)
	res, err := q.NextL1Sequence(ctx, &types.QueryNextL1SequenceRequest{})
	require.NoError(t, err)
	require.Equal(t, types.QueryNextL1SequenceResponse{NextL1Sequence: 100}, *res)
}

func Test_QueryNextL2Sequence(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	// update the next L2 sequence
	require.NoError(t, input.OPChildKeeper.NextL2Sequence.Set(ctx, 100))

	q := keeper.NewQuerier(&input.OPChildKeeper)
	res, err := q.NextL2Sequence(ctx, &types.QueryNextL2SequenceRequest{})
	require.NoError(t, err)
	require.Equal(t, types.QueryNextL2SequenceResponse{NextL2Sequence: 100}, *res)
}

func Test_QueryBaseDenom(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	bz := sha3.Sum256([]byte("base_denom"))
	denom := "l2/" + hex.EncodeToString(bz[:])

	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)
	_, err := ms.FinalizeTokenDeposit(ctx, &types.MsgFinalizeTokenDeposit{
		Sender:    testutil.AddrsStr[0],
		Amount:    sdk.NewInt64Coin(denom, 100),
		From:      testutil.AddrsStr[0],
		To:        testutil.AddrsStr[1],
		Sequence:  1,
		Height:    1,
		BaseDenom: "base_denom",
	})
	require.NoError(t, err)

	q := keeper.NewQuerier(&input.OPChildKeeper)
	res, err := q.BaseDenom(ctx, &types.QueryBaseDenomRequest{Denom: denom})
	require.NoError(t, err)

	require.Equal(t, types.QueryBaseDenomResponse{BaseDenom: "base_denom"}, *res)
}

func Test_QueryMigrationInfo(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
	q := keeper.NewQuerier(&input.OPChildKeeper)

	denom := "l2/denom"
	port := "transfer"
	channel := "channel-0"
	migrationInfo := types.MigrationInfo{
		Denom:        denom,
		IbcChannelId: channel,
		IbcPortId:    port,
	}

	_, err := q.MigrationInfo(ctx, &types.QueryMigrationInfoRequest{Denom: denom})
	require.Error(t, err) // migration info not found

	// register migration info
	require.NoError(t, input.OPChildKeeper.SetMigrationInfo(ctx, migrationInfo))

	_, err = q.MigrationInfo(ctx, &types.QueryMigrationInfoRequest{Denom: denom})
	require.Error(t, err) // base denom not found

	// set base denom
	require.NoError(t, input.OPChildKeeper.DenomPairs.Set(ctx, denom, "test1"))

	res, err := q.MigrationInfo(ctx, &types.QueryMigrationInfoRequest{Denom: denom})
	require.NoError(t, err)

	ibcDenom := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom(migrationInfo.IbcPortId, migrationInfo.IbcChannelId, "test1")).IBCDenom()
	require.Equal(t, types.QueryMigrationInfoResponse{MigrationInfo: types.MigrationInfo{
		Denom:        denom,
		IbcChannelId: channel,
		IbcPortId:    port,
	}, IbcDenom: ibcDenom}, *res)
}
