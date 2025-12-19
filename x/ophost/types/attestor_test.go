package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	codecaddress "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"

	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
)

func Test_ValidateAttestor(t *testing.T) {
	vc := codecaddress.NewBech32Codec("init")
	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := sdk.ValAddress(pubKey.Address())

	addrStr, err := vc.BytesToString(addr)
	require.NoError(t, err)

	pkAny, err := codectypes.NewAnyWithValue(pubKey)
	require.NoError(t, err)

	// valid attestor
	validAttestor := ophosttypes.Attestor{
		OperatorAddress: addrStr,
		ConsensusPubkey: pkAny,
		Moniker:         "valid-attestor",
	}

	err = ophosttypes.ValidateAttestor(validAttestor, vc)
	require.NoError(t, err)

	invalidAddrAttestor := ophosttypes.Attestor{
		OperatorAddress: "invalid-address",
		ConsensusPubkey: pkAny,
		Moniker:         "invalid-addr",
	}

	err = ophosttypes.ValidateAttestor(invalidAddrAttestor, vc)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid attestor address")

	nilPubkeyAttestor := ophosttypes.Attestor{
		OperatorAddress: addrStr,
		ConsensusPubkey: nil,
		Moniker:         "nil-pubkey",
	}

	err = ophosttypes.ValidateAttestor(nilPubkeyAttestor, vc)
	require.Error(t, err)
	require.Contains(t, err.Error(), "consensus pubkey cannot be nil")
}

func Test_ValidateAttestorNoAddrValidation(t *testing.T) {
	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey()

	pkAny, err := codectypes.NewAnyWithValue(pubKey)
	require.NoError(t, err)

	// valid attestor
	validAttestor := ophosttypes.Attestor{
		OperatorAddress: "some-address",
		ConsensusPubkey: pkAny,
		Moniker:         "valid-attestor",
	}

	err = ophosttypes.ValidateAttestorNoAddrValidation(validAttestor)
	require.NoError(t, err)

	emptyAddrAttestor := ophosttypes.Attestor{
		OperatorAddress: "",
		ConsensusPubkey: pkAny,
		Moniker:         "empty-addr",
	}

	err = ophosttypes.ValidateAttestorNoAddrValidation(emptyAddrAttestor)
	require.Error(t, err)
	require.Contains(t, err.Error(), "attestor address cannot be empty")

	nilPubkeyAttestor := ophosttypes.Attestor{
		OperatorAddress: "some-address",
		ConsensusPubkey: nil,
		Moniker:         "nil-pubkey",
	}

	err = ophosttypes.ValidateAttestorNoAddrValidation(nilPubkeyAttestor)
	require.Error(t, err)
	require.Contains(t, err.Error(), "consensus pubkey cannot be nil")
}

func Test_ValidateAttestorSet(t *testing.T) {
	vc := codecaddress.NewBech32Codec("init")

	privKey1 := ed25519.GenPrivKey()
	pubKey1 := privKey1.PubKey()
	addr1 := sdk.ValAddress(pubKey1.Address())
	addrStr1, err := vc.BytesToString(addr1)
	require.NoError(t, err)

	pkAny1, err := codectypes.NewAnyWithValue(pubKey1)
	require.NoError(t, err)

	attestor1 := ophosttypes.Attestor{
		OperatorAddress: addrStr1,
		ConsensusPubkey: pkAny1,
		Moniker:         "attestor1",
	}

	privKey2 := ed25519.GenPrivKey()
	pubKey2 := privKey2.PubKey()
	addr2 := sdk.ValAddress(pubKey2.Address())
	addrStr2, err := vc.BytesToString(addr2)
	require.NoError(t, err)

	pkAny2, err := codectypes.NewAnyWithValue(pubKey2)
	require.NoError(t, err)

	attestor2 := ophosttypes.Attestor{
		OperatorAddress: addrStr2,
		ConsensusPubkey: pkAny2,
		Moniker:         "attestor2",
	}

	validSet := []ophosttypes.Attestor{attestor1, attestor2}
	err = ophosttypes.ValidateAttestorSet(validSet, vc)
	require.NoError(t, err)

	duplicateAddrSet := []ophosttypes.Attestor{
		attestor1,
		{
			OperatorAddress: addrStr1, // Duplicate
			ConsensusPubkey: pkAny2,
			Moniker:         "duplicate-addr",
		},
	}

	err = ophosttypes.ValidateAttestorSet(duplicateAddrSet, vc)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate attestor address")

	duplicatePubkeySet := []ophosttypes.Attestor{
		attestor1,
		{
			OperatorAddress: addrStr2,
			ConsensusPubkey: pkAny1, // Duplicate
			Moniker:         "duplicate-pubkey",
		},
	}

	err = ophosttypes.ValidateAttestorSet(duplicatePubkeySet, vc)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate consensus pubkey")
}

func Test_ValidateAttestorSetNoAddrValidation(t *testing.T) {
	privKey1 := ed25519.GenPrivKey()
	pubKey1 := privKey1.PubKey()
	pkAny1, err := codectypes.NewAnyWithValue(pubKey1)
	require.NoError(t, err)

	attestor1 := ophosttypes.Attestor{
		OperatorAddress: "addr1",
		ConsensusPubkey: pkAny1,
		Moniker:         "attestor1",
	}

	privKey2 := ed25519.GenPrivKey()
	pubKey2 := privKey2.PubKey()
	pkAny2, err := codectypes.NewAnyWithValue(pubKey2)
	require.NoError(t, err)

	attestor2 := ophosttypes.Attestor{
		OperatorAddress: "addr2",
		ConsensusPubkey: pkAny2,
		Moniker:         "attestor2",
	}

	validSet := []ophosttypes.Attestor{attestor1, attestor2}
	err = ophosttypes.ValidateAttestorSetNoAddrValidation(validSet)
	require.NoError(t, err)

	duplicateAddrSet := []ophosttypes.Attestor{
		attestor1,
		{
			OperatorAddress: "addr1", // Duplicate
			ConsensusPubkey: pkAny2,
			Moniker:         "duplicate",
		},
	}

	err = ophosttypes.ValidateAttestorSetNoAddrValidation(duplicateAddrSet)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate attestor address")
}

func Test_NewAttestorSetUpdatePacketData(t *testing.T) {
	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey()
	pkAny, err := codectypes.NewAnyWithValue(pubKey)
	require.NoError(t, err)

	attestor := ophosttypes.Attestor{
		OperatorAddress: "addr1",
		ConsensusPubkey: pkAny,
		Moniker:         "attestor1",
	}

	bridgeId := uint64(1)
	l1BlockHeight := uint64(100)
	attestorSet := []ophosttypes.Attestor{attestor}

	packetData := ophosttypes.NewAttestorSetUpdatePacketData(bridgeId, attestorSet, l1BlockHeight)

	require.Equal(t, bridgeId, packetData.BridgeId)
	require.Equal(t, l1BlockHeight, packetData.L1BlockHeight)
	require.Len(t, packetData.AttestorSet, 1)
	require.Equal(t, attestor.OperatorAddress, packetData.AttestorSet[0].OperatorAddress)

	packetBytes := packetData.GetBytes()
	require.NotNil(t, packetBytes)
	require.NotEmpty(t, packetBytes)
}
