package keeper_test

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"cosmossdk.io/store/dbadapter"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/initia-labs/OPinit/x/opchild/types"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	protoio "github.com/cosmos/gogoproto/io"
	"github.com/cosmos/gogoproto/proto"

	cometabci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v8/modules/core/23-commitment/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	tmclient "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"

	connectcodec "github.com/skip-mev/connect/v2/abci/strategies/codec"
	"github.com/skip-mev/connect/v2/abci/strategies/currencypair"
	vetypes "github.com/skip-mev/connect/v2/abci/ve/types"
	connecttypes "github.com/skip-mev/connect/v2/pkg/types"
	oraclekeeper "github.com/skip-mev/connect/v2/x/oracle/keeper"
	oracletypes "github.com/skip-mev/connect/v2/x/oracle/types"
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

func getConnect(oracleKeeper *oraclekeeper.Keeper) (currencypair.CurrencyPairStrategy, connectcodec.ExtendedCommitCodec, connectcodec.VoteExtensionCodec) {
	cpStrategy := currencypair.NewHashCurrencyPairStrategy(oracleKeeper)
	voteExtensionCodec := connectcodec.NewCompressionVoteExtensionCodec(
		connectcodec.NewDefaultVoteExtensionCodec(),
		connectcodec.NewZLibCompressor(),
	)

	extendedCommitCodec := connectcodec.NewCompressionExtendedCommitCodec(
		connectcodec.NewDefaultExtendedCommitCodec(),
		connectcodec.NewZStdCompressor(),
	)

	return cpStrategy, extendedCommitCodec, voteExtensionCodec
}

func Test_UpdateHostValidatorSet(t *testing.T) {
	defaultClientId := "test-client-id"

	testCases := []struct {
		name         string
		hostClientId string
		hostHeight   int64
		numVals      int
		expectError  bool
	}{
		{
			name:         "empty chain id",
			hostClientId: "",
			hostHeight:   100,
			numVals:      10,
			expectError:  true,
		},
		{
			name:         "different chain id",
			hostClientId: "test-host-12",
			hostHeight:   100,
			numVals:      10,
			expectError:  true,
		},
		{
			name:         "zero height",
			hostClientId: defaultClientId,
			hostHeight:   0,
			numVals:      10,
			expectError:  true,
		},
		{
			name:         "good host chain id, height, validators",
			hostClientId: defaultClientId,
			hostHeight:   100,
			numVals:      10,
			expectError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, input := createDefaultTestInput(t)
			opchildKeeper := input.OPChildKeeper
			hostValidatorStore := opchildKeeper.HostValidatorStore

			bridgeInfo := types.BridgeInfo{
				L1ClientId: defaultClientId,
			}
			err := opchildKeeper.BridgeInfo.Set(ctx, bridgeInfo)
			require.NoError(t, err)

			_, valPubKeys, validatorSet := createCmtValidatorSet(t, tc.numVals)
			err = opchildKeeper.UpdateHostValidatorSet(ctx, tc.hostClientId, tc.hostHeight, validatorSet)
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
	defaultClientId := "test-client-id"
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

			bridgeInfo := types.BridgeInfo{
				L1ChainId:  defaultHostChainId,
				L1ClientId: defaultClientId,
			}
			err := opchildKeeper.BridgeInfo.Set(ctx, bridgeInfo)
			require.NoError(t, err)

			oracleKeeper.InitGenesis(sdk.UnwrapSDKContext(ctx), oracletypes.GenesisState{
				CurrencyPairGenesis: make([]oracletypes.CurrencyPairGenesis, 0),
			})
			for _, currencyPair := range tc.currencyPairs {
				cp, err := connecttypes.CurrencyPairFromString(currencyPair)
				require.NoError(t, err)
				err = oracleKeeper.CreateCurrencyPair(sdk.UnwrapSDKContext(ctx), cp)
				require.NoError(t, err)
			}

			cpStrategy, extendedCommitCodec, voteExtensionCodec := getConnect(oracleKeeper)
			valPrivKeys, _, validatorSet := createCmtValidatorSet(t, tc.numVals)
			err = opchildKeeper.UpdateHostValidatorSet(ctx, defaultClientId, 1, validatorSet)
			require.NoError(t, err)

			eci := cometabci.ExtendedCommitInfo{
				Round: 2,
				Votes: make([]cometabci.ExtendedVoteInfo, tc.numVals),
			}

			for i, privKey := range valPrivKeys {
				convertedPrices := make(map[uint64][]byte)
				for currencyPairID, priceString := range tc.prices[i] {
					cp, err := connecttypes.CurrencyPairFromString(currencyPairID)
					require.NoError(t, err)
					rawPrice, converted := new(big.Int).SetString(priceString, 10)
					require.True(t, converted)

					sdkCtx := sdk.UnwrapSDKContext(ctx)
					encodedPrice, err := cpStrategy.GetEncodedPrice(sdkCtx, cp, rawPrice)
					require.NoError(t, err)

					id, err := currencypair.CurrencyPairToHashID(currencyPairID)
					require.NoError(t, err)

					convertedPrices[id] = encodedPrice
				}
				ove := vetypes.OracleVoteExtension{
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
				cp, err := connecttypes.CurrencyPairFromString(currencyPairID)
				require.NoError(t, err)

				price, err := oracleKeeper.GetPriceForCurrencyPair(sdk.UnwrapSDKContext(ctx), cp)
				require.NoError(t, err)
				require.Equal(t, price.Price.String(), priceString)
			}
		})
	}
}

