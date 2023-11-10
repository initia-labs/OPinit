package keeper_test

import (
	"testing"
	"time"

	"github.com/initia-labs/OPinit/x/ophost/types"
	"github.com/stretchr/testify/require"
)

func Test_BridgeConfig(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	config := types.BridgeConfig{
		Challenger:          addrs[0].String(),
		Proposer:            addrs[1].String(),
		SubmissionInterval:  time.Second * 100,
		FinalizationPeriod:  time.Second * 10,
		SubmissionStartTime: time.Now().UTC(),
		Metadata:            []byte{1, 2, 3},
	}
	require.NoError(t, input.OPHostKeeper.SetBridgeConfig(ctx, 1, config))
	_config, err := input.OPHostKeeper.GetBridgeConfig(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, config, _config)
}

func Test_IterateBridgeConfig(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	config1 := types.BridgeConfig{
		Challenger:          addrs[0].String(),
		Proposer:            addrs[1].String(),
		SubmissionInterval:  time.Second * 100,
		FinalizationPeriod:  time.Second * 10,
		SubmissionStartTime: time.Now().UTC(),
		Metadata:            []byte{1, 2, 3},
	}
	config2 := types.BridgeConfig{
		Challenger:          addrs[2].String(),
		Proposer:            addrs[3].String(),
		SubmissionInterval:  time.Second * 100,
		FinalizationPeriod:  time.Second * 10,
		SubmissionStartTime: time.Now().UTC(),
		Metadata:            []byte{3, 4, 5},
	}
	require.NoError(t, input.OPHostKeeper.SetBridgeConfig(ctx, 1, config1))
	require.NoError(t, input.OPHostKeeper.SetBridgeConfig(ctx, 2, config2))

	input.OPHostKeeper.IterateBridgeConfig(ctx, func(bridgeId uint64, bridgeConfig types.BridgeConfig) bool {
		if bridgeId == 1 {
			require.Equal(t, config1, bridgeConfig)
		} else {
			require.Equal(t, config2, bridgeConfig)
		}

		return false
	})
}
