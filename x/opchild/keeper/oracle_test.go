package keeper_test

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/store/dbadapter"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/initia-labs/OPinit/x/opchild/testutil"
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
			ctx, input := testutil.CreateTestInput(t, false)
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
			ctx, input := testutil.CreateTestInput(t, false)
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
	ctx, input := testutil.CreateTestInput(t, false)

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
		BridgeId:        2, // mismatched bridge ID
		OraclePriceHash: make([]byte, 32),
		Prices: []types.OraclePriceData{
			{
				CurrencyPair:   "BTC/USD",
				Price:          "10000000",
				Decimals:       5,
				CurrencyPairId: 1,
				Nonce:          1,
			},
		},
		L1BlockHeight: 100,
		L1BlockTime:   1000000000,
	}

	err = input.OPChildKeeper.HandleOracleDataPacket(ctx, oracleData)
	require.Error(t, err)
	require.Contains(t, err.Error(), types.ErrInvalidBridgeInfo.Error())
}

func Test_HandleOracleDataPacket_OracleDisabled(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

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
		BridgeId:        1,
		OraclePriceHash: make([]byte, 32),
		Prices: []types.OraclePriceData{
			{
				CurrencyPair:   "BTC/USD",
				Price:          "10000000",
				Decimals:       5,
				CurrencyPairId: 1,
				Nonce:          1,
			},
		},
		L1BlockHeight: 100,
		L1BlockTime:   1000000000,
	}

	err = input.OPChildKeeper.HandleOracleDataPacket(ctx, oracleData)
	require.Error(t, err)
	require.Contains(t, err.Error(), types.ErrOracleDisabled.Error())
}

func Test_ConvertProofOpsToMerkleProof_InvalidProof(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

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
		BridgeId:        1,
		OraclePriceHash: make([]byte, 32),
		Prices: []types.OraclePriceData{
			{
				CurrencyPair:   "BTC/USD",
				Price:          "10000000",
				Decimals:       5,
				CurrencyPairId: 1,
				Nonce:          1,
			},
		},
		L1BlockHeight: 100,
		L1BlockTime:   1000000000,
		Proof:         []byte("invalid-proof"),
		ProofHeight: clienttypes.Height{
			RevisionNumber: 0,
			RevisionHeight: 100,
		},
	}

	err = input.OPChildKeeper.HandleOracleDataPacket(ctx, oracleData)
	require.Error(t, err)
	require.Contains(t, err.Error(), "oracle hash proof verification failed")
}

