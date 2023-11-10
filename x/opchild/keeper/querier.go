package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/initia-labs/OPinit/x/opchild/types"
)

type Querier struct {
	Keeper
}

var _ types.QueryServer = &Querier{}

// NewQuerier return new Querier instance
func NewQuerier(k Keeper) Querier {
	return Querier{k}
}

func (q Querier) Validator(context context.Context, req *types.QueryValidatorRequest) (*types.QueryValidatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.ValidatorAddr == "" {
		return nil, status.Error(codes.InvalidArgument, "validator address cannot be empty")
	}

	valAddr, err := sdk.ValAddressFromBech32(req.ValidatorAddr)
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(context)
	validator, found := q.GetValidator(ctx, valAddr)
	if !found {
		return nil, status.Errorf(codes.NotFound, "validator %s not found", req.ValidatorAddr)
	}

	return &types.QueryValidatorResponse{Validator: validator}, nil
}

func (q Querier) Validators(context context.Context, req *types.QueryValidatorsRequest) (*types.QueryValidatorsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	var validators types.Validators
	ctx := sdk.UnwrapSDKContext(context)

	store := ctx.KVStore(q.storeKey)
	valStore := prefix.NewStore(store, types.ValidatorsKey)

	pageRes, err := query.FilteredPaginate(valStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		val, err := types.UnmarshalValidator(q.cdc, value)
		if err != nil {
			return false, err
		}

		if accumulate {
			validators = append(validators, val)
		}

		return true, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryValidatorsResponse{Validators: validators, Pagination: pageRes}, nil
}

func (q Querier) Params(context context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(context)

	return &types.QueryParamsResponse{Params: q.GetParams(ctx)}, nil
}
