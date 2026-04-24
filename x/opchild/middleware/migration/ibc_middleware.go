package migration

import (
	"fmt"

	"cosmossdk.io/core/address"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
)

// Interface assertions to ensure IBCMiddleware implements required interfaces
var _ porttypes.Middleware = &IBCMiddleware{}

// IBCMiddleware wraps an underlying IBC module and intercepts fungible token
// transfers in order to migrate legacy OP-tokens into their IBC-token equivalent
// when a registered denom is received, acked, or timed out.
type IBCMiddleware struct {
	ac  address.Codec
	cdc codec.Codec

	// app is the underlying IBC module that handles standard IBC operations
	app porttypes.IBCModule
	// ics4Wrapper provides packet sending/receiving capabilities for the middleware
	ics4Wrapper porttypes.ICS4Wrapper
	// bankKeeper is the keeper for the bank module
	bankKeeper BankKeeper
	// opChildKeeper is the keeper for the opchild module
	opChildKeeper OPChildKeeper
}

// NewIBCMiddleware creates a new IBCMiddleware given the keeper and underlying application
//
// Parameters:
//   - app: The underlying IBC module that handles standard IBC operations
//   - ics4Wrapper: Provides packet sending/receiving capabilities
//   - rootModule: The top-level IBC module that handles upgrade-related callbacks
//
// Returns:
//   - IBCMiddleware: A configured middleware instance
func NewIBCMiddleware(
	ac address.Codec,
	cdc codec.Codec,
	app porttypes.IBCModule,
	ics4Wrapper porttypes.ICS4Wrapper,
	bankKeeper BankKeeper,
	opChildKeeper OPChildKeeper,
) IBCMiddleware {
	return IBCMiddleware{
		ac:            ac,
		cdc:           cdc,
		app:           app,
		ics4Wrapper:   ics4Wrapper,
		bankKeeper:    bankKeeper,
		opChildKeeper: opChildKeeper,
	}
}

// OnChanOpenInit implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	return im.app.OnChanOpenInit(ctx, order, connectionHops, portID, channelID, counterparty, version)
}

// OnChanOpenTry implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	return im.app.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, counterparty, counterpartyVersion)
}

// OnChanOpenAck implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	return im.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)
}

// OnChanOpenConfirm implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return im.app.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnChanCloseInit implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return im.app.OnChanCloseInit(ctx, portID, channelID)
}

// OnChanCloseConfirm implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return im.app.OnChanCloseConfirm(ctx, portID, channelID)
}

// SendPacket implements the ICS4 Wrapper interface
func (im IBCMiddleware) SendPacket(
	ctx sdk.Context,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	return im.ics4Wrapper.SendPacket(
		ctx,
		sourcePort,
		sourceChannel,
		timeoutHeight,
		timeoutTimestamp,
		data,
	)
}

// WriteAcknowledgement implements the ICS4 Wrapper interface
func (im IBCMiddleware) WriteAcknowledgement(
	ctx sdk.Context,
	packet ibcexported.PacketI,
	ack ibcexported.Acknowledgement,
) error {
	return im.ics4Wrapper.WriteAcknowledgement(ctx, packet, ack)
}

// GetAppVersion returns the application version of the underlying application
func (im IBCMiddleware) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	return im.ics4Wrapper.GetAppVersion(ctx, portID, channelID)
}

