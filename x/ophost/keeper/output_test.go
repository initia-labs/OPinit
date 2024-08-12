package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/collections"
	"github.com/initia-labs/OPinit/x/ophost/types"
	"github.com/stretchr/testify/require"
)

func Test_OutputProposal(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	output := types.Output{
		OutputRoot:    []byte{1, 2, 3},
		L1BlockTime:   time.Now().UTC(),
		L2BlockNumber: 100,
	}
	err := input.OPHostKeeper.SetOutputProposal(ctx, 1, 1, output)
	require.NoError(t, err)

	_output, err := input.OPHostKeeper.GetOutputProposal(ctx, 1, 1)
	require.NoError(t, err)
	require.Equal(t, output, _output)
}

func Test_IterateOutputProposal(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	output1 := types.Output{
		OutputRoot:    []byte{1, 2, 3},
		L1BlockNumber: 1,
		L1BlockTime:   time.Now().UTC(),
		L2BlockNumber: 101,
	}
	output2 := types.Output{
		OutputRoot:    []byte{4, 5, 6},
		L1BlockNumber: 2,
		L1BlockTime:   time.Now().UTC(),
		L2BlockNumber: 102,
	}
	output3 := types.Output{
		OutputRoot:    []byte{7, 8, 9},
		L1BlockNumber: 3,
		L1BlockTime:   time.Now().UTC(),
		L2BlockNumber: 103,
	}
	output4 := types.Output{
		OutputRoot:    []byte{10, 11, 12},
		L1BlockNumber: 4,
		L1BlockTime:   time.Now().UTC(),
		L2BlockNumber: 104,
	}

	err := input.OPHostKeeper.SetOutputProposal(ctx, 1, 1, output1)
	require.NoError(t, err)

	err = input.OPHostKeeper.SetOutputProposal(ctx, 1, 2, output2)
	require.NoError(t, err)

	err = input.OPHostKeeper.SetOutputProposal(ctx, 1, 3, output3)
	require.NoError(t, err)

	err = input.OPHostKeeper.SetOutputProposal(ctx, 2, 1, output4)
	require.NoError(t, err)

	err = input.OPHostKeeper.IterateOutputProposals(ctx, 1, func(key collections.Pair[uint64, uint64], output types.Output) (stop bool, err error) {
		require.Equal(t, key.K1(), uint64(1))
		switch key.K2() {
		case 1:
			require.Equal(t, output1, output)
		case 2:
			require.Equal(t, output2, output)
		case 3:
			require.Equal(t, output3, output)
		default:
			require.Fail(t, "should not enter here")
		}

		return false, nil
	})
	require.NoError(t, err)
}

func Test_IsFinalized(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	err := input.OPHostKeeper.SetBridgeConfig(ctx, 1, types.BridgeConfig{
		Challengers:           []string{addrsStr[1]},
		Proposer:              addrsStr[0],
		SubmissionInterval:    100,
		FinalizationPeriod:    time.Second * 10,
		SubmissionStartHeight: 1,
		BatchInfo:             types.BatchInfo{Submitter: addrsStr[0], ChainType: types.BatchInfo_CHAIN_TYPE_INITIA},
	})
	require.NoError(t, err)

	proposeTime := time.Now().UTC()
	err = input.OPHostKeeper.SetOutputProposal(ctx, 1, 1, types.Output{
		OutputRoot:    []byte{1, 2, 3},
		L1BlockNumber: 1,
		L1BlockTime:   proposeTime,
		L2BlockNumber: 100,
	})
	require.NoError(t, err)

	ok, err := input.OPHostKeeper.IsFinalized(ctx.WithBlockTime(proposeTime), 1, 1)
	require.NoError(t, err)
	require.False(t, ok)

	ok, err = input.OPHostKeeper.IsFinalized(ctx.WithBlockTime(proposeTime.Add(time.Second*10)), 1, 1)
	require.NoError(t, err)
	require.True(t, ok)
}

func Test_NextOutputIndex(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	index, err := input.OPHostKeeper.IncreaseNextOutputIndex(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, uint64(1), index)
	index, err = input.OPHostKeeper.IncreaseNextOutputIndex(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, uint64(2), index)
	index, err = input.OPHostKeeper.IncreaseNextOutputIndex(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, uint64(3), index)
	index, err = input.OPHostKeeper.IncreaseNextOutputIndex(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, uint64(4), index)

	err = input.OPHostKeeper.SetNextOutputIndex(ctx, 1, 100)
	require.NoError(t, err)

	index, err = input.OPHostKeeper.IncreaseNextOutputIndex(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, uint64(100), index)

	index, err = input.OPHostKeeper.GetNextOutputIndex(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, uint64(101), index)
}

func Test_GetLastFinalizedOutput(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	err := input.OPHostKeeper.SetBridgeConfig(ctx, 1, types.BridgeConfig{
		Proposer:              addrsStr[0],
		Challengers:           []string{addrsStr[1]},
		SubmissionInterval:    100,
		FinalizationPeriod:    time.Second * 10,
		SubmissionStartHeight: 1,
		BatchInfo:             types.BatchInfo{Submitter: addrsStr[0], ChainType: types.BatchInfo_CHAIN_TYPE_INITIA},
	})
	require.NoError(t, err)

	proposeTime := time.Now().UTC()
	err = input.OPHostKeeper.SetOutputProposal(ctx, 1, 1, types.Output{
		OutputRoot:    []byte{1, 2, 3},
		L1BlockNumber: 1,
		L1BlockTime:   proposeTime,
		L2BlockNumber: 100,
	})
	require.NoError(t, err)

	index, output, err := input.OPHostKeeper.GetLastFinalizedOutput(ctx.WithBlockTime(proposeTime), 1)
	require.NoError(t, err)
	require.Empty(t, output)
	require.Equal(t, uint64(0), index)

	index, output, err = input.OPHostKeeper.GetLastFinalizedOutput(ctx.WithBlockTime(proposeTime.Add(time.Second*10)), 1)
	require.NoError(t, err)
	require.Equal(t, uint64(1), index)
	require.Equal(t, types.Output{
		OutputRoot:    []byte{1, 2, 3},
		L1BlockNumber: 1,
		L1BlockTime:   proposeTime,
		L2BlockNumber: 100,
	}, output)
}

func Test_DeleteOutputProposal(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	output := types.Output{
		OutputRoot:    []byte{1, 2, 3},
		L1BlockNumber: 1,
		L1BlockTime:   ctx.BlockTime(),
		L2BlockNumber: 100,
	}
	err := input.OPHostKeeper.SetOutputProposal(ctx, 1, 1, output)
	require.NoError(t, err)

	err = input.OPHostKeeper.SetBridgeConfig(ctx, 1, types.BridgeConfig{
		Proposer:              addrsStr[0],
		Challengers:           []string{addrsStr[1]},
		SubmissionInterval:    100,
		FinalizationPeriod:    time.Second * 10,
		SubmissionStartHeight: 1,
		BatchInfo:             types.BatchInfo{Submitter: addrsStr[0], ChainType: types.BatchInfo_CHAIN_TYPE_INITIA},
	})
	require.NoError(t, err)

	// delete should fail due to already finalized error
	err = input.OPHostKeeper.DeleteOutputProposal(ctx.WithBlockTime(ctx.BlockTime().Add(time.Second*11)), 1, 1)
	require.ErrorIs(t, err, types.ErrAlreadyFinalized)

	// delete should success
	err = input.OPHostKeeper.DeleteOutputProposal(ctx.WithBlockTime(ctx.BlockTime().Add(time.Second*9)), 1, 1)
	require.NoError(t, err)
}
