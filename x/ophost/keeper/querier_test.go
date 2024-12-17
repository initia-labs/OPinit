package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/initia-labs/OPinit/x/ophost/keeper"
	"github.com/initia-labs/OPinit/x/ophost/types"
)

func Test_QueryBridge(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	config := types.BridgeConfig{
		Challenger:            addrs[0].String(),
		Proposer:              addrs[0].String(),
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: addrsStr[0], ChainType: types.BatchInfo_INITIA},
	}
	err := input.OPHostKeeper.SetBridgeConfig(ctx, 1, config)
	require.NoError(t, err)

	q := keeper.NewQuerier(input.OPHostKeeper)
	res, err := q.Bridge(ctx, &types.QueryBridgeRequest{
		BridgeId: 1,
	})

	require.NoError(t, err)
	require.Equal(t, types.QueryBridgeResponse{
		BridgeId:     1,
		BridgeAddr:   types.BridgeAddress(1).String(),
		BridgeConfig: config,
	}, *res)
}

func Test_QueryBridges(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	config1 := types.BridgeConfig{
		Challenger:            addrs[0].String(),
		Proposer:              addrs[0].String(),
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: addrsStr[0], ChainType: types.BatchInfo_INITIA},
	}
	config2 := types.BridgeConfig{
		Challenger:            addrs[1].String(),
		Proposer:              addrs[0].String(),
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{3, 4, 5},
		BatchInfo:             types.BatchInfo{Submitter: addrsStr[0], ChainType: types.BatchInfo_INITIA},
	}
	require.NoError(t, input.OPHostKeeper.SetBridgeConfig(ctx, 1, config1))
	require.NoError(t, input.OPHostKeeper.SetBridgeConfig(ctx, 2, config2))

	q := keeper.NewQuerier(input.OPHostKeeper)
	res, err := q.Bridges(ctx, &types.QueryBridgesRequest{})

	require.NoError(t, err)
	require.Equal(t, []types.QueryBridgeResponse{
		{
			BridgeId:     1,
			BridgeAddr:   types.BridgeAddress(1).String(),
			BridgeConfig: config1,
		}, {
			BridgeId:     2,
			BridgeAddr:   types.BridgeAddress(2).String(),
			BridgeConfig: config2,
		},
	}, res.Bridges)
}

func Test_QueryTokenPair(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	pair := types.TokenPair{
		L1Denom: "l1denom",
		L2Denom: types.L2Denom(1, "l1denom"),
	}
	err := input.OPHostKeeper.SetTokenPair(ctx, 1, pair.L2Denom, pair.L1Denom)
	require.NoError(t, err)

	q := keeper.NewQuerier(input.OPHostKeeper)
	res, err := q.TokenPairByL1Denom(ctx, &types.QueryTokenPairByL1DenomRequest{
		BridgeId: 1,
		L1Denom:  pair.L1Denom,
	})
	require.NoError(t, err)
	require.Equal(t, pair, res.TokenPair)

	res2, err := q.TokenPairByL2Denom(ctx, &types.QueryTokenPairByL2DenomRequest{
		BridgeId: 1,
		L2Denom:  pair.L2Denom,
	})
	require.NoError(t, err)
	require.Equal(t, pair, res2.TokenPair)
}

func Test_QueryTokenPairs(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	pair1 := types.TokenPair{
		L1Denom: "l1denom1",
		L2Denom: types.L2Denom(1, "l1denom1"),
	}
	pair2 := types.TokenPair{
		L1Denom: "l1denom2",
		L2Denom: types.L2Denom(1, "l1denom2"),
	}
	require.NoError(t, input.OPHostKeeper.SetTokenPair(ctx, 1, pair1.L2Denom, pair1.L1Denom))
	require.NoError(t, input.OPHostKeeper.SetTokenPair(ctx, 1, pair2.L2Denom, pair2.L1Denom))

	q := keeper.NewQuerier(input.OPHostKeeper)
	res, err := q.TokenPairs(ctx, &types.QueryTokenPairsRequest{
		BridgeId: 1,
	})

	require.NoError(t, err)
	require.Equal(t, []types.TokenPair{
		pair1, pair2,
	}, res.TokenPairs)
}

func Test_QueryOutputProposal(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	output := types.Output{
		OutputRoot:    []byte{1, 2, 3},
		L1BlockNumber: 1,
		L2BlockNumber: 100,
	}
	require.NoError(t, input.OPHostKeeper.SetOutputProposal(ctx, 1, 1, output))

	q := keeper.NewQuerier(input.OPHostKeeper)
	res, err := q.OutputProposal(ctx, &types.QueryOutputProposalRequest{
		BridgeId:    1,
		OutputIndex: 1,
	})
	require.NoError(t, err)
	require.Equal(t, output, res.OutputProposal)
}

