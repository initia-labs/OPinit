package keeper_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/initia-labs/OPinit/x/opchild/types"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
)

const testBridgeId = uint64(1)

func setTestBridgeInfo(t *testing.T, ctx context.Context, input TestKeepers) {
	if ok, err := input.OPChildKeeper.BridgeInfo.Has(ctx); err == nil && ok {
		return
	}
	require.NoError(t, input.OPChildKeeper.BridgeInfo.Set(ctx, types.BridgeInfo{
		BridgeId:   testBridgeId,
		BridgeAddr: addrsStr[1],
		L1ChainId:  "test-chain-id",
		L1ClientId: "test-client-id",
		BridgeConfig: ophosttypes.BridgeConfig{
			Challenger: addrsStr[2],
			Proposer:   addrsStr[3],
			BatchInfo: ophosttypes.BatchInfo{
				Submitter: addrsStr[4],
				ChainType: ophosttypes.BatchInfo_INITIA,
			},
			SubmissionInterval:    time.Minute,
			FinalizationPeriod:    time.Hour,
			SubmissionStartHeight: 1,
			Metadata:              []byte("metadata"),
		},
	}))
}

func testL2Denom(t *testing.T, ctx context.Context, input TestKeepers, baseDenom string) string {
	setTestBridgeInfo(t, ctx, input)
	return ophosttypes.L2Denom(testBridgeId, baseDenom)
}
