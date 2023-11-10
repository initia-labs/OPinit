package keeper_test

import (
	"testing"
	"time"

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
		L1BlockTime:   time.Now().UTC(),
		L2BlockNumber: 101,
	}
	output2 := types.Output{
		OutputRoot:    []byte{4, 5, 6},
		L1BlockTime:   time.Now().UTC(),
		L2BlockNumber: 102,
	}
	output3 := types.Output{
		OutputRoot:    []byte{7, 8, 9},
		L1BlockTime:   time.Now().UTC(),
		L2BlockNumber: 103,
	}
	output4 := types.Output{
		OutputRoot:    []byte{10, 11, 12},
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

	input.OPHostKeeper.IterateOutputProposals(ctx, 1, func(bridgeId, outputIndex uint64, output types.Output) bool {
		require.Equal(t, bridgeId, uint64(1))
		switch outputIndex {
		case 1:
			require.Equal(t, output1, output)
		case 2:
			require.Equal(t, output2, output)
		case 3:
			require.Equal(t, output3, output)
		default:
			require.Fail(t, "should not enter here")
		}

		return false
	})
}

func Test_IsFinalized(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	input.OPHostKeeper.SetBridgeConfig(ctx, 1, types.BridgeConfig{
		Challenger:          "",
		Proposer:            "",
		SubmissionInterval:  100,
		StartingBlockNumber: 100,
		FinalizationPeriod:  time.Second * 10,
	})

	proposeTime := time.Now().UTC()
	err := input.OPHostKeeper.SetOutputProposal(ctx, 1, 1, types.Output{
		OutputRoot:    []byte{1, 2, 3},
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

	require.Equal(t, uint64(1), input.OPHostKeeper.IncreaseNextOutputIndex(ctx, 1))
	require.Equal(t, uint64(2), input.OPHostKeeper.IncreaseNextOutputIndex(ctx, 1))
	require.Equal(t, uint64(3), input.OPHostKeeper.IncreaseNextOutputIndex(ctx, 1))
	require.Equal(t, uint64(4), input.OPHostKeeper.IncreaseNextOutputIndex(ctx, 1))

	input.OPHostKeeper.SetNextOutputIndex(ctx, 1, 100)
	require.Equal(t, uint64(100), input.OPHostKeeper.IncreaseNextOutputIndex(ctx, 1))
	require.Equal(t, uint64(101), input.OPHostKeeper.GetNextOutputIndex(ctx, 1))
}