func Test_HandleOracleDataPacket_BridgeIdMismatch(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	bridgeInfo := types.BridgeInfo{
		BridgeId:  1,
		L1ChainId: "test-chain-1",
		BridgeConfig: ophosttypes.BridgeConfig{
			OracleEnabled: true,
		},
	}
	err := input.OPChildKeeper.BridgeInfo.Set(ctx, bridgeInfo)
	require.NoError(t, err)

	oracleData := types.OracleData{
		BridgeId:       2, // mismatched bridge ID
		CurrencyPair:   "BTC/USD",
		Price:          "10000000",
		L1BlockHeight:  100,
		L1BlockTime:    1000000000,
		CurrencyPairId: 1,
		Nonce:          1,
	}

	err = input.OPChildKeeper.HandleOracleDataPacket(ctx, oracleData)
	require.Error(t, err)
	require.Contains(t, err.Error(), types.ErrInvalidBridgeInfo.Error())
}

func Test_HandleOracleDataPacket_OracleDisabled(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	bridgeInfo := types.BridgeInfo{
		BridgeId:  1,
		L1ChainId: "test-chain-1",
		BridgeConfig: ophosttypes.BridgeConfig{
			OracleEnabled: false,
		},
	}
	err := input.OPChildKeeper.BridgeInfo.Set(ctx, bridgeInfo)
	require.NoError(t, err)

	oracleData := types.OracleData{
		BridgeId:       1,
		CurrencyPair:   "BTC/USD",
		Price:          "10000000",
		L1BlockHeight:  100,
		L1BlockTime:    1000000000,
		CurrencyPairId: 1,
		Nonce:          1,
	}

	err = input.OPChildKeeper.HandleOracleDataPacket(ctx, oracleData)
	require.Error(t, err)
	require.Contains(t, err.Error(), types.ErrOracleDisabled.Error())
}

func Test_HandleOracleDataPacket_InvalidCurrencyPair(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	bridgeInfo := types.BridgeInfo{
		BridgeId:  1,
		L1ChainId: "test-chain-1",
		BridgeConfig: ophosttypes.BridgeConfig{
			OracleEnabled: true,
		},
	}
	err := input.OPChildKeeper.BridgeInfo.Set(ctx, bridgeInfo)
	require.NoError(t, err)

	oracleData := types.OracleData{
		BridgeId:       1,
		CurrencyPair:   "INVALID-PAIR", // invalid format
		Price:          "10000000",
		L1BlockHeight:  100,
		L1BlockTime:    1000000000,
		CurrencyPairId: 1,
		Nonce:          1,
	}

	err = input.OPChildKeeper.HandleOracleDataPacket(ctx, oracleData)
	require.Error(t, err)
}

