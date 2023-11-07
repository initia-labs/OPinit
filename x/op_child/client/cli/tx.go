package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/initia-labs/OPinit/x/op_child/types"
)

// GetTxCmd returns a root CLI command handler for all x/rollup transaction commands.
func GetTxCmd() *cobra.Command {
	rollupTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Rollup transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	rollupTxCmd.AddCommand(
		NewExecuteMessagesCmd(),
		NewDepositCmd(),
		NewWithdrawCmd(),
		NewLegacyContentParamChangeTxCmd(),
		NewLegacyContentSubmitUpdateClientCmd(),
		NewLegacyContentSubmitUpgradeCmd(),
	)

	return rollupTxCmd
}

// NewDepositCmd returns a CLI command handler for the transaction sending a deposit to an user account.
func NewDepositCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit [sequence] [from_l1] [to_l2] [amount]",
		Args:  cobra.ExactArgs(4),
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

			from, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}
			to, err := sdk.AccAddressFromBech32(args[2])
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinsNormalized(args[3])
			if err != nil {
				return err
			}

			hookMsg, err := cmd.Flags().GetString(FlagHookMsg)
			if err != nil {
				return err
			}

			txf, msg, err := newBuildDepositMsg(clientCtx, txf, cmd.Flags(), sequence, from, to, amount, []byte(hookMsg))
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
func NewWithdrawCmd() *cobra.Command {
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

			to, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			txf, msg, err := newBuildWithdrawMsg(clientCtx, txf, cmd.Flags(), to, amount)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewExecuteMessagesCmd returns a CLI command handler for transaction to administrating the system.
func NewExecuteMessagesCmd() *cobra.Command {
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

			msg, err := types.NewMsgExecuteMessages(clientCtx.GetFromAddress(), msgs)
			if err != nil {
				return fmt.Errorf("invalid message: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func newBuildWithdrawMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet, to sdk.AccAddress, amount sdk.Coin) (tx.Factory, *types.MsgWithdraw, error) {
	sender := clientCtx.GetFromAddress()

	msg := types.NewMsgWithdraw(sender, to, amount)
	if err := msg.ValidateBasic(); err != nil {
		return txf, nil, err
	}

	return txf, msg, nil
}

func newBuildDepositMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet,
	sequence uint64, from, to sdk.AccAddress, amount sdk.Coins, hookMsg []byte,
) (tx.Factory, *types.MsgDeposit, error) {
	sender := clientCtx.GetFromAddress()

	msg := types.NewMsgDeposit(sender, from, to, amount, sequence, hookMsg)
	if err := msg.ValidateBasic(); err != nil {
		return txf, nil, err
	}

	return txf, msg, nil
}
