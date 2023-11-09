package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// RegisterLegacyAminoCodec registers the move types and interface
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgRecordBatch{}, "op_host/MsgRecordBatch")
	legacy.RegisterAminoMsg(cdc, &MsgCreateBridge{}, "op_host/MsgCreateBridge")
	legacy.RegisterAminoMsg(cdc, &MsgProposeOutput{}, "op_host/MsgProposeOutput")
	legacy.RegisterAminoMsg(cdc, &MsgDeleteOutput{}, "op_host/MsgDeleteOutput")
	legacy.RegisterAminoMsg(cdc, &MsgInitiateTokenDeposit{}, "op_host/MsgInitiateTokenDeposit")
	legacy.RegisterAminoMsg(cdc, &MsgFinalizeTokenWithdrawal{}, "op_host/MsgFinalizeTokenWithdrawal")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateProposer{}, "op_host/MsgUpdateProposer")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateChallenger{}, "op_host/MsgUpdateChallenger")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "op_host/MsgUpdateParams")

	cdc.RegisterConcrete(Params{}, "op_host/Params", nil)
	cdc.RegisterConcrete(&BridgeAccount{}, "move/BridgeAccount", nil)
}

// RegisterInterfaces registers the x/market interfaces types with the interface registry
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRecordBatch{},
		&MsgCreateBridge{},
		&MsgProposeOutput{},
		&MsgDeleteOutput{},
		&MsgInitiateTokenDeposit{},
		&MsgFinalizeTokenWithdrawal{},
		&MsgUpdateProposer{},
		&MsgUpdateChallenger{},
		&MsgUpdateParams{},
	)

	// auth account registration
	registry.RegisterImplementations(
		(*authtypes.AccountI)(nil),
		&BridgeAccount{},
	)
	registry.RegisterImplementations(
		(*authtypes.GenesisAccount)(nil),
		&BridgeAccount{},
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
}
