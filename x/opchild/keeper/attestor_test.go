package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testutilsims "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"

	opchildtypes "github.com/initia-labs/OPinit/x/opchild/types"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
)

func Test_HandleAttestorSetUpdatePacket_Success(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	bridgeInfo := opchildtypes.BridgeInfo{
		BridgeId:   1,
		BridgeAddr: addrsStr[0],
		L1ChainId:  "test-chain-1",
		L1ClientId: "test-client-id",
		BridgeConfig: ophosttypes.BridgeConfig{
			AttestorSet: []ophosttypes.Attestor{},
		},
	}
	err := input.OPChildKeeper.BridgeInfo.Set(ctx, bridgeInfo)
	require.NoError(t, err)

	valPubKeys := testutilsims.CreateTestPubKeys(3)
	attestor1 := createAttestor(t, valAddrsStr[1], valPubKeys[0], "attestor1")
	attestor2 := createAttestor(t, valAddrsStr[2], valPubKeys[1], "attestor2")
	attestor3 := createAttestor(t, valAddrsStr[3], valPubKeys[2], "attestor3")

	packetData := ophosttypes.AttestorSetUpdatePacketData{
		BridgeId:      1,
		AttestorSet:   []ophosttypes.Attestor{attestor1, attestor2, attestor3},
		L1BlockHeight: 100,
	}

	ack, err := input.OPChildKeeper.HandleAttestorSetUpdatePacket(ctx, packetData)
	require.NoError(t, err)
	require.True(t, ack.Success)
	require.Empty(t, ack.Error)

	vals, err := input.OPChildKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	require.Len(t, vals, 3)

	for _, val := range vals {
		require.Equal(t, int64(opchildtypes.AttestorConsPower), val.ConsPower)
	}

	updatedBridgeInfo, err := input.OPChildKeeper.BridgeInfo.Get(ctx)
	require.NoError(t, err)
	require.Len(t, updatedBridgeInfo.BridgeConfig.AttestorSet, 3)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	events := sdkCtx.EventManager().Events()
	found := false
	for _, event := range events {
		if event.Type == opchildtypes.EventTypeAttestorSetUpdate {
			found = true
			break
		}
	}
	require.True(t, found, "expected attestor set update event to be emitted")
}

func Test_HandleAttestorSetUpdatePacket_BridgeIdMismatch(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	bridgeInfo := opchildtypes.BridgeInfo{
		BridgeId:   1,
		BridgeAddr: addrsStr[0],
		L1ChainId:  "test-chain-1",
		L1ClientId: "test-client-id",
		BridgeConfig: ophosttypes.BridgeConfig{
			AttestorSet: []ophosttypes.Attestor{},
		},
	}
	err := input.OPChildKeeper.BridgeInfo.Set(ctx, bridgeInfo)
	require.NoError(t, err)

	packetData := ophosttypes.AttestorSetUpdatePacketData{
		BridgeId:      2, // Mismatched bridge ID
		AttestorSet:   []ophosttypes.Attestor{},
		L1BlockHeight: 100,
	}

	ack, err := input.OPChildKeeper.HandleAttestorSetUpdatePacket(ctx, packetData)
	require.NoError(t, err)
	require.False(t, ack.Success)
	require.Contains(t, ack.Error, "bridge ID mismatch")
}

func Test_HandleAttestorSetUpdatePacket_RemoveOldAttestors(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	bridgeInfo := opchildtypes.BridgeInfo{
		BridgeId:   1,
		BridgeAddr: addrsStr[0],
		L1ChainId:  "test-chain-1",
		L1ClientId: "test-client-id",
		BridgeConfig: ophosttypes.BridgeConfig{
			AttestorSet: []ophosttypes.Attestor{},
		},
	}
	err := input.OPChildKeeper.BridgeInfo.Set(ctx, bridgeInfo)
	require.NoError(t, err)

	valPubKeys := testutilsims.CreateTestPubKeys(4)
	attestor1 := createAttestor(t, valAddrsStr[1], valPubKeys[0], "attestor1")
	attestor2 := createAttestor(t, valAddrsStr[2], valPubKeys[1], "attestor2")

	packetData1 := ophosttypes.AttestorSetUpdatePacketData{
		BridgeId:      1,
		AttestorSet:   []ophosttypes.Attestor{attestor1, attestor2},
		L1BlockHeight: 100,
	}

	_, err = input.OPChildKeeper.HandleAttestorSetUpdatePacket(ctx, packetData1)
	require.NoError(t, err)

	vals, err := input.OPChildKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	require.Len(t, vals, 2)

	// update with new set (remove attestor2, add attestor3)
	attestor3 := createAttestor(t, valAddrsStr[3], valPubKeys[2], "attestor3")

	packetData2 := ophosttypes.AttestorSetUpdatePacketData{
		BridgeId:      1,
		AttestorSet:   []ophosttypes.Attestor{attestor1, attestor3},
		L1BlockHeight: 200,
	}

	_, err = input.OPChildKeeper.HandleAttestorSetUpdatePacket(ctx, packetData2)
	require.NoError(t, err)

	vals, err = input.OPChildKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	require.Len(t, vals, 3) // all 3 validators still exist

	// check attestor states
	foundAttestor1 := false
	foundAttestor2WithZeroPower := false
	foundAttestor3 := false
	for _, val := range vals {
		if val.Moniker == "attestor1" {
			foundAttestor1 = true
			require.Equal(t, int64(opchildtypes.AttestorConsPower), val.ConsPower, "attestor1 should have active power")
		}
		if val.Moniker == "attestor2" {
			foundAttestor2WithZeroPower = true
			require.Equal(t, int64(0), val.ConsPower, "attestor2 should have 0 power (removed)")
		}
		if val.Moniker == "attestor3" {
			foundAttestor3 = true
			require.Equal(t, int64(opchildtypes.AttestorConsPower), val.ConsPower, "attestor3 should have active power")
		}
	}
	require.True(t, foundAttestor1, "attestor1 should still exist")
	require.True(t, foundAttestor2WithZeroPower, "attestor2 should exist with 0 power")
	require.True(t, foundAttestor3, "attestor3 should be added")
}

