package types

import (
	"fmt"

	errors "cosmossdk.io/errors"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/gogoproto/proto"
)

// rollup message types
const (
	TypeMsgExecuteMessages       = "execute_messages"
	TypeMsgExecuteLegacyContents = "execute_legacy_contents"

	TypeMsgAddValidator    = "add_validator"
	TypeMsgRemoveValidator = "remove_validator"
	TypeMsgUpdateParams    = "update_params"
	TypeMsgWhitelist       = "whitelist"
	TypeMsgSpendFeePool    = "spend_fee_pool"

	TypeMsgInitiateTokenWithdrawal = "withdraw‰"
	TypeMsgFinalizeTokenDeposit    = "deposit"
)

var (
	_ sdk.Msg = &MsgExecuteMessages{}
	_ sdk.Msg = &MsgExecuteLegacyContents{}
	_ sdk.Msg = &MsgAddValidator{}
	_ sdk.Msg = &MsgRemoveValidator{}
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgSpendFeePool{}
	_ sdk.Msg = &MsgFinalizeTokenDeposit{}
	_ sdk.Msg = &MsgInitiateTokenWithdrawal{}

	_ legacytx.LegacyMsg = &MsgExecuteMessages{}
	_ legacytx.LegacyMsg = &MsgExecuteLegacyContents{}
	_ legacytx.LegacyMsg = &MsgAddValidator{}
	_ legacytx.LegacyMsg = &MsgRemoveValidator{}
	_ legacytx.LegacyMsg = &MsgUpdateParams{}
	_ legacytx.LegacyMsg = &MsgSpendFeePool{}
	_ legacytx.LegacyMsg = &MsgFinalizeTokenDeposit{}
	_ legacytx.LegacyMsg = &MsgInitiateTokenWithdrawal{}
)

// should refer initiavm/precompile/modules/minlib/sources/coin.move
const MAX_TOKEN_NAME_LENGTH = 128
const MAX_TOKEN_SYMBOL_LENGTH = 128

/* MsgExecuteMessages */

// NewMsgExecuteMessages creates a new MsgExecuteMessages instance.
func NewMsgExecuteMessages(
	sender sdk.AccAddress, //nolint:interfacer
	messages []sdk.Msg,
) (*MsgExecuteMessages, error) {
	msg := &MsgExecuteMessages{
		Sender: sender.String(),
	}

	anys, err := sdktx.SetMsgs(messages)
	if err != nil {
		return nil, err
	}
	msg.Messages = anys

	return msg, nil
}

// GetMsgs unpacks m.Messages Any's into sdk.Msg's
func (msg *MsgExecuteMessages) GetMsgs() ([]sdk.Msg, error) {
	return sdktx.GetMsgs(msg.Messages, "sdk.MsgProposal")
}

// Route implements the sdk.Msg interface.
func (msg MsgExecuteMessages) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg MsgExecuteMessages) Type() string {
	return TypeMsgExecuteMessages
}

// ValidateBasic performs basic MsgExecuteMessages message validation.
func (msg MsgExecuteMessages) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return err
	}

	msgs, err := msg.GetMsgs()
	if err != nil {
		return err
	}
	// Check Msgs length is non nil.
	if len(msg.Messages) == 0 {
		return errors.Wrap(govtypes.ErrNoProposalMsgs, "Msgs length must be non-zero")
	}

	for idx, msg := range msgs {
		if err := msg.ValidateBasic(); err != nil {
			return errors.Wrapf(govtypes.ErrInvalidProposalMsg, "msg: %d, err: %s", idx, err.Error())
		}

		signers := msg.GetSigners()
		if len(signers) != 1 {
			return govtypes.ErrInvalidSigner
		}
	}

	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgExecuteMessages) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the signer addresses that are expected to sign the result
// of GetSignBytes.
func (msg MsgExecuteMessages) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{senderAddr}
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgExecuteMessages) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return sdktx.UnpackInterfaces(unpacker, m.Messages)
}

/* MsgExecuteLegacyContents */

// NewMsgExecuteLegacyContents creates a new MsgExecuteLegacyContents instance.
func NewMsgExecuteLegacyContents(
	sender sdk.AccAddress, //nolint:interfacer
	contents []govv1beta1.Content,
) (*MsgExecuteLegacyContents, error) {
	msg := &MsgExecuteLegacyContents{
		Sender: sender.String(),
	}

	if err := msg.SetContents(contents); err != nil {
		return nil, err
	}

	return msg, nil
}

