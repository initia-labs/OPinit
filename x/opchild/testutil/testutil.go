package testutil

import (
	"context"
	"encoding/binary"
	"fmt"
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
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codecaddress "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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

	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"

	"github.com/initia-labs/OPinit/x/opchild"
	opchildkeeper "github.com/initia-labs/OPinit/x/opchild/keeper"
	opchildtypes "github.com/initia-labs/OPinit/x/opchild/types"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
	oraclekeeper "github.com/skip-mev/connect/v2/x/oracle/keeper"
	oracletypes "github.com/skip-mev/connect/v2/x/oracle/types"
)

var (
	// ModuleBasics is the basic manager for the opchild module
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		opchild.AppModuleBasic{},
	)

	// TestDenoms are test token denominations used across tests
	TestDenoms = []string{
		"test1",
		"test2",
		"test3",
		"test4",
		"test5",
	}

	initiaSupply = math.NewInt(100_000_000_000)

	PubKeys = GenPubKeys(5)

	Addrs = []sdk.AccAddress{
		sdk.AccAddress(PubKeys[0].Address()),
		sdk.AccAddress(PubKeys[1].Address()),
		sdk.AccAddress(PubKeys[2].Address()),
		sdk.AccAddress(PubKeys[3].Address()),
		sdk.AccAddress(PubKeys[4].Address()),
	}

	AddrsStr = []string{
		Addrs[0].String(),
		Addrs[1].String(),
		Addrs[2].String(),
		Addrs[3].String(),
		Addrs[4].String(),
	}

	ValAddrs = []sdk.ValAddress{
		sdk.ValAddress(PubKeys[0].Address()),
		sdk.ValAddress(PubKeys[1].Address()),
		sdk.ValAddress(PubKeys[2].Address()),
		sdk.ValAddress(PubKeys[3].Address()),
		sdk.ValAddress(PubKeys[4].Address()),
	}

	ValAddrsStr = []string{
		ValAddrs[0].String(),
		ValAddrs[1].String(),
		ValAddrs[2].String(),
		ValAddrs[3].String(),
		ValAddrs[4].String(),
	}
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
	_, _, addr := KeyPubAddr()
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
	_, _, addr := KeyPubAddr()
	f.Fund(ctx, addr, amounts...)
	return addr
}

type TestKeepers struct {
	Cdc                  codec.Codec
	AccountKeeper        authkeeper.AccountKeeper
	BankKeeper           bankkeeper.Keeper
	OPChildKeeper        opchildkeeper.Keeper
	OracleKeeper         *oraclekeeper.Keeper
	ClientKeeper         *MockIBCClientKeeper
	EncodingConfig       EncodingConfig
	Faucet               *TestFaucet
	TokenCreationFactory *TestTokenCreationFactory
	MockRouter           *MockRouter
}

func CreateTestInput(t testing.TB, isCheckTx bool) (context.Context, TestKeepers) {
	return createTestInput(t, isCheckTx, dbm.NewMemDB())
}

var keyCounter uint64

// KeyPubAddr generates deterministic keys for testing
func KeyPubAddr() (cryptotypes.PrivKey, cryptotypes.PubKey, sdk.AccAddress) {
	keyCounter++
	seed := make([]byte, 8)
	binary.BigEndian.PutUint64(seed, keyCounter)

	key := secp256k1.GenPrivKeyFromSecret(seed)
	pub := key.PubKey()
	addr := sdk.AccAddress(pub.Address())
	return key, pub, addr
}

func initialTotalSupply() sdk.Coins {
	faucetBalance := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initiaSupply))
	for _, testDenom := range TestDenoms {
		faucetBalance = faucetBalance.Add(sdk.NewCoin(testDenom, initiaSupply))
	}
	return faucetBalance
}