func Test_HandleAttestorSetUpdatePacket_SkipExistingAttestors(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	bridgeInfo := opchildtypes.BridgeInfo{
		BridgeId:   1,
		BridgeAddr: addrsStr[0],
		L1ChainId:  "test-chain-1",
		L1ClientId: "test-client-id",
		BridgeConfig: ophosttypes.BridgeConfig{
			AttestorSet: []ophosttypes.Attestor{},
		},
	}
	err := input.OPChildKeeper.BridgeInfo.Set(ctx, bridgeInfo)
	require.NoError(t, err)

	valPubKeys := testutilsims.CreateTestPubKeys(2)
	attestor1 := createAttestor(t, valAddrsStr[1], valPubKeys[0], "attestor1")
	attestor2 := createAttestor(t, valAddrsStr[2], valPubKeys[1], "attestor2")

	packetData := ophosttypes.AttestorSetUpdatePacketData{
		BridgeId:      1,
		AttestorSet:   []ophosttypes.Attestor{attestor1, attestor2},
		L1BlockHeight: 100,
	}

	_, err = input.OPChildKeeper.HandleAttestorSetUpdatePacket(ctx, packetData)
	require.NoError(t, err)

	_, err = input.OPChildKeeper.HandleAttestorSetUpdatePacket(ctx, packetData)
	require.NoError(t, err)

	// still only 2 attestors
	vals, err := input.OPChildKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	require.Len(t, vals, 2)
}

func Test_HandleAttestorSetUpdatePacket_InvalidAttestor(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	bridgeInfo := opchildtypes.BridgeInfo{
		BridgeId:   1,
		BridgeAddr: addrsStr[0],
		L1ChainId:  "test-chain-1",
		L1ClientId: "test-client-id",
		BridgeConfig: ophosttypes.BridgeConfig{
			AttestorSet: []ophosttypes.Attestor{},
		},
	}
	err := input.OPChildKeeper.BridgeInfo.Set(ctx, bridgeInfo)
	require.NoError(t, err)

	packetData := ophosttypes.AttestorSetUpdatePacketData{
		BridgeId: 1,
		AttestorSet: []ophosttypes.Attestor{
			{
				OperatorAddress: "invalid-address",
				ConsensusPubkey: nil,
				Moniker:         "invalid-attestor",
			},
		},
		L1BlockHeight: 100,
	}

	ack, err := input.OPChildKeeper.HandleAttestorSetUpdatePacket(ctx, packetData)
	require.NoError(t, err)
	require.False(t, ack.Success)
	require.Contains(t, ack.Error, "failed to update attestors")
}

func Test_OnRecvAttestorSetUpdatePacket_Success(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	bridgeInfo := opchildtypes.BridgeInfo{
		BridgeId:   1,
		BridgeAddr: addrsStr[0],
		L1ChainId:  "test-chain-1",
		L1ClientId: "test-client-id",
		BridgeConfig: ophosttypes.BridgeConfig{
			AttestorSet: []ophosttypes.Attestor{},
		},
	}
	err := input.OPChildKeeper.BridgeInfo.Set(ctx, bridgeInfo)
	require.NoError(t, err)

	valPubKeys := testutilsims.CreateTestPubKeys(2)
	attestor1 := createAttestor(t, valAddrsStr[1], valPubKeys[0], "attestor1")
	attestor2 := createAttestor(t, valAddrsStr[2], valPubKeys[1], "attestor2")

	packetData := ophosttypes.NewAttestorSetUpdatePacketData(
		1,
		[]ophosttypes.Attestor{attestor1, attestor2},
		100,
	)

	ackBytes, err := input.OPChildKeeper.OnRecvAttestorSetUpdatePacket(ctx, packetData.GetBytes())
	require.NoError(t, err)
	require.NotNil(t, ackBytes)

	vals, err := input.OPChildKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	require.Len(t, vals, 2)
}

func Test_OnRecvAttestorSetUpdatePacket_InvalidPacketData(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	_, err := input.OPChildKeeper.OnRecvAttestorSetUpdatePacket(ctx, []byte("invalid-data"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to unmarshal attestor set update packet")
}
