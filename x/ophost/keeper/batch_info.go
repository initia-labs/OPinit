package keeper

import (
	"context"

	"cosmossdk.io/collections"

	"github.com/initia-labs/OPinit/x/ophost/types"
)

func (k Keeper) GetNextBatchInfoIndex(ctx context.Context, bridgeId uint64) (batchInfoIndex uint64, err error) {
	if err := k.ReverseIterateBatchInfos(ctx, bridgeId, func(key collections.Pair[uint64, uint64], batchInfo types.BatchInfoWithOutput) (stop bool, err error) {
		batchInfoIndex = key.K2() + 1
		return true, nil
	}); err != nil {
		return 0, err
	}

	return batchInfoIndex, nil
}

func (k Keeper) SetBatchInfo(ctx context.Context, bridgeId uint64, batchInfo types.BatchInfo, output types.Output) error {
	nextIndex, err := k.GetNextBatchInfoIndex(ctx, bridgeId)
	if err != nil {
		return err
	}
	return k.BatchInfos.Set(ctx, collections.Join(bridgeId, nextIndex), types.BatchInfoWithOutput{BatchInfo: batchInfo, Output: output})
}

func (k Keeper) GetAllBatchInfos(ctx context.Context, bridgeId uint64) (batchInfos []types.BatchInfoWithOutput, err error) {
	batchInfos = make([]types.BatchInfoWithOutput, 0)
	if err := k.IterateBatchInfos(ctx, bridgeId, func(key collections.Pair[uint64, uint64], batchInfo types.BatchInfoWithOutput) (stop bool, err error) {
		batchInfos = append(batchInfos, batchInfo)
		return false, nil
	}); err != nil {
		return nil, err
	}

	return batchInfos, nil
}

func (k Keeper) IterateBatchInfos(ctx context.Context, bridgeId uint64, cb func(key collections.Pair[uint64, uint64], batchInfo types.BatchInfoWithOutput) (stop bool, err error)) error {
	return k.BatchInfos.Walk(ctx, collections.NewPrefixedPairRange[uint64, uint64](bridgeId), cb)
}

func (k Keeper) ReverseIterateBatchInfos(ctx context.Context, bridgeId uint64, cb func(key collections.Pair[uint64, uint64], batchInfo types.BatchInfoWithOutput) (stop bool, err error)) error {
	return k.BatchInfos.Walk(ctx, collections.NewPrefixedPairRange[uint64, uint64](bridgeId).Descending(), cb)
}
