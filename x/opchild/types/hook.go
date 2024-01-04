package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type BridgeHook = func(ctx context.Context, sender sdk.AccAddress, msgBytes []byte) error
