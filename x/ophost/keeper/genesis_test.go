package keeper_test

import (
	"testing"
	"time"

	"github.com/initia-labs/OPinit/x/ophost/testutil"
	"github.com/initia-labs/OPinit/x/ophost/types"
	"github.com/stretchr/testify/require"
)

func Test_GenesisExport(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	params := input.OPHostKeeper.GetParams(ctx)
	config1 := types.BridgeConfig{
		Challenger:            testutil.AddrsStr[1],
		Proposer:              testutil.AddrsStr[0],
		SubmissionInterval:    100,
		FinalizationPeriod:    100,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: types.BatchInfo_INITIA},
	}
	config2 := types.BridgeConfig{
		Challenger:            testutil.AddrsStr[2],
		Proposer:              testutil.AddrsStr[3],
		SubmissionInterval:    200,
		FinalizationPeriod:    200,
		SubmissionStartHeight: 1,
		Metadata:              []byte{3, 4, 5},
		BatchInfo:             types.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: types.BatchInfo_INITIA},
	}
	require.NoError(t, input.OPHostKeeper.SetBridgeConfig(ctx, 1, config1))
	require.NoError(t, input.OPHostKeeper.SetBridgeConfig(ctx, 2, config2))

	require.NoError(t, input.OPHostKeeper.SetNextBridgeId(ctx, 3))
	require.NoError(t, input.OPHostKeeper.SetNextL1Sequence(ctx, 1, 100))
	require.NoError(t, input.OPHostKeeper.SetNextL1Sequence(ctx, 2, 200))
	require.NoError(t, input.OPHostKeeper.SetNextOutputIndex(ctx, 1, 10))
	require.NoError(t, input.OPHostKeeper.SetNextOutputIndex(ctx, 2, 20))

	output1 := types.Output{
		OutputRoot:    []byte{1, 2, 3},
		L1BlockTime:   time.Now().UTC(),
		L1BlockNumber: 1,
		L2BlockNumber: 100,
	}
	output2 := types.Output{
		OutputRoot:    []byte{1, 2, 3},
		L1BlockTime:   time.Now().UTC(),
		L1BlockNumber: 2,
		L2BlockNumber: 200,
	}
	output3 := types.Output{
		OutputRoot:    []byte{1, 2, 3},
		L1BlockTime:   time.Now().UTC(),
		L1BlockNumber: 1,
		L2BlockNumber: 100,
	}
	require.NoError(t, input.OPHostKeeper.SetOutputProposal(ctx, 1, 1, output1))
	require.NoError(t, input.OPHostKeeper.SetOutputProposal(ctx, 1, 2, output2))
	require.NoError(t, input.OPHostKeeper.SetOutputProposal(ctx, 2, 1, output3))

	require.NoError(t, input.OPHostKeeper.SetTokenPair(ctx, 1, "l2denom", "l1denom"))
	require.NoError(t, input.OPHostKeeper.SetTokenPair(ctx, 2, "l12denom", "l11denom"))

	require.NoError(t, input.OPHostKeeper.RecordProvenWithdrawal(ctx, 1, [32]byte{1, 2, 3}))
	require.NoError(t, input.OPHostKeeper.RecordProvenWithdrawal(ctx, 1, [32]byte{3, 4, 5}))

	require.NoError(t, input.OPHostKeeper.SetBatchInfo(ctx, 1, types.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: types.BatchInfo_INITIA}, types.Output{}))
	require.NoError(t, input.OPHostKeeper.SetBatchInfo(ctx, 1, types.BatchInfo{Submitter: testutil.AddrsStr[1], ChainType: types.BatchInfo_CELESTIA}, output1))
	require.NoError(t, input.OPHostKeeper.SetBatchInfo(ctx, 1, types.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: types.BatchInfo_INITIA}, output3))

	// set migration info
	require.NoError(t, input.OPHostKeeper.SetMigrationInfo(ctx, types.MigrationInfo{
		BridgeId:     1,
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
		L1Denom:      "test1",
	}))

	// set migration info
	require.NoError(t, input.OPHostKeeper.SetMigrationInfo(ctx, types.MigrationInfo{
		BridgeId:     2,
		IbcChannelId: "channel-1",
		IbcPortId:    "transfer",
		L1Denom:      "test2",
	}))

	genState := input.OPHostKeeper.ExportGenesis(ctx)
	require.Equal(t, uint64(3), genState.NextBridgeId)
	require.Equal(t, params, genState.Params)
	require.Equal(t, types.Bridge{
		BridgeId:        1,
		NextL1Sequence:  100,
		NextOutputIndex: 10,
		BridgeConfig:    config1,
		TokenPairs: []types.TokenPair{
			{
				L1Denom: "l1denom",
				L2Denom: "l2denom",
			},
		},
		Proposals: []types.WrappedOutput{
			{
				OutputIndex:    1,
				OutputProposal: output1,
			},
			{
				OutputIndex:    2,
				OutputProposal: output2,
			},
		},
		ProvenWithdrawals: [][]byte{
			{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{3, 4, 5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		BatchInfos: []types.BatchInfoWithOutput{
			{BatchInfo: types.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: types.BatchInfo_INITIA}, Output: types.Output{}},
			{BatchInfo: types.BatchInfo{Submitter: testutil.AddrsStr[1], ChainType: types.BatchInfo_CELESTIA}, Output: output1},
			{BatchInfo: types.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: types.BatchInfo_INITIA}, Output: output3},
		},
	}, genState.Bridges[0])

	require.Equal(t, []types.MigrationInfo{
		{
			BridgeId:     1,
			IbcChannelId: "channel-0",
			IbcPortId:    "transfer",
			L1Denom:      "test1",
		},
		{
			BridgeId:     2,
			IbcChannelId: "channel-1",
			IbcPortId:    "transfer",
			L1Denom:      "test2",
		},
	}, genState.MigrationInfos)
}

