package keeper_test

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/initia-labs/OPinit/x/opchild/types"
	"github.com/stretchr/testify/require"
)

func Test_RegisterExecutorChangePlan(t *testing.T) {
	// Setup
	_, input := createTestInput(t, false)

	// Arguments
	l1ProposalID, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	require.NoError(t, err)
	height, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	require.NoError(t, err)
	nextValAddr := valAddrsStr[0]
	nextExecutorAddr := []string{addrsStr[0], addrsStr[1]}
	consensusPubKey := "l7aqGv+Zjbm0rallfqfqz+3iN31iOmgJCafWV5pGs6o="
	moniker := "moniker"
	info := "info"

	err = input.OPChildKeeper.RegisterExecutorChangePlan(
		l1ProposalID.Uint64(), height.Uint64(), nextValAddr,
		moniker,
		fmt.Sprintf(`{"@type":"/cosmos.crypto.ed25519.PubKey","key":"%s"}`, consensusPubKey),
		info,
		nextExecutorAddr,
	)
	require.NoError(t, err)
	require.Len(t, input.OPChildKeeper.ExecutorChangePlans, 1)

	consensusPubKeyBytes, err := base64.StdEncoding.DecodeString(consensusPubKey)
	require.NoError(t, err)

	expectedValidator, err := types.NewValidator(valAddrs[0], &ed25519.PubKey{
		Key: consensusPubKeyBytes,
	}, moniker)
	require.NoError(t, err)
	require.Equal(t, types.ExecutorChangePlan{
		ProposalID:    l1ProposalID.Uint64(),
		Height:        height.Uint64(),
		NextExecutors: []string{addrsStr[0], addrsStr[1]},
		NextValidator: expectedValidator,
		Info:          info,
	}, input.OPChildKeeper.ExecutorChangePlans[height.Uint64()])
}
