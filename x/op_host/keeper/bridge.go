package keeper

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/initia-labs/OPinit/x/op_host/types"
)

////////////////////////////////////
// BridgeConfig

func (k Keeper) SetBridgeConfig(
	ctx sdk.Context,
	bridgeId uint64,
	bridgeConfig types.BridgeConfig,
) error {
	bz, err := k.cdc.Marshal(&bridgeConfig)
	if err != nil {
		return err
	}

	kvStore := ctx.KVStore(k.storeKey)
	kvStore.Set(types.GetBridgeConfigKey(bridgeId), bz)
	return nil
}

func (k Keeper) GetBridgeConfig(
	ctx sdk.Context,
	bridgeId uint64,
) (bridgeConfig types.BridgeConfig, err error) {
	kvStore := ctx.KVStore(k.storeKey)
	bz := kvStore.Get(types.GetBridgeConfigKey(bridgeId))

	if len(bz) == 0 {
		err = errors.ErrKeyNotFound
		return
	}

	err = k.cdc.Unmarshal(bz, &bridgeConfig)
	return
}

func (k Keeper) IterateBridgeConfig(
	ctx sdk.Context,
	cb func(bridgeId uint64, bridgeConfig types.BridgeConfig) bool,
) error {
	kvStore := ctx.KVStore(k.storeKey)

	prefixStore := prefix.NewStore(kvStore, types.BridgeConfigKey)
	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		val := iterator.Value()

		bridgeId := binary.BigEndian.Uint64(key)

		var bridgeConfig types.BridgeConfig
		if err := k.cdc.Unmarshal(val, &bridgeConfig); err != nil {
			return err
		}

		if cb(bridgeId, bridgeConfig) {
			break
		}
	}

	return nil
}

////////////////////////////////////
// NextL1Sequence

func (k Keeper) SetNextL1Sequence(ctx sdk.Context, bridgeId, nextL1Sequence uint64) {
	_nextL1Sequence := [8]byte{}
	binary.BigEndian.PutUint64(_nextL1Sequence[:], nextL1Sequence)

	kvStore := ctx.KVStore(k.storeKey)
	kvStore.Set(types.GetNextL1SequenceKey(bridgeId), _nextL1Sequence[:])
}

func (k Keeper) GetNextL1Sequence(ctx sdk.Context, bridgeId uint64) uint64 {
	kvStore := ctx.KVStore(k.storeKey)
	bz := kvStore.Get(types.GetNextL1SequenceKey(bridgeId))
	if len(bz) == 0 {
		return 1
	}

	return binary.BigEndian.Uint64(bz)
}

func (k Keeper) IncreaseNextL1Sequence(ctx sdk.Context, bridgeId uint64) uint64 {
	kvStore := ctx.KVStore(k.storeKey)

	// load next bridge sequence
	key := types.GetNextL1SequenceKey(bridgeId)
	bz := kvStore.Get(key)

	nextL1Sequence := uint64(1)
	if len(bz) != 0 {
		nextL1Sequence = binary.BigEndian.Uint64(bz)
	}

	// increase next bridge sequence
	_nextL1Sequence := [8]byte{}
	binary.BigEndian.PutUint64(_nextL1Sequence[:], nextL1Sequence+1)
	kvStore.Set(key, _nextL1Sequence[:])

	return nextL1Sequence
}

////////////////////////////////////
// NextBridgeId

func (k Keeper) SetNextBridgeId(ctx sdk.Context, nextBridgeId uint64) {
	_nextBridgeId := [8]byte{}
	binary.BigEndian.PutUint64(_nextBridgeId[:], nextBridgeId)

	kvStore := ctx.KVStore(k.storeKey)
	kvStore.Set(types.NextBridgeIdKey, _nextBridgeId[:])
}

func (k Keeper) GetNextBridgeId(ctx sdk.Context) uint64 {
	kvStore := ctx.KVStore(k.storeKey)
	bz := kvStore.Get(types.NextBridgeIdKey)
	if len(bz) == 0 {
		return 1
	}

	return binary.BigEndian.Uint64(bz)
}

func (k Keeper) IncreaseNextBridgeId(ctx sdk.Context) uint64 {
	kvStore := ctx.KVStore(k.storeKey)

	// load next bridge id
	key := types.NextBridgeIdKey
	bz := kvStore.Get(key)

	nextBridgeId := uint64(1)
	if len(bz) != 0 {
		nextBridgeId = binary.BigEndian.Uint64(bz)
	}

	// increase next bridge id
	_nextBridgeId := [8]byte{}
	binary.BigEndian.PutUint64(_nextBridgeId[:], nextBridgeId+1)
	kvStore.Set(key, _nextBridgeId[:])

	return nextBridgeId
}
