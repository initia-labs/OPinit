package types

import (
	"slices"
	time "time"

	"cosmossdk.io/core/address"
	"cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (config BridgeConfig) Validate(ac address.Codec) error {
	challengerDupMap := make(map[string]bool, len(config.Challengers))
	for _, challenger := range config.Challengers {
		if _, err := ac.StringToBytes(challenger); err != nil {
			return err
		}

		if _, ok := challengerDupMap[challenger]; ok {
			return errors.Wrapf(sdkerrors.ErrInvalidRequest, "challengers must be unique")
		}

		challengerDupMap[challenger] = true
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

	if !config.isValidChallengers() {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "challengers must be non-empty array")
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

	if !config.isValidChallengers() {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "challengers must be non-empty array")
	}

	if len(config.BatchInfo.Chain) == 0 {
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

func (config BridgeConfig) isValidChallengers() bool {
	if len(config.Challengers) == 0 {
		return false
	}

	if slices.Contains(config.Challengers, "") {
		return false
	}

	return true
}
