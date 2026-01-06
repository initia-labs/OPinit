package ophost_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	"github.com/initia-labs/OPinit/x/ophost"
	"github.com/initia-labs/OPinit/x/ophost/testutil"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
)

func Test_OPHostIBCModule_OnChanOpenInit(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	ibcModule := ophost.NewIBCModule(input.OPHostKeeper)
	capability := &capabilitytypes.Capability{}

	// success case
	version, err := ibcModule.OnChanOpenInit(
		ctx,
		channeltypes.UNORDERED,
		[]string{"connection-0"},
		ophosttypes.PortID,
		"channel-0",
		capability,
		channeltypes.NewCounterparty(ophosttypes.PortID, "channel-1"),
		ophosttypes.Version,
	)
	require.NoError(t, err)
	require.Equal(t, ophosttypes.Version, version)

	// invalid ordering
	_, err = ibcModule.OnChanOpenInit(
		ctx,
		channeltypes.ORDERED,
		[]string{"connection-0"},
		ophosttypes.PortID,
		"channel-1",
		capability,
		channeltypes.NewCounterparty(ophosttypes.PortID, "channel-2"),
		ophosttypes.Version,
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected ORDER_UNORDERED channel")

	// invalid port
	_, err = ibcModule.OnChanOpenInit(
		ctx,
		channeltypes.UNORDERED,
		[]string{"connection-0"},
		"invalid-port",
		"channel-2",
		capability,
		channeltypes.NewCounterparty(ophosttypes.PortID, "channel-3"),
		ophosttypes.Version,
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid port")
}

func Test_OPHostIBCModule_OnChanCloseInit(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	ibcModule := ophost.NewIBCModule(input.OPHostKeeper)

	// user cannot close channel
	err := ibcModule.OnChanCloseInit(
		ctx,
		ophosttypes.PortID,
		"channel-0",
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "user cannot close channel")
}

func Test_OPHostIBCModule_OnRecvPacket(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	ibcModule := ophost.NewIBCModule(input.OPHostKeeper)

	packet := channeltypes.Packet{
		SourcePort:         ophosttypes.PortID,
		SourceChannel:      "channel-0",
		DestinationPort:    ophosttypes.PortID,
		DestinationChannel: "channel-1",
		Data:               []byte("test-data"),
		Sequence:           1,
	}

	// OPHost module on L1 should not receive packets
	ack := ibcModule.OnRecvPacket(ctx, packet, nil)
	require.False(t, ack.Success())
}

func Test_OPHostIBCModule_OnAcknowledgementPacket(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	ibcModule := ophost.NewIBCModule(input.OPHostKeeper)

	packet := channeltypes.Packet{
		SourcePort:         ophosttypes.PortID,
		SourceChannel:      "channel-0",
		DestinationPort:    ophosttypes.PortID,
		DestinationChannel: "channel-1",
		Data:               []byte("test-data"),
		Sequence:           1,
	}

	ack := channeltypes.NewResultAcknowledgement([]byte("success"))

	err := ibcModule.OnAcknowledgementPacket(ctx, packet, ack.Acknowledgement(), nil)
	require.NoError(t, err)
}

func Test_OPHostIBCModule_OnAcknowledgementPacket_Error(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	ibcModule := ophost.NewIBCModule(input.OPHostKeeper)

	packet := channeltypes.Packet{
		SourcePort:         ophosttypes.PortID,
		SourceChannel:      "channel-0",
		DestinationPort:    ophosttypes.PortID,
		DestinationChannel: "channel-1",
		Data:               []byte("test-data"),
		Sequence:           1,
	}

	ack := channeltypes.NewErrorAcknowledgement(fmt.Errorf("test error"))

	err := ibcModule.OnAcknowledgementPacket(ctx, packet, ack.Acknowledgement(), nil)
	require.NoError(t, err)

	events := ctx.EventManager().Events()
	found := false
	for _, event := range events {
		if event.Type == ophosttypes.EventTypePacket {
			for _, attr := range event.Attributes {
				if attr.Key == ophosttypes.AttributeKeyAckError {
					found = true
					break
				}
			}
		}
	}
	require.True(t, found, "expected ack error event to be emitted")
}

func Test_OPHostIBCModule_OnTimeoutPacket(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	ibcModule := ophost.NewIBCModule(input.OPHostKeeper)

	packet := channeltypes.Packet{
		SourcePort:         ophosttypes.PortID,
		SourceChannel:      "channel-0",
		DestinationPort:    ophosttypes.PortID,
		DestinationChannel: "channel-1",
		Data:               []byte("test-data"),
		Sequence:           1,
	}

	err := ibcModule.OnTimeoutPacket(ctx, packet, nil)
	require.NoError(t, err)
}
