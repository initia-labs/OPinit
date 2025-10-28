package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/OPinit/x/opchild/types"
)

var dummyValAddr = sdk.ValAddress(make([]byte, 20))
var dummyPubKey = &ed25519.PubKey{Key: make([]byte, ed25519.PubKeySize)}

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
		// Skip module accounts - they can't sign L1 transactions
		if acc := k.authKeeper.GetAccount(ctx, addr); acc != nil {
			if _, ok := acc.(sdk.ModuleAccountI); ok {
				return false
			}
		}

		_, err = ms.GetBaseDenom(ctx, coin.Denom)
		if err != nil {
			// when the coin is not from the bridge, skip
			err = nil
			return false
		}

		// Only withdraw spendable amount for this denom
		spendable := k.bankKeeper.SpendableCoins(ctx, addr).AmountOf(coin.Denom)
		if !spendable.IsPositive() {
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

		_, err = ms.InitiateTokenWithdrawal(ctx, types.NewMsgInitiateTokenWithdrawal(from, to, sdk.NewCoin(coin.Denom, spendable)))
		return err != nil
	})
	if err != nil {
		return err
	}

	dummyVal, err := types.NewValidator(dummyValAddr, dummyPubKey, "")
	if err != nil {
		return err
	}
	return k.ChangeExecutor(ctx, types.ExecutorChangePlan{
		NextValidator: dummyVal,
	})
}
