package hook_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	"github.com/stretchr/testify/require"

	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
	"github.com/initia-labs/OPinit/x/ophost/types/hook"
)

var _ hook.ChannelKeeper = MockChannelKeeper{}
var _ hook.PermKeeper = MockPermKeeper{}

type MockChannelKeeper struct {
	sequenceIDs map[string]uint64
}

func (k MockChannelKeeper) GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool) {
	seq, ok := k.sequenceIDs[portID+"/"+channelID]
	return seq, ok
}

type MockPermissionRelayers struct {
	Relayers []string
}

func (l MockPermissionRelayers) HasRelayer(relayer string) bool {
	for _, r := range l.Relayers {
		if r == relayer {
			return true
		}
	}
	return false
}

type MockPermRelayerList struct {
	Relayers []sdk.AccAddress
}

func (l MockPermRelayerList) Equals(relayer sdk.AccAddress) bool {
	for _, r := range l.Relayers {
		if r.Equals(relayer) {
			return true
		}
	}
	return false
}

type MockPermKeeper struct {
	perms map[string]MockPermRelayerList
}

func (k MockPermKeeper) HasPermission(ctx context.Context, portID, channelID string, relayer sdk.AccAddress) (bool, error) {
	return k.perms[portID+"/"+channelID].Equals(relayer), nil
}

func (k MockPermKeeper) SetPermissionedRelayers(ctx context.Context, portID, channelID string, relayers []sdk.AccAddress) error {
	k.perms[portID+"/"+channelID] = MockPermRelayerList{Relayers: relayers}
	return nil
}

func (k MockPermKeeper) GetPermissionedRelayers(ctx context.Context, portID, channelID string) ([]sdk.AccAddress, error) {
	if _, ok := k.perms[portID+"/"+channelID]; !ok {
		return nil, collections.ErrNotFound
	}
	return k.perms[portID+"/"+channelID].Relayers, nil
}

func setup() (context.Context, hook.BridgeHook) {
	h := hook.NewBridgeHook(MockChannelKeeper{
		sequenceIDs: map[string]uint64{
			"transfer/channel-0": 1,
			"transfer/channel-1": 2,
			"transfer/channel-2": 1,
		},
	}, MockPermKeeper{
		perms: make(map[string]MockPermRelayerList),
	}, address.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()))

	ms := store.NewCommitMultiStore(dbm.NewMemDB(), log.NewNopLogger(), metrics.NewNoOpMetrics())
	ctx := sdk.NewContext(ms, tmproto.Header{
		Height: 1234567,
		Time:   time.Date(2020, time.April, 22, 12, 0, 0, 0, time.UTC),
	}, false, log.NewNopLogger())

	return ctx, h
}

func acc_addr() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()), sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())}
}

func Test_BridgeHook_BridgeCreated(t *testing.T) {
	ctx, h := setup()

	metadata, err := json.Marshal(hook.PermsMetadata{
		PermChannels: []hook.PortChannelID{
			{
				PortID:    "transfer",
				ChannelID: "channel-0",
			},
		},
	})
	require.NoError(t, err)

	metadata2, err := json.Marshal(hook.PermsMetadata{
		PermChannels: []hook.PortChannelID{
			{
				PortID:    "transfer",
				ChannelID: "channel-1",
			},
		},
	})
	require.NoError(t, err)

	addr := acc_addr()
	err = h.BridgeCreated(ctx, 1, ophosttypes.BridgeConfig{
		Challengers: []string{addr[0].String()},
		Metadata:    metadata,
	})
	require.NoError(t, err)

	// can't create bridge with already taken channel
	err = h.BridgeCreated(ctx, 2, ophosttypes.BridgeConfig{
		Challengers: []string{addr[0].String()},
		Metadata:    metadata,
	})
	require.ErrorIs(t, err, channeltypes.ErrChannelExists)

	// cannot take non-1 sequence channel
	err = h.BridgeCreated(ctx, 1, ophosttypes.BridgeConfig{
		Challengers: []string{addr[0].String()},
		Metadata:    metadata2,
	})
	require.ErrorIs(t, err, channeltypes.ErrChannelExists)

	// check permission is applied
	ok, err := h.IBCPermKeeper.HasPermission(ctx, "transfer", "channel-0", addr[0])
	require.NoError(t, err)
	require.True(t, ok)

	// check permission is applied
	ok, err = h.IBCPermKeeper.HasPermission(ctx, "transfer", "channel-0", addr[1])
	require.NoError(t, err)
	require.False(t, ok)
}

