package launchtools

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"time"

	"github.com/pkg/errors"

	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"

	"github.com/initia-labs/OPinit/contrib/launchtools/utils"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
)

type Config struct {
	L1Config        *L1Config        `json:"l1_config,omitempty"`
	L2Config        *L2Config        `json:"l2_config,omitempty"`
	OpBridge        *OpBridge        `json:"op_bridge,omitempty"`
	SystemKeys      *SystemKeys      `json:"system_keys,omitempty"`
	GenesisAccounts *GenesisAccounts `json:"genesis_accounts,omitempty"`
}

func NewConfig(path string) (*Config, error) {
	if path == "" {
		return &Config{}, nil
	}

	bz, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to read file: %s", path))
	}

	ret := new(Config)
	if err := json.Unmarshal(bz, ret); err != nil {
		return nil, err
	}

	return ret, nil
}

func (i *Config) Finalize(targetNetwork string, buf *bufio.Reader) error {
	if i.L1Config == nil {
		i.L1Config = &L1Config{}
	}
	if i.L2Config == nil {
		i.L2Config = &L2Config{}
	}
	if i.OpBridge == nil {
		i.OpBridge = &OpBridge{}
	}
	if i.SystemKeys == nil {
		i.SystemKeys = &SystemKeys{}
	}
	if i.GenesisAccounts == nil {
		i.GenesisAccounts = &GenesisAccounts{}
	}

	// finalize all fields
	if err := i.L1Config.Finalize(targetNetwork); err != nil {
		return err
	}
	if err := i.L2Config.Finalize(); err != nil {
		return err
	}
	if err := i.OpBridge.Finalize(); err != nil {
		return err
	}
	if err := i.SystemKeys.Finalize(buf); err != nil {
		return err
	}
	if err := i.GenesisAccounts.Finalize(*i.SystemKeys); err != nil {
		return err
	}

	return nil
}

type L2Config struct {
	ChainID string `json:"chain_id,omitempty"`
	Denom   string `json:"denom,omitempty"`
	Moniker string `json:"moniker,omitempty"`

	// BridgeID will be generated after the launch.
	BridgeID uint64 `json:"bridge_id,omitempty"`
}

func (l2config *L2Config) Finalize() error {
	if l2config.ChainID == "" {
		l2config.ChainID = fmt.Sprintf("minitia-%s-1", randString(6))
	}

	if l2config.Denom == "" {
		l2config.Denom = "umin"
	}

	if l2config.Moniker == "" {
		l2config.Moniker = "operator"
	}

	return nil
}

type OpBridge struct {
	// output submission setup
	OutputSubmissionInterval    *time.Duration `json:"output_submission_interval,omitempty"`
	OutputFinalizationPeriod    *time.Duration `json:"output_finalization_period,omitempty"`
	OutputSubmissionStartHeight uint64         `json:"output_submission_start_height,omitempty"`

	// batch submission setup
	BatchSubmissionTarget ophosttypes.BatchInfo_ChainType `json:"batch_submission_target"`
}

func (opBridge *OpBridge) Finalize() error {
	if opBridge.OutputSubmissionStartHeight == 0 {
		opBridge.OutputSubmissionStartHeight = 1
	}

	if opBridge.BatchSubmissionTarget == ophosttypes.BatchInfo_CHAIN_TYPE_UNSPECIFIED {
		opBridge.BatchSubmissionTarget = ophosttypes.BatchInfo_CHAIN_TYPE_INITIA
	}

	if opBridge.OutputSubmissionInterval == nil {
		interval := time.Hour
		opBridge.OutputSubmissionInterval = &interval
	}

	if opBridge.OutputFinalizationPeriod == nil {
		period := time.Hour
		opBridge.OutputFinalizationPeriod = &period
	}

	return nil
}

func (opBridge *OpBridge) UnmarshalJSON(data []byte) error {
	var tmp struct {
		OutputSubmissionInterval    string                          `json:"output_submission_interval,omitempty"`
		OutputFinalizationPeriod    string                          `json:"output_finalization_period,omitempty"`
		OutputSubmissionStartHeight uint64                          `json:"output_submission_start_height,omitempty"`
		BatchSubmissionTarget       ophosttypes.BatchInfo_ChainType `json:"batch_submission_target,omitempty"`
	}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	if tmp.OutputSubmissionInterval != "" {
		d, err := time.ParseDuration(tmp.OutputSubmissionInterval)
		if err != nil {
			return err
		}

		opBridge.OutputSubmissionInterval = &d
	}

	if tmp.OutputFinalizationPeriod != "" {
		d, err := time.ParseDuration(tmp.OutputFinalizationPeriod)
		if err != nil {
			return err
		}

		opBridge.OutputFinalizationPeriod = &d
	}

	opBridge.OutputSubmissionStartHeight = tmp.OutputSubmissionStartHeight
	opBridge.BatchSubmissionTarget = tmp.BatchSubmissionTarget

	return nil
}

