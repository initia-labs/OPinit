package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"
	govcodec "github.com/cosmos/cosmos-sdk/x/gov/codec"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	groupcodec "github.com/cosmos/cosmos-sdk/x/group/codec"
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

	//cdc.RegisterConcrete(&PublishAuthorization{}, "move/PublishAuthorization", nil)
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

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	sdk.RegisterLegacyAminoCodec(amino)

	// Register all Amino interfaces and concrete types on the authz  and gov Amino codec so that this can later be
	// used to properly serialize MsgGrant, MsgExec and MsgSubmitProposal instances
	RegisterLegacyAminoCodec(authzcodec.Amino)
	RegisterLegacyAminoCodec(govcodec.Amino)
	RegisterLegacyAminoCodec(groupcodec.Amino)
}
