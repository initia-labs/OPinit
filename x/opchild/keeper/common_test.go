package keeper_test

import (
	"context"
	"encoding/binary"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	signingmod "cosmossdk.io/x/tx/signing"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codecaddress "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsign "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"

	opchild "github.com/initia-labs/OPinit/x/opchild"
	opchildkeeper "github.com/initia-labs/OPinit/x/opchild/keeper"
	opchildtypes "github.com/initia-labs/OPinit/x/opchild/types"
	oraclekeeper "github.com/skip-mev/slinky/x/oracle/keeper"
	oracletypes "github.com/skip-mev/slinky/x/oracle/types"
)

var ModuleBasics = module.NewBasicManager(
	auth.AppModuleBasic{},
	bank.AppModuleBasic{},
	opchild.AppModuleBasic{},
)

var (
	pubKeys = []cryptotypes.PubKey{
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
	}

	addrs = []sdk.AccAddress{
		sdk.AccAddress(pubKeys[0].Address()),
		sdk.AccAddress(pubKeys[1].Address()),
		sdk.AccAddress(pubKeys[2].Address()),
		sdk.AccAddress(pubKeys[3].Address()),
		sdk.AccAddress(pubKeys[4].Address()),
	}

	addrsStr = []string{
		addrs[0].String(),
		addrs[1].String(),
		addrs[2].String(),
		addrs[3].String(),
		addrs[4].String(),
	}

	valAddrs = []sdk.ValAddress{
		sdk.ValAddress(pubKeys[0].Address()),
		sdk.ValAddress(pubKeys[1].Address()),
		sdk.ValAddress(pubKeys[2].Address()),
		sdk.ValAddress(pubKeys[3].Address()),
		sdk.ValAddress(pubKeys[4].Address()),
	}

	valAddrsStr = []string{
		valAddrs[0].String(),
		valAddrs[1].String(),
		valAddrs[2].String(),
		valAddrs[3].String(),
		valAddrs[4].String(),
	}

	testDenoms = []string{
		"test1",
		"test2",
		"test3",
		"test4",
		"test5",
	}

	initiaSupply = math.NewInt(100_000_000_000)
)

type EncodingConfig struct {
	InterfaceRegistry codectypes.InterfaceRegistry
	Marshaler         codec.Codec
	TxConfig          client.TxConfig
	Amino             *codec.LegacyAmino
}

func MakeTestCodec(t testing.TB) codec.Codec {
	return MakeEncodingConfig(t).Marshaler
}

func MakeEncodingConfig(_ testing.TB) EncodingConfig {
	interfaceRegistry, _ := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signingmod.Options{
			AddressCodec:          codecaddress.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
			ValidatorAddressCodec: codecaddress.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		},
	})
	appCodec := codec.NewProtoCodec(interfaceRegistry)
	legacyAmino := codec.NewLegacyAmino()
	txConfig := authtx.NewTxConfig(appCodec, authtx.DefaultSignModes)

	std.RegisterInterfaces(interfaceRegistry)
	std.RegisterLegacyAminoCodec(legacyAmino)

	ModuleBasics.RegisterLegacyAminoCodec(legacyAmino)
	ModuleBasics.RegisterInterfaces(interfaceRegistry)

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         appCodec,
		TxConfig:          txConfig,
		Amino:             legacyAmino,
	}
}

func initialTotalSupply() sdk.Coins {
	faucetBalance := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initiaSupply))
	for _, testDenom := range testDenoms {
		faucetBalance = faucetBalance.Add(sdk.NewCoin(testDenom, initiaSupply))
	}

	return faucetBalance
}

type TestFaucet struct {
	t                testing.TB
	bankKeeper       bankkeeper.Keeper
	sender           sdk.AccAddress
	balance          sdk.Coins
	minterModuleName string
}

func NewTestFaucet(t testing.TB, ctx context.Context, bankKeeper bankkeeper.Keeper, minterModuleName string, initiaSupply ...sdk.Coin) *TestFaucet {
	require.NotEmpty(t, initiaSupply)
	r := &TestFaucet{t: t, bankKeeper: bankKeeper, minterModuleName: minterModuleName}
	_, _, addr := keyPubAddr()
	r.sender = addr
	r.Mint(ctx, addr, initiaSupply...)
	r.balance = initiaSupply
	return r
}

