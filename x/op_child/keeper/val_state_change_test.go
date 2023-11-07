package keeper_test

import (
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"

	testutilsims "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/initia-labs/OPinit/x/op_child/types"
	"github.com/stretchr/testify/require"
)

func Test_BlockValidatorUpdates(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	valPubKeys := testutilsims.CreateTestPubKeys(2)
	val1, err := types.NewValidator(valAddrs[1], valPubKeys[0], "validator1")
	require.NoError(t, err)

	val2, err := types.NewValidator(valAddrs[2], valPubKeys[1], "validator2")
	require.NoError(t, err)

	// set validators
	input.OPChildKeeper.SetValidator(ctx, val1)
	input.OPChildKeeper.SetValidator(ctx, val2)

	// apply val updates
	updates := input.OPChildKeeper.BlockValidatorUpdates(ctx)
	valTmConsPubKey1, err := val1.TmConsPublicKey()
	valTmConsPubKey2, err := val2.TmConsPublicKey()
	require.Len(t, updates, 2)
	require.Contains(t, updates, abci.ValidatorUpdate{
		PubKey: valTmConsPubKey1,
		Power:  val1.ConsensusPower(),
	})
	require.Contains(t, updates, abci.ValidatorUpdate{
		PubKey: valTmConsPubKey2,
		Power:  val2.ConsensusPower(),
	})

	// no changes
	updates = input.OPChildKeeper.BlockValidatorUpdates(ctx)
	require.Equal(t, []abci.ValidatorUpdate{}, updates)

	// val2 removed
	val2.ConsPower = 0
	input.OPChildKeeper.SetValidator(ctx, val2)
	updates = input.OPChildKeeper.BlockValidatorUpdates(ctx)
	require.Equal(t, []abci.ValidatorUpdate{val2.ABCIValidatorUpdateZero()}, updates)
}
