package launchtools

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
)

type Input struct {
	L2Config        L2Config             `json:"l2_config"`
	L1Config        L1Config             `json:"l1_config"`
	OpBridge        OpBridge             `json:"op_bridge"`
	SystemKeys      SystemKeys           `json:"system_keys"`
	GenesisAccounts []AccountWithBalance `json:"genesis_accounts"`
}

func (input Input) FromFile(path string) (*Input, error) {
	bz, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to read file: %s", path))
	}

	ret := new(Input)
	if err := json.Unmarshal(bz, ret); err != nil {
		return nil, err
	}
	return ret, nil
}

type L2Config struct {
	ChainID       string `json:"chain_id"`
	Denom         string `json:"denom"`
	Moniker       string `json:"moniker"`
	AccountPrefix string `json:"account_prefix"`
	GasPrices     string `json:"gas_prices"`
}

type OpBridge struct {
	SubmissionStartTime time.Time `json:"submission_start_time"`
	SubmitTarget        string    `json:"submit_target"`
	SubmissionInterval  string    `json:"submission_interval"`
	FinalizationPeriod  string    `json:"finalization_period"`
}

type L1Config struct {
	ChainID       string `json:"chain_id"`
	RPCURL        string `json:"rpc_url"`
	Denom         string `json:"denom"`
	RestURL       string `json:"rest_url"`
	GrpcURL       string `json:"grpc_url"`
	WsURL         string `json:"ws_url"`
	AccountPrefix string `json:"account_prefix"`
	GasPrices     string `json:"gas_prices"`
}

type Account struct {
	Address  string `json:"address"`
	Mnemonic string `json:"mnemonic"`
}

type AccountWithBalance struct {
	Account
	Coins string `json:"coins"`
}

type SystemKeys struct {
	Validator  Account `json:"validator"`
	Executor   Account `json:"executor"`
	Output     Account `json:"output"`
	Challenger Account `json:"challenger"`
	Submitter  Account `json:"submitter"`
	Relayer    Account `json:"relayer"`
}

func (i Input) Validate() error {
	return nil
}
