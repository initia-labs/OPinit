package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/initia-labs/OPinit/x/ophost/types"
)

type Querier struct {
	Keeper
}

var _ types.QueryServer = &Querier{}

// NewQuerier return new Querier instance
func NewQuerier(k Keeper) Querier {
	return Querier{k}
}

func (q Querier) Bridge(ctx context.Context, req *types.QueryBridgeRequest) (*types.QueryBridgeResponse, error) {
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

func (q Querier) Bridges(ctx context.Context, req *types.QueryBridgesRequest) (*types.QueryBridgesResponse, error) {
	bridges, pageRes, err := query.CollectionPaginate(ctx, q.Keeper.BridgeConfigs, req.Pagination, func(bridgeId uint64, bridgeConfig types.BridgeConfig) (types.QueryBridgeResponse, error) {
		return types.QueryBridgeResponse{
			BridgeId:     bridgeId,
			BridgeAddr:   types.BridgeAddress(bridgeId).String(),
			BridgeConfig: bridgeConfig,
		}, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryBridgesResponse{
		Bridges:    bridges,
		Pagination: pageRes,
	}, nil
}

func (q Querier) TokenPairByL1Denom(_ context.Context, req *types.QueryTokenPairByL1DenomRequest) (*types.QueryTokenPairByL1DenomResponse, error) {
	l2Denom := types.L2Denom(req.BridgeId, req.L1Denom)
	return &types.QueryTokenPairByL1DenomResponse{
		TokenPair: types.TokenPair{
			L1Denom: req.L1Denom,
			L2Denom: l2Denom,
		},
	}, nil
}

func (q Querier) TokenPairByL2Denom(ctx context.Context, req *types.QueryTokenPairByL2DenomRequest) (*types.QueryTokenPairByL2DenomResponse, error) {
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

func (q Querier) TokenPairs(ctx context.Context, req *types.QueryTokenPairsRequest) (*types.QueryTokenPairsResponse, error) {
	pairs, pageRes, err := query.CollectionPaginate(ctx, q.Keeper.TokenPairs, req.Pagination, func(key collections.Pair[uint64, string], l1Denom string) (types.TokenPair, error) {
		return types.TokenPair{
			L1Denom: l1Denom,
			L2Denom: key.K2(),
		}, nil
	}, query.WithCollectionPaginationPairPrefix[uint64, string](req.BridgeId))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryTokenPairsResponse{
		TokenPairs: pairs,
		Pagination: pageRes,
	}, nil
}

func (q Querier) LastFinalizedOutput(ctx context.Context, req *types.QueryLastFinalizedOutputRequest) (*types.QueryLastFinalizedOutputResponse, error) {
	lastOutputIndex, lastOutput, err := q.GetLastFinalizedOutput(ctx, req.BridgeId)
	if err != nil {
		return nil, err
	}

	return &types.QueryLastFinalizedOutputResponse{
		OutputIndex:    lastOutputIndex,
		OutputProposal: lastOutput,
	}, nil
}

func (q Querier) OutputProposal(ctx context.Context, req *types.QueryOutputProposalRequest) (*types.QueryOutputProposalResponse, error) {
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

func (q Querier) OutputProposals(ctx context.Context, req *types.QueryOutputProposalsRequest) (*types.QueryOutputProposalsResponse, error) {
	outputs, pageRes, err := query.CollectionPaginate(ctx, q.Keeper.OutputProposals, req.Pagination, func(key collections.Pair[uint64, uint64], outputProposal types.Output) (types.QueryOutputProposalResponse, error) {
		return types.QueryOutputProposalResponse{
			BridgeId:       req.BridgeId,
			OutputIndex:    key.K2(),
			OutputProposal: outputProposal,
		}, nil
	}, query.WithCollectionPaginationPairPrefix[uint64, uint64](req.BridgeId))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryOutputProposalsResponse{
		OutputProposals: outputs,
		Pagination:      pageRes,
	}, nil
}

func (q Querier) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	return &types.QueryParamsResponse{Params: q.GetParams(ctx)}, nil
}

func (q Querier) Claimed(ctx context.Context, req *types.QueryClaimedRequest) (*types.QueryClaimedResponse, error) {
	if len(req.WithdrawalHash) != 32 {
		return nil, status.Error(codes.InvalidArgument, "invalid withdrawal hash")
	}

	claimed, err := q.HasProvenWithdrawal(ctx, req.BridgeId, [32]byte(req.WithdrawalHash))
	if err != nil {
		return nil, err
	}

	return &types.QueryClaimedResponse{
		Claimed: claimed,
	}, nil
}

func (q Querier) NextL1Sequence(ctx context.Context, req *types.QueryNextL1SequenceRequest) (*types.QueryNextL1SequenceResponse, error) {
	sequence, err := q.GetNextL1Sequence(ctx, req.BridgeId)
	if err != nil {
		return nil, err
	}

	return &types.QueryNextL1SequenceResponse{
		NextL1Sequence: sequence,
	}, nil
}

func (q Querier) BatchInfos(ctx context.Context, req *types.QueryBatchInfosRequest) (*types.QueryBatchInfosResponse, error) {
	batchInfos, pageRes, err := query.CollectionPaginate(ctx, q.Keeper.BatchInfos, req.Pagination, func(key collections.Pair[uint64, uint64], batchInfoWithOutput types.BatchInfoWithOutput) (types.BatchInfoWithOutput, error) {
		return batchInfoWithOutput, nil
	}, query.WithCollectionPaginationPairPrefix[uint64, uint64](req.BridgeId))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryBatchInfosResponse{
		BatchInfos: batchInfos,
		Pagination: pageRes,
	}, nil
}
