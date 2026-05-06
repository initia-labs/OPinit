package types

import (
	errors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	host "github.com/cosmos/ibc-go/v10/modules/core/24-host"
)

func (m MigrationInfo) Validate() error {
	if err := host.ChannelIdentifierValidator(m.IbcChannelId); err != nil {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	if err := host.PortIdentifierValidator(m.IbcPortId); err != nil {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	if err := sdk.ValidateDenom(m.Denom); err != nil {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	if m.BaseIbcDenomPath != "" {
		denomTrace := transfertypes.ExtractDenomFromPath(m.BaseIbcDenomPath)
		if denomTrace.IsNative() {
			return errors.Wrap(sdkerrors.ErrInvalidRequest, "base IBC denom path must include an IBC trace path")
		}
		if err := denomTrace.Validate(); err != nil {
			return errors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
		}
	}

	return nil
}
