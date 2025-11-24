package types

import (
	"strings"

	"cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (o OracleData) Validate() error {
	if o.BridgeId == 0 {
		return ErrInvalidBridgeInfo.Wrap("bridge id cannot be zero")
	}

	if strings.TrimSpace(o.CurrencyPair) == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("currency pair cannot be empty")
	}

	// format check: should contain '/'
	if !strings.Contains(o.CurrencyPair, "/") {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid currency pair format: %s", o.CurrencyPair)
	}

	if strings.TrimSpace(o.Price) == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("price cannot be empty")
	}

	_, ok := math.NewIntFromString(o.Price)
	if !ok {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid price format: %s", o.Price)
	}

	if o.Decimals > 18 {
		return sdkerrors.ErrInvalidRequest.Wrapf("decimals too large: %d (max 18)", o.Decimals)
	}

	if o.L1BlockHeight == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("l1 block height cannot be zero")
	}

	if o.L1BlockTime <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid l1 block time: %d", o.L1BlockTime)
	}

	if len(o.Proof) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("proof cannot be empty")
	}

	if o.ProofHeight.RevisionHeight == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("proof height revision height cannot be zero")
	}

	// nonce can be zero for the very first update
	
	return nil
}