func Test_ProcessOraclePriceUpdate_InvalidPrice(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	input.OracleKeeper.InitGenesis(sdk.UnwrapSDKContext(ctx), oracletypes.GenesisState{
		CurrencyPairGenesis: make([]oracletypes.CurrencyPairGenesis, 0),
	})

	cp, err := connecttypes.CurrencyPairFromString("BTC/USD")
	require.NoError(t, err)
	err = input.OracleKeeper.CreateCurrencyPair(sdk.UnwrapSDKContext(ctx), cp)
	require.NoError(t, err)

	bridgeInfo := types.BridgeInfo{
		BridgeId:  1,
		L1ChainId: "test-chain-1",
		BridgeConfig: ophosttypes.BridgeConfig{
			OracleEnabled: true,
		},
	}
	err = input.OPChildKeeper.BridgeInfo.Set(ctx, bridgeInfo)
	require.NoError(t, err)

	oracleData := types.OracleData{
		BridgeId:       1,
		CurrencyPair:   "BTC/USD",
		Price:          "invalid-price", // invalid price
		L1BlockHeight:  100,
		L1BlockTime:    1000000000,
		CurrencyPairId: 1,
		Nonce:          1,
	}

	err = input.OPChildKeeper.HandleOracleDataPacket(ctx, oracleData)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid price format")
}

func Test_ConvertProofOpsToMerkleProof_InvalidProof(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	bridgeInfo := types.BridgeInfo{
		BridgeId:   1,
		L1ChainId:  "test-chain-1",
		L1ClientId: "test-client-id",
		BridgeConfig: ophosttypes.BridgeConfig{
			OracleEnabled: true,
		},
	}
	err := input.OPChildKeeper.BridgeInfo.Set(ctx, bridgeInfo)
	require.NoError(t, err)

	oracleData := types.OracleData{
		BridgeId:       1,
		CurrencyPair:   "BTC/USD",
		Price:          "10000000",
		L1BlockHeight:  100,
		L1BlockTime:    1000000000,
		CurrencyPairId: 1,
		Nonce:          1,
		Proof:          []byte("invalid-proof"),
		ProofHeight: clienttypes.Height{
			RevisionNumber: 0,
			RevisionHeight: 100,
		},
	}

	err = input.OPChildKeeper.HandleOracleDataPacket(ctx, oracleData)
	require.Error(t, err)
	require.Contains(t, err.Error(), "oracle state proof verification failed")
}

