package types

import (
	"cosmossdk.io/core/address"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gopkg.in/yaml.v3"
)

var (
	DefaultMinGasPrices = sdk.NewDecCoins(sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyNewDecWithPrec(15, 2))) // 0.15
	DefaultHookMaxGas   = uint64(1_000_000)
)

// DefaultParams returns default move parameters
func DefaultParams() Params {
	return NewParams(
		"",
		[]string{""},
		DefaultMaxValidators,
		DefaultHistoricalEntries,
		DefaultMinGasPrices,
		[]string{},
		DefaultHookMaxGas,
	)
}

// NewParams creates a new Params instance
func NewParams(admin string, bridgeExecutors []string, maxValidators, historicalEntries uint32, minGasPrice sdk.DecCoins, feeWhitelist []string, hookMaxGas uint64) Params {
	return Params{
		Admin:             admin,
		BridgeExecutors:   bridgeExecutors,
		MaxValidators:     maxValidators,
		HistoricalEntries: historicalEntries,
		MinGasPrices:      minGasPrice,
		FeeWhitelist:      feeWhitelist,
		HookMaxGas:        hookMaxGas,
	}

}

func (p Params) String() string {
	out, err := yaml.Marshal(p)
	if err != nil {
		panic(err)
	}
	return string(out)
}

// Validate performs basic validation on move parameters
func (p Params) Validate(ac address.Codec) error {
	if _, err := ac.StringToBytes(p.Admin); err != nil {
		return err
	}
	for _, be := range p.BridgeExecutors {
		if _, err := ac.StringToBytes(be); err != nil {
			return err
		}
	}

	if err := p.MinGasPrices.Validate(); err != nil {
		return err
	}

	if p.MaxValidators == 0 {
		return ErrZeroMaxValidators
	}

	// Validate fee whitelist addresses
	for _, addr := range p.FeeWhitelist {
		if _, err := ac.StringToBytes(addr); err != nil {
			return err
		}
	}

	return nil
}
