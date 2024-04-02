package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/initia-labs/OPinit/x/ophost/types"
)

////////////////////////////////////
// OutputProposal

// GetLastFinalizedOutput returns the last finalized output proposal.
// if there is no finalized output, it returns empty output and zero index.
func (k Keeper) GetLastFinalizedOutput(ctx context.Context, bridgeId uint64) (outputIndex uint64, outputProposal types.Output, err error) {
	bridgeConfig, err := k.GetBridgeConfig(ctx, bridgeId)
	if err != nil {
		return outputIndex, outputProposal, err
	}

	cb := func(key collections.Pair[uint64, uint64], output types.Output) (stop bool, err error) {
		if ok, err := k.isFinalizedWithConfig(ctx, bridgeConfig, output); err != nil {
			return true, err
		} else if ok {
			outputIndex = key.K2()
			outputProposal = output
			return true, nil
		}
		return false, nil
	}
	if err := k.ReverseIterateOutputProposals(ctx, bridgeId, cb); err != nil {
		return outputIndex, outputProposal, err
	}

	return outputIndex, outputProposal, nil
}

func (k Keeper) SetOutputProposal(ctx context.Context, bridgeId, outputIndex uint64, outputProposal types.Output) error {
	return k.OutputProposals.Set(ctx, collections.Join(bridgeId, outputIndex), outputProposal)
}

func (k Keeper) GetOutputProposal(ctx context.Context, bridgeId, outputIndex uint64) (outputProposal types.Output, err error) {
	return k.OutputProposals.Get(ctx, collections.Join(bridgeId, outputIndex))
}

func (k Keeper) DeleteOutputProposal(ctx context.Context, bridgeId, outputIndex uint64) error {
	output, err := k.GetOutputProposal(ctx, bridgeId, outputIndex)
	if err != nil {
		return err
	}

	if isFinalized, err := k.isFinalized(ctx, bridgeId, output); err != nil {
		return err
	} else if isFinalized {
		return types.ErrAlreadyFinalized
	}

	return k.OutputProposals.Remove(ctx, collections.Join(bridgeId, outputIndex))
}

func (k Keeper) IterateOutputProposals(ctx context.Context, bridgeId uint64, cb func(key collections.Pair[uint64, uint64], output types.Output) (stop bool, err error)) error {
	return k.OutputProposals.Walk(ctx, collections.NewPrefixedPairRange[uint64, uint64](bridgeId), cb)
}

func (k Keeper) ReverseIterateOutputProposals(ctx context.Context, bridgeId uint64, cb func(key collections.Pair[uint64, uint64], output types.Output) (stop bool, err error)) error {
	return k.OutputProposals.Walk(ctx, collections.NewPrefixedPairRange[uint64, uint64](bridgeId).Descending(), cb)
}

func (k Keeper) IsFinalized(ctx context.Context, bridgeId, outputIndex uint64) (bool, error) {
	output, err := k.GetOutputProposal(ctx, bridgeId, outputIndex)
	if err != nil {
		return false, err
	}

	return k.isFinalized(ctx, bridgeId, output)
}

func (k Keeper) isFinalized(ctx context.Context, bridgeId uint64, output types.Output) (bool, error) {
	bridgeConfig, err := k.GetBridgeConfig(ctx, bridgeId)
	if err != nil {
		return false, err
	}

	return k.isFinalizedWithConfig(ctx, bridgeConfig, output)
}

func (k Keeper) isFinalizedWithConfig(ctx context.Context, bridgeConfig types.BridgeConfig, output types.Output) (bool, error) {
	return sdk.UnwrapSDKContext(ctx).BlockTime().Unix() >= output.L1BlockTime.Add(bridgeConfig.FinalizationPeriod).Unix(), nil
}

////////////////////////////////////
// NextOutputIndex

func (k Keeper) SetNextOutputIndex(ctx context.Context, bridgeId, nextOutputIndex uint64) error {
	return k.NextOutputIndexes.Set(ctx, bridgeId, nextOutputIndex)
}

func (k Keeper) GetNextOutputIndex(ctx context.Context, bridgeId uint64) (uint64, error) {
	nextOutputIndex, err := k.NextOutputIndexes.Get(ctx, bridgeId)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			nextOutputIndex = 1
		} else {
			return 0, err
		}
	}

	return nextOutputIndex, nil
}

func (k Keeper) IncreaseNextOutputIndex(ctx context.Context, bridgeId uint64) (uint64, error) {
	nextOutputIndex, err := k.GetNextOutputIndex(ctx, bridgeId)
	if err != nil {
		return 0, err
	}

	// increase NextOutputIndex
	if err := k.NextOutputIndexes.Set(ctx, bridgeId, nextOutputIndex+1); err != nil {
		return 0, err
	}

	return nextOutputIndex, nil
}
