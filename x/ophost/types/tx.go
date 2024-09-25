package types

import (
	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgRecordBatch{}
	_ sdk.Msg = &MsgCreateBridge{}
	_ sdk.Msg = &MsgProposeOutput{}
	_ sdk.Msg = &MsgDeleteOutput{}
	_ sdk.Msg = &MsgFinalizeTokenWithdrawal{}
	_ sdk.Msg = &MsgInitiateTokenDeposit{}
	_ sdk.Msg = &MsgUpdateProposer{}
	_ sdk.Msg = &MsgUpdateChallenger{}
	_ sdk.Msg = &MsgUpdateBatchInfo{}
	_ sdk.Msg = &MsgUpdateMetadata{}
	_ sdk.Msg = &MsgUpdateParams{}
)

const (
	MaxMetadataLength = 1024 * 5
)

/* MsgRecordBatch */

// NewMsgRecordBatch creates a new MsgRecordBatch instance.
func NewMsgRecordBatch(
	submitter string,
	bridgeId uint64,
	batchBytes []byte,
) *MsgRecordBatch {
	return &MsgRecordBatch{
		Submitter:  submitter,
		BridgeId:   bridgeId,
		BatchBytes: batchBytes,
	}
}

// Validate performs basic MsgRecordBatch message validation.
func (msg MsgRecordBatch) Validate(ac address.Codec) error {
	if _, err := ac.StringToBytes(msg.Submitter); err != nil {
		return err
	}

	if msg.BridgeId == 0 {
		return ErrInvalidBridgeId
	}

	if len(msg.BatchBytes) == 0 {
		return ErrEmptyBatchBytes
	}

	return nil
}

/* MsgCreateBridge */

// NewMsgCreateBridge creates a new MsgCreateBridge instance.
func NewMsgCreateBridge(
	creator string,
	config BridgeConfig,
) *MsgCreateBridge {
	return &MsgCreateBridge{
		Creator: creator,
		Config:  config,
	}
}

// Validate performs basic MsgCreateBridge message validation.
func (msg MsgCreateBridge) Validate(ac address.Codec) error {
	if _, err := ac.StringToBytes(msg.Creator); err != nil {
		return err
	}

	if err := msg.Config.Validate(ac); err != nil {
		return err
	}

	if len(msg.Config.Metadata) > MaxMetadataLength {
		return ErrInvalidBridgeMetadata.Wrapf("metadata length exceeds %d", MaxMetadataLength)
	}

	return nil
}

/* MsgProposeOutput */

// NewMsgProposeOutput creates a new MsgProposeOutput instance.
// Delegator address and validator address are the same.
func NewMsgProposeOutput(
	proposer string,
	bridgeId uint64,
	outputIndex uint64,
	l2BlockNumber uint64,
	outputRoot []byte,
) *MsgProposeOutput {
	return &MsgProposeOutput{
		Proposer:      proposer,
		BridgeId:      bridgeId,
		OutputIndex:   outputIndex,
		L2BlockNumber: l2BlockNumber,
		OutputRoot:    outputRoot,
	}
}

// Validate performs basic MsgProposeOutput message validation.
func (msg MsgProposeOutput) Validate(ac address.Codec) error {
	_, err := ac.StringToBytes(msg.Proposer)
	if err != nil {
		return err
	}

	if msg.BridgeId == 0 {
		return ErrInvalidBridgeId
	}

	if len(msg.OutputRoot) != 32 {
		return ErrInvalidHashLength.Wrap("output_root")
	}

	return nil
}

/* MsgDeleteOutput */

// NewMsgDeleteOutput creates a new MsgDeleteOutput instance.
func NewMsgDeleteOutput(
	challenger string,
	bridgeId uint64,
	outputIndex uint64,
) *MsgDeleteOutput {
	return &MsgDeleteOutput{
		Challenger:  challenger,
		BridgeId:    bridgeId,
		OutputIndex: outputIndex,
	}
}

// Validate performs basic MsgDeleteOutput message validation.
func (msg MsgDeleteOutput) Validate(ac address.Codec) error {
	if _, err := ac.StringToBytes(msg.Challenger); err != nil {
		return err
	}

	if msg.BridgeId == 0 {
		return ErrInvalidBridgeId
	}

	if msg.OutputIndex == 0 {
		return ErrInvalidOutputIndex
	}

	return nil
}

