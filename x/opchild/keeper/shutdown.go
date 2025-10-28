package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/OPinit/x/opchild/types"
)

func (k Keeper) IsBridgeDisabled(ctx context.Context) (bool, error) {
	bridgeConfig, err := k.BridgeInfo.Get(ctx)
	if errors.Is(err, collections.ErrNotFound) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return bridgeConfig.BridgeConfig.BridgeDisabled, nil
}

func (k *Keeper) Shutdown(ctx context.Context) error {
	ms := NewMsgServerImpl(k)
	var err error
	k.bankKeeper.IterateAllBalances(ctx, func(addr sdk.AccAddress, coin sdk.Coin) bool {
		_, err := ms.GetBaseDenom(ctx, coin.Denom)
		if err != nil {
			// when the coin is not from the bridge, skip
			return false
		}

		var from, to string
		from, err = k.addressCodec.BytesToString(addr)
		if err != nil {
			return true
		}

		to, err = k.l1AddressCodec.BytesToString(addr)
		if err != nil {
			return true
		}

		_, err = ms.InitiateTokenWithdrawal(ctx, types.NewMsgInitiateTokenWithdrawal(from, to, coin))
		return err != nil
	})
	if err != nil {
		return err
	}

	// clear the validator set
	err = k.Validators.Clear(ctx, nil)
	if err != nil {
		return err
	}
	err = k.ValidatorsByConsAddr.Clear(ctx, nil)
	if err != nil {
		return err
	}
	return nil
}
