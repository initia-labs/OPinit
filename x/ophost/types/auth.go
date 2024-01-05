package types

import (
	"encoding/binary"
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var (
	_ sdk.AccountI             = (*BridgeAccount)(nil)
	_ authtypes.GenesisAccount = (*BridgeAccount)(nil)
)

// NewBridgeAccountWithAddress create new object account with the given address.
func NewBridgeAccountWithAddress(addr sdk.AccAddress) *BridgeAccount {
	return &BridgeAccount{
		authtypes.NewBaseAccountWithAddress(addr),
	}
}

// SetPubKey - Implements AccountI
func (ma BridgeAccount) SetPubKey(pubKey cryptotypes.PubKey) error {
	return fmt.Errorf("not supported for object accounts")
}

// BridgeAddress - generate bridge address with bridge id as seed
func BridgeAddress(bridgeId uint64) sdk.AccAddress {
	seed := make([]byte, 8)
	binary.BigEndian.PutUint64(seed, bridgeId)
	return address.Module(ModuleName, seed)
}
