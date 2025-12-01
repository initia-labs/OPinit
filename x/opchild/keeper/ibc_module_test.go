package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	"github.com/initia-labs/OPinit/x/opchild"
	opchildtypes "github.com/initia-labs/OPinit/x/opchild/types"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
)

// helper function to create an Attestor
func createAttestor(t *testing.T, operatorAddr string, pubKey cryptotypes.PubKey, moniker string) ophosttypes.Attestor {
	pkAny, err := codectypes.NewAnyWithValue(pubKey)
	require.NoError(t, err)

	return ophosttypes.Attestor{
		OperatorAddress: operatorAddr,
		ConsensusPubkey: pkAny,
		Moniker:         moniker,
	}
}

func Test_IBCModule_OnChanOpenInit(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	err := input.OPChildKeeper.PortID.Set(ctx, opchildtypes.PortID)
	require.NoError(t, err)

	ibcModule := opchild.NewIBCModule(input.OPChildKeeper)
	capability := &capabilitytypes.Capability{}

	// success case with valid parameters
	version, err := ibcModule.OnChanOpenInit(
		sdkCtx,
		channeltypes.UNORDERED,
		[]string{"connection-0"},
		opchildtypes.PortID,
		"channel-0",
		capability,
		channeltypes.NewCounterparty(opchildtypes.PortID, "channel-1"),
		opchildtypes.Version,
	)
	require.NoError(t, err)
	require.Equal(t, opchildtypes.Version, version)

	// empty version string
	version, err = ibcModule.OnChanOpenInit(
		sdkCtx,
		channeltypes.UNORDERED,
		[]string{"connection-0"},
		opchildtypes.PortID,
		"channel-1",
		capability,
		channeltypes.NewCounterparty(opchildtypes.PortID, "channel-2"),
		"",
	)
	require.NoError(t, err)
	require.Equal(t, opchildtypes.Version, version)

	// invalid channel ordering
	_, err = ibcModule.OnChanOpenInit(
		sdkCtx,
		channeltypes.ORDERED,
		[]string{"connection-0"},
		opchildtypes.PortID,
		"channel-2",
		capability,
		channeltypes.NewCounterparty(opchildtypes.PortID, "channel-3"),
		opchildtypes.Version,
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected ORDER_UNORDERED channel")

	// invalid port
	_, err = ibcModule.OnChanOpenInit(
		sdkCtx,
		channeltypes.UNORDERED,
		[]string{"connection-0"},
		"invalid-port",
		"channel-3",
		capability,
		channeltypes.NewCounterparty(opchildtypes.PortID, "channel-4"),
		opchildtypes.Version,
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid port")

	// invalid version
	_, err = ibcModule.OnChanOpenInit(
		sdkCtx,
		channeltypes.UNORDERED,
		[]string{"connection-0"},
		opchildtypes.PortID,
		"channel-4",
		capability,
		channeltypes.NewCounterparty(opchildtypes.PortID, "channel-5"),
		"invalid-version",
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected "+opchildtypes.Version)
}

func Test_IBCModule_OnChanOpenTry(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	err := input.OPChildKeeper.PortID.Set(ctx, opchildtypes.PortID)
	require.NoError(t, err)

	ibcModule := opchild.NewIBCModule(input.OPChildKeeper)
	capability := &capabilitytypes.Capability{}

	// success case
	version, err := ibcModule.OnChanOpenTry(
		sdkCtx,
		channeltypes.UNORDERED,
		[]string{"connection-0"},
		opchildtypes.PortID,
		"channel-0",
		capability,
		channeltypes.NewCounterparty(opchildtypes.PortID, "channel-1"),
		opchildtypes.Version,
	)
	require.NoError(t, err)
	require.Equal(t, opchildtypes.Version, version)

	// invalid channel ordering
	_, err = ibcModule.OnChanOpenTry(
		sdkCtx,
		channeltypes.ORDERED,
		[]string{"connection-0"},
		opchildtypes.PortID,
		"channel-1",
		capability,
		channeltypes.NewCounterparty(opchildtypes.PortID, "channel-2"),
		opchildtypes.Version,
	)
	require.Error(t, err)

	// invalid counterparty version
	_, err = ibcModule.OnChanOpenTry(
		sdkCtx,
		channeltypes.UNORDERED,
		[]string{"connection-0"},
		opchildtypes.PortID,
		"channel-2",
		capability,
		channeltypes.NewCounterparty(opchildtypes.PortID, "channel-3"),
		"invalid-version",
	)
	require.Error(t, err)
}

func Test_IBCModule_OnChanOpenAck(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	ibcModule := opchild.NewIBCModule(input.OPChildKeeper)

	// success case
	err := ibcModule.OnChanOpenAck(
		sdkCtx,
		opchildtypes.PortID,
		"channel-0",
		"channel-1",
		opchildtypes.Version,
	)
	require.NoError(t, err)

	// invalid version
	err = ibcModule.OnChanOpenAck(
		sdkCtx,
		opchildtypes.PortID,
		"channel-0",
		"channel-1",
		"invalid-version",
	)
	require.Error(t, err)
}

func Test_IBCModule_OnChanOpenConfirm(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	ibcModule := opchild.NewIBCModule(input.OPChildKeeper)

	err := ibcModule.OnChanOpenConfirm(
		sdkCtx,
		opchildtypes.PortID,
		"channel-0",
	)
	require.NoError(t, err)
}

func Test_IBCModule_OnChanCloseInit(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	ibcModule := opchild.NewIBCModule(input.OPChildKeeper)

	// user cannot close channel
	err := ibcModule.OnChanCloseInit(
		sdkCtx,
		opchildtypes.PortID,
		"channel-0",
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "user cannot close channel")
}

func Test_IBCModule_OnChanCloseConfirm(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	ibcModule := opchild.NewIBCModule(input.OPChildKeeper)

	err := ibcModule.OnChanCloseConfirm(
		sdkCtx,
		opchildtypes.PortID,
		"channel-0",
	)
	require.NoError(t, err)
}

func Test_IBCModule_OnRecvPacket_AttestorSetUpdate(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	bridgeInfo := opchildtypes.BridgeInfo{
		BridgeId:   1,
		BridgeAddr: addrsStr[0],
		L1ChainId:  "test-chain-1",
		L1ClientId: "test-client-id",
		BridgeConfig: ophosttypes.BridgeConfig{
			AttestorSet: []ophosttypes.Attestor{},
		},
	}
	err := input.OPChildKeeper.BridgeInfo.Set(ctx, bridgeInfo)
	require.NoError(t, err)

	// attestor set update packet
	attestor1 := createAttestor(t, valAddrsStr[1], pubKeys[1], "attestor1")
	attestor2 := createAttestor(t, valAddrsStr[2], pubKeys[2], "attestor2")

	packetData := ophosttypes.NewAttestorSetUpdatePacketData(
		1,
		[]ophosttypes.Attestor{attestor1, attestor2},
		100,
	)

	packet := channeltypes.Packet{
		SourcePort:         opchildtypes.PortID,
		SourceChannel:      "channel-0",
		DestinationPort:    opchildtypes.PortID,
		DestinationChannel: "channel-1",
		Data:               packetData.GetBytes(),
		Sequence:           1,
	}

	ibcModule := opchild.NewIBCModule(input.OPChildKeeper)

	// successful packet receipt
	ack := ibcModule.OnRecvPacket(sdkCtx, packet, nil)
	require.True(t, ack.Success())

	// attestors were added
	vals, err := input.OPChildKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	require.Len(t, vals, 2)
}

func Test_IBCModule_OnRecvPacket_InvalidData(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	ibcModule := opchild.NewIBCModule(input.OPChildKeeper)

	packet := channeltypes.Packet{
		SourcePort:         opchildtypes.PortID,
		SourceChannel:      "channel-0",
		DestinationPort:    opchildtypes.PortID,
		DestinationChannel: "channel-1",
		Data:               []byte("invalid-data"),
		Sequence:           1,
	}

	ack := ibcModule.OnRecvPacket(sdkCtx, packet, nil)
	require.False(t, ack.Success())
}

func Test_IBCModule_OnAcknowledgementPacket(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	ibcModule := opchild.NewIBCModule(input.OPChildKeeper)

	packet := channeltypes.Packet{
		SourcePort:         opchildtypes.PortID,
		SourceChannel:      "channel-0",
		DestinationPort:    opchildtypes.PortID,
		DestinationChannel: "channel-1",
		Data:               []byte("test-data"),
		Sequence:           1,
	}

	ack := channeltypes.NewResultAcknowledgement([]byte("success"))

	err := ibcModule.OnAcknowledgementPacket(sdkCtx, packet, ack.Acknowledgement(), nil)
	require.NoError(t, err)

	events := sdkCtx.EventManager().Events()
	found := false
	for _, event := range events {
		if event.Type == opchildtypes.EventTypePacket {
			found = true
			break
		}
	}
	require.True(t, found, "expected packet event to be emitted")
}

func Test_IBCModule_OnTimeoutPacket(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	ibcModule := opchild.NewIBCModule(input.OPChildKeeper)

	packet := channeltypes.Packet{
		SourcePort:         opchildtypes.PortID,
		SourceChannel:      "channel-0",
		DestinationPort:    opchildtypes.PortID,
		DestinationChannel: "channel-1",
		Data:               []byte("test-data"),
		Sequence:           1,
	}

	err := ibcModule.OnTimeoutPacket(sdkCtx, packet, nil)
	require.NoError(t, err)

	events := sdkCtx.EventManager().Events()
	found := false
	for _, event := range events {
		if event.Type == opchildtypes.EventTypeTimeout {
			found = true
			break
		}
	}
	require.True(t, found, "expected timeout event to be emitted")
}
