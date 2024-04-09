package cli

import (
	"fmt"
	"strconv"
	"strings"

	"cosmossdk.io/core/address"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/initia-labs/OPinit/x/opchild/types"
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
	)

	return opchildTxCmd
}

// NewDepositCmd returns a CLI command handler for the transaction sending a deposit to an user account.
func NewDepositCmd(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit [sequence] [from_l1] [to_l2] [amount] [base_denom]",
		Args:  cobra.ExactArgs(5),
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

			from, err := ac.StringToBytes(args[1])
			if err != nil {
				return err
			}
			to, err := ac.StringToBytes(args[2])
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinNormalized(args[3])
			if err != nil {
				return err
			}

			baseDenom := args[4]

			hookMsg, err := cmd.Flags().GetString(FlagHookMsg)
			if err != nil {
				return err
			}

			txf, msg, err := newBuildDepositMsg(
				clientCtx, ac, txf, sequence,
				from, to, amount, baseDenom,
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

			to, err := ac.StringToBytes(args[0])
			if err != nil {
				return err
			}
			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			txf, msg, err := newBuildWithdrawMsg(clientCtx, ac, txf, to, amount)
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

func newBuildWithdrawMsg(clientCtx client.Context, ac address.Codec, txf tx.Factory, to sdk.AccAddress, amount sdk.Coin) (tx.Factory, *types.MsgInitiateTokenWithdrawal, error) {
	sender := clientCtx.GetFromAddress()
	senderAddr, err := ac.BytesToString(sender)
	if err != nil {
		return txf, nil, err
	}

	toAddr, err := ac.BytesToString(to)
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
	from, to sdk.AccAddress,
	amount sdk.Coin,
	baseDenom string,
	hookMsg []byte,
) (tx.Factory, *types.MsgFinalizeTokenDeposit, error) {
	sender := clientCtx.GetFromAddress()
	senderAddr, err := ac.BytesToString(sender)
	if err != nil {
		return txf, nil, err
	}

	fromAddr, err := ac.BytesToString(from)
	if err != nil {
		return txf, nil, err
	}

	toAddr, err := ac.BytesToString(to)
	if err != nil {
		return txf, nil, err
	}

	msg := types.NewMsgFinalizeTokenDeposit(senderAddr, fromAddr, toAddr, amount, sequence, baseDenom, hookMsg)
	if err := msg.Validate(ac); err != nil {
		return txf, nil, err
	}

	return txf, msg, nil
}
