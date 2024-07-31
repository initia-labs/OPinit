package types

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
)

func Test_JSONMarshalUnmarshal(t *testing.T) {
	batchInfo := BatchInfo{
		Submitter: "submitter",
		ChainType: BatchInfo_CHAIN_TYPE_INITIA,
	}

	bz, err := json.Marshal(batchInfo)
	require.NoError(t, err)
	require.Equal(t, `{"submitter":"submitter","chain_type":"INITIA"}`, string(bz))

	val, err := proto.MarshalJSONEnum(BatchInfo_ChainType_name, int32(batchInfo.ChainType))
	fmt.Println("SIBONG", string(val))
	require.Error(t, err)

	var batchInfo1 BatchInfo
	err = json.Unmarshal(bz, &batchInfo1)
	require.NoError(t, err)
	require.Equal(t, batchInfo, batchInfo1)
}

func Test_ValidateBridgeConfig(t *testing.T) {
	config := BridgeConfig{
		Proposer:              "proposer",
		Challengers:           []string{"challenger1", "challenger2"},
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
