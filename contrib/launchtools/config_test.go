package launchtools

import (
	"encoding/json"
	"testing"
	"time"

	ophosttypes "github.com/initia-labs/OPinit/v1/x/ophost/types"
	"github.com/stretchr/testify/require"
)

func Test_opBridgeConfig_JSON(t *testing.T) {
	submissionInterval := time.Second
	finalizationPeriod := time.Second * 10

	opBridgeConfig := &OpBridge{
		OutputSubmissionInterval:    &submissionInterval,
		OutputFinalizationPeriod:    &finalizationPeriod,
		OutputSubmissionStartHeight: 1,
		BatchSubmissionTarget:       ophosttypes.BatchInfo_CELESTIA,
	}

	bz, err := json.Marshal(opBridgeConfig)
	require.NoError(t, err)
	require.Equal(t, `{"output_submission_interval":"1s","output_finalization_period":"10s","output_submission_start_height":1,"batch_submission_target":"CELESTIA"}`, string(bz))

	var opBridgeConfig2 OpBridge
	err = json.Unmarshal(bz, &opBridgeConfig2)
	require.NoError(t, err)
	require.Equal(t, opBridgeConfig, &opBridgeConfig2)
}
