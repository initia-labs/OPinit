package types

import (
	"cosmossdk.io/core/address"
)

func (config BridgeConfig) Validate(ac address.Codec) error {
	if _, err := ac.StringToBytes(config.Challenger); err != nil {
		return err
	}

	if _, err := ac.StringToBytes(config.Proposer); err != nil {
		return err
	}

	return nil
}
