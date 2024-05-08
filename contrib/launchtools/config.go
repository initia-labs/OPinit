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
)

type Config struct {
	L1Config        *L1Config        `json:"l1_config,omitempty"`
	L2Config        *L2Config        `json:"l2_config,omitempty"`
	OpBridge        *OpBridge        `json:"op_bridge,omitempty"`
	SystemKeys      *SystemKeys      `json:"system_keys,omitempty"`
	GenesisAccounts *GenesisAccounts `json:"genesis_accounts,omitempty"`
}

func (input Config) FromFile(path string) (*Config, error) {
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
	OutputSubmissionStartTime *time.Time     `json:"output_submission_start_time,omitempty"`
	OutputSubmissionInterval  *time.Duration `json:"output_submission_interval,omitempty"`
	OutputFinalizationPeriod  *time.Duration `json:"output_finalization_period,omitempty"`

	// batch submission setup
	BatchSubmitTarget string `json:"batch_submission_target"`
}

func (opBridge *OpBridge) Finalize() error {
	if opBridge.OutputSubmissionStartTime == nil {
		now := time.Now()
		opBridge.OutputSubmissionStartTime = &now
	}

	if opBridge.BatchSubmitTarget == "" {
		opBridge.BatchSubmitTarget = "l1"
	}

	if opBridge.OutputSubmissionInterval == nil {
		interval := time.Hour
		opBridge.OutputSubmissionInterval = &interval
	}

	if opBridge.OutputFinalizationPeriod == nil {
		period := time.Hour
		opBridge.OutputFinalizationPeriod = &period
	}

	// TODO: validate batch submit target
	// if opBridge.BatchSubmitTarget != "l1" && opBridge.BatchSubmitTarget != "celestia" {
	// 	return errors.New(fmt.Sprintf("invalid batch submit target: %s", opBridge.BatchSubmitTarget))
	// }

	return nil
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

type Account struct {
	Address  string `json:"address,omitempty"`
	Mnemonic string `json:"mnemonic,omitempty"`
}

type AccountWithBalance struct {
	Account
	Coins string `json:"coins,omitempty"`
}

type GenesisAccounts []AccountWithBalance

func (gas *GenesisAccounts) Finalize(systemKeys SystemKeys) error {
	keys := reflect.ValueOf(systemKeys)
	for idx := 0; idx < keys.NumField(); idx++ {
		k, ok := keys.Field(idx).Interface().(*Account)
		if !ok {
			return errors.New("systemKeys must be of type launcher.Account")
		}

		found := false
		for _, ga := range *gas {
			if ga.Address == k.Address {
				found = true
				break
			}
		}
		if found {
			continue
		}

		*gas = append(*gas, AccountWithBalance{
			Account: Account{Address: k.Address},
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
	Validator       *Account `json:"validator,omitempty"`
	BridgeExecutor  *Account `json:"bridge_executor,omitempty"`
	OutputSubmitter *Account `json:"output_submitter,omitempty"`
	BatchSubmitter  *Account `json:"batch_submitter,omitempty"`

	// Challenger does not require mnemonic
	Challenger *Account `json:"challenger,omitempty"`
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

func deriveAddress(mnemonic string) (string, error) {
	algo := hd.Secp256k1
	derivedPriv, err := algo.Derive()(
		mnemonic,
		keyring.DefaultBIP39Passphrase,
		sdk.GetConfig().GetFullBIP44Path(),
	)
	if err != nil {
		return "", errors.Wrap(err, "failed to derive private key")
	}

	privKey := algo.Generate()(derivedPriv)
	return sdk.AccAddress(privKey.PubKey().Address()).String(), nil
}

func (systemKeys *SystemKeys) Finalize(buf *bufio.Reader) error {
	if systemKeys.Validator == nil {
		mnemonic, err := generateMnemonic()
		if err != nil {
			return errors.New("failed to generate mnemonic")
		}

		// derive address
		addr, err := deriveAddress(mnemonic)
		if err != nil {
			return errors.Wrap(err, "failed to derive address")
		}

		systemKeys.Validator = &Account{
			Address:  addr,
			Mnemonic: mnemonic,
		}
	}
	if systemKeys.BatchSubmitter == nil {
		mnemonic, err := generateMnemonic()
		if err != nil {
			return errors.New("failed to generate mnemonic")
		}

		// derive address
		addr, err := deriveAddress(mnemonic)
		if err != nil {
			return errors.Wrap(err, "failed to derive address")
		}

		systemKeys.BatchSubmitter = &Account{
			Address:  addr,
			Mnemonic: mnemonic,
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
		addr, err := deriveAddress(mnemonic)
		if err != nil {
			return errors.Wrap(err, "failed to derive address")
		}

		systemKeys.BridgeExecutor = &Account{
			Address:  addr,
			Mnemonic: mnemonic,
		}
	}
	if systemKeys.Challenger == nil {
		mnemonic, err := generateMnemonic()
		if err != nil {
			return errors.New("failed to generate mnemonic")
		}

		// derive address
		addr, err := deriveAddress(mnemonic)
		if err != nil {
			return errors.Wrap(err, "failed to derive address")
		}

		systemKeys.Challenger = &Account{
			Address:  addr,
			Mnemonic: mnemonic,
		}
	}
	if systemKeys.OutputSubmitter == nil {
		mnemonic, err := generateMnemonic()
		if err != nil {
			return errors.New("failed to generate mnemonic")
		}

		// derive address
		addr, err := deriveAddress(mnemonic)
		if err != nil {
			return errors.Wrap(err, "failed to derive address")
		}

		systemKeys.OutputSubmitter = &Account{
			Address:  addr,
			Mnemonic: mnemonic,
		}
	}

	// validate all accounts
	if systemKeys.Validator.Address == "" || systemKeys.Validator.Mnemonic == "" {
		return errors.New("validator account not initialized")
	}
	if systemKeys.BatchSubmitter.Address == "" || systemKeys.BatchSubmitter.Mnemonic == "" {
		return errors.New("batch_submitter account not initialized")
	}
	if systemKeys.BridgeExecutor.Address == "" || systemKeys.BridgeExecutor.Mnemonic == "" {
		return errors.New("bridge_executor account not initialized")
	}
	if systemKeys.OutputSubmitter.Address == "" || systemKeys.OutputSubmitter.Mnemonic == "" {
		return errors.New("output_submitter account not initialized")
	}
	if systemKeys.Challenger.Address == "" {
		return errors.New("challenger account not initialized")
	}

	return nil
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(b)
}
