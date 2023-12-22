package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// RegisterLegacyAminoCodec registers the move types and interface
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {

	legacy.RegisterAminoMsg(cdc, &MsgExecuteMessages{}, "opchild/MsgExecuteMessages")
	legacy.RegisterAminoMsg(cdc, &MsgExecuteLegacyContents{}, "opchild/MsgExecuteLegacyContents")
	legacy.RegisterAminoMsg(cdc, &MsgAddValidator{}, "opchild/MsgAddValidator")
	legacy.RegisterAminoMsg(cdc, &MsgRemoveValidator{}, "opchild/MsgRemoveAddValidator")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "opchild/MsgUpdateParams")
	legacy.RegisterAminoMsg(cdc, &MsgFinalizeTokenDeposit{}, "opchild/MsgFinalizeTokenDeposit")
	legacy.RegisterAminoMsg(cdc, &MsgInitiateTokenWithdrawal{}, "opchild/MsgInitiateTokenWithdrawal")

	cdc.RegisterConcrete(Params{}, "opchild/Params", nil)
}

// RegisterInterfaces registers the x/market interfaces types with the interface registry
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgExecuteMessages{},
		&MsgAddValidator{},
		&MsgRemoveValidator{},
		&MsgUpdateParams{},
		&MsgFinalizeTokenDeposit{},
		&MsgInitiateTokenWithdrawal{},
	)
	registry.RegisterInterface(
		"cosmos.gov.v1beta1.Content",
		(*govv1beta1.Content)(nil),
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
