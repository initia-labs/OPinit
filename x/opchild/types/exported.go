package types

import (
	context "context"

	tmprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"

	"cosmossdk.io/math"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DelegationI delegation bond for a delegated proof of stake system
type DelegationI interface {
	GetDelegatorAddr() string  // delegator sdk.AccAddress for the bond
	GetValidatorAddr() string  // validator operator address
	GetShares() math.LegacyDec // amount of validator's shares held in this delegation
}

// ValidatorI expected validator functions
type ValidatorI interface {
	GetOperator() string                               // operator address to receive/return validators coins
	ConsPubKey() (cryptotypes.PubKey, error)           // validation consensus pubkey (cryptotypes.PubKey)
	TmConsPublicKey() (tmprotocrypto.PublicKey, error) // validation consensus pubkey (Tendermint)
	GetConsAddr() (sdk.ConsAddress, error)             // validation consensus address
	GetConsensusPower() int64                          // validation power in tendermint
	GetMoniker() string                                // return validator moniker
}

type AnteKeeper interface {
	MinGasPrices(ctx context.Context) (sdk.DecCoins, error)
}
