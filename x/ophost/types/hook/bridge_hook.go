package hook

import (
	"context"

	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"

	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
)

var _ ophosttypes.BridgeHook = BridgeHook{}

type BridgeHook struct {
	IBCChannelKeeper ChannelKeeper
	IBCPermKeeper    PermKeeper
	ac               address.Codec
}

type ChannelKeeper interface {
	GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool)
}

type PermKeeper interface {
	IsTaken(ctx context.Context, portID, channelID string) (bool, error)

	SetAdmin(ctx context.Context, portID, channelID string, admin sdk.AccAddress) error
	HasAdminPermission(ctx context.Context, portID, channelID string, admin sdk.AccAddress) (bool, error)
}

func NewBridgeHook(channelKeeper ChannelKeeper, permKeeper PermKeeper, ac address.Codec) BridgeHook {
	return BridgeHook{channelKeeper, permKeeper, ac}
}

func (h BridgeHook) BridgeCreated(
	ctx context.Context,
	bridgeId uint64,
	bridgeConfig ophosttypes.BridgeConfig,
) error {
	hasPermChannels, metadata := hasPermChannels(bridgeConfig.Metadata)
	if !hasPermChannels {
		return nil
	}

	challenger, err := h.ac.StringToBytes(bridgeConfig.Challenger)
	if err != nil {
		return err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	for _, permChannel := range metadata.PermChannels {
		portID, channelID := permChannel.PortID, permChannel.ChannelID

		// register challenger as channel admin
		if err := h.registerChannelAdmin(sdkCtx, portID, channelID, challenger); err != nil {
			return err
		}
	}

	return nil
}

func (h BridgeHook) BridgeChallengerUpdated(
	ctx context.Context,
	bridgeId uint64,
	bridgeConfig ophosttypes.BridgeConfig,
) error {
	hasPermChannels, metadata := hasPermChannels(bridgeConfig.Metadata)
	if !hasPermChannels {
		return nil
	}

	challenger, err := h.ac.StringToBytes(bridgeConfig.Challenger)
	if err != nil {
		return err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	for _, permChannel := range metadata.PermChannels {
		portID, channelID := permChannel.PortID, permChannel.ChannelID

		// update channel admin to the new challenger
		if err := h.IBCPermKeeper.SetAdmin(sdkCtx, portID, channelID, challenger); err != nil {
			return err
		}
	}

	return nil
}

func (h BridgeHook) BridgeProposerUpdated(
	ctx context.Context,
	bridgeId uint64,
	bridgeConfig ophosttypes.BridgeConfig,
) error {
	return nil
}

// BridgeBatchInfoUpdated implements types.BridgeHook.
func (h BridgeHook) BridgeBatchInfoUpdated(
	ctx context.Context,
	bridgeId uint64,
	bridgeConfig ophosttypes.BridgeConfig,
) error {
	return nil
}

func (h BridgeHook) BridgeMetadataUpdated(
	ctx context.Context,
	bridgeId uint64,
	bridgeConfig ophosttypes.BridgeConfig,
) error {
	hasPermChannels, metadata := hasPermChannels(bridgeConfig.Metadata)
	if !hasPermChannels {
		return nil
	}

	challenger, err := h.ac.StringToBytes(bridgeConfig.Challenger)
	if err != nil {
		return err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	for _, permChannel := range metadata.PermChannels {
		portID, channelID := permChannel.PortID, permChannel.ChannelID

		// check if the challenger is already registered as a channel admin
		if hasPerm, err := h.IBCPermKeeper.HasAdminPermission(ctx, portID, channelID, challenger); err != nil {
			return err
		} else if hasPerm {
			continue
		}

		// register challenger as channel admin
		if err := h.registerChannelAdmin(sdkCtx, portID, channelID, challenger); err != nil {
			return err
		}
	}

	return nil
}

func (h BridgeHook) registerChannelAdmin(
	ctx sdk.Context,
	portID, channelID string,
	admin sdk.AccAddress,
) error {
	if seq, ok := h.IBCChannelKeeper.GetNextSequenceSend(ctx, portID, channelID); !ok {
		return channeltypes.ErrChannelNotFound.Wrap("failed to register ibcperm admin")
	} else if seq != 1 {
		return channeltypes.ErrChannelExists.Wrap("cannot register ibcperm admin for the channel in use")
	}

	// check if the channel has a admin already
	if taken, err := h.IBCPermKeeper.IsTaken(ctx, portID, channelID); err != nil {
		return err
	} else if taken {
		return channeltypes.ErrChannelExists.Wrap("cannot register ibcperm admin for the channel in use")
	}

	// register challenger as channel admin
	if err := h.IBCPermKeeper.SetAdmin(ctx, portID, channelID, admin); err != nil {
		return err
	}

	return nil
}
