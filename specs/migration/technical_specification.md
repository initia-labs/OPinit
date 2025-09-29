# Bridge Replacement System Technical Specification

## Architecture Components

### Core Modules

#### OPChild Module (L2)

- **Purpose**: Handles L2 token operations, L2→L1 migration, and IBC→L2 conversion
- **Key Functions**:
  - `MigrateToken`: Convert OP tokens to IBC tokens (L2→IBC)
  - `HandleMigratedTokenDeposit`: Core logic to convert IBC tokens to OP tokens (IBC→L2)
  - `HandleMigratedTokenWithdrawal`: Handle L2→L1 withdrawal with automatic migration
  - `RegisterMigrationInfo`: Register migration configuration

#### OPHost Module (L1)

- **Purpose**: Handles L1 bridge coordination and L1→L2 migration via IBC transfer
- **Key Functions**:
  - `HandleMigratedTokenDeposit`: Process `MsgInitiateTokenDeposit` by forwarding as IBC transfer
  - `HandleMigratedTokenWithdrawal`: Process in-flight `MsgFinalizeTokenWithdrawal` requests by withdrawing from IBC escrow
  - `SetMigrationInfo`: Register migration info for L1→L2 flow
  - Bridge hook encoding in IBC transfer memos

#### IBC Middleware

- **Purpose**: Intercepts incoming IBC transfer packets and triggers IBC→L2 conversion
- **Key Function**: `OnRecvPacket` - automatically calls `HandleMigratedTokenDeposit`

### Data Structures

#### MigrationInfo (OPChild)

```go
type MigrationInfo struct {
    Denom        string // L2 denom (e.g., "l2/1234567890ABCDEF")
    IbcChannelId string // IBC channel ID (e.g., "channel-0")
    IbcPortId    string // IBC port ID (e.g., "transfer")
}
```

#### MigrationInfo (OPHost)

```go
type MigrationInfo struct {
    BridgeId     uint64 // Bridge ID
    L1Denom      string // L1 denom (e.g., "uinit")
    IbcChannelId string // IBC channel ID
    IbcPortId    string // IBC port ID
}
```

#### IBCToL2DenomMap

- **Purpose**: Maps IBC denoms to L2 denoms for IBC→L2 conversion
- **Format**: `ibc/1234567890ABCDEF` → `l2/1234567890ABCDEF`
- **Usage**: Created during migration info registration

## Technical Implementation

### IBC Middleware Integration

#### Packet Interception Flow

```go
func (im IBCMiddleware) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) ibcexported.Acknowledgement {
    // 1. Parse transfer packet data
    var data transfertypes.FungibleTokenPacketData
    if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
        return im.app.OnRecvPacket(ctx, packet, relayer)
    }
    
    // 2. Skip if token originated from receiving chain
    if transfertypes.ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), data.Denom) {
        return im.app.OnRecvPacket(ctx, packet, relayer)
    }
    
    // 3. Compute IBC denom
    sourcePrefix := transfertypes.GetDenomPrefix(packet.GetDestPort(), packet.GetDestChannel())
    prefixedDenom := sourcePrefix + data.Denom
    denomTrace := transfertypes.ParseDenomTrace(prefixedDenom)
    ibcDenom := denomTrace.IBCDenom()
    
    // 4. Check if registered for conversion
    if hasMigration, err := im.opChildKeeper.HasIBCToL2DenomMap(ctx, ibcDenom); err != nil || !hasMigration {
        return im.app.OnRecvPacket(ctx, packet, relayer)
    }
    
    // 5. Resolve receiver address
    receiver, err := im.ac.StringToBytes(data.Receiver)
    if err != nil {
        return newEmitErrorAcknowledgement(err)
    }

    // 6. Record balance before IBC processing
    beforeBalance := im.bankKeeper.GetBalance(ctx, receiver, ibcDenom)
    
    // 7. Process IBC packet normally
    ack := im.app.OnRecvPacket(ctx, packet, relayer)
    if !ack.Success() {
        return ack
    }
    
    // 8. Check if balance increased
    afterBalance := im.bankKeeper.GetBalance(ctx, receiver, ibcDenom)
    if afterBalance.Amount.LTE(beforeBalance.Amount) {
        return ack
    }
    
    // 9. Trigger IBC→L2 conversion
    diff := afterBalance.Amount.Sub(beforeBalance.Amount)
    ibcCoin := sdk.NewCoin(ibcDenom, diff)
    l2Coin, err := im.opChildKeeper.HandleMigratedTokenDeposit(ctx, receiver, ibcCoin, data.Memo)
    if err != nil {
        return newEmitErrorAcknowledgement(err)
    }
    
    // 10. Emit conversion event
    ctx.EventManager().EmitEvent(sdk.NewEvent(
        EventTypeHandleMigratedTokenDeposit,
        sdk.NewAttribute(AttributeKeyReceiver, data.Receiver),
        sdk.NewAttribute(AttributeKeyIbcDenom, ibcDenom),
        sdk.NewAttribute(AttributeKeyAmount, l2Coin.String()),
    ))
    
    return ack
}
```

