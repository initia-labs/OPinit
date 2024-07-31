package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

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

	input.OPChildKeeper.IncreaseNextL1Sequence(ctx) // 2
	input.OPChildKeeper.IncreaseNextL1Sequence(ctx) // 3

	genState := input.OPChildKeeper.ExportGenesis(ctx)
	require.Nil(t, genState.BridgeInfo)

	// set bridge info
	genState.BridgeInfo = &types.BridgeInfo{
		BridgeId:   1,
		BridgeAddr: addrsStr[1],
		BridgeConfig: ophosttypes.BridgeConfig{
			Challengers: []string{addrsStr[2]},
			Proposer:    addrsStr[3],
			BatchInfo: ophosttypes.BatchInfo{
				Submitter: addrsStr[4],
				ChainType: ophosttypes.BatchInfo_CHAIN_TYPE_INITIA,
			},
			SubmissionInterval:    time.Minute,
			FinalizationPeriod:    time.Hour,
			SubmissionStartHeight: 1,
			Metadata:              []byte("metadata"),
		},
	}

	genState.PendingDeposits = append(
		genState.PendingDeposits,
		types.PendingDeposits{Recipient: addrsStr[0], Coins: sdk.NewCoins(sdk.NewInt64Coin("eth", 100))},
	)

	input.OPChildKeeper.InitGenesis(ctx, genState)
	genState_ := input.OPChildKeeper.ExportGenesis(ctx)
	require.Equal(t, genState, genState_)
}