func (f *TestFaucet) Mint(parentCtx context.Context, addr sdk.AccAddress, amounts ...sdk.Coin) {
	amounts = sdk.Coins(amounts).Sort()
	require.NotEmpty(f.t, amounts)
	ctx := sdk.UnwrapSDKContext(parentCtx).WithEventManager(sdk.NewEventManager()) // discard all faucet related events
	err := f.bankKeeper.MintCoins(ctx, f.minterModuleName, amounts)
	require.NoError(f.t, err)
	err = f.bankKeeper.SendCoinsFromModuleToAccount(ctx, f.minterModuleName, addr, amounts)
	require.NoError(f.t, err)
	f.balance = f.balance.Add(amounts...)
}

func (f *TestFaucet) Fund(parentCtx context.Context, receiver sdk.AccAddress, amounts ...sdk.Coin) {
	require.NotEmpty(f.t, amounts)
	// ensure faucet is always filled
	if !f.balance.IsAllGTE(amounts) {
		f.Mint(parentCtx, f.sender, amounts...)
	}
	ctx := sdk.UnwrapSDKContext(parentCtx).WithEventManager(sdk.NewEventManager()) // discard all faucet related events
	err := f.bankKeeper.SendCoins(ctx, f.sender, receiver, amounts)
	require.NoError(f.t, err)
	f.balance = f.balance.Sub(amounts...)
}

func (f *TestFaucet) NewFundedAccount(ctx context.Context, amounts ...sdk.Coin) sdk.AccAddress {
	_, _, addr := keyPubAddr()
	f.Fund(ctx, addr, amounts...)
	return addr
}

type TestKeepers struct {
	Cdc            codec.Codec
	AccountKeeper  authkeeper.AccountKeeper
	BankKeeper     bankkeeper.Keeper
	OPChildKeeper  opchildkeeper.Keeper
	OracleKeeper   *oraclekeeper.Keeper
	EncodingConfig EncodingConfig
	Faucet         *TestFaucet
}

// createDefaultTestInput common settings for createTestInput
func createDefaultTestInput(t testing.TB) (context.Context, TestKeepers) {
	return createTestInput(t, false)
}

// createTestInput encoders can be nil to accept the defaults, or set it to override some of the message handlers (like default)
func createTestInput(t testing.TB, isCheckTx bool) (context.Context, TestKeepers) {
	return _createTestInput(t, isCheckTx, dbm.NewMemDB())
}

var keyCounter uint64

// we need to make this deterministic (same every test run), as encoded address size and thus gas cost,
// depends on the actual bytes (due to ugly CanonicalAddress encoding)
func keyPubAddr() (cryptotypes.PrivKey, cryptotypes.PubKey, sdk.AccAddress) {
	keyCounter++
	seed := make([]byte, 8)
	binary.BigEndian.PutUint64(seed, keyCounter)

	key := secp256k1.GenPrivKeyFromSecret(seed)
	pub := key.PubKey()
	addr := sdk.AccAddress(pub.Address())
	return key, pub, addr
}

