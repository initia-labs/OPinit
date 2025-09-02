package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
)

func (m MigrationInfo) Validate() error {
	if m.BridgeId == 0 {
		return ErrInvalidBridgeId
	}

	if err := host.ChannelIdentifierValidator(m.IbcChannelId); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
	}

	if err := host.PortIdentifierValidator(m.IbcPortId); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
	}

	if err := sdk.ValidateDenom(m.L1Denom); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
	}

	return nil
}
