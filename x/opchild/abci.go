package opchild

import (
	"context"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/telemetry"

	"github.com/initia-labs/OPinit/x/opchild/keeper"
	"github.com/initia-labs/OPinit/x/opchild/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker will persist the current header and validator set as a historical entry
// and prune the oldest entry based on the HistoricalEntries parameter
func BeginBlocker(ctx context.Context, k *keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	return k.TrackHistoricalInfo(ctx)
}

// Called every block, update validator set
func EndBlocker(ctx context.Context, k *keeper.Keeper) ([]abci.ValidatorUpdate, error) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	height := sdkCtx.BlockHeight()

	if plan, found := k.ExecutorChangePlans[uint64(height)]; found {
		err := k.ChangeExecutor(ctx, plan)
		if err != nil {
			return nil, err
		}
	}

	return k.BlockValidatorUpdates(ctx)
}
