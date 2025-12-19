package types

import (
	"strings"

	"cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	connecttypes "github.com/skip-mev/connect/v2/pkg/types"
	marketmaptypes "github.com/skip-mev/connect/v2/x/marketmap/types"
)

func (o OraclePriceData) Validate() error {
	if strings.TrimSpace(o.CurrencyPair) == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("currency pair cannot be empty")
	}

	if _, err := connecttypes.CurrencyPairFromString(o.CurrencyPair); err != nil {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid currency pair format: %s", o.CurrencyPair)
	}

	if strings.TrimSpace(o.Price) == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("price cannot be empty")
	}

	_, ok := math.NewIntFromString(o.Price)
	if !ok {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid price format: %s", o.Price)
	}

	if o.Decimals > marketmaptypes.DefaultMaxDecimals {
		return sdkerrors.ErrInvalidRequest.Wrapf("decimals too large: %d (max %d)", o.Decimals, marketmaptypes.DefaultMaxDecimals)
	}

	// nonce can be zero for the very first update

	return nil
}

func (o OracleData) Validate() error {
	if o.BridgeId == 0 {
		return ErrInvalidBridgeInfo.Wrap("bridge id cannot be zero")
	}

	if len(o.OraclePriceHash) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("oracle price hash cannot be empty")
	}

	if len(o.OraclePriceHash) != 32 {
		return sdkerrors.ErrInvalidRequest.Wrapf("oracle price hash must be 32 bytes, got %d", len(o.OraclePriceHash))
	}

	if len(o.Prices) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("prices cannot be empty")
	}

	// validate each price data
	for i, priceData := range o.Prices {
		if err := priceData.Validate(); err != nil {
			return sdkerrors.ErrInvalidRequest.Wrapf("invalid price data at index %d: %v", i, err)
		}
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

	return nil
}