func Test_QueryOutputProposals(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	output1 := types.Output{
		OutputRoot:    []byte{1, 2, 3},
		L1BlockNumber: 1,
		L1BlockTime:   ctx.BlockTime(),
		L2BlockNumber: 100,
	}
	output2 := types.Output{
		OutputRoot:    []byte{3, 4, 5},
		L1BlockNumber: 1,
		L1BlockTime:   ctx.BlockTime(),
		L2BlockNumber: 100,
	}
	require.NoError(t, input.OPHostKeeper.SetOutputProposal(ctx, 1, 1, output1))
	require.NoError(t, input.OPHostKeeper.SetOutputProposal(ctx, 1, 2, output2))

	q := keeper.NewQuerier(input.OPHostKeeper)
	res, err := q.OutputProposals(ctx, &types.QueryOutputProposalsRequest{
		BridgeId: 1,
	})

	require.NoError(t, err)
	require.Equal(t, []types.QueryOutputProposalResponse{
		{
			BridgeId:       1,
			OutputIndex:    1,
			OutputProposal: output1,
		}, {
			BridgeId:       1,
			OutputIndex:    2,
			OutputProposal: output2,
		},
	}, res.OutputProposals)
}

func Test_QueryLastFinalizedOutput(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	err := input.OPHostKeeper.SetBridgeConfig(ctx, 1, types.BridgeConfig{
		Proposer:              addrsStr[0],
		Challenger:            addrsStr[1],
		SubmissionInterval:    100,
		FinalizationPeriod:    time.Second * 10,
		SubmissionStartHeight: 1,
		BatchInfo:             types.BatchInfo{Submitter: addrsStr[0], ChainType: types.BatchInfo_INITIA},
	})
	require.NoError(t, err)

	proposeTime := time.Now().UTC()
	err = input.OPHostKeeper.SetOutputProposal(ctx, 1, 1, types.Output{
		OutputRoot:    []byte{1, 2, 3},
		L1BlockNumber: 1,
		L1BlockTime:   ctx.BlockTime(),
		L2BlockNumber: 100,
	})
	require.NoError(t, err)

	q := keeper.NewQuerier(input.OPHostKeeper)
	res, err := q.LastFinalizedOutput(ctx, &types.QueryLastFinalizedOutputRequest{
		BridgeId: 1,
	})
	require.NoError(t, err)
	require.Empty(t, res.OutputProposal)
	require.Zero(t, res.OutputIndex)

	res, err = q.LastFinalizedOutput(ctx.WithBlockTime(proposeTime.Add(time.Second*10)), &types.QueryLastFinalizedOutputRequest{
		BridgeId: 1,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), res.OutputIndex)
	require.Equal(t, types.Output{
		OutputRoot:    []byte{1, 2, 3},
		L1BlockNumber: 1,
		L1BlockTime:   ctx.BlockTime(),
		L2BlockNumber: 100,
	}, res.OutputProposal)
}

func Test_QueryClaimed(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	wh := [32]byte{1, 2, 3}

	// Check if the withdrawal is not claimed
	res, err := keeper.NewQuerier(input.OPHostKeeper).Claimed(ctx, &types.QueryClaimedRequest{
		BridgeId:       1,
		WithdrawalHash: wh[:],
	})
	require.NoError(t, err)
	require.False(t, res.Claimed)

	// Record the withdrawal as claimed
	err = input.OPHostKeeper.RecordProvenWithdrawal(ctx, 1, wh)
	require.NoError(t, err)

	// Check if the withdrawal is claimed
	res, err = keeper.NewQuerier(input.OPHostKeeper).Claimed(ctx, &types.QueryClaimedRequest{
		BridgeId:       1,
		WithdrawalHash: wh[:],
	})
	require.NoError(t, err)
	require.True(t, res.Claimed)
}

func Test_QueryNextL1Sequence(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	// update the next L1 sequence
	require.NoError(t, input.OPHostKeeper.NextL1Sequences.Set(ctx, 100, 100))

	q := keeper.NewQuerier(input.OPHostKeeper)
	res, err := q.NextL1Sequence(ctx, &types.QueryNextL1SequenceRequest{
		BridgeId: 100,
	})
	require.NoError(t, err)
	require.Equal(t, types.QueryNextL1SequenceResponse{NextL1Sequence: 100}, *res)
}

func Test_QueryBatchInfos(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	batchInfo := types.BatchInfo{Submitter: addrsStr[0], ChainType: types.BatchInfo_INITIA}
	require.NoError(t, input.OPHostKeeper.SetBatchInfo(ctx, 1, batchInfo, types.Output{}))

	newBatchInfo := types.BatchInfoWithOutput{
		BatchInfo: types.BatchInfo{
			Submitter: addrsStr[0],
			ChainType: types.BatchInfo_CELESTIA,
		},
		Output: types.Output{
			OutputRoot:    []byte{1, 2, 3},
			L1BlockNumber: 100,
			L2BlockNumber: 300,
		},
	}
	require.NoError(t, input.OPHostKeeper.SetBatchInfo(ctx, 1, newBatchInfo.BatchInfo, newBatchInfo.Output))

	q := keeper.NewQuerier(input.OPHostKeeper)
	res, err := q.BatchInfos(ctx, &types.QueryBatchInfosRequest{BridgeId: 1})

	require.NoError(t, err)
	require.Equal(t, []types.BatchInfoWithOutput{
		{
			BatchInfo: batchInfo,
		},
		{
			BatchInfo: newBatchInfo.BatchInfo,
			Output:    newBatchInfo.Output,
		},
	}, res.BatchInfos,
	)
}