// encoders can be nil to accept the defaults, or set it to override some of the message handlers (like default)
func _createTestInput(
	t testing.TB,
	isCheckTx bool,
	db dbm.DB,
) (context.Context, TestKeepers) {
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, opchildtypes.StoreKey, oracletypes.StoreKey,
	)
	ms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	for _, v := range keys {
		ms.MountStoreWithDB(v, storetypes.StoreTypeIAVL, db)
	}
	memKeys := storetypes.NewMemoryStoreKeys()
	for _, v := range memKeys {
		ms.MountStoreWithDB(v, storetypes.StoreTypeMemory, db)
	}

	require.NoError(t, ms.LoadLatestVersion())

	ctx := sdk.NewContext(ms, tmproto.Header{
		Height: 1234567,
		Time:   time.Date(2020, time.April, 22, 12, 0, 0, 0, time.UTC),
	}, isCheckTx, log.NewNopLogger())

	encodingConfig := MakeEncodingConfig(t)
	appCodec := encodingConfig.Marshaler
	txDecoder := encodingConfig.TxConfig.TxDecoder()
	signModeHandler := encodingConfig.TxConfig.SignModeHandler()

	maccPerms := map[string][]string{ // module account permissions
		authtypes.FeeCollectorName:     nil,
		distributiontypes.ModuleName:   nil,
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		opchildtypes.ModuleName:        {authtypes.Burner, authtypes.Minter},

		// for testing
		authtypes.Minter: {authtypes.Minter, authtypes.Burner},
	}
	accountKeeper := authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]), // target store
		authtypes.ProtoBaseAccount,                          // prototype
		maccPerms,
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
		authtypes.NewModuleAddress(opchildtypes.ModuleName).String(),
	)
	require.NoError(t, accountKeeper.Params.Set(ctx, authtypes.DefaultParams()))
	blockedAddrs := make(map[string]bool)
	for acc := range maccPerms {
		blockedAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	bankKeeper := bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		accountKeeper,
		blockedAddrs,
		authtypes.NewModuleAddress(opchildtypes.ModuleName).String(),
		ctx.Logger(),
	)
	require.NoError(t, bankKeeper.SetParams(ctx, banktypes.DefaultParams()))

	msgRouter := baseapp.NewMsgServiceRouter()
	msgRouter.SetInterfaceRegistry(encodingConfig.InterfaceRegistry)

	// register bank message service to the router
	banktypes.RegisterMsgServer(msgRouter, bankkeeper.NewMsgServerImpl(bankKeeper))

	oracleKeeper := oraclekeeper.NewKeeper(
		runtime.NewKVStoreService(keys[oracletypes.StoreKey]),
		appCodec,
		nil,
		authtypes.NewModuleAddress(opchildtypes.ModuleName),
	)

	opchildKeeper := opchildkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[opchildtypes.StoreKey]),
		accountKeeper,
		bankKeeper,
		&oracleKeeper,
		sdk.ChainAnteDecorators(
			authante.NewSetPubKeyDecorator(accountKeeper),
			authante.NewValidateSigCountDecorator(accountKeeper),
			authante.NewSigGasConsumeDecorator(accountKeeper, authante.DefaultSigVerificationGasConsumer),
			authante.NewSigVerificationDecorator(accountKeeper, signModeHandler),
			authante.NewIncrementSequenceDecorator(accountKeeper),
		),
		txDecoder,
		msgRouter,
		authtypes.NewModuleAddress(opchildtypes.ModuleName).String(),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
		ctx.Logger(),
	)

	opchildParams := opchildtypes.DefaultParams()
	opchildParams.Admin = addrs[0].String()
	opchildParams.BridgeExecutors = []string{addrs[0].String()}
	require.NoError(t, opchildKeeper.SetParams(ctx, opchildParams))

	// register handlers to msg router
	opchildtypes.RegisterMsgServer(msgRouter, opchildkeeper.NewMsgServerImpl(opchildKeeper))

	faucet := NewTestFaucet(t, ctx, bankKeeper, authtypes.Minter, initialTotalSupply()...)

	keepers := TestKeepers{
		Cdc:            appCodec,
		AccountKeeper:  accountKeeper,
		BankKeeper:     bankKeeper,
		OPChildKeeper:  *opchildKeeper,
		OracleKeeper:   &oracleKeeper,
		EncodingConfig: encodingConfig,
		Faucet:         faucet,
	}
	return ctx, keepers
}

func generateTestTx(
	t *testing.T, input TestKeepers, msgs []sdk.Msg,
	privs []cryptotypes.PrivKey, accNums []uint64,
	accSeqs []uint64, chainID string,
) authsign.Tx {
	txConfig := input.EncodingConfig.TxConfig
	txBuilder := txConfig.NewTxBuilder()

	defaultSignMode, err := authsign.APISignModeToInternal(txConfig.SignModeHandler().DefaultMode())
	require.NoError(t, err)

	// set msgs
	txBuilder.SetMsgs(msgs...)

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	var sigsV2 []signing.SignatureV2
	for i, priv := range privs {
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  defaultSignMode,
				Signature: nil,
			},
			Sequence: accSeqs[i],
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err = txBuilder.SetSignatures(sigsV2...)
	require.NoError(t, err)

	// Second round: all signer infos are set, so each signer can sign.
	sigsV2 = []signing.SignatureV2{}
	for i, priv := range privs {
		signerData := authsign.SignerData{
			Address:       sdk.AccAddress(priv.PubKey().Address()).String(),
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
			PubKey:        priv.PubKey(),
		}
		sigV2, err := tx.SignWithPrivKey(
			context.TODO(), defaultSignMode, signerData,
			txBuilder, priv, txConfig, accSeqs[i])
		require.NoError(t, err)

		sigsV2 = append(sigsV2, sigV2)
	}
	err = txBuilder.SetSignatures(sigsV2...)
	require.NoError(t, err)

	return txBuilder.GetTx()
}
