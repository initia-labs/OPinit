package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/initia-labs/OPinit/x/opchild/types"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
)

// OnRecvAttestorSetUpdatePacket is called when an attestor set update packet is received via IBC.
func (k Keeper) OnRecvAttestorSetUpdatePacket(
	ctx context.Context,
	packet []byte,
) ([]byte, error) {
	data, err := types.DecodePacketData(packet)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "failed to unmarshal attestor set update packet")
	}

	ack, err := k.HandleAttestorSetUpdatePacket(ctx, data)
	if err != nil {
		return nil, err
	}

	ackBytes, err := json.Marshal(ack)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrJSONMarshal, "failed to marshal acknowledgement")
	}

	return ackBytes, nil
}

// HandleAttestorSetUpdatePacket handles the attestor set update packet from L1.
func (k Keeper) HandleAttestorSetUpdatePacket(
	ctx context.Context,
	packet ophosttypes.AttestorSetUpdatePacketData,
) (ophosttypes.AttestorSetUpdatePacketAck, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	bridgeInfo, err := k.BridgeInfo.Get(ctx)
	if err != nil {
		return ophosttypes.AttestorSetUpdatePacketAck{
			Success: false,
			Error:   fmt.Sprintf("failed to get bridge info: %v", err),
		}, nil
	}

	if packet.BridgeId != bridgeInfo.BridgeId {
		return ophosttypes.AttestorSetUpdatePacketAck{
			Success: false,
			Error:   fmt.Sprintf("bridge ID mismatch: expected %d, got %d", bridgeInfo.BridgeId, packet.BridgeId),
		}, nil
	}

	if err := k.updateAttestorsFromPacket(ctx, packet.AttestorSet); err != nil {
		return ophosttypes.AttestorSetUpdatePacketAck{
			Success: false,
			Error:   fmt.Sprintf("failed to update attestors: %v", err),
		}, nil
	}

	bridgeInfo.BridgeConfig.AttestorSet = packet.AttestorSet
	if err := k.BridgeInfo.Set(ctx, bridgeInfo); err != nil {
		return ophosttypes.AttestorSetUpdatePacketAck{
			Success: false,
			Error:   fmt.Sprintf("failed to update attestor set in bridge config: %v", err),
		}, nil
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAttestorSetUpdate,
			sdk.NewAttribute(types.AttributeKeyBridgeId, fmt.Sprintf("%d", packet.BridgeId)),
			sdk.NewAttribute(types.AttributeKeyL1BlockHeight, fmt.Sprintf("%d", packet.L1BlockHeight)),
			sdk.NewAttribute(types.AttributeKeyAttestorSetSize, fmt.Sprintf("%d", len(packet.AttestorSet))),
		),
	)

	return ophosttypes.AttestorSetUpdatePacketAck{Success: true}, nil
}

// updateAttestorsFromPacket updates the attestor validators from IBC packet data.
func (k Keeper) updateAttestorsFromPacket(ctx context.Context, attestorSet []ophosttypes.Attestor) error {
	validators, err := k.GetAllValidators(ctx)
	if err != nil {
		return err
	}

	newAttestors := make(map[string]ophosttypes.Attestor)
	for _, attestor := range attestorSet {
		newAttestors[attestor.OperatorAddress] = attestor
	}

	existingAttestors := make(map[string]types.Validator)
	for _, validator := range validators {
		if validator.ConsPower == types.AttestorConsPower {
			existingAttestors[validator.OperatorAddress] = validator
		}
	}

	// remove attestors that are no longer in the new set
	for opAddr := range existingAttestors {
		if _, stillExists := newAttestors[opAddr]; !stillExists {
			vAddr, err := k.validatorAddressCodec.StringToBytes(opAddr)
			if err != nil {
				return errorsmod.Wrapf(err, "invalid existing attestor address: %s", opAddr)
			}

			if err := k.RemoveValidatorByAddress(ctx, vAddr); err != nil {
				return errorsmod.Wrapf(err, "failed to remove attestor: %s", opAddr)
			}
		}
	}

	// add new attestors that weren't in the old set
	for i, attestorData := range attestorSet {
		// skip if already exists
		if _, exists := existingAttestors[attestorData.OperatorAddress]; exists {
			continue
		}

		operatorAddr, err := k.validatorAddressCodec.StringToBytes(attestorData.OperatorAddress)
		if err != nil {
			return errorsmod.Wrapf(err, "invalid attestor address at index %d", i)
		}

		var pubkey cryptotypes.PubKey
		if err := k.cdc.UnpackAny(attestorData.ConsensusPubkey, &pubkey); err != nil {
			return errorsmod.Wrapf(err, "failed to unpack attestor pubkey at index %d", i)
		}

		if err := k.AddValidatorWithPower(ctx, attestorData.Moniker, operatorAddr, pubkey, types.AttestorConsPower); err != nil {
			return errorsmod.Wrapf(err, "failed to add attestor: %s", attestorData.OperatorAddress)
		}
	}

	return nil
}
