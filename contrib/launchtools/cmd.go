package launchtools

import (
	"bufio"
	"encoding/json"
	"fmt"
	"path"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	flagArtifactsDir = "artifacts-dir"
	flagWithConfig   = "with-config"
)

func LaunchCmd(
	appCreator AppCreator,
	defaultGenesisGetter func(denom string) map[string]json.RawMessage,
	steps []LauncherStepFuncFactory[*Config],
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "launch [target-chain-id]",
		Short: "Launch a new instance of the app",
		Long: `Launch a new instance of the app. This command will execute a series of steps to
initialize the app and generate the necessary configuration files. The artifacts will be stored in the
specified directory. The command will output the artifacts to stdout too.

Example:
$ launchtools launch mahalo-3 --artifacts-dir ./ --with-config ./config.json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sdk.SetAddrCacheEnabled(false)
			defer sdk.SetAddrCacheEnabled(true)

			clientCtx := client.GetClientContextFromCmd(cmd)
			serverCtx := server.GetServerContextFromCmd(cmd)

			targetNetwork := args[0]
			if targetNetwork == "" {
				return errors.New("target chain id is required")
			}

			artifactsDir, err := cmd.Flags().GetString(flagArtifactsDir)
			if err != nil {
				return errors.Wrap(err, "failed to get artifacts flag")
			}

			configPath, err := cmd.Flags().GetString(flagWithConfig)
			if err != nil {
				return errors.Wrap(err, "failed to get config flag")
			}

			config, err := NewConfig(configPath)
			if err != nil {
				return err
			}

			if err := config.Finalize(targetNetwork, bufio.NewReader(clientCtx.Input)); err != nil {
				return errors.Wrap(err, "failed to finalize config")
			}

			launcher := NewLauncher(
				cmd,
				&clientCtx,
				serverCtx,
				appCreator,
				defaultGenesisGetter(config.L2Config.Denom),
				artifactsDir,
			)

			stepFns := make([]LauncherStepFunc, len(steps))

			for stepI, step := range steps {
				stepFns[stepI] = step(config)
			}

			for _, stepFn := range stepFns {
				if err := stepFn(launcher); err != nil {
					return errors.Wrapf(err, "failed to run launcher step")
				}
			}

			// print out the artifacts to stdout
			artifacts, err := launcher.FinalizeOutput(config)
			if err != nil {
				return errors.Wrap(err, "failed to get output")
			}

			if _, err := fmt.Fprintf(cmd.OutOrStdout(), `
############################################
Artifact written to 
* %s
* %s

`,
				path.Join(clientCtx.HomeDir, artifactsDir, "config.json"),
				path.Join(clientCtx.HomeDir, artifactsDir, "artifact.json"),
			); err != nil {
				return errors.Wrap(err, "failed to write artifacts to stdout")
			}

			if _, err := fmt.Fprintf(cmd.ErrOrStderr(), `%s`,
				artifacts,
			); err != nil {
				return errors.Wrap(err, "failed to write artifacts to stdout")
			}

			return nil
		},
	}

	cmd.Flags().String(flagArtifactsDir, "artifacts", "Path to the directory where artifacts will be stored")
	cmd.Flags().String(flagWithConfig, "", "Path to the config file to use for the launch")

	return cmd
}
