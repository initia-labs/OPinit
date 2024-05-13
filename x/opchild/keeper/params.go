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

// FeeWhitelist returns params.FeeWhitelist
func (k Keeper) FeeWhitelist(ctx context.Context) ([]string, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	return params.FeeWhitelist, nil
}

// SetParams sets the x/opchild module parameters.
func (k Keeper) SetParams(ctx context.Context, params types.Params) error {
	if err := params.Validate(k.authKeeper.AddressCodec()); err != nil {
		return err
	}

	allValidators, err := k.GetAllValidators(ctx)
	if err != nil {
		return err
	}

	if int(params.MaxValidators) < len(allValidators) {
		return types.ErrMaxValidatorsLowerThanCurrent
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
