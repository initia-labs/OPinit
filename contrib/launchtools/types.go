package launchtools

import (
	"context"
	"cosmossdk.io/log"
	"encoding/json"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	"github.com/initia-labs/OPinit/contrib/launchtools/utils"
	opchildtypes "github.com/initia-labs/OPinit/x/opchild/types"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
	"path"
	"sync"
)

const (
	// OutDir is the directory where the output files are written.
	// All files written are placed under $HOME/{OutDir}
	OutDir = "out"
)

// LauncherStepFuncFactory is a factory function that creates a step function.
type LauncherStepFuncFactory[Manifest any] func(Manifest) LauncherStepFunc

// LauncherStepFunc is a function that takes a launcher and returns an error.
type LauncherStepFunc func(ctx Launcher) error

// LauncherCleanupFunc is a function that cleans up the launcher.
type LauncherCleanupFunc func() error

// ExpectedApp is an extended interface that allows to get the ibc keeper from the app.
type ExpectedApp interface {
	servertypes.Application

	// GetIBCKeeper returns the ibc keeper from the app.
	GetIBCKeeper() *ibckeeper.Keeper
}

// Launcher is an interface that provides the necessary methods to interact with the launcher.
// It is used to abstract away the underlying contexts, and provide a non-intrusive way to interact with the launcher.
type Launcher interface {
	// IsAppInitialized returns true if the app is initialized.
	IsAppInitialized() bool

	// SetApp sets the app to the launcher.
	SetApp(app servertypes.Application) error

	// App returns the app. Must only be called after SetApp is called.
	App() ExpectedApp

	// Logger returns the logger.
	Logger() log.Logger

	// AppCreator returns the app creator.
	AppCreator() servertypes.AppCreator

	// QueryContext returns the query context.
	// Note: Use this instead of Launcher.Context() for any query operations.
	QueryContext() context.Context

	// Context returns the context (from the cobra command).
	Context() context.Context

	// ClientContext returns the client context.
	// Usually used as a carrier for the codec, and whatever global sdk configuration.
	ClientContext() *client.Context

	// ServerContext returns the server context.
	// Usually used as a carrier for the server configuration (cometbft, main() context, viper, ...).
	ServerContext() *server.Context

	// DefaultGenesis returns the default genesis.
	DefaultGenesis() map[string]json.RawMessage

	// SetRPCHelpers sets the rpc helpers.
	SetRPCHelpers(rpcHelperL1, rpcHelperL2 *utils.RPCHelper)
	GetRPCHelperL1() *utils.RPCHelper
	GetRPCHelperL2() *utils.RPCHelper

	// WriteToFile writes data to a file under $HOME/out.
	WriteToFile(filename string, data string) error
}

var _ Launcher = &LauncherContext{}

type LauncherContext struct {
	mtx *sync.Mutex

	log            log.Logger
	app            ExpectedApp
	cleanupFns     []LauncherCleanupFunc
	defaultGenesis map[string]json.RawMessage

	appCreator servertypes.AppCreator
	clientCtx  *client.Context
	serverCtx  *server.Context

	cmd *cobra.Command

	rpcHelperL1 *utils.RPCHelper
	rpcHelperL2 *utils.RPCHelper
}

func NewLauncher(
	cmd *cobra.Command,
	clientCtx *client.Context,
	serverCtx *server.Context,
	appCreator servertypes.AppCreator,
	defaultGenesis map[string]json.RawMessage,
) *LauncherContext {
	kr, err := keyring.New("minitia", keyring.BackendTest, clientCtx.HomeDir, nil, clientCtx.Codec)
	if err != nil {
		panic("failed to create keyring")
	}

	// create a new client context with the keyring
	nextClientCtx := clientCtx.WithKeyring(kr)

	// mute log output
	serverCtx.Logger = log.NewNopLogger()

	// make sure to register both ophost and opchild proto
	// otherwise it fails on creating op bridge on L1
	ophosttypes.RegisterInterfaces(clientCtx.InterfaceRegistry)
	opchildtypes.RegisterInterfaces(clientCtx.InterfaceRegistry)

	if err := os.MkdirAll(path.Join(nextClientCtx.HomeDir, OutDir), os.ModePerm); err != nil {
		panic("failed to create out directory")
	}

	return &LauncherContext{
		log:            log.NewLogger(os.Stderr),
		mtx:            new(sync.Mutex),
		clientCtx:      &nextClientCtx,
		serverCtx:      serverCtx,
		appCreator:     appCreator,
		cmd:            cmd,
		defaultGenesis: defaultGenesis,
	}
}

func (l *LauncherContext) IsAppInitialized() bool {
	return l.app != nil
}

func (l *LauncherContext) SetApp(app servertypes.Application) error {
	nextApp, ok := app.(ExpectedApp)
	if !ok {
		return errors.New("supplied app does not implement expected methods")
	}

	l.app = nextApp
	return nil
}

func (l *LauncherContext) App() ExpectedApp {
	return l.app
}

func (l *LauncherContext) AppCreator() servertypes.AppCreator {
	return l.appCreator
}

func (l *LauncherContext) Logger() log.Logger {
	return l.log
}

func (l *LauncherContext) ClientContext() *client.Context {
	return l.clientCtx
}

func (l *LauncherContext) ServerContext() *server.Context {
	return l.serverCtx
}

func (l *LauncherContext) QueryContext() context.Context {
	return sdk.NewContext(
		l.App().CommitMultiStore(),
		cmtproto.Header{},
		false,
		log.NewNopLogger(),
	)
}

func (l *LauncherContext) Context() context.Context {
	return l.cmd.Context()
}

func (l *LauncherContext) DefaultGenesis() map[string]json.RawMessage {
	return l.defaultGenesis
}

func (l *LauncherContext) SetRPCHelpers(rpcHelperL1, rpcHelperL2 *utils.RPCHelper) {
	l.rpcHelperL1 = rpcHelperL1
	l.rpcHelperL2 = rpcHelperL2
}

func (l *LauncherContext) GetRPCHelperL1() *utils.RPCHelper {
	return l.rpcHelperL1
}

func (l *LauncherContext) GetRPCHelperL2() *utils.RPCHelper {
	return l.rpcHelperL2
}

func (l *LauncherContext) WriteToFile(filename string, data string) error {
	file, err := os.Create(path.Join(l.clientCtx.HomeDir, OutDir, filename))
	if err != nil {
		return errors.Wrap(err, "failed to create file")
	}
	defer file.Close()

	if _, err := file.Write([]byte(data)); err != nil {
		return errors.Wrap(err, "failed to write data to file")
	}

	return nil
}
