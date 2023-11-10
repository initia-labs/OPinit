package types

import (
	"encoding/binary"
	fmt "fmt"

	"golang.org/x/crypto/sha3"
)

const L2_DENOM_PREFIX = "l2/"

func L2Denom(bridgeId uint64, l1Denom string) string {
	var bz []byte
	bz = binary.BigEndian.AppendUint64(bz, bridgeId)
	bz = append(bz, []byte(l1Denom)...)

	hash := sha3.Sum256(bz)
	return fmt.Sprintf("%s%x", L2_DENOM_PREFIX, hash[:])
}
