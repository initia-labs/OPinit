package types

import (
	bytes "bytes"
	"encoding/binary"

	"golang.org/x/crypto/sha3"
)

func (output Output) Validate() error {
	if len(output.OutputRoot) != 32 {
		return ErrInvalidHashLength.Wrap("output_root")
	}

	return nil
}

func (output Output) IsEmpty() bool {
	return len(output.OutputRoot) == 0 && output.L1BlockNumber == 0 && output.L2BlockNumber == 0
}

func GenerateOutputRoot(version byte, storageRoot []byte, latestBlockHash []byte) [32]byte {
	seed := make([]byte, 1+32+32)
	seed[0] = version
	copy(seed[1:], storageRoot[:32])
	copy(seed[1+32:], latestBlockHash[:32])
	return sha3.Sum256(seed)
}

func GenerateWithdrawalHash(bridgeId uint64, l2Sequence uint64, sender string, receiver string, denom string, amount uint64) [32]byte {
	var withdrawalHash [32]byte
	seed := []byte{}
	seed = binary.BigEndian.AppendUint64(seed, bridgeId)
	seed = binary.BigEndian.AppendUint64(seed, l2Sequence)

	// variable length
	senderDigest := sha3.Sum256([]byte(sender))
	seed = append(seed, senderDigest[:]...) // put utf8 encoded address
	// variable length
	receiverDigest := sha3.Sum256([]byte(receiver))
	seed = append(seed, receiverDigest[:]...) // put utf8 encoded address
	// variable length
	denomDigest := sha3.Sum256([]byte(denom))
	seed = append(seed, denomDigest[:]...)
	seed = binary.BigEndian.AppendUint64(seed, amount)

	// double hash the leaf node
	withdrawalHash = sha3.Sum256(seed)
	withdrawalHash = sha3.Sum256(withdrawalHash[:])

	return withdrawalHash
}

func GenerateNodeHash(a, b []byte) [32]byte {
	var data [32]byte
	switch bytes.Compare(a, b) {
	case 0, 1: // equal or greater
		data = sha3.Sum256(append(b, a...))
	case -1: // less
		data = sha3.Sum256(append(a, b...))
	}
	return data
}

func GenerateRootHashFromProofs(data [32]byte, proofs [][]byte) [32]byte {
	for _, proof := range proofs {
		data = GenerateNodeHash(data[:], proof)
	}
	return data
}