// Test_BatchedOracleRelay_WithIBCSetup tests the batched oracle relay flow
// with full IBC client state setup, similar to the single-price test.
// This validates that batched oracle data works through the complete verification flow.
func Test_BatchedOracleRelay_WithIBCSetup(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	tmclient.RegisterInterfaces(input.Cdc.InterfaceRegistry())

	input.OracleKeeper.InitGenesis(sdk.UnwrapSDKContext(ctx), oracletypes.GenesisState{
		CurrencyPairGenesis: make([]oracletypes.CurrencyPairGenesis, 0),
	})

	// create all 16 currency pairs that exist on L1
	allPairs := []string{
		"TIMESTAMP/NANOSECOND", "ATOM/USD", "OSMO/USD", "BNB/USD", "BTC/USD", "NTRN/USD",
		"USDT/USD", "USDC/USD", "SUI/USD", "SOL/USD", "TIA/USD", "BERA/USD",
		"ENA/USD", "APT/USD", "ARB/USD", "ETH/USD",
	}
	for _, pair := range allPairs {
		cp, err := connecttypes.CurrencyPairFromString(pair)
		require.NoError(t, err)
		err = input.OracleKeeper.CreateCurrencyPair(sdk.UnwrapSDKContext(ctx), cp)
		require.NoError(t, err)
	}

	bridgeInfo := types.BridgeInfo{
		BridgeId:   1,
		L1ChainId:  "testchain-1",
		L1ClientId: "test-client-id-1",
		BridgeConfig: ophosttypes.BridgeConfig{
			OracleEnabled: true,
		},
	}
	err := input.OPChildKeeper.BridgeInfo.Set(ctx, bridgeInfo)
	require.NoError(t, err)

	appHash, _ := hex.DecodeString("73A6EB1E530A6F451C4AE7373206D0FF44C859B452DD801FAF3385EA2AB209C2")
	nextValidatorHash, _ := hex.DecodeString("0A03067E7FA76E1A6880933EF515FFD2E391A3E8A883B3AE26A098718E44E7AD")

	clientState := tmclient.NewClientState(
		"testchain-1",
		tmclient.DefaultTrustLevel,
		24*time.Hour*7,
		24*time.Hour*21,
		10*time.Second,
		clienttypes.NewHeight(1, 664),
		commitmenttypes.GetSDKSpecs(),
		[]string{"upgrade", "upgradedIBCState"},
	)

	consensusState := &tmclient.ConsensusState{
		Timestamp:          time.Date(2025, 12, 12, 9, 45, 35, 356266000, time.UTC),
		Root:               commitmenttypes.NewMerkleRoot(appHash),
		NextValidatorsHash: nextValidatorHash,
	}

	input.ClientKeeper.SetClientState("test-client-id-1", clientState)
	input.ClientKeeper.SetConsensusState("test-client-id-1", consensusState)

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	memDB := dbm.NewMemDB()
	clientStore := dbadapter.Store{DB: memDB}

	csBytes, err := clienttypes.MarshalClientState(input.Cdc, clientState)
	require.NoError(t, err)
	clientStore.Set(host.ClientStateKey(), csBytes)

	height := clienttypes.NewHeight(1, 664)
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

	proofBytes, _ := hex.DecodeString("0aeb010a0a69637332333a6961766c1209a100000000000000011ad1010ace010a09a10000000000000001122f0a20d3f45ee700ee1a37b1b114d7e12f44d2514cafaa2308fa0446c324f3491453f710970518f8eab787abd59bc0181a0c0801180120012a040002ae0a222a080112260204ae0a201fdb1e2e60ff88045cd3f68bd0da4593b20d710c13725f98e239dd014fbac36020222a080112260406ae0a20d0a1e370a3c38d3c05c109c408335caa79e61e6005cfba0fb1da3fc640c55cd820222a08011226060aae0a20bbe83488fa11800d0da08a55ae8d9201868378c190c793f5fcc294f7fb61faae200a96020a0c69637332333a73696d706c6512066f70686f73741afd010afa010a066f70686f737412200cecd8af22290cfdd98431e98cccb6418b7b27a3a9914616b0237d3d2f027ccb1a090801180120012a01002225080112210141c04a42e47994d4dbcb3060eedd792d7b623f1bf61f4f7498bd5e3e1119666a22250801122101a9a139f033e62375c436e237e8fb7deddb3a5ade8a485d45e0229d8b8ea9d937222508011221010f31c2c269df8cc8cccdc393e8c168cd397589427bf4fc70646b263f9389529a222708011201011a20cfff89b3f02d88273b060b24913e22bfe25d236a2cd25c21d142c6d93a1eb92e22250801122101bada0d57b7de3594067a4227846a20361bb1429f632d0cd0cde8f31395b96772")

	testL1BlockTime := int64(1765532733321115000)
	timestampPrice, _ := math.NewIntFromString("1765532732375394000")
	testPrices := []ophosttypes.OraclePriceInfo{
		{CurrencyPairId: 0, Price: timestampPrice, Timestamp: testL1BlockTime},           // TIMESTAMP/NANOSECOND
		{CurrencyPairId: 1, Price: math.NewInt(2180500000), Timestamp: testL1BlockTime},  // ATOM/USD
		{CurrencyPairId: 2, Price: math.NewInt(69120736), Timestamp: testL1BlockTime},    // OSMO/USD
		{CurrencyPairId: 3, Price: math.NewInt(8866564969), Timestamp: testL1BlockTime},  // BNB/USD
		{CurrencyPairId: 4, Price: math.NewInt(9222376713), Timestamp: testL1BlockTime},  // BTC/USD
		{CurrencyPairId: 5, Price: math.NewInt(2980894), Timestamp: testL1BlockTime},     // NTRN/USD
		{CurrencyPairId: 6, Price: math.NewInt(1000300090), Timestamp: testL1BlockTime},  // USDT/USD
		{CurrencyPairId: 7, Price: math.NewInt(999800000), Timestamp: testL1BlockTime},   // USDC/USD
		{CurrencyPairId: 8, Price: math.NewInt(16278383515), Timestamp: testL1BlockTime}, // SUI/USD
		{CurrencyPairId: 9, Price: math.NewInt(13750125037), Timestamp: testL1BlockTime}, // SOL/USD
		{CurrencyPairId: 10, Price: math.NewInt(59214999), Timestamp: testL1BlockTime},   // TIA/USD
		{CurrencyPairId: 11, Price: math.NewInt(72521756), Timestamp: testL1BlockTime},   // BERA/USD
		{CurrencyPairId: 12, Price: math.NewInt(26347904), Timestamp: testL1BlockTime},   // ENA/USD
		{CurrencyPairId: 13, Price: math.NewInt(1706011803), Timestamp: testL1BlockTime}, // APT/USD
		{CurrencyPairId: 14, Price: math.NewInt(213564069), Timestamp: testL1BlockTime},  // ARB/USD
		{CurrencyPairId: 15, Price: math.NewInt(3243513053), Timestamp: testL1BlockTime}, // ETH/USD
	}
	computedHash := ophosttypes.OraclePriceInfos(testPrices).ComputeOraclePricesHash()

	oracleData := types.OracleData{
		BridgeId:        1,
		OraclePriceHash: computedHash,
		Prices: []types.OraclePriceData{
			{CurrencyPair: "TIMESTAMP/NANOSECOND", Price: "1765532732375394000", Decimals: 8, CurrencyPairId: 0, Nonce: 571},
			{CurrencyPair: "ATOM/USD", Price: "2180500000", Decimals: 9, CurrencyPairId: 1, Nonce: 567},
			{CurrencyPair: "OSMO/USD", Price: "69120736", Decimals: 9, CurrencyPairId: 2, Nonce: 563},
			{CurrencyPair: "BNB/USD", Price: "8866564969", Decimals: 7, CurrencyPairId: 3, Nonce: 563},
			{CurrencyPair: "BTC/USD", Price: "9222376713", Decimals: 5, CurrencyPairId: 4, Nonce: 567},
			{CurrencyPair: "NTRN/USD", Price: "2980894", Decimals: 8, CurrencyPairId: 5, Nonce: 565},
			{CurrencyPair: "USDT/USD", Price: "1000300090", Decimals: 9, CurrencyPairId: 6, Nonce: 571},
			{CurrencyPair: "USDC/USD", Price: "999800000", Decimals: 9, CurrencyPairId: 7, Nonce: 571},
			{CurrencyPair: "SUI/USD", Price: "16278383515", Decimals: 10, CurrencyPairId: 8, Nonce: 567},
			{CurrencyPair: "SOL/USD", Price: "13750125037", Decimals: 8, CurrencyPairId: 9, Nonce: 567},
			{CurrencyPair: "TIA/USD", Price: "59214999", Decimals: 8, CurrencyPairId: 10, Nonce: 567},
			{CurrencyPair: "BERA/USD", Price: "72521756", Decimals: 8, CurrencyPairId: 11, Nonce: 567},
			{CurrencyPair: "ENA/USD", Price: "26347904", Decimals: 8, CurrencyPairId: 12, Nonce: 567},
			{CurrencyPair: "APT/USD", Price: "1706011803", Decimals: 9, CurrencyPairId: 13, Nonce: 567},
			{CurrencyPair: "ARB/USD", Price: "213564069", Decimals: 9, CurrencyPairId: 14, Nonce: 567},
			{CurrencyPair: "ETH/USD", Price: "3243513053", Decimals: 6, CurrencyPairId: 15, Nonce: 567},
		},
		L1BlockHeight: 663,
		L1BlockTime:   testL1BlockTime,
		Proof:         proofBytes,
		ProofHeight:   clienttypes.NewHeight(1, 664),
	}

	err = input.OPChildKeeper.HandleOracleDataPacket(ctx, oracleData)
	require.NoError(t, err, "batched oracle relay should succeed with all 16 prices matching L1 hash")

	btcPair, _ := connecttypes.CurrencyPairFromString("BTC/USD")
	btcPrice, err := input.OracleKeeper.GetPriceForCurrencyPair(ctx, btcPair)
	require.NoError(t, err)
	require.Equal(t, "9222376713", btcPrice.Price.String())
	require.Equal(t, time.Unix(0, 1765532733321115000).UTC(), btcPrice.BlockTimestamp)

	ethPair, _ := connecttypes.CurrencyPairFromString("ETH/USD")
	ethPrice, err := input.OracleKeeper.GetPriceForCurrencyPair(ctx, ethPair)
	require.NoError(t, err)
	require.Equal(t, "3243513053", ethPrice.Price.String())
	require.Equal(t, time.Unix(0, 1765532733321115000).UTC(), ethPrice.BlockTimestamp)

	atomPair, _ := connecttypes.CurrencyPairFromString("ATOM/USD")
	atomPrice, err := input.OracleKeeper.GetPriceForCurrencyPair(ctx, atomPair)
	require.NoError(t, err)
	require.Equal(t, "2180500000", atomPrice.Price.String())
	require.Equal(t, time.Unix(0, 1765532733321115000).UTC(), atomPrice.BlockTimestamp)

	events := sdkCtx.EventManager().Events()
	found := false
	for _, event := range events {
		if event.Type == types.EventTypeOracleDataRelay {
			found = true
			var bridgeIdFound, l1HeightFound, numPairsFound bool
			for _, attr := range event.Attributes {
				switch attr.Key {
				case types.AttributeKeyBridgeId:
					require.Equal(t, "1", attr.Value)
					bridgeIdFound = true
				case types.AttributeKeyL1BlockHeight:
					require.Equal(t, "663", attr.Value)
					l1HeightFound = true
				case types.AttributeKeyNumCurrencyPair:
					require.Equal(t, "16", attr.Value)
					numPairsFound = true
				}
			}
			require.True(t, bridgeIdFound && l1HeightFound && numPairsFound, "all expected attributes should be present")
			break
		}
	}
	require.True(t, found, "expected oracle data relay event to be emitted")
}
