package keeper

import (
	"context"
	"strconv"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"

	"github.com/initia-labs/OPinit/x/ophost/types"
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
		uint64(sdkCtx.BlockHeight()), //nolint:gosec
	)

	timeoutTimestamp := uint64(sdkCtx.BlockTime().Add(types.DefaultPacketTimeoutTimestamp).UnixNano()) //nolint:gosec

	_, err = k.channelKeeper.SendPacket(
		sdkCtx,
		sourcePort,
		sourceChannel,
		types.DefaultTransferPacketTimeoutHeight,
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
