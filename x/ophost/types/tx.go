package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

// rollup message types
const (
	TypeMsgRecordBatch             = "record_batch"
	TypeMsgCreateBridge            = "create_bridge"
	TypeMsgProposeOutput           = "propose_output"
	TypeMsgDeleteOutput            = "delete_output"
	TypeMsgInitiateTokenDeposit    = "deposit"
	TypeMsgFinalizeTokenWithdrawal = "claim"
	TypeMsgUpdateProposer          = "update_proposer"
	TypeMsgUpdateChallenger        = "update_challenger"
	TypeMsgUpdateParams            = "update_params"
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
	_ sdk.Msg = &MsgUpdateParams{}

	_ legacytx.LegacyMsg = &MsgRecordBatch{}
	_ legacytx.LegacyMsg = &MsgCreateBridge{}
	_ legacytx.LegacyMsg = &MsgProposeOutput{}
	_ legacytx.LegacyMsg = &MsgDeleteOutput{}
	_ legacytx.LegacyMsg = &MsgFinalizeTokenWithdrawal{}
	_ legacytx.LegacyMsg = &MsgInitiateTokenDeposit{}
	_ legacytx.LegacyMsg = &MsgUpdateProposer{}
	_ legacytx.LegacyMsg = &MsgUpdateChallenger{}
	_ legacytx.LegacyMsg = &MsgUpdateParams{}
)

/* MsgRecordBatch */

// NewMsgRecordBatch creates a new MsgRecordBatch instance.
func NewMsgRecordBatch(
	submitter sdk.AccAddress,
	bridgeId uint64,
	batchBytes []byte,
) *MsgRecordBatch {
	return &MsgRecordBatch{
		Submitter:  submitter.String(),
		BridgeId:   bridgeId,
		BatchBytes: batchBytes,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgRecordBatch) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg MsgRecordBatch) Type() string {
	return TypeMsgRecordBatch
}

// ValidateBasic performs basic MsgRecordBatch message validation.
func (msg MsgRecordBatch) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Submitter); err != nil {
		return err
	}

	if msg.BridgeId == 0 {
		return ErrInvalidBridgeId
	}

	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgRecordBatch) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the signer addresses that are expected to sign the result
// of GetSignBytes.
func (msg MsgRecordBatch) GetSigners() []sdk.AccAddress {
	submitterAddr, err := sdk.AccAddressFromBech32(msg.Submitter)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{submitterAddr}
}

/* MsgCreateBridge */

// NewMsgCreateBridge creates a new MsgCreateBridge instance.
func NewMsgCreateBridge(
	creator sdk.AccAddress,
	config BridgeConfig,
) *MsgCreateBridge {
	return &MsgCreateBridge{
		Creator: creator.String(),
		Config:  config,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgCreateBridge) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg MsgCreateBridge) Type() string {
	return TypeMsgCreateBridge
}

// ValidateBasic performs basic MsgCreateBridge message validation.
func (msg MsgCreateBridge) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return err
	}

	if err := msg.Config.Validate(); err != nil {
		return err
	}

	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgCreateBridge) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the signer addresses that are expected to sign the result
// of GetSignBytes.
func (msg MsgCreateBridge) GetSigners() []sdk.AccAddress {
	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{creatorAddr}
}

/* MsgProposeOutput */

// NewMsgProposeOutput creates a new MsgProposeOutput instance.
// Delegator address and validator address are the same.
func NewMsgProposeOutput(
	proposer sdk.AccAddress,
	bridgeId uint64,
	l2BlockNumber uint64,
	outputRoot []byte,
) *MsgProposeOutput {
	return &MsgProposeOutput{
		Proposer:      proposer.String(),
		BridgeId:      bridgeId,
		L2BlockNumber: l2BlockNumber,
		OutputRoot:    outputRoot,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgProposeOutput) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg MsgProposeOutput) Type() string {
	return TypeMsgProposeOutput
}

// ValidateBasic performs basic MsgProposeOutput message validation.
func (msg MsgProposeOutput) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Proposer)
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

// GetSignBytes returns the message bytes to sign over.
func (msg MsgProposeOutput) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the signer addresses that are expected to sign the result
// of GetSignBytes.
func (msg MsgProposeOutput) GetSigners() []sdk.AccAddress {
	proposer, _ := sdk.AccAddressFromBech32(msg.Proposer)

	return []sdk.AccAddress{proposer}
}

/* MsgDeleteOutput */

// NewMsgDeleteOutput creates a new MsgDeleteOutput instance.
func NewMsgDeleteOutput(
	challenger sdk.AccAddress,
	bridgeId uint64,
	outputIndex uint64,
) *MsgDeleteOutput {
	return &MsgDeleteOutput{
		Challenger:  challenger.String(),
		BridgeId:    bridgeId,
		OutputIndex: outputIndex,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgDeleteOutput) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg MsgDeleteOutput) Type() string {
	return TypeMsgDeleteOutput
}

