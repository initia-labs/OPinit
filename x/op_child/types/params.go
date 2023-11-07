package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gopkg.in/yaml.v3"
)

var (
	DefaultMinGasPrices = sdk.NewDecCoins(sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDecWithPrec(15, 2))) // 0.15
)

// DefaultParams returns default move parameters
func DefaultParams() Params {
	return NewParams(
		DefaultMaxValidators,
		DefaultHistoricalEntries,
		"",
		DefaultMinGasPrices,
	)
}

// NewParams creates a new Params instance
func NewParams(maxValidators, historicalEntries uint32, bridgeExecutor string, minGasPrice sdk.DecCoins) Params {
	return Params{
		BridgeExecutor:    bridgeExecutor,
		MaxValidators:     maxValidators,
		HistoricalEntries: historicalEntries,
		MinGasPrices:      minGasPrice,
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
func (p Params) Validate() error {
	if _, err := sdk.AccAddressFromBech32(p.BridgeExecutor); err != nil {
		return err
	}
	if err := p.MinGasPrices.Validate(); err != nil {
		return err
	}

	return nil
}

// unmarshal the current rollup params value from store key or panic
func MustUnmarshalParams(cdc *codec.LegacyAmino, value []byte) Params {
	params, err := UnmarshalParams(cdc, value)
	if err != nil {
		panic(err)
	}

	return params
}

// unmarshal the current rollup params value from store key
func UnmarshalParams(cdc *codec.LegacyAmino, value []byte) (params Params, err error) {
	err = cdc.Unmarshal(value, &params)
	if err != nil {
		return
	}

	return
}
