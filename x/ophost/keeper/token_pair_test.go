package keeper_test

import (
	"testing"

	"github.com/initia-labs/OPinit/x/ophost/types"
	"github.com/stretchr/testify/require"
)

func Test_TokenPair(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	tokenPair := types.TokenPair{
		L1Denom: "l1_denom",
		L2Denom: "l2_denom",
	}
	require.NoError(t, input.OPHostKeeper.SetTokenPair(ctx, 1, tokenPair.L2Denom, tokenPair.L1Denom))

	l1Denom, err := input.OPHostKeeper.GetTokenPair(ctx, 1, tokenPair.L2Denom)
	require.Equal(t, tokenPair.L1Denom, l1Denom)
	require.NoError(t, err)
}

func Test_IterateTokenPair(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	tokenPair1 := types.TokenPair{
		L1Denom: "l11_denom",
		L2Denom: "l12_denom",
	}
	tokenPair2 := types.TokenPair{
		L1Denom: "l21_denom",
		L2Denom: "l22_denom",
	}
	tokenPair3 := types.TokenPair{
		L1Denom: "l31_denom",
		L2Denom: "l32_denom",
	}
	require.NoError(t, input.OPHostKeeper.SetTokenPair(ctx, 1, tokenPair1.L2Denom, tokenPair1.L1Denom))
	require.NoError(t, input.OPHostKeeper.SetTokenPair(ctx, 1, tokenPair2.L2Denom, tokenPair2.L1Denom))
	require.NoError(t, input.OPHostKeeper.SetTokenPair(ctx, 2, tokenPair3.L2Denom, tokenPair3.L1Denom))

	require.NoError(t, input.OPHostKeeper.IterateTokenPair(ctx, 1, func(bridgeId uint64, tokenPair types.TokenPair) (stop bool, err error) {
		require.Equal(t, bridgeId, uint64(1))
		if tokenPair.L1Denom == tokenPair1.L1Denom {
			require.Equal(t, tokenPair1, tokenPair)
		} else {
			require.Equal(t, tokenPair2, tokenPair)
		}
		return false, nil
	}))
}
