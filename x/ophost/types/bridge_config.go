package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (config BridgeConfig) Validate() error {
	if _, err := sdk.AccAddressFromBech32(config.Challenger); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(config.Proposer); err != nil {
		return err
	}

	return nil
}