// OnRecvPacket implements the IBCMiddleware interface
func (im IBCMiddleware) OnRecvPacket(
	ctx sdk.Context,
	channelVersion string,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	// if it is not a transfer packet or receiver chain is source, then execute inner app
	// without any further checks
	data, ibcDenom, ok := lookupPacket(packet, true)
	if !ok {
		return im.app.OnRecvPacket(ctx, channelVersion, packet, relayer)
	}

	// if the token is not registered for migration, do nothing
	if hasMigration, err := im.opChildKeeper.HasIBCToL2DenomMap(ctx, ibcDenom); err != nil || !hasMigration {
		return im.app.OnRecvPacket(ctx, channelVersion, packet, relayer)
	}

	// get the receiver address
	receiver, err := im.ac.StringToBytes(data.Receiver)
	if err != nil {
		return newEmitErrorAcknowledgement(err)
	}

	// get the before balance
	beforeBalance := im.bankKeeper.GetBalance(ctx, receiver, ibcDenom)

	// call the underlying IBC module
	ack := im.app.OnRecvPacket(ctx, channelVersion, packet, relayer)
	if !ack.Success() {
		return ack
	}

	// if the balance is not changed, do nothing
	afterBalance := im.bankKeeper.GetBalance(ctx, receiver, ibcDenom)
	if afterBalance.Amount.LTE(beforeBalance.Amount) {
		return ack
	}

	// compute the difference
	diff := afterBalance.Amount.Sub(beforeBalance.Amount)

	// burn IBC token and mint L2 token
	ibcCoin := sdk.NewCoin(ibcDenom, diff)
	l2Coin, err := im.opChildKeeper.HandleMigratedTokenDeposit(ctx, receiver, ibcCoin, data.Memo)
	if err != nil {
		return newEmitErrorAcknowledgement(err)
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		EventTypeHandleMigratedTokenDeposit,
		sdk.NewAttribute(AttributeKeyReceiver, data.Receiver),
		sdk.NewAttribute(AttributeKeyIbcDenom, ibcDenom),
		sdk.NewAttribute(AttributeKeyAmount, l2Coin.String()),
	))

	return ack
}

// OnAcknowledgementPacket implements the IBCMiddleware interface
func (im IBCMiddleware) OnAcknowledgementPacket(
	ctx sdk.Context,
	channelVersion string,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	// if it is not an error ack, just pass through
	if !isAckError(im.cdc, acknowledgement) {
		return im.app.OnAcknowledgementPacket(ctx, channelVersion, packet, acknowledgement, relayer)
	}

	// if it is not a transfer packet or sender chain is source, then execute inner app
	// without any further checks
	data, ibcDenom, ok := lookupPacket(packet, false)
	if !ok {
		return im.app.OnAcknowledgementPacket(ctx, channelVersion, packet, acknowledgement, relayer)
	}

	// if the token is not registered for migration, just pass through
	if hasMigration, err := im.opChildKeeper.HasIBCToL2DenomMap(ctx, ibcDenom); err != nil || !hasMigration {
		return im.app.OnAcknowledgementPacket(ctx, channelVersion, packet, acknowledgement, relayer)
	}

	// get the sender address
	sender, err := im.ac.StringToBytes(data.Sender)
	if err != nil {
		return err
	}

	// get the before balance
	beforeBalance := im.bankKeeper.GetBalance(ctx, sender, ibcDenom)

	// call the underlying IBC module
	if err := im.app.OnAcknowledgementPacket(ctx, channelVersion, packet, acknowledgement, relayer); err != nil {
		return err
	}

	// if the balance is not changed, do nothing
	afterBalance := im.bankKeeper.GetBalance(ctx, sender, ibcDenom)
	if afterBalance.Amount.LTE(beforeBalance.Amount) {
		return nil
	}

	// compute the difference
	diff := afterBalance.Amount.Sub(beforeBalance.Amount)

	// burn IBC token and mint L2 token
	ibcCoin := sdk.NewCoin(ibcDenom, diff)
	l2Coin, err := im.opChildKeeper.HandleMigratedTokenDeposit(ctx, sender, ibcCoin, "")
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		EventTypeHandleMigratedTokenRefund,
		sdk.NewAttribute(AttributeKeyReceiver, data.Sender),
		sdk.NewAttribute(AttributeKeyIbcDenom, ibcDenom),
		sdk.NewAttribute(AttributeKeyAmount, l2Coin.String()),
	))

	return nil
}

