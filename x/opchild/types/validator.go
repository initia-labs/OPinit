package types

import (
	"bytes"
	"sort"
	"strings"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	"sigs.k8s.io/yaml"

	"cosmossdk.io/core/address"
	"cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// TODO: Why can't we just have one string description which can be JSON by convention
	MaxMonikerLength         = 70
	MaxIdentityLength        = 3000
	MaxWebsiteLength         = 140
	MaxSecurityContactLength = 140
	MaxDetailsLength         = 280
)

var _ ValidatorI = Validator{}

// NewValidator constructs a new Validator
//
//nolint:interfacer
func NewValidator(operator sdk.ValAddress, pubKey cryptotypes.PubKey, moniker string) (Validator, error) {
	pkAny, err := codectypes.NewAnyWithValue(pubKey)
	if err != nil {
		return Validator{}, err
	}

	return Validator{
		Moniker:         moniker,
		OperatorAddress: operator.String(),
		ConsensusPubkey: pkAny,
		ConsPower:       1,
	}, nil
}

// String implements the Stringer interface for a Validator object.
func (v Validator) String() string {
	bz, err := codec.ProtoMarshalJSON(&v, nil)
	if err != nil {
		panic(err)
	}

	out, err := yaml.JSONToYAML(bz)
	if err != nil {
		panic(err)
	}

	return string(out)
}

// GetMoniker return validator moniker.
func (v Validator) GetMoniker() string {
	return v.Moniker
}

// Validators is a collection of Validator
type Validators struct {
	Validators     []Validator
	ValidatorCodec address.Codec
}

func (v Validators) String() (out string) {
	for _, val := range v.Validators {
		out += val.String() + "\n"
	}

	return strings.TrimSpace(out)
}

// ToSDKValidators -  convenience function convert []Validator to []sdk.ValidatorI
func (v Validators) ToSDKValidators() (validators []ValidatorI) {
	for _, val := range v.Validators {
		validators = append(validators, val)
	}

	return validators
}

// Sort Validators sorts validator array in ascending operator address order
func (v Validators) Sort() {
	sort.Sort(v)
}

// Implements sort interface
func (v Validators) Len() int {
	return len(v.Validators)
}

// Implements sort interface
func (v Validators) Less(i, j int) bool {
	vi, err := v.ValidatorCodec.StringToBytes(v.Validators[i].GetOperator())
	if err != nil {
		panic(err)
	}
	vj, err := v.ValidatorCodec.StringToBytes(v.Validators[j].GetOperator())
	if err != nil {
		panic(err)
	}

	return bytes.Compare(vi, vj) == -1
}

// Implements sort interface
func (v Validators) Swap(i, j int) {
	v.Validators[i], v.Validators[j] = v.Validators[j], v.Validators[i]
}

// ValidatorsByVotingPower implements sort.Interface for []Validator based on
// the VotingPower and Address fields.
// The validators are sorted first by their voting power (descending). Secondary index - Address (ascending).
// Copied from tendermint/types/validator_set.go
type ValidatorsByVotingPower []Validator

func (valz ValidatorsByVotingPower) Len() int { return len(valz) }

func (valz ValidatorsByVotingPower) Less(i, j int) bool {
	if valz[i].ConsensusPower() == valz[j].ConsensusPower() {
		addrI, errI := valz[i].GetConsAddr()
		addrJ, errJ := valz[j].GetConsAddr()
		// If either returns error, then return false
		if errI != nil || errJ != nil {
			return false
		}
		return bytes.Compare(addrI, addrJ) == -1
	}
	return valz[i].ConsensusPower() > valz[j].ConsensusPower()
}

func (valz ValidatorsByVotingPower) Swap(i, j int) {
	valz[i], valz[j] = valz[j], valz[i]
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (v Validators) UnpackInterfaces(c codectypes.AnyUnpacker) error {
	for i := range v.Validators {
		if err := v.Validators[i].UnpackInterfaces(c); err != nil {
			return err
		}
	}
	return nil
}

// return the delegation
func MustMarshalValidator(cdc codec.BinaryCodec, validator *Validator) []byte {
	return cdc.MustMarshal(validator)
}

// unmarshal a delegation from a store value
func MustUnmarshalValidator(cdc codec.BinaryCodec, value []byte) Validator {
	validator, err := UnmarshalValidator(cdc, value)
	if err != nil {
		panic(err)
	}

	return validator
}

// unmarshal a delegation from a store value
func UnmarshalValidator(cdc codec.BinaryCodec, value []byte) (v Validator, err error) {
	err = cdc.Unmarshal(value, &v)
	return v, err
}

func (v Validator) ABCIValidatorUpdate() abci.ValidatorUpdate {
	tmProtoPk, err := v.TmConsPublicKey()
	if err != nil {
		panic(err)
	}

	return abci.ValidatorUpdate{
		PubKey: tmProtoPk,
		Power:  v.ConsensusPower(),
	}
}

// ABCIValidatorUpdateZero returns an abci.ValidatorUpdate from a rollup validator type
// with zero power used for validator updates.
func (v Validator) ABCIValidatorUpdateZero() abci.ValidatorUpdate {
	tmProtoPk, err := v.TmConsPublicKey()
	if err != nil {
		panic(err)
	}

	return abci.ValidatorUpdate{
		PubKey: tmProtoPk,
		Power:  0,
	}
}

// ConsensusPower gets the consensus-engine power.
// Return constant value 1 for all validators
func (v Validator) ConsensusPower() int64 {
	return v.ConsPower
}

// Equal checks if the receiver equals the parameter
func (v *Validator) Equal(v2 *Validator) bool {
	return v.OperatorAddress == v2.OperatorAddress &&
		v.ConsensusPubkey.Equal(v2.ConsensusPubkey)
}

func (v Validator) GetOperator() string {
	return v.OperatorAddress
}

// ConsPubKey returns the validator PubKey as a cryptotypes.PubKey.
func (v Validator) ConsPubKey() (cryptotypes.PubKey, error) {
	pk, ok := v.ConsensusPubkey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidType, "expecting cryptotypes.PubKey, got %T", pk)
	}

	return pk, nil
}

// TmConsPublicKey casts Validator.ConsensusPubkey to cmtprotocrypto.PubKey.
func (v Validator) TmConsPublicKey() (cmtprotocrypto.PublicKey, error) {
	pk, err := v.ConsPubKey()
	if err != nil {
		return cmtprotocrypto.PublicKey{}, err
	}

	tmPk, err := cryptocodec.ToCmtProtoPublicKey(pk)
	if err != nil {
		return cmtprotocrypto.PublicKey{}, err
	}

	return tmPk, nil
}

// GetConsAddr extracts Consensus key address
func (v Validator) GetConsAddr() (sdk.ConsAddress, error) {
	pk, ok := v.ConsensusPubkey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidType, "expecting cryptotypes.PubKey, got %T", pk)
	}

	return sdk.ConsAddress(pk.Address()), nil
}

func (v Validator) GetConsensusPower() int64 {
	return v.ConsensusPower()
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (v Validator) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pk cryptotypes.PubKey
	return unpacker.UnpackAny(v.ConsensusPubkey, &pk)
}
