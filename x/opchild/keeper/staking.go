package keeper

import (
	"context"
	"time"
)

// fake staking functions

// MaxValidators - Maximum number of validators
func (k Keeper) MaxValidators(ctx context.Context) (uint32, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return 0, err
	}

	return params.MaxValidators, nil
}

// HistoricalEntries = number of historical info entries
// to persist in store
func (k Keeper) HistoricalEntries(ctx context.Context) (uint32, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return 0, err
	}

	return params.HistoricalEntries, nil
}

// UnbondingTime - The time duration for unbonding
func (k Keeper) UnbondingTime(ctx context.Context) time.Duration {
	return unbondingTime
}
