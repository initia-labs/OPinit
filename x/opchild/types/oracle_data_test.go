package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"

	"github.com/initia-labs/OPinit/x/opchild/types"
)

func TestOracleData_Validate(t *testing.T) {
	validOracleData := types.OracleData{
		BridgeId:        1,
		OraclePriceHash: make([]byte, 32), // 32-byte hash
		Prices: []types.OraclePriceData{
			{
				CurrencyPair:   "BTC/USD",
				Price:          "50000000000",
				Decimals:       8,
				CurrencyPairId: 1,
				Nonce:          1,
			},
		},
		L1BlockHeight: 100,
		L1BlockTime:   1000000000,
		Proof:         []byte("valid-proof"),
		ProofHeight: clienttypes.Height{
			RevisionNumber: 0,
			RevisionHeight: 100,
		},
	}

	testCases := []struct {
		name        string
		oracleData  types.OracleData
		expectedErr string
	}{
		{
			name:        "valid oracle data",
			oracleData:  validOracleData,
			expectedErr: "",
		},
		{
			name: "zero bridge id",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.BridgeId = 0
				return data
			}(),
			expectedErr: "bridge id cannot be zero",
		},
		{
			name: "empty oracle price hash",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.OraclePriceHash = []byte{}
				return data
			}(),
			expectedErr: "oracle price hash cannot be empty",
		},
		{
			name: "nil oracle price hash",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.OraclePriceHash = nil
				return data
			}(),
			expectedErr: "oracle price hash cannot be empty",
		},
		{
			name: "oracle price hash wrong length",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.OraclePriceHash = make([]byte, 16) // Should be 32
				return data
			}(),
			expectedErr: "oracle price hash must be 32 bytes",
		},
		{
			name: "empty prices array",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.Prices = []types.OraclePriceData{}
				return data
			}(),
			expectedErr: "prices cannot be empty",
		},
		{
			name: "nil prices array",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.Prices = nil
				return data
			}(),
			expectedErr: "prices cannot be empty",
		},
		{
			name: "invalid price in batch - empty currency pair",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.Prices = []types.OraclePriceData{
					{
						CurrencyPair:   "",
						Price:          "50000000000",
						Decimals:       8,
						CurrencyPairId: 1,
						Nonce:          1,
					},
				}
				return data
			}(),
			expectedErr: "currency pair cannot be empty",
		},
		{
			name: "invalid price in batch - whitespace currency pair",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.Prices = []types.OraclePriceData{
					{
						CurrencyPair:   "   ",
						Price:          "50000000000",
						Decimals:       8,
						CurrencyPairId: 1,
						Nonce:          1,
					},
				}
				return data
			}(),
			expectedErr: "currency pair cannot be empty",
		},
		{
			name: "invalid price in batch - no slash",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.Prices = []types.OraclePriceData{
					{
						CurrencyPair:   "BTCUSD",
						Price:          "50000000000",
						Decimals:       8,
						CurrencyPairId: 1,
						Nonce:          1,
					},
				}
				return data
			}(),
			expectedErr: "invalid currency pair format",
		},
		{
			name: "invalid price in batch - empty price",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.Prices = []types.OraclePriceData{
					{
						CurrencyPair:   "BTC/USD",
						Price:          "",
						Decimals:       8,
						CurrencyPairId: 1,
						Nonce:          1,
					},
				}
				return data
			}(),
			expectedErr: "price cannot be empty",
		},
		{
			name: "invalid price in batch - not a number",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.Prices = []types.OraclePriceData{
					{
						CurrencyPair:   "BTC/USD",
						Price:          "not-a-number",
						Decimals:       8,
						CurrencyPairId: 1,
						Nonce:          1,
					},
				}
				return data
			}(),
			expectedErr: "invalid price format",
		},
		{
			name: "invalid price in batch - decimals too large",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.Prices = []types.OraclePriceData{
					{
						CurrencyPair:   "BTC/USD",
						Price:          "50000000000",
						Decimals:       37,
						CurrencyPairId: 1,
						Nonce:          1,
					},
				}
				return data
			}(),
			expectedErr: "decimals too large",
		},
		{
			name: "multiple prices in batch - all valid",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.Prices = []types.OraclePriceData{
					{
						CurrencyPair:   "BTC/USD",
						Price:          "50000000000",
						Decimals:       8,
						CurrencyPairId: 1,
						Nonce:          1,
					},
					{
						CurrencyPair:   "ETH/USD",
						Price:          "3000000000",
						Decimals:       8,
						CurrencyPairId: 2,
						Nonce:          1,
					},
					{
						CurrencyPair:   "ATOM/USD",
						Price:          "10000000",
						Decimals:       8,
						CurrencyPairId: 3,
						Nonce:          1,
					},
				}
				return data
			}(),
			expectedErr: "",
		},
		{
			name: "multiple prices - one invalid",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.Prices = []types.OraclePriceData{
					{
						CurrencyPair:   "BTC/USD",
						Price:          "50000000000",
						Decimals:       8,
						CurrencyPairId: 1,
						Nonce:          1,
					},
					{
						CurrencyPair:   "ETH/USD",
						Price:          "", // Invalid
						Decimals:       8,
						CurrencyPairId: 2,
						Nonce:          1,
					},
				}
				return data
			}(),
			expectedErr: "price cannot be empty",
		},
		{
			name: "zero l1 block height",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.L1BlockHeight = 0
				return data
			}(),
			expectedErr: "l1 block height cannot be zero",
		},
		{
			name: "zero l1 block time",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.L1BlockTime = 0
				return data
			}(),
			expectedErr: "invalid l1 block time",
		},
		{
			name: "negative l1 block time",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.L1BlockTime = -1
				return data
			}(),
			expectedErr: "invalid l1 block time",
		},
		{
			name: "empty proof",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.Proof = []byte{}
				return data
			}(),
			expectedErr: "proof cannot be empty",
		},
		{
			name: "nil proof",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.Proof = nil
				return data
			}(),
			expectedErr: "proof cannot be empty",
		},
		{
			name: "zero proof height revision height",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.ProofHeight.RevisionHeight = 0
				return data
			}(),
			expectedErr: "proof height revision height cannot be zero",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.oracleData.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr)
			}
		})
	}
}

