package types

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/gogoproto/jsonpb"
	"github.com/cosmos/gogoproto/proto"
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
	legacy.RegisterAminoMsg(cdc, &MsgUpdateFinalizationPeriod{}, "ophost/MsgUpdateFinalizationPeriod")
	legacy.RegisterAminoMsg(cdc, &MsgRegisterMigrationInfo{}, "ophost/MsgRegisterMigrationInfo")
	legacy.RegisterAminoMsg(cdc, &MsgRegisterAttestorSet{}, "ophost/MsgRegisterAttestorSet")
	legacy.RegisterAminoMsg(cdc, &MsgAddAttestor{}, "ophost/MsgAddAttestor")
	legacy.RegisterAminoMsg(cdc, &MsgRemoveAttestor{}, "ophost/MsgRemoveAttestor")

	cdc.RegisterConcrete(Params{}, "ophost/Params", nil)
	cdc.RegisterConcrete(&BridgeAccount{}, "ophost/BridgeAccount", nil)
	cdc.RegisterConcrete(Attestor{}, "ophost/Attestor", nil)
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
		&MsgUpdateFinalizationPeriod{},
		&MsgRegisterMigrationInfo{},
		&MsgRegisterAttestorSet{},
		&MsgAddAttestor{},
		&MsgRemoveAttestor{},
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

var (
	anyResolver codectypes.InterfaceRegistry
)

func init() {
	anyResolver = codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(anyResolver)
}

// mustProtoMarshalJSON marshals a protobuf message to JSON and panics if there is an error.
func mustProtoMarshalJSON(msg proto.Message) []byte {
	bz, err := protoMarshalJSON(msg, anyResolver)
	if err != nil {
		panic(err)
	}
	return bz
}

// protoMarshalJSON provides an auxiliary function to return Proto3 JSON encoded bytes of a message.
func protoMarshalJSON(msg proto.Message, resolver jsonpb.AnyResolver) ([]byte, error) {
	jm := &jsonpb.Marshaler{OrigName: false, EmitDefaults: false, AnyResolver: resolver}
	err := codectypes.UnpackInterfaces(msg, codectypes.ProtoJSONPacker{JSONPBMarshaler: jm})
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if err := jm.Marshal(buf, msg); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