// ValidateBasic performs basic MsgDeleteOutput message validation.
func (msg MsgDeleteOutput) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Challenger); err != nil {
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

// GetSignBytes returns the message bytes to sign over.
func (msg MsgDeleteOutput) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the signer addresses that are expected to sign the result
// of GetSignBytes.
func (msg MsgDeleteOutput) GetSigners() []sdk.AccAddress {
	challengerAddr, err := sdk.AccAddressFromBech32(msg.Challenger)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{challengerAddr}
}

/* MsgInitiateTokenDeposit */

// NewMsgInitiateTokenDeposit creates a new MsgInitiateTokenDeposit instance.
func NewMsgInitiateTokenDeposit(
	sender sdk.AccAddress,
	bridgeId uint64,
	to sdk.AccAddress,
	amount sdk.Coin,
	data []byte,
) *MsgInitiateTokenDeposit {
	return &MsgInitiateTokenDeposit{
		Sender:   sender.String(),
		To:       to.String(),
		Amount:   amount,
		BridgeId: bridgeId,
		Data:     data,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgInitiateTokenDeposit) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg MsgInitiateTokenDeposit) Type() string {
	return TypeMsgInitiateTokenDeposit
}

// ValidateBasic performs basic MsgInitiateTokenDeposit message validation.
func (msg MsgInitiateTokenDeposit) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(msg.To); err != nil {
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

// GetSignBytes returns the message bytes to sign over.
func (msg MsgInitiateTokenDeposit) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the signer addresses that are expected to sign the result
// of GetSignBytes.
func (msg MsgInitiateTokenDeposit) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{senderAddr}
}

/* MsgFinalizeTokenWithdrawal */

// NewMsgFinalizeTokenWithdrawal creates a new MsgFinalizeTokenWithdrawal
func NewMsgFinalizeTokenWithdrawal(
	bridgeId uint64,
	outputIndex uint64,
	sequence uint64,
	withdrawalProofs [][]byte,
	sender sdk.AccAddress,
	receiver sdk.AccAddress,
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
		Sender:           sender.String(),
		Receiver:         receiver.String(),
		Sequence:         sequence,
		Amount:           amount,
		Version:          version,
		StateRoot:        stateRoot,
		StorageRoot:      storageRoot,
		LatestBlockHash:  latestBlockHash,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgFinalizeTokenWithdrawal) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg MsgFinalizeTokenWithdrawal) Type() string {
	return TypeMsgFinalizeTokenWithdrawal
}

// ValidateBasic performs basic MsgFinalizeTokenWithdrawal message validation.
func (msg MsgFinalizeTokenWithdrawal) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(msg.Receiver); err != nil {
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

// GetSignBytes returns the message bytes to sign over.
func (msg MsgFinalizeTokenWithdrawal) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the signer addresses that are expected to sign the result
// of GetSignBytes.
func (msg MsgFinalizeTokenWithdrawal) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{senderAddr}
}

/* MsgUpdateProposer */

// NewMsgUpdateProposer creates a new MsgUpdateProposer instance.
func NewMsgUpdateProposer(
	authority sdk.AccAddress,
	bridgeId uint64,
	newProposer sdk.AccAddress,
) *MsgUpdateProposer {
	return &MsgUpdateProposer{
		Authority:   authority.String(),
		BridgeId:    bridgeId,
		NewProposer: newProposer.String(),
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgUpdateProposer) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg MsgUpdateProposer) Type() string {
	return TypeMsgUpdateProposer
}

// ValidateBasic performs basic MsgUpdateProposer message validation.
func (msg MsgUpdateProposer) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return err
	}

	if msg.BridgeId == 0 {
		return ErrInvalidBridgeId
	}

	if _, err := sdk.AccAddressFromBech32(msg.NewProposer); err != nil {
		return err
	}

	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgUpdateProposer) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the signer addresses that are expected to sign the result
// of GetSignBytes.
func (msg MsgUpdateProposer) GetSigners() []sdk.AccAddress {
	authorityAddr, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{authorityAddr}
}

/* MsgUpdateChallenger */

// NewMsgUpdateChallenger creates a new MsgUpdateChallenger instance.
func NewMsgUpdateChallenger(
	authority sdk.AccAddress,
	bridgeId uint64,
	newChallenger sdk.AccAddress,
) *MsgUpdateChallenger {
	return &MsgUpdateChallenger{
		Authority:     authority.String(),
		BridgeId:      bridgeId,
		NewChallenger: newChallenger.String(),
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgUpdateChallenger) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg MsgUpdateChallenger) Type() string {
	return TypeMsgUpdateChallenger
}

// ValidateBasic performs basic MsgUpdateChallenger message validation.
func (msg MsgUpdateChallenger) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return err
	}

	if msg.BridgeId == 0 {
		return ErrInvalidBridgeId
	}

	if _, err := sdk.AccAddressFromBech32(msg.NewChallenger); err != nil {
		return err
	}

	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgUpdateChallenger) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the signer addresses that are expected to sign the result
// of GetSignBytes.
func (msg MsgUpdateChallenger) GetSigners() []sdk.AccAddress {
	authorityAddr, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{authorityAddr}
}

/* MsgUpdateParams */

// NewMsgUpdateParams returns a new MsgUpdateParams instance
func NewMsgUpdateParams(authority sdk.AccAddress, params *Params) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority.String(),
		Params:    params,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgUpdateParams) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg MsgUpdateParams) Type() string {
	return TypeMsgUpdateParams
}

// ValidateBasic performs basic MsgUpdateParams message validation.
func (msg MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return err
	}

	if err := msg.Params.Validate(); err != nil {
		return err
	}

	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgUpdateParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the signer addresses that are expected to sign the result
// of GetSignBytes.
func (msg MsgUpdateParams) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{senderAddr}
}
