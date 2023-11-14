package ante

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	rolluptypes "github.com/initia-labs/OPinit/x/opchild/types"
)

// MempoolFeeChecker will check if the transaction's fee is at least as large
// as the local validator's minimum gasFee (defined in validator config).
// If fee is too low, decorator returns error and tx is rejected from mempool.
// Note this only applies when ctx.CheckTx = true
// If fee is high enough or not CheckTx, then call next AnteHandler
// CONTRACT: Tx must implement FeeTx to use MempoolFeeChecker
type MempoolFeeChecker struct {
	keeper rolluptypes.AnteKeeper
}

// NewGasPricesDecorator create MempoolFeeDecorator instance
func NewMempoolFeeChecker(
	keeper rolluptypes.AnteKeeper,
) MempoolFeeChecker {
	return MempoolFeeChecker{
		keeper,
	}
}

func (mfd MempoolFeeChecker) CheckTxFeeWithMinGasPrices(ctx sdk.Context, tx sdk.Tx) (sdk.Coins, int64, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, 0, errors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()

	if ctx.IsCheckTx() {

		minGasPrices := ctx.MinGasPrices()
		if mfd.keeper != nil {
			minGasPrices = CombinedMinGasPrices(minGasPrices, mfd.keeper.MinGasPrices(ctx))
		}

		if !minGasPrices.IsZero() {
			requiredFees := computeRequiredFees(gas, minGasPrices)

			if !feeCoins.IsAnyGTE(requiredFees) {
				return nil, 0, errors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeCoins, requiredFees)
			}
		}
	}

	// TODO - if we want to use ethereum like priority system,
	// then we need to compute all dex prices of all fee coins
	return feeCoins, 1 /* FIFO */, nil
}
