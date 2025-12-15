package types_test

import (
	"testing"

	"cosmossdk.io/core/address"
	"github.com/stretchr/testify/require"

	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"

	"github.com/initia-labs/OPinit/x/opchild/types"
)

func TestNewMsgRelayOracleData(t *testing.T) {
	sender := "init1test"
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
		},
		L1BlockHeight: 100,
		L1BlockTime:   1000000000,
		Proof:         []byte("proof"),
		ProofHeight: clienttypes.Height{
			RevisionNumber: 0,
			RevisionHeight: 100,
		},
	}

	msg := types.NewMsgRelayOracleData(sender, oracleData)

	require.NotNil(t, msg)
	require.Equal(t, sender, msg.Sender)
	require.Equal(t, oracleData, msg.OracleData)
}

func TestMsgRelayOracleData_Validate(t *testing.T) {
	ac := authcodec.NewBech32Codec("init")

	validOracleData := types.OracleData{
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
		},
		L1BlockHeight: 100,
		L1BlockTime:   1000000000,
		Proof:         []byte("proof"),
		ProofHeight: clienttypes.Height{
			RevisionNumber: 0,
			RevisionHeight: 100,
		},
	}

	testCases := []struct {
		name        string
		msg         *types.MsgRelayOracleData
		expectedErr string
	}{
		{
			name: "valid message",
			msg: types.NewMsgRelayOracleData(
				"init1q6jhwnarkw2j5qqgx3qlu20k8nrdglft5ksr0g",
				validOracleData,
			),
			expectedErr: "",
		},
		{
			name: "invalid sender address - empty",
			msg: types.NewMsgRelayOracleData(
				"",
				validOracleData,
			),
			expectedErr: "empty address string is not allowed",
		},
		{
			name: "invalid sender address - wrong prefix",
			msg: types.NewMsgRelayOracleData(
				"cosmos1test",
				validOracleData,
			),
			expectedErr: "decoding bech32 failed",
		},
		{
			name: "invalid sender address - malformed",
			msg: types.NewMsgRelayOracleData(
				"invalid-address",
				validOracleData,
			),
			expectedErr: "decoding bech32 failed",
		},
		{
			name: "invalid oracle data - zero bridge id",
			msg: types.NewMsgRelayOracleData(
				"init1q6jhwnarkw2j5qqgx3qlu20k8nrdglft5ksr0g",
				types.OracleData{
					BridgeId:        0, // Invalid
					OraclePriceHash: make([]byte, 32),
					Prices: []types.OraclePriceData{
						{
							CurrencyPair:   "BTC/USD",
							Price:          "50000",
							Decimals:       5,
							CurrencyPairId: 1,
							Nonce:          1,
						},
					},
					L1BlockHeight: 100,
					L1BlockTime:   1000,
					Proof:         []byte("proof"),
					ProofHeight: clienttypes.Height{
						RevisionHeight: 100,
					},
				},
			),
			expectedErr: "bridge id cannot be zero",
		},
		{
			name: "invalid oracle data - empty currency pair",
			msg: types.NewMsgRelayOracleData(
				"init1q6jhwnarkw2j5qqgx3qlu20k8nrdglft5ksr0g",
				types.OracleData{
					BridgeId:        1,
					OraclePriceHash: make([]byte, 32),
					Prices: []types.OraclePriceData{
						{
							CurrencyPair:   "", // Invalid
							Price:          "50000",
							Decimals:       5,
							CurrencyPairId: 1,
							Nonce:          1,
						},
					},
					L1BlockHeight: 100,
					L1BlockTime:   1000,
					Proof:         []byte("proof"),
					ProofHeight: clienttypes.Height{
						RevisionHeight: 100,
					},
				},
			),
			expectedErr: "currency pair cannot be empty",
		},
		{
			name: "invalid oracle data - invalid price",
			msg: types.NewMsgRelayOracleData(
				"init1q6jhwnarkw2j5qqgx3qlu20k8nrdglft5ksr0g",
				types.OracleData{
					BridgeId:        1,
					OraclePriceHash: make([]byte, 32),
					Prices: []types.OraclePriceData{
						{
							CurrencyPair:   "BTC/USD",
							Price:          "not-a-number", // Invalid
							Decimals:       5,
							CurrencyPairId: 1,
							Nonce:          1,
						},
					},
					L1BlockHeight: 100,
					L1BlockTime:   1000,
					Proof:         []byte("proof"),
					ProofHeight: clienttypes.Height{
						RevisionHeight: 100,
					},
				},
			),
			expectedErr: "invalid price format",
		},
		{
			name: "invalid oracle data - decimals too large",
			msg: types.NewMsgRelayOracleData(
				"init1q6jhwnarkw2j5qqgx3qlu20k8nrdglft5ksr0g",
				types.OracleData{
					BridgeId:        1,
					OraclePriceHash: make([]byte, 32),
					Prices: []types.OraclePriceData{
						{
							CurrencyPair:   "BTC/USD",
							Price:          "50000",
							Decimals:       19, // Invalid - max is 18
							CurrencyPairId: 1,
							Nonce:          1,
						},
					},
					L1BlockHeight: 100,
					L1BlockTime:   1000,
					Proof:         []byte("proof"),
					ProofHeight: clienttypes.Height{
						RevisionHeight: 100,
					},
				},
			),
			expectedErr: "decimals too large",
		},
		{
			name: "invalid oracle data - zero block height",
			msg: types.NewMsgRelayOracleData(
				"init1q6jhwnarkw2j5qqgx3qlu20k8nrdglft5ksr0g",
				types.OracleData{
					BridgeId:        1,
					OraclePriceHash: make([]byte, 32),
					Prices: []types.OraclePriceData{
						{
							CurrencyPair:   "BTC/USD",
							Price:          "50000",
							Decimals:       5,
							CurrencyPairId: 1,
							Nonce:          1,
						},
					},
					L1BlockHeight: 0, // Invalid
					L1BlockTime:   1000,
					Proof:         []byte("proof"),
					ProofHeight: clienttypes.Height{
						RevisionHeight: 100,
					},
				},
			),
			expectedErr: "l1 block height cannot be zero",
		},
		{
			name: "invalid oracle data - invalid block time",
			msg: types.NewMsgRelayOracleData(
				"init1q6jhwnarkw2j5qqgx3qlu20k8nrdglft5ksr0g",
				types.OracleData{
					BridgeId:        1,
					OraclePriceHash: make([]byte, 32),
					Prices: []types.OraclePriceData{
						{
							CurrencyPair:   "BTC/USD",
							Price:          "50000",
							Decimals:       5,
							CurrencyPairId: 1,
							Nonce:          1,
						},
					},
					L1BlockHeight: 100,
					L1BlockTime:   0, // Invalid
					Proof:         []byte("proof"),
					ProofHeight: clienttypes.Height{
						RevisionHeight: 100,
					},
				},
			),
			expectedErr: "invalid l1 block time",
		},
		{
			name: "invalid oracle data - empty proof",
			msg: types.NewMsgRelayOracleData(
				"init1q6jhwnarkw2j5qqgx3qlu20k8nrdglft5ksr0g",
				types.OracleData{
					BridgeId:        1,
					OraclePriceHash: make([]byte, 32),
					Prices: []types.OraclePriceData{
						{
							CurrencyPair:   "BTC/USD",
							Price:          "50000",
							Decimals:       5,
							CurrencyPairId: 1,
							Nonce:          1,
						},
					},
					L1BlockHeight: 100,
					L1BlockTime:   1000,
					Proof:         []byte{}, // Invalid
					ProofHeight: clienttypes.Height{
						RevisionHeight: 100,
					},
				},
			),
			expectedErr: "proof cannot be empty",
		},
		{
			name: "invalid oracle data - zero proof height",
			msg: types.NewMsgRelayOracleData(
				"init1q6jhwnarkw2j5qqgx3qlu20k8nrdglft5ksr0g",
				types.OracleData{
					BridgeId:        1,
					OraclePriceHash: make([]byte, 32),
					Prices: []types.OraclePriceData{
						{
							CurrencyPair:   "BTC/USD",
							Price:          "50000",
							Decimals:       5,
							CurrencyPairId: 1,
							Nonce:          1,
						},
					},
					L1BlockHeight: 100,
					L1BlockTime:   1000,
					Proof:         []byte("proof"),
					ProofHeight: clienttypes.Height{
						RevisionHeight: 0, // Invalid
					},
				},
			),
			expectedErr: "proof height revision height cannot be zero",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.Validate(ac)
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr)
			}
		})
	}
}

