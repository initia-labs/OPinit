package types

import (
	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	_ sdk.Msg = &MsgUpdateParams{}
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
func (msg MsgRecordBatch) Validate(accAddressCodec address.Codec) error {
	if _, err := accAddressCodec.StringToBytes(msg.Submitter); err != nil {
		return err
	}

	if msg.BridgeId == 0 {
		return ErrInvalidBridgeId
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

	return nil
}

/* MsgProposeOutput */

// NewMsgProposeOutput creates a new MsgProposeOutput instance.
// Delegator address and validator address are the same.
func NewMsgProposeOutput(
	proposer string,
	bridgeId uint64,
	l2BlockNumber uint64,
	outputRoot []byte,
) *MsgProposeOutput {
	return &MsgProposeOutput{
		Proposer:      proposer,
		BridgeId:      bridgeId,
		L2BlockNumber: l2BlockNumber,
		OutputRoot:    outputRoot,
	}
}

// Validate performs basic MsgProposeOutput message validation.
func (msg MsgProposeOutput) Validate(accAddressCodec address.Codec) error {
	_, err := accAddressCodec.StringToBytes(msg.Proposer)
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
func (msg MsgDeleteOutput) Validate(accAddressCodec address.Codec) error {
	if _, err := accAddressCodec.StringToBytes(msg.Challenger); err != nil {
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
func (msg MsgInitiateTokenDeposit) Validate(accAddressCodec address.Codec) error {
	if _, err := accAddressCodec.StringToBytes(msg.Sender); err != nil {
		return err
	}

	if _, err := accAddressCodec.StringToBytes(msg.To); err != nil {
		return err
	}

	if !msg.Amount.IsValid() || msg.Amount.IsZero() {
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
	bridgeId uint64,
	outputIndex uint64,
	sequence uint64,
	withdrawalProofs [][]byte,
	sender string,
	receiver string,
	amount sdk.Coin,
	version []byte,
	stateRoot []byte,
	storageRoot []byte,
	latestBlockHash []byte,
) *MsgFinalizeTokenWithdrawal {
	return &MsgFinalizeTokenWithdrawal{
		BridgeId:         bridgeId,
		OutputIndex:      outputIndex,
		WithdrawalProofs: withdrawalProofs,
		Sender:           sender,
		Receiver:         receiver,
		Sequence:         sequence,
		Amount:           amount,
		Version:          version,
		StateRoot:        stateRoot,
		StorageRoot:      storageRoot,
		LatestBlockHash:  latestBlockHash,
	}
}

// Validate performs basic MsgFinalizeTokenWithdrawal message validation.
func (msg MsgFinalizeTokenWithdrawal) Validate(accAddressCodec address.Codec) error {
	if _, err := accAddressCodec.StringToBytes(msg.Sender); err != nil {
		return err
	}

	if _, err := accAddressCodec.StringToBytes(msg.Receiver); err != nil {
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

	if len(msg.Version) != 32 {
		return ErrInvalidHashLength.Wrap("version")
	}

	if len(msg.StateRoot) != 32 {
		return ErrInvalidHashLength.Wrap("state_root")
	}

	if len(msg.StorageRoot) != 32 {
		return ErrInvalidHashLength.Wrap("storage_root")
	}

	if len(msg.LatestBlockHash) != 32 {
		return ErrInvalidHashLength.Wrap("latest_block_hash")
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
func (msg MsgUpdateProposer) Validate(accAddressCodec address.Codec) error {
	if _, err := accAddressCodec.StringToBytes(msg.Authority); err != nil {
		return err
	}

	if msg.BridgeId == 0 {
		return ErrInvalidBridgeId
	}

	if _, err := accAddressCodec.StringToBytes(msg.NewProposer); err != nil {
		return err
	}

	return nil
}

/* MsgUpdateChallenger */

// NewMsgUpdateChallenger creates a new MsgUpdateChallenger instance.
func NewMsgUpdateChallenger(
	authority string,
	bridgeId uint64,
	newChallenger string,
) *MsgUpdateChallenger {
	return &MsgUpdateChallenger{
		Authority:     authority,
		BridgeId:      bridgeId,
		NewChallenger: newChallenger,
	}
}

// Validate performs basic MsgUpdateChallenger message validation.
func (msg MsgUpdateChallenger) Validate(accAddressCodec address.Codec) error {
	if _, err := accAddressCodec.StringToBytes(msg.Authority); err != nil {
		return err
	}

	if msg.BridgeId == 0 {
		return ErrInvalidBridgeId
	}

	if _, err := accAddressCodec.StringToBytes(msg.NewChallenger); err != nil {
		return err
	}

	return nil
}

/* MsgUpdateBatchInfo */

// NewMsgUpdateBatchInfo creates a new MsgUpdateBatchInfo instance.
func NewMsgUpdateBatchInfo(
	authority string,
	bridgeId uint64,
	newBatchInfo *BatchInfo,
) *MsgUpdateBatchInfo {
	return &MsgUpdateBatchInfo{
		Authority:    authority,
		BridgeId:     bridgeId,
		NewBatchInfo: newBatchInfo,
	}
}

// Validate performs basic MsgUpdateChallenger message validation.
func (msg MsgUpdateBatchInfo) Validate(accAddressCodec address.Codec) error {
	if _, err := accAddressCodec.StringToBytes(msg.Authority); err != nil {
		return err
	}

	if msg.BridgeId == 0 {
		return ErrInvalidBridgeId
	}

	if msg.NewBatchInfo != nil && (msg.NewBatchInfo.Chain == "" || msg.NewBatchInfo.Submitter == "") {
		return ErrEmptyBatchInfo
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
func (msg MsgUpdateParams) Validate(accAddressCodec address.Codec) error {
	if _, err := accAddressCodec.StringToBytes(msg.Authority); err != nil {
		return err
	}

	if err := msg.Params.Validate(); err != nil {
		return err
	}

	return nil
}
