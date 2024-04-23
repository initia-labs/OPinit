package l2slinky

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"cosmossdk.io/math"
	cometabci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	slinkyaggregator "github.com/skip-mev/slinky/abci/strategies/aggregator"
	slinkycodec "github.com/skip-mev/slinky/abci/strategies/codec"
	slinkyabci "github.com/skip-mev/slinky/abci/types"
	slinkytypes "github.com/skip-mev/slinky/pkg/types"
	oracletypes "github.com/skip-mev/slinky/x/oracle/types"

	"github.com/initia-labs/OPinit/x/opchild/types"
)

func GetOracleVotes(
	veCodec slinkycodec.VoteExtensionCodec,
	extendedCommitInfo cometabci.ExtendedCommitInfo,
) ([]slinkyaggregator.Vote, error) {
	votes := make([]slinkyaggregator.Vote, len(extendedCommitInfo.Votes))
	for i, voteInfo := range extendedCommitInfo.Votes {
		voteExtension, err := veCodec.Decode(voteInfo.VoteExtension)
		if err != nil {
			return nil, slinkyabci.CodecError{
				Err: fmt.Errorf("error decoding vote-extension: %w", err),
			}
		}

		votes[i] = slinkyaggregator.Vote{
			ConsAddress:         voteInfo.Validator.Address,
			OracleVoteExtension: voteExtension,
		}
	}

	return votes, nil
}

func WritePrices(ctx sdk.Context, ok types.OracleKeeper, updatedTime time.Time, prices map[slinkytypes.CurrencyPair]*big.Int) error {
	currencyPairs := ok.GetAllCurrencyPairs(ctx)
	for _, cp := range currencyPairs {
		price, found := prices[cp]
		if !found || price == nil {
			continue
		}

		qp, err := ok.GetPriceForCurrencyPair(ctx, cp)
		if err == nil && !qp.BlockTimestamp.After(updatedTime) {
			return errors.New("try to update the past price")
		}

		// Convert the price to a quote price and write it to state.
		quotePrice := oracletypes.QuotePrice{
			Price:          math.NewIntFromBigInt(price),
			BlockTimestamp: updatedTime,
			BlockHeight:    uint64(ctx.BlockHeight()),
		}

		if err := ok.SetPriceForCurrencyPair(ctx, cp, quotePrice); err != nil {
			return err
		}
	}

	return nil
}
