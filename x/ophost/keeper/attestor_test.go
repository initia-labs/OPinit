package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/initia-labs/OPinit/x/ophost/testutil"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
)

func Test_SendAttestorSetUpdatePacket_Success(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	bridgeId := uint64(1)
	bridgeConfig := ophosttypes.BridgeConfig{
		Proposer:              testutil.AddrsStr[0],
		Challenger:            testutil.AddrsStr[1],
		SubmissionInterval:    100,
		FinalizationPeriod:    1000,
		SubmissionStartHeight: 1,
		Metadata:              []byte("test-metadata"),
		BatchInfo: ophosttypes.BatchInfo{
			Submitter: testutil.AddrsStr[2],
			ChainType: ophosttypes.BatchInfo_INITIA,
		},
		AttestorSet: []ophosttypes.Attestor{},
	}

	err := input.OPHostKeeper.SetBridgeConfig(ctx, bridgeId, bridgeConfig)
	require.NoError(t, err)

	err = input.OPHostKeeper.SendAttestorSetUpdatePacket(
		ctx,
		bridgeId,
		ophosttypes.PortID,
		"channel-0",
	)
	require.NoError(t, err)

	events := sdk.UnwrapSDKContext(ctx).EventManager().Events()
	found := false
	for _, event := range events {
		if event.Type == ophosttypes.EventTypeAttestorSetPacketSent {
			found = true
			break
		}
	}
	require.True(t, found, "expected attestor set packet sent event")
}

func Test_SendAttestorSetUpdatePacket_BridgeNotFound(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	err := input.OPHostKeeper.SendAttestorSetUpdatePacket(
		ctx,
		999, // non-existent bridge ID
		ophosttypes.PortID,
		"channel-0",
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get bridge config")
}

func Test_SendAttestorSetUpdatePacket_WithAttestors(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	privKey1 := ed25519.GenPrivKey()
	pubKey1 := privKey1.PubKey()

	pkAny1, err := codectypes.NewAnyWithValue(pubKey1)
	require.NoError(t, err)

	attestor1 := ophosttypes.Attestor{
		OperatorAddress: testutil.ValAddrsStr[3],
		ConsensusPubkey: pkAny1,
		Moniker:         "attestor1",
	}

	bridgeId := uint64(1)
	bridgeConfig := ophosttypes.BridgeConfig{
		Proposer:              testutil.AddrsStr[0],
		Challenger:            testutil.AddrsStr[1],
		SubmissionInterval:    100,
		FinalizationPeriod:    1000,
		SubmissionStartHeight: 1,
		Metadata:              []byte("test-metadata"),
		BatchInfo: ophosttypes.BatchInfo{
			Submitter: testutil.AddrsStr[2],
			ChainType: ophosttypes.BatchInfo_INITIA,
		},
		AttestorSet: []ophosttypes.Attestor{attestor1},
	}

	err = input.OPHostKeeper.SetBridgeConfig(ctx, bridgeId, bridgeConfig)
	require.NoError(t, err)

	err = input.OPHostKeeper.SendAttestorSetUpdatePacket(
		ctx,
		bridgeId,
		ophosttypes.PortID,
		"channel-0",
	)
	require.NoError(t, err)
}