// GetContents returns the contents of MsgExecuteLegacyContents.
func (m *MsgExecuteLegacyContents) GetContents() []govv1beta1.Content {
	contents := make([]govv1beta1.Content, len(m.Contents))
	for i, content := range m.Contents {
		content, ok := content.GetCachedValue().(govv1beta1.Content)
		if !ok {
			return nil
		}

		contents[i] = content
	}

	return contents
}

// SetContents sets the contents for MsgExecuteLegacyContents.
func (m *MsgExecuteLegacyContents) SetContents(contents []govv1beta1.Content) error {
	anys := make([]*codectypes.Any, len(contents))
	for i, content := range contents {
		msg, ok := content.(proto.Message)
		if !ok {
			return fmt.Errorf("can't proto marshal %T", msg)
		}
		any, err := codectypes.NewAnyWithValue(msg)
		if err != nil {
			return err
		}

		anys[i] = any
	}

	m.Contents = anys
	return nil
}

// Route implements the sdk.Msg interface.
func (msg MsgExecuteLegacyContents) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg MsgExecuteLegacyContents) Type() string {
	return TypeMsgExecuteLegacyContents
}

// ValidateBasic performs basic MsgExecuteLegacyContents message validation.
func (msg MsgExecuteLegacyContents) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return err
	}

	// Check Msgs length is non nil.
	if len(msg.Contents) == 0 {
		return errors.Wrap(govtypes.ErrNoProposalMsgs, "Contents length must be non-zero")
	}

	contents := msg.GetContents()
	for idx, content := range contents {
		if err := content.ValidateBasic(); err != nil {
			return errors.Wrapf(govtypes.ErrInvalidProposalContent, "content: %d, err: %s", idx, err.Error())
		}
	}

	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgExecuteLegacyContents) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the signer addresses that are expected to sign the result
// of GetSignBytes.
func (msg MsgExecuteLegacyContents) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{senderAddr}
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (c MsgExecuteLegacyContents) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, anyContent := range c.Contents {
		var content govv1beta1.Content
		if err := unpacker.UnpackAny(anyContent, &content); err != nil {
			return err
		}
	}

	return nil
}

/* MsgAddValidator */

// NewMsgAddValidator creates a new MsgAddValidator instance.
// Delegator address and validator address are the same.
func NewMsgAddValidator(
	moniker string, authority sdk.AccAddress,
	valAddr sdk.ValAddress, pubKey cryptotypes.PubKey, //nolint:interfacer
) (*MsgAddValidator, error) {
	var pkAny *codectypes.Any
	if pubKey != nil {
		var err error
		if pkAny, err = codectypes.NewAnyWithValue(pubKey); err != nil {
			return nil, err
		}
	}
	return &MsgAddValidator{
		Moniker:          moniker,
		Authority:        authority.String(),
		ValidatorAddress: valAddr.String(),
		Pubkey:           pkAny,
	}, nil
}

// Route implements the sdk.Msg interface.
func (msg MsgAddValidator) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg MsgAddValidator) Type() string {
	return TypeMsgAddValidator
}

// ValidateBasic performs basic MsgAddValidator message validation.
func (msg MsgAddValidator) ValidateBasic() error {
	// note that unmarshaling from bech32 ensures both non-empty and valid
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid delegator address: %s", err)
	}
	_, err = sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	if msg.Pubkey == nil {
		return ErrEmptyValidatorPubKey
	}

	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgAddValidator) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the signer addresses that are expected to sign the result
// of GetSignBytes.
func (msg MsgAddValidator) GetSigners() []sdk.AccAddress {
	// delegator is first signer so delegator pays fees
	delegator, _ := sdk.AccAddressFromBech32(msg.Authority)
	addrs := []sdk.AccAddress{delegator}
	valAddr, _ := sdk.ValAddressFromBech32(msg.ValidatorAddress)

	valAccAddr := sdk.AccAddress(valAddr)
	if !delegator.Equals(valAccAddr) {
		addrs = append(addrs, valAccAddr)
	}

	return addrs
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgAddValidator) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pubKey cryptotypes.PubKey
	return unpacker.UnpackAny(msg.Pubkey, &pubKey)
}

/* MsgRemoveValidator */

