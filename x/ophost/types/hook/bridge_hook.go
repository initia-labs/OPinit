package hook

import (
	"context"
	"errors"

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
	HasPermission(ctx context.Context, portID, channelID string, relayer sdk.AccAddress) (bool, error)
	SetPermissionedRelayers(ctx context.Context, portID, channelID string, relayers []sdk.AccAddress) error
	GetPermissionedRelayers(ctx context.Context, portID, channelID string) ([]sdk.AccAddress, error)
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

	// TODO (reviewer): This is a temporary solution to allow only one challenger on bridge creation.
	if len(bridgeConfig.Challengers) != 1 {
		return errors.New("bridge must have exactly one challenger on creation")
	}

	challengers := make([]sdk.AccAddress, len(bridgeConfig.Challengers))
	for i, challenger := range bridgeConfig.Challengers {
		challengerAddr, err := h.ac.StringToBytes(challenger)
		if err != nil {
			return err
		}

		challengers[i] = challengerAddr
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	for _, permChannel := range metadata.PermChannels {
		portID, channelID := permChannel.PortID, permChannel.ChannelID
		if seq, ok := h.IBCChannelKeeper.GetNextSequenceSend(sdkCtx, portID, channelID); !ok {
			return channeltypes.ErrChannelNotFound.Wrap("failed to get next sequence send")
		} else if seq != 1 {
			return channeltypes.ErrChannelExists.Wrap("cannot register permissioned relayers for the channel in use")
		}

		// check if the channel has a permissioned relayer
		if relayers, err := h.IBCPermKeeper.GetPermissionedRelayers(ctx, portID, channelID); err != nil {
			return err
		} else if len(relayers) > 0 {
			return channeltypes.ErrChannelExists.Wrap("cannot register permissioned relayers for the channel in use")
		}

		// register challengers as channel relayer
		if err := h.IBCPermKeeper.SetPermissionedRelayers(sdkCtx, portID, channelID, challengers); err != nil {
			return err
		}
	}

	return nil
}

func (h BridgeHook) BridgeChallengersUpdated(
	ctx context.Context,
	bridgeId uint64,
	bridgeConfig ophosttypes.BridgeConfig,
) error {
	hasPermChannels, metadata := hasPermChannels(bridgeConfig.Metadata)
	if !hasPermChannels {
		return nil
	}

	challengers := make([]sdk.AccAddress, len(bridgeConfig.Challengers))
	for i, challenger := range bridgeConfig.Challengers {
		challengerAddr, err := h.ac.StringToBytes(challenger)
		if err != nil {
			return err
		}

		challengers[i] = challengerAddr
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	for _, permChannel := range metadata.PermChannels {
		portID, channelID := permChannel.PortID, permChannel.ChannelID
		// register challengers as channel relayers
		if err := h.IBCPermKeeper.SetPermissionedRelayers(sdkCtx, portID, channelID, challengers); err != nil {
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

	challengers := make([]sdk.AccAddress, len(bridgeConfig.Challengers))
	for i, challenger := range bridgeConfig.Challengers {
		challengerAddr, err := h.ac.StringToBytes(challenger)
		if err != nil {
			return err
		}

		challengers[i] = challengerAddr
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	for _, permChannel := range metadata.PermChannels {
		portID, channelID := permChannel.PortID, permChannel.ChannelID

		hasPermission := true

		// check if the challengers are already registered as a permissioned relayers
		for _, challenger := range challengers {
			if hasPerm, err := h.IBCPermKeeper.HasPermission(ctx, portID, channelID, challenger); err != nil {
				return err
			} else if !hasPerm {
				hasPermission = false
				break
			}
		}

		if hasPermission {
			continue
		}

		if seq, ok := h.IBCChannelKeeper.GetNextSequenceSend(sdkCtx, portID, channelID); !ok {
			return channeltypes.ErrChannelNotFound.Wrap("failed to register permissioned relayer")
		} else if seq != 1 {
			return channeltypes.ErrChannelExists.Wrap("cannot register permissioned relayer for the channel in use")
		}

		// register challengers as channel relayers
		if err := h.IBCPermKeeper.SetPermissionedRelayers(sdkCtx, portID, channelID, challengers); err != nil {
			return err
		}
	}

	return nil
}
