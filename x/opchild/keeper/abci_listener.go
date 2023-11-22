package keeper

import (
	"context"
	"sync"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	store "github.com/cosmos/cosmos-sdk/store/types"
)

var _ baseapp.StreamingService = &ABCIListener{}

// ABCIListener is the abci listener to check whether current block is empty or not.
type ABCIListener struct {
	txCount uint64
	*Keeper
}

func newABCIListener(k *Keeper) ABCIListener {
	return ABCIListener{txCount: 0, Keeper: k}
}

// ListenDeliverTx updates the steaming service with the latest DeliverTx messages
func (listener *ABCIListener) ListenDeliverTx(ctx context.Context, _ abci.RequestDeliverTx, _ abci.ResponseDeliverTx) error {
	listener.txCount++

	return nil
}

// Stream is the streaming service loop, awaits kv pairs and writes them to some destination stream or file
func (listener *ABCIListener) Stream(wg *sync.WaitGroup) error { return nil }

// Listeners returns the streaming service's listeners for the BaseApp to register
func (listener *ABCIListener) Listeners() map[store.StoreKey][]store.WriteListener { return nil }

// ListenBeginBlock updates the streaming service with the latest BeginBlock messages
func (listener *ABCIListener) ListenBeginBlock(ctx context.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) error {
	listener.txCount = 0

	return nil
}

// ListenEndBlock updates the steaming service with the latest EndBlock messages
func (listener *ABCIListener) ListenEndBlock(ctx context.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) error {

	// if a block is empty, then remove historical info.
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// ignore first tx in a block.
	// - https://github.com/skip-mev/block-sdk/issues/215
	if listener.txCount == 1 && sdkCtx.BlockHeight() != 1 {
		listener.DeleteHistoricalInfo(sdkCtx, sdkCtx.BlockHeight())
	}

	return nil
}

// ListenCommit updates the steaming service with the latest Commit event
func (listener *ABCIListener) ListenCommit(ctx context.Context, res abci.ResponseCommit) error {
	return nil
}

// Closer is the interface that wraps the basic Close method.
func (listener *ABCIListener) Close() error {
	return nil
}
