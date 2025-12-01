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

	// set denom pairs
	l2DenomFoo := ophosttypes.L2Denom(1, "foo")
	l2DenomBar := ophosttypes.L2Denom(1, "bar")
	err = input.OPChildKeeper.DenomPairs.Set(ctx, l2DenomFoo, "foo")
	require.NoError(t, err)
	err = input.OPChildKeeper.DenomPairs.Set(ctx, l2DenomBar, "bar")
	require.NoError(t, err)

	// set migration info
	err = input.OPChildKeeper.SetMigrationInfo(ctx, types.MigrationInfo{
		Denom:        l2DenomFoo,
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
	})
	require.NoError(t, err)
	err = input.OPChildKeeper.SetMigrationInfo(ctx, types.MigrationInfo{
		Denom:        l2DenomBar,
		IbcChannelId: "channel-1",
		IbcPortId:    "transfer",
	})
	require.NoError(t, err)

	// set ibc to l2 denom map
	err = input.OPChildKeeper.SetIBCToL2DenomMap(ctx, "ibc/foo", l2DenomFoo)
	require.NoError(t, err)
	err = input.OPChildKeeper.SetIBCToL2DenomMap(ctx, "ibc/bar", l2DenomBar)
	require.NoError(t, err)

	// set port id
	err = input.OPChildKeeper.PortID.Set(ctx, types.PortID)
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
	genState.MigrationInfos = []types.MigrationInfo{
		{
			Denom:        l2DenomFoo,
			IbcChannelId: "channel-0",
			IbcPortId:    "transfer",
		},
		{
			Denom:        l2DenomBar,
			IbcChannelId: "channel-1",
			IbcPortId:    "transfer",
		},
	}

	input.OPChildKeeper.InitGenesis(ctx, genState)
	genState_ := input.OPChildKeeper.ExportGenesis(ctx)
	require.Equal(t, genState, genState_)
}
