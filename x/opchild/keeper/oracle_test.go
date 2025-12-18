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
				Timestamp:      1000000000,
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
				Timestamp:      1000000000,
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
				Timestamp:      1000000000,
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

	appHash, err := hex.DecodeString("5EA27D296380CCBD058286BAB9A7309F4821FDFD5C0EA79B43452F935943A0F7")
	require.NoError(t, err)
	nextValidatorHash, err := hex.DecodeString("AB1BD1DFFAFC92506E5E8A29C3DBC133455D5F6A04F037BE1D865F2E5233ADE8")
	require.NoError(t, err)

	clientState := tmclient.NewClientState(
		"testchain-1",
		tmclient.DefaultTrustLevel,
		24*time.Hour*7,
		24*time.Hour*21,
		10*time.Second,
		clienttypes.NewHeight(1, 77),
		commitmenttypes.GetSDKSpecs(),
		[]string{"upgrade", "upgradedIBCState"},
	)

	consensusState := &tmclient.ConsensusState{
		Timestamp:          time.Date(2025, 12, 15, 9, 48, 57, 131147000, time.UTC),
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

	height := clienttypes.NewHeight(1, 77)
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

	proofBytes, err := hex.DecodeString("0aae010a0a69637332333a6961766c1201a11a9c010a99010a01a1122e0a202426ac051cc798ba02498fea05d5339f9faec21636a86451fd385fefd0cbdcf0104c18e0f2ce80f6d2d6c0181a0c0801180120012a0400029801222a08011226020498012075596157e985487234ce2decc134457070eda3b58b91ae7650cae591d28c5e3420222a080112260406980120a8b6e260143cdc29339aaee80e08969a7aabb37b277a610a05b0929f4b498f7b200a96020a0c69637332333a73696d706c6512066f70686f73741afd010afa010a066f70686f737412206a3019435bcd8ff099cf7fdce28999492238c34861490ed43b291446e1a90e5d1a090801180120012a01002225080112210141c04a42e47994d4dbcb3060eedd792d7b623f1bf61f4f7498bd5e3e1119666a2225080112210164f8a73f872520985d0223e32e62c7b0abe1a4a5b28291971287e31d8d2018f222250801122101ca8b13c9a346351958db29e912ac6c6f077ca5daad31e42b24e0c1f2188279c8222708011201011a2051d6a3064852fb1c963e30cf47a449b53579124da7dcfe72fcd745c73467891e2225080112210158954f6c5b67cd1cd66210fe76713db5a80e8f96f36c357782c87bea6a22d86b")
	require.NoError(t, err)

	testL1BlockTime := int64(1765792135104412000)
	timestampPrice, _ := math.NewIntFromString("1765792134345144000")
	testPrices := []ophosttypes.OraclePriceInfo{
		{CurrencyPairId: 0, CurrencyPairString: "TIMESTAMP/NANOSECOND", Price: timestampPrice, Timestamp: testL1BlockTime},
		{CurrencyPairId: 1, CurrencyPairString: "ATOM/USD", Price: math.NewInt(2147129425), Timestamp: testL1BlockTime},
		{CurrencyPairId: 2, CurrencyPairString: "OSMO/USD", Price: math.NewInt(63231311), Timestamp: testL1BlockTime},
		{CurrencyPairId: 3, CurrencyPairString: "BNB/USD", Price: math.NewInt(8903780756), Timestamp: testL1BlockTime},
		{CurrencyPairId: 4, CurrencyPairString: "BTC/USD", Price: math.NewInt(8983796759), Timestamp: testL1BlockTime},
		{CurrencyPairId: 5, CurrencyPairString: "NTRN/USD", Price: math.NewInt(2785557), Timestamp: testL1BlockTime},
		{CurrencyPairId: 6, CurrencyPairString: "USDT/USD", Price: math.NewInt(1000200040), Timestamp: testL1BlockTime},
		{CurrencyPairId: 7, CurrencyPairString: "USDC/USD", Price: math.NewInt(999800000), Timestamp: testL1BlockTime},
		{CurrencyPairId: 8, CurrencyPairString: "SUI/USD", Price: math.NewInt(15696139227), Timestamp: testL1BlockTime},
		{CurrencyPairId: 9, CurrencyPairString: "SOL/USD", Price: math.NewInt(13250650130), Timestamp: testL1BlockTime},
		{CurrencyPairId: 10, CurrencyPairString: "TIA/USD", Price: math.NewInt(55051010), Timestamp: testL1BlockTime},
		{CurrencyPairId: 11, CurrencyPairString: "BERA/USD", Price: math.NewInt(72214442), Timestamp: testL1BlockTime},
		{CurrencyPairId: 12, CurrencyPairString: "ENA/USD", Price: math.NewInt(23744748), Timestamp: testL1BlockTime},
		{CurrencyPairId: 13, CurrencyPairString: "APT/USD", Price: math.NewInt(1649965023), Timestamp: testL1BlockTime},
		{CurrencyPairId: 14, CurrencyPairString: "ARB/USD", Price: math.NewInt(209041808), Timestamp: testL1BlockTime},
		{CurrencyPairId: 15, CurrencyPairString: "ETH/USD", Price: math.NewInt(3144108821), Timestamp: testL1BlockTime},
	}
	computedHash := ophosttypes.OraclePriceInfos(testPrices).ComputeOraclePricesHash()

	oracleData := types.OracleData{
		BridgeId:        1,
		OraclePriceHash: computedHash,
		Prices: []types.OraclePriceData{
			{CurrencyPair: "TIMESTAMP/NANOSECOND", Price: "1765792134345144000", Decimals: 8, CurrencyPairId: 0, Nonce: 58, Timestamp: testL1BlockTime},
			{CurrencyPair: "ATOM/USD", Price: "2147129425", Decimals: 9, CurrencyPairId: 1, Nonce: 56, Timestamp: testL1BlockTime},
			{CurrencyPair: "OSMO/USD", Price: "63231311", Decimals: 9, CurrencyPairId: 2, Nonce: 53, Timestamp: testL1BlockTime},
			{CurrencyPair: "BNB/USD", Price: "8903780756", Decimals: 7, CurrencyPairId: 3, Nonce: 55, Timestamp: testL1BlockTime},
			{CurrencyPair: "BTC/USD", Price: "8983796759", Decimals: 5, CurrencyPairId: 4, Nonce: 56, Timestamp: testL1BlockTime},
			{CurrencyPair: "NTRN/USD", Price: "2785557", Decimals: 8, CurrencyPairId: 5, Nonce: 53, Timestamp: testL1BlockTime},
			{CurrencyPair: "USDT/USD", Price: "1000200040", Decimals: 9, CurrencyPairId: 6, Nonce: 58, Timestamp: testL1BlockTime},
			{CurrencyPair: "USDC/USD", Price: "999800000", Decimals: 9, CurrencyPairId: 7, Nonce: 58, Timestamp: testL1BlockTime},
			{CurrencyPair: "SUI/USD", Price: "15696139227", Decimals: 10, CurrencyPairId: 8, Nonce: 56, Timestamp: testL1BlockTime},
			{CurrencyPair: "SOL/USD", Price: "13250650130", Decimals: 8, CurrencyPairId: 9, Nonce: 56, Timestamp: testL1BlockTime},
			{CurrencyPair: "TIA/USD", Price: "55051010", Decimals: 8, CurrencyPairId: 10, Nonce: 56, Timestamp: testL1BlockTime},
			{CurrencyPair: "BERA/USD", Price: "72214442", Decimals: 8, CurrencyPairId: 11, Nonce: 56, Timestamp: testL1BlockTime},
			{CurrencyPair: "ENA/USD", Price: "23744748", Decimals: 8, CurrencyPairId: 12, Nonce: 56, Timestamp: testL1BlockTime},
			{CurrencyPair: "APT/USD", Price: "1649965023", Decimals: 9, CurrencyPairId: 13, Nonce: 56, Timestamp: testL1BlockTime},
			{CurrencyPair: "ARB/USD", Price: "209041808", Decimals: 9, CurrencyPairId: 14, Nonce: 57, Timestamp: testL1BlockTime},
			{CurrencyPair: "ETH/USD", Price: "3144108821", Decimals: 6, CurrencyPairId: 15, Nonce: 56, Timestamp: testL1BlockTime},
		},
		L1BlockHeight: 76,
		L1BlockTime:   testL1BlockTime,
		Proof:         proofBytes,
		ProofHeight:   clienttypes.NewHeight(1, 77),
	}

	err = input.OPChildKeeper.HandleOracleDataPacket(ctx, oracleData)
	require.NoError(t, err, "batched oracle relay should succeed with all 16 prices matching L1 hash")

	// verify prices were updated correctly
	btcPair, err := connecttypes.CurrencyPairFromString("BTC/USD")
	require.NoError(t, err)
	btcPrice, err := input.OracleKeeper.GetPriceForCurrencyPair(ctx, btcPair)
	require.NoError(t, err)
	require.Equal(t, "8983796759", btcPrice.Price.String())
	require.Equal(t, time.Unix(0, testL1BlockTime).UTC(), btcPrice.BlockTimestamp)

	ethPair, err := connecttypes.CurrencyPairFromString("ETH/USD")
	require.NoError(t, err)
	ethPrice, err := input.OracleKeeper.GetPriceForCurrencyPair(ctx, ethPair)
	require.NoError(t, err)
	require.Equal(t, "3144108821", ethPrice.Price.String())
	require.Equal(t, time.Unix(0, testL1BlockTime).UTC(), ethPrice.BlockTimestamp)

	atomPair, err := connecttypes.CurrencyPairFromString("ATOM/USD")
	require.NoError(t, err)
	atomPrice, err := input.OracleKeeper.GetPriceForCurrencyPair(ctx, atomPair)
	require.NoError(t, err)
	require.Equal(t, "2147129425", atomPrice.Price.String())
	require.Equal(t, time.Unix(0, testL1BlockTime).UTC(), atomPrice.BlockTimestamp)

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
					require.Equal(t, "76", attr.Value)
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
