package keeper

import (
	"context"
	"strconv"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"

	"github.com/initia-labs/OPinit/x/ophost/types"
)

var (
	// DefaultTransferPacketTimeoutHeight is the timeout height following IBC defaults
	DefaultTransferPacketTimeoutHeight = clienttypes.Height{
		RevisionNumber: 0,
		RevisionHeight: 0,
	}

	// DefaultPacketTimeoutTimestamp is the default packet timeout timestamp (in nanoseconds)
	// The default is currently set to a 10-minute timeout.
	DefaultPacketTimeoutTimestamp = time.Duration((10 * time.Minute).Nanoseconds())
)

// SendAttestorSetUpdatePacket sends an attestor set update packet to L2 via IBC.
func (k Keeper) SendAttestorSetUpdatePacket(
	ctx context.Context,
	bridgeId uint64,
	sourcePort, sourceChannel string,
) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	config, err := k.GetBridgeConfig(ctx, bridgeId)
	if err != nil {
		return errorsmod.Wrap(err, "failed to get bridge config")
	}

	packetData := types.NewAttestorSetUpdatePacketData(
		bridgeId,
		config.AttestorSet,
		uint64(sdkCtx.BlockHeight()),
	)

	channelCap, ok := k.scopedKeeper.GetCapability(sdkCtx, host.ChannelCapabilityPath(sourcePort, sourceChannel))
	if !ok {
		return errorsmod.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	timeoutTimestamp := uint64(sdkCtx.BlockTime().Add(DefaultPacketTimeoutTimestamp).UnixNano())

	_, err = k.channelKeeper.SendPacket(
		sdkCtx,
		channelCap,
		sourcePort,
		sourceChannel,
		DefaultTransferPacketTimeoutHeight,
		timeoutTimestamp,
		packetData.GetBytes(),
	)
	if err != nil {
		return errorsmod.Wrap(err, "failed to send IBC packet")
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAttestorSetPacketSent,
			sdk.NewAttribute(types.AttributeKeyBridgeId, strconv.FormatUint(bridgeId, 10)),
			sdk.NewAttribute(types.AttributeKeyL1BlockHeight, strconv.FormatInt(sdkCtx.BlockHeight(), 10)),
			sdk.NewAttribute(channeltypes.AttributeKeySrcChannel, sourceChannel),
		),
	)

	return nil
}
