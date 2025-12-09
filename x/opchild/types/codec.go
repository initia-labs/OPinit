package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/gogoproto/proto"
)

// RegisterLegacyAminoCodec registers the move types and interface
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {

	legacy.RegisterAminoMsg(cdc, &MsgExecuteMessages{}, "opchild/MsgExecuteMessages")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateSequencer{}, "opchild/MsgUpdateSequencer")
	legacy.RegisterAminoMsg(cdc, &MsgAddFeeWhitelistAddresses{}, "opchild/MsgAddFeeWhitelistAddresses")
	legacy.RegisterAminoMsg(cdc, &MsgRemoveFeeWhitelistAddresses{}, "opchild/MsgRemoveFeeWhitelistAddresses")
	legacy.RegisterAminoMsg(cdc, &MsgAddBridgeExecutor{}, "opchild/MsgAddBridgeExecutor")
	legacy.RegisterAminoMsg(cdc, &MsgRemoveBridgeExecutor{}, "opchild/MsgRemoveBridgeExecutor")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateMinGasPrices{}, "opchild/MsgUpdateMinGasPrices")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateAdmin{}, "opchild/MsgUpdateAdmin")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "opchild/MsgUpdateParams")
	legacy.RegisterAminoMsg(cdc, &MsgFinalizeTokenDeposit{}, "opchild/MsgFinalizeTokenDeposit")
	legacy.RegisterAminoMsg(cdc, &MsgInitiateTokenWithdrawal{}, "opchild/MsgInitiateTokenWithdrawal")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateOracle{}, "opchild/MsgUpdateOracle")
	legacy.RegisterAminoMsg(cdc, &MsgSetBridgeInfo{}, "opchild/MsgSetBridgeInfo")
	legacy.RegisterAminoMsg(cdc, &MsgSpendFeePool{}, "opchild/MsgSpendFeePool")
	legacy.RegisterAminoMsg(cdc, &MsgRegisterMigrationInfo{}, "opchild/MsgRegisterMigrationInfo")
	legacy.RegisterAminoMsg(cdc, &MsgMigrateToken{}, "opchild/MsgMigrateToken")
	legacy.RegisterAminoMsg(cdc, &MsgRelayOracleData{}, "opchild/MsgRelayOracleData")

	cdc.RegisterConcrete(Params{}, "opchild/Params", nil)
}

// RegisterInterfaces registers the x/market interfaces types with the interface registry
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgExecuteMessages{},
		&MsgUpdateSequencer{},
		&MsgAddFeeWhitelistAddresses{},
		&MsgRemoveFeeWhitelistAddresses{},
		&MsgAddBridgeExecutor{},
		&MsgRemoveBridgeExecutor{},
		&MsgUpdateMinGasPrices{},
		&MsgUpdateAdmin{},
		&MsgUpdateParams{},
		&MsgFinalizeTokenDeposit{},
		&MsgInitiateTokenWithdrawal{},
		&MsgUpdateOracle{},
		&MsgSetBridgeInfo{},
		&MsgSpendFeePool{},
		&MsgRegisterMigrationInfo{},
		&MsgMigrateToken{},
		&MsgRelayOracleData{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	anyResolver codectypes.InterfaceRegistry
	protoCodec  = codec.NewProtoCodec(anyResolver)
)

func init() {
	anyResolver = codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(anyResolver)
}

// unmarshalProtoJSON unmarshals JSON-encoded bytes into a protobuf message.
func unmarshalProtoJSON(bz []byte, ptr proto.Message) error {
	return protoCodec.UnmarshalJSON(bz, ptr)
}