func Test_BridgeHook_ChallengersUpdated(t *testing.T) {
	ctx, h := setup()

	metadata, err := json.Marshal(hook.PermsMetadata{
		PermChannels: []hook.PortChannelID{
			{
				PortID:    "transfer",
				ChannelID: "channel-0",
			},
		},
	})
	require.NoError(t, err)

	addr := acc_addr()
	err = h.BridgeCreated(ctx, 1, ophosttypes.BridgeConfig{
		Challengers: []string{addr[0].String()},
		Metadata:    metadata,
	})
	require.NoError(t, err)

	newAddr := acc_addr()
	err = h.BridgeChallengersUpdated(ctx, 1, ophosttypes.BridgeConfig{
		Challengers: []string{newAddr[0].String(), newAddr[1].String()},
		Metadata:    metadata,
	})
	require.NoError(t, err)

	// check permission is applied
	ok, err := h.IBCPermKeeper.HasPermission(ctx, "transfer", "channel-0", newAddr[0])
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = h.IBCPermKeeper.HasPermission(ctx, "transfer", "channel-0", newAddr[1])
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = h.IBCPermKeeper.HasPermission(ctx, "transfer", "channel-0", addr[0])
	require.NoError(t, err)
	require.False(t, ok)
}

func Test_BridgeHook_MetadataUpdated(t *testing.T) {
	ctx, h := setup()

	metadata, err := json.Marshal(hook.PermsMetadata{
		PermChannels: []hook.PortChannelID{
			{
				PortID:    "transfer",
				ChannelID: "channel-0",
			},
		},
	})
	require.NoError(t, err)

	addr := acc_addr()
	err = h.BridgeCreated(ctx, 1, ophosttypes.BridgeConfig{
		Challengers: []string{addr[0].String()},
		Metadata:    metadata,
	})
	require.NoError(t, err)

	// new metadata
	metadata, err = json.Marshal(hook.PermsMetadata{
		PermChannels: []hook.PortChannelID{
			{
				PortID:    "transfer",
				ChannelID: "channel-0",
			},
			{
				PortID:    "transfer",
				ChannelID: "channel-1",
			},
		},
	})
	require.NoError(t, err)

	// cannot take non-1 sequence channel
	err = h.BridgeMetadataUpdated(ctx, 1, ophosttypes.BridgeConfig{
		Challengers: []string{addr[0].String(), addr[1].String()},
		Metadata:    metadata,
	})
	require.Error(t, err)

	// new metadata
	metadata, err = json.Marshal(hook.PermsMetadata{
		PermChannels: []hook.PortChannelID{
			{
				PortID:    "transfer",
				ChannelID: "channel-0",
			},
			{
				PortID:    "transfer",
				ChannelID: "channel-2",
			},
		},
	})
	require.NoError(t, err)

	err = h.BridgeMetadataUpdated(ctx, 1, ophosttypes.BridgeConfig{
		Challengers: []string{addr[0].String(), addr[1].String()},
		Metadata:    metadata,
	})
	require.NoError(t, err)

	// check permission is applied
	ok, err := h.IBCPermKeeper.HasPermission(ctx, "transfer", "channel-2", addr[0])
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = h.IBCPermKeeper.HasPermission(ctx, "transfer", "channel-1", addr[0])
	require.NoError(t, err)
	require.False(t, ok)
}
