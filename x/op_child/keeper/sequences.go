package keeper

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/initia-labs/OPinit/x/op_child/types"
)

func (k Keeper) RecordFinalizedInboundSequence(ctx sdk.Context, inboundSequence uint64) {
	kvStore := ctx.KVStore(k.storeKey)
	kvStore.Set(types.GetFinalizedInboundSequenceKey(inboundSequence), []byte{1})
}

func (k Keeper) HasFinalizedInboundSequence(ctx sdk.Context, inboundSequence uint64) bool {
	kvStore := ctx.KVStore(k.storeKey)
	bz := kvStore.Get(types.GetFinalizedInboundSequenceKey(inboundSequence))
	return len(bz) != 0
}

func (k Keeper) IncreaseOutboundSequence(ctx sdk.Context) uint64 {
	kvStore := ctx.KVStore(k.storeKey)
	bz := kvStore.Get(types.OutboundSequenceKey)
	if len(bz) == 0 {
		bz := [8]byte{}
		binary.BigEndian.PutUint64(bz[:], 1)
		kvStore.Set(types.OutboundSequenceKey, bz[:])
		return 1
	}

	return binary.BigEndian.Uint64(bz)
}
