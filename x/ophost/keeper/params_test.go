package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/initia-labs/OPinit/x/ophost/testutil"
)

func Test_Params(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	params := input.OPHostKeeper.GetParams(ctx)
	params.RegistrationFee = sdk.NewCoins()

	err := input.OPHostKeeper.SetParams(ctx, params)
	require.NoError(t, err)
	require.True(t, input.OPHostKeeper.RegistrationFee(ctx).IsZero())
}
