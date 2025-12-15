package types

import (
	"crypto/sha256"
	"encoding/binary"
	"sort"

	"cosmossdk.io/math"
)

// OraclePriceInfo represents price information for hash computation.
type OraclePriceInfo struct {
	CurrencyPairId     uint64
	CurrencyPairString string
	Price              math.Int
	Timestamp          int64
}

type OraclePriceInfos []OraclePriceInfo

// ComputeOraclePricesHash computes a deterministic hash of oracle prices.
// Prices are sorted by currency pair ID before hashing to ensure determinism.
func (op OraclePriceInfos) ComputeOraclePricesHash() []byte {
	if len(op) == 0 {
		return nil
	}

	sorted := make([]OraclePriceInfo, len(op))
	copy(sorted, op)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CurrencyPairId < sorted[j].CurrencyPairId
	})

	hasher := sha256.New()
	for _, p := range sorted {
		idBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(idBytes, p.CurrencyPairId)
		hasher.Write(idBytes)

		// write currency pair string (authenticated to prevent price misdirection)
		hasher.Write([]byte(p.CurrencyPairString))

		// write price as string
		hasher.Write([]byte(p.Price.String()))

		// write timestamp (8 bytes)
		tsBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(tsBytes, uint64(p.Timestamp)) //nolint:gosec
		hasher.Write(tsBytes)
	}

	return hasher.Sum(nil)
}
