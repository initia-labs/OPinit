package keeper_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	protoio "github.com/cosmos/gogoproto/io"
	"github.com/cosmos/gogoproto/proto"

	cometabci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	slinkycodec "github.com/skip-mev/slinky/abci/strategies/codec"
	"github.com/skip-mev/slinky/abci/strategies/currencypair"
	"github.com/skip-mev/slinky/abci/ve/types"
	slinkytypes "github.com/skip-mev/slinky/pkg/types"
	oraclekeeper "github.com/skip-mev/slinky/x/oracle/keeper"
	oracletypes "github.com/skip-mev/slinky/x/oracle/types"
)

func createCmtValidatorSet(t *testing.T, numVals int) ([]cryptotypes.PrivKey, []cryptotypes.PubKey, *cmtproto.ValidatorSet) {
	privKeys := make([]cryptotypes.PrivKey, numVals)
	pubKeys := make([]cryptotypes.PubKey, numVals)
	for i := 0; i < numVals; i++ {
		privKeys[i] = ed25519.GenPrivKey()
		pubKeys[i] = privKeys[i].PubKey()
	}
	cmtValSet := &cmtproto.ValidatorSet{
		Validators: make([]*cmtproto.Validator, numVals),
	}
	for i, valPubKey := range pubKeys {
		cmtPubKey, err := cryptocodec.ToCmtProtoPublicKey(valPubKey)
		require.NoError(t, err)

		cmtValSet.Validators[i] = &cmtproto.Validator{
			Address:     valPubKey.Address(),
			PubKey:      cmtPubKey,
			VotingPower: 1,
		}
	}

	return privKeys, pubKeys, cmtValSet
}

func getSlinky(oracleKeeper *oraclekeeper.Keeper) (currencypair.CurrencyPairStrategy, slinkycodec.ExtendedCommitCodec, slinkycodec.VoteExtensionCodec) {
	cpStrategy := currencypair.NewDefaultCurrencyPairStrategy(oracleKeeper)
	voteExtensionCodec := slinkycodec.NewCompressionVoteExtensionCodec(
		slinkycodec.NewDefaultVoteExtensionCodec(),
		slinkycodec.NewZLibCompressor(),
	)

	extendedCommitCodec := slinkycodec.NewCompressionExtendedCommitCodec(
		slinkycodec.NewDefaultExtendedCommitCodec(),
		slinkycodec.NewZStdCompressor(),
	)

	return cpStrategy, extendedCommitCodec, voteExtensionCodec
}

func Test_UpdateHostValidatorSet(t *testing.T) {
	defaultHostChainId := "test-host-1"

	testCases := []struct {
		name        string
		hostChainId string
		hostHeight  int64
		numVals     int
		expectError bool
	}{
		{
			name:        "empty chain id",
			hostChainId: "",
			hostHeight:  100,
			numVals:     10,
			expectError: true,
		},
		{
			name:        "different chain id",
			hostChainId: "test-host-12",
			hostHeight:  100,
			numVals:     10,
			expectError: true,
		},
		{
			name:        "zero height",
			hostChainId: defaultHostChainId,
			hostHeight:  0,
			numVals:     10,
			expectError: true,
		},
		{
			name:        "good host chain id, height, validators",
			hostChainId: defaultHostChainId,
			hostHeight:  100,
			numVals:     10,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, input := createDefaultTestInput(t)
			opchildKeeper := input.OPChildKeeper
			hostValidatorStore := opchildKeeper.HostValidatorStore

			params, err := opchildKeeper.GetParams(ctx)
			require.NoError(t, err)
			params.HostChainId = defaultHostChainId
			err = opchildKeeper.SetParams(ctx, params)
			require.NoError(t, err)

			_, valPubKeys, validatorSet := createCmtValidatorSet(t, tc.numVals)
			err = opchildKeeper.UpdateHostValidatorSet(ctx, tc.hostChainId, tc.hostHeight, validatorSet)
			if tc.expectError {
				// no error but no validator update
				require.NoError(t, err)

				vals, err := hostValidatorStore.GetAllValidators(ctx)
				require.NoError(t, err)
				require.Empty(t, vals)
				return
			} else {
				require.NoError(t, err)
			}

			for _, valPubKey := range valPubKeys {
				consAddr := sdk.ConsAddress(valPubKey.Address())
				_, err := hostValidatorStore.GetPubKeyByConsAddr(ctx, consAddr)
				require.NoError(t, err)

				_, err = hostValidatorStore.ValidatorByConsAddr(ctx, consAddr)
				require.NoError(t, err)
			}

		})
	}
}

