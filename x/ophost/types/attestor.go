package types

import (
	"cosmossdk.io/core/address"
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ValidateAttestor validates a single attestor with address validation
func ValidateAttestor(attestor Attestor, vc address.Codec) error {
	if _, err := vc.StringToBytes(attestor.OperatorAddress); err != nil {
		return errors.Wrapf(err, "invalid attestor address")
	}

	if attestor.ConsensusPubkey == nil {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "consensus pubkey cannot be nil")
	}

	cachedValue := attestor.ConsensusPubkey.GetCachedValue()
	if cachedValue == nil {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "cannot find cached value for consensus pubkey")
	}

	pubkey, ok := cachedValue.(cryptotypes.PubKey)
	if !ok || pubkey == nil {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid consensus pubkey")
	}

	return nil
}

// ValidateAttestorNoAddrValidation validates a single attestor without address validation
func ValidateAttestorNoAddrValidation(attestor Attestor) error {
	if len(attestor.OperatorAddress) == 0 {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "attestor address cannot be empty")
	}

	if attestor.ConsensusPubkey == nil {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "consensus pubkey cannot be nil")
	}

	return nil
}

// ValidateAttestorSet validates a set of attestors with address validation
func ValidateAttestorSet(attestorSet []Attestor, vc address.Codec) error {
	seenAddrs := make(map[string]bool)
	seenPubkeys := make(map[string]bool)

	for i, attestor := range attestorSet {
		if err := ValidateAttestor(attestor, vc); err != nil {
			return errors.Wrapf(err, "invalid attestor at index %d", i)
		}

		// check for duplicates
		if seenAddrs[attestor.OperatorAddress] {
			return errors.Wrapf(sdkerrors.ErrInvalidRequest, "duplicate attestor address: %s", attestor.OperatorAddress)
		}
		seenAddrs[attestor.OperatorAddress] = true

		cachedValue := attestor.ConsensusPubkey.GetCachedValue()
		if pubkey, ok := cachedValue.(cryptotypes.PubKey); ok && pubkey != nil {
			pubkeyStr := pubkey.String()
			if seenPubkeys[pubkeyStr] {
				return errors.Wrapf(sdkerrors.ErrInvalidRequest, "duplicate consensus pubkey at index %d", i)
			}
			seenPubkeys[pubkeyStr] = true
		}
	}

	return nil
}

// ValidateAttestorSetNoAddrValidation validates a set of attestors without address validation
func ValidateAttestorSetNoAddrValidation(attestorSet []Attestor) error {
	seenAddrs := make(map[string]bool)
	seenPubkeys := make(map[string]bool)

	for i, attestor := range attestorSet {
		if err := ValidateAttestorNoAddrValidation(attestor); err != nil {
			return errors.Wrapf(err, "invalid attestor at index %d", i)
		}

		// check for duplicates
		if seenAddrs[attestor.OperatorAddress] {
			return errors.Wrapf(sdkerrors.ErrInvalidRequest, "duplicate attestor address: %s", attestor.OperatorAddress)
		}
		seenAddrs[attestor.OperatorAddress] = true

		pubkeyKey := attestor.ConsensusPubkey.TypeUrl + string(attestor.ConsensusPubkey.Value)
		if seenPubkeys[pubkeyKey] {
			return errors.Wrapf(sdkerrors.ErrInvalidRequest, "duplicate consensus pubkey at index %d", i)
		}
		seenPubkeys[pubkeyKey] = true
	}

	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (a Attestor) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pk cryptotypes.PubKey
	return unpacker.UnpackAny(a.ConsensusPubkey, &pk)
}

// NewAttestorSetUpdatePacketData constructs a new AttestorSetUpdatePacketData instance
func NewAttestorSetUpdatePacketData(
	bridgeId uint64, attestorSet []Attestor, l1BlockHeight uint64,
) AttestorSetUpdatePacketData {
	return AttestorSetUpdatePacketData{
		BridgeId:      bridgeId,
		AttestorSet:   attestorSet,
		L1BlockHeight: l1BlockHeight,
	}
}

// GetBytes is a helper for serializing AttestorSetUpdatePacketData
func (apd AttestorSetUpdatePacketData) GetBytes() []byte {
	return sdk.MustSortJSON(mustProtoMarshalJSON(&apd))
}
