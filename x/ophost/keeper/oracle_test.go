package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/initia-labs/OPinit/x/ophost/testutil"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
)

// Test_UpdateOraclePriceHashes_NoCurrencyPairs tests the error handling when no currency pairs exist
func Test_UpdateOraclePriceHashes_NoCurrencyPairs(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	// should fail with no currency pairs
	err := input.OPHostKeeper.UpdateOraclePriceHashes(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no currency pairs found")
}

// Test_GetOraclePriceHash_NotFound tests getting oracle price hash when none exists
func Test_GetOraclePriceHash_NotFound(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	// try to get oracle price hash when none exists
	_, err := input.OPHostKeeper.GetOraclePriceHash(ctx)
	require.Error(t, err)
}

// Test_OraclePriceInfos_ComputeHash tests the hash computation for oracle price infos
func Test_OraclePriceInfos_ComputeHash(t *testing.T) {
	testTime := time.Now()

	prices := ophosttypes.OraclePriceInfos{
		{CurrencyPairId: 0, CurrencyPairString: "BTC/USD", Price: math.NewInt(10000000), Timestamp: testTime.UnixNano()},
		{CurrencyPairId: 1, CurrencyPairString: "ETH/USD", Price: math.NewInt(200000), Timestamp: testTime.UnixNano()},
	}

	hash1 := prices.ComputeOraclePricesHash()
	require.NotEmpty(t, hash1)

	// should be deterministic
	hash2 := prices.ComputeOraclePricesHash()
	require.Equal(t, hash1, hash2)
}

// Test_OraclePriceInfos_ComputeHash_DifferentPrices tests that different prices produce different hashes
func Test_OraclePriceInfos_ComputeHash_DifferentPrices(t *testing.T) {
	testTime := time.Now()

	prices1 := ophosttypes.OraclePriceInfos{
		{CurrencyPairId: 0, CurrencyPairString: "BTC/USD", Price: math.NewInt(10000000), Timestamp: testTime.UnixNano()},
	}

	prices2 := ophosttypes.OraclePriceInfos{
		{CurrencyPairId: 0, CurrencyPairString: "BTC/USD", Price: math.NewInt(11000000), Timestamp: testTime.UnixNano()},
	}

	hash1 := prices1.ComputeOraclePricesHash()
	hash2 := prices2.ComputeOraclePricesHash()

	require.NotEqual(t, hash1, hash2)
}

// Test_OraclePriceInfos_ComputeHash_DifferentTimestamps tests that different timestamps produce different hashes
func Test_OraclePriceInfos_ComputeHash_DifferentTimestamps(t *testing.T) {
	time1 := time.Now()
	time2 := time1.Add(time.Hour)

	prices1 := ophosttypes.OraclePriceInfos{
		{CurrencyPairId: 0, CurrencyPairString: "BTC/USD", Price: math.NewInt(10000000), Timestamp: time1.UnixNano()},
	}

	prices2 := ophosttypes.OraclePriceInfos{
		{CurrencyPairId: 0, CurrencyPairString: "BTC/USD", Price: math.NewInt(10000000), Timestamp: time2.UnixNano()},
	}

	hash1 := prices1.ComputeOraclePricesHash()
	hash2 := prices2.ComputeOraclePricesHash()

	require.NotEqual(t, hash1, hash2)
}

// Test_OraclePriceInfos_ComputeHash_OrderIndependent tests that hash computation is deterministic regardless of order
func Test_OraclePriceInfos_ComputeHash_OrderIndependent(t *testing.T) {
	testTime := time.Now()

	prices1 := ophosttypes.OraclePriceInfos{
		{CurrencyPairId: 0, CurrencyPairString: "BTC/USD", Price: math.NewInt(10000000), Timestamp: testTime.UnixNano()},
		{CurrencyPairId: 1, CurrencyPairString: "ETH/USD", Price: math.NewInt(200000), Timestamp: testTime.UnixNano()},
	}

	prices2 := ophosttypes.OraclePriceInfos{
		{CurrencyPairId: 1, CurrencyPairString: "ETH/USD", Price: math.NewInt(200000), Timestamp: testTime.UnixNano()},
		{CurrencyPairId: 0, CurrencyPairString: "BTC/USD", Price: math.NewInt(10000000), Timestamp: testTime.UnixNano()},
	}

	hash1 := prices1.ComputeOraclePricesHash()
	hash2 := prices2.ComputeOraclePricesHash()

	require.Equal(t, hash1, hash2, "Hash computation should be order-independent")
}
