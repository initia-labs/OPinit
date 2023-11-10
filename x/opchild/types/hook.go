package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type BridgeHook = func(ctx sdk.Context, sender sdk.AccAddress, msgBytes []byte) error
