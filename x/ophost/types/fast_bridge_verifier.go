package types

import (
	"cosmossdk.io/errors"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ FastBridgeVerifierI = FastBridgeVerifier{}

// NewFastBridgeVerifier constructs a new Fast Bridge verifier
func NewFastBridgeVerifier(address sdk.Address, pubKey cryptotypes.PubKey) (FastBridgeVerifier, error) {
	pkAddress := sdk.AccAddress(pubKey.Address())
	if !address.Equals(pkAddress) {
		return FastBridgeVerifier{}, sdkerrors.ErrInvalidPubKey.Wrapf("mismatch pubkey address; expected %s, got %s", address, pkAddress)
	}

	pkAny, err := codectypes.NewAnyWithValue(pubKey)
	if err != nil {
		return FastBridgeVerifier{}, err
	}

	return FastBridgeVerifier{
		Address: address.String(),
		Pubkey:  pkAny,
	}, nil
}

func (fv FastBridgeVerifier) GetPubKey() (cryptotypes.PubKey, error) {
	pk, ok := fv.Pubkey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidType, "expecting cryptotypes.PubKey, got %T", pk)
	}

	return pk, nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (fv FastBridgeVerifier) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pk cryptotypes.PubKey
	return unpacker.UnpackAny(fv.Pubkey, &pk)
}