/* MsgInitiateTokenDeposit */

// NewMsgInitiateTokenDeposit creates a new MsgInitiateTokenDeposit instance.
func NewMsgInitiateTokenDeposit(
	sender string,
	bridgeId uint64,
	to string,
	amount sdk.Coin,
	data []byte,
) *MsgInitiateTokenDeposit {
	return &MsgInitiateTokenDeposit{
		Sender:   sender,
		To:       to,
		Amount:   amount,
		BridgeId: bridgeId,
		Data:     data,
	}
}

// Validate performs basic MsgInitiateTokenDeposit message validation.
func (msg MsgInitiateTokenDeposit) Validate(ac address.Codec) error {
	if _, err := ac.StringToBytes(msg.Sender); err != nil {
		return err
	}

	// cannot validate to address as it can be any format of address based on the chain.
	if len(msg.To) == 0 {
		return sdkerrors.ErrInvalidAddress.Wrap("to address cannot be empty")
	}

	// allow zero amount for creating account
	if !msg.Amount.IsValid() {
		return ErrInvalidAmount
	}

	if msg.BridgeId == 0 {
		return ErrInvalidBridgeId
	}

	return nil
}

/* MsgFinalizeTokenWithdrawal */

// NewMsgFinalizeTokenWithdrawal creates a new MsgFinalizeTokenWithdrawal
func NewMsgFinalizeTokenWithdrawal(
	sender string,
	bridgeId uint64,
	outputIndex uint64,
	sequence uint64,
	withdrawalProofs [][]byte,
	from string,
	to string,
	amount sdk.Coin,
	version []byte,
	storageRoot []byte,
	lastBlockHash []byte,
) *MsgFinalizeTokenWithdrawal {
	return &MsgFinalizeTokenWithdrawal{
		Sender:           sender,
		BridgeId:         bridgeId,
		OutputIndex:      outputIndex,
		WithdrawalProofs: withdrawalProofs,
		From:             from,
		To:               to,
		Sequence:         sequence,
		Amount:           amount,
		Version:          version,
		StorageRoot:      storageRoot,
		LastBlockHash:    lastBlockHash,
	}
}

// Validate performs basic MsgFinalizeTokenWithdrawal message validation.
func (msg MsgFinalizeTokenWithdrawal) Validate(ac address.Codec) error {
	if _, err := ac.StringToBytes(msg.Sender); err != nil {
		return err
	}

	// cannot validate from address as it can be any format of address based on the chain.
	if len(msg.From) == 0 {
		return sdkerrors.ErrInvalidAddress.Wrap("from address cannot be empty")
	}

	if _, err := ac.StringToBytes(msg.To); err != nil {
		return err
	}

	if !msg.Amount.IsValid() || msg.Amount.IsZero() {
		return ErrInvalidAmount
	}

	if msg.Sequence == 0 {
		return ErrInvalidSequence
	}

	if msg.BridgeId == 0 {
		return ErrInvalidBridgeId
	}

	if msg.OutputIndex == 0 {
		return ErrInvalidOutputIndex
	}

	for _, proof := range msg.WithdrawalProofs {
		if len(proof) != 32 {
			return ErrInvalidHashLength.Wrap("withdrawal_proofs")
		}
	}

	if len(msg.Version) != 1 {
		return ErrInvalidHashLength.Wrap("version")
	}

	if len(msg.StorageRoot) != 32 {
		return ErrInvalidHashLength.Wrap("storage_root")
	}

	if len(msg.LastBlockHash) != 32 {
		return ErrInvalidHashLength.Wrap("last_block_hash")
	}

	return nil
}

/* MsgUpdateProposer */

// NewMsgUpdateProposer creates a new MsgUpdateProposer instance.
func NewMsgUpdateProposer(
	authority string,
	bridgeId uint64,
	newProposer string,
) *MsgUpdateProposer {
	return &MsgUpdateProposer{
		Authority:   authority,
		BridgeId:    bridgeId,
		NewProposer: newProposer,
	}
}

// Validate performs basic MsgUpdateProposer message validation.
func (msg MsgUpdateProposer) Validate(ac address.Codec) error {
	if _, err := ac.StringToBytes(msg.Authority); err != nil {
		return err
	}

	if msg.BridgeId == 0 {
		return ErrInvalidBridgeId
	}

	if _, err := ac.StringToBytes(msg.NewProposer); err != nil {
		return err
	}

	return nil
}

