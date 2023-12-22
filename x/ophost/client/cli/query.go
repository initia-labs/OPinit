package cli

import (
	"cosmossdk.io/core/address"
	"github.com/cosmos/cosmos-sdk/client"

	"github.com/initia-labs/OPinit/x/ophost/types"

	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module
// TODO implement query cmd
func GetQueryCmd(ac address.Codec) *cobra.Command {
	ophostQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the ophost module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	ophostQueryCmd.AddCommand()

	return ophostQueryCmd
}
