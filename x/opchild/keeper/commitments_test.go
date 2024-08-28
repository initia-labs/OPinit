package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/initia-labs/OPinit/x/opchild/types"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
)

func Test_TrackWithdrawalCommitments(t *testing.T) {
	now := time.Now()
	ctx, input := createDefaultTestInput(t)

	err := input.OPChildKeeper.BridgeInfo.Set(ctx, types.BridgeInfo{
		BridgeConfig: ophosttypes.BridgeConfig{
			FinalizationPeriod: time.Second * 5,
		},
	})
	require.NoError(t, err)

	// H: 1, T: 0
	ctx = sdk.UnwrapSDKContext(ctx).WithBlockHeight(1).WithBlockTime(now)
	err = input.OPChildKeeper.TrackWithdrawalCommitments(ctx)
	require.NoError(t, err)

	err = input.OPChildKeeper.WithdrawalCommitments.Set(ctx, 1, types.WithdrawalCommitment{
		Commitment: types.CommitWithdrawal(1, "recipient", sdk.NewInt64Coin("uinit", 100)),
		SubmitTime: now,
	})
	require.NoError(t, err)

	err = input.OPChildKeeper.WithdrawalCommitments.Set(ctx, 2, types.WithdrawalCommitment{
		Commitment: types.CommitWithdrawal(1, "recipient", sdk.NewInt64Coin("uinit", 100)),
		SubmitTime: now,
	})
	require.NoError(t, err)

	// H: 2, T: 5
	ctx = sdk.UnwrapSDKContext(ctx).WithBlockHeight(2).WithBlockTime(now.Add(time.Second * 5))
	err = input.OPChildKeeper.TrackWithdrawalCommitments(ctx)
	require.NoError(t, err)

	// record historical withdrawal
	err = input.OPChildKeeper.WithdrawalCommitments.Set(ctx, 3, types.WithdrawalCommitment{
		Commitment: types.CommitWithdrawal(1, "recipient", sdk.NewInt64Coin("uinit", 100)),
		SubmitTime: now.Add(time.Second * 5),
	})
	require.NoError(t, err)

	// should not be removed
	_, err = input.OPChildKeeper.WithdrawalCommitments.Get(ctx, 1)
	require.NoError(t, err)
	_, err = input.OPChildKeeper.WithdrawalCommitments.Get(ctx, 2)
	require.NoError(t, err)
	_, err = input.OPChildKeeper.WithdrawalCommitments.Get(ctx, 3)
	require.NoError(t, err)

	// H: 3, T: 15
	ctx = sdk.UnwrapSDKContext(ctx).WithBlockHeight(3).WithBlockTime(now.Add(time.Second * 15))
	err = input.OPChildKeeper.TrackWithdrawalCommitments(ctx)
	require.NoError(t, err)

	// should be removed entries of height 1
	_, err = input.OPChildKeeper.WithdrawalCommitments.Get(ctx, 1)
	require.ErrorIs(t, err, collections.ErrNotFound)
	_, err = input.OPChildKeeper.WithdrawalCommitments.Get(ctx, 2)
	require.ErrorIs(t, err, collections.ErrNotFound)
	// should not be removed
	_, err = input.OPChildKeeper.WithdrawalCommitments.Get(ctx, 3)
	require.NoError(t, err)

	// H: 4, T: 25
	ctx = sdk.UnwrapSDKContext(ctx).WithBlockHeight(4).WithBlockTime(now.Add(time.Second * 25))
	err = input.OPChildKeeper.TrackWithdrawalCommitments(ctx)
	require.NoError(t, err)

	// should be removed entries of height 2
	_, err = input.OPChildKeeper.WithdrawalCommitments.Get(ctx, 3)
	require.ErrorIs(t, err, collections.ErrNotFound)
}
