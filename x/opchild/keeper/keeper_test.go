package keeper_test

import (
	"testing"

	"cosmossdk.io/collections"
	"github.com/stretchr/testify/require"

	"github.com/initia-labs/OPinit/x/opchild/types"
)

func Test_WithdrawalCommitmentKey(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	sequence := uint64(1)
	err := input.OPChildKeeper.Commitments.Set(
		ctx,
		sequence,
		[]byte{1, 2, 3, 4},
	)
	require.NoError(t, err)

	bz, err := input.OPChildKeeper.StoreService().OpenKVStore(ctx).Get(types.WithdrawalCommitmentKey(sequence))
	require.NoError(t, err)
	ret, err := collections.BytesValue.Decode(bz)
	require.NoError(t, err)
	require.Equal(t, []byte{1, 2, 3, 4}, ret)
}
