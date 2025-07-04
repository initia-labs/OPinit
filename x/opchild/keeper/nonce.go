package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// DefaultFastBridgeNonceStart is the starting nonce for fast bridge operations
	DefaultFastBridgeNonceStart = uint64(1)
)

// SetFastBridgeNonce sets the fast bridge nonce for a specific account
func (k Keeper) SetFastBridgeNonce(ctx context.Context, addr sdk.AccAddress, nonce uint64) error {
	return k.NextFastBridgeNonce.Set(ctx, addr, nonce)
}

// GetFastBridgeNonce returns the current fast bridge nonce for a specific account
func (k Keeper) GetFastBridgeNonce(ctx context.Context, addr sdk.AccAddress) (uint64, error) {
	nonce, err := k.NextFastBridgeNonce.Get(ctx, addr)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return DefaultFastBridgeNonceStart, nil
		}
		return 0, err
	}

	return nonce, nil
}

// IncreaseFastBridgeNonce increases the fast bridge nonce for a specific account and returns the current value
func (k Keeper) IncreaseFastBridgeNonce(ctx context.Context, addr sdk.AccAddress) (uint64, error) {
	currentNonce, err := k.GetFastBridgeNonce(ctx, addr)
	if err != nil {
		return 0, err
	}

	if err := k.SetFastBridgeNonce(ctx, addr, currentNonce+1); err != nil {
		return 0, err
	}

	return currentNonce, nil
}
