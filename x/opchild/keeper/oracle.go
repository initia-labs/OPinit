package keeper

import (
	"context"
	"errors"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"

	slinkyaggregator "github.com/skip-mev/slinky/abci/strategies/aggregator"
	slinkycodec "github.com/skip-mev/slinky/abci/strategies/codec"
	"github.com/skip-mev/slinky/abci/strategies/currencypair"
	"github.com/skip-mev/slinky/pkg/math/voteweighted"
	slinkytypes "github.com/skip-mev/slinky/pkg/types"
	oraclekeeper "github.com/skip-mev/slinky/x/oracle/keeper"

	"github.com/initia-labs/OPinit/x/opchild/l2slinky"
	"github.com/initia-labs/OPinit/x/opchild/types"
)

type L2OracleHandler struct {
	*Keeper

	oracleKeeper        types.OracleKeeper
	extendedCommitCodec slinkycodec.ExtendedCommitCodec
	veCodec             slinkycodec.VoteExtensionCodec
	voteAggregator      slinkyaggregator.VoteAggregator
}

func NewL2OracleHandler(
	k *Keeper,
	oracleKeeper *oraclekeeper.Keeper,
	logger log.Logger,
) *L2OracleHandler {
	return &L2OracleHandler{
		Keeper:       k,
		oracleKeeper: oracleKeeper,
		extendedCommitCodec: slinkycodec.NewCompressionExtendedCommitCodec(
			slinkycodec.NewDefaultExtendedCommitCodec(),
			slinkycodec.NewZStdCompressor(),
		),
		veCodec: slinkycodec.NewCompressionVoteExtensionCodec(
			slinkycodec.NewDefaultVoteExtensionCodec(),
			slinkycodec.NewZLibCompressor(),
		),
		voteAggregator: slinkyaggregator.NewDefaultVoteAggregator(
			logger,
			voteweighted.MedianFromContext(
				logger,
				k.HostValidatorStore,
				voteweighted.DefaultPowerThreshold,
			),
			currencypair.NewHashCurrencyPairStrategy(oracleKeeper),
		),
	}
}

func (k L2OracleHandler) UpdateOracle(ctx context.Context, height uint64, extCommitBz []byte) error {
	hostStoreLastHeight, err := k.HostValidatorStore.GetLastHeight(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.ErrOracleValidatorsNotRegistered
		}
		return err
	}

	if hostStoreLastHeight > int64(height) {
		return types.ErrInvalidOracleHeight
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	hostChainID, err := k.L1ChainId(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.ErrBridgeInfoNotExists
		}
		return err
	}

	extendedCommitInfo, err := k.extendedCommitCodec.Decode(extCommitBz)
	if err != nil {
		return err
	}

	err = l2slinky.ValidateVoteExtensions(sdkCtx, k.HostValidatorStore, int64(height-1), hostChainID, extendedCommitInfo)
	if err != nil {
		return err
	}

	votes, err := l2slinky.GetOracleVotes(k.veCodec, extendedCommitInfo)
	if err != nil {
		return err
	}
	prices, err := k.voteAggregator.AggregateOracleVotes(sdkCtx, votes)
	if err != nil {
		return err
	}

	tsCp, err := slinkytypes.CurrencyPairFromString(l2slinky.ReservedCPTimestamp)
	if err != nil {
		return err
	}

	// if there is no timestamp price, skip the price update
	if _, ok := prices[tsCp]; !ok {
		return types.ErrOracleTimestampNotExists
	}

	updatedTime := time.Unix(0, prices[tsCp].Int64())
	err = l2slinky.WritePrices(sdkCtx, k.oracleKeeper, updatedTime, prices)
	if err != nil {
		return err
	}

	return nil
}
