package types

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/codec"
	codecaddress "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/gogoproto/proto"
)

func Test_JSONMarshalUnmarshal(t *testing.T) {
	batchInfo := BatchInfo{
		Submitter: "submitter",
		ChainType: BatchInfo_CHAIN_TYPE_INITIA,
	}

	bz, err := json.Marshal(batchInfo)
	require.NoError(t, err)
	require.Equal(t, `{"submitter":"submitter","chain_type":"INITIA"}`, string(bz))

	var batchInfo1 BatchInfo
	err = json.Unmarshal(bz, &batchInfo1)
	require.NoError(t, err)
	require.Equal(t, batchInfo, batchInfo1)
}

func Test_ValidateBridgeConfig(t *testing.T) {
	config := BridgeConfig{
		Proposer:              "proposer",
		Challenger:            "challenger",
		SubmissionInterval:    100,
		FinalizationPeriod:    100,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             BatchInfo{Submitter: "submitter", ChainType: BatchInfo_CHAIN_TYPE_INITIA},
	}

	require.NoError(t, config.ValidateWithNoAddrValidation())

	// 1. wrong batch info chain type
	// 1.1 unspecified
	config.BatchInfo.ChainType = BatchInfo_CHAIN_TYPE_UNSPECIFIED
	require.Error(t, config.ValidateWithNoAddrValidation())

	// 1.2 unknown chain type
	config.BatchInfo.ChainType = 100
	require.Error(t, config.ValidateWithNoAddrValidation())
}

func TestGoGoProtoJsonPB(t *testing.T) {
	interfaceRegistry, err := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec:          codecaddress.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
			ValidatorAddressCodec: codecaddress.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		},
	})
	if err != nil {
		panic(err)
	}
	RegisterInterfaces(interfaceRegistry)
	std.RegisterInterfaces(interfaceRegistry)
	appCodec := codec.NewProtoCodec(interfaceRegistry)
	const CHAIN_TYPE = "CELESTIA"
	jsonText := fmt.Sprintf(`{"body":{"messages":[{"@type":"/opinit.ophost.v1.MsgCreateBridge","creator":"init1fgwvmvr8y5ul03aty9647vny7uj0uzt4g5zxev","config":{"challenger":"init1n8x7h4l96wmazlmxpqurccfg37fayrdafpmhfr","proposer":"init1knt8ehj03wk6hrzr4j35rmx2m7gqf4mwwcmrze","batch_info":{"submitter":"celestia13ycasdutemyk6pw6dy3ct3rxwknxrz6ygzjmu7","chain_type":"%s"},"submission_interval":"60s","finalization_period":"604800s","submission_start_height":"1","oracle_enabled":true,"metadata":"eyJwZXJtX2NoYW5uZWxzIjpbeyJwb3J0X2lkIjoibmZ0LXRyYW5zZmVyIiwiY2hhbm5lbF9pZCI6ImNoYW5uZWwtMzUzIn0seyJwb3J0X2lkIjoidHJhbnNmZXIiLCJjaGFubmVsX2lkIjoiY2hhbm5lbC0zNTIifV19"}}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[{"public_key":{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AwKanlZ3AymXu7E54dOD5gL5kSrSBvTdrcow/vOsTVfl"},"mode_info":{"single":{"mode":"SIGN_MODE_DIRECT"}},"sequence":"7"}],"fee":{"amount":[{"denom":"uinit","amount":"3000"}],"gas_limit":"200000","payer":"","granter":""},"tip":null},"signatures":["4x52q4ac4+Xbg/7PXzz73vhQrZgnL+FKsLuCYDK4f1EFOnsXvFAtyn+cqXwzXTJiKY+ZUoi4QfC4u+LEj0IZtg=="]}`, CHAIN_TYPE)

	var tx tx.Tx
	err = appCodec.UnmarshalJSON([]byte(jsonText), &tx)
	require.NoError(t, err)
	require.Len(t, tx.Body.Messages, 1)

	var msg sdk.Msg
	err = appCodec.UnpackAny(tx.Body.Messages[0], &msg)
	require.NoError(t, err)
	require.Equal(t, BatchInfo_CHAIN_TYPE_CELESTIA, msg.(*MsgCreateBridge).Config.BatchInfo.ChainType)
}
