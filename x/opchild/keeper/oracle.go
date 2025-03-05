package keeper

import (
	"context"
	"errors"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"

	connectaggregator "github.com/skip-mev/connect/v2/abci/strategies/aggregator"
	connectcodec "github.com/skip-mev/connect/v2/abci/strategies/codec"
	"github.com/skip-mev/connect/v2/abci/strategies/currencypair"
	"github.com/skip-mev/connect/v2/pkg/math/voteweighted"
	connecttypes "github.com/skip-mev/connect/v2/pkg/types"

	"github.com/initia-labs/OPinit/v1/x/opchild/l2connect"
	"github.com/initia-labs/OPinit/v1/x/opchild/types"
)

type L2OracleHandler struct {
	*Keeper

	oracleKeeper        types.OracleKeeper
	extendedCommitCodec connectcodec.ExtendedCommitCodec
	veCodec             connectcodec.VoteExtensionCodec
	voteAggregator      connectaggregator.VoteAggregator
}

func NewL2OracleHandler(
	k *Keeper,
	oracleKeeper types.OracleKeeper,
	logger log.Logger,
) *L2OracleHandler {
	return &L2OracleHandler{
		Keeper:       k,
		oracleKeeper: oracleKeeper,
		extendedCommitCodec: connectcodec.NewCompressionExtendedCommitCodec(
			connectcodec.NewDefaultExtendedCommitCodec(),
			connectcodec.NewZStdCompressor(),
		),
		veCodec: connectcodec.NewCompressionVoteExtensionCodec(
			connectcodec.NewDefaultVoteExtensionCodec(),
			connectcodec.NewZLibCompressor(),
		),
		voteAggregator: connectaggregator.NewDefaultVoteAggregator(
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

	h := int64(height) //nolint:gosec
	if hostStoreLastHeight > h {
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

	extendedVotes, err := l2connect.ValidateVoteExtensions(sdkCtx, k.HostValidatorStore, h-1, hostChainID, extendedCommitInfo)
	if err != nil {
		return err
	}

	votes, err := l2connect.GetOracleVotes(k.veCodec, extendedVotes)
	if err != nil {
		return err
	}

	prices, err := k.voteAggregator.AggregateOracleVotes(sdkCtx, votes)
	if err != nil {
		return err
	}

	tsCp, err := connecttypes.CurrencyPairFromString(l2connect.ReservedCPTimestamp)
	if err != nil {
		return err
	}

	// if there is no timestamp price, skip the price update
	if _, ok := prices[tsCp]; !ok {
		return types.ErrOracleTimestampNotExists
	}

	updatedTime := time.Unix(0, prices[tsCp].Int64())
	err = l2connect.WritePrices(sdkCtx, k.oracleKeeper, updatedTime, prices)
	if err != nil {
		return err
	}

	return nil
}
