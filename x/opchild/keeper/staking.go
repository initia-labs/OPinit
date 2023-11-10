package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// fake staking functions

// MaxValidators - Maximum number of validators
func (k Keeper) MaxValidators(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).MaxValidators
}

// HistoricalEntries = number of historical info entries
// to persist in store
func (k Keeper) HistoricalEntries(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).HistoricalEntries
}

// UnbondingTime - The time duration for unbonding
func (k Keeper) UnbondingTime(ctx sdk.Context) time.Duration {
	return unbondingTime
}
