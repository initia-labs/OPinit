package keeper

import (
	"context"
	"errors"

	errorsmod "cosmossdk.io/errors"
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

	// compute the hash of all oracle prices
	hash, err := k.computeOraclePricesHash(sdkCtx)
	if err != nil {
		return errorsmod.Wrap(err, "failed to compute oracle prices hash")
	}

	oraclePriceHash := types.OraclePriceHash{
		Hash:          hash,
		L1BlockHeight: uint64(sdkCtx.BlockHeight()), //nolint:gosec
		L1BlockTime:   sdkCtx.BlockTime().UnixNano(),
	}

	if err := k.OraclePriceHash.Set(ctx, oraclePriceHash); err != nil {
		return errorsmod.Wrap(err, "failed to store oracle price hash")
	}

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
			CurrencyPairId:     id,
			CurrencyPairString: cp.String(),
			Price:              price.Price,
			Timestamp:          price.BlockTimestamp.UnixNano(),
		})
	}

	return prices.ComputeOraclePricesHash(), nil
}

// GetOraclePriceHash returns the oracle price hash for a bridge.
func (k Keeper) GetOraclePriceHash(ctx context.Context) (types.OraclePriceHash, error) {
	return k.OraclePriceHash.Get(ctx)
}
