package types

import (
	"encoding/json"
	"strings"
	time "time"

	"cosmossdk.io/core/address"
	"cosmossdk.io/errors"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (config BridgeConfig) Validate(ac address.Codec, vc address.Codec) error {
	if _, err := ac.StringToBytes(config.Challenger); err != nil {
		return err
	}

	if _, err := ac.StringToBytes(config.Proposer); err != nil {
		return err
	}

	if config.BatchInfo.ChainType == BatchInfo_UNSPECIFIED {
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

	if err := ValidateAttestorSet(config.AttestorSet, vc); err != nil {
		return err
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

	if config.BatchInfo.ChainType == BatchInfo_UNSPECIFIED {
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

	if err := ValidateAttestorSetNoAddrValidation(config.AttestorSet); err != nil {
		return err
	}

	return nil
}

// MarshalJSON marshals the BatchInfo_ChainType to JSON
func (cy BatchInfo_ChainType) MarshalJSON() ([]byte, error) {
	return json.Marshal(cy.String())
}

// UnmarshalJSON unmarshals the BatchInfo_ChainType from JSON
func (cy *BatchInfo_ChainType) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}

	chainType, ok := BatchInfo_ChainType_value[strings.ToUpper(str)]
	if !ok {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid chain type")
	}

	*cy = BatchInfo_ChainType(chainType)
	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (config BridgeConfig) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for i := range config.AttestorSet {
		if err := config.AttestorSet[i].UnpackInterfaces(unpacker); err != nil {
			return err
		}
	}
	return nil
}
