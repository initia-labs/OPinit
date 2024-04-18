package types

import (
	"cosmossdk.io/core/address"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gopkg.in/yaml.v3"
)

var (
	DefaultMinGasPrices = sdk.NewDecCoins(sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyNewDecWithPrec(15, 2))) // 0.15
	DefaultHostChainId  = "mahalo-2"
)

// DefaultParams returns default move parameters
func DefaultParams() Params {
	return NewParams(
		"",
		"",
		DefaultMaxValidators,
		DefaultHistoricalEntries,
		DefaultMinGasPrices,
		DefaultHostChainId,
		[]string{},
	)
}

// NewParams creates a new Params instance
func NewParams(admin, bridgeExecutor string, maxValidators, historicalEntries uint32, minGasPrice sdk.DecCoins, hostChainId string, feeWhitelist []string) Params {
	return Params{
		Admin:             admin,
		BridgeExecutor:    bridgeExecutor,
		MaxValidators:     maxValidators,
		HistoricalEntries: historicalEntries,
		MinGasPrices:      minGasPrice,
		HostChainId:       hostChainId,
		FeeWhitelist:      feeWhitelist,
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
	if _, err := ac.StringToBytes(p.BridgeExecutor); err != nil {
		return err
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
