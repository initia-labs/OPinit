package types

import (
	"encoding/binary"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

const (
	// ModuleName is the name of the op_child module
	// module addr: init1vl25je2ntvjy7u9dnz9qzju674vfe25tkhhp92
	ModuleName = "op_child"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// RouterKey is the msg router key for the op_child module
	RouterKey = ModuleName
)

var (
	// Keys for store prefixes
	// Last* values are constant during a block.
	LastValidatorPowerKey = []byte{0x11} // prefix for each key to a validator index, for bonded validators
	ProposerKey           = []byte{0x12} // key for the proposer operator address

	ValidatorsKey           = []byte{0x21} // prefix for each key to a validator
	ValidatorsByConsAddrKey = []byte{0x22} // prefix for each key to a validator index, by pubkey

	HistoricalInfoKey   = []byte{0x31} // prefix for the historical info
	ValidatorUpdatesKey = []byte{0x32} // prefix for the end block validator updates key

	ParamsKey = []byte{0x41} // prefix for parameters for module x/op_child

	NextL2SequenceKey      = []byte{0x51} // key for the outbound sequence number
	FinalizedL1SequenceKey = []byte{0x62} // prefix for finalized deposit sequences
)

// GetValidatorKey creates the key for the validator with address
// VALUE: op_child/Validator
func GetValidatorKey(operatorAddr sdk.ValAddress) []byte {
	return append(ValidatorsKey, address.MustLengthPrefix(operatorAddr)...)
}

// GetValidatorByConsAddrKey creates the key for the validator with pubkey
// VALUE: validator operator address ([]byte)
func GetValidatorByConsAddrKey(addr sdk.ConsAddress) []byte {
	return append(ValidatorsByConsAddrKey, address.MustLengthPrefix(addr)...)
}

// AddressFromValidatorsKey creates the validator operator address from ValidatorsKey
func AddressFromValidatorsKey(key []byte) []byte {
	kv.AssertKeyAtLeastLength(key, 3)
	return key[2:] // remove prefix bytes and address length
}

// AddressFromLastValidatorPowerKey creates the validator operator address from LastValidatorPowerKey
func AddressFromLastValidatorPowerKey(key []byte) []byte {
	kv.AssertKeyAtLeastLength(key, 3)
	return key[2:] // remove prefix bytes and address length
}

// GetLastValidatorPowerKey creates the bonded validator index key for an operator address
func GetLastValidatorPowerKey(operator sdk.ValAddress) []byte {
	return append(LastValidatorPowerKey, address.MustLengthPrefix(operator)...)
}

// GetHistoricalInfoKey returns a key prefix for indexing HistoricalInfo objects.
func GetHistoricalInfoKey(height int64) []byte {
	return append(HistoricalInfoKey, []byte(strconv.FormatInt(height, 10))...)
}

func GetFinalizedL1SequenceKey(sequence uint64) []byte {
	_sequence := [8]byte{}
	binary.BigEndian.PutUint64(_sequence[:], sequence)

	return append(FinalizedL1SequenceKey, _sequence[:]...)
}
