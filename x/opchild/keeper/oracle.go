package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cometabci "github.com/cometbft/cometbft/abci/types"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/initia-labs/OPinit/x/opchild/types"
)

func (k Keeper) UpdateOracle(ctx context.Context, extCommitBz []byte) error {
	if k.slinkyKeeper == nil || k.slinkyPreblockHandler == nil || k.slinkyProposalHandler == nil {
		return types.ErrInactiveOracle
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	consensusParams := sdkCtx.ConsensusParams()
	consensusParams.Abci = &cmtproto.ABCIParams{VoteExtensionsEnableHeight: 1}
	veEnabledCtx := sdkCtx.WithConsensusParams(consensusParams)

	proposalReq := &cometabci.RequestProcessProposal{
		Txs: [][]byte{extCommitBz},
	}

	_, err := k.slinkyProposalHandler.ProcessProposalHandler()(veEnabledCtx, proposalReq)
	if err != nil {
		return err
	}

	req := &cometabci.RequestFinalizeBlock{
		Txs: [][]byte{extCommitBz},
	}

	_, err = k.slinkyPreblockHandler.PreBlocker()(veEnabledCtx, req)
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) UpdateHostValidatorSet(ctx context.Context, validatorSet *cmtproto.ValidatorSet) error {

	return nil
}
