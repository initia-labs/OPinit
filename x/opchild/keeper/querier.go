package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/initia-labs/OPinit/x/opchild/types"

	tmcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
)

type Querier struct {
	Keeper
}

var _ types.QueryServer = &Querier{}

// NewQuerier return new Querier instance
func NewQuerier(k Keeper) Querier {
	return Querier{k}
}

func (q Querier) Validator(ctx context.Context, req *types.QueryValidatorRequest) (*types.QueryValidatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.ValidatorAddr == "" {
		return nil, status.Error(codes.InvalidArgument, "validator address cannot be empty")
	}

	valAddr, err := q.validatorAddressCodec.StringToBytes(req.ValidatorAddr)
	if err != nil {
		return nil, err
	}

	validator, found := q.GetValidator(ctx, valAddr)
	if !found {
		return nil, status.Errorf(codes.NotFound, "validator %s not found", req.ValidatorAddr)
	}

	return &types.QueryValidatorResponse{Validator: validator}, nil
}

func (q Querier) Validators(ctx context.Context, req *types.QueryValidatorsRequest) (*types.QueryValidatorsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	validators, pageRes, err := query.CollectionPaginate(ctx, q.Keeper.Validators, req.Pagination, func(_ []byte, validator types.Validator) (types.Validator, error) {
		return validator, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryValidatorsResponse{Validators: validators, Pagination: pageRes}, nil
}

func (q Querier) ExecutorPubKey(ctx context.Context, req *types.QueryExecutorRequest) (*tmcrypto.PublicKey, error) {
	params, err := q.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	valAddr, err := q.addressCodec.StringToBytes(params.BridgeExecutor)
	if err != nil {
		return nil, err
	}

	validator, found := q.GetValidator(ctx, sdk.ValAddress(valAddr))
	if !found {
		return nil, status.Errorf(codes.NotFound, "executor validator not found", params.BridgeExecutor)
	}

	tmPubKey, err := validator.TmConsPublicKey()
	if err != nil {
		return nil, err
	}

	return &tmPubKey, nil
}

func (q Querier) Params(context context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(context)
	params, err := q.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryParamsResponse{Params: params}, nil
}
