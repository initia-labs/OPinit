package types

import (
	"sort"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// NewHistoricalInfo will create a historical information struct from header and valset
// it will first sort valset before inclusion into historical info
func NewHistoricalInfo(header tmproto.Header, valSet Validators) HistoricalInfo {
	// Must sort in the same way that tendermint does
	sort.SliceStable(valSet, func(i, j int) bool {
		return ValidatorsByVotingPower(valSet).Less(i, j)
	})

	return HistoricalInfo{
		Header: header,
		Valset: valSet,
	}
}

// MustUnmarshalHistoricalInfo wll unmarshal historical info and panic on error
func MustUnmarshalHistoricalInfo(cdc codec.BinaryCodec, value []byte) HistoricalInfo {
	hi, err := UnmarshalHistoricalInfo(cdc, value)
	if err != nil {
		panic(err)
	}

	return hi
}

// UnmarshalHistoricalInfo will unmarshal historical info and return any error
func UnmarshalHistoricalInfo(cdc codec.BinaryCodec, value []byte) (hi HistoricalInfo, err error) {
	err = cdc.Unmarshal(value, &hi)
	return hi, err
}

// ValidateBasic will ensure HistoricalInfo is not nil and sorted
func ValidateBasic(hi HistoricalInfo) error {
	if len(hi.Valset) == 0 {
		return errors.Wrap(ErrInvalidHistoricalInfo, "validator set is empty")
	}

	if !sort.IsSorted(Validators(hi.Valset)) {
		return errors.Wrap(ErrInvalidHistoricalInfo, "validator set is not sorted by address")
	}

	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (hi HistoricalInfo) UnpackInterfaces(c codectypes.AnyUnpacker) error {
	for i := range hi.Valset {
		if err := hi.Valset[i].UnpackInterfaces(c); err != nil {
			return err
		}
	}
	return nil
}
