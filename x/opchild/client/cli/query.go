package cli

import (
	"cosmossdk.io/core/address"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/initia-labs/OPinit/x/opchild/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(vc address.Codec) *cobra.Command {
	opchildQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the opchild module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	opchildQueryCmd.AddCommand()
	return opchildQueryCmd
}
