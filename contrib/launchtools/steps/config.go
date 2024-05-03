package steps

import "github.com/initia-labs/OPinit/contrib/launchtools"

// InitializeConfig sets the config for the server context.
func InitializeConfig(manifest launchtools.Input) launchtools.LauncherStepFunc {
	return func(ctx launchtools.Launcher) error {
		// set config
		config := ctx.ServerContext().Config
		config.Moniker = manifest.L2Config.Moniker

		return nil
	}
}
