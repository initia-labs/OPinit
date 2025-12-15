package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/initia-labs/OPinit/x/ophost/types"
)

// UpdateOraclePriceHashes computes and stores the hash of all oracle prices for each bridge.
// This should be called in EndBlocker to update the oracle price hash for batched relaying.
func (k Keeper) UpdateOraclePriceHashes(ctx context.Context) error {
	if k.oracleKeeper == nil {
		return nil
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// iterate through all bridges
	err := k.IterateBridgeConfig(ctx, func(bridgeId uint64, config types.BridgeConfig) (bool, error) {
		// skip here if oracle is not enabled for the bridge
		if !config.OracleEnabled {
			return false, nil
		}

		// compute the hash of all oracle prices
		hash, err := k.computeOraclePricesHash(sdkCtx)
		if err != nil {
			k.Logger(ctx).Error("failed to compute oracle prices hash",
				"bridge_id", bridgeId,
				"error", err.Error(),
			)
			// continue to the next bridge instead of failing
			return false, nil
		}

		oraclePriceHash := types.OraclePriceHash{
			Hash:          hash,
			L1BlockHeight: uint64(sdkCtx.BlockHeight()), //nolint:gosec
			L1BlockTime:   sdkCtx.BlockTime().UnixNano(),
		}

		if err := k.OraclePriceHashes.Set(ctx, bridgeId, oraclePriceHash); err != nil {
			k.Logger(ctx).Error("failed to store oracle price hash",
				"bridge_id", bridgeId,
				"error", err.Error(),
			)
			return false, nil
		}

		return false, nil
	})

	return err
}

// computeOraclePricesHash computes a deterministic hash of all current oracle prices.
func (k Keeper) computeOraclePricesHash(ctx sdk.Context) ([]byte, error) {
	numPairs, err := k.oracleKeeper.GetNumCurrencyPairs(ctx)
	if err != nil {
		return nil, err
	}
	if numPairs == 0 {
		return nil, errors.New("no currency pairs found")
	}

	prices := make(types.OraclePriceInfos, 0, numPairs)

	// iterate through all currency pair IDs
	for id := uint64(0); id < numPairs; id++ {
		cp, found := k.oracleKeeper.GetCurrencyPairFromID(ctx, id)
		if !found {
			continue
		}

		price, err := k.oracleKeeper.GetPriceForCurrencyPair(ctx, cp)
		if err != nil {
			continue
		}

		prices = append(prices, types.OraclePriceInfo{
			CurrencyPairId: id,
			Price:          price.Price,
			Timestamp:      price.BlockTimestamp.UnixNano(),
		})
	}

	if len(prices) == 0 {
		return nil, errors.New("no valid currency pairs with prices found")
	}

	return prices.ComputeOraclePricesHash(), nil
}

// GetOraclePriceHash returns the oracle price hash for a bridge.
func (k Keeper) GetOraclePriceHash(ctx context.Context, bridgeId uint64) (types.OraclePriceHash, error) {
	return k.OraclePriceHashes.Get(ctx, bridgeId)
}

// HasOraclePriceHash checks if an oracle price hash exists for a bridge.
func (k Keeper) HasOraclePriceHash(ctx context.Context, bridgeId uint64) (bool, error) {
	hash, err := k.OraclePriceHashes.Get(ctx, bridgeId)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return len(hash.Hash) > 0, nil
}
