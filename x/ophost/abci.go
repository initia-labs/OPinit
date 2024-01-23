package ophost

import (
	"context"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/OPinit/x/ophost/types"
)

func BeginBlocker(ctx context.Context, k types.OracleKeeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	if k == nil {
		return nil
	}

	// emit oracle prices for bridge executors

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	pairs := k.GetAllCurrencyPairs(sdkCtx)
	events := make([]sdk.Event, 0, len(pairs))
	for _, pair := range pairs {
		price, err := k.GetPriceForCurrencyPair(sdkCtx, pair)
		if err != nil {
			return err
		}

		events = append(events, sdk.NewEvent(
			types.EventTypeOraclePrice,
			sdk.NewAttribute(types.AttributeKeyBase, pair.Base),
			sdk.NewAttribute(types.AttributeKeyQuote, pair.Quote),
			sdk.NewAttribute(types.AttributeKeyPrice, price.Price.String()),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, strconv.FormatUint(price.BlockHeight, 10)),
			sdk.NewAttribute(types.AttributeKeyBlockTime, price.BlockTimestamp.String()),
		))
	}

	sdkCtx.EventManager().EmitEvents(events)

	return nil
}