/* MsgUpdateChallenger */

// NewMsgUpdateChallenger creates a new MsgUpdateChallenger instance.
func NewMsgUpdateChallenger(
	authority string,
	bridgeId uint64,
	challenger string,
) *MsgUpdateChallenger {
	return &MsgUpdateChallenger{
		Authority:  authority,
		BridgeId:   bridgeId,
		Challenger: challenger,
	}
}

// Validate performs basic MsgUpdateChallenger message validation.
func (msg MsgUpdateChallenger) Validate(ac address.Codec) error {
	if _, err := ac.StringToBytes(msg.Authority); err != nil {
		return err
	}

	if msg.BridgeId == 0 {
		return ErrInvalidBridgeId
	}

	_, err := ac.StringToBytes(msg.Challenger)
	if err != nil {
		return err
	}

	return nil
}

/* MsgUpdateBatchInfo */

// NewMsgUpdateBatchInfo creates a new MsgUpdateBatchInfo instance.
func NewMsgUpdateBatchInfo(
	authority string,
	bridgeId uint64,
	newBatchInfo BatchInfo,
) *MsgUpdateBatchInfo {
	return &MsgUpdateBatchInfo{
		Authority:    authority,
		BridgeId:     bridgeId,
		NewBatchInfo: newBatchInfo,
	}
}

// Validate performs basic MsgUpdateChallenger message validation.
func (msg MsgUpdateBatchInfo) Validate(ac address.Codec) error {
	if _, err := ac.StringToBytes(msg.Authority); err != nil {
		return err
	}

	if msg.BridgeId == 0 {
		return ErrInvalidBridgeId
	}

	if msg.NewBatchInfo.ChainType == BatchInfo_CHAIN_TYPE_UNSPECIFIED || msg.NewBatchInfo.Submitter == "" {
		return ErrEmptyBatchInfo
	}

	return nil
}

/* MsgUpdateOracleConfig */

// NewMsgUpdateOracleConfig creates a new MsgUpdateOracleConfig instance.
func NewMsgUpdateOracleConfig(
	authority string,
	bridgeId uint64,
	oracleEnabled bool,
) *MsgUpdateOracleConfig {
	return &MsgUpdateOracleConfig{
		Authority:     authority,
		BridgeId:      bridgeId,
		OracleEnabled: oracleEnabled,
	}
}

// Validate performs basic MsgUpdateOracleConfig message validation.
func (msg MsgUpdateOracleConfig) Validate(accAddressCodec address.Codec) error {
	if _, err := accAddressCodec.StringToBytes(msg.Authority); err != nil {
		return err
	}

	if msg.BridgeId == 0 {
		return ErrInvalidBridgeId
	}
	return nil
}

/* MsgUpdateMetadata */

// NewMsgUpdateMetadata creates a new MsgUpdateMetadata instance.
func NewMsgUpdateMetadata(
	authority string,
	bridgeId uint64,
	metadata []byte,
) *MsgUpdateMetadata {
	return &MsgUpdateMetadata{
		Authority: authority,
		BridgeId:  bridgeId,
		Metadata:  metadata,
	}
}

// Validate performs basic MsgUpdateMetadata message validation.
func (msg MsgUpdateMetadata) Validate(ac address.Codec) error {
	if _, err := ac.StringToBytes(msg.Authority); err != nil {
		return err
	}

	if msg.BridgeId == 0 {
		return ErrInvalidBridgeId
	}

	if len(msg.Metadata) > MaxMetadataLength {
		return ErrInvalidBridgeMetadata.Wrapf("metadata length exceeds %d", MaxMetadataLength)
	}

	return nil
}

/* MsgUpdateParams */

// NewMsgUpdateParams returns a new MsgUpdateParams instance
func NewMsgUpdateParams(authority string, params *Params) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority,
		Params:    params,
	}
}

// Validate performs basic MsgUpdateParams message validation.
func (msg MsgUpdateParams) Validate(ac address.Codec) error {
	if _, err := ac.StringToBytes(msg.Authority); err != nil {
		return err
	}

	if err := msg.Params.Validate(); err != nil {
		return err
	}

	return nil
}
