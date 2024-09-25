package types

import (
	"encoding/json"
	"strings"
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

	if config.BatchInfo.ChainType == BatchInfo_CHAIN_TYPE_UNSPECIFIED {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "batch chain type must be set")
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

	if config.SubmissionStartHeight == 0 {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "submission start height must be set")
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

	if config.BatchInfo.ChainType == BatchInfo_CHAIN_TYPE_UNSPECIFIED {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "batch chain type must be set")
	} else if _, ok := BatchInfo_ChainType_name[int32(config.BatchInfo.ChainType)]; !ok {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid chain type")
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

	if config.SubmissionStartHeight == 0 {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "submission start height must be set")
	}

	return nil
}

// prefix for chain type enum
const chainTypePrefix = "CHAIN_TYPE_"

// MarshalJSON marshals the BatchInfo_ChainType to JSON
func (cy BatchInfo_ChainType) MarshalJSON() ([]byte, error) {
	return json.Marshal(cy.StringWithoutPrefix())
}

// UnmarshalJSON unmarshals the BatchInfo_ChainType from JSON
func (cy *BatchInfo_ChainType) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}

	chainType, ok := BatchInfo_ChainType_value[chainTypePrefix+strings.ToUpper(str)]
	if !ok {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid chain type")
	}

	*cy = BatchInfo_ChainType(chainType)
	return nil
}

// StringWithoutPrefix returns the string representation of a BatchInfo_ChainType without the prefix
func (cy BatchInfo_ChainType) StringWithoutPrefix() string {
	return cy.String()[len(chainTypePrefix):]
}
