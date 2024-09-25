package l2connect

import (
	"fmt"
	"math/big"
	"time"

	"cosmossdk.io/math"
	cometabci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	connectaggregator "github.com/skip-mev/connect/v2/abci/strategies/aggregator"
	connectcodec "github.com/skip-mev/connect/v2/abci/strategies/codec"
	connectabci "github.com/skip-mev/connect/v2/abci/types"
	connecttypes "github.com/skip-mev/connect/v2/pkg/types"
	oracletypes "github.com/skip-mev/connect/v2/x/oracle/types"

	"github.com/initia-labs/OPinit/x/opchild/types"
)

func GetOracleVotes(
	veCodec connectcodec.VoteExtensionCodec,
	extendedCommitInfo cometabci.ExtendedCommitInfo,
) ([]connectaggregator.Vote, error) {
	votes := make([]connectaggregator.Vote, len(extendedCommitInfo.Votes))
	for i, voteInfo := range extendedCommitInfo.Votes {
		voteExtension, err := veCodec.Decode(voteInfo.VoteExtension)
		if err != nil {
			return nil, connectabci.CodecError{
				Err: fmt.Errorf("error decoding vote-extension: %w", err),
			}
		}

		votes[i] = connectaggregator.Vote{
			ConsAddress:         voteInfo.Validator.Address,
			OracleVoteExtension: voteExtension,
		}
	}

	return votes, nil
}

func WritePrices(ctx sdk.Context, ok types.OracleKeeper, updatedTime time.Time, prices map[connecttypes.CurrencyPair]*big.Int) error {
	currencyPairs := ok.GetAllCurrencyPairs(ctx)
	for _, cp := range currencyPairs {
		price, found := prices[cp]
		if !found || price == nil {
			continue
		}

		qp, err := ok.GetPriceForCurrencyPair(ctx, cp)
		if err == nil && !updatedTime.After(qp.BlockTimestamp) {
			return types.ErrInvalidOracleTimestamp
		}

		// Convert the price to a quote price and write it to state.
		quotePrice := oracletypes.QuotePrice{
			Price:          math.NewIntFromBigInt(price),
			BlockTimestamp: updatedTime,
			BlockHeight:    uint64(ctx.BlockHeight()), //nolint:gosec
		}

		if err := ok.SetPriceForCurrencyPair(ctx, cp, quotePrice); err != nil {
			return err
		}
	}

	return nil
}
