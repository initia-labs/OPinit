package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	slinkypreblock "github.com/skip-mev/slinky/abci/preblock/oracle"

	"github.com/skip-mev/slinky/abci/ve"

	cometabci "github.com/cometbft/cometbft/abci/types"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/initia-labs/OPinit/x/opchild/types"
	"github.com/skip-mev/slinky/abci/strategies/codec"
)

func (k *Keeper) SetOracle(
	slinkyKeeper types.OracleKeeper,
	extendedCommitCodec codec.ExtendedCommitCodec,
	slinkyPreblockHandler *slinkypreblock.PreBlockHandler,
) {
	k.slinkyKeeper = slinkyKeeper
	k.extendedCommitCodec = extendedCommitCodec
	k.slinkyPreblockHandler = slinkyPreblockHandler
}

func (k Keeper) UpdateOracle(ctx context.Context, height uint64, extCommitBz []byte) error {
	if k.slinkyKeeper == nil || k.slinkyPreblockHandler == nil || k.extendedCommitCodec == nil {
		return types.ErrInactiveOracle
	}

	hostStoreLastHeight, err := k.HostValidatorStore.GetLastHeight(ctx)
	if err != nil {
		return err
	}

	if hostStoreLastHeight > int64(height) {
		return errors.New("invalid height")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	consensusParams := sdkCtx.ConsensusParams()
	consensusParams.Abci = &cmtproto.ABCIParams{VoteExtensionsEnableHeight: 1}
	veEnabledCtx := sdkCtx.WithConsensusParams(consensusParams)

	hostChainID, err := k.HostChainId(ctx)
	if err != nil {
		return err
	}

	extendedCommitInfo, err := k.extendedCommitCodec.Decode(extCommitBz)
	if err != nil {
		return err
	}
	err = ve.ValidateVoteExtensionsFromL1(veEnabledCtx, k.HostValidatorStore, int64(height), hostChainID, extendedCommitInfo)
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

func (k Keeper) UpdateHostValidatorSet(ctx context.Context, chainID string, height int64, validatorSet *cmtproto.ValidatorSet) error {
	hostChainID, err := k.HostChainId(ctx)
	if err != nil {
		return err
	}
	if chainID == "" {
		return errors.New("empty chain id")
	}
	if hostChainID != chainID {
		return errors.New("only save host chain validators")
	}

	return k.HostValidatorStore.UpdateValidators(ctx, height, validatorSet)
}
