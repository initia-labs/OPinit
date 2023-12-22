package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"

	"github.com/initia-labs/OPinit/x/ophost/types"
)

////////////////////////////////////
// BridgeConfig

func (k Keeper) SetBridgeConfig(
	ctx context.Context,
	bridgeId uint64,
	bridgeConfig types.BridgeConfig,
) error {
	return k.BridgeConfigs.Set(ctx, bridgeId, bridgeConfig)
}

func (k Keeper) GetBridgeConfig(
	ctx context.Context,
	bridgeId uint64,
) (bridgeConfig types.BridgeConfig, err error) {
	return k.BridgeConfigs.Get(ctx, bridgeId)
}

func (k Keeper) IterateBridgeConfig(
	ctx context.Context,
	cb func(bridgeId uint64, bridgeConfig types.BridgeConfig) (stop bool, err error),
) error {
	return k.BridgeConfigs.Walk(ctx, nil, cb)
}

////////////////////////////////////
// NextL1Sequence

func (k Keeper) SetNextL1Sequence(ctx context.Context, bridgeId, nextL1Sequence uint64) error {
	return k.NextL1Sequences.Set(ctx, bridgeId, nextL1Sequence)
}

func (k Keeper) GetNextL1Sequence(ctx context.Context, bridgeId uint64) (uint64, error) {
	nextSequence, err := k.NextL1Sequences.Get(ctx, bridgeId)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			nextSequence = 1
		} else {
			return 0, err
		}
	}

	return nextSequence, nil
}

func (k Keeper) IncreaseNextL1Sequence(ctx context.Context, bridgeId uint64) (uint64, error) {
	nextL1Sequence, err := k.GetNextL1Sequence(ctx, bridgeId)
	if err != nil {
		return 0, err
	}

	// increase NextL1Sequence
	if err = k.NextL1Sequences.Set(ctx, bridgeId, nextL1Sequence+1); err != nil {
		return 0, err
	}

	return nextL1Sequence, err
}

////////////////////////////////////
// NextBridgeId

func (k Keeper) SetNextBridgeId(ctx context.Context, nextBridgeId uint64) error {
	return k.NextBridgeId.Set(ctx, nextBridgeId)
}

func (k Keeper) GetNextBridgeId(ctx context.Context) (uint64, error) {
	nextBridgeId, err := k.NextBridgeId.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			nextBridgeId = 1
		} else {
			return 0, err
		}
	}

	return nextBridgeId, nil
}

func (k Keeper) IncreaseNextBridgeId(ctx context.Context) (uint64, error) {
	nextBridgeId, err := k.GetNextBridgeId(ctx)
	if err != nil {
		return 0, err
	}

	// increase NextBridgeId
	if err := k.NextBridgeId.Set(ctx, nextBridgeId+1); err != nil {
		return 0, err
	}

	return nextBridgeId, nil
}
