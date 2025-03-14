package keeper_test

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"

	testutilsims "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/initia-labs/OPinit/x/opchild/keeper"
	"github.com/initia-labs/OPinit/x/opchild/types"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
)

func Test_QueryValidator(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	valPubKeys := testutilsims.CreateTestPubKeys(1)
	val, err := types.NewValidator(valAddrs[0], valPubKeys[0], "validator1")
	require.NoError(t, err)

	require.NoError(t, input.OPChildKeeper.SetValidator(ctx, val))
	q := keeper.NewQuerier(&input.OPChildKeeper)

	res, err := q.Validator(ctx, &types.QueryValidatorRequest{ValidatorAddr: val.OperatorAddress})
	require.NoError(t, err)
	require.Equal(t, types.QueryValidatorResponse{Validator: val}, *res)
}

func Test_QueryValidators(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	valPubKeys := testutilsims.CreateTestPubKeys(2)
	val1, err := types.NewValidator(valAddrs[0], valPubKeys[0], "validator1")
	require.NoError(t, err)

	val2, err := types.NewValidator(valAddrs[1], valPubKeys[1], "validator2")
	require.NoError(t, err)
	require.NoError(t, input.OPChildKeeper.SetValidator(ctx, val1))
	require.NoError(t, input.OPChildKeeper.SetValidator(ctx, val2))
	q := keeper.NewQuerier(&input.OPChildKeeper)

	res, err := q.Validators(ctx, &types.QueryValidatorsRequest{})
	require.NoError(t, err)
	require.Len(t, res.Validators, 2)
}

func Test_QuerySetBridgeInfo(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	info := types.BridgeInfo{
		BridgeId:   1,
		BridgeAddr: addrsStr[1],
		BridgeConfig: ophosttypes.BridgeConfig{
			Challenger: addrsStr[2],
			Proposer:   addrsStr[3],
			BatchInfo: ophosttypes.BatchInfo{
				Submitter: addrsStr[4],
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
	ctx, input := createDefaultTestInput(t)

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
	ctx, input := createDefaultTestInput(t)

	// update the next L1 sequence
	require.NoError(t, input.OPChildKeeper.NextL1Sequence.Set(ctx, 100))

	q := keeper.NewQuerier(&input.OPChildKeeper)
	res, err := q.NextL1Sequence(ctx, &types.QueryNextL1SequenceRequest{})
	require.NoError(t, err)
	require.Equal(t, types.QueryNextL1SequenceResponse{NextL1Sequence: 100}, *res)
}

func Test_QueryNextL2Sequence(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	// update the next L2 sequence
	require.NoError(t, input.OPChildKeeper.NextL2Sequence.Set(ctx, 100))

	q := keeper.NewQuerier(&input.OPChildKeeper)
	res, err := q.NextL2Sequence(ctx, &types.QueryNextL2SequenceRequest{})
	require.NoError(t, err)
	require.Equal(t, types.QueryNextL2SequenceResponse{NextL2Sequence: 100}, *res)
}

func Test_QueryBaseDenom(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	bz := sha3.Sum256([]byte("base_denom"))
	denom := "l2/" + hex.EncodeToString(bz[:])

	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)
	_, err := ms.FinalizeTokenDeposit(ctx, &types.MsgFinalizeTokenDeposit{
		Sender:    addrsStr[0],
		Amount:    sdk.NewInt64Coin(denom, 100),
		From:      addrsStr[0],
		To:        addrsStr[1],
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