func createTestInput(
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

	originMessageRouter := baseapp.NewMsgServiceRouter()
	originMessageRouter.SetInterfaceRegistry(encodingConfig.InterfaceRegistry)
	mockRouter := &MockRouter{
		originMessageRouter: originMessageRouter,
	}

	oracleKeeper := oraclekeeper.NewKeeper(
		runtime.NewKVStoreService(keys[oracletypes.StoreKey]),
		appCodec,
		nil,
		authtypes.NewModuleAddress(opchildtypes.ModuleName),
	)

	// Use first address from shared test addresses for admin/executor
	firstAddr := Addrs[0]

	tokenCreationFactory := &TestTokenCreationFactory{Created: make(map[string]bool)}
	opchildKeeper := opchildkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[opchildtypes.StoreKey]),
		&accountKeeper,
		bankKeeper,
		&oracleKeeper,
		sdk.ChainAnteDecorators(
			authante.NewValidateBasicDecorator(),
			authante.NewSetPubKeyDecorator(accountKeeper),
			authante.NewValidateSigCountDecorator(accountKeeper),
			authante.NewSigGasConsumeDecorator(accountKeeper, authante.DefaultSigVerificationGasConsumer),
			authante.NewSigVerificationDecorator(accountKeeper, signModeHandler),
			authante.NewIncrementSequenceDecorator(accountKeeper),
		),
		txDecoder,
		mockRouter,
		authtypes.NewModuleAddress(opchildtypes.ModuleName).String(),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
		authcodec.NewBech32Codec("init"),
		ctx.Logger(),
	).WithTokenCreationFn(tokenCreationFactory.TokenCreationFn)

	opchildParams := opchildtypes.DefaultParams()
	opchildParams.Admin = firstAddr.String()
	opchildParams.BridgeExecutors = []string{firstAddr.String()}
	require.NoError(t, opchildKeeper.SetParams(ctx, opchildParams))

	// register handlers to msg router
	banktypes.RegisterMsgServer(originMessageRouter, bankkeeper.NewMsgServerImpl(bankKeeper))
	opchildtypes.RegisterMsgServer(originMessageRouter, opchildkeeper.NewMsgServerImpl(opchildKeeper))

	faucet := NewTestFaucet(t, ctx, bankKeeper, authtypes.Minter, initialTotalSupply()...)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	consensusParams := tmproto.ConsensusParams{
		Validator: &tmproto.ValidatorParams{
			PubKeyTypes: []string{"ed25519", "secp256k1"},
		},
	}
	ctx = sdkCtx.WithConsensusParams(consensusParams)

	// Set IBC keepers for opchild keeper
	ibcClientKeeper := NewMockIBCClientKeeper()
	portKeeper := &MockIBCPortKeeper{}
	scopedKeeper := &MockIBCScopedKeeper{}
	if err := opchildKeeper.SetIBCKeepers(ibcClientKeeper, portKeeper, scopedKeeper); err != nil {
		panic(err)
	}

	keepers := TestKeepers{
		Cdc:                  appCodec,
		AccountKeeper:        accountKeeper,
		BankKeeper:           bankKeeper,
		OPChildKeeper:        *opchildKeeper,
		ClientKeeper:         ibcClientKeeper,
		OracleKeeper:         &oracleKeeper,
		EncodingConfig:       encodingConfig,
		Faucet:               faucet,
		TokenCreationFactory: tokenCreationFactory,
		MockRouter:           mockRouter,
	}
	return ctx, keepers
}

