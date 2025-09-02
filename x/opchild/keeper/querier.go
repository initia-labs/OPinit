package keeper

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/types/query"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	"github.com/initia-labs/OPinit/x/opchild/types"
)

type Querier struct {
	*Keeper
}

var _ types.QueryServer = &Querier{}

// NewQuerier return new Querier instance
func NewQuerier(k *Keeper) types.QueryServer {
	return &Querier{k}
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

func (q Querier) BridgeInfo(ctx context.Context, req *types.QueryBridgeInfoRequest) (*types.QueryBridgeInfoResponse, error) {
	bridgeInfo, err := q.Keeper.BridgeInfo.Get(ctx)
	if err != nil {
		return nil, status.Error(codes.NotFound, "bridge info not found")
	}

	return &types.QueryBridgeInfoResponse{BridgeInfo: bridgeInfo}, nil
}

func (q Querier) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	params, err := q.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryParamsResponse{Params: params}, nil
}

func (q Querier) NextL1Sequence(ctx context.Context, req *types.QueryNextL1SequenceRequest) (*types.QueryNextL1SequenceResponse, error) {
	nextL1Sequence, err := q.GetNextL1Sequence(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryNextL1SequenceResponse{NextL1Sequence: nextL1Sequence}, nil
}

func (q Querier) NextL2Sequence(ctx context.Context, req *types.QueryNextL2SequenceRequest) (*types.QueryNextL2SequenceResponse, error) {
	nextL2Sequence, err := q.GetNextL2Sequence(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryNextL2SequenceResponse{NextL2Sequence: nextL2Sequence}, nil
}

func (q Querier) BaseDenom(ctx context.Context, req *types.QueryBaseDenomRequest) (*types.QueryBaseDenomResponse, error) {
	baseDenom, err := q.GetBaseDenom(ctx, req.Denom)
	if err != nil {
		return nil, err
	}

	return &types.QueryBaseDenomResponse{BaseDenom: baseDenom}, nil
}

// MigrationInfo implements the Query/MigrationInfo RPC method
func (q Querier) MigrationInfo(ctx context.Context, req *types.QueryMigrationInfoRequest) (*types.QueryMigrationInfoResponse, error) {
	migrationInfo, err := q.GetMigrationInfo(ctx, req.Denom)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return nil, status.Error(codes.NotFound, err.Error())
	} else if err != nil {
		return nil, err
	}

	baseDenom, err := q.GetBaseDenom(ctx, req.Denom)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return nil, status.Error(codes.NotFound, err.Error())
	} else if err != nil {
		return nil, err
	}

	ibcDenom := transfertypes.ParseDenomTrace(fmt.Sprintf("%s/%s/%s", migrationInfo.IbcPortId, migrationInfo.IbcChannelId, baseDenom)).IBCDenom()
	return &types.QueryMigrationInfoResponse{MigrationInfo: migrationInfo, IbcDenom: ibcDenom}, nil
}
