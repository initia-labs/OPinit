package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/stretchr/testify/require"

	"github.com/initia-labs/OPinit/x/opchild/types"
)

func Test_WithdrawalCommitmentKey(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	sequence := uint64(1)
	now := time.Now().UTC()

	commitment := types.WithdrawalCommitment{
		Commitment: []byte{1, 2, 3, 4},
		SubmitTime: now,
	}
	input.OPChildKeeper.WithdrawalCommitments.Set(
		ctx,
		sequence,
		commitment,
	)

	bz, err := input.OPChildKeeper.StoreService().OpenKVStore(ctx).Get(types.WithdrawalCommitmentKey(sequence))
	require.NoError(t, err)
	ret, err := codec.CollValue[types.WithdrawalCommitment](input.Cdc).Decode(bz)
	require.NoError(t, err)
	require.Equal(t, commitment, ret)
}
