package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/version"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	paramscutils "github.com/cosmos/cosmos-sdk/x/params/client/utils"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/initia-labs/OPinit/x/opchild/types"
)

// NewLegacyContentParamChangeTxCmd returns a CLI command handler for creating
// a parameter change legacy content validator transaction.
func NewLegacyContentParamChangeTxCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "param-change [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a parameter change proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a parameter proposal along with an initial deposit.
The proposal details must be supplied via a JSON file. For values that contains
objects, only non-empty fields will be updated.

IMPORTANT: Currently parameter changes are evaluated but not validated, so it is
very important that any "value" change is valid (ie. correct type and within bounds)
for its respective parameter, eg. "MaxValidators" should be an integer and not a decimal.

Proper vetting of a parameter change proposal should prevent this from happening
(no deposits should occur during the governance process), but it should be noted
regardless.

Example:
$ %s tx gov submit-proposal param-change <path/to/proposal.json> --from=<key_or_address>

Where proposal.json contains:

{
  "title": "Staking Param Change",
  "description": "Update max validators",
  "changes": [
    {
      "subspace": "staking",
      "key": "MaxValidators",
      "value": 105
    }
  ]
}
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			proposal, err := paramscutils.ParseParamChangeProposalJSON(clientCtx.LegacyAmino, args[0])
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()
			content := paramproposal.NewParameterChangeProposal(
				proposal.Title, proposal.Description, proposal.Changes.ToParamChanges(),
			)

			msg, err := types.NewMsgExecuteLegacyContents(from, []govv1beta1.Content{content})
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}

// NewLegacyContentSubmitUpdateClientCmd implements a command handler for submitting an update IBC client transaction.
func NewLegacyContentSubmitUpdateClientCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-client [subject-client-id] [substitute-client-id]",
		Args:  cobra.ExactArgs(2),
		Short: "Submit an update IBC client proposal",
		Long: "Submit an update IBC client proposal along with an initial deposit.\n" +
			"Please specify a subject client identifier you want to update..\n" +
			"Please specify the substitute client the subject client will be updated to.",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			title, err := cmd.Flags().GetString(govcli.FlagTitle) //nolint:staticcheck // need this till full govv1 conversion.
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(govcli.FlagDescription) //nolint:staticcheck // need this till full govv1 conversion.
			if err != nil {
				return err
			}

			subjectClientID := args[0]
			substituteClientID := args[1]

			content := ibcclienttypes.NewClientUpdateProposal(title, description, subjectClientID, substituteClientID)

			from := clientCtx.GetFromAddress()
			msg, err := types.NewMsgExecuteLegacyContents(from, []govv1beta1.Content{content})
			if err != nil {
				return err
			}

			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(govcli.FlagTitle, "", "title of proposal")             //nolint:staticcheck // need this till full govv1 conversion.
	cmd.Flags().String(govcli.FlagDescription, "", "description of proposal") //nolint:staticcheck // need this till full govv1 conversion.

	return cmd
}

// NewLegacyContentSubmitUpgradeCmd implements a command handler for submitting an upgrade IBC client transaction.
func NewLegacyContentSubmitUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ibc-upgrade [name] [height] [path/to/upgraded_client_state.json] [flags]",
		Args:  cobra.ExactArgs(3),
		Short: "Submit an IBC upgrade proposal",
		Long: "Submit an IBC client breaking upgrade proposal along with an initial deposit.\n" +
			"The client state specified is the upgraded client state representing the upgraded chain\n" +
			`Example Upgraded Client State JSON: 
{
	"@type":"/ibc.lightclients.tendermint.v1.ClientState",
 	"chain_id":"testchain1",
	"unbonding_period":"1814400s",
	"latest_height":{"revision_number":"0","revision_height":"2"},
	"proof_specs":[{"leaf_spec":{"hash":"SHA256","prehash_key":"NO_HASH","prehash_value":"SHA256","length":"VAR_PROTO","prefix":"AA=="},"inner_spec":{"child_order":[0,1],"child_size":33,"min_prefix_length":4,"max_prefix_length":12,"empty_child":null,"hash":"SHA256"},"max_depth":0,"min_depth":0},{"leaf_spec":{"hash":"SHA256","prehash_key":"NO_HASH","prehash_value":"SHA256","length":"VAR_PROTO","prefix":"AA=="},"inner_spec":{"child_order":[0,1],"child_size":32,"min_prefix_length":1,"max_prefix_length":1,"empty_child":null,"hash":"SHA256"},"max_depth":0,"min_depth":0}],
	"upgrade_path":["upgrade","upgradedIBCState"],
}
			`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			cdc := codec.NewProtoCodec(clientCtx.InterfaceRegistry)

			title, err := cmd.Flags().GetString(govcli.FlagTitle) //nolint:staticcheck // need this till full govv1 conversion.
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(govcli.FlagDescription) //nolint:staticcheck // need this till full govv1 conversion.
			if err != nil {
				return err
			}

			name := args[0]

			height, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return err
			}

			plan := upgradetypes.Plan{
				Name:   name,
				Height: height,
			}

			// attempt to unmarshal client state argument
			var clientState exported.ClientState
			clientContentOrFileName := args[2]
			if err := cdc.UnmarshalInterfaceJSON([]byte(clientContentOrFileName), &clientState); err != nil {

				// check for file path if JSON input is not provided
				contents, err := os.ReadFile(clientContentOrFileName)
				if err != nil {
					return fmt.Errorf("neither JSON input nor path to .json file for client state were provided: %w", err)
				}

				if err := cdc.UnmarshalInterfaceJSON(contents, &clientState); err != nil {
					return fmt.Errorf("error unmarshalling client state file: %w", err)
				}
			}

			content, err := ibcclienttypes.NewUpgradeProposal(title, description, plan, clientState)
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()
			msg, err := types.NewMsgExecuteLegacyContents(from, []govv1beta1.Content{content})
			if err != nil {
				return err
			}

			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(govcli.FlagTitle, "", "title of proposal")             //nolint:staticcheck // need this till full govv1 conversion.
	cmd.Flags().String(govcli.FlagDescription, "", "description of proposal") //nolint:staticcheck // need this till full govv1 conversion.

	return cmd
}