func (opBridge OpBridge) MarshalJSON() ([]byte, error) {
	tmp := struct {
		OutputSubmissionInterval    string                          `json:"output_submission_interval,omitempty"`
		OutputFinalizationPeriod    string                          `json:"output_finalization_period,omitempty"`
		OutputSubmissionStartHeight uint64                          `json:"output_submission_start_height,omitempty"`
		BatchSubmissionTarget       ophosttypes.BatchInfo_ChainType `json:"batch_submission_target,omitempty"`
	}{
		OutputSubmissionStartHeight: opBridge.OutputSubmissionStartHeight,
		BatchSubmissionTarget:       opBridge.BatchSubmissionTarget,
	}

	if opBridge.OutputSubmissionInterval != nil {
		tmp.OutputSubmissionInterval = opBridge.OutputSubmissionInterval.String()
	}

	if opBridge.OutputFinalizationPeriod != nil {
		tmp.OutputFinalizationPeriod = opBridge.OutputFinalizationPeriod.String()
	}

	return json.Marshal(tmp)
}

type L1Config struct {
	ChainID   string `json:"chain_id,omitempty"`
	RPC_URL   string `json:"rpc_url,omitempty"`
	GasPrices string `json:"gas_prices,omitempty"`
}

func (l1config *L1Config) Finalize(targetNetwork string) error {
	if l1config.ChainID == "" {
		l1config.ChainID = targetNetwork
	}

	if l1config.RPC_URL == "" {
		l1config.RPC_URL = fmt.Sprintf("https://rpc.%s.initia.xyz:443", targetNetwork)
	}

	if l1config.GasPrices == "" {
		l1config.GasPrices = "0.15uinit"
	}

	_, err := sdk.ParseDecCoins(l1config.GasPrices)
	if err != nil {
		return errors.Wrap(err, "failed to parse gas prices")
	}

	return nil
}

type SystemAccount struct {
	L1Address string `json:"l1_address,omitempty"`
	L2Address string `json:"l2_address,omitempty"`
	Mnemonic  string `json:"mnemonic,omitempty"`
}

type GenesisAccount struct {
	Address string `json:"address,omitempty"`
	Coins   string `json:"coins,omitempty"`
}

type GenesisAccounts []GenesisAccount

func (gas *GenesisAccounts) Finalize(systemKeys SystemKeys) error {
	keys := reflect.ValueOf(systemKeys)
	for idx := 0; idx < keys.NumField(); idx++ {
		k, ok := keys.Field(idx).Interface().(*SystemAccount)
		if !ok {
			return errors.New("systemKeys must be of type launcher.Account")
		}

		found := false
		for _, ga := range *gas {
			if ga.Address == k.L2Address {
				found = true
				break
			}
		}
		if found {
			continue
		}

		*gas = append(*gas, GenesisAccount{
			Address: k.L2Address,
			Coins:   "",
		})
	}

	for _, ga := range *gas {
		if ga.Address == "" {
			return errors.New("genesis account address cannot be empty")
		}

		_, err := sdk.ParseCoinsNormalized(ga.Coins)
		if err != nil {
			return errors.Wrap(err, "failed to parse genesis account coins")
		}
	}

	return nil
}

type SystemKeys struct {
	Validator       *SystemAccount `json:"validator,omitempty"`
	BridgeExecutor  *SystemAccount `json:"bridge_executor,omitempty"`
	OutputSubmitter *SystemAccount `json:"output_submitter,omitempty"`
	BatchSubmitter  *SystemAccount `json:"batch_submitter,omitempty"`

	// Challenger does not require mnemonic
	Challenger *SystemAccount `json:"challenger,omitempty"`
}

const mnemonicEntropySize = 256

func generateMnemonic() (string, error) {
	entropySeed, err := bip39.NewEntropy(mnemonicEntropySize)
	if err != nil {
		return "", err
	}
	mnemonic, err := bip39.NewMnemonic(entropySeed)
	if err != nil {
		return "", err
	}

	return mnemonic, nil
}