func Test_GenesisImportExport(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
	params := input.OPHostKeeper.GetParams(ctx)
	config1 := types.BridgeConfig{
		Challenger:            testutil.AddrsStr[1],
		Proposer:              testutil.AddrsStr[0],
		SubmissionInterval:    100,
		FinalizationPeriod:    100,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             types.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: types.BatchInfo_INITIA},
	}
	output1 := types.Output{
		OutputRoot:    []byte{1, 2, 3},
		L1BlockNumber: 1,
		L1BlockTime:   time.Now().UTC(),
		L2BlockNumber: 100,
	}
	output2 := types.Output{
		OutputRoot:    []byte{1, 2, 3},
		L1BlockNumber: 2,
		L1BlockTime:   time.Now().UTC(),
		L2BlockNumber: 200,
	}
	output3 := types.Output{
		OutputRoot:    []byte{1, 2, 3},
		L1BlockNumber: 1,
		L1BlockTime:   time.Now().UTC(),
		L2BlockNumber: 100,
	}

	genState := &types.GenesisState{
		Params: params,
		Bridges: []types.Bridge{
			{
				BridgeId:        1,
				NextL1Sequence:  100,
				NextOutputIndex: 10,
				BridgeConfig:    config1,
				TokenPairs: []types.TokenPair{
					{
						L1Denom: "l1denom",
						L2Denom: "l2denom",
					},
				},
				Proposals: []types.WrappedOutput{
					{
						OutputIndex:    1,
						OutputProposal: output1,
					},
					{
						OutputIndex:    2,
						OutputProposal: output2,
					},
				},
				ProvenWithdrawals: [][]byte{
					{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					{3, 4, 5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
				BatchInfos: []types.BatchInfoWithOutput{
					{BatchInfo: types.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: types.BatchInfo_INITIA}, Output: types.Output{}},
					{BatchInfo: types.BatchInfo{Submitter: testutil.AddrsStr[1], ChainType: types.BatchInfo_CELESTIA}, Output: output1},
					{BatchInfo: types.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: types.BatchInfo_INITIA}, Output: output3},
				},
			}},
		NextBridgeId: 2,
		MigrationInfos: []types.MigrationInfo{
			{
				BridgeId:     1,
				IbcChannelId: "channel-0",
				IbcPortId:    "transfer",
				L1Denom:      "test1",
			},
			{
				BridgeId:     2,
				IbcChannelId: "channel-1",
				IbcPortId:    "transfer",
				L1Denom:      "test2",
			},
		},
	}

	input.OPHostKeeper.InitGenesis(ctx, genState)
	_genState := input.OPHostKeeper.ExportGenesis(ctx)
	require.Equal(t, genState, _genState)
}
