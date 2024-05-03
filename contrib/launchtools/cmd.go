package launchtools

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func LaunchCmd(
	appCreator AppCreator,
	defaultGenesisGetter func(denom string) map[string]json.RawMessage,
	steps []LauncherStepFuncFactory[Input],
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "launch [path to manifest]",
		Short: "Launch a new instance of the app",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			serverCtx := server.GetServerContextFromCmd(cmd)

			manifestPath := args[0]
			manifest, err := Input{}.FromFile(manifestPath)
			if err != nil {
				return err
			}

			launcher := NewLauncher(
				cmd,
				&clientCtx,
				serverCtx,
				appCreator,
				defaultGenesisGetter(manifest.L2Config.Denom),
			)

			stepFns := make([]LauncherStepFunc, len(steps))

			for stepI, step := range steps {
				stepFns[stepI] = step(*manifest)
			}

			for _, stepFn := range stepFns {
				if err := stepFn(launcher); err != nil {
					return errors.Wrapf(err, "failed to run launcher step")
				}
			}

			return nil
		},
	}

	return cmd
}
