package migration

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
)

func TestIBCMiddleware_OnRecvPacket(t *testing.T) {
	ctx := sdk.Context{}.WithEventManager(sdk.NewEventManager())
	_, _, relayer := keyPubAddr()
	_, _, accAddr := keyPubAddr()
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
		DestinationChannel: "channel-0",
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
