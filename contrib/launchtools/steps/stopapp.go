package steps

import (
	"errors"
	launchertypes "github.com/initia-labs/OPinit/contrib/launchtools"
)

func StopApp(_ launchertypes.Input) launchertypes.LauncherStepFunc {
	return func(ctx launchertypes.Launcher) error {
		if !ctx.IsAppInitialized() {
			return errors.New("app is not initialized")
		}

		log := ctx.Logger()
		log.Info("cleanup")

		return nil
	}
}
