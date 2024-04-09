package keeper

import (
	"context"

	"cosmossdk.io/math"
	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/skip-mev/slinky/abci/ve"
	"github.com/skip-mev/slinky/pkg/math/voteweighted"
)

var _ ve.ValidatorStore = (*HostValidatorStore)(nil)
var _ voteweighted.ValidatorStore = (*HostValidatorStore)(nil)

type HostValidatorStore struct {
}

func NewHostValidatorStore() *HostValidatorStore {
	return &HostValidatorStore{}
}

func (hv HostValidatorStore) GetPubKeyByConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (cmtprotocrypto.PublicKey, error) {

	return cmtprotocrypto.PublicKey{}, nil
}

func (hv HostValidatorStore) ValidatorByConsAddr(ctx context.Context, addr sdk.ConsAddress) (stakingtypes.ValidatorI, error) {
	return nil, nil
}

func (hv HostValidatorStore) TotalBondedTokens(ctx context.Context) (math.Int, error) {
	return math.Int{}, nil
}
