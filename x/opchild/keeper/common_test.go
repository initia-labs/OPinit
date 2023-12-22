package keeper_test

import (
	"context"
	"encoding/binary"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	errors "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/tx/signing"
	dbm "github.com/cosmos/cosmos-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codecaddress "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"

	opchild "github.com/initia-labs/OPinit/x/opchild"
	opchildkeeper "github.com/initia-labs/OPinit/x/opchild/keeper"
	opchildtypes "github.com/initia-labs/OPinit/x/opchild/types"
)

const baseDenom = "umin"

var ModuleBasics = module.NewBasicManager(
	auth.AppModuleBasic{},
	bank.AppModuleBasic{},
	opchild.AppModuleBasic{},
)

var (
	pubKeys = []crypto.PubKey{
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

	valAddrs = []sdk.ValAddress{
		sdk.ValAddress(pubKeys[0].Address()),
		sdk.ValAddress(pubKeys[1].Address()),
		sdk.ValAddress(pubKeys[2].Address()),
		sdk.ValAddress(pubKeys[3].Address()),
		sdk.ValAddress(pubKeys[4].Address()),
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
		SigningOptions: signing.Options{
			AddressCodec:          codecaddress.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
			ValidatorAddressCodec: codecaddress.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		},
	})
	appCodec := codec.NewProtoCodec(interfaceRegistry)
	legacyAmino := codec.NewLegacyAmino()
	txConfig := tx.NewTxConfig(appCodec, tx.DefaultSignModes)

	std.RegisterInterfaces(interfaceRegistry)
	std.RegisterLegacyAminoCodec(legacyAmino)

	ModuleBasics.RegisterLegacyAminoCodec(legacyAmino)
	ModuleBasics.RegisterInterfaces(interfaceRegistry)

	interfaceRegistry.RegisterImplementations(
		(*govv1beta1.Content)(nil),
		&testLegacyContent{},
	)

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         appCodec,
		TxConfig:          txConfig,
		Amino:             legacyAmino,
	}
}

func initialTotalSupply() sdk.Coins {
	faucetBalance := sdk.NewCoins(sdk.NewCoin(baseDenom, initiaSupply))
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
	AccountKeeper  authkeeper.AccountKeeper
	BankKeeper     bankkeeper.Keeper
	OPChildKeeper  opchildkeeper.Keeper
	BridgeHook     *bridgeHook
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
func keyPubAddr() (crypto.PrivKey, crypto.PubKey, sdk.AccAddress) {
	keyCounter++
	seed := make([]byte, 8)
	binary.BigEndian.PutUint64(seed, keyCounter)

	key := ed25519.GenPrivKeyFromSecret(seed)
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
		authtypes.StoreKey, banktypes.StoreKey, opchildtypes.StoreKey,
	)
	ms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	for _, v := range keys {
		ms.MountStoreWithDB(v, storetypes.StoreTypeIAVL, db)
	}
	memKeys := storetypes.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)
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
	bankKeeper.SetParams(ctx, banktypes.DefaultParams())

	msgRouter := baseapp.NewMsgServiceRouter()
	msgRouter.SetInterfaceRegistry(encodingConfig.InterfaceRegistry)

	bridgeHook := &bridgeHook{}
	opchildKeeper := opchildkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[opchildtypes.StoreKey]),
		accountKeeper,
		bankKeeper,
		bridgeHook.Hook,
		msgRouter,
		authtypes.NewModuleAddress(opchildtypes.ModuleName).String(),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	)

	// set test legacy content handler
	govRouter := govv1beta1.NewRouter()
	govRouter.AddRoute("test", NewTestLegacyContentHandler())
	opchildKeeper.SetLegacyRouter(govRouter)

	opchildParams := opchildtypes.DefaultParams()
	opchildParams.BridgeExecutor = addrs[0].String()
	opchildKeeper.SetParams(ctx, opchildParams)

	// register handlers to msg router
	opchildtypes.RegisterMsgServer(msgRouter, opchildkeeper.NewMsgServerImpl(*opchildKeeper))

	faucet := NewTestFaucet(t, ctx, bankKeeper, authtypes.Minter, initialTotalSupply()...)

	keepers := TestKeepers{
		AccountKeeper:  accountKeeper,
		BankKeeper:     bankKeeper,
		OPChildKeeper:  *opchildKeeper,
		BridgeHook:     bridgeHook,
		EncodingConfig: encodingConfig,
		Faucet:         faucet,
	}
	return ctx, keepers
}

type testLegacyContent struct {
	Title       string
	Description string
	Message     string
}

func (*testLegacyContent) ProtoMessage() {}
func (m *testLegacyContent) Reset()      { *m = testLegacyContent{} }
func (c *testLegacyContent) GetTitle() string {
	return c.Title
}
func (c *testLegacyContent) GetDescription() string {
	return c.Description
}

func (c *testLegacyContent) ProposalRoute() string {
	return "test"
}
func (c *testLegacyContent) ProposalType() string {
	return "test"
}
func (c *testLegacyContent) ValidateBasic() error {
	return nil
}
func (c testLegacyContent) String() string {
	return fmt.Sprintf(`Parameter Change Proposal:
Title:       %s
Description: %s
Message:     %s
`, c.Title, c.Description, c.Message)
}

func NewTestLegacyContentHandler() govv1beta1.Handler {
	return func(ctx sdk.Context, content govv1beta1.Content) error {
		switch c := content.(type) {
		case *testLegacyContent:
			if c.Message != "test" {
				return fmt.Errorf("message must be `test`")
			}
			return nil

		default:
			return errors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized legacy content type: %T", c)
		}
	}
}

type bridgeHook struct {
	msgBytes []byte
	err      error
}

func (h *bridgeHook) Hook(ctx context.Context, sender sdk.AccAddress, msgBytes []byte) error {
	if h.err == nil {
		h.msgBytes = msgBytes
	}

	return h.err
}