### OPHost IBC Transfer Integration

#### HandleMigratedTokenDeposit (OPHost)

```go
func (k Keeper) HandleMigratedTokenDeposit(ctx context.Context, msg *types.MsgInitiateTokenDeposit) (handled bool, err error) {
    l1Denom := msg.Amount.Denom
    migrationInfo, err := k.GetMigrationInfo(ctx, msg.BridgeId, l1Denom)
    if err != nil && errors.Is(err, collections.ErrNotFound) {
        return false, nil // Not configured for migration
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

    // Create IBC transfer message
    transferMsg := transfertypes.NewMsgTransfer(
        migrationInfo.IbcPortId,
        migrationInfo.IbcChannelId,
        msg.Amount,
        msg.Sender,
        msg.To,
        clienttypes.NewHeight(0, 0),
        uint64(sdk.UnwrapSDKContext(ctx).BlockTime().UnixNano())+transfertypes.DefaultRelativePacketTimeoutTimestamp, //nolint:gosec
        memo,
    )

    // Route IBC transfer via message router
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
```

#### HandleMigratedTokenWithdrawal (OPHost) - In-Flight Request Handling

```go
func (k Keeper) HandleMigratedTokenWithdrawal(ctx context.Context, msg *types.MsgFinalizeTokenWithdrawal) (handled bool, err error) {
    l1Denom := msg.Amount.Denom
    migrationInfo, err := k.GetMigrationInfo(ctx, msg.BridgeId, l1Denom)
    if err != nil && errors.Is(err, collections.ErrNotFound) {
        return false, nil // Not configured for migration
    } else if err != nil {
        return false, err
    }

    transferEscrowAddress := transfertypes.GetEscrowAddress(migrationInfo.IbcPortId, migrationInfo.IbcChannelId)
    receiver, err := k.authKeeper.AddressCodec().StringToBytes(msg.To)
    if err != nil {
        return false, err
    }

    withdrawnFunds := sdk.NewCoins(msg.Amount)
    if err := k.bankKeeper.SendCoins(ctx, transferEscrowAddress, receiver, withdrawnFunds); err != nil {
        return false, err
    }

    return true, nil
}
```

## Event System

### Event Types

#### Middleware Events

- **`handle_migrated_token_deposit`**: IBC transfer conversion events (emitted by middleware)
  - `receiver`: Token recipient address
  - `ibc_denom`: IBC token denomination that was burned
  - `amount`: L2 token amount that was minted

#### OPChild Keeper Events

- **`migrate_token`**: Forward migration completion (emitted by keeper)
- **`register_migration_info`**: Migration info registration (emitted by keeper)

#### OPHost Keeper Events

- **`register_migration_info`**: Migration info registration (emitted by keeper)
  - `bridge_id`: Bridge identifier
  - `l1_denom`: L1 token denomination
  - `ibc_channel_id`: IBC channel identifier
  - `ibc_port_id`: IBC port identifier
