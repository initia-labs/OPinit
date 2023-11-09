package keeper

import (
	"context"
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/initia-labs/OPinit/x/op_host/types"
)

type Querier struct {
	Keeper
}

var _ types.QueryServer = &Querier{}

// NewQuerier return new Querier instance
func NewQuerier(k Keeper) Querier {
	return Querier{k}
}

func (q Querier) Bridge(context context.Context, req *types.QueryBridgeRequest) (*types.QueryBridgeResponse, error) {
	ctx := sdk.UnwrapSDKContext(context)

	bridgeId := req.BridgeId
	config, err := q.GetBridgeConfig(ctx, bridgeId)
	if err != nil {
		return nil, err
	}

	return &types.QueryBridgeResponse{
		BridgeId:     bridgeId,
		BridgeAddr:   types.BridgeAddress(bridgeId).String(),
		BridgeConfig: config,
	}, nil
}

func (q Querier) Bridges(context context.Context, req *types.QueryBridgesRequest) (*types.QueryBridgesResponse, error) {
	ctx := sdk.UnwrapSDKContext(context)

	var bridges []types.QueryBridgeResponse

	store := ctx.KVStore(q.storeKey)
	bridgeStore := prefix.NewStore(store, types.BridgeConfigKey)

	pageRes, err := query.FilteredPaginate(bridgeStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		bridgeId := binary.BigEndian.Uint64(key)
		var config types.BridgeConfig
		if err := q.cdc.Unmarshal(value, &config); err != nil {
			return false, err
		}

		if accumulate {
			bridges = append(bridges, types.QueryBridgeResponse{
				BridgeId:     bridgeId,
				BridgeAddr:   types.BridgeAddress(bridgeId).String(),
				BridgeConfig: config,
			})
		}

		return true, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryBridgesResponse{
		Bridges:    bridges,
		Pagination: pageRes,
	}, nil
}

func (q Querier) TokenPairByL1Denom(context context.Context, req *types.QueryTokenPairByL1DenomRequest) (*types.QueryTokenPairByL1DenomResponse, error) {
	l2Denom := types.L2Denom(req.BridgeId, req.L1Denom)
	return &types.QueryTokenPairByL1DenomResponse{
		TokenPair: types.TokenPair{
			L1Denom: req.L1Denom,
			L2Denom: l2Denom,
		},
	}, nil
}

func (q Querier) TokenPairByL2Denom(context context.Context, req *types.QueryTokenPairByL2DenomRequest) (*types.QueryTokenPairByL2DenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(context)
	l1Denom, err := q.GetTokenPair(ctx, req.BridgeId, req.L2Denom)
	if err != nil {
		return nil, err
	}

	return &types.QueryTokenPairByL2DenomResponse{
		TokenPair: types.TokenPair{
			L1Denom: l1Denom,
			L2Denom: req.L2Denom,
		},
	}, nil
}

func (q Querier) TokenPairs(context context.Context, req *types.QueryTokenPairsRequest) (*types.QueryTokenPairsResponse, error) {
	ctx := sdk.UnwrapSDKContext(context)

	var pairs []types.TokenPair
	bridgeId := req.BridgeId

	store := ctx.KVStore(q.storeKey)
	pairStore := prefix.NewStore(store, types.GetTokenPairBridgePrefixKey(bridgeId))

	pageRes, err := query.FilteredPaginate(pairStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		l2Denom := string(key)
		l1Denom := string(value)

		if accumulate {
			pairs = append(pairs, types.TokenPair{
				L1Denom: l1Denom,
				L2Denom: l2Denom,
			})
		}

		return true, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryTokenPairsResponse{
		TokenPairs: pairs,
		Pagination: pageRes,
	}, nil
}

func (q Querier) OutputProposal(context context.Context, req *types.QueryOutputProposalRequest) (*types.QueryOutputProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(context)
	output, err := q.GetOutputProposal(ctx, req.BridgeId, req.OutputIndex)
	if err != nil {
		return nil, err
	}

	return &types.QueryOutputProposalResponse{
		BridgeId:       req.BridgeId,
		OutputIndex:    req.OutputIndex,
		OutputProposal: output,
	}, nil
}

func (q Querier) OutputProposals(context context.Context, req *types.QueryOutputProposalsRequest) (*types.QueryOutputProposalsResponse, error) {
	ctx := sdk.UnwrapSDKContext(context)

	var outputs []types.QueryOutputProposalResponse
	bridgeId := req.BridgeId

	store := ctx.KVStore(q.storeKey)
	outputStore := prefix.NewStore(store, types.GetOutputProposalBridgePrefixKey(bridgeId))

	pageRes, err := query.FilteredPaginate(outputStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		outputIndex := binary.BigEndian.Uint64(key)
		var output types.Output
		if err := q.cdc.Unmarshal(value, &output); err != nil {
			return false, err
		}

		if accumulate {
			outputs = append(outputs, types.QueryOutputProposalResponse{
				BridgeId:       bridgeId,
				OutputIndex:    outputIndex,
				OutputProposal: output,
			})
		}

		return true, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryOutputProposalsResponse{
		OutputProposals: outputs,
		Pagination:      pageRes,
	}, nil
}

func (q Querier) Params(context context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(context)

	return &types.QueryParamsResponse{Params: q.GetParams(ctx)}, nil
}
