package steps

import "time"

const (
	KeyringBackend = "test"

	OperatorKeyName       = "Validator"
	BridgeExecutorKeyName = "BridgeExecutor"

	// this relayer is just for ibc setup
	// so we can use any address for this.
	RelayerKeyName = BridgeExecutorKeyName

	RelayerPathName = "ibc"
	RelayerPathTemp = ".relayer"

	CreateEmptyBlocksInterval = 3 * time.Second
	OpBridgeIDKey             = "bridge_id"
)