// NewMsgRemoveValidator creates a new MsgRemoveValidator instance.
func NewMsgRemoveValidator(
	authority sdk.AccAddress,
	valAddr sdk.ValAddress, //nolint:interfacer
) (*MsgRemoveValidator, error) {
	return &MsgRemoveValidator{
		Authority:        authority.String(),
		ValidatorAddress: valAddr.String(),
	}, nil
}

// Route implements the sdk.Msg interface.
func (msg MsgRemoveValidator) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg MsgRemoveValidator) Type() string {
	return TypeMsgRemoveValidator
}

// ValidateBasic performs basic MsgRemoveValidator message validation.
func (msg MsgRemoveValidator) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return err
	}

	if _, err := sdk.ValAddressFromBech32(msg.ValidatorAddress); err != nil {
		return err
	}

	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgRemoveValidator) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the signer addresses that are expected to sign the result
// of GetSignBytes.
func (msg MsgRemoveValidator) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{senderAddr}
}

/* MsgInitiateTokenWithdrawal */

// NewMsgInitiateTokenWithdrawal creates a new MsgInitiateTokenWithdrawal instance.
func NewMsgInitiateTokenWithdrawal(
	sender sdk.AccAddress,
	to sdk.AccAddress,
	amount sdk.Coin,
) *MsgInitiateTokenWithdrawal {
	return &MsgInitiateTokenWithdrawal{
		Sender: sender.String(),
		To:     to.String(),
		Amount: amount,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgInitiateTokenWithdrawal) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg MsgInitiateTokenWithdrawal) Type() string {
	return TypeMsgInitiateTokenWithdrawal
}

// ValidateBasic performs basic MsgInitiateTokenWithdrawal message validation.
func (msg MsgInitiateTokenWithdrawal) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(msg.To); err != nil {
		return err
	}

	if !msg.Amount.IsValid() || msg.Amount.IsZero() {
		return ErrInvalidAmount
	}

	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgInitiateTokenWithdrawal) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the signer addresses that are expected to sign the result
// of GetSignBytes.
func (msg MsgInitiateTokenWithdrawal) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{senderAddr}
}

/* MsgFinalizeTokenDeposit */

// NewMsgFinalizeTokenDeposit creates a new MsgFinalizeTokenDeposit instance.
func NewMsgFinalizeTokenDeposit(
	sender, from, to sdk.AccAddress,
	amount sdk.Coin,
	sequence uint64,
	data []byte,
) *MsgFinalizeTokenDeposit {
	return &MsgFinalizeTokenDeposit{
		Sender:   sender.String(),
		From:     from.String(),
		To:       to.String(),
		Amount:   amount,
		Sequence: sequence,
		Data:     data,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgFinalizeTokenDeposit) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg MsgFinalizeTokenDeposit) Type() string {
	return TypeMsgFinalizeTokenDeposit
}

// ValidateBasic performs basic MsgFinalizeTokenDeposit message validation.
func (msg MsgFinalizeTokenDeposit) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(msg.From); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(msg.To); err != nil {
		return err
	}

	if !msg.Amount.IsValid() || msg.Amount.IsZero() {
		return ErrInvalidAmount
	}

	if msg.Sequence == 0 {
		return ErrInvalidSequence
	}

	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgFinalizeTokenDeposit) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the signer addresses that are expected to sign the result
// of GetSignBytes.
func (msg MsgFinalizeTokenDeposit) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{senderAddr}
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

/* MsgSpendFeePool */

// NewMsgSpendFeePool creates a new MsgSpendFeePool
func NewMsgSpendFeePool(authority, recipient sdk.AccAddress, amount sdk.Coins) *MsgSpendFeePool {
	return &MsgSpendFeePool{
		Authority: authority.String(),
		Recipient: recipient.String(),
		Amount:    amount,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgSpendFeePool) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg MsgSpendFeePool) Type() string {
	return TypeMsgSpendFeePool
}

// ValidateBasic performs basic MsgSpendFeePool message validation.
func (msg MsgSpendFeePool) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(msg.Recipient); err != nil {
		return err
	}

	if !msg.Amount.IsValid() {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "invalid spend amount")
	}

	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgSpendFeePool) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the signer addresses that are expected to sign the result
// of GetSignBytes.
func (msg MsgSpendFeePool) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{senderAddr}
}
