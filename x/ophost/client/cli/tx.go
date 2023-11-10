package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/initia-labs/OPinit/x/ophost/types"
	"github.com/spf13/cobra"
)

// GetTxCmd returns a root CLI command handler for all x/ophost transaction commands.
func GetTxCmd() *cobra.Command {
	ophostTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "OPChild transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	ophostTxCmd.AddCommand()

	return ophostTxCmd
}
