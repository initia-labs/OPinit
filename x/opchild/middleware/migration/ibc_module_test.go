package migration

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
)

func TestIBCMiddleware_OnRecvPacket(t *testing.T) {
	ctx := sdk.Context{}.WithEventManager(sdk.NewEventManager())
	_, _, relayer := keyPubAddr()
	_, _, accAddr := keyPubAddr()
	cdc := codec.NewProtoCodec(nil)
	ac := address.NewBech32Codec("init")
	addr, err := ac.BytesToString(accAddr.Bytes())
	require.NoError(t, err)

	bankKeeper := &MockBankKeeper{
		ac:       ac,
		balances: make(map[string]sdk.Coins),
	}
	opchildKeeper := &MockOPChildKeeper{
		bankKeeper:      bankKeeper,
		ibcToL2DenomMap: make(map[string]string),
	}
	app := MockTransferApp{
		ac:         ac,
		bankKeeper: bankKeeper,
	}

	middleware := NewIBCMiddleware(
		ac,
		cdc,
		app,
		nil,
		bankKeeper,
		opchildKeeper,
	)

	// case 1. receiving chain is source chain case
	denom := "transfer/channel-0/uinit"
	packet := channeltypes.Packet{
		SourcePort:         "transfer",
		SourceChannel:      "channel-0",
		Data:               buildPacketData(t, denom, "100", addr, addr, ""),
		DestinationPort:    "transfer",
		DestinationChannel: "channel-1",
	}

	ack := middleware.OnRecvPacket(ctx, packet, relayer)
	require.True(t, ack.Success())

	// check balance increased
	require.Equal(t, sdk.NewCoin("uinit", sdkmath.NewInt(100)), bankKeeper.GetBalance(ctx, accAddr, "uinit"))

	// case 2. non-migrated asset transfer
	denom = "uinit"
	packet = channeltypes.Packet{
		SourcePort:         "transfer",
		SourceChannel:      "channel-0",
		Data:               buildPacketData(t, denom, "100", addr, addr, ""),
		DestinationPort:    "transfer",
		DestinationChannel: "channel-0",
	}

	ack = middleware.OnRecvPacket(ctx, packet, relayer)
	require.True(t, ack.Success())

	// check balance increased
	ibcDenom := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom(packet.GetDestPort(), packet.GetDestChannel(), denom)).IBCDenom()
	require.Equal(t, sdk.NewCoin(ibcDenom, sdkmath.NewInt(100)), bankKeeper.GetBalance(ctx, accAddr, ibcDenom))

	// case 3. migrated asset transfer
	opchildKeeper.ibcToL2DenomMap[ibcDenom] = "uinit"
	ack = middleware.OnRecvPacket(ctx, packet, relayer)
	require.True(t, ack.Success())

	// check balance increased; we transferred IBC denom but received uinit due to the migration
	require.Equal(t, sdk.NewCoin("uinit", sdkmath.NewInt(200)), bankKeeper.GetBalance(ctx, accAddr, "uinit"))
	require.Equal(t, sdk.NewCoin(ibcDenom, sdkmath.NewInt(100)), bankKeeper.GetBalance(ctx, accAddr, ibcDenom))
}