func TestOracleData_Validate_EdgeCases(t *testing.T) {
	t.Run("all valid edge values", func(t *testing.T) {
		oracleData := types.OracleData{
			BridgeId:        1,
			OraclePriceHash: make([]byte, 32),
			Prices: []types.OraclePriceData{
				{
					CurrencyPair:   "A/B",
					Price:          "1",
					Decimals:       0,
					CurrencyPairId: 0,
					Nonce:          0,
				},
			},
			L1BlockHeight: 1,
			L1BlockTime:   1,
			Proof:         []byte{0x00},
			ProofHeight: clienttypes.Height{
				RevisionNumber: 0,
				RevisionHeight: 1,
			},
		}

		err := oracleData.Validate()
		require.NoError(t, err)
	})

	t.Run("large values", func(t *testing.T) {
		oracleData := types.OracleData{
			BridgeId:        ^uint64(0),
			OraclePriceHash: make([]byte, 32),
			Prices: []types.OraclePriceData{
				{
					CurrencyPair:   "VERYLONGCURRENCYPAIR/ANOTHERLONGONE",
					Price:          "99999999999999999999999999999999999999",
					Decimals:       18,
					CurrencyPairId: ^uint64(0),
					Nonce:          ^uint64(0),
				},
			},
			L1BlockHeight: ^uint64(0),
			L1BlockTime:   9223372036854775807, // Max int64
			Proof:         make([]byte, 10000),
			ProofHeight: clienttypes.Height{
				RevisionNumber: ^uint64(0),
				RevisionHeight: ^uint64(0),
			},
		}

		err := oracleData.Validate()
		require.NoError(t, err)
	})

	t.Run("max valid decimals in batch", func(t *testing.T) {
		oracleData := types.OracleData{
			BridgeId:        1,
			OraclePriceHash: make([]byte, 32),
			Prices: []types.OraclePriceData{
				{
					CurrencyPair:   "BTC/USD",
					Price:          "50000000000",
					Decimals:       36,
					CurrencyPairId: 1,
					Nonce:          1,
				},
			},
			L1BlockHeight: 100,
			L1BlockTime:   1000000000,
			Proof:         []byte("valid-proof"),
			ProofHeight: clienttypes.Height{
				RevisionNumber: 0,
				RevisionHeight: 100,
			},
		}

		err := oracleData.Validate()
		require.NoError(t, err)
	})

	t.Run("valid with timestamp currency pair", func(t *testing.T) {
		oracleData := types.OracleData{
			BridgeId:        1,
			OraclePriceHash: make([]byte, 32),
			Prices: []types.OraclePriceData{
				{
					CurrencyPair:   "TIMESTAMP/NANOSECOND",
					Price:          "1765532732375394000",
					Decimals:       8,
					CurrencyPairId: 0,
					Nonce:          571,
				},
			},
			L1BlockHeight: 100,
			L1BlockTime:   1000000000,
			Proof:         []byte("valid-proof"),
			ProofHeight: clienttypes.Height{
				RevisionNumber: 0,
				RevisionHeight: 100,
			},
		}

		err := oracleData.Validate()
		require.NoError(t, err)
	})

	t.Run("duplicate currency pairs allowed", func(t *testing.T) {
		oracleData := types.OracleData{
			BridgeId:        1,
			OraclePriceHash: make([]byte, 32),
			Prices: []types.OraclePriceData{
				{
					CurrencyPair:   "BTC/USD",
					Price:          "50000000000",
					Decimals:       8,
					CurrencyPairId: 1,
					Nonce:          1,
				},
				{
					CurrencyPair:   "BTC/USD",
					Price:          "50000000001",
					Decimals:       8,
					CurrencyPairId: 1,
					Nonce:          2,
				},
			},
			L1BlockHeight: 100,
			L1BlockTime:   1000000000,
			Proof:         []byte("valid-proof"),
			ProofHeight: clienttypes.Height{
				RevisionNumber: 0,
				RevisionHeight: 100,
			},
		}

		err := oracleData.Validate()
		require.NoError(t, err)
	})
}
