package opchild_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	connectiontypes "github.com/cosmos/ibc-go/v10/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"

	"github.com/initia-labs/OPinit/x/opchild"
	"github.com/initia-labs/OPinit/x/opchild/testutil"
	opchildtypes "github.com/initia-labs/OPinit/x/opchild/types"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
)

const (
	testL1ClientID    = "test-client-id"
	testConnectionID  = "connection-0"
	testSourceChannel = "channel-0"
	testDestChannel   = "channel-1"
)

func setupAttestorSetUpdateOrigin(t *testing.T, ctx sdk.Context, input testutil.TestKeepers, l1ClientID string) {
	t.Helper()

	input.IBCKeeper.ConnectionKeeper.SetClientConnectionPaths(ctx, l1ClientID, []string{testConnectionID})
	input.IBCKeeper.ConnectionKeeper.SetConnection(ctx, testConnectionID, connectiontypes.ConnectionEnd{
		State:    connectiontypes.OPEN,
		ClientId: l1ClientID,
	})
	input.IBCKeeper.ChannelKeeper.SetChannel(ctx, opchildtypes.PortID, testDestChannel, channeltypes.Channel{
		State: channeltypes.OPEN,
		Counterparty: channeltypes.Counterparty{
			PortId:    opchildtypes.PortID,
			ChannelId: testSourceChannel,
		},
		ConnectionHops: []string{testConnectionID},
	})
}

