package lanes

import (
	"context"

	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/skip-mev/block-sdk/v2/block/base"
)

type FeeWhitelistKeeper interface {
	FeeWhitelist(ctx context.Context) ([]string, error)
}

// FreeLaneMatchHandler returns the default match handler for the free lane. The
// default implementation matches fee payers that are in the fee whitelist.
type FreeLaneMatchHandler struct {
	ac  address.Codec
	fwk FeeWhitelistKeeper
}

func NewFreeLaneMatchHandler(ac address.Codec, fwk FeeWhitelistKeeper) FreeLaneMatchHandler {
	return FreeLaneMatchHandler{
		ac:  ac,
		fwk: fwk,
	}
}

func (h FreeLaneMatchHandler) MatchHandler() base.MatchHandler {
	return func(ctx sdk.Context, tx sdk.Tx) bool {
		feeTx, ok := tx.(sdk.FeeTx)
		if !ok {
			return false
		}

		if whitelist, err := h.fwk.FeeWhitelist(ctx); err != nil {
			return false
		} else if payer, err := h.ac.BytesToString(feeTx.FeePayer()); err != nil {
			return false
		} else {
			for _, addr := range whitelist {
				if addr == payer {
					return true
				}
			}
		}

		return false
	}
}
