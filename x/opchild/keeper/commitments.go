package keeper

import (
	"context"
	"errors"

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

	err = k.WithdrawalCommitments.Walk(ctx, nil, func(sequence uint64, wc types.WithdrawalCommitment) (stop bool, err error) {
		// stop the iteration if the current time is less than the submit time + retension period
		if curTime <= wc.SubmitTime.UTC().Unix()+retensionPeriod {
			return true, nil
		}

		err = k.WithdrawalCommitments.Remove(ctx, sequence)
		if err != nil {
			return true, err
		}

		return false, nil
	})
	if err != nil {
		return err
	}

	return nil
}
