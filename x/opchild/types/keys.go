package types

const (
	// ModuleName is the name of the opchild module
	// module addr: init1gz9n8jnu9fgqw7vem9ud67gqjk5q4m2w0aejne
	ModuleName = "opchild"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// RouterKey is the msg router key for the opchild module
	RouterKey = ModuleName
)

var (
	// Keys for store prefixes
	ParamsKey              = []byte{0x11} // prefix for parameters for module x/opchild
	NextL2SequenceKey      = []byte{0x12} // key for the outbound sequence number
	BridgeInfoKey          = []byte{0x13} // prefix for bridge_info
	FinalizedL1SequenceKey = []byte{0x14} // prefix for finalized deposit sequences

	LastValidatorPowerPrefix   = []byte{0x21} // prefix for each key to a validator index, for bonded validators
	ValidatorsPrefix           = []byte{0x31} // prefix for each key to a validator
	ValidatorsByConsAddrPrefix = []byte{0x41} // prefix for each key to a validator index, by pubkey
	HistoricalInfoPrefix       = []byte{0x61} // prefix for the historical info
	DenomPairPrefix            = []byte{0x71} // prefix for the denom pair

	// HostValidatorStore keys
	HostHeightKey        = []byte{0x81}
	HostValidatorsPrefix = []byte{0x82}
)
