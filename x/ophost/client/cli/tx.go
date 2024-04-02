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
					"submission_start_time" : "rfc3339-datetime",
					"batch_info": {"submitter": "bech32-address","chain": "l1|celestia"},
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

			origConfig := BridgeConfig{}
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

			submissionStartTime, err := time.Parse(time.RFC3339, origConfig.SubmissionStartTime)
			if err != nil {
				return err
			}

			config := types.BridgeConfig{
				Challenger:          origConfig.Challenger,
				Proposer:            origConfig.Proposer,
				SubmissionInterval:  submissionInterval,
				FinalizationPeriod:  finalizationPeriod,
				SubmissionStartTime: submissionStartTime,
				Metadata:            []byte(origConfig.Metadata),
				BatchInfo:           origConfig.BatchInfo,
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
		Use:   "propose-output [bridge-id] [l2-block-number] [output-root-hash]",
		Short: "send a output-proposing tx",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			bridgeId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			l2BlockNumber, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			outputBytes, err := hex.DecodeString(args[2])
			if err != nil {
				return err
			}

			fromAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			msg := types.NewMsgProposeOutput(fromAddr, bridgeId, l2BlockNumber, outputBytes)
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

			toAddr := args[1]
			_, err = ac.StringToBytes(toAddr)
			if err != nil {
				return err
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
		Short: "send a token deposit initiating tx",
		Long: strings.TrimSpace(
			fmt.Sprintf(
				`send a tx to finalize tokwn withdrawal with withdrawal info json.
				Example:
				$ %s tx ophost finalize-token-withdrawal path/to/withdrawal-info.json
				
				Where withrawal-info.json contains:
				{
					"bridge_id": 1,
					"output_index": 0,
					"withdrawal_proofs": [ "proof1", "proof2", ... ],
					"receiver": "bech32-address",
					"sequence": 0,
					"amount": "10000000uatom",
					"version": "version hex",
					"state_root": "state-root hex",
					"storage_root": "storage-root hex",
					"latest_block_hash": "latest-block-hash"
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
			withdrawalInfo := MsgFinalizeTokenWithdrawal{}
			err = json.Unmarshal(withdrawalBytes, &withdrawalInfo)
			if err != nil {
				return err
			}

			withdrawalProofs := make([][]byte, len(withdrawalInfo.WithdrawalProofs))
			for i, wp := range withdrawalInfo.WithdrawalProofs {
				withdrawalProofs[i], err = hex.DecodeString(wp)
				if err != nil {
					return err
				}
			}

			receiver := withdrawalInfo.Receiver
			_, err = ac.StringToBytes(receiver)
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinNormalized(withdrawalInfo.Amount)
			if err != nil {
				return err
			}

			version, err := hex.DecodeString(withdrawalInfo.Version)
			if err != nil {
				return err
			}

			stateRoot, err := hex.DecodeString(withdrawalInfo.StateRoot)
			if err != nil {
				return err
			}

			storageRoot, err := hex.DecodeString(withdrawalInfo.StorageRoot)
			if err != nil {
				return err
			}

			latestBlockHash, err := hex.DecodeString(withdrawalInfo.LatestBlockHash)
			if err != nil {
				return err
			}

			fromAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			msg := types.NewMsgFinalizeTokenWithdrawal(
				withdrawalInfo.BridgeId,
				withdrawalInfo.OutputIndex,
				withdrawalInfo.Sequence,
				withdrawalProofs,
				fromAddr,
				receiver,
				amount,
				version,
				stateRoot,
				storageRoot,
				latestBlockHash,
			)
			if err = msg.Validate(ac); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
