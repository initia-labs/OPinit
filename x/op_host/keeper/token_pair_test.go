package keeper_test

import (
	"testing"

	"github.com/initia-labs/OPinit/x/op_host/types"
	"github.com/stretchr/testify/require"
)

func Test_TokenPair(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	tokenPair := types.TokenPair{
		L1Denom: "l1_denom",
		L2Denom: "l2_denom",
	}
	input.OPHostKeeper.SetTokenPair(ctx, 1, tokenPair.L2Denom, tokenPair.L1Denom)

	l1Denom, err := input.OPHostKeeper.GetTokenPair(ctx, 1, tokenPair.L2Denom)
	require.Equal(t, tokenPair.L1Denom, l1Denom)
	require.NoError(t, err)
}
