package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/pkg/errors"

	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"

	ibcfeetypes "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	"github.com/initia-labs/OPinit/contrib/launchtools"
	"github.com/initia-labs/OPinit/contrib/launchtools/types"
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
		channelWithPorts(srcPort, dstPort, channelVersion),
	)

	return func(ctx launchtools.Launcher) error {
		// ibc relayer seems changing the bech32 prefix for the account,
		// so we need to reset it after the relayer setup is done
		originPrefix := sdk.GetConfig().GetBech32AccountAddrPrefix()
		originPubPrefix := sdk.GetConfig().GetBech32AccountPubPrefix()
		defer sdk.GetConfig().SetBech32PrefixForAccount(originPrefix, originPubPrefix)

		if !ctx.IsAppInitialized() {
			return errors.New("app is not initialized")
		}

		relayer, err := NewRelayer(ctx.Context(), relayerPath, ctx.Logger())
		if err != nil {
			return errors.Wrap(err, "failed to initialize relayer")
		}
		ctx.SetRelayer(relayer)
		return runLifecycle(relayer)
	}
}

// -------------------------------
func initializeConfig(r *Relayer) error {
	return r.run([]string{"config", "init"})
}

// initializeChains creates chain configuration files and initializes chains for the relayer
// "chains" in cosmos/relayer lingo means srcChain and dstChain. Specific ports are not created here.
// see initializePaths.
func initializeChains(config *launchtools.Config, basePath string) func(*Relayer) error {
	// ChainConfig is a struct that represents the configuration of a chain
	// cosmos/relayer specific

	var chainConfigs = [2]types.ChainConfig{
		{
			Type: "cosmos",
			Value: types.CosmosProviderConfig{
				Key:            RelayerKeyName,
				ChainID:        config.L1Config.ChainID,
				RPCAddr:        config.L1Config.RPC_URL,
				AccountPrefix:  "init",
				KeyringBackend: KeyringBackend,
				GasAdjustment:  1.5,
				GasPrices:      config.L1Config.GasPrices,
				Debug:          true,
				Timeout:        "200s",
				OutputFormat:   "json",
				Broadcast:      "batch",
			},
		},
		{
			Type: "cosmos",
			Value: types.CosmosProviderConfig{
				Key:            RelayerKeyName,
				ChainID:        config.L2Config.ChainID,
				RPCAddr:        "http://localhost:26657",
				AccountPrefix:  sdk.GetConfig().GetBech32AccountAddrPrefix(),
				KeyringBackend: KeyringBackend,
				GasAdjustment:  1.5,
				GasPrices:      "", // gas prices required for l2 txs
				Debug:          true,
				Timeout:        "200s",
				OutputFormat:   "json",
				Broadcast:      "batch",
			},
		},
	}

	// write chain configs to files
	for i, chainConfig := range chainConfigs {

		pathName := fmt.Sprintf("chain%d", i)
		fileName := fmt.Sprintf("%s/%s.json", basePath, pathName)

		if err := writeJSONConfig(fileName, chainConfig); err != nil {
			panic(errors.Wrap(err, "failed to write chain config"))
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

	pathConfig := types.Path{
		Src: &types.PathEnd{
			ChainID: config.L2Config.ChainID,
		},
		Dst: &types.PathEnd{
			ChainID: config.L1Config.ChainID,
		},
		Filter: types.ChannelFilter{
			Rule:        "",
			ChannelList: nil,
		},
	}

	if err := writeJSONConfig(fmt.Sprintf("%s/paths.json", basePath), pathConfig); err != nil {
		panic(errors.Wrap(err, "failed to write path config"))
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

	relayerKey := relayerKeyFromInput.Interface().(*launchtools.SystemAccount)
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

func marshalIBCFeeMetadata(appVersion string) ([]byte, error) {
	return json.Marshal(ibcfeetypes.Metadata{
		FeeVersion: ibcfeetypes.Version,
		AppVersion: appVersion,
	})
}

// link creates a default transfer channel between the chains
// it does all the heavy lifting of creating the channel, connection, and client
func link(r *Relayer) error {
	versionBz, err := marshalIBCFeeMetadata(ibctransfertypes.Version)
	if err != nil {
		return err
	}

	r.logger.Info("linking chains for relayer...", "version", string(versionBz))
	return r.run([]string{
		"tx",
		"link",
		RelayerPathName,
		"--version",
		string(versionBz),
		"--override",
	})
}

// channelWithPorts  create a channel reusing the same light client
func channelWithPorts(srcPort string, dstPort string, version string) func(*Relayer) error {
	return func(r *Relayer) error {
		versionBz, err := marshalIBCFeeMetadata(version)
		if err != nil {
			return err
		}

		r.logger.Info("linking chains for relayer...", "version", string(versionBz))
		return r.run([]string{
			"tx",
			"channel",
			RelayerPathName,
			"--src-port",
			srcPort,
			"--dst-port",
			dstPort,
			"--version",
			string(versionBz),
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
	bin    string
}

func NewRelayer(
	ctx context.Context,
	home string,
	logger log.Logger,
) (*Relayer, error) {
	bin, err := ensureRelayerBinary()
	if err != nil {
		return nil, err
	}

	return &Relayer{
		home:   home,
		logger: logger,
		ctx:    ctx,
		bin:    bin,
	}, nil
}

func (r *Relayer) run(args []string) error {
	cmdArgs := append(args, "--home", r.home)
	cmd := exec.CommandContext(r.ctx, r.bin, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (r *Relayer) UpdateClients() error {
	r.logger.Info("update clients...")
	return r.run([]string{
		"tx",
		"update-clients",
		RelayerPathName,
	})
}

func ensureRelayerBinary() (string, error) {
	// Check if rly is already in PATH
	if path, err := exec.LookPath("rly"); err == nil {
		if checkRelayerVersion(path, types.RlyVersion) {
			return path, nil
		}
	}

	// Download to a temporary directory
	destDir := filepath.Join(os.TempDir(), "opinit-bin")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", errors.Wrap(err, "failed to create binary destination directory")
	}

	destBin := filepath.Join(destDir, "rly")
	if _, err := os.Stat(destBin); err == nil {
		if checkRelayerVersion(destBin, types.RlyVersion) {
			return destBin, nil
		}
		// If version mismatch, remove the old binary
		_ = os.Remove(destBin)
	}

	goOS := runtime.GOOS
	goArch := runtime.GOARCH

	var osName string
	switch goOS {
	case "darwin":
		osName = "darwin"
	case "linux":
		osName = "linux"
	default:
		return "", fmt.Errorf("unsupported OS for automatic download: %s", goOS)
	}

	var archName string
	switch goArch {
	case "amd64":
		archName = "amd64"
	case "arm64":
		archName = "arm64"
	default:
		return "", fmt.Errorf("unsupported architecture for automatic download: %s", goArch)
	}

	versionNoV := strings.TrimPrefix(types.RlyVersion, "v")
	downloadURL := fmt.Sprintf("https://github.com/cosmos/relayer/releases/download/%s/Cosmos.Relayer_%s_%s_%s.tar.gz", types.RlyVersion, versionNoV, osName, archName)

	resp, err := http.Get(downloadURL) //nolint:gosec // G107: URL is constructed from constants and system properties
	if err != nil {
		return "", errors.Wrap(err, "failed to download rly binary")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download rly binary: status %d", resp.StatusCode)
	}

	// Extract tar.gz
	tarCmd := exec.Command("tar", "-xzf", "-", "-C", destDir)
	tarCmd.Stdin = resp.Body
	if err := tarCmd.Run(); err != nil {
		return "", errors.Wrap(err, "failed to extract rly binary")
	}

	// The binary might be in a subdirectory (e.g. "Cosmos Relayer_2.6.0-rc.2_darwin_amd64/rly")
	// We need to find it and move it to destBin
	err = filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() == "rly" {
			if path != destBin {
				return os.Rename(path, destBin)
			}
		}
		return nil
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to locate rly binary after extraction")
	}

	if _, err := os.Stat(destBin); err != nil {
		return "", errors.Wrap(err, "rly binary not found after extraction")
	}

	return destBin, nil
}

func checkRelayerVersion(binPath, expectedVersion string) bool {
	cmd := exec.Command(binPath, "version")
	out, err := cmd.Output()
	if err != nil {
		return false
	}

	// Expected output format: "version: 2.6.0" or similar
	// We'll check if the output contains the version string (without 'v' prefix if present in expectedVersion)
	versionStr := expectedVersion
	if len(versionStr) > 0 && versionStr[0] == 'v' {
		versionStr = versionStr[1:]
	}

	return strings.Contains(string(out), versionStr)
}

func writeJSONConfig(fileName string, v interface{}) error {
	bz, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(fileName, bz, 0600)
}