func deriveAddress(mnemonic string) (string, string, error) {
	algo := hd.Secp256k1
	derivedPriv, err := algo.Derive()(
		mnemonic,
		keyring.DefaultBIP39Passphrase,
		sdk.GetConfig().GetFullBIP44Path(),
	)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to derive private key")
	}

	privKey := algo.Generate()(derivedPriv)
	addrBz := privKey.PubKey().Address()

	// use init Bech32 prefix for l1 address
	l1Addr, err := utils.L1AddressCodec().BytesToString(addrBz)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to convert address to bech32")
	}

	l2Addr, err := utils.L2AddressCodec().BytesToString(addrBz)
	return l1Addr, l2Addr, err
}

func (systemKeys *SystemKeys) Finalize(buf *bufio.Reader) error {
	if systemKeys.Validator == nil {
		mnemonic, err := generateMnemonic()
		if err != nil {
			return errors.New("failed to generate mnemonic")
		}

		// derive address
		l1Addr, l2Addr, err := deriveAddress(mnemonic)
		if err != nil {
			return errors.Wrap(err, "failed to derive address")
		}

		systemKeys.Validator = &SystemAccount{
			L1Address: l1Addr,
			L2Address: l2Addr,
			Mnemonic:  mnemonic,
		}
	}
	if systemKeys.BatchSubmitter == nil {
		mnemonic, err := generateMnemonic()
		if err != nil {
			return errors.New("failed to generate mnemonic")
		}

		// derive address
		l1Addr, l2Addr, err := deriveAddress(mnemonic)
		if err != nil {
			return errors.Wrap(err, "failed to derive address")
		}

		systemKeys.BatchSubmitter = &SystemAccount{
			L1Address: l1Addr,
			L2Address: l2Addr,
			Mnemonic:  mnemonic,
		}
	}
	if systemKeys.BridgeExecutor == nil {
		mnemonic, err := input.GetString("Enter L1 gas token funded bridge_executor bip39 mnemonic", buf)
		if err != nil {
			return err
		}

		if !bip39.IsMnemonicValid(mnemonic) {
			return errors.New("invalid mnemonic")
		}

		// derive address
		l1Addr, l2Addr, err := deriveAddress(mnemonic)
		if err != nil {
			return errors.Wrap(err, "failed to derive address")
		}

		systemKeys.BridgeExecutor = &SystemAccount{
			L1Address: l1Addr,
			L2Address: l2Addr,
			Mnemonic:  mnemonic,
		}
	}
	if systemKeys.Challenger == nil {
		mnemonic, err := generateMnemonic()
		if err != nil {
			return errors.New("failed to generate mnemonic")
		}

		// derive address
		l1Addr, l2Addr, err := deriveAddress(mnemonic)
		if err != nil {
			return errors.Wrap(err, "failed to derive address")
		}

		systemKeys.Challenger = &SystemAccount{
			L1Address: l1Addr,
			L2Address: l2Addr,
			Mnemonic:  mnemonic,
		}
	}
	if systemKeys.OutputSubmitter == nil {
		mnemonic, err := generateMnemonic()
		if err != nil {
			return errors.New("failed to generate mnemonic")
		}

		// derive address
		l1Addr, l2Addr, err := deriveAddress(mnemonic)
		if err != nil {
			return errors.Wrap(err, "failed to derive address")
		}

		systemKeys.OutputSubmitter = &SystemAccount{
			L1Address: l1Addr,
			L2Address: l2Addr,
			Mnemonic:  mnemonic,
		}
	}

	// validate all accounts
	if systemKeys.Validator.L2Address == "" || systemKeys.Validator.Mnemonic == "" {
		return errors.New("validator account not initialized")
	}
	if systemKeys.BridgeExecutor.L1Address == "" || systemKeys.BridgeExecutor.L2Address == "" || systemKeys.BridgeExecutor.Mnemonic == "" {
		return errors.New("bridge_executor account not initialized")
	}
	if systemKeys.BatchSubmitter.L1Address == "" {
		return errors.New("batch_submitter account not initialized")
	}
	if systemKeys.OutputSubmitter.L1Address == "" {
		return errors.New("output_submitter account not initialized")
	}
	if systemKeys.Challenger.L1Address == "" {
		return errors.New("challenger account not initialized")
	}

	return nil
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	rand := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(b)
}
