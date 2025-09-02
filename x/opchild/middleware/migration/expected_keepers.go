package migration

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type OPChildKeeper interface {
	HasIBCToL2DenomMap(ctx context.Context, ibcDenom string) (bool, error)
	HandleMigratedTokenDeposit(ctx context.Context, sender sdk.AccAddress, ibcCoin sdk.Coin, memo string) (sdk.Coin, error)
}

type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}
