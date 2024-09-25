package cli

import (
	"github.com/initia-labs/OPinit/x/ophost/types"
)

// BridgeConfig defines the set of bridge config.
// NOTE: it is a modified BridgeConfig from x/ophost/types/types.go to make unmarshal easier
type BridgeCliConfig struct {
	Challenger            string          `json:"challenger"`
	Proposer              string          `json:"proposer"`
	SubmissionInterval    string          `json:"submission_interval"`
	FinalizationPeriod    string          `json:"finalization_period"`
	SubmissionStartHeight string          `json:"submission_start_height"`
	Metadata              string          `json:"metadata"`
	BatchInfo             types.BatchInfo `json:"batch_info"`
	OracleEnabled         bool            `json:"oracle_enabled"`
}
