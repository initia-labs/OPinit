package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/OPinit/x/opchild/types"
)

// BridgeExecutor returns params.BridgeExecutor
func (k Keeper) BridgeExecutor(ctx context.Context) (sdk.AccAddress, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	return k.authKeeper.AddressCodec().StringToBytes(params.BridgeExecutor)
}

// SetParams sets the x/opchild module parameters.
func (k Keeper) SetParams(ctx context.Context, params types.Params) error {
	if err := params.Validate(k.authKeeper.AddressCodec()); err != nil {
		return err
	}
	return k.Params.Set(ctx, params)
}

// GetParams sets the x/opchild module parameters.
func (k Keeper) GetParams(ctx context.Context) (params types.Params, err error) {
	return k.Params.Get(ctx)
}

func (k Keeper) MinGasPrices(ctx context.Context) (sdk.DecCoins, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	return params.MinGasPrices, nil
}
