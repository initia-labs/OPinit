package types

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// FastBridgeVerifierI expected fast bridge verifier functions
type FastBridgeVerifierI interface {
	GetPubKey() (cryptotypes.PubKey, error)
}
