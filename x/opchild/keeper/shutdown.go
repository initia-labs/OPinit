package keeper

import (
	"bytes"
	"context"
	"errors"
	"math"
	"strings"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"

	ibcerrors "github.com/cosmos/ibc-go/v8/modules/core/errors"
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
		var acc sdk.AccountI
		addr, err = iter.Key()
		if err != nil {
			return false, err
		}
		acc, err = iter.Value()
		if err != nil {
			return false, err
		}

		if _, ok := acc.(sdk.ModuleAccountI); ok {
			lastAddr = addr
			continue
		}

		k.bankKeeper.IterateAccountBalances(ctx, addr, func(coin sdk.Coin) bool {
			var from, to string
			from, err = k.addressCodec.BytesToString(addr)
			if err != nil {
				return true
			}

			to, err = k.l1AddressCodec.BytesToString(addr)
			if err != nil {
				return true
			}

			// Only withdraw spendable amount for this denom
			spendable := k.bankKeeper.SpendableCoin(ctx, addr, coin.Denom)
			if !spendable.IsPositive() {
				return false
			}

			_, err = ms.GetBaseDenom(ctx, coin.Denom)
			if err == nil {
				_, err = ms.InitiateTokenWithdrawal(ctx, types.NewMsgInitiateTokenWithdrawal(from, to, spendable))
				if err != nil {
					return true
				}

				withdrawCount++
				if withdrawCount == MaxWithdrawCount {
					return true
				}
			} else if errors.Is(err, types.ErrNonL1Token) {
				err = nil
			} else {
				return true
			}

			if strings.HasPrefix(coin.Denom, "ibc/") {
				var fullDenomPath string
				fullDenomPath, err = k.transferKeeper.DenomPathFromHash(sdk.UnwrapSDKContext(ctx), coin.Denom)
				if err != nil {
					return true
				}

				parts := strings.Split(fullDenomPath, "/")
				if len(parts) < 3 {
					return false
				}

				sourcePort := parts[0]
				sourceChannel := parts[1]

				var connection exported.ConnectionI
				_, connection, err = k.channelKeeper.GetChannelConnection(sdk.UnwrapSDKContext(ctx), sourcePort, sourceChannel)
				if err != nil {
					return true
				}

				var l1ClientId string
				l1ClientId, err = k.L1ClientId(ctx)
				if err != nil {
					return true
				}
				// check if the connection is the same as the l1 client id
				if connection.GetClientID() != l1ClientId {
					return false
				}

				transferMsg := transfertypes.NewMsgTransfer(
					sourcePort,
					sourceChannel,
					spendable,
					from,
					to,
					clienttypes.NewHeight(0, 0),
					math.MaxUint64,
					"",
				)

				// handle the transfer message
				sdkCtx := sdk.UnwrapSDKContext(ctx)
				handler := k.msgRouter.Handler(transferMsg)
				if handler == nil {
					err = errorsmod.Wrap(sdkerrors.ErrNotFound, sdk.MsgTypeURL(transferMsg))
					return true
				}
				var res *sdk.Result
				res, err = handler(sdkCtx, transferMsg)
				if errors.Is(err, transfertypes.ErrSendDisabled) || errors.Is(err, ibcerrors.ErrUnauthorized) {
					err = nil
					return false
				} else if err != nil {
					return true
				}
				sdkCtx.EventManager().EmitEvents(res.GetEvents())

				withdrawCount++
				if withdrawCount == MaxWithdrawCount {
					return true
				}
			}
			return false
		})
		if err != nil || withdrawCount == MaxWithdrawCount {
			break
		}
		lastAddr = addr
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
