package steps

import (
	"context"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/initia-labs/OPinit/contrib/launchtools"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

var _ launchtools.LauncherStepFuncFactory[*launchtools.Config] = RunApp

// RunApp runs the in-process application. This routine temporarily allows creation of empty blocks,
// in order to expedite IBC channel establishment. It waits until the app generates at least 1 block after the genesis.
func RunApp(_ *launchtools.Config) launchtools.LauncherStepFunc {
	return func(ctx launchtools.Launcher) error {
		// temporarily allow creation of empty blocks
		// this should help creation of ibc channels.
		// NOTE: This part is ephemeral only in the context of the launcher.
		ctx.ServerContext().Config.Consensus.CreateEmptyBlocks = true
		ctx.ServerContext().Config.Consensus.CreateEmptyBlocksInterval = CreateEmptyBlocksInterval

		// create a channel to synchronize on app creation
		var syncDone = make(chan interface{})

		// create cobra command context
		startCmd := server.StartCmdWithOptions(
			ctx.AppCreator(),
			ctx.ClientContext().HomeDir,

			// set up a post setup function to set the app in the context
			server.StartCmdOptions{
				PostSetup: func(svrCtx *server.Context, clientCtx client.Context, _ context.Context, g *errgroup.Group) (err error) {
					// set the error group to gracefully shutdown the launch cmd
					ctx.SetErrorGroup(g)

					// wait until latest version goes to 2
					g.Go(func() error {
						for {
							ctx.Logger().Info("waiting for app to be created")

							if ctx.App().CommitMultiStore().LatestVersion() > 1 {
								// Signal that the app is created
								syncDone <- struct{}{}
								break
							}

							time.Sleep(1 * time.Second)
						}

						return nil
					})

					return nil
				},
			},
		)

		// set relevant context; this part is necessary to correctly set up the start command and their start-up flags
		startCmd.SetContext(ctx.Context())

		// Run PreRunE from startCmd. This step is necessary to correctly set up start-up flags,
		// as it is done usually with cometbft start command.
		if err := startCmd.PreRunE(startCmd, nil); err != nil {
			return errors.Wrapf(err, "failed to prerun command")
		}

		// Run RunE command - this part fires up the actual chain
		// Note that the command is run in a separate goroutine, as it is blocking.
		// App should be later cleaned up in another launcher step
		go func() {
			if err := startCmd.RunE(startCmd, nil); err != nil {
				panic(errors.Wrapf(err, "failed to run command"))
			}
		}()

		<-syncDone

		return nil
	}
}
