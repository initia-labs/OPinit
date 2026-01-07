package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"

	"github.com/initia-labs/OPinit/x/opchild/types"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
)

func TestDecodePacketData_Success(t *testing.T) {
	// Create valid attestor
	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey()
	pkAny, err := codectypes.NewAnyWithValue(pubKey)
	require.NoError(t, err)

	attestor := ophosttypes.Attestor{
		OperatorAddress: "init1test",
		ConsensusPubkey: pkAny,
		Moniker:         "test-attestor",
	}

	// Create packet data
	packetData := ophosttypes.NewAttestorSetUpdatePacketData(
		1,
		[]ophosttypes.Attestor{attestor},
		100,
	)

	// Encode packet data
	packetBytes := packetData.GetBytes()
	require.NotNil(t, packetBytes)

	// Decode packet data
	decoded, err := types.DecodePacketData(packetBytes)
	require.NoError(t, err)
	require.Equal(t, uint64(1), decoded.BridgeId)
	require.Equal(t, uint64(100), decoded.L1BlockHeight)
	require.Len(t, decoded.AttestorSet, 1)
	require.Equal(t, "init1test", decoded.AttestorSet[0].OperatorAddress)
	require.Equal(t, "test-attestor", decoded.AttestorSet[0].Moniker)
}

func TestDecodePacketData_EmptyAttestorSet(t *testing.T) {
	// Create packet data with empty attestor set
	packetData := ophosttypes.NewAttestorSetUpdatePacketData(
		1,
		[]ophosttypes.Attestor{},
		100,
	)

	// Encode packet data
	packetBytes := packetData.GetBytes()
	require.NotNil(t, packetBytes)

	// Decode packet data
	decoded, err := types.DecodePacketData(packetBytes)
	require.NoError(t, err)
	require.Equal(t, uint64(1), decoded.BridgeId)
	require.Equal(t, uint64(100), decoded.L1BlockHeight)
	require.Len(t, decoded.AttestorSet, 0)
}

func TestDecodePacketData_MultipleAttestors(t *testing.T) {
	// Create multiple attestors
	attestors := make([]ophosttypes.Attestor, 5)
	for i := 0; i < 5; i++ {
		privKey := ed25519.GenPrivKey()
		pubKey := privKey.PubKey()
		pkAny, err := codectypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)

		attestors[i] = ophosttypes.Attestor{
			OperatorAddress: "init1test" + string(rune('0'+i)),
			ConsensusPubkey: pkAny,
			Moniker:         "attestor-" + string(rune('0'+i)),
		}
	}

	// Create packet data
	packetData := ophosttypes.NewAttestorSetUpdatePacketData(
		1,
		attestors,
		100,
	)

	// Encode packet data
	packetBytes := packetData.GetBytes()
	require.NotNil(t, packetBytes)

	// Decode packet data
	decoded, err := types.DecodePacketData(packetBytes)
	require.NoError(t, err)
	require.Equal(t, uint64(1), decoded.BridgeId)
	require.Len(t, decoded.AttestorSet, 5)

	// Verify all attestors
	for i := 0; i < 5; i++ {
		require.Equal(t, attestors[i].OperatorAddress, decoded.AttestorSet[i].OperatorAddress)
		require.Equal(t, attestors[i].Moniker, decoded.AttestorSet[i].Moniker)
	}
}

func TestDecodePacketData_InvalidData(t *testing.T) {
	testCases := []struct {
		name        string
		packetData  []byte
		expectedErr string
	}{
		{
			name:        "empty data",
			packetData:  []byte{},
			expectedErr: "invalid request",
		},
		{
			name:        "invalid json",
			packetData:  []byte("invalid-json"),
			expectedErr: "invalid request",
		},
		{
			name:        "random bytes",
			packetData:  []byte{0x00, 0x01, 0x02, 0x03},
			expectedErr: "invalid request",
		},
		{
			name:        "partial json",
			packetData:  []byte(`{"bridge_id": 1`),
			expectedErr: "invalid request",
		},
		{
			name:        "malformed json",
			packetData:  []byte(`{"bridge_id": "not-a-number"}`),
			expectedErr: "invalid request",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := types.DecodePacketData(tc.packetData)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

func TestDecodePacketData_LargeBridgeId(t *testing.T) {
	// Create packet data with max uint64 bridge ID
	packetData := ophosttypes.NewAttestorSetUpdatePacketData(
		^uint64(0), // Max uint64
		[]ophosttypes.Attestor{},
		100,
	)

	// Encode packet data
	packetBytes := packetData.GetBytes()
	require.NotNil(t, packetBytes)

	// Decode packet data
	decoded, err := types.DecodePacketData(packetBytes)
	require.NoError(t, err)
	require.Equal(t, ^uint64(0), decoded.BridgeId)
}

func TestDecodePacketData_LargeBlockHeight(t *testing.T) {
	// Create packet data with max uint64 block height
	packetData := ophosttypes.NewAttestorSetUpdatePacketData(
		1,
		[]ophosttypes.Attestor{},
		^uint64(0), // Max uint64
	)

	// Encode packet data
	packetBytes := packetData.GetBytes()
	require.NotNil(t, packetBytes)

	// Decode packet data
	decoded, err := types.DecodePacketData(packetBytes)
	require.NoError(t, err)
	require.Equal(t, ^uint64(0), decoded.L1BlockHeight)
}
