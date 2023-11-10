package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/OPinit/x/opchild/types"
)

// BridgeExecutor returns params.BridgeExecutor
func (k Keeper) BridgeExecutor(ctx sdk.Context) sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(k.GetParams(ctx).BridgeExecutor)
}

// SetParams sets the x/opchild module parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	store.Set(types.ParamsKey, bz)

	return nil
}

// GetParams sets the x/opchild module parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return params
	}

	k.cdc.MustUnmarshal(bz, &params)
	return params
}

func (k Keeper) MinGasPrices(ctx sdk.Context) sdk.DecCoins {
	return k.GetParams(ctx).MinGasPrices
}
