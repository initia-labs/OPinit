package keeper_test

// TODO - implement test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
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
	input.OPChildKeeper.InitGenesis(ctx, genState)
	_genState := input.OPChildKeeper.ExportGenesis(ctx)
	require.Equal(t, genState, _genState)
	fmt.Printf("genState: %v\n", genState)
}
