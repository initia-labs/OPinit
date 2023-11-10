package types

import (
	"encoding/binary"
)

const (
	// ModuleName is the name of the ophost module
	ModuleName = "ophost"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// RouterKey is the msg router key for the ophost module
	RouterKey = ModuleName
)

var (

	// Keys for store prefixes
	NextBridgeIdKey = []byte{0x11}

	BridgeConfigKey   = []byte{0x21}
	NextL1SequenceKey = []byte{0x22}

	TokenPairKey = []byte{0x31}

	OutputProposalKey  = []byte{0x41}
	NextOutputIndexKey = []byte{0x42}

	ProvenWithdrawalKey = []byte{0x51}

	ParamsKey = []byte{0x61}
)

// GetBridgeConfigKey creates the key for the bridge config with bridge id
func GetBridgeConfigKey(bridgeId uint64) []byte {
	_bridgeId := [8]byte{}
	binary.BigEndian.PutUint64(_bridgeId[:], bridgeId)
	return append(BridgeConfigKey, _bridgeId[:]...)
}

// GetNextL1SequenceKey creates the key for the next bridge sequence with bridge id
func GetNextL1SequenceKey(bridgeId uint64) []byte {
	_bridgeId := [8]byte{}
	binary.BigEndian.PutUint64(_bridgeId[:], bridgeId)
	return append(NextL1SequenceKey, _bridgeId[:]...)
}

// GetTokenPairBridgePrefixKey creates the prefix key for the token pair info with bridge id
func GetTokenPairBridgePrefixKey(bridgeId uint64) []byte {
	_bridgeId := [8]byte{}
	binary.BigEndian.PutUint64(_bridgeId[:], bridgeId)
	return append(TokenPairKey, _bridgeId[:]...)
}

// GetTokenPairKey creates the key for the token pair info with bridge id and l2Denom
func GetTokenPairKey(bridgeId uint64, l2Denom string) []byte {
	_bridgeId := [8]byte{}
	binary.BigEndian.PutUint64(_bridgeId[:], bridgeId)
	return append(append(TokenPairKey, _bridgeId[:]...), []byte(l2Denom)...)
}

func GetOutputProposalBridgePrefixKey(bridgeId uint64) []byte {
	_bridgeId := [8]byte{}
	binary.BigEndian.PutUint64(_bridgeId[:], bridgeId)

	return append(OutputProposalKey, _bridgeId[:]...)
}

// GetOutputProposalKey creates the key for the output proposal with bridge id and outputIndex.
func GetOutputProposalKey(bridgeId, outputIndex uint64) []byte {
	_bridgeId := [8]byte{}
	binary.BigEndian.PutUint64(_bridgeId[:], bridgeId)

	_outputIndex := [8]byte{}
	binary.BigEndian.PutUint64(_outputIndex[:], outputIndex)

	return append(append(OutputProposalKey, _bridgeId[:]...), _outputIndex[:]...)
}

// GetNextOutputIndexKey creates the key for the next output index with bridge id.
func GetNextOutputIndexKey(bridgeId uint64) []byte {
	_bridgeId := [8]byte{}
	binary.BigEndian.PutUint64(_bridgeId[:], bridgeId)
	return append(NextOutputIndexKey, _bridgeId[:]...)
}

func GetProvenWithdrawalPrefixKey(bridgeId uint64) []byte {
	_bridgeId := [8]byte{}
	binary.BigEndian.PutUint64(_bridgeId[:], bridgeId)

	return append(ProvenWithdrawalKey, _bridgeId[:]...)
}

// GetProvenWithdrawalKey creates the key for the proven withdrawal with bridge id and withdrawal hash.
func GetProvenWithdrawalKey(bridgeId uint64, withdrawalHash [32]byte) []byte {
	_bridgeId := [8]byte{}
	binary.BigEndian.PutUint64(_bridgeId[:], bridgeId)
	return append(append(ProvenWithdrawalKey, _bridgeId[:]...), withdrawalHash[:]...)
}
