package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/math"
	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/skip-mev/slinky/abci/ve"
	"github.com/skip-mev/slinky/pkg/math/voteweighted"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

var _ ve.ValidatorStore = (*HostValidatorStore)(nil)
var _ voteweighted.ValidatorStore = (*HostValidatorStore)(nil)

type HostValidatorStore struct {
	lastHeight            collections.Item[int64]
	validators            collections.Map[[]byte, stakingtypes.Validator]
	consensusAddressCodec address.Codec
}

func NewHostValidatorStore(
	lastHeight collections.Item[int64],
	validators collections.Map[[]byte, stakingtypes.Validator],
	consensusAddressCodec address.Codec,
) *HostValidatorStore {
	return &HostValidatorStore{
		lastHeight:            lastHeight,
		validators:            validators,
		consensusAddressCodec: consensusAddressCodec,
	}
}

func (hv HostValidatorStore) GetPubKeyByConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (cmtprotocrypto.PublicKey, error) {
	validator, err := hv.validators.Get(ctx, consAddr)
	if err != nil {
		return cmtprotocrypto.PublicKey{}, err
	}
	return validator.CmtConsPublicKey()
}

func (hv HostValidatorStore) ValidatorByConsAddr(ctx context.Context, addr sdk.ConsAddress) (stakingtypes.ValidatorI, error) {
	return hv.validators.Get(ctx, addr)
}

func (hv HostValidatorStore) TotalBondedTokens(ctx context.Context) (math.Int, error) {
	validators, err := hv.GetAllValidators(ctx)
	if err != nil {
		return math.Int{}, nil
	}
	totalBondedTokens := math.ZeroInt()
	for _, val := range validators {
		totalBondedTokens = totalBondedTokens.Add(val.BondedTokens())
	}
	return totalBondedTokens, nil
}

func (hv *HostValidatorStore) UpdateValidators(ctx context.Context, height int64, validatorSet *cmtproto.ValidatorSet) error {
	found, err := hv.lastHeight.Has(ctx)
	if err != nil {
		return err
	}

	lastHeight := int64(0)
	if found {
		lastHeight, err = hv.lastHeight.Get(ctx)
		if err != nil {
			return err
		}
	}

	if lastHeight >= height {
		return errors.New("invalid height")
	}

	err = hv.DeleteAllValidators(ctx)
	if err != nil {
		return err
	}

	for _, val := range validatorSet.GetValidators() {
		sdkPubKey, err := cryptocodec.FromCmtProtoPublicKey(val.PubKey)
		if err != nil {
			return err
		}

		validator, err := stakingtypes.NewValidator("", sdkPubKey, stakingtypes.Description{})
		if err != nil {
			return err
		}

		validator.Status = stakingtypes.Bonded
		validator.Tokens = math.NewInt(val.VotingPower)

		err = hv.SetValidator(ctx, validator)
		if err != nil {
			return err
		}
	}
	return hv.lastHeight.Set(ctx, height)
}

func (hv HostValidatorStore) SetValidator(ctx context.Context, validator stakingtypes.Validator) error {
	consAddr, err := validator.GetConsAddr()
	if err != nil {
		return err
	}
	return hv.validators.Set(ctx, consAddr, validator)
}

// get a single validator by consensus address
func (hv HostValidatorStore) GetValidator(ctx context.Context, consAddr sdk.ConsAddress) (validator stakingtypes.Validator, found bool) {
	validator, err := hv.validators.Get(ctx, consAddr)
	if errors.Is(err, collections.ErrNotFound) {
		return validator, false
	}

	return validator, true
}

func (hv HostValidatorStore) GetAllValidators(ctx context.Context) (validators []stakingtypes.Validator, err error) {
	err = hv.validators.Walk(ctx, nil, func(key []byte, validator stakingtypes.Validator) (stop bool, err error) {
		validators = append(validators, validator)
		return false, nil
	})

	return validators, err
}

func (hv HostValidatorStore) DeleteAllValidators(ctx context.Context) error {
	return hv.validators.Walk(ctx, nil, func(key []byte, _ stakingtypes.Validator) (stop bool, err error) {
		if err = hv.validators.Remove(ctx, key); err != nil {
			return true, err
		}
		return false, nil
	})
}

func (hv HostValidatorStore) GetLastHeight(ctx context.Context) (int64, error) {
	return hv.lastHeight.Get(ctx)
}
