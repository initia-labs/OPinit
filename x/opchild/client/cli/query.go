package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/initia-labs/OPinit/x/opchild/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	rollupQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the rollup module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	rollupQueryCmd.AddCommand(
		GetCmdQueryValidator(),
		GetCmdQueryValidators(),
		//GetCmdQueryHistoricalInfo(), TODO: remove it if not needed
		GetCmdQueryParams(),
	)

	return rollupQueryCmd
}

// GetCmdQueryValidator implements the validator query command.
func GetCmdQueryValidator() *cobra.Command {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "validator [validator-addr]",
		Short: "Query a validator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details about an individual validator.

Example:
$ %s query rollup validator %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.AppName, bech32PrefixValAddr,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			addr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := &types.QueryValidatorRequest{ValidatorAddr: addr.String()}
			res, err := queryClient.Validator(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Validator)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryValidators implements the query all validators command.
func GetCmdQueryValidators() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validators",
		Short: "Query for all validators",
		Args:  cobra.NoArgs,
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details about all validators on a network.

Example:
$ %s query rollup validators
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			result, err := queryClient.Validators(cmd.Context(), &types.QueryValidatorsRequest{
				// Leaving status empty on purpose to query all validators.
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(result)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "validators")

	return cmd
}

/* TODO: remove it if not needed
// GetCmdQueryHistoricalInfo implements the historical info query command
func GetCmdQueryHistoricalInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "historical-info [height]",
		Args:  cobra.ExactArgs(1),
		Short: "Query historical info at given height",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query historical info at given height.

Example:
$ %s query rollup historical-info 5
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			height, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || height < 0 {
				return fmt.Errorf("height argument provided must be a non-negative-integer: %v", err)
			}

			params := &types.QueryHistoricalInfoRequest{Height: height}
			res, err := queryClient.HistoricalInfo(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res.Hist)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
*/

// GetCmdQueryParams implements the params query command.
func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query the current rollup parameters information",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query values set as rollup parameters.

Example:
$ %s query rollup params
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
