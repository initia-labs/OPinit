package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"

	"github.com/initia-labs/OPinit/x/opchild/types"
)

func TestOracleData_Validate(t *testing.T) {
	validOracleData := types.OracleData{
		BridgeId:       1,
		CurrencyPair:   "BTC/USD",
		Price:          "50000000000",
		Decimals:       8,
		L1BlockHeight:  100,
		L1BlockTime:    1000000000,
		CurrencyPairId: 1,
		Nonce:          1,
		Proof:          []byte("valid-proof"),
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
			name: "empty currency pair",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.CurrencyPair = ""
				return data
			}(),
			expectedErr: "currency pair cannot be empty",
		},
		{
			name: "whitespace currency pair",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.CurrencyPair = "   "
				return data
			}(),
			expectedErr: "currency pair cannot be empty",
		},
		{
			name: "invalid currency pair format - no slash",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.CurrencyPair = "BTCUSD"
				return data
			}(),
			expectedErr: "invalid currency pair format",
		},
		{
			name: "empty price",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.Price = ""
				return data
			}(),
			expectedErr: "price cannot be empty",
		},
		{
			name: "whitespace price",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.Price = "   "
				return data
			}(),
			expectedErr: "price cannot be empty",
		},
		{
			name: "invalid price format - not a number",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.Price = "not-a-number"
				return data
			}(),
			expectedErr: "invalid price format",
		},
		{
			name: "decimals too large",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.Decimals = 19
				return data
			}(),
			expectedErr: "decimals too large",
		},
		{
			name: "max valid decimals",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.Decimals = 18
				return data
			}(),
			expectedErr: "",
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
		{
			name: "zero nonce is valid",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.Nonce = 0
				return data
			}(),
			expectedErr: "",
		},
		{
			name: "valid data with different currency pair",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.CurrencyPair = "ETH/USD"
				return data
			}(),
			expectedErr: "",
		},
		{
			name: "valid data with timestamp currency pair",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.CurrencyPair = "TIMESTAMP/NANOSECOND"
				return data
			}(),
			expectedErr: "",
		},
		{
			name: "very large price",
			oracleData: func() types.OracleData {
				data := validOracleData
				data.Price = "999999999999999999999999999999"
				return data
			}(),
			expectedErr: "",
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
			BridgeId:       1,
			CurrencyPair:   "A/B",
			Price:          "1",
			Decimals:       0,
			L1BlockHeight:  1,
			L1BlockTime:    1,
			CurrencyPairId: 0,
			Nonce:          0,
			Proof:          []byte{0x00},
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
			BridgeId:       ^uint64(0),
			CurrencyPair:   "VERYLONGCURRENCYPAIR/ANOTHERLONGONE",
			Price:          "99999999999999999999999999999999999999",
			Decimals:       18,
			L1BlockHeight:  ^uint64(0),
			L1BlockTime:    9223372036854775807, // Max int64
			CurrencyPairId: ^uint64(0),
			Nonce:          ^uint64(0),
			Proof:          make([]byte, 10000),
			ProofHeight: clienttypes.Height{
				RevisionNumber: ^uint64(0),
				RevisionHeight: ^uint64(0),
			},
		}

		err := oracleData.Validate()
		require.NoError(t, err)
	})
}