// OnTimeoutPacket implements the IBCMiddleware interface
func (im IBCMiddleware) OnTimeoutPacket(
	ctx sdk.Context,
	channelVersion string,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	// if it is not a transfer packet or sender chain is source, then execute inner app
	// without any further checks
	data, ibcDenom, ok := lookupPacket(packet, false)
	if !ok {
		return im.app.OnTimeoutPacket(ctx, channelVersion, packet, relayer)
	}

	// if the token is not registered for migration, just pass through
	if hasMigration, err := im.opChildKeeper.HasIBCToL2DenomMap(ctx, ibcDenom); err != nil || !hasMigration {
		return im.app.OnTimeoutPacket(ctx, channelVersion, packet, relayer)
	}

	// get the sender address
	sender, err := im.ac.StringToBytes(data.Sender)
	if err != nil {
		return err
	}

	// get the before balance
	beforeBalance := im.bankKeeper.GetBalance(ctx, sender, ibcDenom)

	// call the underlying IBC module
	if err := im.app.OnTimeoutPacket(ctx, channelVersion, packet, relayer); err != nil {
		return err
	}

	// if the balance is not changed, do nothing
	afterBalance := im.bankKeeper.GetBalance(ctx, sender, ibcDenom)
	if afterBalance.Amount.LTE(beforeBalance.Amount) {
		return nil
	}

	// compute the difference
	diff := afterBalance.Amount.Sub(beforeBalance.Amount)

	// burn IBC token and mint L2 token
	ibcCoin := sdk.NewCoin(ibcDenom, diff)
	l2Coin, err := im.opChildKeeper.HandleMigratedTokenDeposit(ctx, sender, ibcCoin, "")
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		EventTypeHandleMigratedTokenRefund,
		sdk.NewAttribute(AttributeKeyReceiver, data.Sender),
		sdk.NewAttribute(AttributeKeyIbcDenom, ibcDenom),
		sdk.NewAttribute(AttributeKeyAmount, l2Coin.String()),
	))

	return nil
}

// lookupPacket checks if the packet is a fungible token transfer packet and not originated from the
// receiving chain (if receive=true) or sending chain (if receive=false). If so, it computes the IBC denom
// and returns it along with the parsed packet data. Otherwise, it returns ok=false.
func lookupPacket(packet channeltypes.Packet, receive bool) (data transfertypes.FungibleTokenPacketData, ibcDenom string, needCheck bool) {
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return data, "", false
	}

	denom := transfertypes.ExtractDenomFromPath(data.Denom)
	// if the token is originated from the receiving chain, do nothing
	if receive && denom.HasPrefix(packet.GetSourcePort(), packet.GetSourceChannel()) {
		return data, "", false
	}

	// if the token is originated from the sending chain, do nothing
	if !receive && !denom.HasPrefix(packet.GetSourcePort(), packet.GetSourceChannel()) {
		return data, "", false
	}

	// compute the prefixed ibc denom
	if receive {
		destHop := transfertypes.NewHop(packet.GetDestPort(), packet.GetDestChannel())
		denom = transfertypes.NewDenom(denom.Base, append([]transfertypes.Hop{destHop}, denom.Trace...)...)
	}

	return data, denom.IBCDenom(), true
}

// newEmitErrorAcknowledgement creates a new error acknowledgement after having emitted an event with the
// details of the error.
func newEmitErrorAcknowledgement(err error) channeltypes.Acknowledgement {
	return channeltypes.Acknowledgement{
		Response: &channeltypes.Acknowledgement_Error{
			Error: fmt.Sprintf("ibc middleware migration error: %s", err.Error()),
		},
	}
}

// isAckError checks an IBC acknowledgement to see if it's an error.
func isAckError(appCodec codec.Codec, acknowledgement []byte) bool {
	var ack channeltypes.Acknowledgement
	if err := appCodec.UnmarshalJSON(acknowledgement, &ack); err == nil && !ack.Success() {
		return true
	}

	return false
}
