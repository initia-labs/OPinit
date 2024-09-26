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

var (
	_ sdk.Msg = &MsgExecuteMessages{}
	_ sdk.Msg = &MsgAddValidator{}
	_ sdk.Msg = &MsgRemoveValidator{}
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgSpendFeePool{}
	_ sdk.Msg = &MsgSetBridgeInfo{}
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
	sender string,
	messages []sdk.Msg,
) (*MsgExecuteMessages, error) {
	msg := &MsgExecuteMessages{
		Sender: sender,
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
func (msg MsgExecuteMessages) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return sdktx.UnpackInterfaces(unpacker, msg.Messages)
}

/* MsgAddValidator */

// NewMsgAddValidator creates a new MsgAddValidator instance.
// Delegator address and validator address are the same.
func NewMsgAddValidator(
	moniker string, authority string,
	valAddr string, pubKey cryptotypes.PubKey,
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
		Authority:        authority,
		ValidatorAddress: valAddr,
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
	authority string,
	valAddr string,
) (*MsgRemoveValidator, error) {
	return &MsgRemoveValidator{
		Authority:        authority,
		ValidatorAddress: valAddr,
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
	sender string,
	to string,
	amount sdk.Coin,
) *MsgInitiateTokenWithdrawal {
	return &MsgInitiateTokenWithdrawal{
		Sender: sender,
		To:     to,
		Amount: amount,
	}
}

// Validate performs basic MsgInitiateTokenWithdrawal message validation.
func (msg MsgInitiateTokenWithdrawal) Validate(ac address.Codec) error {
	if _, err := ac.StringToBytes(msg.Sender); err != nil {
		return err
	}

	if len(msg.To) == 0 {
		return sdkerrors.ErrInvalidAddress.Wrap("to address cannot be empty")
	}

	if !msg.Amount.IsValid() || !msg.Amount.IsPositive() {
		return ErrInvalidAmount
	}

	return nil
}

/* MsgSetBridgeInfo */
func NewMsgSetBridgeInfo(
	sender string,
	bridgeInfo BridgeInfo,
) *MsgSetBridgeInfo {
	return &MsgSetBridgeInfo{
		Sender:     sender,
		BridgeInfo: bridgeInfo,
	}
}

func (msg MsgSetBridgeInfo) Validate(ac address.Codec) error {
	if _, err := ac.StringToBytes(msg.Sender); err != nil {
		return err
	}

	if err := msg.BridgeInfo.Validate(ac); err != nil {
		return err
	}

	return nil
}

func (info BridgeInfo) Validate(ac address.Codec) error {
	if info.BridgeId == 0 {
		return ErrInvalidBridgeInfo.Wrap("bridge id cannot be zero")
	}

	if len(info.BridgeAddr) == 0 {
		return ErrInvalidBridgeInfo.Wrap("bridge address cannot be empty")
	}

	if err := info.BridgeConfig.ValidateWithNoAddrValidation(); err != nil {
		return ErrInvalidBridgeInfo.Wrap(err.Error())
	}

	return nil
}

/* MsgFinalizeTokenDeposit */

// NewMsgFinalizeTokenDeposit creates a new MsgFinalizeTokenDeposit instance.
func NewMsgFinalizeTokenDeposit(
	sender, from, to string,
	amount sdk.Coin,
	sequence uint64,
	height uint64,
	baseDenom string,
	data []byte,
) *MsgFinalizeTokenDeposit {
	return &MsgFinalizeTokenDeposit{
		Sender:    sender,
		From:      from,
		To:        to,
		Amount:    amount,
		Sequence:  sequence,
		Height:    height,
		BaseDenom: baseDenom,
		Data:      data,
	}
}

// Validate performs basic MsgFinalizeTokenDeposit message validation.
func (msg MsgFinalizeTokenDeposit) Validate(ac address.Codec) error {
	if _, err := ac.StringToBytes(msg.Sender); err != nil {
		return err
	}

	// from address can be different form of address (e.g. L1 address)
	if len(msg.From) == 0 {
		return sdkerrors.ErrInvalidAddress.Wrap("from address cannot be empty")
	}

	// allow zero amount
	if !msg.Amount.IsValid() {
		return ErrInvalidAmount
	}

	if err := sdk.ValidateDenom(msg.BaseDenom); err != nil {
		return err
	}

	if msg.Sequence == 0 {
		return ErrInvalidSequence
	}

	if msg.Height == 0 {
		return ErrInvalidBlockHeight
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

func NewMsgUpdateOracle(sender string, height uint64, data []byte) *MsgUpdateOracle {
	return &MsgUpdateOracle{
		Sender: sender,
		Height: height,
		Data:   data,
	}
}

func (msg MsgUpdateOracle) Validate(ac address.Codec) error {
	if _, err := ac.StringToBytes(msg.Sender); err != nil {
		return err
	}

	if msg.Height == 0 {
		return ErrInvalidHeight
	}
	return nil
}
