package launchtools

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	FlagArtifacts = "FlagArtifacts"
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

			artifactsDir, err := cmd.Flags().GetString(FlagArtifacts)
			if err != nil {
				return errors.Wrap(err, "failed to get artifacts flag")
			}

			launcher := NewLauncher(
				cmd,
				&clientCtx,
				serverCtx,
				appCreator,
				defaultGenesisGetter(manifest.L2Config.Denom),
				artifactsDir,
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

			// print out the artifacts to stdout
			artifacts, err := launcher.FinalizeOutput()
			if err != nil {
				return errors.Wrap(err, "failed to get output")
			}

			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s", artifacts); err != nil {
				return errors.Wrap(err, "failed to write artifacts to stdout")
			}

			return nil
		},
	}

	cmd.Flags().String(FlagArtifacts, "artifacts", "Path to the directory where artifacts will be stored")

	return cmd
}