func Test_UpdateOracle(t *testing.T) {
	defaultHostChainId := "test-host-1"

	testCases := []struct {
		name          string
		currencyPairs []string
		prices        []map[string]string
		result        map[string]string
		numVals       int
		expectError   bool
	}{
		{
			name:          "good currency pairs, updates",
			currencyPairs: []string{"BTC/USD", "ETH/USD", "ATOM/USD", "TIMESTAMP/NANOSECOND"},
			prices: []map[string]string{
				{"BTC/USD": "10000000", "ETH/USD": "200000", "ATOM/USD": "5000", "TIMESTAMP/NANOSECOND": "10000"},
				{"BTC/USD": "10000000", "ETH/USD": "200000", "ATOM/USD": "5000", "TIMESTAMP/NANOSECOND": "10000"},
				{"BTC/USD": "10000000", "ETH/USD": "200000", "ATOM/USD": "5000", "TIMESTAMP/NANOSECOND": "10000"},
				{"BTC/USD": "10000000", "ETH/USD": "200000", "ATOM/USD": "5000", "TIMESTAMP/NANOSECOND": "10000"},
				{"BTC/USD": "10000000", "ETH/USD": "200000", "ATOM/USD": "5000", "TIMESTAMP/NANOSECOND": "10000"},
			},
			result:      map[string]string{"BTC/USD": "10000000", "ETH/USD": "200000", "ATOM/USD": "5000", "TIMESTAMP/NANOSECOND": "10000"},
			numVals:     5,
			expectError: false,
		},
		{
			name:          "only BTC, ETH",
			currencyPairs: []string{"BTC/USD", "ETH/USD", "TIMESTAMP/NANOSECOND"},
			prices: []map[string]string{
				{"BTC/USD": "10000000", "ETH/USD": "200000", "ATOM/USD": "5000", "TIMESTAMP/NANOSECOND": "10000"},
				{"BTC/USD": "10000000", "ETH/USD": "200000", "ATOM/USD": "5000", "TIMESTAMP/NANOSECOND": "10000"},
				{"BTC/USD": "10000000", "ETH/USD": "210000", "ATOM/USD": "5000", "TIMESTAMP/NANOSECOND": "10000"},
				{"BTC/USD": "11000000", "ETH/USD": "210000", "ATOM/USD": "5000", "TIMESTAMP/NANOSECOND": "10000"},
				{"BTC/USD": "11000000", "ETH/USD": "210000", "ATOM/USD": "5000", "TIMESTAMP/NANOSECOND": "10000"},
			},
			result:      map[string]string{"BTC/USD": "10000000", "ETH/USD": "210000", "TIMESTAMP/NANOSECOND": "10000"},
			numVals:     5,
			expectError: false,
		},
		{
			name:          "reverse order ATOM, ETH, BTC",
			currencyPairs: []string{"ATOM/USD", "ETH/USD", "BTC/USD", "TIMESTAMP/NANOSECOND"},
			prices: []map[string]string{
				{"BTC/USD": "10000000", "ETH/USD": "200000", "ATOM/USD": "5000", "TIMESTAMP/NANOSECOND": "10000"},
				{"BTC/USD": "10000000", "ETH/USD": "200000", "ATOM/USD": "5000", "TIMESTAMP/NANOSECOND": "10000"},
				{"BTC/USD": "10000000", "ETH/USD": "210000", "ATOM/USD": "5000", "TIMESTAMP/NANOSECOND": "10000"},
				{"BTC/USD": "11000000", "ETH/USD": "210000", "ATOM/USD": "5000", "TIMESTAMP/NANOSECOND": "10000"},
				{"BTC/USD": "11000000", "ETH/USD": "210000", "ATOM/USD": "5000", "TIMESTAMP/NANOSECOND": "10000"},
			},
			result:      map[string]string{"BTC/USD": "10000000", "ETH/USD": "210000", "ATOM/USD": "5000", "TIMESTAMP/NANOSECOND": "10000"},
			numVals:     5,
			expectError: false,
		},
		{
			name:          "2 votes",
			currencyPairs: []string{"ATOM/USD", "ETH/USD", "BTC/USD", "TIMESTAMP/NANOSECOND"},
			prices: []map[string]string{
				{"BTC/USD": "10000000", "ETH/USD": "200000", "ATOM/USD": "5000", "TIMESTAMP/NANOSECOND": "10000"},
				{},
				{},
				{"BTC/USD": "11000000", "ETH/USD": "210000", "ATOM/USD": "51000", "TIMESTAMP/NANOSECOND": "10000"},
				{},
			},
			result:      map[string]string{"BTC/USD": "10000000", "ETH/USD": "200000", "ATOM/USD": "5000", "TIMESTAMP/NANOSECOND": "10000"},
			numVals:     5,
			expectError: true,
		},
		{
			name:          "4 votes",
			currencyPairs: []string{"ATOM/USD", "ETH/USD", "BTC/USD", "TIMESTAMP/NANOSECOND"},
			prices: []map[string]string{
				{"BTC/USD": "10000000", "ETH/USD": "200000", "ATOM/USD": "5000", "TIMESTAMP/NANOSECOND": "10000"},
				{},
				{"BTC/USD": "10000000", "ETH/USD": "200000", "ATOM/USD": "5000", "TIMESTAMP/NANOSECOND": "10000"},
				{"BTC/USD": "11000000", "ETH/USD": "220000", "ATOM/USD": "5100", "TIMESTAMP/NANOSECOND": "10000"},
				{"BTC/USD": "10000000", "ETH/USD": "220000", "ATOM/USD": "5000", "TIMESTAMP/NANOSECOND": "10000"},
			},
			result:      map[string]string{"BTC/USD": "10000000", "ETH/USD": "200000", "ATOM/USD": "5000", "TIMESTAMP/NANOSECOND": "10000"},
			numVals:     5,
			expectError: false,
		},
	}

	marshalDelimitedFn := func(msg proto.Message) ([]byte, error) {
		var buf bytes.Buffer
		if err := protoio.NewDelimitedWriter(&buf).WriteMsg(msg); err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, input := createDefaultTestInput(t)
			opchildKeeper := input.OPChildKeeper
			oracleKeeper := input.OracleKeeper

			params, err := opchildKeeper.GetParams(ctx)
			require.NoError(t, err)
			params.HostChainId = defaultHostChainId
			err = opchildKeeper.SetParams(ctx, params)
			require.NoError(t, err)

			oracleKeeper.InitGenesis(sdk.UnwrapSDKContext(ctx), oracletypes.GenesisState{
				CurrencyPairGenesis: make([]oracletypes.CurrencyPairGenesis, 0),
			})
			for _, currencyPair := range tc.currencyPairs {
				cp, err := slinkytypes.CurrencyPairFromString(currencyPair)
				require.NoError(t, err)
				err = oracleKeeper.CreateCurrencyPair(sdk.UnwrapSDKContext(ctx), cp)
				require.NoError(t, err)
			}

			cpStrategy, extendedCommitCodec, voteExtensionCodec := getSlinky(oracleKeeper)
			valPrivKeys, _, validatorSet := createCmtValidatorSet(t, tc.numVals)
			err = opchildKeeper.UpdateHostValidatorSet(ctx, defaultHostChainId, 1, validatorSet)
			require.NoError(t, err)

			eci := cometabci.ExtendedCommitInfo{
				Round: 2,
				Votes: make([]cometabci.ExtendedVoteInfo, tc.numVals),
			}

			for i, privKey := range valPrivKeys {
				convertedPrices := make(map[uint64][]byte)
				for currencyPairID, priceString := range tc.prices[i] {
					cp, err := slinkytypes.CurrencyPairFromString(currencyPairID)
					require.NoError(t, err)
					rawPrice, converted := new(big.Int).SetString(priceString, 10)
					require.True(t, converted)

					encodedPrice, err := cpStrategy.GetEncodedPrice(sdk.UnwrapSDKContext(ctx), cp, rawPrice)
					require.NoError(t, err)

					id := oracletypes.CurrencyPairToID(currencyPairID)
					convertedPrices[id] = encodedPrice
				}
				ove := types.OracleVoteExtension{
					Prices: convertedPrices,
				}

				exCommitBz, err := voteExtensionCodec.Encode(ove)
				require.NoError(t, err)

				cve := cmtproto.CanonicalVoteExtension{
					Extension: exCommitBz,
					Height:    10,
					Round:     2,
					ChainId:   defaultHostChainId,
				}

				extSignBytes, err := marshalDelimitedFn(&cve)
				require.NoError(t, err)

				signedBytes, err := privKey.Sign(extSignBytes)
				require.NoError(t, err)

				eci.Votes[i] = cometabci.ExtendedVoteInfo{
					Validator: cometabci.Validator{
						Address: validatorSet.Validators[i].Address,
						Power:   1,
					},
					VoteExtension:      exCommitBz,
					ExtensionSignature: signedBytes,
					BlockIdFlag:        cmtproto.BlockIDFlagCommit,
				}
			}

			extCommitBz, err := extendedCommitCodec.Encode(eci)
			require.NoError(t, err)

			err = opchildKeeper.ApplyOracleUpdate(ctx, 11, extCommitBz)
			if tc.expectError {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			for currencyPairID, priceString := range tc.result {
				cp, err := slinkytypes.CurrencyPairFromString(currencyPairID)
				require.NoError(t, err)

				price, err := oracleKeeper.GetPriceForCurrencyPair(sdk.UnwrapSDKContext(ctx), cp)
				require.NoError(t, err)
				require.Equal(t, price.Price.String(), priceString)
			}
		})
	}
}
