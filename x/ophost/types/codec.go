package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// RegisterLegacyAminoCodec registers the move types and interface
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgRecordBatch{}, "ophost/MsgRecordBatch")
	legacy.RegisterAminoMsg(cdc, &MsgCreateBridge{}, "ophost/MsgCreateBridge")
	legacy.RegisterAminoMsg(cdc, &MsgProposeOutput{}, "ophost/MsgProposeOutput")
	legacy.RegisterAminoMsg(cdc, &MsgDeleteOutput{}, "ophost/MsgDeleteOutput")
	legacy.RegisterAminoMsg(cdc, &MsgInitiateTokenDeposit{}, "ophost/MsgInitiateTokenDeposit")
	legacy.RegisterAminoMsg(cdc, &MsgFinalizeTokenWithdrawal{}, "ophost/MsgFinalizeTokenWithdrawal")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateProposer{}, "ophost/MsgUpdateProposer")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateChallenger{}, "ophost/MsgUpdateChallenger")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateBatchInfo{}, "ophost/MsgUpdateBatchInfo")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "ophost/MsgUpdateParams")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateMetadata{}, "ophost/MsgUpdateMetadata")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateOracleConfig{}, "ophost/MsgUpdateOracleConfig")

	cdc.RegisterConcrete(Params{}, "ophost/Params", nil)
	cdc.RegisterConcrete(&BridgeAccount{}, "ophost/BridgeAccount", nil)
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
		&MsgUpdateBatchInfo{},
		&MsgUpdateParams{},
		&MsgUpdateMetadata{},
		&MsgUpdateOracleConfig{},
	)

	// auth account registration
	registry.RegisterImplementations(
		(*sdk.AccountI)(nil),
		&BridgeAccount{},
	)
	registry.RegisterImplementations(
		(*authtypes.GenesisAccount)(nil),
		&BridgeAccount{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
