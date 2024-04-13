package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	slinkypreblock "github.com/skip-mev/slinky/abci/preblock/oracle"
	slinkyproposals "github.com/skip-mev/slinky/abci/proposals"

	cometabci "github.com/cometbft/cometbft/abci/types"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/initia-labs/OPinit/x/opchild/types"
)

func (k *Keeper) SetOracle(
	slinkyKeeper types.OracleKeeper,
	slinkyProposalHandler *slinkyproposals.ProposalHandler,
	slinkyPreblockHandler *slinkypreblock.PreBlockHandler,
) {
	k.slinkyKeeper = slinkyKeeper
	k.slinkyProposalHandler = slinkyProposalHandler
	k.slinkyPreblockHandler = slinkyPreblockHandler
}

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

func (k Keeper) UpdateHostValidatorSet(ctx context.Context, chainId string, height int64, validatorSet *cmtproto.ValidatorSet) error {
	if k.HostValidatorStore == nil {
		return errors.New("not set host validator set")
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	if chainId == "" || params.HostChainId != chainId {
		return errors.New("only save host chain validators")
	}

	return k.HostValidatorStore.UpdateValidators(ctx, height, validatorSet)
}
