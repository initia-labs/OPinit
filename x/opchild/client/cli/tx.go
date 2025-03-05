package cli

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"cosmossdk.io/core/address"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/initia-labs/OPinit/v1/x/opchild/types"
	ophostcli "github.com/initia-labs/OPinit/v1/x/ophost/client/cli"
	ophosttypes "github.com/initia-labs/OPinit/v1/x/ophost/types"
)

// GetTxCmd returns a root CLI command handler for all x/opchild transaction commands.
func GetTxCmd(ac address.Codec) *cobra.Command {
	opchildTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "OPChild transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	opchildTxCmd.AddCommand(
		NewExecuteMessagesCmd(ac),
		NewDepositCmd(ac),
		NewWithdrawCmd(ac),
		NewSetBridgeInfoCmd(ac),
		NewUpdateOracleCmd(ac),
	)

	return opchildTxCmd
}

// NewDepositCmd returns a CLI command handler for the transaction sending a deposit to an user account.
func NewDepositCmd(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit [sequence] [height] [from_l1] [to_l2] [amount] [base_denom]",
		Args:  cobra.ExactArgs(6),
		Short: "send a deposit to an user account",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			sequence, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			height, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			// we can't validate the address here because the address is not l2 address.
			fromAddr := args[2]

			to, err := ac.StringToBytes(args[3])
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinNormalized(args[4])
			if err != nil {
				return err
			}

			baseDenom := args[5]

			hookMsg, err := cmd.Flags().GetString(FlagHookMsg)
			if err != nil {
				return err
			}

			txf, msg, err := newBuildDepositMsg(
				clientCtx, ac, txf, sequence, height,
				fromAddr, to, amount, baseDenom,
				[]byte(hookMsg),
			)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().String(FlagHookMsg, "", "Hook message passed from the upper layer")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewWithdrawCmd returns a CLI command handler for the transaction sending a deposit to an user account.
func NewWithdrawCmd(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw [to_l1] [amount]",
		Short: "withdraw a token from L2 to L1",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			// we can't validate the address here because the address is not l2 address.
			toAddr := args[0]

			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			txf, msg, err := newBuildWithdrawMsg(clientCtx, ac, txf, toAddr, amount)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewUpdateOracleCmd returns a CLI command handler for the transaction updating oracle data.
func NewUpdateOracleCmd(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-oracle [height] [data]",
		Short: "update oracle data",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			height, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			data, err := base64.StdEncoding.DecodeString(args[1])
			if err != nil {
				return err
			}

			fromAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateOracle(fromAddr, height, data)
			if err := msg.Validate(ac); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewExecuteMessagesCmd returns a CLI command handler for transaction to administrating the system.
func NewExecuteMessagesCmd(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execute-messages [path/to/messages.json]",
		Short: "send a system administrating tx",
		Long: strings.TrimSpace(
			fmt.Sprintf(
				`send a system administating tx with some messages.
They should be defined in a JSON file.

Example:
$ %s tx rollup execute-messages path/to/proposal.json

Where proposal.json contains:

{
  // array of proto-JSON-encoded sdk.Msgs
  "messages": [
    {
      "@type": "/cosmos.bank.v1beta1.MsgSend",
      "from_address": "init1...",
      "to_address": "init11...",
      "amount":[{"denom": "uminit","amount": "10"}]
    }
  ],
}
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msgs, err := parseExecuteMessages(clientCtx.Codec, args[0])
			if err != nil {
				return err
			}

			fromAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			msg, err := types.NewMsgExecuteMessages(fromAddr, msgs)
			if err != nil {
				return fmt.Errorf("invalid message: %w", err)
			}

			if err := msg.Validate(ac); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewSetBridgeInfoCmd returns a CLI command handler for transaction to setting a bridge info.
func NewSetBridgeInfoCmd(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-bridge-info [bridge-id] [bridge-addr] [l1-chain-id] [l1-client-id] [path/to/bridge-config.json]",
		Short: "send a bridge creating tx",
		Long: strings.TrimSpace(
			fmt.Sprintf(
				`send a tx to set a bridge info with a config file as a json.
				Example:
				$ %s tx opchild set-bridge-info 1 init10d07y265gmmuvt4z0w9aw880jnsr700j55nka3 mahalo-2 07-tendermint-0 path/to/bridge-config.json
				
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
		Args: cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			bridgeId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			bridgeAddr := args[1]
			l1ChainId := args[2]
			l1ClientId := args[3]

			configBytes, err := os.ReadFile(args[4])
			if err != nil {
				return err
			}

			var origConfig ophostcli.BridgeCliConfig
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

			bridgeConfig := ophosttypes.BridgeConfig{
				Challenger:            origConfig.Challenger,
				Proposer:              origConfig.Proposer,
				SubmissionInterval:    submissionInterval,
				FinalizationPeriod:    finalizationPeriod,
				SubmissionStartHeight: submissionStartHeight,
				Metadata:              []byte(origConfig.Metadata),
				BatchInfo:             origConfig.BatchInfo,
				OracleEnabled:         origConfig.OracleEnabled,
			}

			if err = bridgeConfig.ValidateWithNoAddrValidation(); err != nil {
				return err
			}

			fromAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			msg := types.NewMsgSetBridgeInfo(fromAddr, types.BridgeInfo{
				BridgeId:     bridgeId,
				BridgeAddr:   bridgeAddr,
				L1ChainId:    l1ChainId,
				L1ClientId:   l1ClientId,
				BridgeConfig: bridgeConfig,
			})
			if err = msg.Validate(ac); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func newBuildWithdrawMsg(clientCtx client.Context, ac address.Codec, txf tx.Factory, toAddr string, amount sdk.Coin) (tx.Factory, *types.MsgInitiateTokenWithdrawal, error) {
	sender := clientCtx.GetFromAddress()
	senderAddr, err := ac.BytesToString(sender)
	if err != nil {
		return txf, nil, err
	}

	msg := types.NewMsgInitiateTokenWithdrawal(senderAddr, toAddr, amount)
	if err := msg.Validate(ac); err != nil {
		return txf, nil, err
	}

	return txf, msg, nil
}

func newBuildDepositMsg(
	clientCtx client.Context,
	ac address.Codec,
	txf tx.Factory,
	sequence uint64,
	height uint64,
	fromAddr string,
	to sdk.AccAddress,
	amount sdk.Coin,
	baseDenom string,
	hookMsg []byte,
) (tx.Factory, *types.MsgFinalizeTokenDeposit, error) {
	sender := clientCtx.GetFromAddress()
	senderAddr, err := ac.BytesToString(sender)
	if err != nil {
		return txf, nil, err
	}

	toAddr, err := ac.BytesToString(to)
	if err != nil {
		return txf, nil, err
	}

	msg := types.NewMsgFinalizeTokenDeposit(senderAddr, fromAddr, toAddr, amount, sequence, height, baseDenom, hookMsg)
	if err := msg.Validate(ac); err != nil {
		return txf, nil, err
	}

	return txf, msg, nil
}
