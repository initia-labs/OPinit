package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"reflect"

	"cosmossdk.io/log"
	relayercmd "github.com/cosmos/relayer/v2/cmd"
	relayertypes "github.com/cosmos/relayer/v2/relayer"
	relayerconfig "github.com/cosmos/relayer/v2/relayer/chains/cosmos"
	"github.com/initia-labs/OPinit/contrib/launchtools"
	"github.com/pkg/errors"
)

// EstablishIBCChannelsWithNFTTransfer creates a new IBC channel for fungible transfer, and one with NFT transfer
// Note that srcPort, dstPort, channelVersion must be specified by the caller.
// - For minimove, this is usually fixed to "nft-transfer", "nft-transfer", "ics721-1".
// - For miniwasm, this depends on the contract addresses; the caller must specify the correct values.
// - getPorts needs to be a callback because the address is derived correctly only after sdkConfig is set
func EstablishIBCChannelsWithNFTTransfer(getPorts func() (srcPort, dstPort, channelVersion string)) launchtools.LauncherStepFuncFactory[*launchtools.Config] {
	return func(config *launchtools.Config) launchtools.LauncherStepFunc {
		src, dst, cv := getPorts()
		return establishIBCChannels(config, src, dst, cv)
	}
}

func establishIBCChannels(
	config *launchtools.Config,
	srcPort,
	dstPort,
	channelVersion string,
) launchtools.LauncherStepFunc {
	relayerPath, err := os.MkdirTemp("", RelayerPathTemp)
	if err != nil {
		panic(err)
	}

	runLifecycle := lifecycle(
		initializeConfig,
		initializeChains(config, relayerPath),
		initializePaths(config, relayerPath),
		initializeRelayerKeyring(config),

		// create default transfer ports
		link,

		// create nft-transfer ports as well
		linkWithPorts(srcPort, dstPort, channelVersion),
	)

	return func(ctx launchtools.Launcher) error {
		if !ctx.IsAppInitialized() {
			return errors.New("app is not initialized")
		}

		relayer := NewRelayer(ctx.Context(), relayerPath, ctx.Logger())
		ctx.SetRelayer(relayer)

		return runLifecycle(relayer)
	}
}

// -------------------------------
func initializeConfig(r *Relayer) error {
	return r.run([]string{"config", "init"})
}

// initializeChains creates chain configuration files and initializes chains for the relayer
// "chains" in cosmos/relayer lingo means srcChain and dstChain. Speficic ports are not created here.
// see initializePaths.
func initializeChains(config *launchtools.Config, basePath string) func(*Relayer) error {
	// ChainConfig is a struct that represents the configuration of a chain
	// cosmos/relayer specific
	type ChainConfig struct {
		Type  string                             `json:"type"`
		Value relayerconfig.CosmosProviderConfig `json:"value"`
	}

	var chainConfigs = [2]ChainConfig{
		{
			Type: "cosmos",
			Value: relayerconfig.CosmosProviderConfig{
				Key:            RelayerKeyName,
				ChainID:        config.L1Config.ChainID,
				RPCAddr:        config.L1Config.RPC_URL,
				AccountPrefix:  "init",
				KeyringBackend: KeyringBackend,
				GasAdjustment:  1.5,
				GasPrices:      config.L1Config.GasPrices,
				Debug:          true,
				Timeout:        "160s",
				OutputFormat:   "json",
			},
		},
		{
			Type: "cosmos",
			Value: relayerconfig.CosmosProviderConfig{
				Key:            RelayerKeyName,
				ChainID:        config.L2Config.ChainID,
				RPCAddr:        "http://localhost:26657",
				AccountPrefix:  "init",
				KeyringBackend: KeyringBackend,
				GasAdjustment:  1.5,
				GasPrices:      "", // gas prices required for l2 txs
				Debug:          true,
				Timeout:        "160s",
				OutputFormat:   "json",
			},
		},
	}

	// write chain configs to files
	for i, chainConfig := range chainConfigs {
		bz, err := json.MarshalIndent(chainConfig, "", " ")
		if err != nil {
			panic(errors.New("failed to create chain config"))
		}

		pathName := fmt.Sprintf("chain%d", i)
		fileName := fmt.Sprintf("%s/%s.json", basePath, pathName)

		if err := os.WriteFile(fileName, bz, 0644); err != nil {
			panic(errors.New("failed to write chain config"))
		}
	}

	return func(r *Relayer) error {
		r.logger.Info("initializing chains for relayer...",
			"chains-len", len(chainConfigs),
			"chain-0", chainConfigs[0].Value.ChainID,
			"chain-1", chainConfigs[1].Value.ChainID,
		)

		for i, chainConfig := range chainConfigs {
			if err := r.run([]string{
				"chains",
				"add",
				"--file",
				path.Join(basePath, fmt.Sprintf("chain%d.json", i)),
				chainConfig.Value.ChainID,
			}); err != nil {
				return err
			}
		}

		return nil
	}
}

