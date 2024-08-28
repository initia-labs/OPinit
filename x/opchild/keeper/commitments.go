package keeper

import (
	"context"
	"errors"
	"time"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/OPinit/x/opchild/types"
)

func (k Keeper) TrackWithdrawalCommitments(ctx context.Context) error {
	info, err := k.BridgeInfo.Get(ctx)
	if err != nil && errors.Is(err, collections.ErrNotFound) {
		return nil
	} else if err != nil {
		return err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	curTime := sdkCtx.BlockTime().UTC().Unix()
	retensionPeriod := int64(info.BridgeConfig.FinalizationPeriod.Seconds() * 2)

	err = k.CommitmentTimes.Walk(ctx, nil, func(sequence uint64, submitTime time.Time) (stop bool, err error) {
		// stop the iteration if the current time is less than the submit time + retension period
		if curTime <= submitTime.UTC().Unix()+retensionPeriod {
			return true, nil
		}

		return false, k.RemoveWithdrawalCommitment(ctx, sequence)
	})
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) SetWithdrawalCommitment(ctx context.Context, sequence uint64, wc types.WithdrawalCommitment) error {
	if err := k.CommitmentTimes.Set(ctx, sequence, wc.SubmitTime); err != nil {
		return err
	}
	if err := k.Commitments.Set(ctx, sequence, wc.Commitment); err != nil {
		return err
	}

	return nil
}

func (k Keeper) GetWithdrawalCommitment(ctx context.Context, sequence uint64) (types.WithdrawalCommitment, error) {
	commitment, err := k.Commitments.Get(ctx, sequence)
	if err != nil {
		return types.WithdrawalCommitment{}, err
	}

	submitTime, err := k.CommitmentTimes.Get(ctx, sequence)
	if err != nil {
		return types.WithdrawalCommitment{}, err
	}

	return types.WithdrawalCommitment{
		Commitment: commitment,
		SubmitTime: submitTime,
	}, nil
}

func (k Keeper) RemoveWithdrawalCommitment(ctx context.Context, sequence uint64) error {
	if err := k.CommitmentTimes.Remove(ctx, sequence); err != nil {
		return err
	}
	if err := k.Commitments.Remove(ctx, sequence); err != nil {
		return err
	}

	return nil
}
