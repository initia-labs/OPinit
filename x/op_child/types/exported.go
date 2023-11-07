package types

import (
	tmprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DelegationI delegation bond for a delegated proof of stake system
type DelegationI interface {
	GetDelegatorAddr() sdk.AccAddress // delegator sdk.AccAddress for the bond
	GetValidatorAddr() sdk.ValAddress // validator operator address
	GetShares() sdk.Dec               // amount of validator's shares held in this delegation
}

// ValidatorI expected validator functions
type ValidatorI interface {
	GetOperator() sdk.ValAddress                       // operator address to receive/return validators coins
	ConsPubKey() (cryptotypes.PubKey, error)           // validation consensus pubkey (cryptotypes.PubKey)
	TmConsPublicKey() (tmprotocrypto.PublicKey, error) // validation consensus pubkey (Tendermint)
	GetConsAddr() (sdk.ConsAddress, error)             // validation consensus address
	GetConsensusPower() int64                          // validation power in tendermint
	GetMoniker() string                                // return validator moniker
}

type AnteKeeper interface {
	MinGasPrices(ctx sdk.Context) sdk.DecCoins
}
