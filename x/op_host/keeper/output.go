package keeper

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/initia-labs/OPinit/x/op_host/types"
)

////////////////////////////////////
// OutputProposal

func (k Keeper) SetOutputProposal(ctx sdk.Context, bridgeId, outputIndex uint64, outputProposal types.Output) error {
	bz, err := k.cdc.Marshal(&outputProposal)
	if err != nil {
		return err
	}

	kvStore := ctx.KVStore(k.storeKey)
	kvStore.Set(types.GetOutputProposalKey(bridgeId, outputIndex), bz)

	return nil
}

func (k Keeper) GetOutputProposal(ctx sdk.Context, bridgeId, outputIndex uint64) (outputProposal types.Output, err error) {
	kvStore := ctx.KVStore(k.storeKey)
	bz := kvStore.Get(types.GetOutputProposalKey(bridgeId, outputIndex))
	if len(bz) == 0 {
		err = errors.ErrKeyNotFound.Wrap("failed to fetch output_proposal")
		return
	}

	err = k.cdc.Unmarshal(bz, &outputProposal)
	return
}

func (k Keeper) DeleteOutputProposal(ctx sdk.Context, bridgeId, outputIndex uint64) {
	kvStore := ctx.KVStore(k.storeKey)
	kvStore.Delete(types.GetOutputProposalKey(bridgeId, outputIndex))
}

func (k Keeper) IterateOutputProposals(ctx sdk.Context, bridgeId uint64, cb func(bridgeId, outputIndex uint64, output types.Output) bool) error {
	kvStore := ctx.KVStore(k.storeKey)

	prefixStore := prefix.NewStore(kvStore, types.GetOutputProposalBridgePrefixKey(bridgeId))
	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		val := iterator.Value()

		outputIndex := binary.BigEndian.Uint64(key)

		var output types.Output
		if err := k.cdc.Unmarshal(val, &output); err != nil {
			return err
		}

		if cb(bridgeId, outputIndex, output) {
			break
		}
	}

	return nil
}

func (k Keeper) IsFinalized(ctx sdk.Context, bridgeId, outputIndex uint64) (bool, error) {
	output, err := k.GetOutputProposal(ctx, bridgeId, outputIndex)
	if err != nil {
		return false, err
	}

	bridgeConfig, err := k.GetBridgeConfig(ctx, bridgeId)
	if err != nil {
		return false, err
	}

	return ctx.BlockTime().Unix() >= output.L1BlockTime.Add(bridgeConfig.FinalizationPeriod).Unix(), nil
}

////////////////////////////////////
// NextOutputIndex

func (k Keeper) SetNextOutputIndex(ctx sdk.Context, bridgeId, nextOutputIndex uint64) {
	_nextOutputIndex := [8]byte{}
	binary.BigEndian.PutUint64(_nextOutputIndex[:], nextOutputIndex)

	kvStore := ctx.KVStore(k.storeKey)
	kvStore.Set(types.GetNextOutputIndexKey(bridgeId), _nextOutputIndex[:])
}

func (k Keeper) GetNextOutputIndex(ctx sdk.Context, bridgeId uint64) uint64 {
	kvStore := ctx.KVStore(k.storeKey)
	bz := kvStore.Get(types.GetNextOutputIndexKey(bridgeId))
	if len(bz) == 0 {
		return 1
	}

	return binary.BigEndian.Uint64(bz)
}

func (k Keeper) IncreaseNextOutputIndex(ctx sdk.Context, bridgeId uint64) uint64 {
	kvStore := ctx.KVStore(k.storeKey)

	// load next output index
	key := types.GetNextOutputIndexKey(bridgeId)
	bz := kvStore.Get(key)

	nextOutputIndex := uint64(1)
	if len(bz) != 0 {
		nextOutputIndex = binary.BigEndian.Uint64(bz)
	}

	// increase next output index
	_nextOutputIndex := [8]byte{}
	binary.BigEndian.PutUint64(_nextOutputIndex[:], nextOutputIndex+1)
	kvStore.Set(key, _nextOutputIndex[:])

	return nextOutputIndex
}