func Test_ProcessOraclePriceUpdate_Success(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	tmclient.RegisterInterfaces(input.Cdc.InterfaceRegistry())

	input.OracleKeeper.InitGenesis(sdk.UnwrapSDKContext(ctx), oracletypes.GenesisState{
		CurrencyPairGenesis: make([]oracletypes.CurrencyPairGenesis, 0),
	})

	cp, err := connecttypes.CurrencyPairFromString("BTC/USD")
	require.NoError(t, err)
	err = input.OracleKeeper.CreateCurrencyPair(sdk.UnwrapSDKContext(ctx), cp)
	require.NoError(t, err)

	bridgeInfo := types.BridgeInfo{
		BridgeId:   1,
		L1ChainId:  "mock-network-1",
		L1ClientId: "test-client-id-1",
		BridgeConfig: ophosttypes.BridgeConfig{
			OracleEnabled: true,
		},
	}
	err = input.OPChildKeeper.BridgeInfo.Set(ctx, bridgeInfo)
	require.NoError(t, err)

	// setup l1 client
	appHash, _ := hex.DecodeString("5EFAD542D8F32C8E4D23BD5F27D4E8441FEB4D857EC49AB1985C3ADA30BDD932")
	nextValidatorHash, _ := hex.DecodeString("1D7083EEA750237397B09BC331092CB301D36BCCF6F793A41F049590CB607B39")

	clientState := tmclient.NewClientState(
		"mock-network-1",
		tmclient.DefaultTrustLevel,
		24*time.Hour*7,
		24*time.Hour*21,
		10*time.Second,
		clienttypes.NewHeight(1, 260),
		commitmenttypes.GetSDKSpecs(),
		[]string{"upgrade", "upgradedIBCState"},
	)
	consensusState := &tmclient.ConsensusState{
		Timestamp:          time.Date(2025, 12, 1, 9, 42, 43, 311862000, time.UTC),
		Root:               commitmenttypes.NewMerkleRoot(appHash),
		NextValidatorsHash: nextValidatorHash,
	}

	input.ClientKeeper.SetClientState("test-client-id-1", clientState)
	input.ClientKeeper.SetConsensusState("test-client-id-1", consensusState)

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// we use a simple in-memory kv store here
	memDB := dbm.NewMemDB()
	clientStore := dbadapter.Store{DB: memDB}

	// store consensus state at proof height
	csBytes, err := clienttypes.MarshalClientState(input.Cdc, clientState)
	require.NoError(t, err)
	clientStore.Set(host.ClientStateKey(), csBytes)

	height := clienttypes.NewHeight(1, 260)
	bz, err := clienttypes.MarshalConsensusState(input.Cdc, consensusState)
	require.NoError(t, err)
	clientStore.Set(host.ConsensusStateKey(height), bz)

	clientStore.Set(tmclient.IterationKey(height), []byte{})

	processedTime := uint64(consensusState.Timestamp.UnixNano())
	processedTimeBz := sdk.Uint64ToBigEndian(processedTime)
	clientStore.Set(tmclient.ProcessedTimeKey(height), processedTimeBz)

	processedHeight := uint64(sdkCtx.BlockHeight())
	processedHeightBz := sdk.Uint64ToBigEndian(processedHeight)
	clientStore.Set(tmclient.ProcessedHeightKey(height), processedHeightBz)

	input.ClientKeeper.SetClientStore("test-client-id-1", clientStore)

	// verifiable against consensus state at height 260
	proofBytes, _ := hex.DecodeString("0aea020a0a69637332333a6961766c1208004254432f5553441ad1020ace020a08004254432f55534412240a1d0a0a38363736313034303030120c0891c6b5c90610b0baed880118830210ec0118041a0c0801180120012a0400028604222a0801122602048604206ee927733ce2b4ffb368c814b27abfed923364a7b54b2efdbea77a48950efdd420222c0801120504088604201a21207a504df3f3c9f2a7b53d047a3d7403c588e3903ee325485297320350f3b1055d222a0801122606108604201efcc2e0d095cd4dcb10edcfc3c7e65892f0720274c33b4f98d69e9dc57cb4e920222c0801120508208604201a2120161800032403778b6cd0f0b65de07a6ef5ec5d39366574974eef3c5d34c4156b222c080112050a408604201a21203fccd731c12d8850d02254f575b6528c13edcc242dc690d2361cbd4fae8ff6bd222c080112050c668604201a212093f3567e97dbfefe86385f760a6af5904e19c8b535646a239ab0c3395f7bcfab0a9a020a0c69637332333a73696d706c6512066f7261636c651a81020afe010a066f7261636c651220207cad2e8b66dea0835794b3d64bb57f2a5d16ce5a707276dc07e5d259e77e011a090801180120012a0100222708011201011a20780b19babb93a320e633fbd41315c0933f7b0d74e4fed7ece3cd7a2eb98a3429222708011201011a204d08b46784e130550d4a800c32d9c68a7be3cf0e6bbe167c693f87702d473052222708011201011a20078064c6806fd328ae2c5c62c05890470465411d136d3b81a0ff76180b1a06902225080112210186fceb2a2ecfad732820a4462bf727771a37efc0ec68ddf290856ac8f45e189122250801122101c826b470b22b9b72643fe938b2c3191bf0d4423654e1066093b15401f5c6207e")
	oracleData := types.OracleData{
		BridgeId:       1,
		CurrencyPair:   "BTC/USD",
		Price:          "8676104000",
		Decimals:       5,
		L1BlockHeight:  259,
		L1BlockTime:    1764582161287006000,
		Proof:          proofBytes,
		ProofHeight:    clienttypes.NewHeight(1, 260),
		Nonce:          236,
		CurrencyPairId: 4,
	}

	err = input.OPChildKeeper.HandleOracleDataPacket(ctx, oracleData)
	require.NoError(t, err)

	// now verify data
	updatedPrice, err := input.OracleKeeper.GetPriceForCurrencyPair(ctx, cp)
	require.NoError(t, err)
	require.Equal(t, "8676104000", updatedPrice.Price.String())
	require.Equal(t, time.Unix(0, 1764582161287006000).UTC(), updatedPrice.BlockTimestamp)

	events := sdkCtx.EventManager().Events()
	found := false
	for _, event := range events {
		if event.Type == types.EventTypeOracleDataRelay {
			found = true
			require.Equal(t, "1", event.Attributes[0].Value)
			require.Equal(t, "259", event.Attributes[1].Value)
			require.Equal(t, "BTC/USD", event.Attributes[2].Value)
			require.Equal(t, "8676104000", event.Attributes[3].Value)
			break
		}
	}
	require.True(t, found, "expected oracle data relay event to be emitted")
}
