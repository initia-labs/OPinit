package keeper

import (
	"bytes"
	"context"
	"errors"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/OPinit/x/opchild/types"
)

var dummyValAddr = sdk.ValAddress(make([]byte, 20))
var dummyPubKey = &ed25519.PubKey{Key: make([]byte, ed25519.PubKeySize)}

const MaxWithdrawCount = 100

func (k Keeper) IsBridgeDisabled(ctx context.Context) (bool, error) {
	bridgeConfig, err := k.BridgeInfo.Get(ctx)
	if errors.Is(err, collections.ErrNotFound) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return bridgeConfig.BridgeConfig.BridgeDisabled, nil
}

func (k *Keeper) Shutdown(ctx context.Context) (bool, error) {
	ms := NewMsgServerImpl(k)
	withdrawCount := 0

	shutdownInfo, err := k.ShutdownInfo.Get(ctx)
	if errors.Is(err, collections.ErrNotFound) {
		err = nil
	} else if err != nil {
		return false, err
	}
	var lastAddr sdk.AccAddress

	k.authKeeper.IterateAccounts(ctx, func(acc sdk.AccountI) bool {
		addr := acc.GetAddress()

		defer func() {
			lastAddr = addr
		}()

		if bytes.Compare(shutdownInfo, addr.Bytes()) >= 0 {
			return false
		} else if _, ok := acc.(sdk.ModuleAccountI); ok {
			return false
		}

		k.bankKeeper.IterateAccountBalances(ctx, addr, func(coin sdk.Coin) bool {
			_, err = ms.GetBaseDenom(ctx, coin.Denom)
			if errors.Is(err, types.ErrNonL1Token) {
				// when the coin is not from the bridge, skip
				err = nil
				return false
			} else if err != nil {
				return true
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
			if err != nil {
				return true
			}

			withdrawCount++
			return withdrawCount == MaxWithdrawCount
		})
		if err != nil || withdrawCount == MaxWithdrawCount {
			return true
		}
		return false
	})

	if err != nil {
		return false, err
	} else if withdrawCount == MaxWithdrawCount {
		err := k.ShutdownInfo.Set(ctx, lastAddr.Bytes())
		if err != nil {
			return false, err
		}
		return false, nil
	}

	dummyVal, err := types.NewValidator(dummyValAddr, dummyPubKey, "")
	if err != nil {
		return false, err
	}
	return true, k.ChangeExecutor(ctx, types.ExecutorChangePlan{
		NextValidator: dummyVal,
	})
}
