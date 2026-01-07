package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	hosttypes "github.com/initia-labs/OPinit/x/ophost/types"
)

func TestUnmarshalProtoJSON_AttestorSetUpdatePacket(t *testing.T) {
	jsonBytes := []byte(`{
		"attestorSet": [
			{
				"consensusPubkey": {
					"@type": "/cosmos.crypto.ed25519.PubKey",
					"key": "VDspFXG0jq16Ec9De3mFMutuxVTJrp1YEfNczph6Dxk="
				},
				"moniker": "validator",
				"operatorAddress": "initvaloper1vz2ahm0rf2t5m5vaehd5lsjpzdu8pdkdg04qsc"
			}
		],
		"bridgeId": "1",
		"l1BlockHeight": "3384"
	}`)

	packet := &hosttypes.AttestorSetUpdatePacketData{}
	err := unmarshalProtoJSON(jsonBytes, packet)
	require.NoError(t, err)

	require.Equal(t, uint64(1), packet.BridgeId)
	require.Equal(t, uint64(3384), packet.L1BlockHeight)

	require.Len(t, packet.AttestorSet, 1)
	attestor := packet.AttestorSet[0]
	require.Equal(t, "validator", attestor.Moniker)
	require.Equal(t, "initvaloper1vz2ahm0rf2t5m5vaehd5lsjpzdu8pdkdg04qsc", attestor.OperatorAddress)

	require.NotNil(t, attestor.ConsensusPubkey)
	require.Equal(t, "/cosmos.crypto.ed25519.PubKey", attestor.ConsensusPubkey.TypeUrl)
}
