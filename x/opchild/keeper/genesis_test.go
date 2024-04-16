package keeper_test

// TODO - implement test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/initia-labs/OPinit/x/opchild/types"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
)

func Test_GenesisImportExport(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	input.OPChildKeeper.SetNextL2Sequence(ctx, 1)

	seq, err := input.OPChildKeeper.IncreaseNextL2Sequence(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(1), seq)
	seq, err = input.OPChildKeeper.IncreaseNextL2Sequence(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(2), seq)

	input.OPChildKeeper.RecordFinalizedL1Sequence(ctx, 1)
	input.OPChildKeeper.RecordFinalizedL1Sequence(ctx, 2)

	genState := input.OPChildKeeper.ExportGenesis(ctx)
	require.Nil(t, genState.BridgeInfo)

	// set bridge info
	genState.BridgeInfo = &types.BridgeInfo{
		BridgeId:   1,
		BridgeAddr: addrsStr[1],
		BridgeConfig: ophosttypes.BridgeConfig{
			Challenger: addrsStr[2],
			Proposer:   addrsStr[3],
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

	input.OPChildKeeper.InitGenesis(ctx, genState)
	genState_ := input.OPChildKeeper.ExportGenesis(ctx)
	require.Equal(t, genState, genState_)
}