func TestIBCMiddleware_OnAcknowledgementPacket_ErrorRefundsMigratedTokens(t *testing.T) {
	ctx := sdk.Context{}.WithEventManager(sdk.NewEventManager())
	_, _, relayer := keyPubAddr()
	_, _, senderAcc := keyPubAddr()
	_, _, receiverAcc := keyPubAddr()
	cdc := codec.NewProtoCodec(nil)
	ac := address.NewBech32Codec("init")

	senderStr, err := ac.BytesToString(senderAcc.Bytes())
	require.NoError(t, err)
	receiverStr, err := ac.BytesToString(receiverAcc.Bytes())
	require.NoError(t, err)

	bankKeeper := &MockBankKeeper{
		ac:       ac,
		balances: make(map[string]sdk.Coins),
	}
	opchildKeeper := &MockOPChildKeeper{
		bankKeeper:      bankKeeper,
		ibcToL2DenomMap: make(map[string]string),
	}
	app := MockTransferApp{
		ac:         ac,
		bankKeeper: bankKeeper,
	}

	baseDenom := "uinit"
	fullDenom := transfertypes.GetPrefixedDenom("transfer", "channel-0", baseDenom)
	ibcDenom := transfertypes.ParseDenomTrace(fullDenom).IBCDenom()
	opchildKeeper.ibcToL2DenomMap[ibcDenom] = baseDenom

	app.onAcknowledgement = func(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error {
		bankKeeper.AddBalances(ctx, senderAcc, sdk.NewCoins(sdk.NewCoin(ibcDenom, sdkmath.NewInt(100))))
		return nil
	}

	middleware := NewIBCMiddleware(
		ac,
		cdc,
		app,
		nil,
		bankKeeper,
		opchildKeeper,
	)

	packet := channeltypes.Packet{
		SourcePort:         "transfer",
		SourceChannel:      "channel-0",
		Data:               buildPacketData(t, fullDenom, "100", senderStr, receiverStr, ""),
		DestinationPort:    "transfer",
		DestinationChannel: "channel-1",
	}
	ackBz := channeltypes.NewErrorAcknowledgement(errors.New("failed")).Acknowledgement()

	err = middleware.OnAcknowledgementPacket(ctx, packet, ackBz, relayer)
	require.NoError(t, err)

	require.Equal(t, sdk.NewCoin(baseDenom, sdkmath.NewInt(100)), bankKeeper.GetBalance(ctx, senderAcc, baseDenom))
	require.True(t, bankKeeper.GetBalance(ctx, senderAcc, ibcDenom).Amount.IsZero())

	found := false
	for _, evt := range ctx.EventManager().Events() {
		if evt.Type != EventTypeHandleMigratedTokenRefund {
			continue
		}

		found = true

		var receiverAttr, ibcAttr, amountAttr string
		for _, attr := range evt.Attributes {
			switch attr.Key {
			case AttributeKeyReceiver:
				receiverAttr = attr.Value
			case AttributeKeyIbcDenom:
				ibcAttr = attr.Value
			case AttributeKeyAmount:
				amountAttr = attr.Value
			}
		}

		require.Equal(t, senderStr, receiverAttr)
		require.Equal(t, ibcDenom, ibcAttr)
		require.Equal(t, sdk.NewCoin(baseDenom, sdkmath.NewInt(100)).String(), amountAttr)
		break
	}
	require.True(t, found, "expected EventTypeHandleMigratedTokenRefund event")
}

func TestIBCMiddleware_OnTimeoutPacket_RefundsMigratedTokens(t *testing.T) {
	ctx := sdk.Context{}.WithEventManager(sdk.NewEventManager())
	_, _, relayer := keyPubAddr()
	_, _, senderAcc := keyPubAddr()
	_, _, receiverAcc := keyPubAddr()
	cdc := codec.NewProtoCodec(nil)
	ac := address.NewBech32Codec("init")

	senderStr, err := ac.BytesToString(senderAcc.Bytes())
	require.NoError(t, err)
	receiverStr, err := ac.BytesToString(receiverAcc.Bytes())
	require.NoError(t, err)

	bankKeeper := &MockBankKeeper{
		ac:       ac,
		balances: make(map[string]sdk.Coins),
	}
	opchildKeeper := &MockOPChildKeeper{
		bankKeeper:      bankKeeper,
		ibcToL2DenomMap: make(map[string]string),
	}
	app := MockTransferApp{
		ac:         ac,
		bankKeeper: bankKeeper,
	}

	baseDenom := "uinit"
	fullDenom := transfertypes.GetPrefixedDenom("transfer", "channel-0", baseDenom)
	ibcDenom := transfertypes.ParseDenomTrace(fullDenom).IBCDenom()
	opchildKeeper.ibcToL2DenomMap[ibcDenom] = baseDenom

	app.onTimeout = func(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
		bankKeeper.AddBalances(ctx, senderAcc, sdk.NewCoins(sdk.NewCoin(ibcDenom, sdkmath.NewInt(100))))
		return nil
	}

	middleware := NewIBCMiddleware(
		ac,
		cdc,
		app,
		nil,
		bankKeeper,
		opchildKeeper,
	)

	packet := channeltypes.Packet{
		SourcePort:         "transfer",
		SourceChannel:      "channel-0",
		Data:               buildPacketData(t, fullDenom, "100", senderStr, receiverStr, ""),
		DestinationPort:    "transfer",
		DestinationChannel: "channel-1",
	}

	err = middleware.OnTimeoutPacket(ctx, packet, relayer)
	require.NoError(t, err)

	require.Equal(t, sdk.NewCoin(baseDenom, sdkmath.NewInt(100)), bankKeeper.GetBalance(ctx, senderAcc, baseDenom))
	require.True(t, bankKeeper.GetBalance(ctx, senderAcc, ibcDenom).Amount.IsZero())

	found := false
	for _, evt := range ctx.EventManager().Events() {
		if evt.Type != EventTypeHandleMigratedTokenRefund {
			continue
		}

		found = true

		var receiverAttr, ibcAttr, amountAttr string
		for _, attr := range evt.Attributes {
			switch attr.Key {
			case AttributeKeyReceiver:
				receiverAttr = attr.Value
			case AttributeKeyIbcDenom:
				ibcAttr = attr.Value
			case AttributeKeyAmount:
				amountAttr = attr.Value
			}
		}

		require.Equal(t, senderStr, receiverAttr)
		require.Equal(t, ibcDenom, ibcAttr)
		require.Equal(t, sdk.NewCoin(baseDenom, sdkmath.NewInt(100)).String(), amountAttr)
		break
	}
	require.True(t, found, "expected EventTypeHandleMigratedTokenRefund event")
}

func TestIBCMiddleware_OnAcknowledgementPacket_SenderChainSource_NoMigration(t *testing.T) {
	ctx := sdk.Context{}.WithEventManager(sdk.NewEventManager())
	_, _, relayer := keyPubAddr()
	_, _, senderAcc := keyPubAddr()
	_, _, receiverAcc := keyPubAddr()
	cdc := codec.NewProtoCodec(nil)
	ac := address.NewBech32Codec("init")

	senderStr, err := ac.BytesToString(senderAcc.Bytes())
	require.NoError(t, err)
	receiverStr, err := ac.BytesToString(receiverAcc.Bytes())
	require.NoError(t, err)

	bankKeeper := &MockBankKeeper{
		ac:       ac,
		balances: make(map[string]sdk.Coins),
	}
	opchildKeeper := &MockOPChildKeeper{
		bankKeeper:      bankKeeper,
		ibcToL2DenomMap: make(map[string]string),
	}
	app := MockTransferApp{
		ac:         ac,
		bankKeeper: bankKeeper,
	}

	baseDenom := "uinit"
	opchildKeeper.ibcToL2DenomMap[baseDenom] = "l2/" + baseDenom

	app.onAcknowledgement = func(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error {
		bankKeeper.AddBalances(ctx, senderAcc, sdk.NewCoins(sdk.NewCoin(baseDenom, sdkmath.NewInt(100))))
		return nil
	}

	middleware := NewIBCMiddleware(
		ac,
		cdc,
		app,
		nil,
		bankKeeper,
		opchildKeeper,
	)

	packet := channeltypes.Packet{
		SourcePort:         "transfer",
		SourceChannel:      "channel-0",
		Data:               buildPacketData(t, baseDenom, "100", senderStr, receiverStr, ""),
		DestinationPort:    "transfer",
		DestinationChannel: "channel-1",
	}
	ackBz := channeltypes.NewErrorAcknowledgement(errors.New("failed")).Acknowledgement()

	err = middleware.OnAcknowledgementPacket(ctx, packet, ackBz, relayer)
	require.NoError(t, err)

	require.Equal(t, sdk.NewCoin(baseDenom, sdkmath.NewInt(100)), bankKeeper.GetBalance(ctx, senderAcc, baseDenom))
	require.True(t, bankKeeper.GetBalance(ctx, senderAcc, opchildKeeper.ibcToL2DenomMap[baseDenom]).Amount.IsZero())

	for _, evt := range ctx.EventManager().Events() {
		require.NotEqual(t, EventTypeHandleMigratedTokenRefund, evt.Type)
	}
}

func TestIBCMiddleware_OnTimeoutPacket_SenderChainSource_NoMigration(t *testing.T) {
	ctx := sdk.Context{}.WithEventManager(sdk.NewEventManager())
	_, _, relayer := keyPubAddr()
	_, _, senderAcc := keyPubAddr()
	_, _, receiverAcc := keyPubAddr()
	cdc := codec.NewProtoCodec(nil)
	ac := address.NewBech32Codec("init")

	senderStr, err := ac.BytesToString(senderAcc.Bytes())
	require.NoError(t, err)
	receiverStr, err := ac.BytesToString(receiverAcc.Bytes())
	require.NoError(t, err)

	bankKeeper := &MockBankKeeper{
		ac:       ac,
		balances: make(map[string]sdk.Coins),
	}
	opchildKeeper := &MockOPChildKeeper{
		bankKeeper:      bankKeeper,
		ibcToL2DenomMap: make(map[string]string),
	}
	app := MockTransferApp{
		ac:         ac,
		bankKeeper: bankKeeper,
	}

	baseDenom := "uinit"
	opchildKeeper.ibcToL2DenomMap[baseDenom] = "l2/" + baseDenom

	app.onTimeout = func(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
		bankKeeper.AddBalances(ctx, senderAcc, sdk.NewCoins(sdk.NewCoin(baseDenom, sdkmath.NewInt(100))))
		return nil
	}

	middleware := NewIBCMiddleware(
		ac,
		cdc,
		app,
		nil,
		bankKeeper,
		opchildKeeper,
	)

	packet := channeltypes.Packet{
		SourcePort:         "transfer",
		SourceChannel:      "channel-0",
		Data:               buildPacketData(t, baseDenom, "100", senderStr, receiverStr, ""),
		DestinationPort:    "transfer",
		DestinationChannel: "channel-1",
	}

	err = middleware.OnTimeoutPacket(ctx, packet, relayer)
	require.NoError(t, err)

	require.Equal(t, sdk.NewCoin(baseDenom, sdkmath.NewInt(100)), bankKeeper.GetBalance(ctx, senderAcc, baseDenom))
	require.True(t, bankKeeper.GetBalance(ctx, senderAcc, opchildKeeper.ibcToL2DenomMap[baseDenom]).Amount.IsZero())

	for _, evt := range ctx.EventManager().Events() {
		require.NotEqual(t, EventTypeHandleMigratedTokenRefund, evt.Type)
	}
}
