package steps

import "github.com/initia-labs/OPinit/contrib/launchtools"

var _ launchtools.LauncherStepFuncFactory[*launchtools.Config] = InitializeConfig

// InitializeConfig sets the config for the server context.
func InitializeConfig(config *launchtools.Config) launchtools.LauncherStepFunc {
	return func(ctx launchtools.Launcher) error {
		// set config
		serverConfig := ctx.ServerContext().Config
		serverConfig.Moniker = config.L2Config.Moniker

		return nil
	}
}