func Test_IBCModule_OnChanOpenInit(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	ibcModule := opchild.NewIBCModule(input.OPChildKeeper)

	// success case with valid parameters
	version, err := ibcModule.OnChanOpenInit(
		sdkCtx,
		channeltypes.UNORDERED,
		[]string{"connection-0"},
		opchildtypes.PortID,
		"channel-0",
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
		channeltypes.NewCounterparty(opchildtypes.PortID, "channel-3"),
		opchildtypes.Version,
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected ORDER_UNORDERED channel")

	// invalid version
	_, err = ibcModule.OnChanOpenInit(
		sdkCtx,
		channeltypes.UNORDERED,
		[]string{"connection-0"},
		opchildtypes.PortID,
		"channel-4",
		channeltypes.NewCounterparty(opchildtypes.PortID, "channel-5"),
		"invalid-version",
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected "+opchildtypes.Version)
}

func Test_IBCModule_OnChanOpenTry(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	ibcModule := opchild.NewIBCModule(input.OPChildKeeper)

	// success case
	version, err := ibcModule.OnChanOpenTry(
		sdkCtx,
		channeltypes.UNORDERED,
		[]string{"connection-0"},
		opchildtypes.PortID,
		"channel-0",
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
		channeltypes.NewCounterparty(opchildtypes.PortID, "channel-3"),
		"invalid-version",
	)
	require.Error(t, err)
}

func Test_IBCModule_OnChanOpenAck(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
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
	ctx, input := testutil.CreateTestInput(t, false)
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
	ctx, input := testutil.CreateTestInput(t, false)
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
	ctx, input := testutil.CreateTestInput(t, false)
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
	ctx, input := testutil.CreateTestInput(t, false)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	bridgeInfo := opchildtypes.BridgeInfo{
		BridgeId:   1,
		BridgeAddr: testutil.AddrsStr[0],
		L1ChainId:  "test-chain-1",
		L1ClientId: testL1ClientID,
		BridgeConfig: ophosttypes.BridgeConfig{
			AttestorSet: []ophosttypes.Attestor{},
			ChannelId:   testSourceChannel,
		},
	}
	err := input.OPChildKeeper.BridgeInfo.Set(ctx, bridgeInfo)
	require.NoError(t, err)
	setupAttestorSetUpdateOrigin(t, sdkCtx, input, testL1ClientID)

	// attestor set update packet
	attestor1 := testutil.CreateAttestor(t, testutil.ValAddrsStr[1], testutil.PubKeys[1], "attestor1")
	attestor2 := testutil.CreateAttestor(t, testutil.ValAddrsStr[2], testutil.PubKeys[2], "attestor2")

	packetData := ophosttypes.NewAttestorSetUpdatePacketData(
		1,
		[]ophosttypes.Attestor{attestor1, attestor2},
		100,
	)

	packet := channeltypes.Packet{
		SourcePort:         opchildtypes.PortID,
		SourceChannel:      testSourceChannel,
		DestinationPort:    opchildtypes.PortID,
		DestinationChannel: testDestChannel,
		Data:               packetData.GetBytes(),
		Sequence:           1,
	}

	ibcModule := opchild.NewIBCModule(input.OPChildKeeper)

	// successful packet receipt
	ack := ibcModule.OnRecvPacket(sdkCtx, opchildtypes.Version, packet, nil)
	require.True(t, ack.Success())

	// attestors were added
	vals, err := input.OPChildKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	require.Len(t, vals, 2)
}

func Test_IBCModule_OnRecvPacket_InvalidOrigin(t *testing.T) {
	testCases := []struct {
		name           string
		channelVersion string
		setupOrigin    func(t *testing.T, ctx sdk.Context, input testutil.TestKeepers)
		mutatePacket   func(packet *channeltypes.Packet)
		expErr         string
	}{
		{
			name:           "invalid channel version",
			channelVersion: "invalid-version",
			setupOrigin: func(t *testing.T, ctx sdk.Context, input testutil.TestKeepers) {
				setupAttestorSetUpdateOrigin(t, ctx, input, testL1ClientID)
			},
			expErr: "expected channel version",
		},
		{
			name:           "invalid source port",
			channelVersion: opchildtypes.Version,
			setupOrigin: func(t *testing.T, ctx sdk.Context, input testutil.TestKeepers) {
				setupAttestorSetUpdateOrigin(t, ctx, input, testL1ClientID)
			},
			mutatePacket: func(packet *channeltypes.Packet) {
				packet.SourcePort = "attacker"
			},
			expErr: "expected source port",
		},
		{
			name:           "invalid destination port",
			channelVersion: opchildtypes.Version,
			setupOrigin: func(t *testing.T, ctx sdk.Context, input testutil.TestKeepers) {
				setupAttestorSetUpdateOrigin(t, ctx, input, testL1ClientID)
			},
			mutatePacket: func(packet *channeltypes.Packet) {
				packet.DestinationPort = "attacker"
			},
			expErr: "expected destination port",
		},
		{
			name:           "invalid source channel",
			channelVersion: opchildtypes.Version,
			setupOrigin: func(t *testing.T, ctx sdk.Context, input testutil.TestKeepers) {
				setupAttestorSetUpdateOrigin(t, ctx, input, testL1ClientID)
			},
			mutatePacket: func(packet *channeltypes.Packet) {
				packet.SourceChannel = "channel-99"
			},
			expErr: "expected source channel",
		},
		{
			name:           "invalid l1 client",
			channelVersion: opchildtypes.Version,
			setupOrigin: func(t *testing.T, ctx sdk.Context, input testutil.TestKeepers) {
				setupAttestorSetUpdateOrigin(t, ctx, input, "attacker-client-id")
			},
			expErr: "expected l1 client id",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, input := testutil.CreateTestInput(t, false)
			sdkCtx := sdk.UnwrapSDKContext(ctx)

			bridgeInfo := opchildtypes.BridgeInfo{
				BridgeId:   1,
				BridgeAddr: testutil.AddrsStr[0],
				L1ChainId:  "test-chain-1",
				L1ClientId: testL1ClientID,
				BridgeConfig: ophosttypes.BridgeConfig{
					ChannelId: testSourceChannel,
				},
			}
			err := input.OPChildKeeper.BridgeInfo.Set(ctx, bridgeInfo)
			require.NoError(t, err)
			tc.setupOrigin(t, sdkCtx, input)

			packetData := ophosttypes.NewAttestorSetUpdatePacketData(1, nil, 100)
			packet := channeltypes.Packet{
				SourcePort:         opchildtypes.PortID,
				SourceChannel:      testSourceChannel,
				DestinationPort:    opchildtypes.PortID,
				DestinationChannel: testDestChannel,
				Data:               packetData.GetBytes(),
				Sequence:           1,
			}
			if tc.mutatePacket != nil {
				tc.mutatePacket(&packet)
			}

			originErr := input.OPChildKeeper.ValidateAttestorSetUpdatePacketOrigin(sdkCtx, tc.channelVersion, packet)
			require.Error(t, originErr)
			require.Contains(t, originErr.Error(), tc.expErr)

			ibcModule := opchild.NewIBCModule(input.OPChildKeeper)
			ack := ibcModule.OnRecvPacket(sdkCtx, tc.channelVersion, packet, nil)
			require.False(t, ack.Success())
		})
	}
}

func Test_IBCModule_OnRecvPacket_InvalidData(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	bridgeInfo := opchildtypes.BridgeInfo{
		BridgeId:   1,
		BridgeAddr: testutil.AddrsStr[0],
		L1ChainId:  "test-chain-1",
		L1ClientId: testL1ClientID,
		BridgeConfig: ophosttypes.BridgeConfig{
			ChannelId: testSourceChannel,
		},
	}
	err := input.OPChildKeeper.BridgeInfo.Set(ctx, bridgeInfo)
	require.NoError(t, err)
	setupAttestorSetUpdateOrigin(t, sdkCtx, input, testL1ClientID)

	ibcModule := opchild.NewIBCModule(input.OPChildKeeper)

	packet := channeltypes.Packet{
		SourcePort:         opchildtypes.PortID,
		SourceChannel:      testSourceChannel,
		DestinationPort:    opchildtypes.PortID,
		DestinationChannel: testDestChannel,
		Data:               []byte("invalid-data"),
		Sequence:           1,
	}

	ack := ibcModule.OnRecvPacket(sdkCtx, opchildtypes.Version, packet, nil)
	require.False(t, ack.Success())
}

func Test_IBCModule_OnAcknowledgementPacket(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)
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

	err := ibcModule.OnAcknowledgementPacket(sdkCtx, opchildtypes.Version, packet, ack.Acknowledgement(), nil)
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
	ctx, input := testutil.CreateTestInput(t, false)
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

	err := ibcModule.OnTimeoutPacket(sdkCtx, opchildtypes.Version, packet, nil)
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
