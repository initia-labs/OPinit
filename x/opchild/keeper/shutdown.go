package keeper

import (
	"bytes"
	"context"
	"errors"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

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
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return false, err
	}

	// get the auth keeper
	authKeeper, ok := k.authKeeper.(*authkeeper.AccountKeeper)
	if !ok {
		return false, errors.New("unexpected auth keeper type")
	}

	// iterate all accounts from the last processed address
	iter, err := authKeeper.Accounts.Iterate(ctx, new(collections.Range[sdk.AccAddress]).StartExclusive(shutdownInfo))
	if err != nil {
		return false, err
	}
	defer iter.Close()

	var lastAddr sdk.AccAddress
	for ; iter.Valid(); iter.Next() {
		var addr sdk.AccAddress
		addr, err = iter.Key()
		if err != nil {
			return false, err
		}
		acc, err := iter.Value()
		if err != nil {
			return false, err
		}

		if _, ok := acc.(sdk.ModuleAccountI); ok {
			lastAddr = addr
			continue
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
			if withdrawCount == MaxWithdrawCount {
				return true
			}
			lastAddr = addr
			return false
		})
		if err != nil || withdrawCount == MaxWithdrawCount {
			break
		}
	}

	if err != nil {
		return false, err
	} else if !bytes.Equal(shutdownInfo, lastAddr.Bytes()) {
		err := k.ShutdownInfo.Set(ctx, lastAddr.Bytes())
		if err != nil {
			return false, err
		}
	}

	if withdrawCount == MaxWithdrawCount {
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
