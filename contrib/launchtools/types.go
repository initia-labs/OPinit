package launchtools

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"sync"

	"cosmossdk.io/log"
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
	"golang.org/x/sync/errgroup"
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

// ExpectedApp is an extended interface that allows to get the ibc keeper from the app.
type ExpectedApp interface {
	servertypes.Application

	// GetIBCKeeper returns the ibc keeper from the app.
	GetIBCKeeper() *ibckeeper.Keeper
}

type AppCreator interface {
	AppCreator() servertypes.AppCreator
	App() servertypes.Application
}

type Relayer interface {
	UpdateClients() error
}

// Launcher is an interface that provides the necessary methods to interact with the launcher.
// It is used to abstract away the underlying contexts, and provide a non-intrusive way to interact with the launcher.
type Launcher interface {
	// IsAppInitialized returns true if the app is initialized.
	IsAppInitialized() bool

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

	// SetErrorGroup sets the error group.
	SetErrorGroup(g *errgroup.Group)
	GetErrorGroup() *errgroup.Group

	// SetBridgeId sets the bridge id.
	SetBridgeId(id uint64)
	GetBridgeId() *uint64

	// SetRelayer sets the relayer.
	SetRelayer(relayer Relayer)
	GetRelayer() Relayer

	// WriteOutput writes data to internal artifacts buffer.
	WriteOutput(name string, data string) error

	// FinalizeOutput returns the output data in JSON.
	FinalizeOutput(config *Config) (string, error)
}

var _ Launcher = &LauncherContext{}

type LauncherContext struct {
	mtx *sync.Mutex

	log            log.Logger
	defaultGenesis map[string]json.RawMessage

	appCreator AppCreator
	clientCtx  *client.Context
	serverCtx  *server.Context

	cmd *cobra.Command

	rpcHelperL1 *utils.RPCHelper
	rpcHelperL2 *utils.RPCHelper

	// errorgroup is used to manage the lifecycle of the app.
	errorgroup *errgroup.Group

	// artifacts is a map of artifacts that are created during the launch process.
	artifactsDir string
	artifacts    map[string]string

	bridgeId *uint64
	relayer  Relayer
}

func NewLauncher(
	cmd *cobra.Command,
	clientCtx *client.Context,
	serverCtx *server.Context,
	appCreator AppCreator,
	defaultGenesis map[string]json.RawMessage,
	artifactsDir string,
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

	// prepare artifacts output
	artifactsDirFQ := path.Join(nextClientCtx.HomeDir, artifactsDir)
	if err := os.MkdirAll(artifactsDirFQ, os.ModePerm); err != nil {
		panic("failed to create artifacts directory")
	}

	return &LauncherContext{
		log:            log.NewLogger(os.Stderr),
		mtx:            new(sync.Mutex),
		clientCtx:      &nextClientCtx,
		serverCtx:      serverCtx,
		appCreator:     appCreator,
		cmd:            cmd,
		defaultGenesis: defaultGenesis,
		artifactsDir:   artifactsDirFQ,
		artifacts:      map[string]string{},
	}
}

func (l *LauncherContext) IsAppInitialized() bool {
	return l.appCreator.App() != nil
}

func (l *LauncherContext) App() ExpectedApp {
	return l.appCreator.App().(ExpectedApp)
}

func (l *LauncherContext) AppCreator() servertypes.AppCreator {
	return l.appCreator.AppCreator()
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

func (l *LauncherContext) SetBridgeId(id uint64) {
	l.bridgeId = &id
}

func (l *LauncherContext) GetBridgeId() *uint64 {
	return l.bridgeId
}

func (l *LauncherContext) SetRelayer(relayer Relayer) {
	l.relayer = relayer
}

func (l *LauncherContext) GetRelayer() Relayer {
	return l.relayer
}

func (l *LauncherContext) WriteOutput(filename string, data string) error {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	l.artifacts[filename] = data

	return nil
}

func (l *LauncherContext) FinalizeOutput(config *Config) (string, error) {
	// write the artifacts to a file
	bz, err := json.MarshalIndent(l.artifacts, "", " ")
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal artifacts")
	}

	if err := os.WriteFile(path.Join(l.artifactsDir, "artifacts.json"), bz, os.ModePerm); err != nil {
		return "", errors.Wrap(err, "failed to write artifacts to file")
	}

	// write the config to a file
	configBz, err := json.MarshalIndent(config, "", " ")
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal config")
	}

	if err := os.WriteFile(path.Join(l.artifactsDir, "config.json"), configBz, os.ModePerm); err != nil {
		return "", errors.Wrap(err, "failed to write config to file")
	}

	return string(bz), nil
}

func (l *LauncherContext) SetErrorGroup(g *errgroup.Group) {
	l.errorgroup = g
}

func (l *LauncherContext) GetErrorGroup() *errgroup.Group {
	return l.errorgroup
}
