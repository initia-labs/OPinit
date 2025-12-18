package types

const (
	// ModuleName is the name of the ophost module
	ModuleName = "ophost"

	// Version defines the current version for ophost IBC module
	Version = "opinit-1"

	// PortID is the default port id for ophost module
	PortID = "opinit"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// RouterKey is the msg router key for the ophost module
	RouterKey = ModuleName
)

var (
	// Keys for store prefixes
	NextBridgeIdKey = []byte{0x11}
	ParamsKey       = []byte{0x12}

	BridgeConfigPrefix     = []byte{0x21}
	NextL1SequencePrefix   = []byte{0x31}
	TokenPairPrefix        = []byte{0x41}
	OutputProposalPrefix   = []byte{0x51}
	NextOutputIndexPrefix  = []byte{0x61}
	ProvenWithdrawalPrefix = []byte{0x71}
	BatchInfoPrefix        = []byte{0x81}
	MigrationInfoPrefix    = []byte{0x91}
	OraclePriceHashPrefix  = []byte{0xa1}
)
