package types

import (
	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
)

const (
	DefaultBridgeIdStart   = 1
	DefaultL1SequenceStart = 1
)

// NewGenesisState creates a new GenesisState instance
func NewGenesisState(params Params, bridges []Bridge, nextBridgeId uint64, migrationInfos []MigrationInfo, portID string) *GenesisState {
	return &GenesisState{
		Params:         params,
		Bridges:        bridges,
		NextBridgeId:   nextBridgeId,
		MigrationInfos: migrationInfos,
		PortId:         portID,
	}
}

// DefaultGenesisState gets the raw genesis raw message for testing
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:         DefaultParams(),
		Bridges:        []Bridge{},
		NextBridgeId:   DefaultBridgeIdStart,
		MigrationInfos: []MigrationInfo{},
		PortId:         PortID,
	}
}

// ValidateGenesis performs basic validation of rollup genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data *GenesisState, ac address.Codec, vc address.Codec) error {
	for _, bridge := range data.Bridges {
		if err := bridge.BridgeConfig.Validate(ac, vc); err != nil {
			return err
		}

		if bridge.BridgeId == 0 {
			return ErrInvalidBridgeId
		}

		if bridge.NextL1Sequence < DefaultL1SequenceStart {
			return ErrInvalidSequence
		}

		for _, tokenPair := range bridge.TokenPairs {
			if err := sdk.ValidateDenom(tokenPair.L1Denom); err != nil {
				return err
			}
			if err := sdk.ValidateDenom(tokenPair.L2Denom); err != nil {
				return err
			}
		}

		for _, withdrawalHash := range bridge.ProvenWithdrawals {
			if len(withdrawalHash) != 32 {
				return ErrInvalidHashLength
			}
		}

		for _, proposal := range bridge.Proposals {
			if proposal.OutputIndex == 0 {
				return ErrInvalidOutputIndex
			}

			if err := proposal.OutputProposal.Validate(); err != nil {
				return err
			}
		}

		if len(bridge.BatchInfos) == 0 {
			return ErrEmptyBatchInfo
		}

		// last batchinfo should be same with bridgeconfig batchinfo
		if bridge.BatchInfos[len(bridge.BatchInfos)-1].BatchInfo != bridge.BridgeConfig.BatchInfo ||
			!bridge.BatchInfos[0].Output.IsEmpty() {
			return ErrInvalidBatchInfo
		}
	}

	if data.NextBridgeId < DefaultBridgeIdStart {
		return ErrInvalidBridgeId
	}

	for _, migrationInfo := range data.MigrationInfos {
		if err := migrationInfo.Validate(); err != nil {
			return err
		}
	}

	if err := host.PortIdentifierValidator(data.PortId); err != nil {
		return err
	}

	return data.Params.Validate()
}
