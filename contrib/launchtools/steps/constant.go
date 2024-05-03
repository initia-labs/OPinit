package steps

import "time"

const (
	KeyringBackend  = "test"
	RelayerPathName = "ibc"
	RelayerKeyName  = "Relayer"
	RelayerPathTemp = ".relayer"

	CreateEmptyBlocksInterval = 3 * time.Second
	OpBridgeIDKey             = "bridge_id"
)
