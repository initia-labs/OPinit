package steps

import (
	"errors"
	"syscall"

	launchertypes "github.com/initia-labs/OPinit/contrib/launchtools"
)

var _ launchertypes.LauncherStepFuncFactory[*launchertypes.Config] = StopApp

func StopApp(_ *launchertypes.Config) launchertypes.LauncherStepFunc {
	return func(ctx launchertypes.Launcher) error {
		if !ctx.IsAppInitialized() {
			return errors.New("app is not initialized")
		}

		log := ctx.Logger()
		log.Info("cleanup")
		log.Info("waiting for app to stop")

		err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		if err != nil {
			log.Error("failed to raise a kill signal", "error", err)
		}

		// wait for the app to stop
		err = ctx.GetErrorGroup().Wait()
		if err != nil {
			log.Error("cleanup failed", "error", err)
			return err
		}

		// wait for the app to stop completely (release ports)
		ctx.WaitApp()

		log.Info("cleanup finished")
		return nil
	}
}
