package keeper

import (
	"context"
	"encoding/json"
	"errors"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"

	"github.com/initia-labs/OPinit/x/ophost/types"
)

// SetMigrationInfo sets the migration info
func (k Keeper) SetMigrationInfo(ctx context.Context, migrationInfo types.MigrationInfo) error {
	return k.MigrationInfos.Set(ctx, collections.Join(migrationInfo.BridgeId, migrationInfo.L1Denom), migrationInfo)
}

// GetMigrationInfo gets the migration info
func (k Keeper) GetMigrationInfo(ctx context.Context, bridgeId uint64, l1Denom string) (types.MigrationInfo, error) {
	migrationInfo, err := k.MigrationInfos.Get(ctx, collections.Join(bridgeId, l1Denom))
	if err != nil {
		return types.MigrationInfo{}, err
	}

	return migrationInfo, nil
}

// HasMigrationInfo checks if the migration info is registered
func (k Keeper) HasMigrationInfo(ctx context.Context, bridgeId uint64, l1Denom string) (bool, error) {
	return k.MigrationInfos.Has(ctx, collections.Join(bridgeId, l1Denom))
}

// IterateMigrationInfos iterates over the migration infos
func (k Keeper) IterateMigrationInfos(ctx context.Context, cb func(key collections.Pair[uint64, string], migrationInfo types.MigrationInfo) (stop bool, err error)) error {
	return k.MigrationInfos.Walk(ctx, nil, func(key collections.Pair[uint64, string], migrationInfo types.MigrationInfo) (stop bool, err error) {
		if stop, err := cb(key, migrationInfo); err != nil {
			return true, err
		} else if stop {
			return true, nil
		}
		return false, nil
	})
}

// HandleMigratedTokenDeposit handles the migrated token deposit by forwarding it to ibc transfer module
func (k Keeper) HandleMigratedTokenDeposit(ctx context.Context, msg *types.MsgInitiateTokenDeposit) (handled bool, err error) {
	l1Denom := msg.Amount.Denom
	migrationInfo, err := k.GetMigrationInfo(ctx, msg.BridgeId, l1Denom)
	if err != nil && errors.Is(err, collections.ErrNotFound) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	memo := "forwarded from ophost module"
	if len(msg.Data) > 0 {
		memoBz, err := json.Marshal(&types.MigratedTokenDepositMemo{
			OPinit: msg.Data,
		})
		if err != nil {
			return false, err
		}

		memo = string(memoBz)
	}

	// create a transfer message to the destination chain
	transferMsg := transfertypes.NewMsgTransfer(
		migrationInfo.IbcPortId,
		migrationInfo.IbcChannelId,
		msg.Amount,
		msg.Sender,
		msg.To,
		clienttypes.NewHeight(0, 0),
		// use default timeout 10 minutes
		uint64(sdk.UnwrapSDKContext(ctx).BlockTime().UnixNano())+transfertypes.DefaultRelativePacketTimeoutTimestamp,
		memo,
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
