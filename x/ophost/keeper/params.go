package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/OPinit/x/ophost/types"
)

// RegistrationFee returns params.RegistrationFee
func (k Keeper) RegistrationFee(ctx context.Context) sdk.Coins {
	return k.GetParams(ctx).RegistrationFee
}

// SetParams sets the x/opchild module parameters.
func (k Keeper) SetParams(ctx context.Context, params types.Params) error {
	return k.Params.Set(ctx, params)
}

// GetParams sets the x/opchild module parameters.
func (k Keeper) GetParams(ctx context.Context) (params types.Params) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		panic(err)
	}

	return params
}
