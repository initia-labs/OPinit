package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/initia-labs/OPinit/x/opchild/types"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
)

func Test_GenesisImportExport(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	require.NoError(t, input.OPChildKeeper.SetNextL2Sequence(ctx, 1))

	seq, err := input.OPChildKeeper.IncreaseNextL2Sequence(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(1), seq)
	seq, err = input.OPChildKeeper.IncreaseNextL2Sequence(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(2), seq)

	_, err = input.OPChildKeeper.IncreaseNextL1Sequence(ctx) // 2
	require.NoError(t, err)
	_, err = input.OPChildKeeper.IncreaseNextL1Sequence(ctx) // 3
	require.NoError(t, err)

	err = input.OPChildKeeper.DenomPairs.Set(ctx, "foo", "bar")
	require.NoError(t, err)

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
				ChainType: ophosttypes.BatchInfo_INITIA,
			},
			SubmissionInterval:    time.Minute,
			FinalizationPeriod:    time.Hour,
			SubmissionStartHeight: 1,
			Metadata:              []byte("metadata"),
		},
	}

	input.OPChildKeeper.InitGenesis(ctx, genState)
	genState_ := input.OPChildKeeper.ExportGenesis(ctx)
	require.Equal(t, genState, genState_)
}
