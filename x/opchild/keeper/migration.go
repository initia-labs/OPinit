package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"

	"github.com/initia-labs/OPinit/x/opchild/types"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
)

// SetMigrationInfo sets the migration info
func (k Keeper) SetMigrationInfo(ctx context.Context, migrationInfo types.MigrationInfo) error {
	return k.MigrationInfos.Set(ctx, migrationInfo.Denom, migrationInfo)
}

// GetMigrationInfo gets the migration info
func (k Keeper) GetMigrationInfo(ctx context.Context, denom string) (types.MigrationInfo, error) {
	return k.MigrationInfos.Get(ctx, denom)
}

// HasMigrationInfo checks if the migration info is registered
func (k Keeper) HasMigrationInfo(ctx context.Context, denom string) (bool, error) {
	return k.MigrationInfos.Has(ctx, denom)
}

// IterateMigrationInfos iterates over the migration infos
func (k Keeper) IterateMigrationInfos(ctx context.Context, cb func(denom string, migrationInfo types.MigrationInfo) (stop bool, err error)) error {
	return k.MigrationInfos.Walk(ctx, nil, func(denom string, migrationInfo types.MigrationInfo) (stop bool, err error) {
		if stop, err := cb(denom, migrationInfo); err != nil {
			return true, err
		} else if stop {
			return true, nil
		}
		return false, nil
	})
}

// SetIBCToL2DenomMap sets the ibc to l2 denom map
func (k Keeper) SetIBCToL2DenomMap(ctx context.Context, ibcDenom, l2Denom string) error {
	return k.IBCToL2DenomMap.Set(ctx, ibcDenom, l2Denom)
}

// GetIBCToL2DenomMap gets the ibc to l2 denom map
func (k Keeper) GetIBCToL2DenomMap(ctx context.Context, ibcDenom string) (string, error) {
	return k.IBCToL2DenomMap.Get(ctx, ibcDenom)
}

// HasIBCToL2DenomMap checks if the ibc to l2 denom map is registered
func (k Keeper) HasIBCToL2DenomMap(ctx context.Context, ibcDenom string) (bool, error) {
	return k.IBCToL2DenomMap.Has(ctx, ibcDenom)
}

// MigrateToken implements migrating a token from the OP token to the IBC token
func (k Keeper) MigrateToken(ctx context.Context, migrationInfo types.MigrationInfo, sender sdk.AccAddress, amount sdk.Coin) (sdk.Coin, error) {
	// check if the amount is positive
	if !amount.IsPositive() {
		return sdk.Coin{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "amount is not positive")
	}

	if migrationInfo.Denom != amount.Denom {
		return sdk.Coin{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "migration info denom does not match")
	}

	// send coins to the module
	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, sdk.NewCoins(amount))
	if err != nil {
		return sdk.Coin{}, err
	}

	// burn coins from the module
	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(amount))
	if err != nil {
		return sdk.Coin{}, err
	}

	// mint IBC token to the module
	baseDenom, err := k.GetBaseDenom(ctx, amount.Denom)
	if err != nil {
		return sdk.Coin{}, err
	}
	ibcCoin := sdk.NewCoin(ibcDenom(migrationInfo, baseDenom), amount.Amount)
	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(ibcCoin))
	if err != nil {
		return sdk.Coin{}, err
	}

	// send IBC token to the sender
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sender, sdk.NewCoins(ibcCoin))
	if err != nil {
		return sdk.Coin{}, err
	}

	return ibcCoin, nil
}

// HandleMigratedTokenDeposit implements handling a migrated token deposit: convert IBC token to L2 token
func (k Keeper) HandleMigratedTokenDeposit(ctx context.Context, sender sdk.AccAddress, ibcCoin sdk.Coin, memo string) (sdk.Coin, error) {
	// check if the amount is positive
	if !ibcCoin.IsPositive() {
		return sdk.Coin{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "amount is not positive")
	}

	// compute l2Denom
	l2Denom, err := k.GetIBCToL2DenomMap(ctx, ibcCoin.Denom)
	if err != nil {
		return sdk.Coin{}, err
	}

	// send IBC token to the module
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, sdk.NewCoins(ibcCoin))
	if err != nil {
		return sdk.Coin{}, err
	}

	// burn IBC token
	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(ibcCoin))
	if err != nil {
		return sdk.Coin{}, err
	}

	// mint L2 token
	l2Coin := sdk.NewCoin(l2Denom, ibcCoin.Amount)
	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(l2Coin))
	if err != nil {
		return sdk.Coin{}, err
	}

	// send L2 token to the sender
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sender, sdk.NewCoins(l2Coin))
	if err != nil {
		return sdk.Coin{}, err
	}

	// handle deposit hook if exists
	var migratedTokenDepositMemo ophosttypes.MigratedTokenDepositMemo
	decoder := json.NewDecoder(strings.NewReader(memo))
	decoder.DisallowUnknownFields()

	// if the memo is not valid, return the L2 token
	if err := decoder.Decode(&migratedTokenDepositMemo); err != nil {
		return l2Coin, nil
	}

	// if the OPinit is not empty, handle the deposit hook
	if len(migratedTokenDepositMemo.OPinit) > 0 {
		params, err := k.GetParams(ctx)
		if err != nil {
			return sdk.Coin{}, err
		}

		if success, reason := k.handleBridgeHook(sdk.UnwrapSDKContext(ctx), migratedTokenDepositMemo.OPinit, params.HookMaxGas); !success {
			return sdk.Coin{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, reason)
		}
	}

	return l2Coin, nil
}

// HandleMigratedTokenWithdrawal implements handling a migrated token withdrawal
func (k Keeper) HandleMigratedTokenWithdrawal(ctx context.Context, msg *types.MsgInitiateTokenWithdrawal) (handled bool, err error) {
	denom := msg.Amount.Denom
	migrationInfo, err := k.GetMigrationInfo(ctx, denom)
	if err != nil && errors.Is(err, collections.ErrNotFound) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	sender, err := k.addressCodec.StringToBytes(msg.Sender)
	if err != nil {
		return false, err
	}

	// migrate the token from OP token to IBC token
	ibcCoin, err := k.MigrateToken(ctx, migrationInfo, sender, msg.Amount)
	if err != nil {
		return false, err
	}

	// create a transfer message to the destination chain
	transferMsg := transfertypes.NewMsgTransfer(
		migrationInfo.IbcPortId,
		migrationInfo.IbcChannelId,
		ibcCoin,
		msg.Sender,
		msg.To,
		clienttypes.NewHeight(0, 0),
		// use default timeout 10 minutes
		uint64(sdk.UnwrapSDKContext(ctx).BlockTime().UnixNano())+transfertypes.DefaultRelativePacketTimeoutTimestamp,
		"forwarded from opchild module",
	)

	// handle the transfer message
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if handler := k.msgRouter.Handler(transferMsg); handler == nil {
		return false, errorsmod.Wrap(sdkerrors.ErrNotFound, sdk.MsgTypeURL(transferMsg))
	} else if res, err := handler(sdkCtx, transferMsg); err != nil {
		return false, err
	} else {
		sdkCtx.EventManager().EmitEvents(res.GetEvents())
	}

	return true, nil
}

// ibcDenom computes the IBC denom from the migration info and base denom
func ibcDenom(migrationInfo types.MigrationInfo, baseDenom string) string {
	return transfertypes.ParseDenomTrace(fmt.Sprintf("%s/%s/%s", migrationInfo.IbcPortId, migrationInfo.IbcChannelId, baseDenom)).IBCDenom()
}
