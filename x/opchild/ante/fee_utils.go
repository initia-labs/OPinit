package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CombinedMinGasPrices will combine the on-chain fee and min_gas_prices.
func CombinedMinGasPrices(minGasPrices sdk.DecCoins, configMinGasPrices sdk.DecCoins) sdk.DecCoins {
	for _, cmgp := range configMinGasPrices {
		if minGasPrices.AmountOf(cmgp.Denom).IsZero() {
			minGasPrices = minGasPrices.Add(cmgp)
		} else if minGasPrices.AmountOf(cmgp.Denom).LT(cmgp.Amount) {
			minGasPrices = minGasPrices.Add(cmgp.Sub(sdk.NewDecCoinFromDec(cmgp.Denom, minGasPrices.AmountOf(cmgp.Denom))))
		} // else, GTE, use the original minGasPrice {
	}
	return minGasPrices.Sort()
}

// computeRequiredFees returns required fees
func computeRequiredFees(gas uint64, minGasPrices sdk.DecCoins) sdk.Coins {
	// special case: if minGasPrices=[], requiredFees=[]
	requiredFees := make(sdk.Coins, len(minGasPrices))

	// if not all coins are zero, check fee with min_gas_price
	if !minGasPrices.IsZero() {
		// Determine the required fees by multiplying each required minimum gas
		// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
		for i, gp := range minGasPrices {
			fee := gp.Amount.MulInt64(int64(gas))
			requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
		}
	}

	return requiredFees.Sort()
}
