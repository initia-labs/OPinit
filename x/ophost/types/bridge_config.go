package types

import (
	time "time"

	"cosmossdk.io/core/address"
	"cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (config BridgeConfig) Validate(ac address.Codec) error {
	if _, err := ac.StringToBytes(config.Challenger); err != nil {
		return err
	}

	if _, err := ac.StringToBytes(config.Proposer); err != nil {
		return err
	}

	if config.BatchInfo.Chain == "" {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "batch chain must be set")
	}

	if config.BatchInfo.Submitter == "" {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "batch submitter must be set")
	}

	if config.FinalizationPeriod == time.Duration(0) {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "finalization period must be greater than 0")
	}

	if config.SubmissionInterval == time.Duration(0) {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "submission interval must be greater than 0")
	}

	if config.SubmissionStartTime.IsZero() {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "submission start time must be set")
	}

	return nil
}

func (config BridgeConfig) ValidateWithNoAddrValidation() error {
	if len(config.Proposer) == 0 {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "proposer must be set")
	}

	if len(config.Challenger) == 0 {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "challenger must be set")
	}

	if len(config.BatchInfo.Chain) == 0 {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "batch chain must be set")
	}

	if len(config.BatchInfo.Submitter) == 0 {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "batch submitter must be set")
	}

	if config.FinalizationPeriod == time.Duration(0) {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "finalization period must be greater than 0")
	}

	if config.SubmissionInterval == time.Duration(0) {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "submission interval must be greater than 0")
	}

	if config.SubmissionStartTime.IsZero() {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "submission start time must be set")
	}

	return nil
}
