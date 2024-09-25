package cli

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"cosmossdk.io/core/address"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/initia-labs/OPinit/x/ophost/types"
	"github.com/spf13/cobra"
)

// GetTxCmd returns a root CLI command handler for all x/ophost transaction commands.
func GetTxCmd(ac address.Codec) *cobra.Command {
	ophostTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "OPHost transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	ophostTxCmd.AddCommand(
		NewRecordBatchCmd(ac),
		NewCreateBridge(ac),
		NewProposeOutput(ac),
		NewDeleteOutput(ac),
		NewInitiateTokenDeposit(ac),
		NewFinalizeTokenWithdrawal(ac),
	)

	return ophostTxCmd
}

// NewRecordBatchCmd returns a CLI command handler for transaction to submitting a batch record.
func NewRecordBatchCmd(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "record-batch [bridge_id] [base64-encoded-batch-bytes]",
		Short: "send a batch-recording tx",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			bridgeId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			batchBytes, err := base64.StdEncoding.DecodeString(args[1])
			if err != nil {
				return err
			}

			fromAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			msg := types.NewMsgRecordBatch(fromAddr, bridgeId, batchBytes)
			if err = msg.Validate(ac); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewCreateBridgeCmd returns a CLI command handler for transaction to creating a bridge.
func NewCreateBridge(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-bridge [path/to/bridge-config.json]",
		Short: "send a bridge creating tx",
		Long: strings.TrimSpace(
			fmt.Sprintf(
				`send a tx to create a bridge with a config file as a json.
				Example:
				$ %s tx ophost create-bridge path/to/bridge-config.json
				
				Where bridge-config.json contains:
				{
					"challenger": "bech32-address",
					"proposer": "bech32-addresss",
					"submission_interval": "duration",
					"finalization_period": "duration",
					"submission_start_height" : "l2-block-height",
					"batch_info": {"submitter": "bech32-address","chain": "INITIA|CELESTIA"},
					"metadata": "{\"perm_channels\":[{\"port_id\":\"transfer\", \"channel_id\":\"channel-0\"}, {\"port_id\":\"icqhost\", \"channel_id\":\"channel-1\"}]}"
				}`, version.AppName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			configBytes, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}

			var origConfig BridgeCliConfig
			err = json.Unmarshal(configBytes, &origConfig)
			if err != nil {

				return err
			}

			submissionInterval, err := time.ParseDuration(origConfig.SubmissionInterval)
			if err != nil {
				return err
			}

			finalizationPeriod, err := time.ParseDuration(origConfig.FinalizationPeriod)
			if err != nil {
				return err
			}

			submissionStartHeight, err := strconv.ParseUint(origConfig.SubmissionStartHeight, 10, 64)
			if err != nil {
				return err
			}

			config := types.BridgeConfig{
				Challenger:            origConfig.Challenger,
				Proposer:              origConfig.Proposer,
				SubmissionInterval:    submissionInterval,
				FinalizationPeriod:    finalizationPeriod,
				SubmissionStartHeight: submissionStartHeight,
				Metadata:              []byte(origConfig.Metadata),
				BatchInfo:             origConfig.BatchInfo,
				OracleEnabled:         origConfig.OracleEnabled,
			}

			if err = config.Validate(ac); err != nil {

				return err
			}

			fromAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {

				return err
			}

			msg := types.NewMsgCreateBridge(fromAddr, config)
			if err = msg.Validate(ac); err != nil {

				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewProposeOutput returns a CLI command handler for transaction to propose an output.
func NewProposeOutput(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "propose-output [bridge-id] [output-index] [l2-block-number] [output-root-hash]",
		Short: "send a output-proposing tx",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			bridgeId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			outputIndex, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			l2BlockNumber, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			outputBytes, err := base64.StdEncoding.DecodeString(args[3])
			if err != nil {
				return err
			}

			fromAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			msg := types.NewMsgProposeOutput(fromAddr, bridgeId, outputIndex, l2BlockNumber, outputBytes)
			if err = msg.Validate(ac); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewDeleteOutput returns a CLI command handler for transaction to remove an output.
func NewDeleteOutput(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-output [bridge-id] [output-index]",
		Short: "send a output-proposing tx",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			bridgeId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			outputIndex, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			fromAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			msg := types.NewMsgDeleteOutput(fromAddr, bridgeId, outputIndex)
			if err = msg.Validate(ac); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewInitiateTokenDeposit returns a CLI command handler for transaction to initiate token deposit.
func NewInitiateTokenDeposit(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "initiate-token-deposit [bridge-id] [to] [amount] [data]",
		Short: "send a token deposit initiating tx",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			bridgeId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			// cannot validate to address here because it is l2 address.
			toAddr := args[1]
			if len(toAddr) == 0 {
				return fmt.Errorf("to address is required")
			}

			amount, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return err
			}

			data, err := hex.DecodeString(args[3])
			if err != nil {
				return err
			}

			fromAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			msg := types.NewMsgInitiateTokenDeposit(fromAddr, bridgeId, toAddr, amount, data)
			if err = msg.Validate(ac); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewFinalizeTokenWithdrawal returns a CLI command handler for transaction to finalize token withdrawal.
func NewFinalizeTokenWithdrawal(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "finalize-token-withdrawal [path/to/withdrawal-info.json]",
		Short: "send a tx to finalize token withdrawal",
		Long: strings.TrimSpace(
			fmt.Sprintf(
				`send a tx to finalize token withdrawal with withdrawal info json.
				Example:
				$ %s tx ophost finalize-token-withdrawal path/to/withdrawal-info.json
				
				Where withrawal-info.json contains:
				{
					"bridge_id": "1",
					"output_index": "1",
					"sequence": "1",
					"from" : "l2-bech32-address",
					"to" : "l1-bech32-address",
					"amount": {"amount": "10000000", "denom": "uinit"},
					"withdrawal_proofs": [ "base64-encoded proof1", "proof2", ... ],
					"version": "base64-encoded version",
					"storage_root": "base64-encoded storage-root",
					"last_block_hash": "base64-encoded latest-block-hash"
				}`, version.AppName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			withdrawalBytes, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}

			var msg types.MsgFinalizeTokenWithdrawal
			err = clientCtx.Codec.UnmarshalJSON(withdrawalBytes, &msg)
			if err != nil {
				return err
			}

			sender, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			msg.Sender = sender
			if err = msg.Validate(ac); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