func TestMsgRelayOracleData_ValidateBasic(t *testing.T) {
	// Test with different address codecs
	testCases := []struct {
		name   string
		codec  address.Codec
		sender string
	}{
		{
			name:   "init prefix",
			codec:  authcodec.NewBech32Codec("init"),
			sender: "init1q6jhwnarkw2j5qqgx3qlu20k8nrdglft5ksr0g",
		},
		{
			name:   "cosmos prefix",
			codec:  authcodec.NewBech32Codec("cosmos"),
			sender: "cosmos1q6jhwnarkw2j5qqgx3qlu20k8nrdglft6qssy2",
		},
		{
			name:   "osmo prefix",
			codec:  authcodec.NewBech32Codec("osmo"),
			sender: "osmo1q6jhwnarkw2j5qqgx3qlu20k8nrdglftjmrqjc",
		},
	}

	validOracleData := types.OracleData{
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
		},
		L1BlockHeight: 100,
		L1BlockTime:   1000000000,
		Proof:         []byte("proof"),
		ProofHeight: clienttypes.Height{
			RevisionNumber: 0,
			RevisionHeight: 100,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgRelayOracleData(tc.sender, validOracleData)
			err := msg.Validate(tc.codec)
			require.NoError(t, err)
		})
	}
}

func TestMsgRelayOracleData_EdgeCases(t *testing.T) {
	ac := authcodec.NewBech32Codec("init")

	t.Run("minimal valid oracle data", func(t *testing.T) {
		minimalData := types.OracleData{
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

		msg := types.NewMsgRelayOracleData("init1q6jhwnarkw2j5qqgx3qlu20k8nrdglft5ksr0g", minimalData)
		err := msg.Validate(ac)
		require.NoError(t, err)
	})

	t.Run("max valid decimals", func(t *testing.T) {
		maxDecimalsData := types.OracleData{
			BridgeId:        1,
			OraclePriceHash: make([]byte, 32),
			Prices: []types.OraclePriceData{
				{
					CurrencyPair:   "BTC/USD",
					Price:          "50000",
					Decimals:       18, // Max allowed
					CurrencyPairId: 1,
					Nonce:          1,
				},
			},
			L1BlockHeight: 100,
			L1BlockTime:   1000,
			Proof:         []byte("proof"),
			ProofHeight: clienttypes.Height{
				RevisionHeight: 100,
			},
		}

		msg := types.NewMsgRelayOracleData("init1q6jhwnarkw2j5qqgx3qlu20k8nrdglft5ksr0g", maxDecimalsData)
		err := msg.Validate(ac)
		require.NoError(t, err)
	})

	t.Run("very large price", func(t *testing.T) {
		largePriceData := types.OracleData{
			BridgeId:        1,
			OraclePriceHash: make([]byte, 32),
			Prices: []types.OraclePriceData{
				{
					CurrencyPair:   "BTC/USD",
					Price:          "999999999999999999999999999999",
					Decimals:       8,
					CurrencyPairId: 1,
					Nonce:          1,
				},
			},
			L1BlockHeight: 100,
			L1BlockTime:   1000,
			Proof:         []byte("proof"),
			ProofHeight: clienttypes.Height{
				RevisionHeight: 100,
			},
		}

		msg := types.NewMsgRelayOracleData("init1q6jhwnarkw2j5qqgx3qlu20k8nrdglft5ksr0g", largePriceData)
		err := msg.Validate(ac)
		require.NoError(t, err)
	})
}
