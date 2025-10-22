package migration

import (
	"context"
	"encoding/binary"
	"testing"

	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
)

var _ porttypes.IBCModule = &MockTransferApp{}
var _ BankKeeper = &MockBankKeeper{}
var _ OPChildKeeper = &MockOPChildKeeper{}

type MockBankKeeper struct {
	ac       address.Codec
	balances map[string]sdk.Coins
}

func (m MockBankKeeper) GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	addrStr, err := m.ac.BytesToString(addr.Bytes())
	if err != nil {
		return sdk.Coin{}
	}
	if val, ok := m.balances[addrStr]; !ok {
		return sdk.NewCoin(denom, sdkmath.ZeroInt())
	} else {
		return sdk.NewCoin(denom, val.AmountOf(denom))
	}
}

func (m MockBankKeeper) GetBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins {
	addrStr, err := m.ac.BytesToString(addr.Bytes())
	if err != nil {
		return sdk.NewCoins()
	}
	return m.balances[addrStr]
}

func (m MockBankKeeper) AddBalances(ctx context.Context, addr sdk.AccAddress, coins sdk.Coins) {
	addrStr, err := m.ac.BytesToString(addr.Bytes())
	if err != nil {
		return
	}
	balances := m.GetBalances(ctx, addr)
	m.balances[addrStr] = balances.Add(coins...)
}

func (m MockBankKeeper) SubBalances(ctx context.Context, addr sdk.AccAddress, coins sdk.Coins) {
	addrStr, err := m.ac.BytesToString(addr.Bytes())
	if err != nil {
		return
	}
	balances := m.GetBalances(ctx, addr)
	m.balances[addrStr] = balances.Sub(coins...)
}

type MockOPChildKeeper struct {
	bankKeeper      BankKeeper
	ibcToL2DenomMap map[string]string
}

func (m MockOPChildKeeper) HasIBCToL2DenomMap(ctx context.Context, ibcDenom string) (bool, error) {
	_, ok := m.ibcToL2DenomMap[ibcDenom]
	return ok, nil
}

// burn IBC token and mint L2 token
func (m MockOPChildKeeper) HandleMigratedTokenDeposit(ctx context.Context, sender sdk.AccAddress, ibcCoin sdk.Coin, memo string) (sdk.Coin, error) {
	l2Denom, ok := m.ibcToL2DenomMap[ibcCoin.Denom]
	if !ok {
		return sdk.Coin{}, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "ibc denom not found in ibcToL2DenomMap")
	}

	l2Coin := sdk.NewCoin(l2Denom, ibcCoin.Amount)
	m.bankKeeper.(*MockBankKeeper).SubBalances(ctx, sender, sdk.NewCoins(ibcCoin))
	m.bankKeeper.(*MockBankKeeper).AddBalances(ctx, sender, sdk.NewCoins(l2Coin))
	return l2Coin, nil
}

type MockTransferApp struct {
	ac                address.Codec
	bankKeeper        BankKeeper
	onAcknowledgement func(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error
	onTimeout         func(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error
}

// OnChanOpenInit implements the IBCMiddleware interface
func (im MockTransferApp) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	channelCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	return transfertypes.Version, nil
}

// OnChanOpenTry implements the IBCMiddleware interface
func (im MockTransferApp) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	channelCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	return transfertypes.Version, nil
}

// OnChanOpenAck implements the IBCMiddleware interface
func (im MockTransferApp) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	return nil
}

// OnChanOpenConfirm implements the IBCMiddleware interface
func (im MockTransferApp) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return nil
}

// OnChanCloseInit implements the IBCMiddleware interface
func (im MockTransferApp) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return nil
}

// OnChanCloseConfirm implements the IBCMiddleware interface
func (im MockTransferApp) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return nil
}

// OnAcknowledgementPacket implements the IBCMiddleware interface
func (im MockTransferApp) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	if im.onAcknowledgement != nil {
		return im.onAcknowledgement(ctx, packet, acknowledgement, relayer)
	}
	return nil
}

// OnTimeoutPacket implements the IBCMiddleware interface
func (im MockTransferApp) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	if im.onTimeout != nil {
		return im.onTimeout(ctx, packet, relayer)
	}
	return nil
}

// OnRecvPacket implements the IBCMiddleware interface
//
// unescrow IBC token if the token is originated from the receiving chain (in test, it is same with minting)
// mint IBC token if the token is not originated from the receiving chain
func (im MockTransferApp) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	// if it is not a transfer packet, do nothing
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return newEmitErrorAcknowledgement(err)
	}

	// validate the data
	if err := data.ValidateBasic(); err != nil {
		return newEmitErrorAcknowledgement(err)
	}

	// validate the amount
	transferAmount, ok := sdkmath.NewIntFromString(data.Amount)
	if !ok {
		return newEmitErrorAcknowledgement(errorsmod.Wrapf(transfertypes.ErrInvalidAmount, "invalid amount"))
	}

	// if the token is originated from the receiving chain, do nothing
	var denom string
	if transfertypes.ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), data.Denom) {
		voucherPrefix := transfertypes.GetDenomPrefix(packet.GetSourcePort(), packet.GetSourceChannel())
		unprefixedDenom := data.Denom[len(voucherPrefix):]

		// coin denomination used in sending from the escrow address
		denom = unprefixedDenom
	} else {
		// compute the ibc denom
		sourcePrefix := transfertypes.GetDenomPrefix(packet.GetDestPort(), packet.GetDestChannel())
		prefixedDenom := sourcePrefix + data.Denom
		denomTrace := transfertypes.ParseDenomTrace(prefixedDenom)
		ibcDenom := denomTrace.IBCDenom()
		denom = ibcDenom
	}

	receiverAddr, err := im.ac.StringToBytes(data.Receiver)
	if err != nil {
		return newEmitErrorAcknowledgement(err)
	}
	im.bankKeeper.(*MockBankKeeper).AddBalances(ctx, receiverAddr, sdk.NewCoins(sdk.NewCoin(denom, transferAmount)))

	return channeltypes.NewResultAcknowledgement([]byte{byte(1)})
}

var keyCounter uint64

// we need to make this deterministic (same every test run), as encoded address size and thus gas cost,
// depends on the actual bytes (due to ugly CanonicalAddress encoding)
func keyPubAddr() (cryptotypes.PrivKey, cryptotypes.PubKey, sdk.AccAddress) {
	keyCounter++
	seed := make([]byte, 8)
	binary.BigEndian.PutUint64(seed, keyCounter)

	key := secp256k1.GenPrivKeyFromSecret(seed)
	pub := key.PubKey()
	addr := sdk.AccAddress(pub.Address())
	return key, pub, addr
}

func buildPacketData(t *testing.T, denom, amount, sender, receiver, memo string) []byte {
	data, err := transfertypes.ModuleCdc.MarshalJSON(&transfertypes.FungibleTokenPacketData{
		Denom:    denom,
		Amount:   amount,
		Sender:   sender,
		Receiver: receiver,
		Memo:     memo,
	})
	require.NoError(t, err)
	return data
}
