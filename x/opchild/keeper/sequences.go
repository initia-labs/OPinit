package keeper

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/initia-labs/OPinit/x/opchild/types"
)

func (k Keeper) RecordFinalizedL1Sequence(ctx sdk.Context, l1Sequence uint64) {
	kvStore := ctx.KVStore(k.storeKey)
	kvStore.Set(types.GetFinalizedL1SequenceKey(l1Sequence), []byte{1})
}

func (k Keeper) HasFinalizedL1Sequence(ctx sdk.Context, l1Sequence uint64) bool {
	kvStore := ctx.KVStore(k.storeKey)
	bz := kvStore.Get(types.GetFinalizedL1SequenceKey(l1Sequence))
	return len(bz) != 0
}

func (k Keeper) IterateFinalizedL1Sequences(ctx sdk.Context, cb func(l1Sequence uint64) bool) {
	kvStore := ctx.KVStore(k.storeKey)

	prefixStore := prefix.NewStore(kvStore, types.FinalizedL1SequenceKey)
	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		l1Sequence := binary.BigEndian.Uint64(iterator.Key())

		if cb(l1Sequence) {
			break
		}
	}
}

func (k Keeper) SetNextL2Sequence(ctx sdk.Context, l2Sequence uint64) {
	kvStore := ctx.KVStore(k.storeKey)

	_l2Sequence := [8]byte{}
	binary.BigEndian.PutUint64(_l2Sequence[:], l2Sequence)
	kvStore.Set(types.NextL2SequenceKey, _l2Sequence[:])
}

func (k Keeper) GetNextL2Sequence(ctx sdk.Context) uint64 {
	kvStore := ctx.KVStore(k.storeKey)

	bz := kvStore.Get(types.NextL2SequenceKey)
	if len(bz) == 0 {
		return 1
	}

	return binary.BigEndian.Uint64(bz)
}

func (k Keeper) IncreaseNextL2Sequence(ctx sdk.Context) uint64 {
	kvStore := ctx.KVStore(k.storeKey)
	bz := kvStore.Get(types.NextL2SequenceKey)
	sequence := uint64(1)
	if len(bz) > 0 {
		sequence = binary.BigEndian.Uint64(bz)
	}

	k.SetNextL2Sequence(ctx, sequence+1)

	return sequence
}