// initializePaths creates a path configuration file and initializes paths for the relayer
// Paths are nothing more than a pair of chains that are connected by a channel
func initializePaths(config *launchtools.Config, basePath string) func(*Relayer) error {
	pathConfig := relayertypes.Path{
		Src: &relayertypes.PathEnd{
			ChainID: config.L2Config.ChainID,
		},
		Dst: &relayertypes.PathEnd{
			ChainID: config.L1Config.ChainID,
		},
		Filter: relayertypes.ChannelFilter{
			Rule:        "",
			ChannelList: nil,
		},
	}
	pathConfigJSON, err := json.Marshal(pathConfig)
	if err != nil {
		panic(errors.New("failed to create path config"))
	}

	if err := os.WriteFile(fmt.Sprintf("%s/paths.json", basePath), pathConfigJSON, 0644); err != nil {
		panic(errors.New("failed to write path config"))
	}

	return func(r *Relayer) error {
		r.logger.Info("initializing paths for relayer...",
			"src-chain", pathConfig.Src.ChainID,
			"dst-chain", pathConfig.Dst.ChainID,
		)

		return r.run([]string{
			"paths",
			"add",
			config.L2Config.ChainID,
			config.L1Config.ChainID,
			RelayerPathName,
			"-f",
			fmt.Sprintf("%s/paths.json", basePath),
		})
	}
}

// initializeRelayerKeyring initializes the keyring for the relayer
// cosmos/relayer uses its own keyring to manage keys. for this, we need to restore the relayer key
func initializeRelayerKeyring(config *launchtools.Config) func(*Relayer) error {
	relayerKeyFromInput := reflect.ValueOf(*config.SystemKeys).FieldByName(RelayerKeyName)
	if !relayerKeyFromInput.IsValid() {
		panic(errors.New("relayer key not found in config"))
	}

	relayerKey := relayerKeyFromInput.Interface().(*launchtools.Account)
	return func(r *Relayer) error {
		r.logger.Info("initializing keyring for relayer...",
			"key-name", RelayerKeyName,
		)

		for _, chainName := range []string{
			config.L2Config.ChainID,
			config.L1Config.ChainID,
		} {
			if err := r.run([]string{
				"keys",
				"restore",
				chainName,
				RelayerKeyName,
				relayerKey.Mnemonic,
			}); err != nil {
				return err
			}
		}

		return nil
	}
}

// link creates a default transfer channel between the chains
// it does all the heavy lifting of creating the channel, connection, and client
func link(r *Relayer) error {
	r.logger.Info("linking chains for relayer...")
	return r.run([]string{
		"tx",
		"link",
		RelayerPathName,
	})
}

// linkWithports is the same as link, however ports are specified
func linkWithPorts(srcPort string, dstPort string, version string) func(*Relayer) error {
	return func(r *Relayer) error {
		r.logger.Info("linking chains for relayer...")
		return r.run([]string{
			"tx",
			"link",
			RelayerPathName,
			"--src-port",
			srcPort,
			"--dst-port",
			dstPort,
			"--version",
			version,
		})
	}
}

// -------------------------------
// lifecycle manager
func lifecycle(lfc ...func(*Relayer) error) func(*Relayer) error {
	return func(rly *Relayer) error {
		for i, lf := range lfc {
			if err := lf(rly); err != nil {
				return errors.Wrapf(err, "failed to run lifecycle during ibc step %d", i+1)
			}
		}

		return nil
	}
}

// Relayer cmd proxy caller
type Relayer struct {
	// home is Relayer home directory
	home   string
	logger log.Logger
	ctx    context.Context
}

func NewRelayer(
	ctx context.Context,
	home string,
	logger log.Logger,
) *Relayer {
	return &Relayer{
		home:   home,
		logger: logger,
		ctx:    ctx,
	}
}

func (r *Relayer) run(args []string) error {
	cmd := relayercmd.NewRootCmd(nil)
	cmd.SilenceUsage = true

	cmd.SetArgs(append(args, []string{"--home", r.home}...))
	return cmd.ExecuteContext(context.Background())
}

func (r *Relayer) UpdateClients() error {
	r.logger.Info("update clients...")
	return r.run([]string{
		"tx",
		"update-clients",
		RelayerPathName,
	})
}
