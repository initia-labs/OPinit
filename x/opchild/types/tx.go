package types

import (
	"cosmossdk.io/core/address"
	errors "cosmossdk.io/errors"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// rollup message types
const (
	TypeMsgExecuteMessages = "execute_messages"

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
	_ sdk.Msg = &MsgAddValidator{}
	_ sdk.Msg = &MsgRemoveValidator{}
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgSpendFeePool{}
	_ sdk.Msg = &MsgFinalizeTokenDeposit{}
	_ sdk.Msg = &MsgInitiateTokenWithdrawal{}

	_ codectypes.UnpackInterfacesMessage = &MsgExecuteMessages{}
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

// Validate performs basic MsgExecuteMessages message validation.
func (msg MsgExecuteMessages) Validate(ac address.Codec) error {
	if _, err := ac.StringToBytes(msg.Sender); err != nil {
		return err
	}

	// Check Msgs length is non nil.
	if len(msg.Messages) == 0 {
		return errors.Wrap(govtypes.ErrNoProposalMsgs, "Msgs length must be non-zero")
	}

	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgExecuteMessages) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return sdktx.UnpackInterfaces(unpacker, m.Messages)
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

// Validate performs basic MsgAddValidator message validation.
func (msg MsgAddValidator) Validate(ac address.Codec, vc address.Codec) error {
	// note that unmarshaling from bech32 ensures both non-empty and valid
	_, err := ac.StringToBytes(msg.Authority)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid delegator address: %s", err)
	}
	_, err = vc.StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	if msg.Pubkey == nil {
		return ErrEmptyValidatorPubKey
	}

	return nil
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

// Validate performs basic MsgRemoveValidator message validation.
func (msg MsgRemoveValidator) Validate(ac, vc address.Codec) error {
	if _, err := ac.StringToBytes(msg.Authority); err != nil {
		return err
	}

	if _, err := vc.StringToBytes(msg.ValidatorAddress); err != nil {
		return err
	}

	return nil
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

// Validate performs basic MsgInitiateTokenWithdrawal message validation.
func (msg MsgInitiateTokenWithdrawal) Validate(ac address.Codec) error {
	if _, err := ac.StringToBytes(msg.Sender); err != nil {
		return err
	}

	if _, err := ac.StringToBytes(msg.To); err != nil {
		return err
	}

	if !msg.Amount.IsValid() || msg.Amount.IsZero() {
		return ErrInvalidAmount
	}

	return nil
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

// Validate performs basic MsgFinalizeTokenDeposit message validation.
func (msg MsgFinalizeTokenDeposit) Validate(ac address.Codec) error {
	if _, err := ac.StringToBytes(msg.Sender); err != nil {
		return err
	}

	if _, err := ac.StringToBytes(msg.From); err != nil {
		return err
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

	return nil
}

/* MsgUpdateParams */

// NewMsgUpdateParams returns a new MsgUpdateParams instance
func NewMsgUpdateParams(authority sdk.AccAddress, params *Params) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority.String(),
		Params:    params,
	}
}

// Validate performs basic MsgUpdateParams message validation.
func (msg MsgUpdateParams) Validate(ac address.Codec) error {
	if _, err := ac.StringToBytes(msg.Authority); err != nil {
		return err
	}

	if err := msg.Params.Validate(ac); err != nil {
		return err
	}

	return nil
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

// Validate performs basic MsgSpendFeePool message validation.
func (msg MsgSpendFeePool) Validate(ac address.Codec) error {
	if _, err := ac.StringToBytes(msg.Authority); err != nil {
		return err
	}

	if _, err := ac.StringToBytes(msg.Recipient); err != nil {
		return err
	}

	if !msg.Amount.IsValid() {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "invalid spend amount")
	}

	return nil
}