func GenerateTestTx(
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

type TestTokenCreationFactory struct {
	Created map[string]bool
}

func (t *TestTokenCreationFactory) TokenCreationFn(ctx context.Context, denom string, decimals uint8) error {
	t.Created[denom] = true
	return nil
}

type MockRouter struct {
	originMessageRouter baseapp.MessageRouter
	handledMsgs         []*transfertypes.MsgTransfer
	shouldFail          bool
}

func (router *MockRouter) Handler(msg sdk.Msg) baseapp.MsgServiceHandler {
	return router.HandlerByTypeURL(sdk.MsgTypeURL(msg))
}

func (router *MockRouter) HandlerByTypeURL(typeURL string) baseapp.MsgServiceHandler {
	switch typeURL {
	case sdk.MsgTypeURL(&transfertypes.MsgTransfer{}):
		return func(ctx sdk.Context, _msg sdk.Msg) (*sdk.Result, error) {
			if router.shouldFail {
				return nil, sdkerrors.ErrInvalidRequest
			}

			msg := _msg.(*transfertypes.MsgTransfer)

			// Store the handled message for verification
			router.handledMsgs = append(router.handledMsgs, msg)

			// Emit an event to track the transfer
			ctx.EventManager().EmitEvent(sdk.NewEvent("ibc_transfer",
				sdk.NewAttribute("sender", msg.Sender),
				sdk.NewAttribute("receiver", msg.Receiver),
				sdk.NewAttribute("token", msg.Token.String()),
				sdk.NewAttribute("source_port", msg.SourcePort),
				sdk.NewAttribute("source_channel", msg.SourceChannel),
				sdk.NewAttribute("timeout_height", msg.TimeoutHeight.String()),
				sdk.NewAttribute("timeout_timestamp", fmt.Sprintf("%d", msg.TimeoutTimestamp)),
				sdk.NewAttribute("memo", msg.Memo),
			))

			return &sdk.Result{}, nil
		}
	}

	return router.originMessageRouter.HandlerByTypeURL(typeURL)
}

func (router *MockRouter) GetHandledMsgs() []*transfertypes.MsgTransfer {
	return router.handledMsgs
}

func (router *MockRouter) Reset() {
	router.handledMsgs = nil
}

func (router *MockRouter) SetShouldFail(shouldFail bool) {
	router.shouldFail = shouldFail
}

type MockIBCClientKeeper struct {
	stores    map[string]storetypes.KVStore
	states    map[string]exported.ClientState
	consensus map[string]exported.ConsensusState
}

func NewMockIBCClientKeeper() *MockIBCClientKeeper {
	return &MockIBCClientKeeper{
		stores:    make(map[string]storetypes.KVStore),
		states:    make(map[string]exported.ClientState),
		consensus: make(map[string]exported.ConsensusState),
	}
}

func (k *MockIBCClientKeeper) SetClientState(clientID string, cs exported.ClientState) {
	k.states[clientID] = cs
}

func (k *MockIBCClientKeeper) SetConsensusState(clientID string, cs exported.ConsensusState) {
	k.consensus[clientID] = cs
}

func (k *MockIBCClientKeeper) GetClientState(ctx sdk.Context, clientID string) (exported.ClientState, bool) {
	cs, ok := k.states[clientID]
	return cs, ok
}

func (k *MockIBCClientKeeper) SetClientStore(key string, store storetypes.KVStore) {
	k.stores[key] = store
}

func (k *MockIBCClientKeeper) ClientStore(ctx sdk.Context, clientID string) storetypes.KVStore {
	kvStore, ok := k.stores[clientID]
	if !ok {
		panic(fmt.Sprintf("client %s not found in store", clientID))
	}
	return kvStore
}

type MockIBCPortKeeper struct{}

func (k *MockIBCPortKeeper) BindPort(ctx sdk.Context, portID string) *capabilitytypes.Capability {
	return &capabilitytypes.Capability{}
}

type MockIBCScopedKeeper struct{}

func (k *MockIBCScopedKeeper) GetCapability(ctx sdk.Context, name string) (*capabilitytypes.Capability, bool) {
	return &capabilitytypes.Capability{}, true
}

func (k *MockIBCScopedKeeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return nil
}

func GenPubKeys(n int) []cryptotypes.PubKey {
	pubKeys := make([]cryptotypes.PubKey, n)
	for i := 0; i < n; i++ {
		pubKeys[i] = secp256k1.GenPrivKey().PubKey()
	}
	return pubKeys
}

func CreateAttestor(t *testing.T, operatorAddr string, pubKey cryptotypes.PubKey, moniker string) ophosttypes.Attestor {
	pkAny, err := codectypes.NewAnyWithValue(pubKey)
	require.NoError(t, err)

	return ophosttypes.Attestor{
		OperatorAddress: operatorAddr,
		ConsensusPubkey: pkAny,
		Moniker:         moniker,
	}
}
