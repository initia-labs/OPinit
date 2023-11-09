package types

import (
	"gopkg.in/yaml.v3"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	DefaultRegistrationFee = sdk.Coins{}
)

func DefaultParams() Params {
	return Params{
		RegistrationFee: DefaultRegistrationFee,
	}
}

func NewParams(registrationFee ...sdk.Coin) Params {
	return Params{
		RegistrationFee: registrationFee,
	}
}

func (p Params) String() string {
	out, err := yaml.Marshal(p)
	if err != nil {
		panic(err)
	}
	return string(out)
}

func (p Params) Validate() error {
	if err := p.RegistrationFee.Validate(); err != nil {
		return err
	}

	return nil
}
