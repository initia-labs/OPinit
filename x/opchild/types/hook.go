package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

type BridgeHook = func(ctx context.Context, sender sdk.AccAddress, msgBytes []byte) error

const senderPrefix = "opinit-hook-intermediary"

// DeriveIntermediateSender compute intermediate sender address
// Bech32(Hash(Hash("opinit-hook-intermediary") + sender))
func DeriveIntermediateSender(originalSender string) sdk.AccAddress {
	senderAddr := sdk.AccAddress(address.Hash(senderPrefix, []byte(originalSender)))
	return senderAddr
}
